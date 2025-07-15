// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/ipaddressresource"
	netlinks "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/netlinkresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/workloadresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	localaccount_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	remoteaccessv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
)

const (
	RandomSha256v1    = "49dbd0fbc4332a30651435ed20b5d3b79176b14d40b0339f245ade38f1afce2f"
	RandomSha256v2    = "2d7819c0c56756caece741d1795253fb5f13cf0fd367cc860744864679a2da70"
	RandomSha256v3    = "06e396c94b166ef554a456a7cff0341f4cb2b72e56dd24e2f8db5fdb9e343b9e"
	DummyProviderName = "for unit testing purposes"
	dummyHostName     = "for unit testing purposes"
	dummyInstanceName = "for unit testing purposes"
)

// ProtoEqualOrDiff Deep-compares two protobuf messages and returns a human-readable diff if they
// are not equal. This can be used to create useful error messages in unit tests:
//
//	if eq, diff := ProtoEqualOrDiff(a, b); !eq {
//		t.Errorf("messages are not equal: %v", diff)
//	}
func ProtoEqualOrDiff(a, b proto.Message) (bool, string) {
	if !proto.Equal(a, b) {
		return false, cmp.Diff(a, b, protocmp.Transform())
	}
	return true, ""
}

// Helper function to run as pre-requisite for each helper
// using the Testing clients.
func GetClient(tb testing.TB, clientName ClientType) client.InventoryClient {
	tb.Helper()

	invClient := TestClients[clientName]
	require.NotNil(tb, invClient)

	return invClient
}

func NewInvResourceDAOOrFail(tb testing.TB) *InvResourceDAO {
	tb.Helper()
	return &InvResourceDAO{
		apiClient:        GetClient(tb, APIClient).GetTenantAwareInventoryClient(),
		apiClientWatcher: TestClientsEvents[APIClient],
		rmClient:         GetClient(tb, RMClient).GetTenantAwareInventoryClient(),
		rmClientWatcher:  TestClientsEvents[RMClient],
		tcClient:         GetClient(tb, TCClient).GetTenantAwareInventoryClient(),
		tcClientWatcher:  TestClientsEvents[TCClient],
	}
}

func NewInvResourceDAO() (*InvResourceDAO, error) {
	apiClient := TestClients[APIClient]
	if apiClient == nil {
		return nil, fmt.Errorf("APIClient is not initialized yet")
	}
	apiClientWatcher := TestClientsEvents[APIClient]
	if apiClientWatcher == nil {
		return nil, fmt.Errorf("APIClientWatcher is not initialized yet")
	}
	rmClient := TestClients[RMClient]
	if rmClient == nil {
		return nil, fmt.Errorf("RMClient is not initialized yet")
	}
	rmClientWatcher := TestClientsEvents[RMClient]
	if rmClientWatcher == nil {
		return nil, fmt.Errorf("RMClientWatcher is not initialized yet")
	}
	tcClient := TestClients[TCClient]
	if tcClient == nil {
		return nil, fmt.Errorf("TCClient is not initialized yet")
	}
	tcClientWatcher := TestClientsEvents[TCClient]
	if tcClientWatcher == nil {
		return nil, fmt.Errorf("TCClientWatcher is not initialized yet")
	}

	return &InvResourceDAO{
		apiClient:        apiClient.GetTenantAwareInventoryClient(),
		apiClientWatcher: apiClientWatcher,
		rmClient:         rmClient.GetTenantAwareInventoryClient(),
		rmClientWatcher:  rmClientWatcher,
		tcClient:         tcClient.GetTenantAwareInventoryClient(),
		tcClientWatcher:  tcClientWatcher,
	}, nil
}

// InvResourceDAO provides set of functions allowing for simple inv resource creation/deletion.
type InvResourceDAO struct {
	apiClient        client.TenantAwareInventoryClient
	rmClient         client.TenantAwareInventoryClient
	tcClient         client.TenantAwareInventoryClient
	apiClientWatcher chan *client.WatchEvents
	rmClientWatcher  chan *client.WatchEvents
	tcClientWatcher  chan *client.WatchEvents
}

func (c *InvResourceDAO) GetAPIClient() client.TenantAwareInventoryClient {
	return c.apiClient
}

func (c *InvResourceDAO) GetRMClient() client.TenantAwareInventoryClient {
	return c.rmClient
}

func (c *InvResourceDAO) GetTCClient() client.TenantAwareInventoryClient {
	return c.tcClient
}

func (c *InvResourceDAO) GetAPIClientWatcher() chan *client.WatchEvents {
	return c.apiClientWatcher
}

func (c *InvResourceDAO) GetRMClientWatcher() chan *client.WatchEvents {
	return c.rmClientWatcher
}

func (c *InvResourceDAO) GetTCClientWatcher() chan *client.WatchEvents {
	return c.tcClientWatcher
}

// The following are convenience functions to delete resources;
// they can be used at test exit with CleanUp.

func (c *InvResourceDAO) DeleteResource(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()

	err := c.DeleteResourceAndReturnError(tb, tenantID, resourceID)
	require.NoError(tb, err)
}

//nolint:revive // this is the test tool + want to keep testing.TB on first position
func (c *InvResourceDAO) DeleteAllResources(
	tb testing.TB, ctx context.Context, tenantID string, kind inv_v1.ResourceKind, enforce bool,
) error {
	tb.Helper()

	return c.GetTCClient().DeleteAllResources(ctx, tenantID, kind, enforce)
}

func (c *InvResourceDAO) DeleteResourceAndReturnError(tb testing.TB, tenantID, resourceID string) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := c.apiClient.Delete(ctx, tenantID, resourceID)
	if err != nil {
		return err
	}
	return nil
}

// HardDeleteTenant - hard deletes the given tenant via 2-phase deletion.
func (c *InvResourceDAO) HardDeleteTenant(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()
	err := c.HardDeleteTenantAndReturnError(tb, tenantID, resourceID)
	require.NoError(tb, err, "UpdateHost() failed")
}

// HardDeleteTenantAndReturnError - hard deletes the given host via 2-phase deletion.
func (c *InvResourceDAO) HardDeleteTenantAndReturnError(tb testing.TB, tenantID, resourceID string) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.DeleteResource(tb, tenantID, resourceID)
	_, err := c.tcClient.Update(
		ctx,
		tenantID,
		resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{tenantv1.TenantFieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Tenant{
				Tenant: &tenantv1.Tenant{
					CurrentState: tenantv1.TenantState_TENANT_STATE_DELETED,
				},
			},
		},
	)
	return err
}

// HardDeleteHost - hard deletes the given host via 2-phase deletion.
func (c *InvResourceDAO) HardDeleteHost(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()
	err := c.HardDeleteHostAndReturnError(tb, tenantID, resourceID)
	require.NoError(tb, err, "UpdateHost() failed")
}

