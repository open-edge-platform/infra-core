// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/common"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/scenario"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/cert"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	schedule_cache "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache/schedule"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
	"github.com/open-edge-platform/orch-library/go/pkg/grpc/auth"
)

var zlog = logging.GetLogger("nbi")

// serviceServersSignature defines a signature for a gRPC service registration function.
type serviceServersSignature func(*grpc.Server, *InventorygRPCServer)

// servicesServers maps gRPC service names to their registration functions.
// These are used to register services conditionally based on the scenario allowlist.
// service name must match one of the service names used in api/proto/services/v1/services.proto.
var servicesServers = map[string]serviceServersSignature{
	"RegionService":   func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterRegionServiceServer(s, is) },
	"SiteService":     func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterSiteServiceServer(s, is) },
	"LocationService": func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterLocationServiceServer(s, is) },
	"HostService":     func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterHostServiceServer(s, is) },
	"OperatingSystemService": func(s *grpc.Server, is *InventorygRPCServer) {
		restv1.RegisterOperatingSystemServiceServer(s, is)
	},
	"InstanceService": func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterInstanceServiceServer(s, is) },
	"ScheduleService": func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterScheduleServiceServer(s, is) },
	"WorkloadService": func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterWorkloadServiceServer(s, is) },
	"WorkloadMemberService": func(s *grpc.Server, is *InventorygRPCServer) {
		restv1.RegisterWorkloadMemberServiceServer(s, is)
	},
	"ProviderService": func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterProviderServiceServer(s, is) },
	"TelemetryLogsGroupService": func(s *grpc.Server, is *InventorygRPCServer) {
		restv1.RegisterTelemetryLogsGroupServiceServer(s, is)
	},
	"TelemetryMetricsGroupService": func(s *grpc.Server, is *InventorygRPCServer) {
		restv1.RegisterTelemetryMetricsGroupServiceServer(s, is)
	},
	"TelemetryMetricsProfileService": func(s *grpc.Server, is *InventorygRPCServer) {
		restv1.RegisterTelemetryMetricsProfileServiceServer(s, is)
	},
	"TelemetryLogsProfileService": func(s *grpc.Server, is *InventorygRPCServer) {
		restv1.RegisterTelemetryLogsProfileServiceServer(s, is)
	},
	"LocalAccountService":   func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterLocalAccountServiceServer(s, is) },
	"CustomConfigService":   func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterCustomConfigServiceServer(s, is) },
	"OSUpdatePolicyService": func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterOSUpdatePolicyServer(s, is) },
	"OSUpdateRunService":    func(s *grpc.Server, is *InventorygRPCServer) { restv1.RegisterOSUpdateRunServer(s, is) },
}

type InventorygRPCServer struct {
	InvClient       client.InventoryClient
	InvHCacheClient *schedule_cache.HScheduleCacheClient
}

// GetAPIRoles helper function to get role used by API component.
// It can be used to feed the expected roles of interceptor in APIv2.
func GetAPIRoles() []string {
	return []string{
		"im-rw",
		"im-r",
	}
}

func NewInventoryServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	config *common.GlobalConfig,
) (*InventorygRPCServer, error) {
	InvClient, err := NewInventoryClient(ctx, wg, config)
	if err != nil {
		return nil, err
	}
	invHCacheClient, err := NewInventoryHCacheClient(ctx, config)
	if err != nil {
		return nil, err
	}
	return &InventorygRPCServer{
		InvClient:       InvClient,
		InvHCacheClient: invHCacheClient,
	}, nil
}

func invalidSecureConfig(caCertPath, tlsCertPath, tlsKeyPath string) bool {
	return caCertPath == "" || tlsCertPath == "" || tlsKeyPath == ""
}

