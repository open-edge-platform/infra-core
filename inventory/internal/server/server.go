// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	inv_impl "github.com/open-edge-platform/infra-core/inventory/v2/internal/inventory"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/auditing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/cert"
	inv_errors "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
	"github.com/open-edge-platform/orch-library/go/pkg/grpc/auth"
)

var zlog = logging.GetLogger("InfraInvSrvgRPC")

var (
	// Ping the client if it is idle for 1 minute to ensure the connection is still active.
	serverKeepAliveTime = 1 * time.Minute
	// Wait 1 second for the ping ack before assuming the connection is dead.
	serverKeepAliveTimeout = 1 * time.Second
	// If a client pings more than once every 1 minute, terminate the connection.
	clientMinTime = 1 * time.Minute
)

type Options struct {
	EnableTracing  bool
	EnableAuth     bool
	EnableMetrics  bool
	MetricsAddress string
	EnableAuditing bool
	InsecureGrpc   bool
	CaCertPath     string
	TLSCertPath    string
	TLSKeyPath     string
}

// Metrics server definition, you need to register a gRPC server and start the server to actually serve metrics.
var srvMetrics = metrics.GetServerMetricsWithLatency()

func invalidSecureConfig(caCertPath, tlsCertPath, tlsKeyPath string) bool {
	return caCertPath == "" || tlsCertPath == "" || tlsKeyPath == ""
}

func GetAuthOpts(caCertPath, tlsCertPath, tlsKeyPath string) (grpc.ServerOption, error) {
	// setting secure gRPC connection
	if invalidSecureConfig(caCertPath, tlsCertPath, tlsKeyPath) {
		zlog.InfraSec().Error().Msgf("CaCertPath %s or TlsCerPath %s or TlsKeyPath %s were not provided\n",
			caCertPath, tlsCertPath, tlsKeyPath,
		)
		return nil, inv_errors.Errorf("CaCertPath %s or TlsCerPath %s or TlsKeyPath %s were not provided",
			caCertPath, tlsCertPath, tlsKeyPath,
		)
	}
	creds, err := cert.HandleCertPaths(caCertPath, tlsKeyPath, tlsCertPath, true)
	if err != nil {
		zlog.InfraSec().Err(err).Msgf("an error occurred while loading credentials to server %v, %v, %v: %v\n",
			caCertPath, tlsCertPath, tlsKeyPath, err,
		)
		return nil, inv_errors.Wrap(err)
	}
	return grpc.Creds(creds), nil
}

func GetServerOpts(opts Options) ([]grpc.ServerOption, error) {
	var srvOpts []grpc.ServerOption
	var unaryInter []grpc.UnaryServerInterceptor
	var streamInter []grpc.StreamServerInterceptor

	unaryInter = append(unaryInter, TenantContextExtractingInterceptor())

	if opts.EnableMetrics {
		zlog.Info().Msgf("Metrics exporter is enabled")
		unaryInter = append(unaryInter, srvMetrics.UnaryServerInterceptor())
		streamInter = append(streamInter, srvMetrics.StreamServerInterceptor())
	}

	if !opts.InsecureGrpc {
		authOpts, err := GetAuthOpts(opts.CaCertPath, opts.TLSCertPath, opts.TLSKeyPath)
		if err != nil {
			return nil, err
		}
		srvOpts = append(srvOpts, authOpts)
	}

	if opts.EnableAuth {
		zlog.InfraSec().Info().Msgf("Authentication is enabled")
		unaryInter = append(unaryInter, grpc_auth.UnaryServerInterceptor(auth.AuthenticationInterceptor))
		streamInter = append(streamInter, grpc_auth.StreamServerInterceptor(auth.AuthenticationInterceptor))
	}

	if opts.EnableTracing {
		srvOpts = append(srvOpts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	// TODO ITEP-2566 move this before Auth
	if opts.EnableAuditing {
		unaryInter = append(unaryInter, auditing.GrpcInterceptor)
	}

	keepAliveOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			// MaxConnectionIdle:     ,	// Leave it to default, infinite.
			// MaxConnectionAge:      ,	// Leave it to default, infinite.
			// MaxConnectionAgeGrace: ,	// Leave it to default, infinite.
			Time:    serverKeepAliveTime,
			Timeout: serverKeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             clientMinTime,
			PermitWithoutStream: true,
		}),
	}
	srvOpts = append(srvOpts, keepAliveOpts...)

	// adding unary and stream interceptors
	srvOpts = append(srvOpts,
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(unaryInter...)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(streamInter...)))
	return srvOpts, nil
}

func inventoryGrpcServer(
	lis net.Listener,
	termChan chan bool,
	readyChan chan bool,
	wg *sync.WaitGroup,
	dbURLWriter string,
	dbURLReader string,
	policyBundle string,
	opts Options,
) {
	srvOpts, err := GetServerOpts(opts)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to get server opts")
	}

	gsrv := grpc.NewServer(srvOpts...)

	// register server - inventoryServer
	inv_v1.RegisterInventoryServiceServer(
		gsrv, inv_impl.NewInventoryServer(dbURLWriter, dbURLReader, policyBundle, opts.EnableTracing, opts.EnableAuth))

	// enable reflection
	reflection.Register(gsrv)

	if opts.EnableMetrics {
		// Register metrics
		srvMetrics.InitializeMetrics(gsrv)
		// Start metrics exporter server
		metrics.StartMetricsExporter([]prometheus.Collector{srvMetrics}, metrics.WithListenAddress(opts.MetricsAddress))
	}

	// in goroutine signal is ready and then serve
	go func() {
		// On testing will be nil
		if readyChan != nil {
			readyChan <- true
		}

		err := gsrv.Serve(lis)
		if err != nil {
			zlog.InfraSec().Fatal().Err(err).Msg("failed to serve")
		}
	}()

	// handle termination signals
	termSig := <-termChan
	if termSig {
		gsrv.Stop()
		zlog.Info().Msg("stopping server")
	}

	// exit WaitGroup when done
	wg.Done()
}

func StartInventoryGrpcServer(
	termChan chan bool,
	readyChan chan bool,
	wg *sync.WaitGroup,
	lis net.Listener,
	dbURLWriter string,
	dbURLReader string,
	policyBundle string,
	opts Options,
) {
	zlog.InfraSec().Info().Str("address", lis.Addr().String()).Msg("started to listen")
	inventoryGrpcServer(lis, termChan, readyChan, wg, dbURLWriter, dbURLReader, policyBundle, opts)
}