// HardDeleteHostAndReturnError - hard deletes the given host via 2-phase deletion.
func (c *InvResourceDAO) HardDeleteHostAndReturnError(tb testing.TB, tenantID, resourceID string) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.DeleteResource(tb, tenantID, resourceID)
	_, err := c.rmClient.Update(
		ctx,
		tenantID,
		resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: &computev1.HostResource{
					CurrentState: computev1.HostState_HOST_STATE_DELETED,
				},
			},
		},
	)
	return err
}

// HardDeleteIPAddress - hard deletes is done without explicit delete. IPAddresses are removed without an explicit desired state.
func (c *InvResourceDAO) HardDeleteIPAddress(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()

	err := c.HardDeleteIPAddressAndReturnError(tb, tenantID, resourceID)
	require.NoError(tb, err, "UpdateIPAddress() failed")
}

func (c *InvResourceDAO) HardDeleteIPAddressAndReturnError(tb testing.TB, tenantID, resourceID string) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := c.rmClient.Update(
		ctx,
		tenantID,
		resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{ipaddressresource.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Ipaddress{
				Ipaddress: &network_v1.IPAddressResource{
					CurrentState: network_v1.IPAddressState_IP_ADDRESS_STATE_DELETED,
				},
			},
		},
	)
	return err
}

// HardDeleteInstance - hard deletes the given VM via 2-phase deletion.
func (c *InvResourceDAO) HardDeleteInstance(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()
	err := c.HardDeleteInstanceAndReturnError(tb, tenantID, resourceID)
	require.NoError(tb, err, "UpdateInstance() failed")
}

func (c *InvResourceDAO) HardDeleteInstanceAndReturnError(tb testing.TB, tenantID, resourceID string) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.DeleteResource(tb, tenantID, resourceID)
	_, err := c.rmClient.Update(
		ctx,
		tenantID,
		resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					CurrentState: computev1.InstanceState_INSTANCE_STATE_DELETED,
				},
			},
		},
	)
	return err
}

// The following are convenience functions to set up entities for testing.
// They are automatically deleted at test exit.

func (c *InvResourceDAO) CreateTenantWithOpts(
	tb testing.TB, tenantID string, doCleanup bool, opts ...Opt[tenantv1.Tenant],
) *tenantv1.Tenant {
	tb.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tenant := &tenantv1.Tenant{
		TenantId: tenantID,
	}

	collections.ForEach(opts, func(opt Opt[tenantv1.Tenant]) { opt(tenant) })

	rsp, err := c.GetTCClient().Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Tenant{Tenant: tenant},
		},
	)
	require.NoError(tb, err)

	tenantResp := rsp.GetTenant()
	if doCleanup {
		tb.Cleanup(func() { c.HardDeleteTenant(tb, tenantID, tenantResp.ResourceId) })
	}
	return tenantResp
}

// CreateHostWithOpts - creates Host with given options. Note this helper is not really meant to be used for the
// test of HostResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateHostWithOpts(
	tb testing.TB, tenantID string, doCleanup bool, opts ...Opt[computev1.HostResource],
) (host *computev1.HostResource) {
	tb.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	host = &computev1.HostResource{
		Name:         dummyHostName,
		DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
		Note:         "some note",

		HardwareKind: "XDgen2",
		Uuid:         uuid.NewString(),
		MemoryBytes:  64 * util.Gigabyte, //nolint:mnd // Teting only

		CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
		CpuSockets:      1,
		CpuCores:        14, //nolint:mnd // Teting only
		CpuCapabilities: "",
		CpuArchitecture: "x86_64",
		CpuThreads:      10, //nolint:mnd // Teting only

		MgmtIp: "192.168.10.10",

		BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
		BmcIp:       "10.0.0.10",
		BmcUsername: "user",
		BmcPassword: "pass",
		PxeMac:      "90:49:fa:ff:ff:ff",

		Hostname:     "testhost1",
		SerialNumber: "12345678",

		DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
		TenantId:          tenantID,
	}
	collections.ForEach(opts, func(o Opt[computev1.HostResource]) { o(host) })
	resp, err := c.apiClient.Create(ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{Host: host},
		})
	require.NoError(tb, err)

	hostResp := resp.GetHost()
	if doCleanup {
		tb.Cleanup(func() { c.HardDeleteHost(tb, tenantID, hostResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	hostResp.Site = nil
	hostResp.Provider = nil
	hostResp.Instance = nil

	return hostResp
}

func (c *InvResourceDAO) CreateHost(
	tb testing.TB, tenantID string, opts ...Opt[computev1.HostResource],
) *computev1.HostResource {
	tb.Helper()

	return c.CreateHostWithOpts(tb, tenantID, true, opts...)
}

func (c *InvResourceDAO) CreateHostNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[computev1.HostResource],
) *computev1.HostResource {
	tb.Helper()

	return c.CreateHostWithOpts(tb, tenantID, false, opts...)
}

func (c *InvResourceDAO) createHostnic(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
	doCleanup bool,
) *computev1.HostnicResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	nic := &computev1.HostnicResource{
		Host:     host,
		TenantId: tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Hostnic{Hostnic: nic},
		},
	)
	require.NoError(tb, err)
	nicResp := resp.GetHostnic()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, nicResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	nicResp.Host = nil

	return nicResp
}

// Create host nic. Note this helper is not really meant to be used for the
// test of HostnicResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateHostNic(tb testing.TB, tenantID string, host *computev1.HostResource) *computev1.HostnicResource {
	tb.Helper()

	return c.createHostnic(tb, tenantID, host, true)
}

func (c *InvResourceDAO) CreateHostNicNoCleanup(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
) (nic *computev1.HostnicResource) {
	tb.Helper()

	return c.createHostnic(tb, tenantID, host, false)
}

func (c *InvResourceDAO) createHostGPU(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
	cleanup bool,
) (gpu *computev1.HostgpuResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	gpu = &computev1.HostgpuResource{
		DeviceName:  "Test GPU",
		PciId:       "00:00.1",
		Product:     "some product name",
		Vendor:      "some vendor",
		Description: "for unit testing purposes",

		Host:     host,
		TenantId: tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Hostgpu{Hostgpu: gpu},
		})
	require.NoError(tb, err)

	gpuResp := resp.GetHostgpu()
	if cleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, gpuResp.ResourceId) })
	}

	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	gpuResp.Host = nil

	return gpuResp
}

func (c *InvResourceDAO) CreateHostGPU(
	tb testing.TB, tenantID string, host *computev1.HostResource,
) (gpu *computev1.HostgpuResource) {
	tb.Helper()

	return c.createHostGPU(tb, tenantID, host, true)
}

func (c *InvResourceDAO) CreateHostGPUNoCleanup(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
) (gpu *computev1.HostgpuResource) {
	tb.Helper()

	return c.createHostGPU(tb, tenantID, host, false)
}

