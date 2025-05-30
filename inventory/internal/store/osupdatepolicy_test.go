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

	oss "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/operatingsystemresource"
	oup "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdatepolicyresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	os_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

//nolint:funlen // test function is long.
func Test_Create_Get_Delete_Update_OSUpdatePolicy(t *testing.T) {
	os := inv_testing.CreateOs(t)
	testcases := map[string]struct {
		in    *computev1.OSUpdatePolicyResource
		valid bool
	}{
		"CreateGoodOsUpdatePolicyTargetMut": {
			in: &computev1.OSUpdatePolicyResource{
				Name:              "Test OS Update Policy",
				Description:       "Test Description",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				UpdateSources:     []string{"test entry1", "test entry2"},
				KernelCommand:     "test command",
				UpdatePolicy:      computev1.UpdatePolicy_UPDATE_POLICY_TARGET,
			},
			valid: true,
		},
		"CreateGoodOsUpdatePolicyTargetImm": {
			in: &computev1.OSUpdatePolicyResource{
				Name:         "Test OS Update Policy",
				Description:  "Test Description",
				TargetOs:     os,
				UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_TARGET,
			},
			valid: true,
		},
		"CreateGoodOsUpdatePolicyLatest": {
			in: &computev1.OSUpdatePolicyResource{
				Name:         "Test OS Update Policy",
				Description:  "Test Description",
				UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_LATEST,
			},
			valid: true,
		},
		"CreateBadOsUpdatePolicyUnspecified": {
			in: &computev1.OSUpdatePolicyResource{
				Name:         "Test OS Update Policy",
				UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_UNSPECIFIED,
			},
			valid: false,
		},
		"CreateBadOsUpdatePolicyTarget1": {
			in: &computev1.OSUpdatePolicyResource{
				Name:              "Test OS Update Policy",
				Description:       "Test Description",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				UpdateSources:     []string{"test entry1", "test entry2"},
				KernelCommand:     "test command",
				TargetOs:          os,
				UpdatePolicy:      computev1.UpdatePolicy_UPDATE_POLICY_TARGET,
			},
			valid: false,
		},
		"CreateBadOsUpdatePolicyTarget2": {
			in: &computev1.OSUpdatePolicyResource{
				Name:         "Test OS Update Policy",
				TargetOs:     &os_v1.OperatingSystemResource{ResourceId: "os-12345678"},
				UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_TARGET,
			},
			valid: false,
		},
		"CreateBadOsUpdatePolicyLatest1": {
			in: &computev1.OSUpdatePolicyResource{
				Name:         "Test OS Update Policy",
				Description:  "Test Description",
				TargetOs:     os,
				UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_LATEST,
			},
			valid: false,
		},
		"CreateBadOsUpdatePolicyLatest2": {
			in: &computev1.OSUpdatePolicyResource{
				Name:              "Test OS Update Policy",
				Description:       "Test Description",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				UpdateSources:     []string{"test entry1", "test entry2"},
				KernelCommand:     "test command",
				UpdatePolicy:      computev1.UpdatePolicy_UPDATE_POLICY_LATEST,
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_OsUpdatePolicy{OsUpdatePolicy: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create
			cupdatesourceResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			oupResID := cupdatesourceResp.GetOsUpdatePolicy().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateOsUpdatePolicy() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = oupResID // Update with created resource ID.
				tc.in.CreatedAt = cupdatesourceResp.GetOsUpdatePolicy().GetCreatedAt()
				tc.in.UpdatedAt = cupdatesourceResp.GetOsUpdatePolicy().GetUpdatedAt()
				assertSameResource(t, createresreq, cupdatesourceResp, nil)
				if !tc.valid {
					t.Errorf("CreateOsUpdatePolicy() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "osupdatepolicy-12345678")
				require.Error(t, err)

				// get
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, oupResID)
				require.NoError(t, err, "GetOSUpdatePolicy() failed")

				// verify data
				tc.in.CreatedAt = getresp.GetResource().GetOsUpdatePolicy().GetCreatedAt()
				tc.in.UpdatedAt = getresp.GetResource().GetOsUpdatePolicy().GetUpdatedAt()
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetOsUpdatePolicy()); !eq {
					t.Errorf("GetOSUpdatePolicy() data not equal: %v", diff)
				}

				// update
				updateresreq := &inv_v1.Resource{
					Resource: &inv_v1.Resource_OsUpdatePolicy{
						OsUpdatePolicy: &computev1.OSUpdatePolicyResource{
							Name:        "Updated Name",
							Description: "Updated Description",
						},
					},
				}
				fieldMask := &fieldmaskpb.FieldMask{
					Paths: []string{oup.FieldName, oup.FieldDescription},
				}
				upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(
					ctx,
					tc.in.ResourceId,
					fieldMask,
					updateresreq,
				)
				if err != nil {
					t.Errorf("UpdateOSUpdatePolicy() failed: %s", err)
				}

				assertSameResource(t, updateresreq, upRes, fieldMask)

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "osupdatepolicy-12345678")
				require.Error(t, err)

				// delete
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(
					ctx,
					oupResID,
				)
				if err != nil {
					t.Errorf("DeleteOsUpdatePolicy() failed %s", err)
				}

				_, err = inv_testing.TestClients[inv_testing.APIClient].Get(ctx, oupResID)
				require.Error(t, err, "Failure - OSUpdatePolicy was not deleted, but should be deleted")
			}
		})
	}
}

