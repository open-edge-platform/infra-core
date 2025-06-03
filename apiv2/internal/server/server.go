// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/common"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/cert"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	schedule_cache "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache/schedule"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
	"github.com/open-edge-platform/orch-library/go/pkg/grpc/auth"
)

var zlog = logging.GetLogger("nbi")

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
) {
	srvOpts, err := getServerOpts(enableTracing, enableAuth, insecureGrpc, caCertPath, tlsCertPath, tlsKeyPath)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to get server opts")
	}

	gsrv := grpc.NewServer(srvOpts...)

	// register server - inventoryServer
	restv1.RegisterRegionServiceServer(gsrv, is)
	restv1.RegisterSiteServiceServer(gsrv, is)
	restv1.RegisterLocationServiceServer(gsrv, is)
	restv1.RegisterHostServiceServer(gsrv, is)
	restv1.RegisterInstanceServiceServer(gsrv, is)
	restv1.RegisterScheduleServiceServer(gsrv, is)
	restv1.RegisterOperatingSystemServiceServer(gsrv, is)
	restv1.RegisterWorkloadServiceServer(gsrv, is)
	restv1.RegisterWorkloadMemberServiceServer(gsrv, is)
	restv1.RegisterTelemetryLogsGroupServiceServer(gsrv, is)
	restv1.RegisterTelemetryMetricsGroupServiceServer(gsrv, is)
	restv1.RegisterTelemetryLogsProfileServiceServer(gsrv, is)
	restv1.RegisterTelemetryMetricsProfileServiceServer(gsrv, is)
	restv1.RegisterProviderServiceServer(gsrv, is)
	restv1.RegisterLocalAccountServiceServer(gsrv, is)
	restv1.RegisterOSUpdateRunServer(gsrv, is)
	restv1.RegisterOSUpdatePolicyServer(gsrv, is)

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
