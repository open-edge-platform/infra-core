// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostgpuresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostnicresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hoststorageresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostusbresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/localaccountresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/operatingsystemresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/ouresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/providerresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	res_status "github.com/open-edge-platform/infra-core/inventory/v2/pkg/status"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
)

func Test_Metadata_Inheritance_Host(t *testing.T) {
	// Create required Regions, OUs and Sites
	region0 := inv_testing.CreateRegion(t, nil)
	region1 := inv_testing.CreateRegionWithMeta(t, metaR1, nil)
	region2 := inv_testing.CreateRegionWithMeta(t, metaR2, region1)
	region3 := inv_testing.CreateRegionWithMeta(t, metaR3, region2)
	region4 := inv_testing.CreateRegion(t, region3)

	ou0 := inv_testing.CreateOu(t, nil)
	ou1 := inv_testing.CreateOuWithMeta(t, metaO1, nil)
	ou2 := inv_testing.CreateOuWithMeta(t, metaO2, ou1)
	ou3 := inv_testing.CreateOuWithMeta(t, metaO3, ou2)
	ou4 := inv_testing.CreateOu(t, ou3)
	ou5 := inv_testing.CreateOuWithMeta(t, metaO6, ou3)

	site0 := inv_testing.CreateSite(t, region0, ou0)
	siteOnlyPhy := inv_testing.CreateSite(t, region4, nil)
	siteOnlyLogi := inv_testing.CreateSite(t, nil, ou4)
	siteBoth0 := inv_testing.CreateSite(t, region4, ou4)
	siteBoth1 := inv_testing.CreateSite(t, region4, ou5)
	siteBoth2 := inv_testing.CreateSiteWithMeta(t, metaO5, region4, ou5)
	siteNoHiera := inv_testing.CreateSiteWithMeta(t, metaO6, nil, nil)

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	testcases := map[string]struct {
		in          *computev1.HostResource
		expPhyMeta  *string
		expLogiMeta *string
	}{
		"NoParentSite": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &emptyString,
			expLogiMeta: &emptyString,
		},
		"NoMetadataFromParent1": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: site0,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &emptyString,
			expLogiMeta: &emptyString,
		},
		"MetadataFromSiteOnly": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: siteNoHiera,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &metaO6,
			expLogiMeta: &emptyString,
		},
		"NoMetadataFromParent2": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &emptyString,
			expLogiMeta: &emptyString,
		},
		"PhyMetadataFromParent": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: siteOnlyPhy,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &expPhyMeta1,
			expLogiMeta: nil,
		},
		"LogiMetadataFromParent": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: siteOnlyLogi,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  nil,
			expLogiMeta: &expLogiMeta1,
		},
		"PhyMetadataOverrideFromPhy": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: siteBoth0,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &expPhyMeta1,
			expLogiMeta: &emptyString,
		},
		"PhyMetadataOverrideFromOu": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: siteBoth1,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &expPhyMeta1,
			expLogiMeta: &metaO6,
		},
		"PhyMetadataOverrideFromSite": {
			in: &computev1.HostResource{
				Name: "Test Host 1",
				Site: siteBoth2,
				Uuid: uuid.NewString(),
			},
			expPhyMeta:  &expPhyMeta3,
			expLogiMeta: &metaO6,
		},
		"InheritMetadataFromOuParentAndLocal": {
			in: &computev1.HostResource{
				Name:     "Test Host 1",
				Site:     siteOnlyLogi,
				Metadata: metaO5,
				Uuid:     uuid.NewString(),
			},
			expPhyMeta:  nil,
			expLogiMeta: &expLogiMeta2,
		},
		"InheritMetadataFromRegionParentAndLocal": {
			in: &computev1.HostResource{
				Name:     "Test Host 1",
				Site:     siteOnlyPhy,
				Metadata: metaR5,
				Uuid:     uuid.NewString(),
			},
			expPhyMeta:  &expPhyMeta2,
			expLogiMeta: nil,
		},
		"InheritMetadataFromRegionAndOuParentAndLocal": {
			in: &computev1.HostResource{
				Name:     "Test Host 1",
				Site:     siteBoth0,
				Metadata: metaR5,
				Uuid:     uuid.NewString(),
			},
			expPhyMeta:  &expPhyMeta2,
			expLogiMeta: &emptyString,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: tc.in},
			}
			chostResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			require.NoError(t, err, "CreateHost() failed")
			hostResID := inv_testing.GetResourceIDOrFail(t, chostResp)
			t.Cleanup(func() { inv_testing.HardDeleteHost(t, hostResID) })

			t.Run("Check with GET", func(t *testing.T) {
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, hostResID)
				require.NoError(t, err, "GetHost() failed")
				if tc.expPhyMeta != nil && !CompareMetadata(t, getresp.RenderedMetadata.PhyMetadata, *tc.expPhyMeta) {
					t.Errorf("Physical Metadata data not equal - want: %s, got: %s",
						*tc.expPhyMeta, getresp.RenderedMetadata.PhyMetadata,
					)
				}
				if tc.expLogiMeta != nil && !CompareMetadata(t, getresp.RenderedMetadata.LogiMetadata, *tc.expLogiMeta) {
					t.Errorf("Logical Metadata data not equal - want: %s, got: %s",
						*tc.expLogiMeta, getresp.RenderedMetadata.LogiMetadata,
					)
				}
			})
			t.Run("Check with LIST", func(t *testing.T) {
				listResp, err := inv_testing.TestClients[inv_testing.APIClient].
					List(ctx, &inv_v1.ResourceFilter{
						Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
						Filter: filters.NewBuilderWith(filters.ValEq("tenant_id", client.FakeTenantID)).
							Build(),
					})
				require.NoError(t, err, "GetHost() failed")
				require.NotEmpty(t, listResp.GetResources())
				resource := collections.Filter(listResp.GetResources(), func(v *inv_v1.GetResourceResponse) bool {
					return v.GetResource().GetHost().GetResourceId() == hostResID
				})[0]

				if tc.expPhyMeta != nil && !CompareMetadata(t, resource.RenderedMetadata.PhyMetadata, *tc.expPhyMeta) {
					t.Errorf("Physical Metadata data not equal - want: %s, got: %s",
						*tc.expPhyMeta, resource.RenderedMetadata.PhyMetadata,
					)
				}
				if tc.expLogiMeta != nil && !CompareMetadata(t, resource.RenderedMetadata.LogiMetadata, *tc.expLogiMeta) {
					t.Errorf("Logical Metadata data not equal - want: %s, got: %s",
						*tc.expLogiMeta, resource.RenderedMetadata.LogiMetadata,
					)
				}
			})
		})
	}
}

func Test_Metadata_Inheritance_Host_Filter(t *testing.T) {
	// Create required Regions, OUs and Sites
	region1 := inv_testing.CreateRegionWithMeta(t, metaR1, nil)
	region2 := inv_testing.CreateRegionWithMeta(t, metaR2, region1)

	ou1 := inv_testing.CreateOuWithMeta(t, metaO1, nil)
	ou2 := inv_testing.CreateOuWithMeta(t, metaO2, ou1)

	siteOnlyPhy := inv_testing.CreateSite(t, region2, nil)
	siteOnlyLogi := inv_testing.CreateSite(t, nil, ou2)
	siteBoth1 := inv_testing.CreateSite(t, region1, ou1)
	siteBoth2 := inv_testing.CreateSiteWithMeta(t, metaO5, region1, ou1)

	h1 := inv_testing.CreateHost(t, siteOnlyPhy, nil)
	h2 := inv_testing.CreateHost(t, siteOnlyLogi, nil)
	h3 := inv_testing.CreateHost(t, siteBoth1, nil)
	h4 := inv_testing.CreateHost(t, siteBoth2, nil)

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	testcases := map[string]struct {
		in          *inv_v1.ResourceFilter
		expectedIDs []*client.ResourceTenantIDCarrier
	}{
		"FilterInheritedOnlyPhy": {
			in: &inv_v1.ResourceFilter{
				Filter: "", // Special case filtering on inherited metadata
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							Metadata: `[{"key":"key1-test","value":"region_key1_lvl2-test"}]`,
						},
					},
				},
			},
			expectedIDs: []*client.ResourceTenantIDCarrier{{TenantId: h1.TenantId, ResourceId: h1.ResourceId}},
		},
		"FilterInheritedOnlyLogi1": {
			in: &inv_v1.ResourceFilter{
				Filter: "", // Special case filtering on inherited metadata
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							Metadata: `[{"key":"key1-test","value":"ou_key1_lvl2-test"}]`,
						},
					},
				},
			},
			expectedIDs: []*client.ResourceTenantIDCarrier{{TenantId: h2.TenantId, ResourceId: h2.ResourceId}},
		},
		"FilterInheritedOnlyLogi2": {
			in: &inv_v1.ResourceFilter{
				Filter: "", // Special case filtering on inherited metadata
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							Metadata: `[{"key":"key3-test", "value":"ou_key3_lvl1-test"}]`,
						},
					},
				},
			},
			expectedIDs: []*client.ResourceTenantIDCarrier{{TenantId: h2.TenantId, ResourceId: h2.ResourceId}},
		},
		"FilterInheritedBoth1": {
			in: &inv_v1.ResourceFilter{
				Filter: "", // Special case filtering on inherited metadata
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							Metadata: `[{"key":"key1-test", "value":"region_key1_lvl1-test"}]`,
						},
					},
				},
			},
			expectedIDs: []*client.ResourceTenantIDCarrier{{TenantId: h3.TenantId, ResourceId: h3.ResourceId}},
		},
		"FilterInheritedBoth2": {
			in: &inv_v1.ResourceFilter{
				Filter: "", // Special case filtering on inherited metadata
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							Metadata: `[{"key":"key1-test", "value":"ou_key1_lvl4-test"}]`,
						},
					},
				},
			},
			expectedIDs: []*client.ResourceTenantIDCarrier{{TenantId: h4.TenantId, ResourceId: h4.ResourceId}},
		},
		"FilterInheritedSameRoot": {
			in: &inv_v1.ResourceFilter{
				Filter: "", // Special case filtering on inherited metadata
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							Metadata: `[{"key":"key3-test", "value":"region_key3_lvl1-test"}]`,
						},
					},
				},
			},
			expectedIDs: []*client.ResourceTenantIDCarrier{
				{TenantId: h1.TenantId, ResourceId: h1.ResourceId},
				{TenantId: h3.TenantId, ResourceId: h3.ResourceId},
				{TenantId: h4.TenantId, ResourceId: h4.ResourceId},
			},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			findRes, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)
			require.NoError(t, err, "FilterHosts() failed %s", err)
			inv_testing.SortHasResourceIDAndTenantID(tc.expectedIDs)
			inv_testing.SortHasResourceIDAndTenantID(findRes.Resources)
			if !reflect.DeepEqual(tc.expectedIDs, findRes.Resources) {
				t.Fatalf(
					"FilterHosts() failed - want: %s, got: %s",
					tc.expectedIDs,
					findRes.Resources,
				)
			}
		})
	}
}

