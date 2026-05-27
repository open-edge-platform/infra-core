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

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/workloadmember"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/workloadresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

func Test_Create_Get_Delete_Update_WorkloadMember(t *testing.T) {
	workload := inv_testing.CreateWorkload(t)
	os := inv_testing.CreateOs(t)
	instance := inv_testing.CreateInstance(t, nil, os)

	testcases := map[string]struct {
		in     *computev1.WorkloadMember
		valid  bool
		errVal require.ErrorAssertionFunc
	}{
		"CreateGoodWorkloadMember": {
			in: &computev1.WorkloadMember{
				Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
				Workload: workload,
				Instance: instance,
			},
			valid: true,
		},
		"CreateBadEmptyWorkloadMember": {
			// Workload member should always be populated
			in:    &computev1.WorkloadMember{},
			valid: false,
			errVal: func(t require.TestingT, err error, _ ...interface{}) {
				s := status.Convert(err)
				require.NotNil(t, s)
				assert.Equal(t, codes.InvalidArgument, s.Code())
				assert.Contains(t, s.Message(), "missing required field")
			},
		},
		"CreateBadWorkloadMemberWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.WorkloadMember{
				ResourceId: "workloadmember-12345678",
				Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			},
			valid: false,
			errVal: func(t require.TestingT, err error, _ ...interface{}) {
				s := status.Convert(err)
				require.NotNil(t, s)
				assert.Equal(t, codes.InvalidArgument, s.Code())
				assert.Contains(t, s.Message(), "resource ID can't be set")
			},
		},
		"CreateBadWorkloadMemberWithInvalidResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.WorkloadMember{
				ResourceId: "Workloadmember-12345678",
				Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			},
			valid: false,
			errVal: func(t require.TestingT, err error, _ ...interface{}) {
				s := status.Convert(err)
				require.NotNil(t, s)
				assert.Equal(t, codes.InvalidArgument, s.Code())
				assert.Contains(t, s.Message(), "does not match regex pattern `^workloadmember-[0-9a-f]{8}$`")
			},
		},
		"CreateBadWorkloadMemberWithTooLongResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.WorkloadMember{
				ResourceId: "workload-member-12345678",
				Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			},
			valid: false,
			errVal: func(t require.TestingT, err error, _ ...interface{}) {
				s := status.Convert(err)
				require.NotNil(t, s)
				assert.Equal(t, codes.InvalidArgument, s.Code())
				assert.Contains(t, s.Message(), "must be at most 23 bytes")
			},
		},
		"CreateBadWorkloadMemberWithNonExistingWorkload": {
			in: &computev1.WorkloadMember{
				Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
				Workload: &computev1.WorkloadResource{
					ResourceId: "workload-12345678",
				},
			},
			valid: false,
			errVal: func(t require.TestingT, err error, _ ...interface{}) {
				s := status.Convert(err)
				require.NotNil(t, s)
				assert.Equal(t, codes.NotFound, s.Code())
				assert.Contains(t, s.Message(), "workload_resource not found")
			},
		},
		"CreateBadWorkloadMemberWithNonExistingInstance": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.WorkloadMember{
				Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
				Instance: &computev1.InstanceResource{
					ResourceId: "inst-12345678",
				},
			},
			valid: false,
			errVal: func(t require.TestingT, err error, _ ...interface{}) {
				s := status.Convert(err)
				require.NotNil(t, s)
				assert.Equal(t, codes.NotFound, s.Code())
				assert.Contains(t, s.Message(), "instance_resource not found")
			},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: tc.in},
			}
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create
			cmemberResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			wmemberResID := cmemberResp.GetWorkloadMember().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateWorkloadMember() failed: %s", err)
				}
				tc.errVal(t, err)
			} else {
				tc.in.ResourceId = wmemberResID // Update with created resource ID.
				tc.in.CreatedAt = cmemberResp.GetWorkloadMember().GetCreatedAt()
				tc.in.UpdatedAt = cmemberResp.GetWorkloadMember().GetUpdatedAt()
				assertSameResource(t, createresreq, cmemberResp, nil)
				if !tc.valid {
					t.Errorf("CreateWorkloadMember() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "workloadmember-12345678")
				require.Error(t, err)

				// get
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, wmemberResID)
				if err != nil {
					require.NoError(t, err, "GetWorkloadMember() failed")
				}

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetWorkloadMember()); !eq {
					t.Errorf("GetWorkloadMember() data not equal: %v", diff)
				}

				// update
				updateresreq := &inv_v1.Resource{
					Resource: &inv_v1.Resource_WorkloadMember{
						WorkloadMember: &computev1.WorkloadMember{
							Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
						},
					},
				}

				fm := &fieldmaskpb.FieldMask{Paths: []string{workloadmember.FieldKind}}
				upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(
					ctx,
					wmemberResID,
					fm,
					updateresreq,
				)
				if err != nil {
					t.Errorf("UpdateWorkloadMember() failed: %s", err)
				}

				// Validate returned resource
				assertSameResource(t, updateresreq, upRes, fm)

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "workloadmember-12345678")
				require.Error(t, err)

				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(
					ctx,
					wmemberResID,
				)
				if err != nil {
					t.Errorf("DeleteWorkloadMember() failed %s", err)
				}

				// get after complete Delete of workload, should fail as workload is 2-phase deleted
				_, err = inv_testing.TestClients[inv_testing.APIClient].Get(ctx, wmemberResID)
				if err == nil {
					t.Errorf("Failure - WorkloadMember was not deleted, but should be deleted")
				}
			}
		})
	}
}

