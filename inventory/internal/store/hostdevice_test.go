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

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostdeviceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

//nolint:funlen // length due to test cases
func Test_Create_Get_Delete_Hostdevice(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site := inv_testing.CreateSite(t, region, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host := inv_testing.CreateHost(t, site, provider)

	testcases := map[string]struct {
		in    *computev1.HostdeviceResource
		valid bool
	}{
		"CreateGoodHostdevice": {
			in: &computev1.HostdeviceResource{
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
		"CreateBadHostdeviceWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.HostdeviceResource{
				ResourceId:       "hostdevice-12345678",
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
		"CreateBadHostdeviceWithInvalidResourceIdSet": {
			// This tests case verifies that create requests with a invalid resource ID
			// already set are rejected.
			in: &computev1.HostdeviceResource{
				ResourceId:       "host-device-12345678",
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
		"CreateBadHostdevice_NoHostAssociated": {
			in:    &computev1.HostdeviceResource{},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_HostDevice{HostDevice: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create hostdevice
			chostdeviceResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			hostdeviceResID := chostdeviceResp.GetHostDevice().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateHostdevice() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = hostdeviceResID // Update with created resource ID.
				tc.in.CreatedAt = chostdeviceResp.GetHostDevice().GetCreatedAt()
				tc.in.UpdatedAt = chostdeviceResp.GetHostDevice().GetUpdatedAt()
				assertSameResource(t, createresreq, chostdeviceResp, nil)
				if !tc.valid {
					t.Errorf("CreateHostdevice() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "hostdevice-12345678")
				require.Error(t, err)

				// get hostdevice
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, hostdeviceResID)
				require.NoError(t, err, "GetHostdevice() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetHostDevice()); !eq {
					t.Errorf("GetHostDevice() data not equal: %v", diff)
				}

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "hostdevice-12345678")
				require.Error(t, err)

				// delete hostdevice from API
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, hostdeviceResID)
				if err != nil {
					t.Errorf("DeleteHostdevice() failed: %s", err)
				}

				// get after complete Delete of hostdevice, should fail as Hostdevice is 2-phase deleted
				_, err = inv_testing.TestClients[inv_testing.RMClient].Get(ctx, hostdeviceResID)
				require.Error(t, err, "Failure - Hostdevice was not deleted, but should be deleted")
			}
		})
	}
}

//nolint:funlen // length due to test cases
func Test_UpdateHostdevice(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site := inv_testing.CreateSite(t, region, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host := inv_testing.CreateHost(t, site, provider)

	// create Hostdevice to update
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostDevice{
			HostDevice: &computev1.HostdeviceResource{
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
	hostdeviceResID := inv_testing.GetResourceIDOrFail(t, cvmResp)
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostdeviceResID) })

	testcases := map[string]struct {
		in           *computev1.HostdeviceResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdateHostdevice1": {
			in: &computev1.HostdeviceResource{
				DeviceName: "device0",
			},
			resourceID: hostdeviceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostdeviceresource.FieldDeviceName},
			},
			valid: true,
		},
		"UpdateHostdevice2": {
			in: &computev1.HostdeviceResource{
				Version: "some version",
			},
			resourceID: hostdeviceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					hostdeviceresource.FieldVersion,
					hostdeviceresource.FieldDeviceName,
				},
			},
			valid: true,
		},
		"UpdateHostdevice3": {
			in: &computev1.HostdeviceResource{
				Host: host,
			},
			resourceID: hostdeviceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostdeviceresource.FieldDeviceName},
			},
			valid: true,
		},
		"UpdateNoFieldMask": {
			in: &computev1.HostdeviceResource{
				Host:    host,
				Version: "some version",
			},
			resourceID:   hostdeviceResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateInvalidFieldMask": {
			in: &computev1.HostdeviceResource{
				Version: "some version",
			},
			resourceID: hostdeviceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateFieldMaskNonClearableField": {
			in: &computev1.HostdeviceResource{
				ResourceId: "proj-fb123457",
			},
			resourceID: hostdeviceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"resource"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateResourceIDNotFound": {
			in: &computev1.HostdeviceResource{
				DeviceName: "device0",
			},
			resourceID: "hostdevice-12345678",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{hostdeviceresource.FieldDeviceName},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_HostDevice{HostDevice: tc.in},
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
func Test_FilterHostdevices(t *testing.T) {
	region := inv_testing.CreateRegion(t, nil)
	site1 := inv_testing.CreateSite(t, region, nil)
	site2 := inv_testing.CreateSite(t, region, nil)
	provider1 := inv_testing.CreateProvider(t, "Test Provider1")
	provider2 := inv_testing.CreateProvider(t, "Test Provider2")
	host1 := inv_testing.CreateHost(t, site1, provider1)
	host2 := inv_testing.CreateHost(t, site2, provider2)
	host3 := inv_testing.CreateHost(t, nil, nil)

	// create Hostdevices to find
	createresreq1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostDevice{
			HostDevice: &computev1.HostdeviceResource{
				Host:       host1,
				DeviceName: "testhost",
			},
		},
	}

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostDevice{
			HostDevice: &computev1.HostdeviceResource{
				Host:       host2,
				DeviceName: "testhost",
			},
		},
	}

	createresreqEmpty := &inv_v1.Resource{
		Resource: &inv_v1.Resource_HostDevice{
			HostDevice: &computev1.HostdeviceResource{
				Host: host3,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//nolint:errcheck // creation of test client, error does not need to be checked
	chostdeviceResp1, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq1)
	hostdeviceResID1 := inv_testing.GetResourceIDOrFail(t, chostdeviceResp1)
	//nolint:errcheck // creation of test client, error does not need to be checked
	chostdeviceResp2, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	hostdeviceResID2 := inv_testing.GetResourceIDOrFail(t, chostdeviceResp2)
	//nolint:errcheck // creation of test client, error does not need to be checked
	chostdeviceEmpty, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreqEmpty)
	hostdeviceResIDEmpty := inv_testing.GetResourceIDOrFail(t, chostdeviceEmpty)
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostdeviceResID1) })
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostdeviceResID2) })
	t.Cleanup(func() { inv_testing.DeleteResource(t, hostdeviceResIDEmpty) })

	expHostdevice1 := createresreq1.GetHostDevice()
	expHostdevice1.ResourceId = hostdeviceResID1

	expHostdevice2 := createresreq2.GetHostDevice()
	expHostdevice2.ResourceId = hostdeviceResID2

	expHostdeviceEmpty := createresreqEmpty.GetHostDevice()
	expHostdeviceEmpty.ResourceId = hostdeviceResIDEmpty

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.HostdeviceResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.HostdeviceResource{expHostdevice1, expHostdevice2, expHostdeviceEmpty},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: hostdeviceresource.FieldResourceID,
			},
			resources: []*computev1.HostdeviceResource{expHostdevice1, expHostdevice2, expHostdeviceEmpty},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostusb{}},
				Filter:   fmt.Sprintf(`%s = ""`, hostdeviceresource.FieldResourceID),
			},
			resources: []*computev1.HostdeviceResource{},
			valid:     true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, hostdeviceresource.FieldResourceID, expHostdevice2.ResourceId),
			},
			resources: []*computev1.HostdeviceResource{expHostdevice2},
			valid:     true,
		},
		"FilterHost": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, hostdeviceresource.EdgeHost,
					hostresource.FieldResourceID, host2.GetResourceId()),
			},
			resources: []*computev1.HostdeviceResource{expHostdevice2},
			valid:     true,
		},
		"FilterHostEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, hostdeviceresource.EdgeHost),
			},
			valid: true, // HostDevice must have a Host
		},
		"FilterByHasHostHasSite": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s.%s)`, hostdeviceresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.HostdeviceResource{expHostdevice1, expHostdevice2},
			valid:     true,
		},
		"FilterDeviceName": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, hostdeviceresource.FieldDeviceName, expHostdevice2.GetDeviceName()),
			},
			resources: []*computev1.HostdeviceResource{expHostdevice2},
			valid:     true,
		},
		"FilterDeviceNameEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, hostdeviceresource.FieldDeviceName),
			},
			resources: []*computev1.HostdeviceResource{expHostdeviceEmpty},
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
			resources: []*computev1.HostdeviceResource{expHostdevice1, expHostdevice2, expHostdeviceEmpty},
			valid:     true,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_HostDevice{}} // Set the resource kind
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterHostdevice() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterHostdevice() succeeded but should have failed")
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
						"FilterHostdevice() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListHostdevice() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListHostdevice() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.HostdeviceResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHostDevice())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					hostdeviceEdgesOnlyResourceID(expected)
					hostdeviceEdgesOnlyResourceID(resources[i])

					// Skip check of CreatedAt and UpdatedAt.
					resources[i].CreatedAt = expected.CreatedAt
					resources[i].UpdatedAt = expected.UpdatedAt
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListHostdevice() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func hostdeviceEdgesOnlyResourceID(expected *computev1.HostdeviceResource) {
	if expected.Host != nil {
		expected.Host = &computev1.HostResource{ResourceId: expected.Host.ResourceId}
	}
}

//nolint:funlen // length due to test cases
func Test_NestedFilterHostdevice(t *testing.T) {
	region1 := inv_testing.CreateRegion(t, nil)
	site1 := inv_testing.CreateSite(t, region1, nil)
	host1 := inv_testing.CreateHost(t, site1, nil)
	host2 := inv_testing.CreateHost(t, nil, nil)

	hostDevice1 := inv_testing.CreateHostdevice(t, host1)
	hostDevice1.Host = host1
	hostDevice2 := inv_testing.CreateHostdevice(t, host2)
	hostDevice2.Host = host2

	testcases := map[string]struct {
		in                *inv_v1.ResourceFilter
		resources         []*computev1.HostdeviceResource
		valid             bool
		expectedCodeError codes.Code
	}{
		"FilterBySiteID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, hostdeviceresource.EdgeHost, hostresource.EdgeSite,
					siteresource.FieldResourceID, site1.GetResourceId()),
			},
			resources: []*computev1.HostdeviceResource{hostDevice1},
			valid:     true,
		},
		"FilterBySiteEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s.%s)`, hostdeviceresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.HostdeviceResource{hostDevice2},
			valid:     true,
		},
		"FilterByHasSite": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s.%s)`, hostdeviceresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.HostdeviceResource{hostDevice1},
			valid:     true,
		},
		"FilterByRegionID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s = %q`, hostdeviceresource.EdgeHost, hostresource.EdgeSite,
					siteresource.EdgeRegion, regionresource.FieldResourceID, region1.GetResourceId()),
			},
			resources: []*computev1.HostdeviceResource{hostDevice1},
			valid:     true,
		},
		"FailTooDeep": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s.%s.%s = %q`, hostdeviceresource.EdgeHost, hostresource.EdgeSite,
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_HostDevice{}} // Set the resource kind

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

				resources := make([]*computev1.HostdeviceResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetHostDevice())
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

func Test_StrongRelations_On_Delete_HostDevice(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)
	inv_testing.CreateHostdevice(t, host)

	err := inv_testing.HardDeleteHostAndReturnError(t, host.ResourceId)
	assertStrongRelationError(t, err, "violates foreign key constraint")
}

func TestHostDeviceMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				parent := dao.CreateHost(t, tenantID)
				child := dao.CreateHostDevice(t, tenantID, parent)
				res, err := util.WrapResource(child)
				require.NoError(t, err)
				return child.GetResourceId(), res
			},
		},
	})
}

func TestDeleteResources_HostDevices(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				host := dao.CreateHost(t, tenantID)
				return tenantID, len(
					[]any{
						dao.CreateHostDeviceNoCleanup(t, tenantID, host),
						dao.CreateHostDeviceNoCleanup(t, tenantID, host),
						dao.CreateHostDeviceNoCleanup(t, tenantID, host),
					},
				)
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE,
		},
	})
}