func Test_Host_BackReferences_Read(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)
	storage := inv_testing.CreateHostStorage(t, host)
	nic := inv_testing.CreateHostNic(t, host)
	usb := inv_testing.CreateHostusb(t, host)
	os := inv_testing.CreateOs(t)
	instance := inv_testing.CreateInstance(t, host, os)
	instance.DesiredOs = os
	instance.CurrentOs = os
	instance.Os = os
	host.Instance = instance
	gpu := inv_testing.CreatHostGPU(t, host)
	host.HostStorages = append(host.HostStorages, storage)
	host.HostNics = append(host.HostNics, nic)
	host.HostUsbs = append(host.HostUsbs, usb)
	host.HostGpus = append(host.HostGpus, gpu)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Fetch the host and check that all back-references are present.
	resp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, host.ResourceId)
	require.NoError(t, err)
	eq, diff := inv_testing.ProtoEqualOrDiff(host, resp.GetResource().GetHost())
	require.Truef(t, eq, "host storage back-ref not equal %v", diff)
}

func Test_Create_Get_Delete_Host(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site := inv_testing.CreateSite(t, region, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")

	testcases := map[string]struct {
		in         *computev1.HostResource
		clientName inv_testing.ClientType
		valid      bool
	}{
		"CreateGoodHost": {
			in: &computev1.HostResource{
				Name:         "Test Host 1",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
				CurrentState: computev1.HostState_HOST_STATE_UNSPECIFIED,

				Site:     site,
				Provider: provider,

				HardwareKind: "XDgen2",
				SerialNumber: "12345678",
				Uuid:         "E5E53D99-708D-4AF5-8378-63880FF62712",
				MemoryBytes:  64 * util.Gigabyte,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      20,
				CpuTopology:     `{"some_json":[]}`,

				MgmtIp: "192.168.10.10",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.10",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:ff",

				Hostname:        "testhost1",
				ProductName:     "PowerEdge R750",
				BiosVersion:     "1.0.0",
				BiosReleaseDate: "09/14/2022",
				BiosVendor:      "Dell Inc.",

				DesiredPowerState: computev1.PowerState_POWER_STATE_ON,

				DesiredAmtState: computev1.AmtState_AMT_STATE_PROVISIONED,

				Metadata: "[{\"key\":\"cluster-name\",\"value\":\"\"},{\"key\":\"app-id\",\"value\":\"\"}]",
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		"CreateGoodHost2": {
			in: &computev1.HostResource{
				Name:         "Test Host 1",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				CurrentState: computev1.HostState_HOST_STATE_UNSPECIFIED,

				Site:     site,
				Provider: provider,

				HardwareKind: "XDgen2",
				SerialNumber: "12345678",
				Uuid:         "E5E53D99-708D-4AF5-8378-63880FF62712",
				MemoryBytes:  64 * util.Gigabyte,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      20,
				CpuTopology:     `{"some_json":[]}`,

				MgmtIp: "192.168.10.10",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.10",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:ff",

				Hostname:        "testhost1",
				ProductName:     "PowerEdge R750",
				BiosVersion:     "1.0.0",
				BiosReleaseDate: "09/14/2022",
				BiosVendor:      "Dell Inc.",

				DesiredPowerState: computev1.PowerState_POWER_STATE_ON,

				Metadata: "[{\"key\":\"cluster-name\",\"value\":\"\"},{\"key\":\"app-id\",\"value\":\"\"}]",
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		"CreateGoodEmptyHost": {
			in: &computev1.HostResource{
				Uuid: uuid.NewString(),
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		"CreateBadHostWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.HostResource{
				ResourceId: "host-12345678",
				Uuid:       uuid.NewString(),
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		"CreateBadHostWithInvalidResourceIdSet": {
			// This tests case verifies that create requests with a invalid resource ID
			// already set are rejected.
			in: &computev1.HostResource{
				ResourceId: "ho-12345678",
				Uuid:       uuid.NewString(),
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		"CreateBadHostNonExistingSite": {
			// This tests case verifies that hosts must point to a valid
			// existing site.
			in: &computev1.HostResource{
				Uuid: uuid.NewString(),
				Site: &location_v1.SiteResource{
					ResourceId: "site-12345678",
				},
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		"CreateBadHostWitCurrentState": {
			// This tests case verifies that a host cannot be created with a current state from API client
			in: &computev1.HostResource{
				Uuid:         uuid.NewString(),
				CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		// TODO: negative tests should be added with clientKind API
		"CreateFromRM": {
			// This tests case verifies that a host can be created from RM with RM-updatable fields
			in: &computev1.HostResource{
				Uuid:         uuid.NewString(),
				CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,

				CurrentPowerState:    computev1.PowerState_POWER_STATE_OFF,
				PowerStatus:          "Powered On",
				PowerStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				PowerStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test

				CurrentAmtState:    computev1.AmtState_AMT_STATE_UNPROVISIONED,
				AmtStatus:          "Provisioned",
				AmtStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				AmtStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
				AmtSku:             "vPRO Corporate 16.5.2",
			},
			clientName: inv_testing.RMClient,
			valid:      true,
		},
		"CreateBadMetadataHost": {
			in: &computev1.HostResource{
				Uuid:     uuid.NewString(),
				Metadata: metaDuplicatedKeys,
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create host
			chostResp, err := inv_testing.TestClients[tc.clientName].Create(ctx, createresreq)
			hostResID := chostResp.GetHost().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateHost() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = hostResID // Update with created resource ID.
				tc.in.HostStatus = res_status.DefaultHostStatus
				tc.in.OnboardingStatus = res_status.DefaultOnboardingStatus
				tc.in.RegistrationStatus = res_status.DefaultRegistrationStatus
				tc.in.CreatedAt = chostResp.GetHost().GetCreatedAt()
				tc.in.UpdatedAt = chostResp.GetHost().GetUpdatedAt()
				if tc.in.AmtSku == "" {
					tc.in.AmtSku = "Unknown"
				}
				if tc.in.AmtStatus == "" {
					tc.in.AmtStatus = res_status.DefaultAmtStatus
				}
				if tc.in.PowerStatus == "" {
					tc.in.PowerStatus = res_status.DefaultPowerStatus
				}
				assertSameResource(t, createresreq, chostResp, nil)
				if !tc.valid {
					t.Errorf("CreateHost() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "host-12345678")
				require.Error(t, err)

				// get host
				getresp, err := inv_testing.TestClients[tc.clientName].Get(ctx, hostResID)
				require.NoError(t, err, "GetHost() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetHost()); !eq {
					t.Errorf("GetHost() data not equal: %v", diff)
				}

				// delete non-existent first
				err = inv_testing.DeleteResourceAndReturnError(t, "host-12345678")
				require.Error(t, err)

				// Remove host.
				inv_testing.HardDeleteHost(t, hostResID)

				// get after complete Delete of host, should fail as Host is 2-phase deleted
				_, err = inv_testing.TestClients[inv_testing.RMClient].Get(ctx, hostResID)
				require.Error(t, err, "Failure - Host was not deleted, but should be deleted")
			}
		})
	}
}

func Test_UpdateHost(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site1 := inv_testing.CreateSite(t, region, nil)
	site2 := inv_testing.CreateSite(t, region, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host1 := inv_testing.CreateHost(t, nil, nil)

	// create Host to update
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 2",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,

				Site:     site1,
				Provider: provider,

				HardwareKind: "XDgen2",
				SerialNumber: "12345679",
				Uuid:         "E5E53D99-708D-4AF5-8378-63880FF62712",
				MemoryBytes:  64 * util.Gigabyte,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      2,
				CpuTopology:     "{\"some_json\":[]}",

				MgmtIp: "192.168.10.11",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.11",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:f0",

				Hostname:        "testhost2",
				ProductName:     "PowerEdge R750",
				BiosVersion:     "1.0.0",
				BiosReleaseDate: "09/14/2022",
				BiosVendor:      "Dell Inc.",

				DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
			},
		},
	}

	putHost := computev1.HostResource{
		ResourceId:   host1.ResourceId,
		Name:         "TEST",
		DesiredState: computev1.HostState_HOST_STATE_UNSPECIFIED,
		// CurrentState: provided by SB
		HostStatus:                  "some host status",
		HostStatusIndicator:         statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		HostStatusTimestamp:         uint64(time.Now().Unix()), //nolint:gosec // This is a test
		RegistrationStatus:          "some registration status",
		RegistrationStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED,
		RegistrationStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
		OnboardingStatus:            "some onboarding status",
		OnboardingStatusIndicator:   statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED,
		OnboardingStatusTimestamp:   uint64(time.Now().Unix()), //nolint:gosec // This is a test
		Site:                        site1,
		Provider:                    provider,
		HardwareKind:                "TEST",
		SerialNumber:                "TEST",
		Uuid:                        uuid.NewString(),
		MemoryBytes:                 1,
		CpuModel:                    "TEST",
		CpuSockets:                  1,
		CpuCores:                    1,
		CpuCapabilities:             "TEST",
		CpuThreads:                  1,
		MgmtIp:                      "192.168.1.1",
		BmcIp:                       "192.168.1.2",
		BmcUsername:                 "TEST",
		BmcPassword:                 "TEST",
		PxeMac:                      "TEST",
		Hostname:                    "TEST",
		Metadata:                    metaHost1,
		DesiredPowerState:           computev1.PowerState_POWER_STATE_OFF,
		// CurrentPowerState: should be provided by SB
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chostResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
	require.NoError(t, err, "CreateHost() failed")
	hostResID := inv_testing.GetResourceIDOrFail(t, chostResp)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, hostResID) })

	testcases := []struct {
		name         string
		in           *computev1.HostResource
		resourceID   string
		clientName   inv_testing.ClientType
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		{
			name:         "UpdatePut",
			in:           &putHost,
			resourceID:   host1.ResourceId,
			clientName:   inv_testing.APIClient,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		{
			name: "UpdateHost1",
			in: &computev1.HostResource{
				CpuCores:          8,
				BmcIp:             "10.11.12.14",
				CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
			},
			resourceID: hostResID,
			clientName: inv_testing.RMClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostresource.FieldCPUCores, hostresource.FieldBmcIP,
					hostresource.FieldDesiredState, hostresource.FieldCurrentPowerState,
					hostresource.FieldHostStatus,
				},
			},
			valid: true,
		},
		{
			name:       "UpdateHost2",
			in:         &computev1.HostResource{},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostresource.FieldBmcIP},
			},
			valid: true,
		},
		{
			name: "UpdateHost3",
			in: &computev1.HostResource{
				HostStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{computev1.HostResourceFieldHostStatusIndicator},
			},
			valid: true,
		},
		{
			name: "UpdateHost4",
			in: &computev1.HostResource{
				CurrentState: computev1.HostState_HOST_STATE_UNSPECIFIED,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostresource.FieldCurrentState},
			},
			valid: true,
		},
		{
			name: "UpdateHost5",
			in: &computev1.HostResource{
				CpuCores:     10,
				BmcIp:        "10.11.12.15",
				DesiredState: computev1.HostState_HOST_STATE_UNSPECIFIED,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostresource.FieldCPUCores, hostresource.FieldBmcIP,
					hostresource.FieldBmcKind, hostresource.FieldDesiredState,
				},
			},
			valid: true,
		},
		{
			name: "UpdateHostNewSite",
			in: &computev1.HostResource{
				Site: site2,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{hostresource.EdgeSite}},
			valid:      true,
		},
		{
			name: "UpdateHostNonExistingSite",
			in: &computev1.HostResource{
				Site: &location_v1.SiteResource{
					ResourceId: "site-12345678",
				},
			},
			resourceID:   hostResID,
			clientName:   inv_testing.APIClient,
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{hostresource.EdgeSite}},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
		{
			name:       "UpdateHostNoSite",
			in:         &computev1.HostResource{},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{hostresource.EdgeSite}},
			valid:      true,
		},
		{
			name: "UpdateHostUUID",
			in: &computev1.HostResource{
				Uuid: uuid.NewString(),
			},
			resourceID:   hostResID,
			clientName:   inv_testing.APIClient,
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldUUID}},
			valid:        false,
			expErrorCode: codes.PermissionDenied,
		},
		{
			name: "UpdateHostSerialNumber",
			in: &computev1.HostResource{
				SerialNumber: "11111111",
			},
			resourceID:   hostResID,
			clientName:   inv_testing.APIClient,
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldSerialNumber}},
			valid:        false,
			expErrorCode: codes.PermissionDenied,
		},
		{
			name: "UpdateHostUUIDFromRM",
			in: &computev1.HostResource{
				Uuid: uuid.NewString(),
			},
			resourceID: hostResID,
			clientName: inv_testing.RMClient,
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldUUID}},
			valid:      true,
		},
		{
			name: "UpdateHostSerialNumberFromRM",
			in: &computev1.HostResource{
				SerialNumber: "11111111",
			},
			resourceID: hostResID,
			clientName: inv_testing.RMClient,
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldSerialNumber}},
			valid:      true,
		},
		{
			name: "UpdateHostInvalidFieldMask",
			in: &computev1.HostResource{
				CpuCores: 12,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		{
			name: "UpdateHostWithUnwantedCurrentState",
			// with API client, this should fail
			in: &computev1.HostResource{
				CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostresource.FieldCurrentState},
			},
			valid:        false,
			expErrorCode: codes.PermissionDenied,
		},
		{
			name: "UpdateHostFromRM",
			in: &computev1.HostResource{
				CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
				CurrentPowerState: computev1.PowerState_POWER_STATE_OFF,
			},
			resourceID: hostResID,
			clientName: inv_testing.RMClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostresource.FieldCurrentState,
					hostresource.FieldCurrentPowerState,
				},
			},
			valid: true,
		},
		{
			name: "UpdateHostFromAPI_Fail",
			in: &computev1.HostResource{
				CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
				CurrentPowerState: computev1.PowerState_POWER_STATE_OFF,
			},
			resourceID: hostResID,
			clientName: inv_testing.APIClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostresource.FieldCurrentState,
					hostresource.FieldCurrentPowerState,
				},
			},
			valid:        false,
			expErrorCode: codes.PermissionDenied,
		},
		{
			name: "UpdateBadMetadataHost",
			in: &computev1.HostResource{
				Metadata: metaDuplicatedKeys,
			},
			resourceID:   hostResID,
			clientName:   inv_testing.APIClient,
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldMetadata}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		{
			name: "UpdateResourceIDNotFound",
			in: &computev1.HostResource{
				CpuCores:          8,
				BmcIp:             "10.11.12.14",
				CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
			},
			resourceID: "host-12345678",
			clientName: inv_testing.RMClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostresource.FieldCPUCores, hostresource.FieldBmcIP,
					hostresource.FieldDesiredState, hostresource.FieldCurrentPowerState,
					hostresource.FieldHostStatus,
				},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
		{
			name: "UpdateAmtInfoPowerInfoFromRM",
			in: &computev1.HostResource{
				CpuCores: 8,
				BmcIp:    "10.11.12.14",

				CurrentPowerState:    computev1.PowerState_POWER_STATE_ON,
				PowerStatus:          "Powered On",
				PowerStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				PowerStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test

				CurrentAmtState:    computev1.AmtState_AMT_STATE_UNPROVISIONED,
				AmtStatus:          "Provisioned",
				AmtStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				AmtStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
				AmtSku:             "vPRO Corporate 16.5.2",
			},
			resourceID: hostResID,
			clientName: inv_testing.RMClient,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostresource.FieldCPUCores, hostresource.FieldBmcIP,
					hostresource.FieldCurrentPowerState, hostresource.FieldPowerStatus,
					hostresource.FieldPowerStatusIndicator, hostresource.FieldPowerStatusTimestamp,
					hostresource.FieldCurrentAmtState, hostresource.FieldAmtStatus,
					hostresource.FieldAmtStatusIndicator, hostresource.FieldAmtStatusTimestamp,
					hostresource.FieldAmtSku,
				},
			},
			valid: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: tc.in},
			}
			upRes, err := inv_testing.TestClients[tc.clientName].Update(ctx, tc.resourceID, tc.fieldMask, updateresreq)

			if !tc.valid {
				require.Errorf(t, err, "UpdateResource() succeeded but should have failed")
				assert.Equal(t, tc.expErrorCode, status.Code(err))
				assert.Nil(t, upRes)
				return
			}
			require.NoErrorf(t, err, "UpdateResource() failed: %s", err)

			// Validate returned resource
			assertSameResource(t, updateresreq, upRes, tc.fieldMask)

			// validate update via a get
			var getresp *inv_v1.GetResourceResponse
			getresp, err = inv_testing.TestClients[tc.clientName].Get(ctx, tc.resourceID)
			require.NoError(t, err, "GetHost() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

func Test_Register_Host(t *testing.T) {
	var registeredHostIDs []string
	defaultUUIDAPI := "E5E53D99-708D-4AF5-8378-63880FF62712"
	defaultUUIDRM := "E5E53D99-708D-4AF5-8378-63880FF62714"

	testcases := []struct {
		name       string
		in         *computev1.HostResource
		clientName inv_testing.ClientType
		valid      bool
	}{
		{
			name: "RegisterGoodHostUniqueUUIDUniqueSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
				Uuid:         defaultUUIDAPI,
				SerialNumber: "SN000001",
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		{
			name: "RegisterGoodHostUniqueUUIDNoSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				Uuid:         uuid.New().String(),
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		{
			name: "RegisterGoodHostNoUUIDUniqueSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				SerialNumber: "SN000002",
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		{
			name: "RegisterGoodHostUniqueUUIDDuplicateSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				Uuid:         uuid.New().String(),
				SerialNumber: "SN000001",
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
		{
			name: "RegisterInvalidHostDuplicateUUIDNoSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				Uuid:         defaultUUIDAPI,
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		{
			name: "RegisterInvalidHostNoUUIDDuplicateSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				SerialNumber: "SN000002",
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		{
			name: "RegisterInvalidHostDuplicateUUIDUniqueSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
				Uuid:         defaultUUIDAPI,
				SerialNumber: "SN000004",
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		{
			name: "RegisterInvalidHostNoUUIDNoSN",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
			},
			clientName: inv_testing.APIClient,
			valid:      false,
		},
		{
			name: "RegisterGoodHostUniqueUUIDUniqueSNFromRM",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				Uuid:         defaultUUIDRM,
				SerialNumber: "SN100001",
			},
			clientName: inv_testing.RMClient,
			valid:      true,
		},
		{
			name: "RegisterGoodHostUniqueUUIDNoSNFromRM",
			in: &computev1.HostResource{
				Name: "Test Host Register",
				Uuid: uuid.New().String(),
			},
			clientName: inv_testing.RMClient,
			valid:      true,
		},
		{
			name: "RegisterGoodHostNoUUIDUniqueSNFromRM",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				SerialNumber: "SN100002",
			},
			clientName: inv_testing.RMClient,
			valid:      true,
		},
		{
			name: "RegisterGoodHostUniqueUUIDDuplicateSNFromRM",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				Uuid:         uuid.New().String(),
				SerialNumber: "SN100001",
			},
			clientName: inv_testing.RMClient,
			valid:      true,
		},
		{
			name: "RegisterInvalidHostDuplicateUUIDNoSNFromRM",
			in: &computev1.HostResource{
				Name: "Test Host Register",
				Uuid: defaultUUIDRM,
			},
			clientName: inv_testing.RMClient,
			valid:      false,
		},
		{
			name: "RegisterInvalidHostNoUUIDDuplicateSNFromRM",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				SerialNumber: "SN100002",
			},
			clientName: inv_testing.RMClient,
			valid:      false,
		},
		{
			name: "RegisterInvalidHostDuplicateUUIDUniqueSNFromRM",
			in: &computev1.HostResource{
				Name:         "Test Host Register",
				Uuid:         defaultUUIDRM,
				SerialNumber: "SN100004",
			},
			clientName: inv_testing.RMClient,
			valid:      false,
		},
		{
			name: "RegisterInvalidHostNoUUIDNoSNFromRM",
			in: &computev1.HostResource{
				Name: "Test Host Register",
			},
			clientName: inv_testing.RMClient,
			valid:      false,
		},
		{
			name: "RegisterHostWithAMT",
			in: &computev1.HostResource{
				Name:            "Test Host Register",
				Uuid:            uuid.New().String(),
				DesiredState:    computev1.HostState_HOST_STATE_ONBOARDED,
				DesiredAmtState: computev1.AmtState_AMT_STATE_PROVISIONED,
			},
			clientName: inv_testing.APIClient,
			valid:      true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create host
			chostResp, err := inv_testing.TestClients[tc.clientName].Create(ctx, createresreq)
			hostResID := chostResp.GetHost().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateHost() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = hostResID // Update with created resource ID.
				tc.in.HostStatus = res_status.DefaultHostStatus
				tc.in.OnboardingStatus = res_status.DefaultOnboardingStatus
				tc.in.RegistrationStatus = res_status.DefaultRegistrationStatus
				tc.in.CreatedAt = chostResp.GetHost().GetCreatedAt()
				tc.in.UpdatedAt = chostResp.GetHost().GetUpdatedAt()
				if tc.in.AmtSku == "" {
					tc.in.AmtSku = "Unknown"
				}
				if tc.in.AmtStatus == "" {
					tc.in.AmtStatus = res_status.DefaultAmtStatus
				}
				if tc.in.PowerStatus == "" {
					tc.in.PowerStatus = res_status.DefaultPowerStatus
				}
				assertSameResource(t, createresreq, chostResp, nil)
				if !tc.valid {
					t.Errorf("CreateHost() succeeded but should have failed")
				}
			}

			// verify if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				registeredHostIDs = append(registeredHostIDs, hostResID)
				// get host
				getresp, err := inv_testing.TestClients[tc.clientName].Get(ctx, hostResID)
				require.NoError(t, err, "GetHost() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetHost()); !eq {
					t.Errorf("GetHost() data not equal: %v", diff)
				}
			}
		})
	}

	// remove registered hosts
	for _, hostResID := range registeredHostIDs {
		// build a context for gRPC
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		// Remove host.
		inv_testing.HardDeleteHost(t, hostResID)

		// get after complete Delete of host, should fail as Host is 2-phase deleted
		_, err := inv_testing.TestClients[inv_testing.RMClient].Get(ctx, hostResID)
		require.Error(t, err, "Failure - Host was not deleted, but should be deleted")
		cancel()
	}
}

