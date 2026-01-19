// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/common"
	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/oam"
	_ "github.com/open-edge-platform/infra-core/inventory/v2/pkg/perf" // Adds support for pprof.
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
)

var zlog = logging.GetLogger("server-api")

var (
	RepoURL   = "https://github.com/open-edge-platform/infra-core/apiv2/v2.git"
	Version   = "<unset>"
	Revision  = "<unset>"
	BuildDate = "<unset>"
)

var (
	wg        = sync.WaitGroup{}        // waitgroup so main will wait for all go routines to exit cleanly
	readyChan = make(chan bool, 1)      // channel to signal the readiness.
	termChan  = make(chan bool, 1)      // channel to signal termination of main process.
	sigChan   = make(chan os.Signal, 1) // channel to handle any interrupt signals
)
var errFatal error

func fatal(e error) {
	errFatal = e
	zlog.Fatal().Err(e).Msg("fatal error")
}

func printSummary() {
	zlog.Info().Msgf("Starting Server API")
	zlog.InfraSec().Info().Msgf("RepoURL: %s, Version: %s, Revision: %s, BuildDate: %s\n",
		RepoURL, Version, Revision, BuildDate)
}

func traces(cfg *common.GlobalConfig) func(context.Context) error {
	cleanup, exportErr := tracing.NewTraceExporterHTTP(cfg.Traces.TraceURL, "rest-api", nil)
	if exportErr != nil {
		zlog.Err(exportErr).Msg("Error creating trace exporter")
	}
	if cleanup != nil {
		zlog.Info().Msg("Tracing enabled")
	} else {
		zlog.Info().Msg("Tracing disabled")
	}
	return cleanup
}

func setOAM(cfg *common.GlobalConfig, termChan, readyChan chan bool, wg *sync.WaitGroup) {
	if cfg.RestServer.OamServerAddr != "" {
		// Add oam grpc server
		wg.Add(1)
		go func() {
			if err := oam.StartOamGrpcServer(termChan, readyChan, wg,
				cfg.RestServer.OamServerAddr, cfg.Traces.EnableTracing); err != nil {
				zlog.InfraSec().Err(err).Msg("failed to start OAM")
				fatal(err)
			}
		}()
	}
}

func main() {
	// Print a summary of the build
	printSummary()

	defer func() {
		if errFatal != nil {
			os.Exit(1)
		}
	}()

	cfg, err := common.Config()
	if err != nil {
		zlog.InfraSec().Err(err).Msg("Failed to get gRPC server configuration")
		fatal(err)
	}

	if cfg.Traces.EnableTracing {
		cleanup := traces(cfg)
		if cleanup != nil {
			defer func() {
				cleanErr := cleanup(context.Background())
				if cleanErr != nil {
					zlog.Err(cleanErr).Msg("Error in tracing cleanup")
				}
			}()
		}
	}

	setOAM(cfg, termChan, readyChan, &wg)

	ctx, cancel := context.WithCancel(context.Background())
	invServer, err := server.NewInventoryServer(ctx, &wg, cfg)
	if err != nil {
		zlog.InfraSec().Err(err).Msg("Failed to start gRPC server")
		fatal(err)
	}

	// Add inventory grpc server Start
	wg.Add(1)
	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCAddress)
		if err != nil {
			zlog.InfraSec().Err(err).Msgf("Error listening with TCP on address %s", cfg.GRPCAddress)
			fatal(err)
		}
		defer lis.Close()
		invServer.Start(
			lis,
			termChan,
			readyChan,
			&wg,
			cfg.Traces.EnableTracing,
			true,
			cfg.Inventory.CAPath,
			cfg.Inventory.CertPath,
			cfg.Inventory.KeyPath,
			cfg.RestServer.Authentication,
			cfg.Scenario,
		)
	}()

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		exit := <-sigChan
		zlog.Info().Msgf("Received exit signal %v", exit)
		cancel()
		termChan <- true
		close(termChan)
	}()

	// wait until servers terminate
	wg.Wait()
	zlog.Info().Msgf("Shutdown Server API")
}
