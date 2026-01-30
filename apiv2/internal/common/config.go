// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"flag"
	"time"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/flags"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/oam"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy/rbac"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
)

const (
	DefaultMaxJobs                      = 100
	DefaultMaxWorkers                   = 20
	DefaultTimeout                      = 10
	BaseRESTURL                         = "baseRESTURL"
	BaseRESTURLDescription              = "The REST server base URL"
	RestTimeout                         = "restTimeout"
	RestTimeoutDescription              = "Timeout for a REST API call (in seconds)"
	EnableRateLimiter                   = "enableRateLimiter"
	EnableRateLimiterDescription        = "Flag to enable rate limiter for the Echo REST server (default: true)"
	AllowedCorsOrigins                  = "allowedCorsOrigins"
	AllowedCorsOriginsDescription       = "Comma separated list of allowed CORS origins"
	EchoDebug                           = "echoDebug"
	EchoDebugDescription                = "Flag to enable debug mode in Echo REST API server (i.e., log every request to server)"
	MaxWorkers                          = "maxWorkers"
	MaxWorkersDescription               = "The maximum amount of workers running in parallel to attend job requests"
	MaxJobs                             = "maxJobs"
	MaxJobsDescription                  = "The maximum amount of jobs allowed to be queued to attend job requests"
	WsMaxConnections                    = "wsMaxConnections"
	WsMaxConnectionsDefault             = 10
	WstDefaultMaxConnectionsDescription = "The maximum number of concurrent websocket connections"
	EnableAuditing                      = "enableAuditing"
	EnableAuditingDescription           = "Flag to enable audit logs for REST API calls."
	DefaultScenario                     = "fulleim"
)

type Traces struct {
	EnableTracing bool
	TraceURL      string
}

type RestServer struct {
	Address           string
	BaseURL           string
	Timeout           time.Duration
	Cors              string
	EchoDebug         bool
	Authentication    bool
	EnableRateLimiter bool
	EnableMetrics     bool
	MetricsAddress    string
	OamServerAddr     string
}

type Worker struct {
	// the number of concurrent worker.Worker
	MaxWorkers int
	// the number of worker.Job that can sit in the queue for a specific worker
	MaxJobs int
}

type Southbound struct {
	Address       string
	CAPath        string
	KeyPath       string
	CertPath      string
	Retry         bool
	EnableMetrics bool
}

type GlobalConfig struct {
	Traces         Traces
	RestServer     RestServer
	Worker         Worker
	GRPCEndpoint   string
	GRPCAddress    string
	Inventory      Southbound
	Websocket      Websocket
	EnableAuditing bool
	EIMScenario    string
}

type Websocket struct {
	MaxConnections uint
}

func DefaultConfig() *GlobalConfig {
	return &GlobalConfig{
		Traces: Traces{
			EnableTracing: false,
			TraceURL:      "",
		},
		RestServer: RestServer{
			Address:           "0.0.0.0:8080",
			BaseURL:           "/edge-infra.orchestrator.apis/v2",
			Timeout:           DefaultTimeout * time.Second,
			Cors:              "",
			EchoDebug:         false,
			Authentication:    false,
			EnableRateLimiter: true,
			EnableMetrics:     false,
			MetricsAddress:    metrics.MetricsAddressDefault,
			OamServerAddr:     "",
		},
		Worker: Worker{
			MaxWorkers: DefaultMaxWorkers,
			MaxJobs:    DefaultMaxJobs,
		},
		Inventory: Southbound{
			Address:  "infra-inventory:50051",
			CAPath:   "",
			KeyPath:  "",
			CertPath: "",
			Retry:    true,
		},
		Websocket: Websocket{
			MaxConnections: WsMaxConnectionsDefault,
		},
		EnableAuditing: true,
		GRPCAddress:    "0.0.0.0:8090",
		GRPCEndpoint:   "localhost:8090",
		EIMScenario:    DefaultScenario,
	}
}