type StateTransitionCase struct {
	in         *computev1.HostResource
	resourceID string
	fieldMask  *fieldmaskpb.FieldMask
	valid      bool
}

func Test_HostStateTransition(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	hostUntrusted := createHostWithCurrentState(ctx, t, computev1.HostState_HOST_STATE_UNTRUSTED)

	testcases := map[string]StateTransitionCase{
		// try to update other fields, this should be allowed
		"OtherFields": {
			in: &computev1.HostResource{
				Name: "test",
			},
			resourceID: hostUntrusted.ResourceId,
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldName}},
			valid:      true,
		},
	}

	testcases["UntrustedToUntrusted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNTRUSTED, hostUntrusted.ResourceId, true)
	testcases["UntrustedToDeleted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_DELETED, hostUntrusted.ResourceId, true)
	testcases["UntrustedToUnspecified"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNSPECIFIED, hostUntrusted.ResourceId, false)
	testcases["UntrustedToRegistered"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_REGISTERED, hostUntrusted.ResourceId, false)
	testcases["UntrustedToOnboarded"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_ONBOARDED, hostUntrusted.ResourceId, false)

	hostUnspecified := createHostWithCurrentState(ctx, t, computev1.HostState_HOST_STATE_UNSPECIFIED)

	testcases["UnspecifiedToUnspecified"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNSPECIFIED, hostUnspecified.ResourceId, true)
	testcases["UnspecifiedToDeleted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_DELETED, hostUnspecified.ResourceId, true)
	testcases["UnspecifiedToRegistered"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_REGISTERED, hostUnspecified.ResourceId, true)
	testcases["UnspecifiedToOnboarded"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_ONBOARDED, hostUnspecified.ResourceId, true)
	testcases["UnspecifiedToUntrusted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNTRUSTED, hostUnspecified.ResourceId, true)

	hostOnboarded := createHostWithCurrentState(ctx, t, computev1.HostState_HOST_STATE_ONBOARDED)

	testcases["OnboardedToDeleted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_DELETED, hostOnboarded.ResourceId, true)
	testcases["OnboardedToOnboarded"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_ONBOARDED, hostOnboarded.ResourceId, true)
	testcases["OnboardedToUntrusted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNTRUSTED, hostOnboarded.ResourceId, true)
	testcases["OnboardedToRegistered"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_REGISTERED, hostOnboarded.ResourceId, false)
	testcases["OnboardedToUnspecified"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNSPECIFIED, hostOnboarded.ResourceId, false)

	hostRegistered := createHostWithCurrentState(ctx, t, computev1.HostState_HOST_STATE_REGISTERED)

	testcases["RegisteredToDeleted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_DELETED, hostRegistered.ResourceId, true)
	testcases["RegisteredToRegistered"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_REGISTERED, hostRegistered.ResourceId, true)
	testcases["RegisteredToOnboarded"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_ONBOARDED, hostRegistered.ResourceId, true)
	testcases["RegisteredToUnspecified"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNSPECIFIED, hostRegistered.ResourceId, false)
	testcases["RegisteredToUntrusted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNTRUSTED, hostRegistered.ResourceId, false)

	hostDeleted := createHostWithCurrentState(ctx, t, computev1.HostState_HOST_STATE_DELETED)

	testcases["DeletedToDeleted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_DELETED, hostDeleted.ResourceId, true)
	testcases["DeletedToRegistered"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_REGISTERED, hostDeleted.ResourceId, false)
	testcases["DeletedToOnboarded"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_ONBOARDED, hostDeleted.ResourceId, false)
	testcases["DeletedToUnspecified"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNSPECIFIED, hostDeleted.ResourceId, false)
	testcases["DeletedToUntrusted"] = getStateTransitionCase(
		t, computev1.HostState_HOST_STATE_UNTRUSTED, hostDeleted.ResourceId, false)

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			requiredHost := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{Host: tc.in},
			}
			_, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, tc.resourceID, tc.fieldMask, requiredHost)

			if tc.valid {
				require.NoErrorf(t, err, "Update() failed: %s", err)
				return
			}
			require.Errorf(t, err, "Update() succeeded but should have failed")
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func createHostWithCurrentState(ctx context.Context, t *testing.T, currentState computev1.HostState) *computev1.HostResource {
	t.Helper()
	host := inv_testing.CreateHost(t, nil, nil)

	_, err := inv_testing.GetClient(t, inv_testing.RMClient).Update(
		ctx,
		host.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: &computev1.HostResource{
					CurrentState: currentState,
				},
			},
		},
	)
	require.NoError(t, err, "UpdateHost() failed")
	return host
}

func getStateTransitionCase(t *testing.T, state computev1.HostState, resourceID string, isValid bool) StateTransitionCase {
	t.Helper()
	return StateTransitionCase{
		in: &computev1.HostResource{
			DesiredState: state,
		},
		resourceID: resourceID,
		fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{hostresource.FieldDesiredState}},
		valid:      isValid,
	}
}

func Test_LimitAndOffset(t *testing.T) {
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Create a large amount of hosts
	for i := 0; i < 99; i++ {
		inv_testing.CreateHost(t, nil, nil)
	}

	findresreq1 := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{},
		},
		Limit: 50,
	}

	findres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, findresreq1)
	require.NoError(t, err)
	assert.Equal(t, 50, len(findres.GetResources()))
	assert.True(t, findres.HasNext)
	assert.Equal(t, findres.TotalElements, int32(99))

	findresreq2 := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{},
		},
		Limit:  50,
		Offset: 50,
	}

	findres2, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, findresreq2)
	require.NoError(t, err)

	assert.Equal(t, 49, len(findres2.GetResources()))
	assert.False(t, findres2.HasNext)
	assert.Equal(t, findres2.TotalElements, int32(99))
}

