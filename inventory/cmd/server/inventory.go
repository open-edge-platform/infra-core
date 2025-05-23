// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Inventory (server)

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/golang/mock/mockgen/model" // Needed to register the mockgen model.

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/migrate/migrations"
	_ "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/runtime"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/server"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/utils/migrate"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/flags"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/oam"
	_ "github.com/open-edge-platform/infra-core/inventory/v2/pkg/perf" // Adds support for pprof.
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy/rbac"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

var (
	zlog          = logging.GetLogger("InventoryMain")
	servaddr      = flag.String(flags.ServerAddress, "0.0.0.0:50051", flags.ServerAddressDescription)
	migrationsDir = flag.String(
		migrations.MigrationsDir,
		"/usr/share/migrations/",
		migrations.MigrationsDirDescription+
			"Try './internal/ent/migrate/migrations' when running locally.",
	)
	oamservaddr    = flag.String(oam.OamServerAddress, "", oam.OamServerAddressDescription)
	enableTracing  = flag.Bool(tracing.EnableTracing, false, tracing.EnableTracingDescription)
	traceURL       = flag.String(tracing.TraceURL, "", tracing.TraceURLDescription)
	policyBundle   = flag.String(policy.PolicyBundlePath, "/rego/policy_bundle.tar.gz", policy.PolicyBundlePathDescription)
	insecureGrpc   = flag.Bool(client.InsecureGrpc, true, client.InsecureGrpcDescription)
	caCertPath     = flag.String(client.CaCertPath, "", client.CaCertPathDescription)
	tlsCertPath    = flag.String(client.TLSCertPath, "", client.TLSCertPathDescription)
	tlsKeyPath     = flag.String(client.TLSKeyPath, "", client.TLSKeyPathDescription)
	enableAuth     = flag.Bool(rbac.EnableAuth, false, rbac.EnableAuthDescription)
	enableMetrics  = flag.Bool(metrics.EnableMetrics, false, metrics.EnableMetricsDescription)
	metricsAddress = flag.String(metrics.MetricsAddress, metrics.MetricsAddressDefault, metrics.MetricsAddressDescription)
	enableAuditing = flag.Bool(flags.EnableAuditing, false, flags.EnableAuditingDescription)
)

var (
	RepoURL   = "https://github.com/open-edge-platform/infra-core/inventory.git"
	Version   = "<unset>"
	Revision  = "<unset>"
	BuildDate = "<unset>"
)

type Config struct {
	EnableTracing bool
	TraceURL      string
}

func (c Config) String() string {
	str := "{"
	str += fmt.Sprintf("EnableTracing: %t, ", c.EnableTracing)
	str += fmt.Sprintf("TraceURL: %s", c.TraceURL)
	str += "}"
	return str
}

func printSummary() {
	zlog.Info().Msgf("Starting Inventory")
	zlog.InfraSec().Info().Msgf("RepoURL: %s, Version: %s, Revision: %s, BuildDate: %s\n", RepoURL, Version, Revision, BuildDate)
}

//nolint:cyclop // complexity is 11
func main() {
	// Print a summary of the build
	printSummary()

	flag.Parse()

	cfg := Config{
		EnableTracing: *enableTracing,
		TraceURL:      *traceURL,
	}

	zlog.Info().Msgf("Cfg: %s", cfg)

	if *enableTracing {
		cleanup, err := tracing.NewTraceExporterHTTP(*traceURL, "inventory", nil)
		if err != nil {
			zlog.InfraErr(err).Msg("Error creating trace exporter")
		}
		if cleanup != nil {
			defer func() {
				err := cleanup(context.Background())
				if err != nil {
					zlog.InfraErr(err).Msg("Error in tracing cleanup")
				}
			}()
			zlog.Info().Msg("Tracing enabled")
		} else {
			zlog.Info().Msg("Tracing disabled")
		}
	}

	// Fetch the DB config.
	envPrimary, envReadOnly, err := util.LookupDBEnv()
	if err != nil {
		zlog.Fatal().Msgf("Could not fetch DB config from env")
	}
	// DB connection string for Atlas.
	atlasDBURLWriter := util.GetDBURL(envPrimary)
	var atlasDBURLReader string
	if envReadOnly != nil {
		atlasDBURLReader = util.GetDBURL(envReadOnly)
	}
	if out, migrateErr := migrate.RunAtlasMigrations(atlasDBURLWriter, *migrationsDir); migrateErr != nil {
		zlog.Fatal().Err(migrateErr).Msgf("Database migration failed. Aborting. Atlas output: %v", string(out))
	}

	// channels to handle termination and capture signals
	termChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		close(termChan)
	}()

	// waitgroup so main will wait for servers to exit cleanly
	wg := sync.WaitGroup{}

	// channel to signal the readiness
	var readyChan chan bool

	if *oamservaddr != "" {
		// Add oam grpc server
		wg.Add(1)

		readyChan = make(chan bool)
		go func() {
			if oamErr := oam.StartOamGrpcServer(termChan, readyChan, &wg, *oamservaddr, cfg.EnableTracing); oamErr != nil {
				zlog.InfraSec().Fatal().Err(oamErr).Msg("Cannot start Inventory OAM gRPC server")
			}
		}()
	}

	// Populate tenant values when the migration envs are set
	if envMig, exists := migrate.LookupMigrationEnv(); exists {
		err = migrate.PopulateTenantValues(atlasDBURLWriter, atlasDBURLReader, envMig.ProjectID) // tenantID equals to projectID
		if err != nil {
			zlog.InfraSec().Fatal().Err(err).Msgf("Error updating tenant_id in database")
		}
	}

	// Add inventory grpc server
	wg.Add(1)

	go func() {
		lis, err := net.Listen("tcp", *servaddr)
		if err != nil {
			zlog.InfraSec().Fatal().Err(err).Msgf("Error listening with TCP on address %s", *servaddr)
		}
		server.StartInventoryGrpcServer(
			termChan, readyChan, &wg, lis, atlasDBURLWriter, atlasDBURLReader, *policyBundle, getOpts())
	}()

	// wait until servers terminate
	wg.Wait()
}

func getOpts() server.Options {
	return server.Options{
		EnableTracing:  *enableTracing,
		EnableAuth:     *enableAuth,
		InsecureGrpc:   *insecureGrpc,
		EnableMetrics:  *enableMetrics,
		MetricsAddress: *metricsAddress,
		CaCertPath:     *caCertPath,
		TLSCertPath:    *tlsCertPath,
		TLSKeyPath:     *tlsKeyPath,
		EnableAuditing: *enableAuditing,
	}
}