// Create IPAddress. Note this helper is not really meant to be used for the
// test of IPAddressResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateIPAddress(
	tb testing.TB,
	tenantID string,
	hostNic *computev1.HostnicResource,
	cleanup bool,
) (
	ipaddress *network_v1.IPAddressResource,
) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ipaddress = &network_v1.IPAddressResource{
		Address:      "192.168.0.1/24",
		CurrentState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
		ConfigMethod: network_v1.IPAddressConfigMethod_IP_ADDRESS_CONFIG_METHOD_DYNAMIC,
		Nic:          hostNic,
		TenantId:     tenantID,
	}
	resp, err := c.rmClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Ipaddress{Ipaddress: ipaddress},
		})
	if err != nil {
		tb.Error(err)
		tb.FailNow()
	}
	ipAddressResp := resp.GetIpaddress()
	if cleanup {
		tb.Cleanup(func() { c.HardDeleteIPAddress(tb, tenantID, ipAddressResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	ipAddressResp.Nic = nil

	return ipAddressResp
}

// CreateRepeatedSchedule - creates repeated schedule instance and registers cleanup hooks.
// Note this helper is not really meant to be used for the test of RepeatedScheduleResource,
// but they are typically leveraged in case of wider tests involving long chain of relations
// that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateRepeatedSchedule(
	tb testing.TB,
	tenantID string,
	opts ...Opt[schedule_v1.RepeatedScheduleResource],
) *schedule_v1.RepeatedScheduleResource {
	tb.Helper()
	return c.createRepeatedSchedule(tb, tenantID, true, opts...)
}

// CreateRepeatedScheduleNoCleanup - creates repeated schedule instance, do not register cleanup hooks.
// Note this helper is not really meant to be used for the test of RepeatedScheduleResource,
// but they are typically leveraged in case of wider tests involving long chain of relations
// that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateRepeatedScheduleNoCleanup(
	tb testing.TB,
	tenantID string,
	opts ...Opt[schedule_v1.RepeatedScheduleResource],
) *schedule_v1.RepeatedScheduleResource {
	tb.Helper()
	return c.createRepeatedSchedule(tb, tenantID, false, opts...)
}

func (c *InvResourceDAO) createRepeatedSchedule(
	tb testing.TB,
	tenantID string,
	doCleanup bool,
	opts ...Opt[schedule_v1.RepeatedScheduleResource],
) *schedule_v1.RepeatedScheduleResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repeatedSchedule := &schedule_v1.RepeatedScheduleResource{
		Name:            "for unit testing purposes",
		ScheduleStatus:  schedule_v1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds: uint32(100), //nolint:mnd // Teting only
		CronMinutes:     "3",
		CronHours:       "4",
		CronDayMonth:    "5",
		CronMonth:       "6",
		CronDayWeek:     "0",
		TenantId:        tenantID,
	}
	collections.ForEach(opts, func(opt Opt[schedule_v1.RepeatedScheduleResource]) { opt(repeatedSchedule) })

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_Repeatedschedule{Repeatedschedule: repeatedSchedule}},
	)
	require.NoError(tb, err)
	repeatedScheduleResp := resp.GetRepeatedschedule()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, repeatedScheduleResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	repeatedScheduleResp.Relation = nil

	return repeatedScheduleResp
}

// CreateSingleSchedule - creates single schedule instance and registers cleanup hooks.
// Note this helper is not really meant to be used for the test of RepeatedScheduleResource,
// but they are typically leveraged in case of wider tests involving long chain of relations
// that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateSingleSchedule(
	tb testing.TB, tenantID string, opts ...Opt[schedule_v1.SingleScheduleResource],
) *schedule_v1.SingleScheduleResource {
	tb.Helper()

	return c.createSingleSchedule(tb, tenantID, true, opts...)
}

// CreateSingleScheduleNoCleanup - creates single schedule instance, do not register cleanup hooks.
// Note this helper is not really meant to be used for the test of RepeatedScheduleResource,
// but they are typically leveraged in case of wider tests involving long chain of relations
// that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateSingleScheduleNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[schedule_v1.SingleScheduleResource],
) *schedule_v1.SingleScheduleResource {
	tb.Helper()

	return c.createSingleSchedule(tb, tenantID, false, opts...)
}

func (c *InvResourceDAO) createSingleSchedule(
	tb testing.TB,
	tenantID string,
	doCleanup bool,
	opts ...Opt[schedule_v1.SingleScheduleResource],
) *schedule_v1.SingleScheduleResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := uint64(time.Now().Unix()) //nolint:gosec // no overflow for a few billion years
	singleSchedule := &schedule_v1.SingleScheduleResource{
		Name:           "for unit testing purposes",
		ScheduleStatus: schedule_v1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		StartSeconds:   now + 3600, //nolint:mnd // Testing only
		EndSeconds:     now + 7200, //nolint:mnd // Testing only
		TenantId:       tenantID,
	}

	collections.ForEach(opts, func(opt Opt[schedule_v1.SingleScheduleResource]) { opt(singleSchedule) })

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_Singleschedule{Singleschedule: singleSchedule}},
	)
	require.NoError(tb, err)
	singleScheduleResp := resp.GetSingleschedule()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, singleScheduleResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	singleScheduleResp.Relation = nil

	return singleScheduleResp
}

func (c *InvResourceDAO) CreateTelemetryGroupMetrics(
	tb testing.TB, tenantID string, cleanup bool,
) *telemetry_v1.TelemetryGroupResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tr := &telemetry_v1.TelemetryGroupResource{
		Name:          "for unit testing purposes",
		Kind:          telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		CollectorKind: telemetry_v1.CollectorKind_COLLECTOR_KIND_HOST,
		Groups:        []string{"cpu", "memory"},
		TenantId:      tenantID,
	}

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{TelemetryGroup: tr}},
	)
	require.NoError(tb, err)

	trResp := resp.GetTelemetryGroup()
	if cleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, trResp.ResourceId) })
	}

	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	trResp.Profiles = nil

	return trResp
}

func (c *InvResourceDAO) CreateTelemetryGroupLogs(
	tb testing.TB, tenantID string, cleanup bool,
) *telemetry_v1.TelemetryGroupResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tr := &telemetry_v1.TelemetryGroupResource{
		Name:          "for unit testing purposes",
		Kind:          telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		CollectorKind: telemetry_v1.CollectorKind_COLLECTOR_KIND_HOST,
		Groups:        []string{"kmsg"},
		TenantId:      tenantID,
	}

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{TelemetryGroup: tr}},
	)
	require.NoError(tb, err)

	trResp := resp.GetTelemetryGroup()
	if cleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, trResp.ResourceId) })
	}

	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	trResp.Profiles = nil

	return trResp
}

type TelemetryProfileTargetConfigurator Opt[telemetry_v1.TelemetryProfile]