func Test_LimitAndOffsetWithMetadataFilter(t *testing.T) {
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Create a large amount of hosts with metadata
	for i := 0; i < 99; i++ {
		inv_testing.CreateHostWithMetadata(t, metaHost1, nil, nil, true)
	}

	listResReq1 := &inv_v1.ResourceFilter{
		Filter: fmt.Sprintf("%s = '%s'", hostresource.FieldMetadata, metaHost1),
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: &computev1.HostResource{
					Metadata: metaHost1,
				},
			},
		},
		Limit: 50,
	}
	listRes, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, listResReq1)
	require.NoError(t, err)
	assert.Equal(t, 50, len(listRes.GetResources()))
	assert.True(t, listRes.HasNext)
	assert.Equal(t, listRes.TotalElements, int32(99))

	findRes, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, listResReq1)
	require.NoError(t, err)
	assert.Equal(t, 50, len(findRes.GetResources()))
	assert.True(t, findRes.HasNext)
	assert.Equal(t, findRes.TotalElements, int32(99))

	listResReq2 := &inv_v1.ResourceFilter{
		Filter: fmt.Sprintf("%s = '%s'", hostresource.FieldMetadata, metaHost1),
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: &computev1.HostResource{
					Metadata: metaHost1,
				},
			},
		},
		Limit:  50,
		Offset: 50,
	}
	listRes2, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, listResReq2)
	require.NoError(t, err)
	assert.Equal(t, 49, len(listRes2.GetResources()))
	assert.False(t, listRes2.HasNext)
	assert.Equal(t, listRes2.TotalElements, int32(99))

	findRes2, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, listResReq2)
	require.NoError(t, err)
	assert.Equal(t, 49, len(findRes2.GetResources()))
	assert.False(t, findRes2.HasNext)
	assert.Equal(t, findRes2.TotalElements, int32(99))
}

