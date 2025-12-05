// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
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

	our "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdaterunresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

//nolint:funlen // test function is long.
func Test_Create_Get_Delete_Update_OSUpdateRun(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	host := dao.CreateHost(t, tenantID)
	os := dao.CreateOs(t, tenantID)
	instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
	osUpdatePolicy := dao.CreateOSUpdatePolicy(
		t, tenantID,
		inv_testing.OsUpdatePolicyName("test-policy"),
		inv_testing.OSUpdatePolicyLatest())
	testcases := map[string]struct {
		in    *computev1.OSUpdateRunResource
		valid bool
	}{
		"CreateGoodOsUpdateRunTargetMut": {
			in: &computev1.OSUpdateRunResource{
				Name:            "Test OS Update Run",
				Description:     "Test Description",
				AppliedPolicy:   osUpdatePolicy,
				Instance:        instance,
				StartTime:       uint64(time.Now().Unix()), //nolint:gosec // This is a test
				StatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS,
				StatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			valid: true,
		},
		"CreateBadOsUpdateRunPolicyMissing": {
			in: &computev1.OSUpdateRunResource{
				Name:        "Test OS Update Run",
				Description: "Test Description",
				Instance:    instance,
				StartTime:   uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			valid: false,
		},
		"CreateBadOsUpdateRunInstanceMissing": {
			in: &computev1.OSUpdateRunResource{
				Name:          "Test OS Update Run",
				Description:   "Test Description",
				AppliedPolicy: osUpdatePolicy,
				StartTime:     uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			valid: false,
		},
		"CreateBadOsUpdateRunStartTimeMissing": {
			in: &computev1.OSUpdateRunResource{
				Name:          "Test OS Update Run",
				Description:   "Test Description",
				AppliedPolicy: osUpdatePolicy,
				Instance:      instance,
			},
			valid: false,
		},
		"CreateBadOsUpdateRunEndTimePresent": {
			in: &computev1.OSUpdateRunResource{
				Name:          "Test OS Update Run",
				Description:   "Test Description",
				AppliedPolicy: osUpdatePolicy,
				Instance:      instance,
				StartTime:     uint64(time.Now().Unix()),                //nolint:gosec // This is a test
				EndTime:       uint64(time.Now().Add(time.Hour).Unix()), //nolint:gosec // This is a test
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			tc.in.TenantId = tenantID
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_OsUpdateRun{OsUpdateRun: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
			defer cancel()

			invClient := inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient()

			// create
			cupdatesourceResp, err := invClient.Create(ctx, tenantID, createresreq)
			ourResID := cupdatesourceResp.GetOsUpdateRun().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateOsUpdateRun() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = ourResID // Update with created resource ID.
				tc.in.CreatedAt = cupdatesourceResp.GetOsUpdateRun().GetCreatedAt()
				tc.in.UpdatedAt = cupdatesourceResp.GetOsUpdateRun().GetUpdatedAt()
				assertSameResource(t, createresreq, cupdatesourceResp, nil)
				if !tc.valid {
					t.Errorf("CreateOsUpdateRun() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := invClient.Get(ctx, tenantID, "osupdaterun-12345678")
				require.Error(t, err)

				// get
				getresp, err := invClient.Get(ctx, tenantID, ourResID)
				require.NoError(t, err, "GetOSUpdateRun() failed")

				// verify data
				tc.in.CreatedAt = getresp.GetResource().GetOsUpdateRun().GetCreatedAt()
				tc.in.UpdatedAt = getresp.GetResource().GetOsUpdateRun().GetUpdatedAt()
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetOsUpdateRun()); !eq {
					t.Errorf("GetOSUpdateRun() data not equal: %v", diff)
				}

				// update
				updateresreq := &inv_v1.Resource{
					Resource: &inv_v1.Resource_OsUpdateRun{
						OsUpdateRun: &computev1.OSUpdateRunResource{
							StatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
							StatusDetails:   "Updated details",
						},
					},
				}
				fieldMask := &fieldmaskpb.FieldMask{
					Paths: []string{our.FieldStatusTimestamp, our.FieldStatusDetails},
				}
				upRes, err := invClient.Update(
					ctx,
					tenantID,
					tc.in.ResourceId,
					fieldMask,
					updateresreq,
				)
				if err != nil {
					t.Errorf("UpdateOSUpdateRun() failed: %s", err)
				}

				assertSameResource(t, updateresreq, upRes, fieldMask)

				// delete non-existent first
				_, err = invClient.Delete(ctx, tenantID, "osupdaterun-12345678")
				require.Error(t, err)

				// delete
				_, err = invClient.Delete(
					ctx,
					tenantID,
					ourResID,
				)
				if err != nil {
					t.Errorf("DeleteOsUpdateRun() failed %s", err)
				}

				_, err = invClient.Get(ctx, tenantID, ourResID)
				require.Error(t, err, "Failure - OSUpdateRun was not deleted, but should be deleted")
			}
		})
	}
}

//nolint:funlen,cyclop // test function is long and complex.
func Test_FilterOSUpdateRuns(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	host := dao.CreateHost(t, tenantID)
	os := dao.CreateOs(t, tenantID)
	instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
	osUpdatePolicy := dao.CreateOSUpdatePolicy(
		t, tenantID,
		inv_testing.OsUpdatePolicyName("test-policy"),
		inv_testing.OSUpdatePolicyTarget(),
		inv_testing.OSUpdatePolicyTargetOS(os))
	osUpRun1 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test1"), inv_testing.OsUpdateRunDescription("test1 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Unix())), //nolint:gosec // This is a test
	)
	osUpRun2 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test2"), inv_testing.OsUpdateRunDescription("test2 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS),
		inv_testing.OSUpdateRunStatus("In Progress"),
		inv_testing.OSUpdateRunStatusDetails("In Progress details"),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Add(time.Hour).Unix()))) //nolint:gosec // This is a test
	osUpRun3 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test3"), inv_testing.OsUpdateRunDescription("test3 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IDLE),
		inv_testing.OSUpdateRunStatus("Success"),
		inv_testing.OSUpdateRunStatusDetails("Success details"),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Add(time.Hour).Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunEndTime(uint64(time.Now().Add(time.Hour).Unix())))         //nolint:gosec // This is a test
	osUpRun4 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test4"), inv_testing.OsUpdateRunDescription("test4 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_ERROR),
		inv_testing.OSUpdateRunStatus("Failed"),
		inv_testing.OSUpdateRunStatusDetails("Failed details"),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Add(time.Hour).Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunEndTime(uint64(time.Now().Add(time.Hour).Unix())))         //nolint:gosec // This is a test

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.OSUpdateRunResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.OSUpdateRunResource{osUpRun1, osUpRun2, osUpRun3, osUpRun4},
			valid:     true,
		},
		"NoFilterOrderByName": {
			in: &inv_v1.ResourceFilter{
				OrderBy: our.FieldName,
			},
			resources: []*computev1.OSUpdateRunResource{osUpRun1, osUpRun2, osUpRun3, osUpRun4},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostusb{}},
				Filter:   fmt.Sprintf(`%s = ""`, our.FieldResourceID),
			},
			resources: []*computev1.OSUpdateRunResource{},
			valid:     true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, our.FieldResourceID, osUpRun1.GetResourceId()),
			},
			resources: []*computev1.OSUpdateRunResource{osUpRun1},
			valid:     true,
		},
		"FilterStatusIndicator": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %s`,
					our.FieldStatusIndicator, statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS),
			},
			resources: []*computev1.OSUpdateRunResource{osUpRun1, osUpRun2},
			valid:     true,
		},
		"FilterStatus": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, our.FieldStatus, osUpRun3.GetStatus()),
			},
			resources: []*computev1.OSUpdateRunResource{osUpRun3},
			valid:     true,
		},
		"FilterStatusDetails": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, our.FieldStatusDetails, osUpRun3.GetStatusDetails()),
			},
			resources: []*computev1.OSUpdateRunResource{osUpRun3},
			valid:     true,
		},
		// Unsupported filters commented out as they are not implemented yet.
		// We cannot filter on timestamp, because they are implemented as timestamp, and we provide filter
		// using ILIKE, that's not supported on TIMESTAMP fields.
		// "FilterStatusTimestamp": {
		//	in: &inv_v1.ResourceFilter{
		//		Filter: fmt.Sprintf(`%s = %q`, our.FieldStatusTimestamp, osUpRun3.GetStatusTimestamp()),
		//	},
		//	resources: []*computev1.OSUpdateRunResource{osUpRun3},
		//	valid:     true,
		// },
		// "FilterStartTime": {
		//	in: &inv_v1.ResourceFilter{
		//		Filter: fmt.Sprintf(`%s = %q`, our.FieldStartTime, osUpRun3.GetStartTime()),
		//	},
		//	resources: []*computev1.OSUpdateRunResource{osUpRun3},
		//	valid:     true,
		// },
		// "FilterEndTime": {
		//	in: &inv_v1.ResourceFilter{
		//		Filter: fmt.Sprintf(`%s = %q`, our.FieldEndTime, osUpRun3.GetEndTime()),
		//	},
		//	resources: []*computev1.OSUpdateRunResource{osUpRun3},
		//	valid:     true,
		// },
		"FilterLimit": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  5,
			},
			resources: []*computev1.OSUpdateRunResource{osUpRun1, osUpRun2, osUpRun3, osUpRun4},
			valid:     true,
		},
		"FilterWithOffsetLimit1": {
			in: &inv_v1.ResourceFilter{
				Offset: 5,
				Limit:  0,
			},
			valid: true,
		},
		"FilterInvalidEdge": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, "invalid_edge"),
			},
			valid: false,
		},
		"FilterInvalidField": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, "invalid_field", "some-value"),
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			invClient := inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient()
			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_OsUpdateRun{}} // Set the resource kind
			findres, err := invClient.Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterOss() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterOss() succeeded but should have failed")
				}
			}

			// only get/delete if valid test with non-zero returned response and hasn't failed, otherwise may segfault
			if !t.Failed() && tc.valid {
				if len(findres.Resources) != len(tc.resources) {
					t.Errorf("Expected to obtain %d Resource IDs, but obtained back %d Resource IDs",
						len(tc.resources), len(findres.Resources))
				}

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findres.Resources)

				if !reflect.DeepEqual(resIDs, findres.Resources) {
					t.Errorf(
						"FilterOss() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := invClient.List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListOSUpdateRuns() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListOSUpdateRuns() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.OSUpdateRunResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetOsUpdateRun())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					// Normalize edge fields in the actual retrieved resources to match expected format
					// (expected resources from DAO have nil edge fields, actual resources have them populated)
					actual := resources[i]
					osUpdateRunEdgesOnlyResourceID(actual)

					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, actual); !eq {
						t.Errorf("ListOss() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

// osUpdateRunEdgesOnlyResourceID normalizes OSUpdateRunResource edge fields
// to nil, matching the format returned by the DAO helper functions which
// strip embedded edge objects to avoid nested message comparisons.
func osUpdateRunEdgesOnlyResourceID(resource *computev1.OSUpdateRunResource) {
	// Set to nil to match DAO helper format (see testing_utils.go createOSUpdateRun)
	resource.AppliedPolicy = nil
	resource.Instance = nil
}

//nolint:funlen // test function is long.
func Test_ImmutableFieldsOnUpdateOsUpdateRun(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	host := dao.CreateHost(t, tenantID)
	host2 := dao.CreateHost(t, tenantID)
	os := dao.CreateOs(t, tenantID)
	instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
	instance2 := dao.CreateInstance(t, tenantID, host2, os)
	osUpdatePolicy := dao.CreateOSUpdatePolicy(
		t, tenantID,
		inv_testing.OsUpdatePolicyName("test-policy"),
		inv_testing.OSUpdatePolicyTarget(),
		inv_testing.OSUpdatePolicyTargetOS(os))
	osUpRun1 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test1"), inv_testing.OsUpdateRunDescription("test1 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Add(time.Hour).Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Add(time.Hour).Unix()))) //nolint:gosec // This is a test
	osUpRun2 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test2"), inv_testing.OsUpdateRunDescription("test2 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Add(time.Hour).Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS),
		inv_testing.OSUpdateRunStatus("In Progress"),
		inv_testing.OSUpdateRunStatusDetails("In Progress details"),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Add(time.Hour).Unix()))) //nolint:gosec // This is a test
	osUpRun3 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test3"), inv_testing.OsUpdateRunDescription("test3 description"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Add(time.Hour).Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IDLE),
		inv_testing.OSUpdateRunStatus("Success"),
		inv_testing.OSUpdateRunStatusDetails("Success details"),
		inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Add(time.Hour).Unix())), //nolint:gosec // This is a test
		inv_testing.OSUpdateRunEndTime(uint64(time.Now().Add(time.Hour).Unix())))         //nolint:gosec // This is a test

	testcases := map[string]struct {
		in           *computev1.OSUpdateRunResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdateImmutableName": {
			in: &computev1.OSUpdateRunResource{
				Name: "New Name",
			},
			resourceID:   osUpRun1.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{our.FieldName}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableDescription": {
			in: &computev1.OSUpdateRunResource{
				Description: "New Description",
			},
			resourceID:   osUpRun1.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{our.FieldDescription}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableAppliedPolicy": {
			in: &computev1.OSUpdateRunResource{
				AppliedPolicy: dao.CreateOSUpdatePolicy(
					t, tenantID,
					inv_testing.OsUpdatePolicyName("applied-policy-update-policy"),
					inv_testing.OSUpdatePolicyTarget(),
					inv_testing.OSUpdatePolicyTargetOS(os)),
			},
			resourceID:   osUpRun1.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{our.EdgeAppliedPolicy}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableInstance": {
			in: &computev1.OSUpdateRunResource{
				Name:     "TESTONE",
				Instance: instance2,
			},
			resourceID:   osUpRun2.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{our.EdgeInstance}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableStartTime": {
			in: &computev1.OSUpdateRunResource{
				StartTime: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			resourceID:   osUpRun3.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{our.FieldStartTime}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateStatusIndicator": {
			in: &computev1.OSUpdateRunResource{
				StatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
			},
			resourceID: osUpRun3.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{our.FieldStatusIndicator}},
			valid:      true,
		},
		"UpdateStatus": {
			in: &computev1.OSUpdateRunResource{
				Status: "Good Status",
			},
			resourceID: osUpRun3.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{our.FieldStatus}},
			valid:      true,
		},
		"UpdateStatusDetails": {
			in: &computev1.OSUpdateRunResource{
				StatusDetails: "Good Status Details",
			},
			resourceID: osUpRun3.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{our.FieldStatusDetails}},
			valid:      true,
		},
		"UpdateStatusTimestamp": {
			in: &computev1.OSUpdateRunResource{
				StatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			resourceID: osUpRun3.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{our.FieldStatusTimestamp}},
			valid:      true,
		},
		"UpdateEndTime": {
			in: &computev1.OSUpdateRunResource{
				EndTime: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			resourceID: osUpRun3.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{our.FieldEndTime}},
			valid:      true,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_OsUpdateRun{OsUpdateRun: tc.in},
			}
			invClient := inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient()
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := invClient.Update(ctx, tenantID, tc.resourceID,
				tc.fieldMask, updateresreq)

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
			getresp, err := invClient.Get(ctx, tenantID, tc.resourceID)
			require.NoError(t, err, "GetResource() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

func TestOSUpdateRunMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)

	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				host := dao.CreateHost(t, tenantID)
				os := dao.CreateOs(t, tenantID)
				instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
				osUpdatePolicy := dao.CreateOSUpdatePolicy(
					t, tenantID,
					inv_testing.OsUpdatePolicyName("test-policy"),
					inv_testing.OSUpdatePolicyTarget(),
					inv_testing.OSUpdatePolicyTargetOS(os))
				oup := dao.CreateOSUpdateRun(
					t, tenantID, inv_testing.OsUpdateRunName("OsRun1"),
					inv_testing.OsUpdateRunDescription("OsRun1 description"),
					inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
					inv_testing.OSUpdateRunInstance(instance),
					inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
					inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IDLE),
					inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Unix())), //nolint:gosec // This is a test
				)
				res, err := util.WrapResource(oup)
				require.NoError(t, err)
				return oup.GetResourceId(), res
			},
		},
	})
}

func TestDeleteResources_OSUpdateRuns(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				host := dao.CreateHost(t, tenantID)
				os := dao.CreateOs(t, tenantID)
				instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
				osUpdatePolicy := dao.CreateOSUpdatePolicy(
					t, tenantID,
					inv_testing.OsUpdatePolicyName("test-policy"),
					inv_testing.OSUpdatePolicyTarget(),
					inv_testing.OSUpdatePolicyTargetOS(os),
				)
				return tenantID, len(
					[]any{
						dao.CreateOSUpdateRunNoCleanup(
							t, tenantID,
							inv_testing.OsUpdateRunName("OsRun1"),
							inv_testing.OsUpdateRunDescription("OsRun1 description"),
							inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
							inv_testing.OSUpdateRunInstance(instance),
							inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
							inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IDLE),
							inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Unix()))), //nolint:gosec // This is a test
						dao.CreateOSUpdateRunNoCleanup(
							t, tenantID,
							inv_testing.OsUpdateRunName("OsRun2"),
							inv_testing.OsUpdateRunDescription("OsRun2 description"),
							inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
							inv_testing.OSUpdateRunInstance(instance),
							inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
							inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IDLE),
							inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Unix()))), //nolint:gosec // This is a test
						dao.CreateOSUpdateRunNoCleanup(
							t, tenantID,
							inv_testing.OsUpdateRunName("OsRun3"),
							inv_testing.OsUpdateRunDescription("OsRun2 description"),
							inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
							inv_testing.OSUpdateRunInstance(instance),
							inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())), //nolint:gosec // This is a test
							inv_testing.OSUpdateRunStatusIndicator(statusv1.StatusIndication_STATUS_INDICATION_IDLE),
							inv_testing.OSUpdateRunStatusTimestamp(uint64(time.Now().Unix()))), //nolint:gosec // This is a test
					},
				)
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN,
		},
	})
}