func Config() (*GlobalConfig, error) {
	defaultCfg := DefaultConfig()

	serverAddress := flag.String(flags.ServerAddress, defaultCfg.RestServer.Address, flags.ServerAddressDescription)
	baseURL := flag.String(BaseRESTURL, defaultCfg.RestServer.BaseURL, BaseRESTURLDescription)
	restTimeout := flag.Duration(RestTimeout, defaultCfg.RestServer.Timeout, RestTimeoutDescription)
	enableRateLimiter := flag.Bool(
		EnableRateLimiter,
		defaultCfg.RestServer.EnableRateLimiter,
		EnableRateLimiterDescription,
	)
	maxWorkers := flag.Int(MaxWorkers, defaultCfg.Worker.MaxWorkers, MaxWorkersDescription)
	maxJobs := flag.Int(MaxJobs, defaultCfg.Worker.MaxJobs, MaxJobsDescription)
	inventoryAddress := flag.String(
		client.InventoryAddress,
		defaultCfg.Inventory.Address,
		client.InventoryAddressDescription,
	)
	caPath := flag.String(client.CaCertPath, "", client.CaCertPathDescription)
	keyPath := flag.String(client.TLSKeyPath, "", client.TLSKeyPathDescription)
	certPath := flag.String(client.TLSCertPath, "", client.TLSCertPathDescription)
	enableTracing := flag.Bool(tracing.EnableTracing, defaultCfg.Traces.EnableTracing, tracing.EnableTracingDescription)
	traceURL := flag.String(tracing.TraceURL, defaultCfg.Traces.TraceURL, tracing.TraceURLDescription)
	cors := flag.String(AllowedCorsOrigins, defaultCfg.RestServer.Cors, AllowedCorsOriginsDescription)
	echoDebug := flag.Bool(EchoDebug, defaultCfg.RestServer.EchoDebug, EchoDebugDescription)
	enableAuth := flag.Bool(rbac.EnableAuth, defaultCfg.RestServer.Authentication, rbac.EnableAuthDescription)
	oamservaddr := flag.String(oam.OamServerAddress, "", oam.OamServerAddressDescription)
	wsMaxConnections := flag.Uint(WsMaxConnections, defaultCfg.Websocket.MaxConnections, WstDefaultMaxConnectionsDescription)
	defaultMetricsPort := flag.String(
		metrics.MetricsAddress, defaultCfg.RestServer.MetricsAddress, metrics.MetricsAddressDescription)
	enableMetrics := flag.Bool(
		metrics.EnableMetrics, defaultCfg.RestServer.EnableMetrics, metrics.EnableMetricsDescription)
	enableAuditing := flag.Bool(EnableAuditing, defaultCfg.EnableAuditing, EnableAuditingDescription)
	gRPCEndpoint := flag.String("grpcEndpoint", defaultCfg.GRPCEndpoint, "The endpoint of the gRPC server")
	gRPCAddress := flag.String("grpcAddress", defaultCfg.GRPCEndpoint, "The gRPC server address")
	eimScenario := flag.String("eimScenario", defaultCfg.EIMScenario, "The EIM deployment scenario name (e.g., 'fulleim', 'vpro')")
	flag.Parse()

	return &GlobalConfig{
		Traces: Traces{
			EnableTracing: *enableTracing,
			TraceURL:      *traceURL,
		},
		RestServer: RestServer{
			Address:           *serverAddress,
			BaseURL:           *baseURL,
			Timeout:           *restTimeout,
			Cors:              *cors,
			EchoDebug:         *echoDebug,
			Authentication:    *enableAuth,
			EnableRateLimiter: *enableRateLimiter,
			OamServerAddr:     *oamservaddr,
			EnableMetrics:     *enableMetrics,
			MetricsAddress:    *defaultMetricsPort,
		},
		Worker: Worker{
			MaxWorkers: *maxWorkers,
			MaxJobs:    *maxJobs,
		},
		Inventory: Southbound{
			Address:       *inventoryAddress,
			CAPath:        *caPath,
			KeyPath:       *keyPath,
			CertPath:      *certPath,
			Retry:         defaultCfg.Inventory.Retry,
			EnableMetrics: *enableMetrics,
		},
		Websocket: Websocket{
			MaxConnections: *wsMaxConnections,
		},
		EnableAuditing: *enableAuditing,
		GRPCEndpoint:   *gRPCEndpoint,
		GRPCAddress:    *gRPCAddress,
		EIMScenario:    *eimScenario,
	}, nil
}