func (c *InvResourceDAO) CreateTelemetryProfile(
	tb testing.TB,
	tenantID string,
	configureTarget TelemetryProfileTargetConfigurator,
	group *telemetry_v1.TelemetryGroupResource,
	cleanup bool,
) *telemetry_v1.TelemetryProfile {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tp := &telemetry_v1.TelemetryProfile{
		Kind:     group.Kind,
		Group:    group,
		TenantId: tenantID,
	}

	switch group.Kind {
	case telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS:
		tp.MetricsInterval = 300
	case telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS:
		tp.LogLevel = telemetry_v1.SeverityLevel_SEVERITY_LEVEL_INFO
	case telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_UNSPECIFIED:
		break
	}

	configureTarget(tp)
	require.NotNil(tb, tp.Relation, "TelemetryProfile has to any target (Region, Site, Instance)")

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryProfile{TelemetryProfile: tp}},
	)
	require.NoError(tb, err)

	tpResp := resp.GetTelemetryProfile()
	if cleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, tpResp.ResourceId) })
	}

	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	tpResp.Group = nil
	tpResp.Relation = nil

	return tpResp
}

func (c *InvResourceDAO) createOsWithOpts(
	tb testing.TB,
	tenantID string,
	doCleanup bool,
	opts ...Opt[osv1.OperatingSystemResource],
) *osv1.OperatingSystemResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// a default OS resource, can be overwritten by opts
	osCreateReq := &osv1.OperatingSystemResource{
		Name:                 "for unit testing purposes",
		UpdateSources:        []string{"test entries"},
		ImageUrl:             "Repo URL Test",
		ImageId:              "some image ID",
		ProfileName:          "test profile name",
		ProfileVersion:       "1.0.0",
		Sha256:               "test sha256",
		InstalledPackages:    "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
		InstalledPackagesUrl: "https://manifest-url.example.com/installed-packages.txt",
		SecurityFeature:      osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
		OsType:               osv1.OsType_OS_TYPE_MUTABLE,
		OsProvider:           osv1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
		TenantId:             tenantID,
	}

	for _, opt := range opts {
		opt(osCreateReq)
	}

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_Os{Os: osCreateReq}},
	)
	require.NoError(tb, err)

	osResp := resp.GetOs()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, osResp.ResourceId) })
	}

	return osResp
}

// Deprecated: Use CreateOsWithOpts instead.
func (c *InvResourceDAO) CreateOsWithArgs(
	tb testing.TB,
	tenantID, sha256Hex, name, profileName string,
	feature osv1.SecurityFeature, osType osv1.OsType,
) *osv1.OperatingSystemResource {
	tb.Helper()

	return c.createOsWithOpts(tb, tenantID, true, func(osr *osv1.OperatingSystemResource) {
		osr.Sha256 = sha256Hex
		osr.ProfileName = profileName
		osr.Name = name
		osr.SecurityFeature = feature
		osr.OsType = osType
	})
}

func (c *InvResourceDAO) CreateOsWithOpts(
	tb testing.TB,
	tenantID string,
	doCleanup bool,
	opts ...Opt[osv1.OperatingSystemResource],
) *osv1.OperatingSystemResource {
	tb.Helper()
	return c.createOsWithOpts(tb, tenantID, doCleanup, opts...)
}

// CreateOs creates mutable OSResource by default. Use CreateOsWithOpts to customize OS fields, if needed.
func (c *InvResourceDAO) CreateOs(tb testing.TB, tenantID string) *osv1.OperatingSystemResource {
	tb.Helper()

	return c.createOsWithOpts(
		tb,
		tenantID,
		true,
		func(osr *osv1.OperatingSystemResource) {
			osr.Sha256 = GenerateRandomSha256()
			osr.ProfileName = GenerateRandomProfileName()
			osr.Name = GenerateRandomOsResourceName()
			osr.SecurityFeature = osv1.SecurityFeature_SECURITY_FEATURE_UNSPECIFIED
			osr.OsType = osv1.OsType_OS_TYPE_MUTABLE
		},
	)
}

// CreateOsNoCleanup creates mutable OSResource by default (with no cleanup).
// Use CreateOsWithOpts to customize OS fields, if needed.
func (c *InvResourceDAO) CreateOsNoCleanup(tb testing.TB, tenantID string) *osv1.OperatingSystemResource {
	tb.Helper()

	return c.createOsWithOpts(
		tb,
		tenantID,
		false,
		func(osr *osv1.OperatingSystemResource) {
			osr.Sha256 = GenerateRandomSha256()
			osr.ProfileName = GenerateRandomProfileName()
			osr.Name = GenerateRandomOsResourceName()
			osr.SecurityFeature = osv1.SecurityFeature_SECURITY_FEATURE_UNSPECIFIED
			osr.OsType = osv1.OsType_OS_TYPE_MUTABLE
		},
	)
}

// CreateSite - creates site and takes care about cleanup. Note this helper is not really meant to be used for the
// test of SiteResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateSite(
	tb testing.TB, tenantID string, opts ...Opt[location_v1.SiteResource],
) *location_v1.SiteResource {
	tb.Helper()

	return c.createSite(tb, tenantID, true, opts...)
}

// CreateSiteNoCleanup - creates site and does not register cleanup handlers.
// Note this helper is not really meant to be used for the test of SiteResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateSiteNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[location_v1.SiteResource],
) *location_v1.SiteResource {
	tb.Helper()

	return c.createSite(tb, tenantID, false, opts...)
}

func (c *InvResourceDAO) createSite(
	tb testing.TB, tenantID string, doCleanup bool, opts ...Opt[location_v1.SiteResource],
) *location_v1.SiteResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	site := &location_v1.SiteResource{
		Name:             "for unit testing purposes",
		DnsServers:       []string{},
		DockerRegistries: []string{},
		TenantId:         tenantID,
	}
	collections.ForEach(opts, func(o Opt[location_v1.SiteResource]) { o(site) })
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Site{Site: site},
		})
	require.NoError(tb, err)

	siteResp := resp.GetSite()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, siteResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	siteResp.Region = nil
	siteResp.Ou = nil

	return siteResp
}

// CreateInstance creates instance with a cleanup. Note this helper is not really meant to be used for the
// test of InstanceResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateInstance(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		true,
		func(_ *computev1.InstanceResource) {})
}