// Note the delicacy of handling NULL columns. With a naive NEQ filter, nothing is returned as a SQL NULL is not equal
// to anything, nor is it unequal to anything.
// See: https://stackoverflow.com/a/18243804
// To address this, an additional `OR <> = null` clause can be used.
func Test_FilterHosts(t *testing.T) {
	region1 := inv_testing.CreateRegion(t, nil)
	region2 := inv_testing.CreateRegion(t, region1)
	site1 := inv_testing.CreateSiteWithArgs(t, "SJI1", 100, 100, "", region2, nil, nil)
	site2 := inv_testing.CreateSiteWithArgs(t, "RNB1", 200, 200, "", nil, nil, nil)
	provider1 := inv_testing.CreateProvider(t, "Test Provider1")
	provider2 := inv_testing.CreateProvider(t, "Test Provider2")
	os1 := inv_testing.CreateOs(t)

	uuidH1 := uuid.NewString()
	// create Hosts to find
	createresreq1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 3",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,

				Site:         site1,
				Provider:     provider1,
				HardwareKind: "XDgen4",
				SerialNumber: "1001",
				MemoryBytes:  64 * util.Gigabyte,
				Uuid:         uuidH1,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      13,

				MgmtIp: "192.168.10.13",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.13",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:f3",

				Hostname: "testhost3",
				Metadata: metaHost1,
			},
		},
	}

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 4",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,

				Site:         site2,
				Provider:     provider2,
				HardwareKind: "XDgen3",
				SerialNumber: "1002",
				Uuid:         uuid.NewString(),
				MemoryBytes:  2 * util.Gigabyte,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      10,

				MgmtIp: "192.168.10.14",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.14",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:f4",

				Hostname: "testhost4",
				Metadata: metaHost2,
			},
		},
	}

	createresreq3 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 4",
				DesiredState: computev1.HostState_HOST_STATE_UNSPECIFIED,
				Uuid:         uuid.NewString(),
				MemoryBytes:  64 * util.Gigabyte,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      12,

				MgmtIp: "192.168.10.14",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.14",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:f4",

				Hostname: "testhost4",
			},
		},
	}

	createresreq4 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 4",
				CurrentState: computev1.HostState_HOST_STATE_DELETED,
				DesiredState: computev1.HostState_HOST_STATE_UNSPECIFIED,

				Site:         site2,
				Provider:     provider2,
				HardwareKind: "XDgen3",
				SerialNumber: "1002",
				Uuid:         uuid.NewString(),
				MemoryBytes:  128 * util.Gigabyte,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-13900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      10,

				MgmtIp: "192.168.10.14",

				BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.14",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:f4",

				Hostname: "testhost4",
				Metadata: metaHost2,

				CurrentPowerState:    computev1.PowerState_POWER_STATE_ON,
				PowerStatus:          "Powered On",
				PowerStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				PowerStatusTimestamp: uint64(time.Now().UnixNano()), //nolint:gosec // This is a test

				CurrentAmtState:    computev1.AmtState_AMT_STATE_UNPROVISIONED,
				AmtStatus:          "Provisioned",
				AmtStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				AmtStatusTimestamp: uint64(time.Now().UnixNano()), //nolint:gosec // This is a test
				AmtSku:             "vPRO Corporate 16.5.2",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	// Create four test hosts.
	chostResp1, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expHost1 := *chostResp1.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, expHost1.GetResourceId()) })

	chostResp2, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expHost2 := *chostResp2.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, expHost2.GetResourceId()) })

	chostResp3, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expHost3 := *chostResp3.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, expHost3.GetResourceId()) })

	chostResp4, err := inv_testing.TestClients[inv_testing.RMClient].Create(ctx, createresreq4)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expHost4 := *chostResp4.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, expHost4.GetResourceId()) })

	instance1 := inv_testing.CreateInstance(t, &expHost1, os1)
	instance1.DesiredOs = os1
	instance1.CurrentOs = os1
	instance1.Os = os1
	expHost1.Instance = instance1

	hostStorage1 := inv_testing.CreateHostStorage(t, &expHost1)
	expHost1.HostStorages = append(expHost1.HostStorages, hostStorage1)

	hostNic1 := inv_testing.CreateHostNic(t, &expHost1)
	expHost1.HostNics = append(expHost1.HostNics, hostNic1)

	hostUsb1 := inv_testing.CreateHostusb(t, &expHost1)
	expHost1.HostUsbs = append(expHost1.HostUsbs, hostUsb1)

	hostGpu1 := inv_testing.CreatHostGPU(t, &expHost1)
	expHost1.HostGpus = append(expHost1.HostGpus, hostGpu1)

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.HostResource
		valid     bool
	}{
		"NoFilter": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3, &expHost4},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
				OrderBy: hostresource.FieldResourceID,
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3, &expHost4},
			valid:     true,
		},
		"InvalidNilFilter": {
			in:        nil,
			resources: nil,
			valid:     false,
		},
		"FilterBySiteID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldResourceID, site1.GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterBySiteIDJSON": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`site.resourceId = %q`, site1.GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasSite": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`has(%s)`, hostresource.EdgeSite),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost4},
			valid:     true,
		},
		"FilterByNotHasSite": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeSite),
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterByHasSiteHasRegion": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`has(%s.%s)`, hostresource.EdgeSite, siteresource.EdgeRegion),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasSiteHasRegionHasParentRegion": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`has(%s.%s.%s)`, hostresource.EdgeSite,
					siteresource.EdgeRegion, regionresource.EdgeParentRegion),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterBySiteNameNonExistent": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldName, "San Jose"),
			},
			resources: []*computev1.HostResource{},
			valid:     true,
		},
		"FilterBySiteName": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldName, site2.GetName()),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteByNameIgnoreCase": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldName,
					strings.ToLower(site2.GetName())),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteNamePartialPrefix": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldName, "RN*" /*site2*/),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteNamePartialSuffix": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldName, "*NB1" /*site2*/),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteNameWildcard": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldName, "R*1" /*site2*/),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteLat": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.%s = %v`, hostresource.EdgeSite, siteresource.FieldSiteLat, site2.GetSiteLat()),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteLatJSON": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s.siteLat = %v`, hostresource.EdgeSite, site2.GetSiteLat()),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterBySiteByRegionByID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeSite, siteresource.EdgeRegion,
					siteresource.FieldResourceID, region2.GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasSiteAndMemoryBytesEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`has(%s) AND %s = %v`, hostresource.EdgeSite,
					hostresource.FieldMemoryBytes, expHost2.GetMemoryBytes()),
			},
			resources: []*computev1.HostResource{&expHost2},
			valid:     true,
		},
		"FilterByUuidEqTwice": {
			// This test ensures we can safely construct filter queries with "duplicate" clauses.
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = %q AND %s = %q`, hostresource.FieldUUID, expHost1.GetUuid(),
					hostresource.FieldUUID, expHost1.GetUuid()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByCPUThreads": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %v`, hostresource.FieldCPUThreads, expHost1.GetCpuThreads()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByProvider": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeProvider, providerresource.FieldResourceID,
					provider1.GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasProvider": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`has(%s)`, hostresource.EdgeProvider),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost4},
			valid:     true,
		},
		"FilterByProviderID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeProvider, providerresource.FieldResourceID,
					expHost1.GetProvider().GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByInstanceID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeInstance, instanceresource.FieldResourceID,
					expHost1.GetInstance().GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByNicID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostNics, hostnicresource.FieldResourceID,
					expHost1.GetHostNics()[0].GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByUsbID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostUsbs, hostusbresource.FieldResourceID,
					expHost1.GetHostUsbs()[0].GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByGpuID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostGpus, hostgpuresource.FieldResourceID,
					expHost1.GetHostGpus()[0].GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByInstance": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeInstance,
					instanceresource.FieldResourceID, instance1.GetResourceId()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasInstance": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`has(%s)`, hostresource.EdgeInstance),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByNotHasInstance": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeInstance),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost3, &expHost4},
			valid:     true,
		},
		"FilterHardwareKind": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("%s = %q", hostresource.FieldHardwareKind, "XDgen4"),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterDesiredState": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("%s = %s", hostresource.FieldDesiredState, computev1.HostState_HOST_STATE_ONBOARDED),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2},
			valid:     true,
		},
		"FilterCurrentState": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("%s = %s", hostresource.FieldCurrentState, computev1.HostState_HOST_STATE_DELETED),
			},
			resources: []*computev1.HostResource{&expHost4},
			valid:     true,
		},
		"FilterUUID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("%s = %q", hostresource.FieldUUID, uuidH1),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByMetadata0": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = '%s'`, hostresource.FieldMetadata, `{"key":"key1-test","value":"host_key1-test"}`),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost4},
			valid:     true,
		},
		"FilterByMetadata1": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = '%s'`, hostresource.FieldMetadata, `{"key":"key4-test","value":"host_key4-test"}`),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByMetadata2": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = '%s'`, hostresource.FieldMetadata,
					`{"key":"key1-test","value":"host_key1-test"},{"key":"key2-test","value":"host_key2-test"}`),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost4},
			valid:     true,
		},
		"FilterByMetadata2b": {
			// Using an AND expression the JSON key order does not matter, compare to above.
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = '%s' AND %s = '%s'`,
					hostresource.FieldMetadata, `{"key":"key2-test","value":"host_key2-test"}`,
					hostresource.FieldMetadata, `{"key":"key1-test","value":"host_key1-test"}`,
				),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost4},
			valid:     true,
		},
		"FilterByMetadata3": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = '%s' AND %s = '%s'`,
					hostresource.FieldMetadata, `{"key":"key1-test","value":"host_key1-test"}`,
					hostresource.FieldMetadata, `{"key":"key3-test","value":"host_key3-test"}`,
				),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByMetadata4": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = '%s' AND %s = '%s'`,
					hostresource.FieldMetadata, `{"key":"key1-test","value":"host_key1-test"}`,
					hostresource.FieldMetadata, `{"key":"key4-test","value":"host_key4-test"}`,
				),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByMetadata5": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = '%s'`,
					hostresource.FieldMetadata, `{"key":"key3-test","value":"host_key3_mod-test"}`,
				),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterSiteEmpty": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("NOT has(%s)", hostresource.EdgeSite),
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterProviderEmpty": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("NOT has(%s)", hostresource.EdgeProvider),
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterSiteAndProviderEmpty": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("NOT has(%s) AND NOT has(%s)", hostresource.EdgeProvider, hostresource.EdgeSite),
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterInstancesEmpty": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf("NOT has(%s)", hostresource.EdgeInstance),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost3, &expHost4},
			valid:     true,
		},
		// No support for string + null comparison
		"FilterHardwareKindEmpty": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = ""`, hostresource.FieldHardwareKind),
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterDesiredStateEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %s`, hostresource.FieldDesiredState, computev1.HostState_HOST_STATE_UNSPECIFIED),
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
			},
			resources: []*computev1.HostResource{&expHost3, &expHost4},
			valid:     true,
		},
		"FilterCurrentStateEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %s`, hostresource.FieldCurrentState, computev1.HostState_HOST_STATE_UNSPECIFIED),
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3},
			valid:     true,
		},
		"FilterUUIDEmpty": {
			// UUID cannot be empty
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, hostresource.FieldUUID),
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
			},
			valid: true,
		},
		"FilterSerialNumberEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, hostresource.FieldSerialNumber),
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterFieldMaskMetadataEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, hostresource.FieldMetadata),
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{},
				},
			},
			resources: []*computev1.HostResource{&expHost3},
			valid:     true,
		},
		"FilterWithOffsetLimit": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Offset:   5,
				Limit:    0,
			},
			valid: true,
		},
		"FilterInvalidEdge": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`has(%s)`, "invalid_edge"),
			},
			valid: false,
		},
		"FilterInvalidField": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %q`, "invalid_field", "some-value"),
			},
			valid: false,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %q`, hostresource.FieldResourceID, expHost1.ResourceId),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByResourceIdEqJSON": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`resourceId = %q`, expHost1.ResourceId),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByNameEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %q`, hostresource.FieldName, expHost1.GetName()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByCurrentPowerStateUnspecified": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = %s`, hostresource.FieldCurrentPowerState,
					computev1.PowerState_POWER_STATE_UNSPECIFIED),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3},
			valid:     true,
		},
		"FilterByCurrentPowerStateNotUnspecified": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s != %s`, hostresource.FieldCurrentPowerState,
					computev1.PowerState_POWER_STATE_UNSPECIFIED),
			},
			resources: []*computev1.HostResource{&expHost4},
			valid:     true,
		},
		"FilterByCurrentPowerStateNotOn": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s != %s`, hostresource.FieldCurrentPowerState, computev1.PowerState_POWER_STATE_ON),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3},
			valid:     true,
		},
		"FilterByCurrentPowerStateNotOnOrNull": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s != %s OR %s = null`, hostresource.FieldCurrentPowerState,
					computev1.PowerState_POWER_STATE_ON, hostresource.FieldCurrentPowerState),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3},
			valid:     true,
		},
		"InvalidFilterByCurrentStateLt": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s < %s`, hostresource.FieldCurrentState, computev1.HostState_HOST_STATE_UNTRUSTED),
			},
			resources: nil,
			valid:     false,
		},
		"FilterByMemoryBytesEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %v`, hostresource.FieldMemoryBytes, expHost2.GetMemoryBytes()),
			},
			resources: []*computev1.HostResource{&expHost2},
			valid:     true,
		},
		"FilterByMemoryBytesNeq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s != %v`, hostresource.FieldMemoryBytes, expHost2.GetMemoryBytes()),
			},

			resources: []*computev1.HostResource{&expHost1, &expHost3, &expHost4},
			valid:     true,
		},
		"FilterByMemoryBytesOr": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s = %v OR %s = %v`, hostresource.FieldMemoryBytes,
					expHost2.GetMemoryBytes(), hostresource.FieldMemoryBytes, expHost4.GetMemoryBytes()),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost4},
			valid:     true,
		},
		"FilterByMemoryBytesGe": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s >= %v`, hostresource.FieldMemoryBytes, 64*util.Gigabyte),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost3, &expHost4},
			valid:     true,
		},
		"FilterByMemoryBytesLe": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s <= %v`, hostresource.FieldMemoryBytes, expHost2.GetMemoryBytes()),
			},
			resources: []*computev1.HostResource{&expHost2},
			valid:     true,
		},
		"FilterByMemoryBytesLt": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s < %v`, hostresource.FieldMemoryBytes, expHost2.GetMemoryBytes()),
			},
			resources: []*computev1.HostResource{},
			valid:     true,
		},
		"FilterByHasHostStorages": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`has(%s)`, hostresource.EdgeHostStorages),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasHostStoragesWithCapacityGt": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s > %d`, hostresource.EdgeHostStorages,
					hoststorageresource.FieldCapacityBytes, hostStorage1.GetCapacityBytes()),
			},
			resources: []*computev1.HostResource{},
			valid:     true,
		},
		"FilterByHasHostStoragesWithCapacityGe": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter: fmt.Sprintf(`%s.%s >= %d`, hostresource.EdgeHostStorages,
					hoststorageresource.FieldCapacityBytes, hostStorage1.GetCapacityBytes()),
			},
			resources: []*computev1.HostResource{&expHost1},
			valid:     true,
		},
		"FilterByHasNotHostStorages": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeHostStorages),
			},
			resources: []*computev1.HostResource{&expHost2, &expHost3, &expHost4},
			valid:     true,
		},
		"Injection": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   "; BEGIN;COMMIT;",
			},
			resources: nil,
			valid:     false,
		},
		"InvalidNonBoolResult": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   "1",
			},
			resources: nil,
			valid:     false,
		},
		"InvalidBool": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   "true",
			},
			resources: nil,
			valid:     false,
		},
		"FilterByPowerStatus": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %q`, hostresource.FieldPowerStatus, "Powered On"),
			},
			resources: []*computev1.HostResource{&expHost4},
			valid:     true,
		},
		"FilterByAmtSku": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %q`, hostresource.FieldAmtSku, "Unknown"),
			},
			resources: []*computev1.HostResource{&expHost1, &expHost2, &expHost3},
			valid:     true,
		},
		"FilterByAmtStatus": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				Filter:   fmt.Sprintf(`%s = %q`, hostresource.FieldAmtStatus, "Provisioned"),
			},
			resources: []*computev1.HostResource{&expHost4},
			valid:     true,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			findresIDs, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterHosts() failed: %s", errors.ErrorToStringWithDetails(err))
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterHosts() succeeded but should have failed")
				}
			}

			// only get/delete if valid test with non-zero returned response and hasn't failed, otherwise may segfault
			if !t.Failed() && tc.valid {
				if len(findresIDs.GetResources()) != len(tc.resources) {
					t.Errorf("Expected to obtain %d Resource IDs, but obtained back %d Resource IDs",
						len(tc.resources), len(findresIDs.GetResources()))
				}

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findresIDs.Resources)

				if !reflect.DeepEqual(resIDs, findresIDs.Resources) {
					t.Errorf(
						"FilterHosts() failed - want: %s, got: %s",
						resIDs,
						findresIDs.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListHosts() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListHosts() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.HostResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHost())
				}

				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					expCopy := *expected //nolint:govet // ok to copy lock in test
					hostEdgesOnlyResourceID(&expCopy)
					hostEdgesOnlyResourceID(resources[i])

					// Compare metadata separately
					assert.True(t, CompareMetadata(t, resources[i].Metadata, expCopy.Metadata))
					expCopy.Metadata = ""
					resources[i].Metadata = ""
					if eq, diff := inv_testing.ProtoEqualOrDiff(&expCopy, resources[i]); !eq {
						t.Errorf("ListHost() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

// Tests Get and List hosts with and without a metadata hierarchy.
//
//nolint:gosec // ok to use math/rand in tests
func BenchmarkInvStore_BenchHostRetrieval(b *testing.B) {
	l := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	b.Cleanup(func() { zerolog.SetGlobalLevel(l) })

	// Loop for different number of hosts.
	for _, i := range []int{100, 500, 1000} {
		// Loop for with and without metadata inheritance.
		for _, inheritance := range []bool{true, false} {
			inheritanceString := ""
			if inheritance {
				inheritanceString = "With"
			} else {
				inheritanceString = "Without"
			}
			b.Run(fmt.Sprintf("%sInheritance%d", inheritanceString, i), func(b *testing.B) {
				resIDs := createRandomHostsWithHierarchy(b, i, inheritance)
				b.Run("List", func(b *testing.B) {
					benchmarkInvStoreListHost(b)
				})
				b.Run("Get", func(b *testing.B) {
					benchmarkInvStoreGetRes(b, resIDs[rand.Intn(len(resIDs))])
				})
			})
		}
	}
}

func benchmarkInvStoreGetRes(b *testing.B, resID string) {
	b.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, resID)
		require.NoError(b, err)
	}
	b.StopTimer()
}

