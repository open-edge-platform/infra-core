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

// servicesClients defines a list of all gRPC service clients that must be
// registered to serve REST API.
var servicesClients = []serviceClientsSignature{
	restv1.RegisterRegionServiceHandlerFromEndpoint,
	restv1.RegisterSiteServiceHandlerFromEndpoint,
	restv1.RegisterLocationServiceHandlerFromEndpoint,
	restv1.RegisterHostServiceHandlerFromEndpoint,
	restv1.RegisterOperatingSystemServiceHandlerFromEndpoint,
	restv1.RegisterInstanceServiceHandlerFromEndpoint,
	restv1.RegisterScheduleServiceHandlerFromEndpoint,
	restv1.RegisterWorkloadServiceHandlerFromEndpoint,
	restv1.RegisterWorkloadMemberServiceHandlerFromEndpoint,
	restv1.RegisterProviderServiceHandlerFromEndpoint,
	restv1.RegisterTelemetryLogsGroupServiceHandlerFromEndpoint,
	restv1.RegisterTelemetryMetricsGroupServiceHandlerFromEndpoint,
	restv1.RegisterTelemetryMetricsProfileServiceHandlerFromEndpoint,
	restv1.RegisterTelemetryLogsProfileServiceHandlerFromEndpoint,
	restv1.RegisterLocalAccountServiceHandlerFromEndpoint,
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
	for _, serviceClient := range servicesClients {
		err := serviceClient(m.ctx, mux, m.cfg.GRPCEndpoint,
			[]grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			})
		if err != nil {
			zlog.InfraErr(err).Msgf("failed to set service client %v", serviceClient)
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
