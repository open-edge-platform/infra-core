// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invclient

import (
	"context"
	"sync"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var (
	clientName = "TenantControllerInvClient"
	log        = logging.GetLogger(clientName)

	SupportedEventKinds = []inv_v1.ResourceKind{
		inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD,
	}

	// eventsWatcherBufSize is the buffer size for the events channel.
	eventsWatcherBufSize = 10
)

type TCInventoryClient struct {
	client.TenantAwareInventoryClient
	Watcher  chan *client.WatchEvents
	termChan chan bool
}

type Options struct {
	InventoryAddress string
	EnableTracing    bool
	EnableMetrics    bool
}

type Option func(*Options)

// WithInventoryAddress sets the Inventory Address.
func WithInventoryAddress(invAddr string) Option {
	return func(options *Options) {
		options.InventoryAddress = invAddr
	}
}

// WithEnableTracing enables tracing.
func WithEnableTracing(enableTracing bool) Option {
	return func(options *Options) {
		options.EnableTracing = enableTracing
	}
}

// WithEnableMetrics enables client-side gRPC metrics.
func WithEnableMetrics(enableMetrics bool) Option {
	return func(options *Options) {
		options.EnableMetrics = enableMetrics
	}
}

// Creates an inventory client config, then returns client.
func NewInventoryClientWithOptions(
	readyChan chan bool,
	termChan chan bool,
	wg *sync.WaitGroup,
	sc *client.SecurityConfig,
	opts ...Option,
) (*TCInventoryClient, error) {
	ctx := context.Background()
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	eventsWatcher := make(chan *client.WatchEvents, eventsWatcherBufSize)

	cfg := client.InventoryClientConfig{
		Name:                      clientName,
		Address:                   options.InventoryAddress,
		AbortOnUnknownClientError: true,
		SecurityCfg:               sc,
		Events:                    eventsWatcher,
		ClientKind:                inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER,
		ResourceKinds:             SupportedEventKinds,
		Wg:                        wg,
		EnableTracing:             options.EnableTracing,
		EnableMetrics:             options.EnableMetrics,
	}

	invClient, err := client.NewTenantAwareInventoryClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	log.InfraSec().Info().Msgf("Inventory client started")

	rmInvClient, err := NewTCInventoryClient(invClient, eventsWatcher, termChan)

	readyChan <- true // tell OAM that Inventory Client is ready

	return rmInvClient, err
}

func NewTCInventoryClient(
	invClient client.TenantAwareInventoryClient,
	watcher chan *client.WatchEvents,
	termChan chan bool,
) (*TCInventoryClient, error) {
	cli := &TCInventoryClient{
		TenantAwareInventoryClient: invClient,
		Watcher:                    watcher,
		termChan:                   termChan,
	}
	return cli, nil
}
