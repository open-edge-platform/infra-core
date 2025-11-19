// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package datamodel

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/flags"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	inv_util "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/controller"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/invclient"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/nexus"
	testutils "github.com/open-edge-platform/infra-core/tenant-controller/internal/testing"
	baseconfiginfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/config.edge-orchestrator.intel.com/v1"
	basefolderinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/folder.edge-orchestrator.intel.com/v1"
	baseorginfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/org.edge-orchestrator.intel.com/v1"
	baseprojectinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/project.edge-orchestrator.intel.com/v1"
	baseprojectactivewatcherinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	baseruntimeinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtime.edge-orchestrator.intel.com/v1"
	baseruntimefolderinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimefolder.edge-orchestrator.intel.com/v1"
	baseruntimeorginfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimeorg.edge-orchestrator.intel.com/v1"
	baseruntimeprojectinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimeproject.edge-orchestrator.intel.com/v1"
	tenancyv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/tenancy.edge-orchestrator.intel.com/v1"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
)

var (
	eventuallyTimeout  = 20 * time.Second
	eventuallyInterval = time.Second / 2
)

const (
	mtName         = "default"
	configName     = "default"
	runtimeName    = "default"
	org1Name       = "org1"
	org1FolderName = "default"
)

//nolint:gochecknoinits //just for test purposes
func init() {
	nexus.CreateObjectMeta = func(name string) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name:            name,
			ResourceVersion: "1",
		}
	}
}

const (
	lenovoResourceDefinitionFile = "../../configuration/default/resources-lenovo.json"
	resourceDefinitionFile       = "../../configuration/default/resources.json"
)

