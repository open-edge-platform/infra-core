// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"flag"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	sites "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// To be passed to CreateClient function.
var certPath string

// Used in the streaming channel.
const eventsBufSize = 128

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Currently unused
	flag.String(
		"policyBundle",
		wd+"/../../out/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	flag.Parse()
	projectRoot := filepath.Dir(filepath.Dir(wd))

	policyPath := projectRoot + "/out"
	certPath = projectRoot + "/cert/certificates"
	migrationsDir := projectRoot + "/out"

	inv_testing.StartTestingEnvironment(policyPath, certPath, migrationsDir)
	run := m.Run() // run all tests
	inv_testing.StopTestingEnvironment()

	os.Exit(run)
}

func TestNewInventoryClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	newConfig := func() client.InventoryClientConfig {
		events := make(chan *client.WatchEvents, eventsBufSize)
		wg := &sync.WaitGroup{}
		clientCfg := client.InventoryClientConfig{
			Name:    "test_client",
			Address: "bufconn",
			DialOptions: []grpc.DialOption{
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return inv_testing.BufconnLis.Dial() }),
			},
			SecurityCfg: &client.SecurityConfig{
				Insecure: false,
				CaPath:   certPath + "/ca-cert.pem",
				CertPath: certPath + "/client-cert.pem",
				KeyPath:  certPath + "/client-key.pem",
			},
			Events:        events,
			ClientKind:    inv_v1.ClientKind_CLIENT_KIND_API,
			ResourceKinds: nil,
			Wg:            wg,
			EnableTracing: true,
		}
		return clientCfg
	}

	createClient := func(kind inv_v1.ClientKind) error {
		cfg := newConfig()
		cfg.ClientKind = kind
		cli, err := client.NewInventoryClient(ctx, cfg)
		if err != nil {
			return err
		}
		err = cli.Close()
		if err != nil {
			return err
		}
		err = cli.Close() // Close should be idempotent.
		if err != nil {
			return err
		}
		return nil
	}

	t.Run("API", func(t *testing.T) {
		require.NoError(t, createClient(inv_v1.ClientKind_CLIENT_KIND_API))
	})
	t.Run("RM", func(t *testing.T) {
		require.NoError(t, createClient(inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER))
	})
	t.Run("unspecified", func(t *testing.T) {
		err := createClient(inv_v1.ClientKind_CLIENT_KIND_UNSPECIFIED)
		require.Error(t, err)
		// Do the other assertions
		s := status.Convert(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})
	t.Run("missingContextFail", func(t *testing.T) {
		_, err := client.NewInventoryClient(nil, newConfig()) //nolint:staticcheck // passing nil for testing
		require.Error(t, err)
		s := status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "context")
	})
	t.Run("missingWaitgroupFail", func(t *testing.T) {
		cfg := newConfig()
		cfg.Wg = nil
		_, err := client.NewInventoryClient(ctx, cfg)
		require.Error(t, err)
		s := status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "waitgroup")
	})
	t.Run("missingEventsFail", func(t *testing.T) {
		cfg := newConfig()
		cfg.Events = nil
		_, err := client.NewInventoryClient(ctx, cfg)
		require.Error(t, err)
		s := status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "events")
	})
	t.Run("conflictingCfgRetryAndAbortFail", func(t *testing.T) {
		cfg := newConfig()
		cfg.EnableRegisterRetry = true
		cfg.AbortOnUnknownClientError = true
		_, err := client.NewInventoryClient(ctx, cfg)
		require.Error(t, err)
		s := status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "Both EnableRegisterRetry and AbortOnUnknownClientError")
	})
}

func TestCreate(t *testing.T) {
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	resp, err := apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")
	resID := inv_testing.GetResourceIDOrFail(t, resp)

	// get after create
	_, err = apiClient.Get(ctx, resID)
	require.NoError(t, err, "GetRegion() failed")
}

