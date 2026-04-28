// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	grpc_prom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
)

func TestMain(m *testing.M) {
	// Only needed to suppress the error
	flag.String(
		"policyBundle",
		"/rego/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	flag.Parse()

	run := m.Run() // run all tests
	os.Exit(run)
}

func TestParseOptions(t *testing.T) {
	t.Run("OnlyEndpoint", func(t *testing.T) {
		opts := metrics.ParseOptions(metrics.WithEndpoint("testEndpoint"))
		assert.Equal(t, "testEndpoint", opts.Endpoint)
		assert.Equal(t, metrics.MetricsAddressDefault, opts.ListenAddress)
	})

	t.Run("OnlyAddress", func(t *testing.T) {
		opts := metrics.ParseOptions(metrics.WithListenAddress("testListenAddress"))
		assert.Equal(t, metrics.DefaultEndpoint, opts.Endpoint)
		assert.Equal(t, "testListenAddress", opts.ListenAddress)
	})

	t.Run("BothEndpointAndAddress", func(t *testing.T) {
		opts := metrics.ParseOptions(metrics.WithEndpoint("testEndpoint"), metrics.WithListenAddress("testListenAddress"))
		assert.Equal(t, "testEndpoint", opts.Endpoint)
		assert.Equal(t, "testListenAddress", opts.ListenAddress)
	})
}

func TestStartMetricsExporter(t *testing.T) {
	srvMetrics := grpc_prom.NewServerMetrics()
	go metrics.StartMetricsExporter([]prometheus.Collector{srvMetrics})
	// Wait for the server to start
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/metrics", http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetServerMetricsWithLatency(t *testing.T) {
	srvMetrics := metrics.GetServerMetricsWithLatency()
	assert.NotNil(t, srvMetrics)
}
