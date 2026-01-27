// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package oam

import (
	"context"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
)

const PollingInterval = 5000

func oamGrpcServer(
	lis net.Listener,
	termChan chan bool,
	readyChan chan bool,
	wg *sync.WaitGroup,
	enableTracing bool,
) {
	var srvOpts []grpc.ServerOption

	if enableTracing {
		srvOpts = tracing.EnableGrpcServerTracing(srvOpts)
	}
	gsrv := grpc.NewServer(srvOpts...)
	oam := NewOAM()
	// register server - oamServer
	grpc_health_v1.RegisterHealthServer(gsrv, oam)

	// enable reflection
	reflection.Register(gsrv)

	// serve in goroutine
	go func() {
		err := gsrv.Serve(lis)
		if err != nil {
			zlog.InfraSec().Fatal().Err(err).Msg("failed to serve")
		}
	}()

	for {
		// Blocking wait on:
		// 1) readiness changes
		// 2) term signals
		select {
		case ready := <-readyChan:
			oam.SetReady(ready)
			if ready {
				zlog.InfraSec().Info().Msg("service is set to ready")
			} else {
				zlog.InfraSec().Info().Msg("service is set to not ready")
			}
		case termSig := <-termChan:
			// handle termination signals
			if termSig {
				gsrv.Stop()
				zlog.InfraSec().Info().Msg("stopping server")
			}
			// exit WaitGroup when done
			wg.Done()
			return
		}
	}
}

// StartOamGrpcServer is the the functional interface to create and start the OAM server.
// termChan is used to signal the graceful shutdown of the application, instead readyChan
// to signal the readiness of the service, wg is used to coordinate the termination of
// several flow of executions (see in the cmd package for usage examples), servAddr is the
// server address to listen for and enableTracing to turn on the Infra tracing.
//
// Note that the coordination around the readiness of the pod is realized by establishing a
// channel between the OAM gRPC server and a second control flow handling the initialization
// of the pod - this is the main purpose of the readyChan which is used to asynchronously notify
// the server about the readiness of the pod.
//
// Users should close the termChan for the graceful shutdown of the server.
func StartOamGrpcServer(
	termChan chan bool,
	readyChan chan bool,
	wg *sync.WaitGroup,
	servaddr string,
	enableTracing bool,
) error {
	zlog.InfraSec().Info().Str("address", servaddr).Msg("started to listen")

	var lis net.Listener
	err := error(nil)

	// if testing, use a bufconn, otherwise TCP
	if servaddr == "bufconn" {
		lis = net.Listener(TestBufconn)
	} else {
		lc := &net.ListenConfig{}
		lis, err = lc.Listen(context.Background(), "tcp", servaddr)
		if err != nil {
			zlog.Fatal().Err(err).Msgf("Error listening with TCP: %s", servaddr)
			return errors.Wrap(err)
		}
	}

	oamGrpcServer(lis, termChan, readyChan, wg, enableTracing)
	return err
}