//nolint:funlen,cyclop // test function is long and complex.
func Test_FilterOSUpdatePolicies(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	os := dao.CreateOs(t, tenantID)
	osUpPolicy1 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test1"), inv_testing.OsUpdatePolicyDescription("test description"),
		inv_testing.OSUpdatePolicyLatest())
	osUpPolicy2 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test2"),
		inv_testing.OSUpdatePolicyLatest())
	osUpPolicy3 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test3"),
		inv_testing.OSUpdatePolicyTarget(),
		inv_testing.OSUpdatePolicyInstalledPackages("test package"), inv_testing.OSUpdatePolicyKernelCommand("test command"),
		inv_testing.OSUpdatePolicyUpdateSources([]string{"test update source"}))
	osUpPolicy4 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test4"),
		inv_testing.OSUpdatePolicyTarget(),
		inv_testing.OSUpdatePolicyTargetOS(os))

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.OSUpdatePolicyResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy1, osUpPolicy2, osUpPolicy3, osUpPolicy4},
			valid:     true,
		},
		"NoFilterOrderByName": {
			in: &inv_v1.ResourceFilter{
				OrderBy: oup.FieldName,
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy1, osUpPolicy2, osUpPolicy3, osUpPolicy4},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostusb{}},
				Filter:   fmt.Sprintf(`%s = ""`, oup.FieldResourceID),
			},
			resources: []*computev1.OSUpdatePolicyResource{},
			valid:     true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oup.FieldResourceID, osUpPolicy1.GetResourceId()),
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy1},
			valid:     true,
		},
		"FilterUpdateSources": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oup.FieldUpdateSources, osUpPolicy3.GetUpdateSources()[0]),
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy3},
			valid:     true,
		},
		"FilterInstalledPackages": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oup.FieldInstalledPackages, osUpPolicy3.GetInstalledPackages()),
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy3},
			valid:     true,
		},
		"FilterKernelCommand": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oup.FieldKernelCommand, osUpPolicy3.GetKernelCommand()),
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy3},
			valid:     true,
		},
		"FilterTargetOs": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, oup.EdgeTargetOs, oss.FieldResourceID, os.GetResourceId()),
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy4},
			valid:     true,
		},
		"FilterUpdatePolicy": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %s`, oup.FieldUpdatePolicy, computev1.UpdatePolicy_UPDATE_POLICY_LATEST),
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy1, osUpPolicy2},
			valid:     true,
		},
		"FilterLimit": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  5,
			},
			resources: []*computev1.OSUpdatePolicyResource{osUpPolicy1, osUpPolicy2, osUpPolicy3, osUpPolicy4},
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
			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_OsUpdatePolicy{}} // Set the resource kind
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
					t.Errorf("ListOSUpdatePolicies() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListOSUpdatePolicies() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.OSUpdatePolicyResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetOsUpdatePolicy())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListOss() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

//nolint:funlen // test function is long.
func Test_ImmutableFieldsOnUpdateOsUpdatePolicy(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	os := dao.CreateOs(t, tenantID)
	osUpPolicy1 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test1"), inv_testing.OsUpdatePolicyDescription("test description"),
		inv_testing.OSUpdatePolicyLatest())
	osUpPolicy2 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test2"),
		inv_testing.OSUpdatePolicyTarget(), inv_testing.OSUpdatePolicyTargetOS(os))
	osUpPolicy3 := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("test3"),
		inv_testing.OSUpdatePolicyTarget(),
		inv_testing.OSUpdatePolicyInstalledPackages("test package"), inv_testing.OSUpdatePolicyKernelCommand("test command"),
		inv_testing.OSUpdatePolicyUpdateSources([]string{"test update source"}))

	testcases := map[string]struct {
		in           *computev1.OSUpdatePolicyResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdateName": {
			in: &computev1.OSUpdatePolicyResource{
				Name: "New Name",
			},
			resourceID: osUpPolicy1.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{oup.FieldName}},
			valid:      true,
		},
		"UpdateDescription": {
			in: &computev1.OSUpdatePolicyResource{
				Description: "New Description",
			},
			resourceID: osUpPolicy1.GetResourceId(),
			fieldMask:  &fieldmaskpb.FieldMask{Paths: []string{oup.FieldDescription}},
			valid:      true,
		},
		"UpdateImmutableUpdatePolicy": {
			in: &computev1.OSUpdatePolicyResource{
				UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_TARGET,
			},
			resourceID:   osUpPolicy1.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{oup.FieldUpdatePolicy}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableTargetOs": {
			in: &computev1.OSUpdatePolicyResource{
				Name:     "TESTONE",
				TargetOs: dao.CreateOs(t, tenantID),
			},
			resourceID:   osUpPolicy2.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{oup.EdgeTargetOs}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableInstalledPackages": {
			in: &computev1.OSUpdatePolicyResource{
				InstalledPackages: "test_package",
			},
			resourceID:   osUpPolicy3.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{oup.FieldInstalledPackages}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableUpdateSources": {
			in: &computev1.OSUpdatePolicyResource{
				UpdateSources: []string{"source1", "source2"},
			},
			resourceID:   osUpPolicy3.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{oup.FieldUpdateSources}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableKernelCommand": {
			in: &computev1.OSUpdatePolicyResource{
				KernelCommand: "test kernel command",
			},
			resourceID:   osUpPolicy3.GetResourceId(),
			fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{oup.FieldKernelCommand}},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_OsUpdatePolicy{OsUpdatePolicy: tc.in},
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

// TODO: add strong relations tests when adding link to the Instance

func Test_OsUpdatePolicyEnumStateMap(t *testing.T) {
	v, err := store.OSUpdatePolicyEnumStateMap("invalid_input", int32(computev1.UpdatePolicy_UPDATE_POLICY_LATEST))
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestOSUpdatePolocyMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				oup := dao.CreateOSUpdatePolicy(
					t, tenantID, inv_testing.OsUpdatePolicyName("OsPolicy1"), inv_testing.OSUpdatePolicyLatest())
				res, err := util.WrapResource(oup)
				require.NoError(t, err)
				return oup.GetResourceId(), res
			},
		},
	})
}

func TestDeleteResources_OSUpdatePolicies(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				return tenantID, len(
					[]any{
						dao.CreateOSUpdatePolicyNoCleanup(
							t, tenantID, inv_testing.OsUpdatePolicyName("OsPolicy1"), inv_testing.OSUpdatePolicyLatest()),
						dao.CreateOSUpdatePolicyNoCleanup(
							t, tenantID, inv_testing.OsUpdatePolicyName("OsPolicy2"), inv_testing.OSUpdatePolicyLatest()),
						dao.CreateOSUpdatePolicyNoCleanup(
							t, tenantID, inv_testing.OsUpdatePolicyName("OsPolicy3"), inv_testing.OSUpdatePolicyLatest()),
					},
				)
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY,
		},
	})
}