func (c *InvResourceDAO) createInstanceWithOpts(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	doCleanup bool,
	opts ...Opt[computev1.InstanceResource],
) *computev1.InstanceResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// a default Instance resource, can be overwritten by opts
	instCreateReq := &computev1.InstanceResource{
		Kind:            computev1.InstanceKind_INSTANCE_KIND_METAL,
		Name:            dummyInstanceName,
		DesiredState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
		DesiredOs:       osRes,
		CurrentOs:       osRes, // always create with desired OS == current OS for testing
		Os:              osRes,
		Host:            hostRes,
		SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_UNSPECIFIED,
		TenantId:        tenantID,
	}

	for _, opt := range opts {
		opt(instCreateReq)
	}

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_Instance{Instance: instCreateReq}},
	)
	require.NoError(tb, err)

	instResp := resp.GetInstance()
	if doCleanup {
		tb.Cleanup(func() { c.HardDeleteInstance(tb, tenantID, instResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	instResp.DesiredOs = nil
	instResp.CurrentOs = nil
	instResp.Os = nil
	instResp.Host = nil
	instResp.WorkloadMembers = nil
	instResp.Provider = nil
	instResp.Localaccount = nil
	instResp.CustomConfig = nil

	return instResp
}

func (c *InvResourceDAO) CreateInstanceWithOpts(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	doCleanup bool,
	opts ...Opt[computev1.InstanceResource],
) *computev1.InstanceResource {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		doCleanup,
		opts...)
}

// Deprecated: Use CreateInstanceWithOpts instead.
func (c *InvResourceDAO) CreateInstanceWithArgs(
	tb testing.TB,
	tenantID, instanceName string,
	securityFeature osv1.SecurityFeature,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	proRes *provider_v1.ProviderResource,
	lcRes *localaccount_v1.LocalAccountResource,
	doCleanup bool,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		doCleanup,
		func(inst *computev1.InstanceResource) {
			inst.Name = instanceName
			inst.SecurityFeature = securityFeature
			inst.Provider = proRes
			inst.Localaccount = lcRes
		})
}

// CreateInstanceNoCleanup creates instance with NO cleanup. Note this helper is not really meant to be used for the
// test of InstanceResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateInstanceNoCleanup(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		false,
		func(_ *computev1.InstanceResource) {})
}

func (c *InvResourceDAO) CreateInstanceWithProvider(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	proRes *provider_v1.ProviderResource,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		true,
		func(inst *computev1.InstanceResource) {
			inst.Provider = proRes
		})
}

func (c *InvResourceDAO) CreateInstanceWithCustomConfig(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	customConfig []*computev1.CustomConfigResource,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		true,
		func(inst *computev1.InstanceResource) {
			inst.CustomConfig = customConfig // Assigning from slice
		})
}

func (c *InvResourceDAO) CreateInstanceWithLocalAccount(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	accRes *localaccount_v1.LocalAccountResource,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		true,
		func(inst *computev1.InstanceResource) {
			inst.Localaccount = accRes
		})
}

func (c *InvResourceDAO) CreateInstanceWithProviderNoCleanup(
	tb testing.TB,
	tenantID string,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	proRes *provider_v1.ProviderResource,
) (ins *computev1.InstanceResource) {
	tb.Helper()

	return c.createInstanceWithOpts(
		tb,
		tenantID,
		hostRes,
		osRes,
		false,
		func(inst *computev1.InstanceResource) {
			inst.Provider = proRes
		})
}

// Create host storage. Note this helper is not really meant to be used for the
// test of HoststorageResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) createHostStorage(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
	doCleanup bool,
) (storage *computev1.HoststorageResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	storage = &computev1.HoststorageResource{
		Host:          host,
		CapacityBytes: 1000 * util.Gigabyte, //nolint:mnd // Teting only
		TenantId:      tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Hoststorage{Hoststorage: storage},
		})
	require.NoError(tb, err)
	storageResp := resp.GetHoststorage()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, storageResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	storageResp.Host = nil

	return storageResp
}

// Create host storage. Note this helper is not really meant to be used for the
// test of HoststorageResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateHostStorage(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
) (storage *computev1.HoststorageResource) {
	tb.Helper()

	return c.createHostStorage(tb, tenantID, host, true)
}

func (c *InvResourceDAO) CreateHostStorageNoCleanup(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
) (storage *computev1.HoststorageResource) {
	tb.Helper()

	return c.createHostStorage(tb, tenantID, host, false)
}

func (c *InvResourceDAO) createHostusb(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
	doCleanup bool,
) (usb *computev1.HostusbResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	usb = &computev1.HostusbResource{
		Kind:     "",
		Host:     host,
		TenantId: tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Hostusb{Hostusb: usb},
		})
	require.NoError(tb, err)
	usbResp := resp.GetHostusb()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, usbResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	usbResp.Host = nil

	return usbResp
}

// Create host usb. Note this helper is not really meant to be used for the
// test of HostusbResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateHostUsb(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
) (usb *computev1.HostusbResource) {
	tb.Helper()

	return c.createHostusb(tb, tenantID, host, true)
}

func (c *InvResourceDAO) CreateHostUsbNoCleanup(
	tb testing.TB,
	tenantID string,
	host *computev1.HostResource,
) (usb *computev1.HostusbResource) {
	tb.Helper()

	return c.createHostusb(tb, tenantID, host, false)
}

// Create CustomConfig. Note this helper is not really meant to be used for the
// test of CustomConfigResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateCustomConfig(
	tb testing.TB, tenantID, name, description, config string,
) (cconfig *computev1.CustomConfigResource) {
	tb.Helper()

	return c.createCustomConfig(tb, tenantID, name, description, config, true)
}

func (c *InvResourceDAO) CreateCustomConfigNoCleanup(
	tb testing.TB, tenantID, name, description, config string,
) (cconfig *computev1.CustomConfigResource) {
	tb.Helper()

	return c.createCustomConfig(tb, tenantID, name, description, config, false)
}

// Create local account. Note this helper is not really meant to be used for the
// test of LocalAccountResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateLocalAccount(
	tb testing.TB, tenantID, username, sshKey string,
) (account *localaccount_v1.LocalAccountResource) {
	tb.Helper()

	return c.createLocalAccount(tb, tenantID, username, sshKey, true)
}

func (c *InvResourceDAO) CreateLocalAccountNoCleanup(
	tb testing.TB, tenantID, username, sshKey string,
) (provider *localaccount_v1.LocalAccountResource) {
	tb.Helper()

	return c.createLocalAccount(tb, tenantID, username, sshKey, false)
}

// Create provider. Note this helper is not really meant to be used for the
// test of ProviderResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateProvider(
	tb testing.TB, tenantID, name string, opts ...Opt[provider_v1.ProviderResource],
) (provider *provider_v1.ProviderResource) {
	tb.Helper()

	apiEndpoint := "https://192.168.201.3/discovery"
	vendor := provider_v1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LXCA
	credentials := []string{"test", "test"}

	return c.createProvider(tb, tenantID, name, apiEndpoint, credentials, vendor, true, opts...)
}

func (c *InvResourceDAO) CreateProviderNoCleanup(
	tb testing.TB, tenantID, name string, opts ...Opt[provider_v1.ProviderResource],
) (provider *provider_v1.ProviderResource) {
	tb.Helper()

	apiEndpoint := "https://192.168.201.3/discovery"
	vendor := provider_v1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LXCA
	credentials := []string{"test", "test"}

	return c.createProvider(tb, tenantID, name, apiEndpoint, credentials, vendor, false, opts...)
}

func (c *InvResourceDAO) CreateProviderWithArgs(tb testing.TB,
	tenantID, name, apiEndpoint string,
	credentials []string,
	vendor provider_v1.ProviderVendor,
	opts ...Opt[provider_v1.ProviderResource],
) (provider *provider_v1.ProviderResource) {
	tb.Helper()

	return c.createProvider(tb, tenantID, name, apiEndpoint, credentials, vendor, true, opts...)
}

