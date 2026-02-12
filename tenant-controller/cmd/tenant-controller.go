// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/oam"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/controller"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/datamodel"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/invclient"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/nexus"
)

// Configuration variables, mostly set by flags.
var (
	zlog = logging.GetLogger(configuration.AppName + "Main")

	inventoryAddress = flag.String(client.InventoryAddress, "localhost:50051", client.InventoryAddressDescription)
	oamServerAddress = flag.String(oam.OamServerAddress, "", oam.OamServerAddressDescription)

	enableTracing = flag.Bool(tracing.EnableTracing, false, tracing.EnableTracingDescription)
	traceURL      = flag.String(tracing.TraceURL, "", tracing.TraceURLDescription)

	insecureGrpc = flag.Bool(client.InsecureGrpc, true, client.InsecureGrpcDescription)
	caCertPath   = flag.String(client.CaCertPath, "", client.CaCertPathDescription)
	tlsCertPath  = flag.String(client.TLSCertPath, "", client.TLSCertPathDescription)
	tlsKeyPath   = flag.String(client.TLSKeyPath, "", client.TLSKeyPathDescription)

	enableMetrics  = flag.Bool(metrics.EnableMetrics, false, metrics.EnableMetricsDescription)
	metricsAddress = flag.String(metrics.MetricsAddress, metrics.MetricsAddressDefault, metrics.MetricsAddressDescription)

	initResourcesDefinitionPath = flag.String(
		"initResourcesDefinitionPath",
		"",
		"path to the file containing list of resources to be initialized on every new tenant creation")
	lenovoResourcesDefinitionPath = flag.String(
		"lenovoResourcesDefinitionPath",
		"",
		"path to the file containing list of Lenovo resources to be initialized on every new tenant creation")
	skipOSProvisioning = flag.Bool(
		"skipOSProvisioning",
		false,
		"Skip waiting for OS Resource Manager watcher acknowledgment when true")
)

// Project related variables. Overwritten by build process.
var (
	RepoURL   = "https://github.com/open-edge-platform/infra-core/tenant-controller.git"
	Version   = "<unset>"
	Revision  = "<unset>"
	BuildDate = "<unset>"
)

// Waitgroups and channels used for readiness and program exit.
var (
	wg           = sync.WaitGroup{} // all goroutines added to this, blocks program exit
	invReadyChan = make(chan bool, 1)
	oamReadyChan = make(chan bool, 1) // used for readiness indicators to OAM
	termChan     = make(chan bool, 1) // used to pass on termination signals
	sigChan      = make(chan os.Signal, 1)
)

func StartupSummary() {
	zlog.Info().Msg("Starting " + configuration.AppName)
	zlog.Info().Msgf("RepoURL: %s, Version: %s, Revision: %s, BuildDate: %s\n", RepoURL, Version, Revision, BuildDate)
}

func SetupTracing(traceURL string) func(context.Context) error {
	cleanup, exportErr := tracing.NewTraceExporterHTTP(traceURL, configuration.AppName, nil)
	if exportErr != nil {
		zlog.Err(exportErr).Msg("Error creating trace exporter")
	}
	if cleanup != nil {
		zlog.Info().Msgf("Tracing enabled %s", traceURL)
	} else {
		zlog.Info().Msg("Tracing disabled")
	}
	return cleanup
}

func GetSecurityConfig() *client.SecurityConfig {
	secCfg := &client.SecurityConfig{
		CaPath:   *caCertPath,
		CertPath: *tlsCertPath,
		KeyPath:  *tlsKeyPath,
		Insecure: *insecureGrpc,
	}
	return secCfg
}

func SetupOamServerAndSetReady(enableTracing bool, oamServerAddress string) {
	if oamServerAddress != "" {
		wg.Add(1) // Add oam grpc server to waitgroup

		go func() {
			if err := oam.StartOamGrpcServer(termChan, oamReadyChan, &wg, oamServerAddress, enableTracing); err != nil {
				zlog.InfraSec().Fatal().Err(err).Msg("Cannot start " + configuration.AppName + " gRPC server")
			}
		}()
	}
}

func startMetricsServer() {
	metrics.StartMetricsExporter([]prometheus.Collector{metrics.GetClientMetricsWithLatency()},
		metrics.WithListenAddress(*metricsAddress))
}

func main() {
	// Print a summary of build information
	StartupSummary()

	// Parse flags
	flag.Parse()

	// Tracing, if enabled
	if *enableTracing {
		cleanup := SetupTracing(*traceURL)
		if cleanup != nil {
			defer func() {
				err := cleanup(context.Background())
				if err != nil {
					zlog.Err(err).Msg("Error in tracing cleanup")
				}
			}()
		}
	}

	if *enableMetrics {
		startMetricsServer()
	}

	// connect to Inventory
	invClient, err := invclient.NewInventoryClientWithOptions(
		invReadyChan,
		termChan,
		&wg,
		GetSecurityConfig(),
		invclient.WithInventoryAddress(*inventoryAddress),
		invclient.WithEnableTracing(*enableTracing),
		invclient.WithEnableMetrics(*enableMetrics),
	)
	if err != nil {
		zlog.InfraSec().Fatal().Err(err).Msgf("Unable to start Inventory client")
	}

	initialResourcesProviders, err := getInitResourceProviders()
	if err != nil {
		zlog.InfraSec().Fatal().Err(err).Msgf("")
	}

	nxc, err := nexus.SetupClient()
	if err != nil {
		zlog.InfraSec().Fatal().Err(err).Msgf("Unable to setup Nexus Client")
	}

	tenantTerminationCtrl := controller.NewTerminationController(invClient)
	tenantInitializationCtrl := controller.NewTenantInitializationController(
		initialResourcesProviders, invClient, nxc, *skipOSProvisioning)

	controller.NewEventDispatcher(invClient, tenantInitializationCtrl, tenantTerminationCtrl).Start(termChan)

	dmc := datamodel.NewDataModelController(nxc, *enableTracing, tenantTerminationCtrl, tenantInitializationCtrl)
	if err := dmc.Start(termChan); err != nil {
		zlog.InfraSec().Fatal().Err(err).Msgf("Unable to start DataModel controller")
	}

	// set up OAM (health check) server
	SetupOamServerAndSetReady(*enableTracing, *oamServerAddress)

	// Handle OS signals (ctrl-c, etc.)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan // block until signals received

		close(termChan) // closes the SBgRPC server, OAM Server

		invClient.Close() // stop inventory client
	}()

	// wait for Inventory API and SB gRPC API to be ready, then set OAM ready
	go func() {
		<-invReadyChan
		oamReadyChan <- true
	}()

	wg.Wait()
}

func getInitResourceProviders() ([]configuration.InitResourcesProvider, error) {
	loader, err := configuration.NewInitResourcesProvider(*initResourcesDefinitionPath)
	if err != nil {
		return nil, fmt.Errorf("given path('%s') points to not existing file", *initResourcesDefinitionPath)
	}
	lenovoLoader, err := configuration.NewLenovoInitResourcesDefinitionLoader(*lenovoResourcesDefinitionPath)
	if err != nil {
		return nil, fmt.Errorf("given path('%s') points to not existing file", *lenovoResourcesDefinitionPath)
	}
	return []configuration.InitResourcesProvider{loader, lenovoLoader}, nil
}