func benchmarkInvStoreListHost(b *testing.B) {
	b.Helper()

	// TODO: add parameter for limit
	listHostReq := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{},
		},
		Limit: 100,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, listHostReq)
		assert.NoError(b, err)
	}
	b.StopTimer()
}

func createRandomHostsWithHierarchy(b *testing.B, num int, createHierarchy bool) []string {
	b.Helper()

	resIDs := make([]string, 0, num)
	var site1 *location_v1.SiteResource
	site1 = nil
	if createHierarchy {
		// Create a linear graph of 4 regions.
		region1 := inv_testing.CreateRegionWithMeta(b, metaR1, nil)
		region2 := inv_testing.CreateRegionWithMeta(b, metaR2, region1)
		region3 := inv_testing.CreateRegionWithMeta(b, metaR3, region2)
		region4 := inv_testing.CreateRegionWithMeta(b, metaO1, region3)
		region5 := inv_testing.CreateRegionWithMeta(b, metaO2, region4)
		site1 = inv_testing.CreateSiteWithMeta(b, metaO3, region5, nil)
	}
	// Create a bunch of hosts in order to benchmark reads
	createHostRequest := &inv_v1.Resource{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(num)*500*time.Millisecond)
	b.Cleanup(cancel)
	for id := 0; id < num; id++ {
		createHostRequest.Resource = &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Site: site1,
				Uuid: uuid.NewString(),
			},
		}
		hostResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createHostRequest)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		hostResID := inv_testing.GetResourceIDOrFail(b, hostResp)
		b.Cleanup(func() { inv_testing.HardDeleteHost(b, hostResID) })
		resIDs = append(resIDs, hostResID)
	}
	return resIDs
}