func (c *InvResourceDAO) CreateProviderWithArgsNoCleanup(tb testing.TB,
	tenantID, name, apiEndpoint string,
	credentials []string,
	vendor provider_v1.ProviderVendor,
	opts ...Opt[provider_v1.ProviderResource],
) (provider *provider_v1.ProviderResource) {
	tb.Helper()

	return c.createProvider(tb, tenantID, name, apiEndpoint, credentials, vendor, false, opts...)
}

// Create provider.
func (c *InvResourceDAO) createProvider(tb testing.TB,
	tenantID, name, apiEndpoint string,
	credentials []string,
	vendor provider_v1.ProviderVendor,
	doCleanup bool,
	opts ...Opt[provider_v1.ProviderResource],
) (provider *provider_v1.ProviderResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	provider = &provider_v1.ProviderResource{
		ProviderVendor: vendor,
		Name:           name,
		ApiEndpoint:    apiEndpoint,
		ApiCredentials: credentials,
		TenantId:       tenantID,
	}
	collections.ForEach(opts, func(opt Opt[provider_v1.ProviderResource]) { opt(provider) })
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Provider{Provider: provider},
		})
	require.NoError(tb, err)
	providerResp := resp.GetProvider()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, providerResp.ResourceId) })
	}

	return providerResp
}

// Create CustomConfig.
func (c *InvResourceDAO) createCustomConfig(
	tb testing.TB, tenantID, name, description, config string, doCleanup bool,
) (cconfig *computev1.CustomConfigResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cconfig = &computev1.CustomConfigResource{
		Name:        name,
		Description: description,
		Config:      config,
		TenantId:    tenantID,
	}

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_CustomConfig{CustomConfig: cconfig}},
	)
	require.NoError(tb, err)
	cconfigResp := resp.GetCustomConfig()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, cconfigResp.ResourceId) })
	}

	return cconfigResp
}

// Create LocalAccount.
func (c *InvResourceDAO) createLocalAccount(tb testing.TB,
	tenantID, username, sshKey string,
	doCleanup bool,
) (account *localaccount_v1.LocalAccountResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	account = &localaccount_v1.LocalAccountResource{
		Username: username,
		SshKey:   sshKey,
		TenantId: tenantID,
	}

	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_LocalAccount{LocalAccount: account},
		})
	require.NoError(tb, err)
	accountResp := resp.GetLocalAccount()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, accountResp.ResourceId) })
	}

	return accountResp
}

func (c *InvResourceDAO) CreateOSUpdatePolicyNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[computev1.OSUpdatePolicyResource],
) *computev1.OSUpdatePolicyResource {
	tb.Helper()

	oup, err := c.createOSUpdatePolicy(tb, tenantID, opts...)
	require.NoError(tb, err)
	return oup
}

func (c *InvResourceDAO) CreateOSUpdatePolicy(
	tb testing.TB, tenantID string, opts ...Opt[computev1.OSUpdatePolicyResource],
) *computev1.OSUpdatePolicyResource {
	tb.Helper()

	oup, err := c.createOSUpdatePolicy(tb, tenantID, opts...)
	require.NoError(tb, err)
	tb.Cleanup(func() { c.DeleteResource(tb, tenantID, oup.ResourceId) })
	return oup
}

func (c *InvResourceDAO) createOSUpdatePolicy(
	tb testing.TB, tenantID string, opts ...Opt[computev1.OSUpdatePolicyResource],
) (*computev1.OSUpdatePolicyResource, error) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	oup := &computev1.OSUpdatePolicyResource{
		TenantId:     tenantID,
		UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_LATEST,
	}
	collections.ForEach(opts, func(opt Opt[computev1.OSUpdatePolicyResource]) { opt(oup) })
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_OsUpdatePolicy{OsUpdatePolicy: oup},
		})
	if err != nil {
		return nil, err
	}
	oupResp := resp.GetOsUpdatePolicy()
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	oupResp.TargetOs = nil

	return oupResp, nil
}

// CreateOuAndReturnError - creates ou and return error if any. This function does not take care about cleanup.
// Note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateOuAndReturnError(
	tb testing.TB, tenantID string, opts ...Opt[ou_v1.OuResource],
) (ou *ou_v1.OuResource, err error) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ou = &ou_v1.OuResource{
		Name:     "for unit testing purposes",
		OuKind:   "test BU",
		TenantId: tenantID,
	}
	collections.ForEach(opts, func(opt Opt[ou_v1.OuResource]) { opt(ou) })
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Ou{Ou: ou},
		})
	if err != nil {
		return nil, err
	}
	ouResp := resp.GetOu()
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	ouResp.ParentOu = nil

	return ouResp, nil
}

// CreateOu - create ou and takes care about cleanup. Note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateOu(tb testing.TB, tenantID string, opts ...Opt[ou_v1.OuResource]) *ou_v1.OuResource {
	tb.Helper()

	ou, err := c.CreateOuAndReturnError(tb, tenantID, opts...)
	require.NoError(tb, err)
	tb.Cleanup(func() { c.DeleteResource(tb, tenantID, ou.ResourceId) })
	return ou
}

// CreateOuNoCleanup - creates ou and does not take care about cleanup. Note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateOuNoCleanup(tb testing.TB, tenantID string, opts ...Opt[ou_v1.OuResource]) *ou_v1.OuResource {
	tb.Helper()

	ou, err := c.CreateOuAndReturnError(tb, tenantID, opts...)
	require.NoError(tb, err)
	return ou
}

// CreateRegion - create region and take care about cleanup. Note this helper is not really meant to be used for the
// test of RegionResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateRegion(
	tb testing.TB, tenantID string, opts ...Opt[location_v1.RegionResource],
) (region *location_v1.RegionResource) {
	tb.Helper()

	return c.createRegion(tb, tenantID, true, opts...)
}

// CreateRegionNoCleanup - create region and does not take care about cleanup.
// Note this helper is not really meant to be used for the test of RegionResource, but they are typically leveraged
// in case of wider tests involving long chain of relations that are not usually fulfilled by the eager loading.
func (c *InvResourceDAO) CreateRegionNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[location_v1.RegionResource],
) *location_v1.RegionResource {
	tb.Helper()
	return c.createRegion(tb, tenantID, false, opts...)
}

func (c *InvResourceDAO) createRegion(
	tb testing.TB, tenantID string, doCleanup bool, opts ...Opt[location_v1.RegionResource],
) *location_v1.RegionResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	region := &location_v1.RegionResource{
		Name:       "for unit testing purposes",
		RegionKind: "test region",
		TenantId:   tenantID,
	}
	for _, opt := range opts {
		opt(region)
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{Region: region},
		})
	require.NoError(tb, err)
	regionResp := resp.GetRegion()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, regionResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	regionResp.ParentRegion = nil

	return regionResp
}