//nolint:funlen // just a table test
func TestDataModelSanity(t *testing.T) {
	*flags.FlagDisableCredentialsManagement = true

	cl, err := configuration.NewInitResourcesProvider(resourceDefinitionFile)
	require.NoError(t, err)

	lcl, err := configuration.NewLenovoInitResourcesDefinitionLoader(lenovoResourceDefinitionFile)
	require.NoError(t, err)

	ic := testutils.CreateInvClient(t)
	nxc := nexus.NewClient(nexus_client.NewFakeClient())

	terminationCtrl := controller.NewTerminationController(ic)
	initializationCtrl := controller.NewTenantInitializationController(
		[]configuration.InitResourcesProvider{cl, lcl},
		ic,
		nxc,
	)

	termChan := make(chan bool, 1)
	// ensure proper cleanup avoiding fatal during tests
	defer close(termChan)
	controller.NewEventDispatcher(ic, initializationCtrl, terminationCtrl).Start(termChan)
	dmc := NewDataModelController(nxc, true, terminationCtrl, initializationCtrl)
	require.NoError(t, dmc.Start(termChan))
	// create org and project
	createOrg(t, nxc)

	t.Run("Data Model", func(t *testing.T) {
		projectInfo := inv_util.NewTuple(t.Name(), uuid.NewString())
		t.Run("create project under default org", func(t *testing.T) {
			createProjectUnderDefaultOrg(t, nxc, projectInfo)
		})
	})

	t.Run("Tenant already deleted on Edge Infrastructure Manager side", func(t *testing.T) {
		projectInfo := inv_util.NewTuple(t.Name(), uuid.NewString())

		t.Run("create project under default org", func(t *testing.T) {
			createProjectUnderDefaultOrg(t, nxc, projectInfo)
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		t.Run("tenant exists and has desired_state eq CREATED", func(t *testing.T) {
			require.Eventuallyf(t, func() bool {
				tenant, e := ic.GetTenantResourceInstance(ctx, projectInfo.B)
				if e != nil {
					return false
				}
				return tenantv1.TenantState_TENANT_STATE_CREATED == tenant.GetDesiredState()
			}, eventuallyTimeout, eventuallyInterval,
				"EXPECTED: tenant[%s] with current_state == created was expected", projectInfo.B)
		})

		t.Run("emulate OSRM Reporting Tenant Watcher Eq True", func(t *testing.T) {
			tenant, err := ic.GetTenantResourceInstance(ctx, projectInfo.B)
			require.NoError(t, err)
			emulateOSRMReportingTenantWatcherEq(t, tenant.GetTenantId(), tenant.GetResourceId(), true)
			// Give the TenantInitializationController time to process the event
			time.Sleep(100 * time.Millisecond)
		})

		t.Run("ActiveWatcher's state shall be IDLE", func(t *testing.T) {
			require.Eventuallyf(t,
				func() bool {
					time.Sleep(150 * time.Millisecond)
					aw, e := getActiveWatcher(ctx, nxc, projectInfo.A)
					if e != nil || aw == nil {
						return false
					}
					return aw.Spec.StatusIndicator == baseprojectactivewatcherinfrahostcomv1.StatusIndicationIdle
				},
				eventuallyTimeout,
				eventuallyInterval,
				"EXPECTED: ACTIVEWATCHER[%s].spec.statusIndicator == IDLE was expected to appear on datamodel",
				configuration.AppName,
			)
		})

		t.Run("emulate OSRM Reporting Tenant Watcher Eq False", func(t *testing.T) {
			tenant, err := ic.GetTenantResourceInstance(ctx, projectInfo.B)
			require.NoError(t, err)
			emulateOSRMReportingTenantWatcherEq(t, tenant.GetTenantId(), tenant.GetResourceId(), false)
		})

		t.Run("terminate tenant", func(t *testing.T) {
			require.NoError(t, terminationCtrl.TerminateTenant(ctx, projectInfo.B))
		})

		// there is no tenant existing on infra side, lets trigger project deletion, and execute tenant termination one more time
		t.Run("delete runtime project", func(t *testing.T) {
			deleteRuntimeProject(ctx, t, nxc, projectInfo.A)
		})

		t.Run("AW shall be deleted", func(t *testing.T) {
			assertActiveWatcherDoesNotExist(t, nxc, projectInfo.A)
		})

		t.Run("terminate terminated tenant", func(t *testing.T) {
			require.NoError(t, terminationCtrl.TerminateTenant(ctx, projectInfo.B))
		})
	})

	t.Run("Happy Path - only initial resources", func(t *testing.T) {
		projectInfo := inv_util.NewTuple(t.Name(), uuid.NewString())
		t.Run("create project under default org", func(t *testing.T) {
			createProjectUnderDefaultOrg(t, nxc, projectInfo)
		})

		verifyIntegrationWithDataModel(t, projectInfo, nxc, ic)
	})

	t.Run("Happy Path - simple model", func(t *testing.T) {
		projectInfo := inv_util.NewTuple(t.Name(), uuid.NewString())
		createProjectUnderDefaultOrg(t, nxc, projectInfo)

		// create additional INV resources
		dao := inv_testing.NewInvResourceDAOOrFail(t)

		region := dao.CreateRegionNoCleanup(t, projectInfo.B)
		site := dao.CreateSiteNoCleanup(t, projectInfo.B, inv_testing.SiteRegion(region))

		provider := dao.CreateProviderNoCleanup(t, projectInfo.B, "provider1",
			inv_testing.ProviderKind(providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL))

		host := dao.CreateHostNoCleanup(t, projectInfo.B, inv_testing.HostSite(site), inv_testing.HostProvider(provider))
		os := dao.CreateOsNoCleanup(t, projectInfo.B)
		dao.CreateInstanceNoCleanup(t, projectInfo.B, host, os)

		verifyIntegrationWithDataModel(t, projectInfo, nxc, ic)
	})

	t.Run("Happy Path - complex", func(t *testing.T) {
		projectInfo := inv_util.NewTuple(t.Name(), uuid.NewString())
		createProjectUnderDefaultOrg(t, nxc, projectInfo)

		// create additional INV resources
		dao := inv_testing.NewInvResourceDAOOrFail(t)
		region := dao.CreateRegionNoCleanup(t, projectInfo.B)
		provider := dao.CreateProviderNoCleanup(t, projectInfo.B, "provider",
			inv_testing.ProviderKind(providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL))
		ou := dao.CreateOuNoCleanup(t, projectInfo.B)
		site := dao.CreateSiteNoCleanup(t, projectInfo.B, inv_testing.SiteProvider(provider), inv_testing.SiteOu(ou),
			inv_testing.SiteRegion(region))
		host := dao.CreateHostNoCleanup(t, projectInfo.B, inv_testing.HostSite(site), inv_testing.HostProvider(provider))
		os := dao.CreateOsNoCleanup(t, projectInfo.B)
		instance := dao.CreateInstanceNoCleanup(t, projectInfo.B, host, os)
		workload := dao.CreateWorkloadNoCleanup(t, projectInfo.B)
		dao.CreateWorkloadMemberNoCleanup(t, projectInfo.B, workload, instance)

		verifyIntegrationWithDataModel(t, projectInfo, nxc, ic)
	})

	t.Run("Happy Path - even more complex", func(t *testing.T) {
		projectInfo := inv_util.NewTuple(t.Name(), uuid.NewString())
		createProjectUnderDefaultOrg(t, nxc, projectInfo)

		// create additional INV resources
		dao := inv_testing.NewInvResourceDAOOrFail(t)

		provider1 := dao.CreateProviderNoCleanup(t, projectInfo.B, "provider1",
			inv_testing.ProviderKind(providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL))
		provider2 := dao.CreateProviderNoCleanup(t, projectInfo.B, "provider2",
			inv_testing.ProviderKind(providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL))

		ou1 := dao.CreateOuNoCleanup(t, projectInfo.B)
		ou2 := dao.CreateOuNoCleanup(t, projectInfo.B, inv_testing.OuParent(ou1))
		ou3 := dao.CreateOuNoCleanup(t, projectInfo.B, inv_testing.OuParent(ou2))
		dao.CreateOuNoCleanup(t, projectInfo.B, inv_testing.OuParent(ou3))

		topLevelRegion := dao.CreateRegionNoCleanup(t, projectInfo.B)
		region1 := dao.CreateRegionNoCleanup(t, projectInfo.B, inv_testing.RegionParentRegion(topLevelRegion))
		region2 := dao.CreateRegionNoCleanup(t, projectInfo.B, inv_testing.RegionParentRegion(region1))
		region3 := dao.CreateRegionNoCleanup(t, projectInfo.B)

		site1 := dao.CreateSiteNoCleanup(t,
			projectInfo.B,
			inv_testing.SiteProvider(provider1),
			inv_testing.SiteOu(ou1),
			inv_testing.SiteRegion(region2))

		_ = dao.CreateSiteNoCleanup(t, projectInfo.B, inv_testing.SiteRegion(region3))

		host1 := dao.CreateHostNoCleanup(t, projectInfo.B, inv_testing.HostSite(site1), inv_testing.HostProvider(provider2))
		host2 := dao.CreateHostNoCleanup(t, projectInfo.B, inv_testing.HostSite(site1), inv_testing.HostProvider(provider2))
		dao.CreateNetworkSegment(t, projectInfo.B, "ns1", site1, 1234, false)

		dao.CreateEndpointNoCleanup(t, projectInfo.B, host1)

		tg1 := dao.CreateTelemetryGroupLogs(t, projectInfo.B, false)
		tg2 := dao.CreateTelemetryGroupMetrics(t, projectInfo.B, false)

		dao.CreateTelemetryProfile(t, projectInfo.B, inv_testing.TelemetryProfileTarget(site1), tg1, false)
		dao.CreateTelemetryProfile(t, projectInfo.B, inv_testing.TelemetryProfileTarget(site1), tg2, false)

		os := dao.CreateOsNoCleanup(t, projectInfo.B)

		instance1 := dao.CreateInstanceNoCleanup(t, projectInfo.B, host1, os)
		instance2 := dao.CreateInstanceNoCleanup(t, projectInfo.B, host2, os)

		workload1 := dao.CreateWorkloadNoCleanup(t, projectInfo.B)
		workload2 := dao.CreateWorkloadNoCleanup(t,
			projectInfo.B, inv_testing.WorkloadKind(computev1.WorkloadKind_WORKLOAD_KIND_DHCP))
		dao.CreateWorkloadMemberNoCleanup(t, projectInfo.B, workload1, instance1)
		dao.CreateWorkloadMemberNoCleanup(t, projectInfo.B, workload2, instance2)
		dao.CreateRepeatedScheduleNoCleanup(t, projectInfo.B, inv_testing.RSRRegion(region1))
		dao.CreateRepeatedScheduleNoCleanup(t, projectInfo.B, inv_testing.RSRTargetSite(site1))
		dao.CreateRepeatedScheduleNoCleanup(t, projectInfo.B, inv_testing.RSRTargetHost(host1))
		dao.CreateRepeatedScheduleNoCleanup(t, projectInfo.B, inv_testing.RSRTargetWorkload(workload1))

		dao.CreateSingleScheduleNoCleanup(t, projectInfo.B, inv_testing.SSRRegion(region1))
		dao.CreateSingleScheduleNoCleanup(t, projectInfo.B, inv_testing.SSRTargetSite(site1))
		dao.CreateSingleScheduleNoCleanup(t, projectInfo.B, inv_testing.SSRTargetHost(host1))
		dao.CreateSingleScheduleNoCleanup(t, projectInfo.B, inv_testing.SSRTargetWorkload(workload1))

		verifyIntegrationWithDataModel(t, projectInfo, nxc, ic)
	})
}

func verifyIntegrationWithDataModel(
	t *testing.T, projectInfo *inv_util.Tuple[string, string], nxc *nexus.Client, ic *invclient.TCInventoryClient,
) {
	t.Helper()

	dao := inv_testing.NewInvResourceDAOOrFail(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("AW InProgress", func(t *testing.T) {
		assertActiveWatcherInProgress(t, nxc, projectInfo.A)
	})

	t.Run("tenant exists desired_state eq CREATED", func(t *testing.T) {
		assertTenantExistsDesiredStateCreated(t, ic, projectInfo.B)
	})

	t.Run("providers exist", func(t *testing.T) {
		assertProvidersExist(t, ic, projectInfo.B)
	})

	t.Run("lenovo providers exist", func(t *testing.T) {
		assertLenovoProvidersExist(t, ic, projectInfo.B)
	})

	t.Run("telemetry groups exist", func(t *testing.T) {
		assertTelemetryGroupsExist(t, ic, projectInfo.B)
	})

	t.Run("emulate OSRM Reporting Tenant Watcher Eq true", func(t *testing.T) {
		tenant, err := ic.GetTenantResourceInstance(ctx, projectInfo.B)
		require.NoError(t, err)
		emulateOSRMReportingTenantWatcherEq(t, tenant.GetTenantId(), tenant.GetResourceId(), true)

		// Allow background controller to process the event before checking state
		// This prevents race condition with testify 1.11.0 which checks conditions immediately
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("tenant current_state eq CREATED", func(t *testing.T) {
		assertTenantExistsCurrentStateCreated(t, ic, projectInfo.B)
	})

	t.Run("AW in IDLE", func(t *testing.T) {
		assertActiveWatcherIdle(t, nxc, projectInfo.A)
	})

	t.Run("emulate OSRM Reporting Tenant Watcher Eq false", func(t *testing.T) {
		tenant, err := ic.GetTenantResourceInstance(ctx, projectInfo.B)
		require.NoError(t, err)
		emulateOSRMReportingTenantWatcherEq(t, tenant.GetTenantId(), tenant.GetResourceId(), false)
	})

	t.Run("delete runtime project", func(t *testing.T) {
		deleteRuntimeProject(ctx, t, nxc, projectInfo.A)
	})

	t.Run("emulate CO deleting members/workloads", func(t *testing.T) {
		require.NoError(t, ic.DeleteAllResources(ctx, projectInfo.B,
			inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER, true))
		require.NoError(t, ic.DeleteAllResources(ctx, projectInfo.B,
			inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD, true))
	})

	t.Run("instances desired_state eq deleted", func(t *testing.T) {
		assertInstancesDesiredStateEqDeleted(t, ic, projectInfo.B)
	})

	t.Run("emulate OM deleting instances", func(t *testing.T) {
		require.NoError(t, dao.GetTCClient().
			DeleteAllResources(ctx, projectInfo.B, inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE, true))
	})

	t.Run("hosts desired_state eq deleted", func(t *testing.T) {
		assertHostDesiredStateEqDeleted(t, ic, projectInfo.B)
	})

	t.Run("emulate OM deleting hosts", func(t *testing.T) {
		require.NoError(t, dao.GetTCClient().
			DeleteAllResources(ctx, projectInfo.B, inv_v1.ResourceKind_RESOURCE_KIND_HOST, true))
	})

	t.Run("aw does not exist", func(t *testing.T) {
		assertActiveWatcherDoesNotExist(t, nxc, projectInfo.A)
	})
}

func emulateOSRMReportingTenantWatcherEq(t *testing.T, tid, rid string, watcher bool) {
	t.Helper()
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	ctx := context.TODO()
	_, err := dao.GetRMClient().Update(
		ctx, tid, rid,
		&fieldmaskpb.FieldMask{Paths: []string{tenantv1.TenantFieldWatcherOsmanager}},
		&inv_v1.Resource{Resource: &inv_v1.Resource_Tenant{
			Tenant: &tenantv1.Tenant{WatcherOsmanager: watcher},
		}})
	require.NoError(t, err)
}

func assertActiveWatcherDoesNotExist(t *testing.T, nxc *nexus.Client, projectName string) {
	t.Helper()

	// Wait a bit to allow any in-flight controller operations to complete
	// This helps reduce race conditions between background controllers and test assertions
	time.Sleep(50 * time.Millisecond)

	ctx := context.TODO()
	require.Eventuallyf(t,
		func() bool {
			_, e := getActiveWatcher(ctx, nxc, projectName)
			return nexus_client.IsNotFound(e)
		},
		eventuallyTimeout,
		eventuallyInterval,
		"EXPECTED: ACTIVEWATCHER[%s] shall not exist",
		configuration.AppName,
	)
}

func assertActiveWatcherInProgress(t *testing.T, nxc *nexus.Client, projectName string) {
	t.Helper()

	// Wait a bit to allow any in-flight controller operations to complete
	// This helps reduce race conditions between background controllers and test assertions
	time.Sleep(50 * time.Millisecond)

	ctx := context.TODO()
	require.Eventuallyf(t, func() bool {
		aw, e := getActiveWatcher(ctx, nxc, projectName)
		if e != nil || aw == nil {
			return false
		}
		return baseprojectactivewatcherinfrahostcomv1.StatusIndicationInProgress == aw.Spec.StatusIndicator
	},
		eventuallyTimeout,
		eventuallyInterval,
		"EXPECTED: ACTIVEWATCHER[%s] statusIndicator == IN_PROGRESS",
		configuration.AppName,
	)
}

func assertActiveWatcherIdle(t *testing.T, nxc *nexus.Client, projectName string) {
	t.Helper()

	// Wait a bit to allow any in-flight controller operations to complete
	// This helps reduce race conditions between background controllers and test assertions
	time.Sleep(50 * time.Millisecond)

	require.Eventuallyf(t, func() bool {
		aw, e := getActiveWatcher(context.TODO(), nxc, projectName)
		if e != nil || aw == nil {
			return false
		}
		return baseprojectactivewatcherinfrahostcomv1.StatusIndicationIdle == aw.Spec.StatusIndicator
	},
		eventuallyTimeout,
		eventuallyInterval,
		"EXPECTED: ACTIVEWATCHER[%s] statusIndicator == IDLE",
		configuration.AppName,
	)
}

func assertTenantExistsCurrentStateCreated(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	require.Eventuallyf(t, func() bool {
		tid, rid, err := ic.GetTenantResource(context.TODO(), projectID)
		if err != nil {
			return false
		}
		rsp, e := ic.Get(context.TODO(), tid, rid)
		if e != nil {
			return false
		}
		return tenantv1.TenantState_TENANT_STATE_CREATED == rsp.GetResource().GetTenant().GetCurrentState()
	}, eventuallyTimeout, eventuallyInterval, "EXPECTED: tenant[%s] with current_state == created was expected", projectID)
}

func assertTenantExistsDesiredStateCreated(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	require.Eventuallyf(t, func() bool {
		tenant, e := ic.GetTenantResourceInstance(context.TODO(), projectID)
		if e != nil {
			return false
		}
		return tenantv1.TenantState_TENANT_STATE_CREATED == tenant.GetDesiredState()
	}, eventuallyTimeout, eventuallyInterval,
		"EXPECTED: tenant[%s] with current_state == created was expected", projectID)
}

func assertLenovoProvidersExist(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	provider, err := inv_util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER)
	require.NoError(t, err)

	craftedFilter := fmt.Sprintf("%s = %q AND %s=%s AND %s=%s",
		providerv1.ProviderResourceFieldTenantId, projectID,
		providerv1.ProviderResourceFieldProviderKind, providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL.String(),
		providerv1.ProviderResourceFieldProviderVendor, providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA.String(),
	)
	filter := &inv_v1.ResourceFilter{
		Resource: provider,
		Filter:   craftedFilter,
	}
	lenovoproviders, err := ic.ListAll(context.TODO(), filter)
	require.NoError(t, err)
	require.NotEmptyf(t, lenovoproviders,
		"EXPECTED: Lenovo providers for Tenant[%s] shall be created on INV side", projectID,
	)
}

func assertHostDesiredStateEqDeleted(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	require.Eventually(t, func() bool {
		all, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
			Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
			Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", projectID)).Build(),
		})
		if err != nil {
			return false
		}
		for _, instance := range all {
			if instance.GetHost().DesiredState != computev1.HostState_HOST_STATE_DELETED {
				return false
			}
		}
		return true
	}, eventuallyTimeout, eventuallyInterval)
}

func assertInstancesDesiredStateEqDeleted(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	require.Eventually(t, func() bool {
		all, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
			Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Instance{}},
			Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", projectID)).Build(),
		})
		if err != nil {
			return false
		}
		for _, instance := range all {
			if instance.GetInstance().DesiredState != computev1.InstanceState_INSTANCE_STATE_DELETED {
				return false
			}
		}
		return true
	}, eventuallyTimeout, eventuallyInterval)
}

