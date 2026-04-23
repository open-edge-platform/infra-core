// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tc_testing

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/invclient"
)

const (
	clientName = "TestRMInventoryClient"
	loggerName = "TestLogger"
)

var (
	SupportedEvents = []inv_v1.ResourceKind{
		inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD,
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT,
	}

	log = logging.GetLogger(loggerName)
)

func CreateInvClient(tb testing.TB) *invclient.TCInventoryClient {
	tb.Helper()
	var err error
	err = inv_testing.CreateClient(clientName, inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER, SupportedEvents, "")
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create Inventory Testing client")
	}

	client, err := invclient.NewTCInventoryClient(
		inv_testing.TestClients[clientName].GetTenantAwareInventoryClient(),
		inv_testing.TestClientsEvents[clientName],
		make(chan bool))
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create ResourceManager client")
	}

	tb.Cleanup(func() {
		client.Close()
		delete(inv_testing.TestClients, clientName)
		delete(inv_testing.TestClientsEvents, clientName)
	})

	return client
}

var startTestEnv = sync.OnceFunc(func() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	projectRoot := filepath.Dir(filepath.Dir(wd))
	policyPath := projectRoot + "/out"
	migrationsDir := projectRoot + "/out"
	inv_testing.StartTestingEnvironment(policyPath, "", migrationsDir)
})

var stopTestEnv = sync.OnceFunc(func() {
	wg.Wait()
	inv_testing.StopTestingEnvironment()
})

var wg = sync.WaitGroup{}

func InitTestEnvironment() func(m *testing.M) {
	return func(m *testing.M) {
		wg.Add(1)
		go stopTestEnv()
		defer wg.Done()
		startTestEnv()
		run := m.Run()
		wg.Done()
		os.Exit(run)
	}
}