// Hard deletes the given workload via 2-phase deletion if needed.
func (c *InvResourceDAO) HardDeleteWorkload(
	tb testing.TB,
	tenantID, resourceID string,
	workloadKind computev1.WorkloadKind,
) {
	tb.Helper()

	err := c.HardDeleteWorkloadAndReturnError(tb, tenantID, resourceID, workloadKind)
	if err != nil {
		tb.Errorf("UpdateWorkload() failed: %s", err)
		tb.FailNow()
	}
}

// Hard deletes the given workload via 2-phase deletion if needed.
func (c *InvResourceDAO) HardDeleteWorkloadAndReturnError(
	tb testing.TB,
	tenantID, resourceID string,
	workloadKind computev1.WorkloadKind,
) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := error(nil)
	err = c.DeleteResourceAndReturnError(tb, tenantID, resourceID)

	if err == nil && workloadKind != computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER {
		// 2-phase delete for non-cluster workload
		_, err = c.rmClient.Update(
			ctx,
			tenantID,
			resourceID,
			&fieldmaskpb.FieldMask{Paths: []string{workloadresource.FieldCurrentState}},
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Workload{
					Workload: &computev1.WorkloadResource{
						CurrentState: computev1.WorkloadState_WORKLOAD_STATE_DELETED,
					},
				},
			},
		)
	}
	return err
}

// Create workload. Note this helper is not really meant to be used for the
// test of WorkloadResource but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateWorkload(
	tb testing.TB, tenantID string, opts ...Opt[computev1.WorkloadResource],
) (workload *computev1.WorkloadResource) {
	tb.Helper()

	return c.createWorkload(tb, tenantID, true, opts...)
}

func (c *InvResourceDAO) CreateWorkloadNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[computev1.WorkloadResource],
) (workload *computev1.WorkloadResource) {
	tb.Helper()

	return c.createWorkload(tb, tenantID, false, opts...)
}

func (c *InvResourceDAO) createWorkload(
	tb testing.TB, tenantID string, doCleanup bool, opts ...Opt[computev1.WorkloadResource],
) (workload *computev1.WorkloadResource) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	workload = &computev1.WorkloadResource{
		Kind:         computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
		Name:         "for unit testing purposes",
		DesiredState: computev1.WorkloadState_WORKLOAD_STATE_PROVISIONED,
		Status:       "provisioned",
		Metadata:     "",
		ExternalId:   uuid.NewString(),
		TenantId:     tenantID,
	}
	for _, opt := range opts {
		opt(workload)
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Workload{Workload: workload},
		})
	if err != nil {
		tb.Error(err)
		tb.FailNow()
	}
	workloadResp := resp.GetWorkload()
	if doCleanup {
		tb.Cleanup(func() { c.HardDeleteWorkload(tb, tenantID, workloadResp.ResourceId, workloadResp.Kind) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	workloadResp.Members = nil

	return workloadResp
}

// Create WorkloadMember. Note this helper is not really meant to be used for the
// test of WorkloadMember but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateWorkloadMember(
	tb testing.TB,
	tenantID string,
	workload *computev1.WorkloadResource,
	instance *computev1.InstanceResource,
) *computev1.WorkloadMember {
	tb.Helper()
	return c.createWorkloadMember(tb, tenantID, workload, instance, true)
}

func (c *InvResourceDAO) CreateWorkloadMemberNoCleanup(
	tb testing.TB,
	tenantID string,
	workload *computev1.WorkloadResource,
	instance *computev1.InstanceResource,
) (workloadMember *computev1.WorkloadMember) {
	tb.Helper()

	return c.createWorkloadMember(tb, tenantID, workload, instance, false)
}

func (c *InvResourceDAO) createWorkloadMember(
	tb testing.TB,
	tenantID string,
	workload *computev1.WorkloadResource,
	instance *computev1.InstanceResource,
	doCleanup bool,
) (
	workloadMember *computev1.WorkloadMember,
) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	workloadMember = &computev1.WorkloadMember{
		Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
		Workload: workload,
		Instance: instance,
		TenantId: tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: workloadMember},
		})
	if err != nil {
		tb.Error(err)
		tb.FailNow()
	}
	workloadMemberResp := resp.GetWorkloadMember()
	if doCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, workloadMemberResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	workloadMemberResp.Workload = nil
	workloadMemberResp.Instance = nil

	return workloadMemberResp
}

// HardDeleteNetlink - hard deletes the given netlink via 2-phase deletion.
func (c *InvResourceDAO) HardDeleteNetlink(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.DeleteResource(tb, tenantID, resourceID)
	_, err := c.rmClient.Update(
		ctx,
		tenantID,
		resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{netlinks.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Netlink{
				Netlink: &network_v1.NetlinkResource{
					CurrentState: network_v1.NetlinkState_NETLINK_STATE_DELETED,
				},
			},
		},
	)
	require.NoError(tb, err, "UpdateNetlink() failed")
}

func (c *InvResourceDAO) CreateNetLink(
	tb testing.TB, tenantID string, doCleanUp bool, opts ...Opt[network_v1.NetlinkResource],
) *network_v1.NetlinkResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	nl := &network_v1.NetlinkResource{
		DesiredState: network_v1.NetlinkState_NETLINK_STATE_ONLINE,
		TenantId:     tenantID,
	}
	for _, opt := range opts {
		opt(nl)
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Netlink{
				Netlink: nl,
			},
		})
	require.NoError(tb, err)
	nlResp := resp.GetNetlink()
	if doCleanUp {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, nlResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	nlResp.Src = nil
	nlResp.Dst = nil
	return nlResp
}

// Create network segment. Note this helper is not really meant to be used for the
// test of NetworkSegment but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) CreateNetworkSegment(
	tb testing.TB,
	tenantID, name string,
	site *location_v1.SiteResource,
	vlanID int32,
	doCleanUp bool,
) *network_v1.NetworkSegment {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	netseg := &network_v1.NetworkSegment{
		Name:     name,
		Site:     site,
		VlanId:   vlanID,
		TenantId: tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_NetworkSegment{
				NetworkSegment: netseg,
			},
		})
	require.NoError(tb, err)
	netsegResp := resp.GetNetworkSegment()
	if doCleanUp {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, netsegResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	netsegResp.Site = nil

	return netsegResp
}

func (c *InvResourceDAO) CreateEndpointNoCleanup(
	tb testing.TB, tenantID string, host *computev1.HostResource,
) *network_v1.EndpointResource {
	tb.Helper()
	return c.createEndpoint(tb, tenantID, host, true)
}

func (c *InvResourceDAO) CreateEndpoint(
	tb testing.TB, tenantID string, host *computev1.HostResource,
) *network_v1.EndpointResource {
	tb.Helper()
	return c.createEndpoint(tb, tenantID, host, false)
}