func assertProvidersExist(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	provider, err := inv_util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER)
	require.NoError(t, err)
	providers, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
		Resource: provider,
		Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", projectID)).Build(),
	})
	require.NoError(t, err)
	require.NotEmptyf(t, providers, "EXPECTED: providers for Tenant[%s] shall be created on INV side", projectID)
}

func assertTelemetryGroupsExist(t *testing.T, ic *invclient.TCInventoryClient, projectID string) {
	t.Helper()
	telemetryGroup, err := inv_util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP)
	require.NoError(t, err)
	tgs, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
		Resource: telemetryGroup,
		Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", projectID)).Build(),
	})
	require.NoError(t, err)
	require.NotEmptyf(t, tgs, "EXPECTED: Telemetry Groups for Tenant[%s] shall be created on INV side", projectID)
}

func getActiveWatcher(
	ctx context.Context, nxc *nexus.Client, projectName string,
) (*nexus_client.ProjectactivewatcherProjectActiveWatcher, error) {
	return nxc.TenancyMultiTenancy().
		Runtime().
		Orgs(org1Name).
		Folders(org1FolderName).
		Projects(projectName).
		GetActiveWatchers(ctx, configuration.AppName)
}

func createOrg(t *testing.T, nxc *nexus.Client) {
	t.Helper()

	// create mt
	tenancyClient, err := nxc.AddTenancyMultiTenancy(
		context.Background(), &tenancyv1.MultiTenancy{ObjectMeta: metav1.ObjectMeta{Name: mtName, ResourceVersion: "1"}},
	)
	require.NoError(t, err)

	/////////////////////
	// config tree
	/////////////////////
	// create config tree
	configClient, err := tenancyClient.AddConfig(
		context.Background(),
		&baseconfiginfrahostcomv1.Config{ObjectMeta: metav1.ObjectMeta{Name: configName, ResourceVersion: "1"}},
	)
	require.NoError(t, err)
	// create config org
	orgConfigClient, err := configClient.AddOrgs(
		context.TODO(),
		&baseorginfrahostcomv1.Org{ObjectMeta: metav1.ObjectMeta{Name: org1Name, ResourceVersion: "1"}})
	require.NoError(t, err)

	require.NoError(t, err)
	// create config folder
	_, err = orgConfigClient.AddFolders(
		context.TODO(),
		&basefolderinfrahostcomv1.Folder{ObjectMeta: metav1.ObjectMeta{Name: org1FolderName, ResourceVersion: "1"}})
	require.NoError(t, err)
	// end of config tree /////////////////

	///////////////////////
	// runtime tree
	//////////////////////
	// create runtime tree
	runtimeClient, err := tenancyClient.AddRuntime(
		context.Background(),
		&baseruntimeinfrahostcomv1.Runtime{ObjectMeta: metav1.ObjectMeta{Name: runtimeName, ResourceVersion: "1"}},
	)
	require.NoError(t, err)
	// create runtime org
	orgClient, err := runtimeClient.AddOrgs(
		context.TODO(),
		&baseruntimeorginfrahostcomv1.RuntimeOrg{ObjectMeta: metav1.ObjectMeta{Name: org1Name, ResourceVersion: "1"}})
	require.NoError(t, err)
	// create runtime folder
	_, err = orgClient.AddFolders(
		context.TODO(),
		&baseruntimefolderinfrahostcomv1.RuntimeFolder{
			ObjectMeta: metav1.ObjectMeta{Name: org1FolderName, ResourceVersion: "1"},
		},
	)
	require.NoError(t, err)
	// end of runtime tree /////////////////////
}

