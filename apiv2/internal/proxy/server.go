// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/common"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	api "github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	inv_client "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	ginutils "github.com/open-edge-platform/orch-library/go/pkg/middleware/gin"
)

var zlog = logging.GetLogger("proxy")

// serviceClientsSignature defines a signature for a gRPC client registration function.
type serviceClientsSignature func(
	ctx context.Context,
	mux *runtime.ServeMux,
	endpoint string,
	opts []grpc.DialOption) (err error)

// servicesClients maps gRPC service names to their grpc-gateway registration functions.
// Each function registers HTTP handlers for the corresponding service,
// enabling REST → gRPC proxying through the gateway.
// Keys must match the protobuf service names; registration funcs are generated in
// apiv2/v2/internal/pbapi/services/v1 by grpc-gateway.
// setupClients() filters and registers only the services allowed for the active SCENARIO.
var servicesClients = map[string]serviceClientsSignature{
	"RegionService":                  restv1.RegisterRegionServiceHandlerFromEndpoint,
	"SiteService":                    restv1.RegisterSiteServiceHandlerFromEndpoint,
	"LocationService":                restv1.RegisterLocationServiceHandlerFromEndpoint,
	"HostService":                    restv1.RegisterHostServiceHandlerFromEndpoint,
	"OperatingSystemService":         restv1.RegisterOperatingSystemServiceHandlerFromEndpoint,
	"InstanceService":                restv1.RegisterInstanceServiceHandlerFromEndpoint,
	"ScheduleService":                restv1.RegisterScheduleServiceHandlerFromEndpoint,
	"WorkloadService":                restv1.RegisterWorkloadServiceHandlerFromEndpoint,
	"WorkloadMemberService":          restv1.RegisterWorkloadMemberServiceHandlerFromEndpoint,
	"ProviderService":                restv1.RegisterProviderServiceHandlerFromEndpoint,
	"TelemetryLogsGroupService":      restv1.RegisterTelemetryLogsGroupServiceHandlerFromEndpoint,
	"TelemetryMetricsGroupService":   restv1.RegisterTelemetryMetricsGroupServiceHandlerFromEndpoint,
	"TelemetryMetricsProfileService": restv1.RegisterTelemetryMetricsProfileServiceHandlerFromEndpoint,
	"TelemetryLogsProfileService":    restv1.RegisterTelemetryLogsProfileServiceHandlerFromEndpoint,
	"LocalAccountService":            restv1.RegisterLocalAccountServiceHandlerFromEndpoint,
	"CustomConfigService":            restv1.RegisterCustomConfigServiceHandlerFromEndpoint,
	"OSUpdatePolicy":                 restv1.RegisterOSUpdatePolicyHandlerFromEndpoint,
	"OSUpdateRun":                    restv1.RegisterOSUpdateRunHandlerFromEndpoint,
}

const (
	apiTraceName = "miAPIEchoServer"
)

type Manager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	echoServer *echo.Echo
	// wsHandler  *worker_handlers.WebsocketHandler
	cfg   *common.GlobalConfig
	Ready chan bool
}

func NewManager(cfg *common.GlobalConfig, ready chan bool) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:        ctx,
		cancel:     cancel,
		echoServer: echo.New(),
		cfg:        cfg,
		Ready:      ready,
	}, nil
}

// WrapH wraps an http Handler into an echo Middleware implementation.
func WrapH(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		h.ServeHTTP(res, req)
		return nil
	}
}

func (m *Manager) setupClients(mux *runtime.ServeMux) error {
	scenarioName := "vpro" // strings.TrimSpace(os.Getenv("SCENARIO"))
	if scenarioName == "" {
		return fmt.Errorf("SCENARIO env var is not set")
	}

	// build a map of allowed services for quick lookup
	allowed, err := BuildAllowedClientList(scenarioName)
	if err != nil {
		return err
	}

	for serviceName, serviceClient := range servicesClients {
		if _, isAllowed := allowed[serviceName]; !isAllowed {
			zlog.Debug().Str("service", serviceName).Str("scenario", scenarioName).Msg("skipping service client not allowed for scenario")
			continue
		}

		if err := serviceClient(
			m.ctx,
			mux,
			m.cfg.GRPCEndpoint,
			[]grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				// Use Inventory client max message size, to keep Inventory and API consistent.
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(inv_client.MaxMessageSize)),
			},
		); err != nil {
			zlog.InfraErr(err).Str("service", serviceName).Str("scenario", scenarioName).Msg("failed to set service client")
			return err
		}
	}

	return nil
}

const ActiveProjectID = "ActiveProjectID"

func (m *Manager) Start() error {
	// creating mux for gRPC gateway. This will multiplex or route request different gRPC services.
	mux := runtime.NewServeMux(
		// convert header in response(going from gateway) from metadata received.
		runtime.WithMetadata(func(_ context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			uaHeader := request.Header.Get("User-Agent")
			projectIDHeader := request.Header.Get(ActiveProjectID)
			// send all the headers received from the client
			md := metadata.Pairs("authorization", authHeader, "user-agent", uaHeader, "activeprojectid", projectIDHeader)
			return md
		}),
		runtime.WithRoutingErrorHandler(ginutils.HandleRoutingError),
		runtime.WithErrorHandler(customErrorHandler),
	)

	err := m.setupClients(mux)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to setup gRPC clients")
		return err
	}

	e := echo.New()
	m.setOptions(e)

	openAPIDefinition, err := api.GetSwagger()
	if err != nil {
		zlog.InfraErr(err).Msg("failed to GetSwagger")
		return err
	}

	for _, s := range openAPIDefinition.Servers {
		zlog.Info().Str("url", s.URL).Msgf("Servers")
		s.URL = strings.ReplaceAll(s.URL, "{apiRoot}", "")
	}

	if m.cfg.RestServer.EnableMetrics {
		zlog.Info().Msgf("Metrics exporter is enabled")
		m.enableMetrics(e)
	}

	zlog.Info().Str("baseUrl", m.cfg.RestServer.BaseURL).Msgf("Registering handlers")
	gatewayURL := fmt.Sprintf("%s/*{grpc_gateway}", m.cfg.RestServer.BaseURL)
	zlog.Info().Str("gatewayURL", m.cfg.RestServer.BaseURL).Msgf("Group Proxy URL")
	g := e.Group(gatewayURL)
	g.Match(allowMethods, "", WrapH(mux))

	zlog.Info().Str("address", m.cfg.RestServer.Address).Msgf("Starting REST server")

	m.echoServer = e
	m.Ready <- true
	return e.Start(m.cfg.RestServer.Address)
}

func (m *Manager) Stop(ctx context.Context) error {
	m.cancel()
	return m.echoServer.Shutdown(ctx)
}