func hostEdgesOnlyResourceID(expected *computev1.HostResource) {
	if expected.Site != nil {
		expected.Site = &location_v1.SiteResource{ResourceId: expected.Site.ResourceId}
	}
	if expected.Provider != nil {
		expected.Provider = &provider_v1.ProviderResource{ResourceId: expected.Provider.ResourceId}
	}
}

func Test_NestedFilterHost(t *testing.T) {
	region1 := inv_testing.CreateRegion(t, nil)
	region2 := inv_testing.CreateRegion(t, region1)
	ou1 := inv_testing.CreateOu(t, nil)
	site1 := inv_testing.CreateSite(t, region2, nil)
	site2 := inv_testing.CreateSite(t, region1, ou1)

	host1 := inv_testing.CreateHost(t, site1, nil)
	hostStorage := inv_testing.CreateHostStorage(t, host1)
	hostNic1 := inv_testing.CreateHostNic(t, host1)
	hostNic2 := inv_testing.CreateHostNic(t, host1)
	hostUsb := inv_testing.CreateHostusb(t, host1)
	os := inv_testing.CreateOs(t)

	localaccount := inv_testing.CreateLocalAccount(t,
		"test-user",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ test-user1@example.com",
	)
	instance := inv_testing.CreateInstanceWithLocalAccount(t, host1, os, localaccount)

	instance.DesiredOs = os
	instance.CurrentOs = os
	instance.Os = os

	hostGpu1 := inv_testing.CreatHostGPU(t, host1)
	host1.Site = site1
	host1.Instance = instance
	host1.HostStorages = append(host1.HostStorages, hostStorage)
	host1.HostNics = append(host1.HostNics, hostNic1, hostNic2)
	host1.HostUsbs = append(host1.HostUsbs, hostUsb)
	host1.HostGpus = append(host1.HostGpus, hostGpu1)

	host2 := inv_testing.CreateHost(t, site2, nil)
	host2.Site = site2

	host3 := inv_testing.CreateHost(t, nil, nil)
	hostStorage2 := inv_testing.CreateHostStorage(t, host3)
	host3.HostStorages = append(host3.HostStorages, hostStorage2)

	testcases := map[string]struct {
		in                *inv_v1.ResourceFilter
		resources         []*computev1.HostResource
		valid             bool
		expectedCodeError codes.Code
	}{
		"FilterBySiteID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeSite, siteresource.FieldResourceID, site1.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByUUID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeSite, siteresource.EdgeOu,
					ouresource.FieldResourceID, ou1.GetResourceId()),
			},
			resources: []*computev1.HostResource{host2},
			valid:     true,
		},
		"FilterByRegionID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeSite, siteresource.EdgeRegion,
					regionresource.FieldResourceID, region1.GetResourceId()),
			},
			resources: []*computev1.HostResource{host2},
			valid:     true,
		},
		"FilterByOsID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeInstance, instanceresource.EdgeDesiredOs,
					operatingsystemresource.FieldResourceID, os.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByOsID2": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeInstance, instanceresource.EdgeOs,
					operatingsystemresource.FieldResourceID, os.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByHasHostStorages": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, hostresource.EdgeHostStorages),
			},
			resources: []*computev1.HostResource{host1, host3},
			valid:     true,
		},
		"FilterByHostStorageID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostStorages,
					hoststorageresource.FieldResourceID, hostStorage2.GetResourceId()),
			},
			resources: []*computev1.HostResource{host3},
			valid:     true,
		},
		"FilterByNotHasHostStorages": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeHostStorages),
			},
			resources: []*computev1.HostResource{host2},
			valid:     true,
		},
		"FilterByHostNicID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostNics,
					hostnicresource.FieldResourceID, hostNic1.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByHasHostNics": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, hostresource.EdgeHostNics),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByNotHasHostNics": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeHostNics),
			},
			resources: []*computev1.HostResource{host2, host3},
			valid:     true,
		},
		"FilterByHostUsbID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostUsbs,
					hostusbresource.FieldResourceID, hostUsb.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByHasHostUsbs": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, hostresource.EdgeHostUsbs),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByNotHasHostUsbs": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeHostUsbs),
			},
			resources: []*computev1.HostResource{host2, host3},
			valid:     true,
		},
		"FilterByHasInstances": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, hostresource.EdgeInstance),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByNotHasInstances": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeInstance),
			},
			resources: []*computev1.HostResource{host2, host3},
			valid:     true,
		},
		"FilterByHostGpuID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostresource.EdgeHostGpus,
					hostgpuresource.FieldResourceID, hostGpu1.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByHasHostGpus": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, hostresource.EdgeHostGpus),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByNotHasGpus": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostresource.EdgeHostGpus),
			},
			resources: []*computev1.HostResource{host2, host3},
			valid:     true,
		},
		"FilterByLocalAccountID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeInstance, instanceresource.EdgeLocalaccount,
					localaccountresource.FieldResourceID, localaccount.GetResourceId()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FilterByLocalAccountUsername": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostresource.EdgeInstance, instanceresource.EdgeLocalaccount,
					localaccountresource.FieldUsername, localaccount.GetUsername()),
			},
			resources: []*computev1.HostResource{host1},
			valid:     true,
		},
		"FailTooDeep": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s.%s.%s.%s.%s.%s.%s)`, hostresource.EdgeSite, siteresource.EdgeRegion,
					regionresource.EdgeParentRegion, regionresource.EdgeParentRegion,
					regionresource.EdgeParentRegion, regionresource.EdgeParentRegion,
					regionresource.EdgeParentRegion),
			},
			valid:             false,
			expectedCodeError: codes.InvalidArgument,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}} // Set the resource kind

			// Test FIND
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expectedCodeError, status.Code(err))
			} else {
				require.NoError(t, err)

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findres.Resources)

				if !reflect.DeepEqual(resIDs, findres.Resources) {
					t.Errorf(
						"FilterInstances() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			// Test LIST
			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expectedCodeError, status.Code(err))
			} else {
				require.NoError(t, err)
				require.Len(t, listres.Resources, len(tc.resources))

				resources := make([]*computev1.HostResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHost())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListInstances() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func Test_StrongRelations_On_Delete_Host(t *testing.T) {
	t.Run("Host_Instance", func(t *testing.T) {
		os := inv_testing.CreateOs(t)
		host := inv_testing.CreateHost(t, nil, nil)
		inv_testing.CreateInstance(t, host, os)

		err := inv_testing.HardDeleteHostAndReturnError(t, host.ResourceId)
		assertStrongRelationError(t, err, "host has a relation with Instance and cannot be deleted")
	})
}

func Test_OrderBy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	createresreq1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 1",
				DesiredState: computev1.HostState_HOST_STATE_DELETED,
				MemoryBytes:  64 * util.Gigabyte,
				CpuCores:     8,
				Uuid:         uuid.NewString(),
			},
		},
	}
	chostResp1, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	host1 := chostResp1.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host1.GetResourceId()) })

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 2",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
				MemoryBytes:  16 * util.Gigabyte,
				CpuCores:     8,
				Uuid:         uuid.NewString(),
			},
		},
	}
	chostResp2, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	host2 := chostResp2.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host2.GetResourceId()) })

	createresreq3 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &computev1.HostResource{
				Name:         "Test Host 3",
				DesiredState: computev1.HostState_HOST_STATE_DELETED,
				MemoryBytes:  128 * util.Gigabyte,
				CpuCores:     16,
				Uuid:         uuid.NewString(),
			},
		},
	}
	chostResp3, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	host3 := chostResp3.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host3.GetResourceId()) })

	type testCase struct {
		orderBy           string
		expectedErrorCode codes.Code
		expectedOrder     []*computev1.HostResource
	}
	tests := map[string]testCase{
		"Empty": {
			orderBy:           "",
			expectedErrorCode: codes.OK,
			expectedOrder:     inv_testing.GetOrderByResourceID([]*computev1.HostResource{host3, host1, host2}),
		},
		"Name": {
			orderBy:           hostresource.FieldName,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host1, host2, host3},
		},
		"NameDesc": {
			orderBy:           hostresource.FieldName + " desc",
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host3, host2, host1},
		},
		"NameAsc": {
			orderBy:           hostresource.FieldName + " asc",
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host1, host2, host3},
		},
		// First occurrence wins for duplicate or conflicting order.
		"NameTwice": {
			orderBy:           hostresource.FieldName + "," + hostresource.FieldName,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host1, host2, host3},
		},
		"NameTwiceConflictingOrder": {
			orderBy:           hostresource.FieldName + " desc," + hostresource.FieldName,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host3, host2, host1},
		},
		"MemoryBytes": {
			orderBy:           hostresource.FieldMemoryBytes,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host2, host1, host3},
		},
		"MemoryBytesJSON": {
			orderBy:           hostresource.FieldMemoryBytes,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host2, host1, host3},
		},
		"DesiredStateEnum": { // Sorted alphabetically by string representation of enum
			orderBy:           hostresource.FieldDesiredState,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host1, host3, host2},
		},
		"CpuCoresAndMemoryBytes": {
			orderBy:           hostresource.FieldCPUCores + ", " + hostresource.FieldMemoryBytes,
			expectedErrorCode: codes.OK,
			expectedOrder:     []*computev1.HostResource{host2, host1, host3},
		},
		"InvalidSiteEdge": {
			orderBy:           hostresource.EdgeSite,
			expectedErrorCode: codes.InvalidArgument,
			expectedOrder:     nil,
		},
		"InvalidUnknownField": {
			orderBy:           "asdf",
			expectedErrorCode: codes.InvalidArgument,
			expectedOrder:     nil,
		},
		"InvalidTrailingComma": {
			orderBy:           hostresource.EdgeSite + ",",
			expectedErrorCode: codes.InvalidArgument,
			expectedOrder:     nil,
		},
		"InvalidJustComma": {
			orderBy:           ",",
			expectedErrorCode: codes.InvalidArgument,
			expectedOrder:     nil,
		},
		"InvalidMultipleComma": {
			orderBy:           ",,,",
			expectedErrorCode: codes.InvalidArgument,
			expectedOrder:     nil,
		},
		"InvalidWhitespace": {
			orderBy:           " ",
			expectedErrorCode: codes.InvalidArgument,
			expectedOrder:     nil,
		},
	}
	for tcname, tc := range tests {
		t.Run(tcname, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			findresIDs, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				OrderBy:  tc.orderBy,
			})

			if tc.expectedErrorCode != codes.OK {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorCode, status.Code(err))
			} else {
				require.NoError(t, err)
				require.Len(t, findresIDs.Resources, len(tc.expectedOrder))
				var expectedIDs []*client.ResourceTenantIDCarrier
				for _, id := range tc.expectedOrder {
					expectedIDs = append(
						expectedIDs,
						&client.ResourceTenantIDCarrier{TenantId: id.GetTenantId(), ResourceId: id.GetResourceId()})
				}
				assert.Truef(t, reflect.DeepEqual(expectedIDs, findresIDs.Resources),
					"FilterHosts() failed - want: %s, got: %s",
					expectedIDs,
					findresIDs.Resources,
				)
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Host{}},
				OrderBy:  tc.orderBy,
			})

			if tc.expectedErrorCode != codes.OK {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorCode, status.Code(err))
			} else {
				require.NoError(t, err)

				var resources []*computev1.HostResource
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHost())
				}

				require.Len(t, resources, len(tc.expectedOrder))
				for i, expected := range tc.expectedOrder {
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListInstances() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func Test_HostEnumStateMap(t *testing.T) {
	v, err := store.HostEnumStateMap(hostresource.FieldHostStatusIndicator,
		int32(statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS))
	val0, ok0 := v.(hostresource.HostStatusIndicator)
	assert.True(t, ok0)
	assert.Nil(t, err)
	assert.Equal(t, "STATUS_INDICATION_IN_PROGRESS", val0.String())

	v, err = store.HostEnumStateMap(hostresource.FieldRegistrationStatusIndicator,
		int32(statusv1.StatusIndication_STATUS_INDICATION_IDLE))
	val1, ok1 := v.(hostresource.RegistrationStatusIndicator)
	assert.True(t, ok1)
	assert.Nil(t, err)
	assert.Equal(t, "STATUS_INDICATION_IDLE", val1.String())

	v, err = store.HostEnumStateMap(hostresource.FieldOnboardingStatusIndicator,
		int32(statusv1.StatusIndication_STATUS_INDICATION_IDLE))
	val2, ok2 := v.(hostresource.OnboardingStatusIndicator)
	assert.True(t, ok2)
	assert.Nil(t, err)
	assert.Equal(t, "STATUS_INDICATION_IDLE", val2.String())

	v, err = store.HostEnumStateMap("invalid_option",
		int32(statusv1.StatusIndication_STATUS_INDICATION_IDLE))
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestHostMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)

	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				host := dao.CreateHost(t, tenantID)
				res, err := util.WrapResource(host)
				require.NoError(t, err)
				return host.GetResourceId(), res
			},
		},
	})
}

func TestSoftDeleteResources_Hosts(t *testing.T) {
	suite.Run(t, &softDeleteAllResourcesSuite{
		createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
			tenantID := uuid.NewString()
			return tenantID, len([]any{dao.CreateHostNoCleanup(t, tenantID), dao.CreateHostNoCleanup(t, tenantID)})
		},
		resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		deletedClause: filters.ValEq(
			computev1.HostResourceFieldDesiredState, computev1.HostState_HOST_STATE_DELETED),
		notDeletedClause: filters.ValNotEq(
			computev1.HostResourceFieldDesiredState, computev1.HostState_HOST_STATE_DELETED),
	})
}

func TestHardDeleteResources_Hosts(t *testing.T) {
	suite.Run(t, &hardDeleteAllResourcesSuite{
		createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
			tenantID := uuid.NewString()
			return tenantID, len([]any{dao.CreateHostNoCleanup(t, tenantID), dao.CreateHostNoCleanup(t, tenantID)})
		},
		resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_HOST,
	})
}