func deleteRuntimeProject(ctx context.Context, t *testing.T, nxc *nexus.Client, projectName string) {
	t.Helper()

	runtimeProject, err := nxc.TenancyMultiTenancy().Runtime().Orgs(org1Name).Folders(org1FolderName).
		GetProjects(context.TODO(), projectName)
	require.NoError(t, err)
	runtimeProject.Spec.Deleted = true
	require.NoError(t, runtimeProject.Update(ctx))
}

func createProjectUnderDefaultOrg(
	t *testing.T, nxc *nexus.Client, projectInfo *inv_util.Tuple[string, string],
) *nexus_client.RuntimeprojectRuntimeProject {
	t.Helper()
	_, err := nxc.TenancyMultiTenancy().Config().Orgs(org1Name).Folders(org1FolderName).
		AddProjects(context.TODO(), &baseprojectinfrahostcomv1.Project{
			ObjectMeta: metav1.ObjectMeta{Name: projectInfo.A, UID: types.UID(projectInfo.B), ResourceVersion: "1"},
		})
	require.NoError(t, err)
	rp, err := nxc.TenancyMultiTenancy().Runtime().Orgs(org1Name).Folders(org1FolderName).
		AddProjects(context.TODO(), &baseruntimeprojectinfrahostcomv1.RuntimeProject{
			ObjectMeta: metav1.ObjectMeta{Name: projectInfo.A, UID: types.UID(projectInfo.B), ResourceVersion: "1"},
		})
	require.NoError(t, err)

	return rp
}