func Test_FilterWorkloadMember(t *testing.T) {
	os := inv_testing.CreateOs(t)
	workload1 := inv_testing.CreateWorkload(t)
	instance1 := inv_testing.CreateInstance(t, nil, os)
	workload2 := inv_testing.CreateWorkload(t)
	instance2 := inv_testing.CreateInstance(t, nil, os)
	member1 := inv_testing.CreateWorkloadMember(t, workload1, instance1)
	member1.Workload = workload1
	member1.Instance = instance1
	member2 := inv_testing.CreateWorkloadMember(t, workload2, instance2)
	member2.Workload = workload2
	member2.Instance = instance2

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.WorkloadMember
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.WorkloadMember{member1, member2},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: workloadmember.FieldResourceID,
			},
			resources: []*computev1.WorkloadMember{member1, member2},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, workloadmember.FieldResourceID),
			},
			valid: true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, workloadmember.FieldResourceID, member1.ResourceId),
			},
			resources: []*computev1.WorkloadMember{member1},
			valid:     true,
		},
		"FilterInstance": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, workloadmember.EdgeInstance,
					instanceresource.FieldResourceID, instance1.GetResourceId()),
			},
			resources: []*computev1.WorkloadMember{member1},
			valid:     true,
		},
		"FilterEmptyInstance": {
			// Instance can never be empty since it's a mandatory field
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, workloadmember.EdgeInstance),
			},
			valid: true,
		},
		"FilterByHasInstance": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{}},
				Filter:   fmt.Sprintf(`has(%s)`, workloadmember.EdgeInstance),
			},
			resources: []*computev1.WorkloadMember{member1, member2},
			valid:     true,
		},
		"FilterByInstanceID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, workloadmember.EdgeInstance, instanceresource.FieldResourceID,
					instance1.GetResourceId()),
			},
			resources: []*computev1.WorkloadMember{member1},
			valid:     true,
		},
		"FilterByWorkloadID": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{}},
				Filter: fmt.Sprintf(`%s.%s = %q`, workloadmember.EdgeWorkload, workloadresource.FieldResourceID,
					workload1.GetResourceId()),
			},
			resources: []*computev1.WorkloadMember{member1},
			valid:     true,
		},
		"FilterLimit": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  2,
			},
			resources: []*computev1.WorkloadMember{member1, member2},
			valid:     true,
		},
		"FilterWithOffsetLimit": {
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{}} // Set the resource kind
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterWorkloadMember() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterWorkloadMember() succeeded but should have failed")
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
						"FilterWorkloadMember() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListWorkloadMember() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListWorkloadMember() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.WorkloadMember, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetWorkloadMember())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListWorkloadMember() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func Test_UpdateWorkloadMember(t *testing.T) {
	os := inv_testing.CreateOs(t)
	workload1 := inv_testing.CreateWorkload(t)
	workload2 := inv_testing.CreateWorkload(t)
	workload3 := inv_testing.CreateWorkload(t)
	instance1 := inv_testing.CreateInstance(t, nil, os)
	instance2 := inv_testing.CreateInstance(t, nil, os)
	instance3 := inv_testing.CreateInstance(t, nil, os)

	member1 := inv_testing.CreateWorkloadMember(t, workload1, instance1)

	putAllFieldsWorkloadMember := computev1.WorkloadMember{
		ResourceId: member1.ResourceId,
		Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
		Workload:   workload2,
		Instance:   instance2,
	}
	testcases := map[string]struct {
		in           *computev1.WorkloadMember
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdatePut": {
			in:           &putAllFieldsWorkloadMember,
			resourceID:   member1.ResourceId,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateMultipleFields": {
			in: &computev1.WorkloadMember{
				Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
				Workload: workload3,
			},
			resourceID: member1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{workloadmember.FieldKind, workloadmember.EdgeWorkload},
			},
			valid: true,
		},
		"UpdateNoFieldmask": {
			in: &computev1.WorkloadMember{
				Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			},
			resourceID:   member1.ResourceId,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateHost": {
			in: &computev1.WorkloadMember{
				Instance: instance3,
			},
			resourceID: member1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{workloadmember.EdgeInstance},
			},
			valid: true,
		},
		"UpdateInvalidResetInstance": {
			in:         &computev1.WorkloadMember{},
			resourceID: member1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{workloadmember.EdgeInstance},
			},
			valid:        false,
			expErrorCode: codes.Internal,
		},
		"UpdateInvalidResetWorkload": {
			in:         &computev1.WorkloadMember{},
			resourceID: member1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{workloadmember.EdgeWorkload},
			},
			valid:        false,
			expErrorCode: codes.Internal,
		},
		"UpdateInvalidFieldMask1": {
			in: &computev1.WorkloadMember{
				Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			},
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateResourceIDNotFound": {
			in: &computev1.WorkloadMember{
				Kind: computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			},
			resourceID: "workloadmember-1234578",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{workloadmember.FieldKind},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(
				ctx, tc.resourceID, tc.fieldMask, updateresreq)

			if !tc.valid {
				require.Errorf(t, err, "UpdateResource() succeeded but should have failed")
				assert.Equal(t, tc.expErrorCode, status.Code(err))
				assert.Nil(t, upRes)
				return
			}
			require.NoErrorf(t, err, "UpdateResource() failed: %s", err)

			// Validate returned resource
			assertSameResource(t, updateresreq, upRes, tc.fieldMask)

			if err != nil {
				if tc.valid {
					t.Errorf("UpdateResource() failed: %s", err)
				}
				// Useless to check if the value was updated if the Update was unsuccessful
				return
			} else if !tc.valid {
				t.Errorf("UpdateResource() succeeded but should have failed")
				return
			}

			// validate update via a get
			getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, tc.resourceID)
			require.NoError(t, err, "GetWorkloadMember() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

func Test_FilterNestedWorkloadMember(t *testing.T) {
	os := inv_testing.CreateOs(t)
	workload1 := inv_testing.CreateWorkload(t)
	host1 := inv_testing.CreateHost(t, nil, nil)
	instance1 := inv_testing.CreateInstance(t, host1, os)
	workload2 := inv_testing.CreateWorkload(t)
	instance2 := inv_testing.CreateInstance(t, nil, os)
	member1 := inv_testing.CreateWorkloadMember(t, workload1, instance1)
	member1.Workload = workload1
	member1.Instance = instance1
	member2 := inv_testing.CreateWorkloadMember(t, workload2, instance2)
	member2.Workload = workload2
	member2.Instance = instance2

	testcases := map[string]struct {
		in                *inv_v1.ResourceFilter
		resources         []*computev1.WorkloadMember
		valid             bool
		expectedCodeError codes.Code
	}{
		"FilterByHostID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, workloadmember.EdgeInstance, instanceresource.EdgeHost,
					hostresource.FieldResourceID, host1.GetResourceId()),
			},
			resources: []*computev1.WorkloadMember{member1},
			valid:     true,
		},
		"FilterByWorkloadExternalID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, workloadmember.EdgeWorkload,
					workloadresource.FieldExternalID, workload2.GetExternalId()),
			},
			resources: []*computev1.WorkloadMember{member2},
			valid:     true,
		},
		"FailTooDeep": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s.%s.%s = %q`, workloadmember.EdgeInstance,
					instanceresource.EdgeHost, hostresource.EdgeSite, siteresource.EdgeRegion,
					regionresource.EdgeParentRegion, regionresource.FieldResourceID, ""),
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{}} // Set the resource kind

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

				resources := make([]*computev1.WorkloadMember, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetWorkloadMember())
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

func Test_EnforceWorkloadMemberInstanceUniqueness(t *testing.T) {
	workload1 := inv_testing.CreateWorkload(t)
	workload2 := inv_testing.CreateWorkload(t)
	os := inv_testing.CreateOs(t)
	instance1 := inv_testing.CreateInstance(t, nil, os)
	instance2 := inv_testing.CreateInstance(t, nil, os)

	t.Run("OnCreate", func(t *testing.T) {
		res1 := &computev1.WorkloadMember{
			Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			Workload: workload1,
			Instance: instance1,
		}
		res2 := &computev1.WorkloadMember{
			Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			Workload: workload2,
			Instance: instance1,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req1 := &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: res1}}
		resp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, req1)
		require.NoError(t, err)
		wmemberResID1 := inv_testing.GetResourceIDOrFail(t, resp)
		res1.ResourceId = wmemberResID1
		defer inv_testing.DeleteResource(t, res1.ResourceId)

		req2 := &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: res2}}
		_, err = inv_testing.TestClients[inv_testing.APIClient].Create(ctx, req2)
		require.Error(t, err)
		s := status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.AlreadyExists, s.Code())
		assert.Contains(t, s.Message(), "instance")
	})

	t.Run("OnUpdate", func(t *testing.T) {
		res1 := &computev1.WorkloadMember{
			Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			Workload: workload1,
			Instance: instance1,
		}
		res2 := &computev1.WorkloadMember{
			Kind:     computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
			Workload: workload2,
			Instance: instance2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req1 := &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: res1}}
		resp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, req1)
		require.NoError(t, err)
		wmemberResID1 := inv_testing.GetResourceIDOrFail(t, resp)
		res1.ResourceId = wmemberResID1
		defer inv_testing.DeleteResource(t, res1.ResourceId)

		req2 := &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: res2}}
		resp, err = inv_testing.TestClients[inv_testing.APIClient].Create(ctx, req2)
		require.NoError(t, err)
		wmemberResID2 := inv_testing.GetResourceIDOrFail(t, resp)
		res2.ResourceId = wmemberResID2
		defer inv_testing.DeleteResource(t, res2.ResourceId)

		// update
		res2.Instance = instance1
		updateReq := &inv_v1.Resource{Resource: &inv_v1.Resource_WorkloadMember{WorkloadMember: res2}}
		_, err = inv_testing.TestClients[inv_testing.APIClient].Update(
			ctx,
			res2.ResourceId,
			&fieldmaskpb.FieldMask{Paths: []string{workloadmember.EdgeInstance}},
			updateReq,
		)
		require.Error(t, err)
		s := status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.AlreadyExists, s.Code())
		assert.Contains(t, s.Message(), "instance")
	})
}

func Test_WorkloadMemberEnumStatusMap(t *testing.T) {
	v, err := store.WorkloadMemberEnumStatusMap("invalid_input",
		int32(computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE))
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestWorkloadMemberMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				host := dao.CreateHost(t, tenantID)
				os := dao.CreateOs(t, tenantID)
				instance := dao.CreateInstance(t, tenantID, host, os)
				workload := dao.CreateWorkload(t, tenantID)
				member := dao.CreateWorkloadMember(t, tenantID, workload, instance)
				res, err := util.WrapResource(member)
				require.NoError(t, err)
				return member.GetResourceId(), res
			},
		},
	})
}

func TestDeleteResources_WorkloadMembers(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				host1 := dao.CreateHost(t, tenantID)
				host2 := dao.CreateHost(t, tenantID)
				os := dao.CreateOs(t, tenantID)
				instance1 := dao.CreateInstance(t, tenantID, host1, os)
				instance2 := dao.CreateInstance(t, tenantID, host2, os)
				workload := dao.CreateWorkload(t, tenantID)
				return tenantID, len([]any{
					dao.CreateWorkloadMemberNoCleanup(t, tenantID, workload, instance1),
					dao.CreateWorkloadMemberNoCleanup(t, tenantID, workload, instance2),
				})
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER,
		},
	})
}
