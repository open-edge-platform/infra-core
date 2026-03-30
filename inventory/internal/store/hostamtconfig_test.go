// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostamtconfigresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

//nolint:funlen // length due to test cases
func Test_Create_Get_Delete_Hostamtconfig(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site := inv_testing.CreateSite(t, region, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host := inv_testing.CreateHost(t, site, provider)

	testcases := map[string]struct {
		in    *computev1.HostamtconfigResource
		valid bool
	}{
		"CreateGoodHostamtconfig": {
			in: &computev1.HostamtconfigResource{
				Host:             host,
				Version:          "1.2.34",
				DeviceName:       "testhost",
				OperationalState: "enabled",
				BuildNumber:      "1234",
				Sku:              "5678",
				Features:         "test features",
				DeviceGuid:       "1234abcd-ef56-7890-12ab-34567890cdef",
				ControlMode:      "client",
				DnsSuffix:        "testhost.com",
				NetworkStatus:    "direct",
				RemoteStatus:     "not connected",
				RemoteTrigger:    "user initiated",
				MpsHostname:      "",
			},
			valid: true,
		},
		"CreateBadHostamtconfigWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.HostamtconfigResource{
				ResourceId:       "hostamtconfig-12345678",
				Host:             host,
				Version:          "1.2.34",
				DeviceName:       "testhost",
				OperationalState: "enabled",
				BuildNumber:      "1234",
				Sku:              "5678",
				Features:         "test features",
				DeviceGuid:       "1234abcd-ef56-7890-12ab-34567890cdef",
				ControlMode:      "client",
				DnsSuffix:        "testhost.com",
				NetworkStatus:    "direct",
				RemoteStatus:     "not connected",
				RemoteTrigger:    "user initiated",
				MpsHostname:      "",
			},
			valid: false,
		},
		"CreateBadHostamtconfigWithInvalidResourceIdSet": {
			// This tests case verifies that create requests with a invalid resource ID
			// already set are rejected.
			in: &computev1.HostamtconfigResource{
				ResourceId:       "host-amtconfig-12345678",
				Host:             host,
				Version:          "1.2.34",
				DeviceName:       "testhost",
				OperationalState: "enabled",
				BuildNumber:      "1234",
				Sku:              "5678",
				Features:         "test features",
				DeviceGuid:       "1234abcd-ef56-7890-12ab-34567890cdef",
				ControlMode:      "client",
				DnsSuffix:        "testhost.com",
				NetworkStatus:    "direct",
				RemoteStatus:     "not connected",
				RemoteTrigger:    "user initiated",
				MpsHostname:      "",
			},
			valid: false,
		},
		"CreateBadHostamtconfig_NoHostAssociated": {
			in:    &computev1.HostamtconfigResource{},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_HostAmtconfig{HostAmtconfig: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create hostamtconfig
			chostamtconfigResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			hostamtconfigResID := chostamtconfigResp.GetHostAmtconfig().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateHostamtconfig() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = hostamtconfigResID // Update with created resource ID.
				tc.in.CreatedAt = chostamtconfigResp.GetHostAmtconfig().GetCreatedAt()
				tc.in.UpdatedAt = chostamtconfigResp.GetHostAmtconfig().GetUpdatedAt()
				assertSameResource(t, createresreq, chostamtconfigResp, nil)
				if !tc.valid {
					t.Errorf("CreateHostamtconfig() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "hostamtconfig-12345678")
				require.Error(t, err)

				// get hostamtconfig
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, hostamtconfigResID)
				require.NoError(t, err, "GetHostamtconfig() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetHostAmtconfig()); !eq {
					t.Errorf("GetHostAmtconfig() data not equal: %v", diff)
				}

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "hostamtconfig-12345678")
				require.Error(t, err)

				// delete hostamtconfig from API
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, hostamtconfigResID)
				if err != nil {
					t.Errorf("DeleteHostamtconfig() failed: %s", err)
				}

				// get after complete Delete of hostamtconfig, should fail as Hostamtconfig is 2-phase deleted
				_, err = inv_testing.TestClients[inv_testing.RMClient].Get(ctx, hostamtconfigResID)
				require.Error(t, err, "Failure - Hostamtconfig was not deleted, but should be deleted")
			}
		})
	}
}