func TestUpdate(t *testing.T) {
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	resp, err := apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")
	resID := inv_testing.GetResourceIDOrFail(t, resp)

	res = &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				ResourceId: resID,
				Name:       "Test Region 2",
				Metadata:   `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}

	// update after create
	_, err = apiClient.Update(ctx, resID, &fieldmaskpb.FieldMask{}, res)
	require.NoError(t, err, "UpdateRegion() failed")

	// get after update
	_, err = apiClient.Get(ctx, resID)
	require.NoError(t, err, "GetRegion() failed")
}

func TestDelete(t *testing.T) {
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	resp, err := apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")
	resID := inv_testing.GetResourceIDOrFail(t, resp)

	// delete after create
	_, err = apiClient.Delete(ctx, resID)
	require.NoError(t, err, "DeleteRegion() failed")

	// get after delete
	_, err = apiClient.Get(ctx, resID)
	require.Error(t, err, "GetRegion() should have failed")
}

func TestFind(t *testing.T) {
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	_, err := apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")
	_, err = apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{},
		},
	}
	_, err = apiClient.Find(ctx, filter)
	require.NoError(t, err, "FindRegion() failed")
}

func TestFindAll(t *testing.T) {
	nRegions := client.BatchSize + client.BatchSize/2
	for i := 0; i < nRegions; i++ {
		inv_testing.CreateSite(t, nil, nil)
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Site{}},
		OrderBy:  sites.FieldResourceID,
	}
	res, err := apiClient.FindAll(ctx, filter)
	require.NoError(t, err, "FindAllRegions() failed")
	assert.Equal(t, nRegions, len(res))
}

func TestList(t *testing.T) {
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	_, err := apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")
	_, err = apiClient.Create(ctx, res)
	require.NoError(t, err, "CreateRegion() failed")

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{},
		},
	}
	_, err = apiClient.List(ctx, filter)
	require.NoError(t, err, "ListRegion() failed")
}

func TestListAll(t *testing.T) {
	nRegions := client.BatchSize + client.BatchSize/2
	for i := 0; i < nRegions; i++ {
		inv_testing.CreateSite(t, nil, nil)
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Site{}},
		OrderBy:  sites.FieldResourceID,
	}
	res, err := apiClient.ListAll(ctx, filter)
	require.NoError(t, err, "ListAllRegions() failed")
	assert.Equal(t, nRegions, len(res))
}

func TestInventoryClient_Subscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	apiClient := inv_testing.TestClients[inv_testing.APIClient]
	rmClient := inv_testing.TestClients[inv_testing.RMClient]
	// We need to use RM event channel because a client never sees its own events.
	rmEvents := inv_testing.TestClientsEvents[inv_testing.RMClient]
	// Drain the channel from previous events.
	for len(rmEvents) > 0 {
		<-rmEvents
	}
	require.NoError(t, rmClient.UpdateSubscriptions(ctx, []inv_v1.ResourceKind{inv_v1.ResourceKind_RESOURCE_KIND_OS}))

	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Os{
			Os: &osv1.OperatingSystemResource{
				Name:              "Test OS resource",
				Sha256:            inv_testing.GenerateRandomSha256(),
				ImageUrl:          "test",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				OsType:            osv1.OsType_OS_TYPE_MUTABLE,
				OsProvider:        osv1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
		},
	}

	// 1. Create an OS resource.
	resp, err := apiClient.Create(ctx, res)
	require.NoError(t, err)
	resID := inv_testing.GetResourceIDOrFail(t, resp)
	res.GetOs().ResourceId = resID
	select {
	case ev, ok := <-rmEvents:
		require.True(t, ok, "resource manager did not receive event")
		validateErr := validator.ValidateMessage(ev.Event)
		require.NoError(t, validateErr)
		var kind inv_v1.ResourceKind
		kind, err = util.GetResourceKindFromResourceID(ev.Event.ResourceId)
		require.NoError(t, err, "resource manager did receive a strange event")
		assert.Equal(t, inv_v1.ResourceKind_RESOURCE_KIND_OS, kind)
		assert.Equal(t, inv_v1.SubscribeEventsResponse_EVENT_KIND_CREATED, ev.Event.EventKind)
		assert.Equal(t, resID, ev.Event.ResourceId)
	case <-ctx.Done():
		t.Fatalf("resource manager did not receive event")
	}

	// 2. Unsubscribe from all kinds and trigger OS resource update.
	err = rmClient.UpdateSubscriptions(ctx, []inv_v1.ResourceKind{})
	require.NoError(t, err)
	_, err = apiClient.Update(ctx, resID, &fieldmaskpb.FieldMask{}, res)
	require.NoError(t, err)
	select {
	case ev, ok := <-rmEvents:
		require.Fail(t, "resource manager received an event: %v, ok: %v", ev, ok)
	case <-time.After(time.Second):
		// no event received, pass
	}

	// 3. Add OS to subscriptions and delete OS resource.
	err = rmClient.UpdateSubscriptions(ctx, []inv_v1.ResourceKind{inv_v1.ResourceKind_RESOURCE_KIND_OS})
	require.NoError(t, err)
	_, err = apiClient.Delete(ctx, resID)
	require.NoError(t, err)
	select {
	case ev, ok := <-rmEvents:
		require.True(t, ok, "resource manager did not receive event")
		validateErr := validator.ValidateMessage(ev.Event)
		require.NoError(t, validateErr)
		var kind inv_v1.ResourceKind
		kind, err = util.GetResourceKindFromResourceID(ev.Event.ResourceId)
		require.NoError(t, err, "resource manager did receive a strange event", ev.Event)
		assert.Equal(t, inv_v1.ResourceKind_RESOURCE_KIND_OS, kind)
		assert.Equal(t, inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED, ev.Event.EventKind)
		assert.Equal(t, resID, ev.Event.ResourceId)
		// To do proper comparison we have to reset timestamps.
		eventOsRes := ev.Event.GetResource().GetOs()
		eventOsRes.CreatedAt = ""
		eventOsRes.UpdatedAt = ""
		if eq, diff := inv_testing.ProtoEqualOrDiff(res, ev.Event.Resource); !eq {
			assert.Fail(t, "resources not equal", diff)
		}
	case <-ctx.Done():
		t.Fatalf("resource manager did not receive event")
	}
}

func TestInventoryClient_registerRetry(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var conn net.Conn
	newConfig := func() client.InventoryClientConfig {
		events := make(chan *client.WatchEvents, eventsBufSize)
		wg := &sync.WaitGroup{}
		clientCfg := client.InventoryClientConfig{
			Name:    "test_client",
			Address: "bufconn",
			DialOptions: []grpc.DialOption{
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					var err error
					conn, err = inv_testing.BufconnLis.Dial() // save the connection
					return conn, err
				}),
			},
			SecurityCfg: &client.SecurityConfig{
				Insecure: false,
				CaPath:   certPath + "/ca-cert.pem",
				CertPath: certPath + "/client-cert.pem",
				KeyPath:  certPath + "/client-key.pem",
			},
			Events:                    events,
			ClientKind:                inv_v1.ClientKind_CLIENT_KIND_API,
			ResourceKinds:             nil,
			Wg:                        wg,
			EnableTracing:             true,
			AbortOnUnknownClientError: false,
		}
		return clientCfg
	}

	t.Run("GetWithRetrySuccess", func(t *testing.T) {
		cfg := newConfig()
		cfg.EnableRegisterRetry = true
		cli, err := client.NewInventoryClient(ctx, cfg)
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond) // Need to wait for the event handler loop to pick up the closed conn.

		resp, err := cli.Get(ctx, host.GetResourceId())
		require.NoError(t, err)
		require.Equal(t, host.GetResourceId(), resp.GetResource().GetHost().GetResourceId())

		err = cli.Close()
		require.NoError(t, err)
	})
	t.Run("GetWithoutRetryFail", func(t *testing.T) {
		cfg := newConfig()
		cfg.EnableRegisterRetry = false
		cli, err := client.NewInventoryClient(ctx, cfg)
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)
		time.Sleep(200 * time.Millisecond) // Need to wait for the event handler loop to pick up the closed conn.

		_, err = cli.Get(ctx, host.GetResourceId())
		require.Error(t, err)
	})
}

func TestGet(t *testing.T) {
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:     "Test Host 1",
				Uuid:     uuid.NewString(),
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheClient]

	// create resource.
	rsp, err := clientCache.Create(ctx, res)
	require.NoError(t, err, "CreateHost() failed")

	resID := inv_testing.GetResourceIDOrFail(t, rsp)

	// step 1: Cache shoudln't have resource before GET op.
	res1, err1 := clientCache.TestGetClientCache().GetResourceByID(client.FakeTenantID, resID)
	require.NotNil(t, err1, "client cache shouldn't have resource")
	require.Nil(t, res1, "client cache shouldn't have resource")

	// step 2: try normal Get, this shall cache resource.
	_, err = clientCache.Get(ctx, resID)
	require.NoError(t, err, "GetHost() failed")

	// step 3: cache should have resource now.
	res3, err3 := clientCache.TestGetClientCache().GetResourceByID(client.FakeTenantID, resID)
	require.Nil(t, err3, "client cache should have resource")
	require.Equal(t, resID, res3.Resource.GetHost().ResourceId, "client cache should have resource")

	// step 4: delete resource, this should delete resource from catch too.
	_, err4 := clientCache.Delete(ctx, resID)
	require.NoError(t, err4, "DeleteHost() failed")

	// step 5: cache shouldn't have resource now
	res5, err5 := clientCache.TestGetClientCache().GetResourceByID(client.FakeTenantID, resID)
	require.NotNil(t, err5, "client cache shouldn't have resource")
	require.Nil(t, res5, "client cache shouldn't have resource")
}

func TestCacheList(t *testing.T) {
	hostUUID := uuid.NewString()
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:     "Test Host 1",
				Uuid:     hostUUID,
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheClient]

	// create
	rsp, err := clientCache.Create(ctx, res)
	require.NoError(t, err, "CreateHost() failed")

	resID := inv_testing.GetResourceIDOrFail(t, rsp)

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{},
		},
		Filter: `uuid = "` + hostUUID + `"`,
	}

	// step 1: verify that resource doesn't exist in cache.
	res1, err1 := clientCache.TestGetClientCache().GetResourceByFilter(filter)
	require.NotNil(t, err1, "resource shoudn't exist in cache")
	require.Nil(t, res1, "resource shouldn't exist in cache")

	// step 2: List the resource, this will add resource in cache.
	_, err = clientCache.List(ctx, filter)
	require.NoError(t, err, "ListHost() failed")

	// step 3: verify that resource exist in cache.
	res3, err3 := clientCache.TestGetClientCache().GetResourceByFilter(filter)
	require.Nil(t, err3, "resource should exist in cache")
	require.NotNil(t, res3, "resource should exist in cache")

	// step 4: delete resource, this should delete resource from catch too.
	_, err4 := clientCache.Delete(ctx, resID)
	require.NoError(t, err4, "DeleteHost() failed")

	// step 5: cache shouldn't have resource now
	res5, err5 := clientCache.TestGetClientCache().GetResourceByFilter(filter)
	require.NotNil(t, err5, "client cache shouldn't have resource")
	require.Nil(t, res5, "client cache shouldn't have resource")
}

func TestCacheListTimeout(t *testing.T) {
	hostUUID := uuid.NewString()
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:     "Test Host 1",
				Uuid:     hostUUID,
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheClient]

	// create
	_, err := clientCache.Create(ctx, res)
	require.NoError(t, err, "CreateHost() failed")

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{},
		},
		Filter: `uuid = "` + hostUUID + `"`,
	}

	staleTime := 2 * time.Second

	// step 1: change cache config with stale time to 2 sec.
	clientCache.TestGetClientCache().UpdateStaleTime(staleTime)

	// step 2: List the resource, this will add resource in cache.
	_, err = clientCache.List(ctx, filter)
	require.NoError(t, err, "ListHost() failed")

	// step 3: verify that resource exist in cache.
	res3, err3 := clientCache.TestGetClientCache().GetResourceByFilter(filter)
	require.Nil(t, err3, "resource should exist in cache")
	require.NotNil(t, res3, "resource should exist in cache")

	// step 4: delay checking cache for "staleTime"
	time.Sleep(staleTime)

	// step 5: cache shouldn't have resource now
	res5, err5 := clientCache.TestGetClientCache().GetResourceByFilter(filter)
	require.NotNil(t, err5, "client cache shouldn't have resource")
	require.Nil(t, res5, "client cache shouldn't have resource")
}

func TestGetHostByUUID(t *testing.T) {
	// Host without subres
	hostUUID := uuid.NewString()
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:     "Test Host 1",
				Uuid:     hostUUID,
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]

	// create resource.
	rsp, err := clientCache.Create(ctx, res)
	require.NoError(t, err, "CreateHost() failed")

	resID := inv_testing.GetResourceIDOrFail(t, rsp)

	// step 1: Cache shouldn't have resource before GET op.
	res1, err1 := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.NotNil(t, err1, "client cache shouldn't have resource")
	require.Nil(t, res1, "client cache shouldn't have resource")

	// step 2: try normal Get, this shall cache resource.
	_, err = clientCache.GetHostByUUID(ctx, hostUUID)
	require.NoError(t, err, "GetHost() failed")

	// step 3: cache should have resource now.
	res3, err3 := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.Nil(t, err3, "client cache should have resource")
	require.Equal(t, resID, res3.ResourceId, "client cache should have resource")

	// step 4: delete resource, this should delete resource from catch too.
	_, err4 := clientCache.Delete(ctx, resID)
	require.NoError(t, err4, "DeleteHost() failed")

	// step 5: cache shouldn't have resource now
	res5, err5 := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.NotNil(t, err5, "client cache shouldn't have resource")
	require.Nil(t, res5, "client cache shouldn't have resource")
}

func TestCacheUuidTimeout(t *testing.T) {
	hostUUID := uuid.NewString()
	res := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:     "Test Host 1",
				Uuid:     hostUUID,
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]

	// create
	_, err := clientCache.Create(ctx, res)
	require.NoError(t, err, "CreateHost() failed")

	staleTime := 2 * time.Second

	// step 1: change cache config with stale time to 2 sec.
	clientCache.TestGetClientCacheUUID().UpdateStaleTime(staleTime)

	// step 2: List the resource, this will add resource in cache.
	_, err = clientCache.GetHostByUUID(ctx, hostUUID)
	require.NoError(t, err, "GetHost() failed")

	// step 3: verify that resource exist in cache.
	res3, err3 := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.Nil(t, err3, "resource should exist in cache")
	require.NotNil(t, res3, "resource should exist in cache")

	// step 4: delay checking cache for "staleTime"
	time.Sleep(staleTime + (1 * time.Second))

	// step 5: cache shouldn't have resource now
	res5, err5 := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.NotNil(t, err5, "client cache shouldn't have resource")
	require.Nil(t, res5, "client cache shouldn't have resource")
}

func TestCacheUuidDelete(t *testing.T) {
	host := inv_testing.CreateHostNoCleanup(t, nil, nil)
	hostUUID := host.GetUuid()
	tenantID := host.GetTenantId()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]
	// Populate cache
	_, err := clientCache.GetHostByUUID(ctx, hostUUID)
	require.NoError(t, err, "GetHost() failed")
	res3, err3 := clientCache.TestGetClientCacheUUID().GetHostByUUID(tenantID, hostUUID)
	require.Nil(t, err3, "resource should exist in cache")
	require.NotNil(t, res3, "resource should exist in cache")

	inv_testing.HardDeleteHost(t, host.ResourceId)

	res5, err5 := clientCache.TestGetClientCacheUUID().GetHostByUUID(tenantID, hostUUID)
	require.NotNil(t, err5, "client cache shouldn't have resource")
	require.Nil(t, res5, "client cache shouldn't have resource")

	_, err = clientCache.GetHostByUUID(ctx, hostUUID)
	require.Error(t, err)
	s := status.Convert(err)
	require.NotNil(t, s)
	assert.Equal(t, codes.NotFound, s.Code())
}

func TestCacheUuidUpdate(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)
	osRes := inv_testing.CreateOs(t)
	instance := inv_testing.CreateInstance(t, host, osRes)
	hostGpu := inv_testing.CreatHostGPU(t, host)
	hostUsb := inv_testing.CreateHostusb(t, host)
	hostStorage := inv_testing.CreateHostStorage(t, host)
	hostDevice := inv_testing.CreateHostDevice(t, host)
	hostNic := inv_testing.CreateHostNic(t, host)
	// TODO: create Host GPU as well
	hostUUID := host.GetUuid()
	host2UUID := host2.GetUuid()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]
	// Populate cache
	loadAndAssertUUIDInCache(t, hostUUID)
	loadAndAssertUUIDInCache(t, host2UUID)

	// Reset desired states, we are updating from RM
	host.DesiredState = computev1.HostState_HOST_STATE_UNSPECIFIED
	host.DesiredPowerState = computev1.PowerState_POWER_STATE_UNSPECIFIED
	instance.DesiredState = computev1.InstanceState_INSTANCE_STATE_UNSPECIFIED

	testcases := map[string]struct {
		resID string
		upRes *inv_v1.Resource
	}{
		"Host": {
			resID: host.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: host},
			},
		},
		"Instance": {
			resID: instance.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: instance},
			},
		},
		"HostGpu": {
			resID: hostGpu.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostgpu{Hostgpu: hostGpu},
			},
		},
		"HostUsb": {
			resID: hostUsb.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostusb{Hostusb: hostUsb},
			},
		},
		"HostStorage": {
			resID: hostStorage.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hoststorage{Hoststorage: hostStorage},
			},
		},
		"HostDevice": {
			resID: hostDevice.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostdevice{Hostdevice: hostDevice},
			},
		},
		"HostNic": {
			resID: hostNic.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostnic{Hostnic: hostNic},
			},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// Populate cache
			loadAndAssertUUIDInCache(t, hostUUID)

			// Update that should invalidate the cache
			_, err := clientCache.Update(ctx, tc.resID, &fieldmaskpb.FieldMask{}, tc.upRes)
			require.NoError(t, err)

			// Cache should be empty now
			assertUUIDNotInCache(t, hostUUID)
			// Other host should still be in cache
			assertUUIDInCache(t, host2UUID)
		})
	}
}

func TestCacheUuidInvalidateViaUpdateEvent(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)
	osRes := inv_testing.CreateOs(t)
	instance := inv_testing.CreateInstance(t, host, osRes)
	hostGpu := inv_testing.CreatHostGPU(t, host)
	hostUsb := inv_testing.CreateHostusb(t, host)
	hostStorage := inv_testing.CreateHostStorage(t, host)
	hostDevice := inv_testing.CreateHostDevice(t, host)
	hostNic := inv_testing.CreateHostNic(t, host)
	hostUUID := host.GetUuid()
	host2UUID := host2.GetUuid()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Populate cache
	loadAndAssertUUIDInCache(t, hostUUID)
	loadAndAssertUUIDInCache(t, host2UUID)

	// Reset desired states, we are updating from RM
	host.DesiredState = computev1.HostState_HOST_STATE_UNSPECIFIED
	host.DesiredPowerState = computev1.PowerState_POWER_STATE_UNSPECIFIED
	instance.DesiredState = computev1.InstanceState_INSTANCE_STATE_UNSPECIFIED

	rmClient := inv_testing.GetClient(t, inv_testing.RMClient)

	testcases := map[string]struct {
		resID string
		upRes *inv_v1.Resource
	}{
		"Host": {
			resID: host.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: host},
			},
		},
		"Instance": {
			resID: instance.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: instance},
			},
		},
		"HostGpu": {
			resID: hostGpu.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostgpu{Hostgpu: hostGpu},
			},
		},
		"HostUsb": {
			resID: hostUsb.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostusb{Hostusb: hostUsb},
			},
		},
		"HostStorage": {
			resID: hostStorage.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hoststorage{Hoststorage: hostStorage},
			},
		},
		"HostDevice": {
			resID: hostDevice.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostdevice{Hostdevice: hostDevice},
			},
		},
		"HostNic": {
			resID: hostNic.GetResourceId(),
			upRes: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostnic{Hostnic: hostNic},
			},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// Populate cache
			loadAndAssertUUIDInCache(t, hostUUID)

			// Update that should invalidate the cache
			// Hardcode fieldmask
			_, tErr := rmClient.Update(ctx, tc.resID, &fieldmaskpb.FieldMask{}, tc.upRes)
			require.NoError(t, tErr)
			// Wait a little bit for the event to be propagated
			time.Sleep(10 * time.Millisecond)

			// Cache should be empty now
			assertUUIDNotInCache(t, hostUUID)
			// Other host should still be in cache
			assertUUIDInCache(t, host2UUID)
		})
	}
}

func TestCacheUuidInvalidateViaCreateEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rmClient := inv_testing.GetClient(t, inv_testing.RMClient)
	host := inv_testing.CreateHost(t, nil, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)
	hostUUID := host.Uuid
	host2UUID := host2.Uuid
	osRes := inv_testing.CreateOs(t)

	loadAndAssertUUIDInCache(t, hostUUID)
	loadAndAssertUUIDInCache(t, host2UUID)

	testcases := map[string]struct {
		createResource *inv_v1.Resource
	}{
		"HostUsb": {
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostusb{Hostusb: &computev1.HostusbResource{
					Host: host,
				}},
			},
		},
		"HostStorage": {
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hoststorage{Hoststorage: &computev1.HoststorageResource{
					Host: host,
				}},
			},
		},
		"HostDevice": {
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostdevice{Hostdevice: &computev1.HostdeviceResource{
					Host: host,
				}},
			},
		},
		"HostNic": {
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostnic{Hostnic: &computev1.HostnicResource{
					Host: host,
				}},
			},
		},
		"HostGpu": {
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostgpu{Hostgpu: &computev1.HostgpuResource{
					Host: host,
				}},
			},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			resp, err := rmClient.Create(ctx, tc.createResource)
			require.NoError(t, err)
			t.Cleanup(func() { inv_testing.DeleteResource(t, inv_testing.GetResourceIDOrFail(t, resp)) })
			time.Sleep(10 * time.Millisecond)
			assertUUIDNotInCache(t, hostUUID)
			assertUUIDInCache(t, host2UUID)
		})
	}

	t.Run("Instance", func(t *testing.T) {
		resp, err := rmClient.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: &computev1.InstanceResource{
					Kind: computev1.InstanceKind_INSTANCE_KIND_METAL,
					Host: host,
					Os:   osRes,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.HardDeleteInstance(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		time.Sleep(10 * time.Millisecond)
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})
}

func TestCacheUuidInvalidateViaDeleteEvent(t *testing.T) {
	host := inv_testing.CreateHostNoCleanup(t, nil, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)
	osRes := inv_testing.CreateOs(t)
	instance := inv_testing.CreateInstanceNoCleanup(t, host, osRes)
	hostGpu := inv_testing.CreatHostGPUNoCleanup(t, host)
	hostUsb := inv_testing.CreateHostusbNoCleanup(t, host)
	hostStorage := inv_testing.CreateHostStorageNoCleanup(t, host)
	hostDevice := inv_testing.CreateHostDeviceNoCleanup(t, host)
	hostNic := inv_testing.CreateHostNicNoCleanup(t, host)
	hostUUID := host.GetUuid()
	host2UUID := host2.GetUuid()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Populate cache
	loadAndAssertUUIDInCache(t, hostUUID)
	loadAndAssertUUIDInCache(t, host2UUID)

	// Reset desired states, we are updating from RM
	host.DesiredState = computev1.HostState_HOST_STATE_UNSPECIFIED
	host.DesiredPowerState = computev1.PowerState_POWER_STATE_UNSPECIFIED
	instance.DesiredState = computev1.InstanceState_INSTANCE_STATE_UNSPECIFIED

	rmClient := inv_testing.GetClient(t, inv_testing.RMClient)

	testcases := map[string]struct {
		resID string
	}{
		"HostGpu": {
			resID: hostGpu.GetResourceId(),
		},
		"HostUsb": {
			resID: hostUsb.GetResourceId(),
		},
		"HostStorage": {
			resID: hostStorage.GetResourceId(),
		},
		"HostDevice": {
			resID: hostDevice.GetResourceId(),
		},
		"HostNic": {
			resID: hostNic.GetResourceId(),
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// Populate cache
			loadAndAssertUUIDInCache(t, hostUUID)

			// Delete should invalidate the cache
			// Hardcode fieldmask
			_, tErr := rmClient.Delete(ctx, tc.resID)
			require.NoError(t, tErr)
			// Wait a little bit for the event to be propagated
			time.Sleep(10 * time.Millisecond)

			// Cache should be empty now
			assertUUIDNotInCache(t, hostUUID)
			// Other host should still be in cache
			assertUUIDInCache(t, host2UUID)
		})
	}
	t.Run("Instance", func(t *testing.T) {
		// Populate cache
		loadAndAssertUUIDInCache(t, hostUUID)

		// Delete should invalidate the cache
		tErr := inv_testing.HardDeleteInstanceAndReturnError(t, instance.GetResourceId())
		require.NoError(t, tErr)
		// Wait a little bit for the event to be propagated
		time.Sleep(10 * time.Millisecond)

		// Cache should be empty now
		assertUUIDNotInCache(t, hostUUID)
		// Other host should still be in cache
		assertUUIDInCache(t, host2UUID)
	})

	t.Run("Host", func(t *testing.T) {
		// Populate cache
		loadAndAssertUUIDInCache(t, hostUUID)

		// Delete should invalidate the cache
		// Hardcode fieldmask
		tErr := inv_testing.HardDeleteHostAndReturnError(t, host.GetResourceId())
		require.NoError(t, tErr)
		// Wait a little bit for the event to be propagated
		time.Sleep(10 * time.Millisecond)

		// Cache should be empty now
		assertUUIDNotInCache(t, hostUUID)
		// Other host should still be in cache
		assertUUIDInCache(t, host2UUID)
	})
}

func TestCacheUuidCreateWithWarmCache(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]
	host := inv_testing.CreateHost(t, nil, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)
	hostUUID := host.Uuid
	host2UUID := host2.Uuid

	loadAndAssertUUIDInCache(t, host2UUID)

	t.Run("Instance", func(t *testing.T) {
		loadAndAssertUUIDInCache(t, hostUUID)
		osRes := inv_testing.CreateOs(t)
		resp, err := clientCache.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: &computev1.InstanceResource{
					Kind: computev1.InstanceKind_INSTANCE_KIND_METAL,
					Os:   osRes,
					Host: host,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.HardDeleteInstance(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})

	t.Run("HostUsb", func(t *testing.T) {
		loadAndAssertUUIDInCache(t, hostUUID)
		resp, err := clientCache.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostusb{Hostusb: &computev1.HostusbResource{
					Host: host,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.DeleteResource(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})

	t.Run("HostStorage", func(t *testing.T) {
		loadAndAssertUUIDInCache(t, hostUUID)
		resp, err := clientCache.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hoststorage{Hoststorage: &computev1.HoststorageResource{
					Host: host,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.DeleteResource(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})

	t.Run("HostDevice", func(t *testing.T) {
		loadAndAssertUUIDInCache(t, hostUUID)
		resp, err := clientCache.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostdevice{Hostdevice: &computev1.HostdeviceResource{
					Host: host,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.DeleteResource(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})

	t.Run("HostNic", func(t *testing.T) {
		loadAndAssertUUIDInCache(t, hostUUID)
		resp, err := clientCache.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostnic{Hostnic: &computev1.HostnicResource{
					Host: host,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.DeleteResource(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})

	t.Run("HostGpu", func(t *testing.T) {
		loadAndAssertUUIDInCache(t, hostUUID)
		resp, err := clientCache.Create(ctx,
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Hostgpu{Hostgpu: &computev1.HostgpuResource{
					Host: host,
				}},
			})
		require.NoError(t, err)
		t.Cleanup(func() { inv_testing.DeleteResource(t, inv_testing.GetResourceIDOrFail(t, resp)) })
		assertUUIDNotInCache(t, hostUUID)
		assertUUIDInCache(t, host2UUID)
	})
}

func TestGetTreeHierarchy(t *testing.T) {
	res1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	res2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	r1, err := apiClient.Create(ctx, res1)
	require.NoError(t, err, "CreateRegion() failed")
	res2.GetRegion().ParentRegion = &location_v1.RegionResource{
		ResourceId: inv_testing.GetResourceIDOrFail(t, r1),
	}
	r2, err := apiClient.Create(ctx, res2)
	require.NoError(t, err, "CreateRegion() failed")

	treeReq := &inv_v1.GetTreeHierarchyRequest{
		Filter: []string{inv_testing.GetResourceIDOrFail(t, r2)},
	}
	resp, err := apiClient.GetTreeHierarchy(ctx, treeReq)
	require.NoError(t, err, "GetTreeHierarchy() failed")
	require.Len(t, resp, 2)
}

func TestGetSitesPerRegion(t *testing.T) {
	res1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name: "Test Region 1",
			},
		},
	}
	res2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name: "Test Region 1",
			},
		},
	}

	res3 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Site{
			Site: &location_v1.SiteResource{
				Name: "Test Site 1",
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]

	// create
	r1, err := apiClient.Create(ctx, res1)
	require.NoError(t, err, "CreateRegion() failed")
	r1ID := inv_testing.GetResourceIDOrFail(t, r1)
	res2.GetRegion().ParentRegion = &location_v1.RegionResource{
		ResourceId: r1ID,
	}
	r2, err := apiClient.Create(ctx, res2)
	require.NoError(t, err, "CreateRegion() failed")
	r2ID := inv_testing.GetResourceIDOrFail(t, r2)

	request := &inv_v1.GetSitesPerRegionRequest{
		Filter: []string{
			r1ID,
			r2ID,
		},
	}

	resp, err := apiClient.GetSitesPerRegion(ctx, request)
	require.NoError(t, err, "GetSitesPerRegion() failed")
	respRegions := resp.GetRegions()
	require.Len(t, respRegions, 2)

	respRegion0 := respRegions[0]
	require.Equal(t, 0, int(respRegion0.GetChildSites()))
	respRegion1 := respRegions[1]
	require.Equal(t, 0, int(respRegion1.GetChildSites()))

	res3.GetSite().Region = &location_v1.RegionResource{
		ResourceId: r2ID,
	}
	r3, err := apiClient.Create(ctx, res3)
	require.NoError(t, err, "CreateSite() failed")
	require.NotNil(t, r3)

	resp, err = apiClient.GetSitesPerRegion(ctx, request)
	require.NoError(t, err, "GetSitesPerRegion() failed")
	respRegions = resp.GetRegions()
	require.Len(t, respRegions, 2)

	respRegion0 = respRegions[0]
	require.Equal(t, 1, int(respRegion0.GetChildSites()))
	respRegion1 = respRegions[1]
	require.Equal(t, 1, int(respRegion1.GetChildSites()))
}

func TestMaxMessageSize(t *testing.T) {
	// TODO: extend with more lengthy fields to validate max message size is large enough.
	// This test validate that the max message size is large enough to handle 100 hosts with large size of installed_packages.
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	randString10k := generateRandomString(50000) // 50k characters, max supported from API perspective.
	createdOs := dao.CreateOsWithOpts(t, tenantID, true,
		inv_testing.InstalledPackages(randString10k),
		inv_testing.Sha256(inv_testing.RandomSha256v1),
	)
	// Create 100 Hosts with large installed_packages
	for i := 0; i < 100; i++ {
		host := dao.CreateHost(t, tenantID)
		dao.CreateInstance(t, tenantID, host, createdOs)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apiClient := inv_testing.TestClients[inv_testing.APIClient]
	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{},
		},
	}
	_, err := apiClient.List(ctx, filter)
	require.NoError(t, err, "ListHost() failed")
}

func generateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		//nolint:gosec // no need for cryptographic security here
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func assertUUIDNotInCache(t *testing.T, hostUUID string) {
	t.Helper()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]
	res, tErr := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.NotNil(t, tErr, "client cache shouldn't have resource")
	require.Nil(t, res, "client cache shouldn't have resource")
}

func assertUUIDInCache(t *testing.T, hostUUID string) {
	t.Helper()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]
	res, tErr := clientCache.TestGetClientCacheUUID().GetHostByUUID(client.FakeTenantID, hostUUID)
	require.Nil(t, tErr, "client cache should have resource")
	require.NotNil(t, res, "client cache should have resource")
}

func loadAndAssertUUIDInCache(t *testing.T, hostUUID string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clientCache := inv_testing.TestClients[inv_testing.CacheUUIDClient]
	_, err := clientCache.GetHostByUUID(ctx, hostUUID)
	require.NoError(t, err, "GetHost() failed")
	assertUUIDInCache(t, hostUUID)
}