// CreateEndpoint - Note this helper is not really meant to be used for the
// test of EndpointResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func (c *InvResourceDAO) createEndpoint(
	tb testing.TB, tenantID string, host *computev1.HostResource, noCleanup bool,
) *network_v1.EndpointResource {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	endpoint := &network_v1.EndpointResource{
		Name:     "for unit testing purposes",
		Host:     host,
		TenantId: tenantID,
	}
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Endpoint{Endpoint: endpoint},
		})
	require.NoError(tb, err)
	endpointResp := resp.GetEndpoint()
	if !noCleanup {
		tb.Cleanup(func() { c.DeleteResource(tb, tenantID, endpointResp.ResourceId) })
	}
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	endpointResp.Host = nil

	return endpointResp
}

func (c *InvResourceDAO) CreateRemoteAccessConfiguration(
	t *testing.T, tenantID string, opts ...Opt[remoteaccessv1.RemoteAccessConfiguration],
) *remoteaccessv1.RemoteAccessConfiguration {
	t.Helper()
	return c.createRemoteAccessConfiguration(t, tenantID, true, opts...)
}

func (c *InvResourceDAO) CreateRemoteAccessConfigurationNoCleanup(
	t *testing.T, tenantID string, opts ...Opt[remoteaccessv1.RemoteAccessConfiguration],
) *remoteaccessv1.RemoteAccessConfiguration {
	t.Helper()
	return c.createRemoteAccessConfiguration(t, tenantID, true, opts...)
}

func (c *InvResourceDAO) createRemoteAccessConfiguration(
	t *testing.T, tenantID string, cleanup bool, opts ...Opt[remoteaccessv1.RemoteAccessConfiguration],
) *remoteaccessv1.RemoteAccessConfiguration {
	t.Helper()

	host := c.CreateHost(t, tenantID)
	os := c.CreateOs(t, tenantID)
	instance := c.CreateInstance(t, tenantID, host, os)
	racCreateReq := &remoteaccessv1.RemoteAccessConfiguration{
		DesiredState:        remoteaccessv1.RemoteAccessState_REMOTE_ACCESS_STATE_ENABLED,
		Instance:            instance,
		ExpirationTimestamp: uint64(time.Now().Add(time.Second * 601).Unix()), //nolint:mnd,gosec // Teting only
		TenantId:            tenantID,
	}

	for _, opt := range opts {
		opt(racCreateReq)
	}

	rsp, err := c.apiClient.Create(
		context.TODO(),
		tenantID,
		&inv_v1.Resource{Resource: &inv_v1.Resource_RemoteAccess{RemoteAccess: racCreateReq}},
	)
	require.NoError(t, err, "creation request has been rejected")
	resID := rsp.GetRemoteAccess().GetResourceId()
	require.NotEmpty(t, resID, "resource creation response shall contain not empty resourceId")
	if cleanup {
		t.Cleanup(func() { c.HardDeleteRemoteAccessConfiguration(t, tenantID, resID) })
	}
	return rsp.GetRemoteAccess()
}

func (c *InvResourceDAO) HardDeleteRemoteAccessConfiguration(tb testing.TB, tenantID, resourceID string) {
	tb.Helper()
	err := c.HardDeleteRemoteAccessConfigurationAndReturnError(tb, tenantID, resourceID)
	require.NoError(tb, err, "UpdateHost() failed")
}

func (c *InvResourceDAO) HardDeleteRemoteAccessConfigurationAndReturnError(tb testing.TB, tenantID, resourceID string) error {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.DeleteResource(tb, tenantID, resourceID)
	_, err := c.rmClient.Update(
		ctx,
		tenantID,
		resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{remoteaccessv1.RemoteAccessConfigurationFieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_RemoteAccess{
				RemoteAccess: &remoteaccessv1.RemoteAccessConfiguration{
					CurrentState: remoteaccessv1.RemoteAccessState_REMOTE_ACCESS_STATE_DELETED,
				},
			},
		},
	)
	return err
}

func GetResourceIDOrFail(tb testing.TB, resource *inv_v1.Resource) string {
	tb.Helper()

	resID, err := util.GetResourceIDFromResource(resource)
	require.NoError(tb, err, "Failed to extract resource ID from Resource")
	return resID
}

type hasResourceID interface {
	GetResourceId() string
}

func OrderByResourceID[T hasResourceID](slice []T) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].GetResourceId() < slice[j].GetResourceId()
	})
}

func GetOrderByResourceID[T hasResourceID](slice []T) []T {
	s := slices.Clone(slice)
	OrderByResourceID(s)
	return s
}

type hasResourceIDAndTenantID interface {
	GetResourceId() string
	GetTenantId() string
}

func toString(h hasResourceIDAndTenantID) string {
	return fmt.Sprintf("[tenantID=%s, resourceID=%s]", h.GetTenantId(), h.GetResourceId())
}

func SortHasResourceIDAndTenantID(slice []*client.ResourceTenantIDCarrier) {
	sort.Slice(slice, func(i, j int) bool {
		return toString(slice[i]) < toString(slice[j])
	})
}

func GetSortedResourceIDSlice[T hasResourceIDAndTenantID](slice []T) []*client.ResourceTenantIDCarrier {
	resIDs := make([]*client.ResourceTenantIDCarrier, 0, len(slice))
	for _, r := range slice {
		resIDs = append(resIDs, &client.ResourceTenantIDCarrier{TenantId: r.GetTenantId(), ResourceId: r.GetResourceId()})
	}
	SortHasResourceIDAndTenantID(resIDs)
	return resIDs
}

func (c *InvResourceDAO) CreateOSUpdateRunNoCleanup(
	tb testing.TB, tenantID string, opts ...Opt[computev1.OSUpdateRunResource],
) *computev1.OSUpdateRunResource {
	tb.Helper()

	our, err := c.createOSUpdateRun(tb, tenantID, opts...)
	require.NoError(tb, err)
	return our
}

func (c *InvResourceDAO) CreateOSUpdateRun(
	tb testing.TB, tenantID string, opts ...Opt[computev1.OSUpdateRunResource],
) *computev1.OSUpdateRunResource {
	tb.Helper()

	our, err := c.createOSUpdateRun(tb, tenantID, opts...)
	require.NoError(tb, err)
	tb.Cleanup(func() { c.DeleteResource(tb, tenantID, our.ResourceId) })
	return our
}

func (c *InvResourceDAO) createOSUpdateRun(
	tb testing.TB, tenantID string, opts ...Opt[computev1.OSUpdateRunResource],
) (*computev1.OSUpdateRunResource, error) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	our := &computev1.OSUpdateRunResource{
		TenantId: tenantID,
	}
	collections.ForEach(opts, func(opt Opt[computev1.OSUpdateRunResource]) { opt(our) })
	resp, err := c.apiClient.Create(
		ctx,
		tenantID,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_OsUpdateRun{OsUpdateRun: our},
		})
	if err != nil {
		return nil, err
	}
	ourResp := resp.GetOsUpdateRun()
	// When this test object is used in protobuf comparisons as part of another
	// resource, we do not expect further embedded messages. This matches the
	// structure of objects returned by ent queries, i.e. no two layers of
	// embedded objects for edges.
	ourResp.AppliedPolicy = nil
	ourResp.Instance = nil

	return ourResp, nil
}