//nolint:funlen // length due to test cases
func Test_UpdateHostamtconfig(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site := inv_testing.CreateSite(t, region, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host := inv_testing.CreateHost(t, site, provider)

	// create Hostamtconfig to update
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostAmtconfig{
			HostAmtconfig: &computev1.HostamtconfigResource{
				Host:             host,
				Version:          "1.2.34",
				DeviceName:       "testhost",
				OperationalState: "enabled",
				BuildNumber:      "1234",
				Sku:              "5678",
				Features:         "test features",
				DeviceGuid:       "1234abcd-ef56-7890-12ab-34567890cdef",
				ControlMode:      "client",
				DnsSuffix:        "testhost.com",
				NetworkStatus:    "direct",
				RemoteStatus:     "not connected",
				RemoteTrigger:    "user initiated",
				MpsHostname:      "",
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cvmResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
	require.NoError(t, err)
	hostamtconfigResID := inv_testing.GetResourceIDOrFail(t, cvmResp)
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostamtconfigResID) })

	testcases := map[string]struct {
		in           *computev1.HostamtconfigResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdateHostamtconfig1": {
			in: &computev1.HostamtconfigResource{
				DeviceName: "amtconfig0",
			},
			resourceID: hostamtconfigResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostamtconfigresource.FieldDeviceName},
			},
			valid: true,
		},
		"UpdateHostamtconfig2": {
			in: &computev1.HostamtconfigResource{
				Version: "some version",
			},
			resourceID: hostamtconfigResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostamtconfigresource.FieldVersion,
					hostamtconfigresource.FieldDeviceName,
				},
			},
			valid: true,
		},
		"UpdateHostamtconfig3": {
			in: &computev1.HostamtconfigResource{
				Host: host,
			},
			resourceID: hostamtconfigResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostamtconfigresource.FieldDeviceName},
			},
			valid: true,
		},
		"UpdateNoFieldMask": {
			in: &computev1.HostamtconfigResource{
				Host:    host,
				Version: "some version",
			},
			resourceID:   hostamtconfigResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateInvalidFieldMask": {
			in: &computev1.HostamtconfigResource{
				Version: "some version",
			},
			resourceID: hostamtconfigResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateFieldMaskNonClearableField": {
			in: &computev1.HostamtconfigResource{
				ResourceId: "proj-fb123457",
			},
			resourceID: hostamtconfigResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"resource"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateResourceIDNotFound": {
			in: &computev1.HostamtconfigResource{
				DeviceName: "amtconfig0",
			},
			resourceID: "hostamtconfig-12345678",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostamtconfigresource.FieldDeviceName},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_HostAmtconfig{HostAmtconfig: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := inv_testing.TestClients[inv_testing.RMClient].Update(
				ctx,
				tc.resourceID,
				tc.fieldMask,
				updateresreq,
			)

			if !tc.valid {
				require.Errorf(t, err, "UpdateResource() succeeded but should have failed")
				assert.Equal(t, tc.expErrorCode, status.Code(err))
				assert.Nil(t, upRes)
				return
			}
			require.NoErrorf(t, err, "UpdateResource() failed: %s", err)

			// validate update via a get
			getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, tc.resourceID)
			require.NoError(t, err, "GetResource() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

//nolint:cyclop,funlen // high cyclomatic complexity and length due to number of tests
func Test_FilterHostamtconfigs(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site1 := inv_testing.CreateSite(t, region, nil)
	site2 := inv_testing.CreateSite(t, region, nil)
	provider1 := inv_testing.CreateProvider(t, "Test Provider1")
	provider2 := inv_testing.CreateProvider(t, "Test Provider2")
	host1 := inv_testing.CreateHost(t, site1, provider1)
	host2 := inv_testing.CreateHost(t, site2, provider2)
	host3 := inv_testing.CreateHost(t, nil, nil)

	// create Hostamtconfigs to find
	createresreq1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostAmtconfig{
			HostAmtconfig: &computev1.HostamtconfigResource{
				Host:       host1,
				DeviceName: "testhost",
			},
		},
	}

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostAmtconfig{
			HostAmtconfig: &computev1.HostamtconfigResource{
				Host:       host2,
				DeviceName: "testhost1",
			},
		},
	}

	createresreqEmpty := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostAmtconfig{
			HostAmtconfig: &computev1.HostamtconfigResource{
				Host: host3,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//nolint:errcheck // creation of test client, error does not need to be checked
	chostamtconfigResp1, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq1)
	hostamtconfigResID1 := inv_testing.GetResourceIDOrFail(t, chostamtconfigResp1)
	//nolint:errcheck // creation of test client, error does not need to be checked
	chostamtconfigResp2, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	hostamtconfigResID2 := inv_testing.GetResourceIDOrFail(t, chostamtconfigResp2)
	//nolint:errcheck // creation of test client, error does not need to be checked
	chostamtconfigEmpty, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreqEmpty)
	hostamtconfigResIDEmpty := inv_testing.GetResourceIDOrFail(t, chostamtconfigEmpty)
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostamtconfigResID1) })
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostamtconfigResID2) })
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostamtconfigResIDEmpty) })

	expHostamtconfig1 := createresreq1.GetHostAmtconfig()
	expHostamtconfig1.ResourceId = hostamtconfigResID1

	expHostamtconfig2 := createresreq2.GetHostAmtconfig()
	expHostamtconfig2.ResourceId = hostamtconfigResID2

	expHostamtconfigEmpty := createresreqEmpty.GetHostAmtconfig()
	expHostamtconfigEmpty.ResourceId = hostamtconfigResIDEmpty

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.HostamtconfigResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig1, expHostamtconfig2, expHostamtconfigEmpty},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: hostamtconfigresource.FieldResourceID,
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig1, expHostamtconfig2, expHostamtconfigEmpty},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostusb{}},
				Filter:   fmt.Sprintf(`%s = ""`, hostamtconfigresource.FieldResourceID),
			},
			resources: []*computev1.HostamtconfigResource{},
			valid:     true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, hostamtconfigresource.FieldResourceID, expHostamtconfig2.ResourceId),
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig2},
			valid:     true,
		},
		"FilterHost": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostamtconfigresource.EdgeHost,
					hostresource.FieldResourceID, host2.GetResourceId()),
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig2},
			valid:     true,
		},
		"FilterHostEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostamtconfigresource.EdgeHost),
			},
			valid: true, // HostAmtconfig must have a Host
		},
		"FilterByHasHostHasSite": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s.%s)`, hostamtconfigresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig1, expHostamtconfig2},
			valid:     true,
		},
		"FilterDeviceName": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, hostamtconfigresource.FieldDeviceName, expHostamtconfig2.GetDeviceName()),
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig2},
			valid:     true,
		},
		"FilterDeviceNameEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, hostamtconfigresource.FieldDeviceName),
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfigEmpty},
			valid:     true,
		},
		"FilterInvalidField": {
			in: &inv_v1.ResourceFilter{
				Filter: `invalid_field = "foo"`,
			},
			valid: false,
		},
		"FilterInvalidEdge": {
			in: &inv_v1.ResourceFilter{
				Filter: `has(invalid_edge)`,
			},
			valid: false,
		},
		"FilterWithOffsetLimit1": {
			in: &inv_v1.ResourceFilter{
				Offset: 5,
				Limit:  0,
			},
			valid: true,
		},
		"FilterWithOffsetLimit2": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  5,
			},
			resources: []*computev1.HostamtconfigResource{expHostamtconfig1, expHostamtconfig2, expHostamtconfigEmpty},
			valid:     true,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_HostAmtconfig{}} // Set the resource kind
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterHostamtconfig() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterHostamtconfig() succeeded but should have failed")
				}
			}

			// only get/delete if valid test with non-zero returned response and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				if len(findres.Resources) != len(tc.resources) {
					t.Errorf("Expected to obtain %d Resource IDs, but obtained back %d Resource IDs",
						len(tc.resources), len(findres.Resources))
				}

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findres.Resources)

				if !reflect.DeepEqual(resIDs, findres.Resources) {
					t.Errorf(
						"FilterHostamtconfig() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListHostamtconfig() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListHostamtconfig() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.HostamtconfigResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHostAmtconfig())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					hostamtconfigEdgesOnlyResourceID(expected)
					hostamtconfigEdgesOnlyResourceID(resources[i])

					// Skip check of CreatedAt and UpdatedAt.
					resources[i].CreatedAt = expected.CreatedAt
					resources[i].UpdatedAt = expected.UpdatedAt
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListHostamtconfig() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func hostamtconfigEdgesOnlyResourceID(expected *computev1.HostamtconfigResource) {
	if expected.Host != nil {
		expected.Host = &computev1.HostResource{ResourceId: expected.Host.ResourceId}
	}
}

//nolint:funlen // length due to test cases
func Test_NestedFilterHostamtconfig(t *testing.T) {
	region1 := inv_testing.CreateRegion(t, nil)
	site1 := inv_testing.CreateSite(t, region1, nil)
	host1 := inv_testing.CreateHost(t, site1, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)

	hostAmtconfig1 := inv_testing.CreateHostamtconfig(t, host1)
	hostAmtconfig1.Host = host1
	hostAmtconfig2 := inv_testing.CreateHostamtconfig(t, host2)
	hostAmtconfig2.Host = host2

	testcases := map[string]struct {
		in                *inv_v1.ResourceFilter
		resources         []*computev1.HostamtconfigResource
		valid             bool
		expectedCodeError codes.Code
	}{
		"FilterBySiteID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostamtconfigresource.EdgeHost, hostresource.EdgeSite,
					siteresource.FieldResourceID, site1.GetResourceId()),
			},
			resources: []*computev1.HostamtconfigResource{hostAmtconfig1},
			valid:     true,
		},
		"FilterBySiteEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s.%s)`, hostamtconfigresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.HostamtconfigResource{hostAmtconfig2},
			valid:     true,
		},
		"FilterByHasSite": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s.%s)`, hostamtconfigresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.HostamtconfigResource{hostAmtconfig1},
			valid:     true,
		},
		"FilterByRegionID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s = %q`, hostamtconfigresource.EdgeHost, hostresource.EdgeSite,
					siteresource.EdgeRegion, regionresource.FieldResourceID, region1.GetResourceId()),
			},
			resources: []*computev1.HostamtconfigResource{hostAmtconfig1},
			valid:     true,
		},
		"FailTooDeep": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s.%s.%s = %q`, hostamtconfigresource.EdgeHost, hostresource.EdgeSite,
					siteresource.EdgeRegion, regionresource.EdgeParentRegion, regionresource.EdgeParentRegion,
					regionresource.FieldResourceID, region1.GetResourceId()),
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_HostAmtconfig{}} // Set the resource kind

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

				resources := make([]*computev1.HostamtconfigResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHostAmtconfig())
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

func Test_StrongRelations_On_Delete_HostAmtconfig(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)
	inv_testing.CreateHostamtconfig(t, host)

	err := inv_testing.HardDeleteHostAndReturnError(t, host.ResourceId)
	assertStrongRelationError(t, err, "violates foreign key constraint")
}

func TestHostAmtconfigMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				parent := dao.CreateHost(t, tenantID)
				child := dao.CreateHostAmtconfig(t, tenantID, parent)
				res, err := util.WrapResource(child)
				require.NoError(t, err)
				return child.GetResourceId(), res
			},
		},
	})
}

func TestDeleteResources_HostAmtconfigs(t *testing.T) {
	suite.Run(t, &hardDeleteAllResourcesSuite{
		createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
			tenantID := uuid.NewString()
			host := dao.CreateHost(t, tenantID)
			return tenantID, len([]any{
				dao.CreateHostAmtconfigNoCleanup(t, tenantID, host),
			})
		},
		resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_HOSTAMTCONFIG,
	})
}