func getServerOpts(enableTracing, enableAuth, insecureGrpc bool,
	caCertPath, tlsCertPath, tlsKeyPath string,
) ([]grpc.ServerOption, error) {
	var srvOpts []grpc.ServerOption

	if !insecureGrpc {
		// setting secure gRPC connection
		if invalidSecureConfig(caCertPath, tlsCertPath, tlsCertPath) {
			zlog.InfraSec().Fatal().Msgf("CaCertPath %s or TlsCerPath %s or TlsKeyPath %s were not provided\n",
				caCertPath, tlsCertPath, tlsKeyPath,
			)
			return nil, errors.Errorf("CaCertPath %s or TlsCerPath %s or TlsKeyPath %s were not provided",
				caCertPath, tlsCertPath, tlsKeyPath,
			)
		}
		creds, err := cert.HandleCertPaths(caCertPath, tlsKeyPath, tlsCertPath, true)
		if err != nil {
			zlog.InfraSec().Fatal().Err(err).Msgf("an error occurred while loading credentials to server %v, %v, %v: %v\n",
				caCertPath, tlsCertPath, tlsKeyPath, err,
			)
			return nil, errors.Wrap(err)
		}
		srvOpts = append(srvOpts, grpc.Creds(creds))
	}

	unaryInter := []grpc.UnaryServerInterceptor{}
	streamInter := []grpc.StreamServerInterceptor{}

	if enableAuth {
		zlog.InfraSec().Info().Msgf("Authentication is enabled")
		// Adds tenantID interceptor before Authenticator
		unaryInter = append(unaryInter,
			tenant.GetExtractTenantIDInterceptor(GetAPIRoles()),
			grpc_auth.UnaryServerInterceptor(auth.AuthenticationInterceptor))
		streamInter = append(streamInter, grpc_auth.StreamServerInterceptor(auth.AuthenticationInterceptor))
	}

	if enableTracing {
		srvOpts = append(srvOpts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	// adding unary and stream interceptors
	srvOpts = append(srvOpts,
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(unaryInter...)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(streamInter...)))
	return srvOpts, nil
}

// setupServices registers gRPC services based on the scenario allowlist.
func (is *InventorygRPCServer) setupServices(gsrv *grpc.Server, scenarioName string) error {
	if scenarioName == "" {
		return fmt.Errorf("scenario name is not set")
	}

	// build a map of allowed services for quick lookup
	allowed, err := BuildAllowedServiceList(scenarioName, scenario.Allowlist)
	if err != nil {
		return err
	}

	for serviceName, registerFunc := range servicesServers {
		if _, isAllowed := allowed[serviceName]; !isAllowed {
			zlog.Debug().Str("service", serviceName).Str("scenario", scenarioName).
				Msg("skipping service not allowed for scenario")
			continue
		}

		registerFunc(gsrv, is)
		zlog.Info().Str("service", serviceName).Str("scenario", scenarioName).
			Msg("registered gRPC service")
	}

	return nil
}

func (is *InventorygRPCServer) Start(
	lis net.Listener,
	termChan chan bool,
	readyChan chan bool,
	wg *sync.WaitGroup,
	enableTracing bool,
	insecureGrpc bool,
	caCertPath,
	tlsCertPath,
	tlsKeyPath string,
	enableAuth bool,
	scenarioName string,
) {
	srvOpts, err := getServerOpts(enableTracing, enableAuth, insecureGrpc, caCertPath, tlsCertPath, tlsKeyPath)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to get server opts")
	}

	gsrv := grpc.NewServer(srvOpts...)

	// register server - inventoryServer based on scenario
	if err := is.setupServices(gsrv, scenarioName); err != nil {
		zlog.Fatal().Err(err).Msg("failed to setup services")
	}

	// enable reflection
	reflection.Register(gsrv)

	// in goroutine signal is ready and then serve
	go func() {
		// On testing will be nil
		if readyChan != nil {
			readyChan <- true
		}

		zlog.Info().Msg("Starting gRPC server")
		err := gsrv.Serve(lis)
		if err != nil {
			zlog.InfraSec().Fatal().Err(err).Msg("failed to serve")
		}
	}()
	zlog.Info().Msg("Started gRPC server")

	// handle termination signals
	termSig := <-termChan
	if termSig {
		zlog.Info().Msg("Stopping gRPC server")
		gsrv.GracefulStop()
		zlog.Info().Msg("Stopped gRPC server")
	}

	// exit WaitGroup when done
	wg.Done()
}