func TestOnCreateOnUpdateHooks(t *testing.T) {
	tcs := []struct {
		name                  string
		rp                    *nexus_client.RuntimeprojectRuntimeProject
		assertInitializations func(*testing.T, chan *nexus_client.RuntimeprojectRuntimeProject)
		assertTerminations    func(*testing.T, chan *nexus_client.RuntimeprojectRuntimeProject)
	}{
		{
			name: "initialization",
			rp: &nexus_client.RuntimeprojectRuntimeProject{
				RuntimeProject: &baseruntimeprojectinfrahostcomv1.RuntimeProject{
					ObjectMeta: metav1.ObjectMeta{
						UID: types.UID(uuid.NewString()),
					},
					Spec: baseruntimeprojectinfrahostcomv1.RuntimeProjectSpec{},
				},
			},
			assertInitializations: func(t *testing.T, projects chan *nexus_client.RuntimeprojectRuntimeProject) {
				t.Helper()
				require.NotEmpty(t, projects)
			},
			assertTerminations: func(t *testing.T, projects chan *nexus_client.RuntimeprojectRuntimeProject) {
				t.Helper()
				require.Empty(t, projects)
			},
		},
		{
			name: "termination",
			rp: &nexus_client.RuntimeprojectRuntimeProject{
				RuntimeProject: &baseruntimeprojectinfrahostcomv1.RuntimeProject{
					ObjectMeta: metav1.ObjectMeta{
						UID: types.UID(uuid.NewString()),
					},
					Spec: baseruntimeprojectinfrahostcomv1.RuntimeProjectSpec{
						Deleted: true,
					},
				},
			},
			assertInitializations: func(t *testing.T, projects chan *nexus_client.RuntimeprojectRuntimeProject) {
				t.Helper()
				require.Empty(t, projects)
			},
			assertTerminations: func(t *testing.T, projects chan *nexus_client.RuntimeprojectRuntimeProject) {
				t.Helper()
				require.NotEmpty(t, projects)
			},
		},
	}

	for _, tc := range tcs {
		t.Run("OnCreate "+tc.name, func(t *testing.T) {
			termChan := make(chan bool)
			defer close(termChan)

			initializations := make(chan *nexus_client.RuntimeprojectRuntimeProject, 1)
			terminations := make(chan *nexus_client.RuntimeprojectRuntimeProject, 1)

			dmc := Controller{
				projectsToBeInitialized: initializations,
				projectsToBeTerminated:  terminations,
			}
			dmc.onCreate()(tc.rp)

			tc.assertInitializations(t, initializations)
			tc.assertTerminations(t, terminations)
		})
	}

	for _, tc := range tcs {
		t.Run("OnUpdate "+tc.name, func(t *testing.T) {
			termChan := make(chan bool)
			defer close(termChan)

			initializations := make(chan *nexus_client.RuntimeprojectRuntimeProject, 1)
			terminations := make(chan *nexus_client.RuntimeprojectRuntimeProject, 1)

			dmc := Controller{
				projectsToBeInitialized: initializations,
				projectsToBeTerminated:  terminations,
			}
			dmc.onUpdate()(nil, tc.rp)

			tc.assertInitializations(t, initializations)
			tc.assertTerminations(t, terminations)
		})
	}
}
