// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	os_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

func Test_Create_Get_Delete_Update_Os(t *testing.T) {
	testcases := map[string]struct {
		in    *os_v1.OperatingSystemResource
		valid bool
	}{
		"CreateGoodOs": {
			in: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entry1", "test entry2"},
				ImageUrl:          "Repo test entry",
				ImageId:           "some ID",
				Sha256:            inv_testing.RandomSha256v1,
				ProfileName:       "Test OS profile name",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				SecurityFeature:   os_v1.SecurityFeature_SECURITY_FEATURE_NONE,
				ProfileVersion:    "1.0.0",
				OsType:            os_v1.OsType_OS_TYPE_IMMUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
				PlatformBundle:    "test platform bundle",
			},
			valid: true,
		},
		"CreateGoodOsLenovoProvider": {
			in: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entry1", "test entry2"},
				ImageUrl:          "Repo test entry",
				ImageId:           "some ID",
				Sha256:            inv_testing.RandomSha256v1,
				ProfileName:       "Test OS profile name",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				SecurityFeature:   os_v1.SecurityFeature_SECURITY_FEATURE_NONE,
				ProfileVersion:    "1.0.0",
				OsType:            os_v1.OsType_OS_TYPE_IMMUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_LENOVO,
			},
			valid: true,
		},
		"CreateBadOsWrongSha": {
			in: &os_v1.OperatingSystemResource{
				Name:          "Test Os 1",
				UpdateSources: []string{"test entry1", "test entry2"},
				ImageUrl:      "Repo test entry",
				Sha256:        "________________________________________________________________",
			},
			valid: false,
		},
		"CreateBadWithTooLongUpdateSource": {
			in: &os_v1.OperatingSystemResource{
				Name:          "Test Os 1",
				UpdateSources: []string{"test entry1", inv_testing.RandomString(4001)},
				ImageUrl:      "Repo test entry",
				Sha256:        inv_testing.RandomSha256v1,
			},
			valid: false,
		},
		"CreateGoodOsMissingSha": {
			in: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entry1", "test entry2"},
				ImageUrl:          "Repo test entry",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				OsType:            os_v1.OsType_OS_TYPE_MUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
			valid: true,
		},
		"CreateBadOsWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &os_v1.OperatingSystemResource{
				ResourceId:    "os-12345678",
				Name:          "Test Os 2",
				UpdateSources: []string{"test entries"},
				ImageUrl:      "Repo test entry",
				Sha256:        inv_testing.RandomSha256v1,
				ProfileName:   "Test OS profile name",
			},
			valid: false,
		},
		"CreateBadResourceId": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &os_v1.OperatingSystemResource{
				ResourceId:    "os-1234678",
				Name:          "Test Os 2",
				UpdateSources: []string{"test entries"},
				ImageUrl:      "Repo test entry",
				Sha256:        inv_testing.RandomSha256v1,
				ProfileName:   "Test OS profile name",
			},
			valid: false,
		},
		"CreateGoodOsBadSHA256": {
			in: &os_v1.OperatingSystemResource{
				Name:          "Test Os 1",
				UpdateSources: []string{"test entry1", "test entry2"},
				ImageUrl:      "Repo test entry",
				Sha256:        strings.ToUpper(inv_testing.RandomSha256v1),
				ProfileName:   "Test OS profile name",
			},
			valid: false,
		},
		"CreateGoodOsNoRepoURL": {
			in: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entry1", "test entry2"},
				Sha256:            inv_testing.RandomSha256v1,
				ProfileName:       "Test OS profile name",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				SecurityFeature:   os_v1.SecurityFeature_SECURITY_FEATURE_NONE,
				OsType:            os_v1.OsType_OS_TYPE_MUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
			valid: true,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Os{Os: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create
			cupdatesourceResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			osResID := cupdatesourceResp.GetOs().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateOs() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = osResID // Update with created resource ID.
				if !tc.valid {
					t.Errorf("CreateOs() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "os-12345678")
				require.Error(t, err)

				// get
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, osResID)
				require.NoError(t, err, "GetOs() failed")

				// verify data
				tc.in.CreatedAt = getresp.GetResource().GetOs().GetCreatedAt()
				tc.in.UpdatedAt = getresp.GetResource().GetOs().GetUpdatedAt()
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetOs()); !eq {
					t.Errorf("GetOs() data not equal: %v", diff)
				}

				// update
				updateresreq := &inv_v1.Resource{
					Resource: &inv_v1.Resource_Os{
						Os: &os_v1.OperatingSystemResource{
							Name:              "Updated Name",
							InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
						},
					},
				}
				fieldMask := &fieldmaskpb.FieldMask{
					Paths: []string{oss.FieldName, oss.FieldInstalledPackages},
				}
				upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(
					ctx,
					tc.in.ResourceId,
					fieldMask,
					updateresreq,
				)
				if err != nil {
					t.Errorf("UpdateOs() failed: %s", err)
				}

				assertSameResource(t, updateresreq, upRes, fieldMask)

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "os-12345678")
				require.Error(t, err)

				// delete
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(
					ctx,
					osResID,
				)
				if err != nil {
					t.Errorf("DeleteOs() failed %s", err)
				}

				_, err = inv_testing.TestClients[inv_testing.APIClient].Get(ctx, osResID)
				require.Error(t, err, "Failure - OS was not deleted, but should be deleted")
			}
		})
	}
}

func Test_FilterOss(t *testing.T) {
	cupdatesourceResp1 := inv_testing.CreateOsWithArgs(t, inv_testing.RandomSha256v1, "Test OS 1", "Test OS profile name 1",
		os_v1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION, os_v1.OsType_OS_TYPE_MUTABLE)
	cupdatesourceResp2 := inv_testing.CreateOsWithArgs(t, inv_testing.RandomSha256v2, "Test OS 2", "Test OS profile name 2",
		os_v1.SecurityFeature_SECURITY_FEATURE_NONE, os_v1.OsType_OS_TYPE_MUTABLE)

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*os_v1.OperatingSystemResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1, cupdatesourceResp2},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: oss.FieldResourceID,
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1, cupdatesourceResp2},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostusb{}},
				Filter:   fmt.Sprintf(`%s = ""`, oss.FieldResourceID),
			},
			resources: []*os_v1.OperatingSystemResource{},
			valid:     true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oss.FieldResourceID, cupdatesourceResp1.ResourceId),
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1},
			valid:     true,
		},
		"FilterUpdateSources": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oss.FieldUpdateSources, cupdatesourceResp1.UpdateSources[0]),
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1, cupdatesourceResp2},
			valid:     true,
		},
		"FilterBySecurityFeatures": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %v`, oss.FieldSecurityFeature, cupdatesourceResp1.SecurityFeature),
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1},
			valid:     true,
		},
		"FilterBySHA256": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, oss.FieldSha256, cupdatesourceResp1.Sha256),
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1},
			valid:     true,
		},
		"NoFilterOrderBySHA256": {
			in: &inv_v1.ResourceFilter{
				OrderBy: oss.FieldSha256,
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1, cupdatesourceResp2},
			valid:     true,
		},
		"FilterBySHA256Invalid": {
			// We have special internal handling for field names comprised of letters and numbers. This test makes sure
			// these workarounds are NOT exposed outside.
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, "sha_256", cupdatesourceResp1.Sha256),
			},
			resources: []*os_v1.OperatingSystemResource{},
			valid:     false,
		},
		"FilterLimit": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  2,
			},
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1, cupdatesourceResp2},
			valid:     true,
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
			resources: []*os_v1.OperatingSystemResource{cupdatesourceResp1, cupdatesourceResp2},
			valid:     true,
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_Os{}} // Set the resource kind
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

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

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListOss() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListOss() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*os_v1.OperatingSystemResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetOs())
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

// This test does NOT cover SHA256 and Profile Name (immutable fields) invalid cases.
// They are covered in the Test_ImmutableFieldsOnUpdate test.
func Test_UpdateOs(t *testing.T) {
	// create Os to update
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Os{
			Os: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entries"},
				ImageUrl:          "Repo test entry",
				Sha256:            inv_testing.RandomSha256v1,
				ProfileName:       "Test OS profile name 1",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				SecurityFeature:   os_v1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
				OsType:            os_v1.OsType_OS_TYPE_MUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cosResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
	require.NoError(t, err)
	osResID := inv_testing.GetResourceIDOrFail(t, cosResp)
	t.Cleanup(func() { inv_testing.DeleteResource(t, osResID) })

	testcases := map[string]struct {
		in           *os_v1.OperatingSystemResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdateName": {
			in: &os_v1.OperatingSystemResource{
				Name:        "Updated Name",
				Sha256:      inv_testing.RandomSha256v3,
				ProfileName: "Test OS profile name 3",
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{oss.FieldName},
			},
			valid: true,
		},
		"UpdateMultipleFields": {
			in: &os_v1.OperatingSystemResource{
				Name:          "Updated Name 2",
				KernelCommand: "linux",
				UpdateSources: []string{"update 2"},
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldKernelCommand, oss.FieldName, oss.FieldUpdateSources,
				},
			},
			valid: true,
		},
		"UpdateImmutableSecurityFeatureFail": {
			in: &os_v1.OperatingSystemResource{
				SecurityFeature: os_v1.SecurityFeature_SECURITY_FEATURE_NONE,
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldSecurityFeature,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutablePlatformBundleFail": {
			in: &os_v1.OperatingSystemResource{
				PlatformBundle: "some platform bundle",
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldPlatformBundle,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableProfileNameWithSameValue": {
			in: &os_v1.OperatingSystemResource{
				ProfileName: "Test OS profile name 1",
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldProfileName,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableSHA256WithSameValue": {
			in: &os_v1.OperatingSystemResource{
				Sha256: inv_testing.RandomSha256v1,
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldSha256,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateInvalidFieldMask1": {
			in: &os_v1.OperatingSystemResource{
				Name: "Updated Name 5",
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateNoFieldMask": {
			in: &os_v1.OperatingSystemResource{
				Name: "Updated Name 5",
			},
			resourceID:   osResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateResourceIDNotFound": {
			in: &os_v1.OperatingSystemResource{
				Name:        "Updated Name",
				ImageUrl:    "Repo test entry update",
				Sha256:      inv_testing.RandomSha256v3,
				ProfileName: "Test OS profile name 3",
			},
			resourceID: "os-12345678",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{oss.FieldName, oss.FieldImageURL},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Os{Os: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, tc.resourceID,
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
			getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, tc.resourceID)
			require.NoError(t, err, "GetResource() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

func Test_ImmutableFieldsOnUpdate(t *testing.T) {
	// create Os to update
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Os{
			Os: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entries"},
				ImageUrl:          "Repo test entry",
				Sha256:            inv_testing.RandomSha256v1,
				ProfileName:       "Test OS profile name 1",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				OsType:            os_v1.OsType_OS_TYPE_MUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cosResp, _ := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
	osResID := inv_testing.GetResourceIDOrFail(t, cosResp)
	t.Cleanup(func() { inv_testing.DeleteResource(t, osResID) })

	os1 := inv_testing.CreateOsWithArgs(t, inv_testing.RandomSha256v2, "Test OS 2", "Test OS profile name 2",
		os_v1.SecurityFeature_SECURITY_FEATURE_UNSPECIFIED, os_v1.OsType_OS_TYPE_MUTABLE)

	getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, os1.ResourceId)
	require.NoError(t, err, "GetResource() failed")
	t.Logf("SHA256 in OS resource is %v", getresp.GetResource().GetOs().GetSha256())

	allFields := os_v1.OperatingSystemResource{
		ResourceId:        os1.ResourceId,
		Name:              "TEST",
		Architecture:      "TEST",
		KernelCommand:     "TEST",
		UpdateSources:     []string{"TEST"},
		ImageUrl:          "TEST",
		Sha256:            inv_testing.RandomSha256v2,
		ProfileName:       "Test OS profile name 2",
		InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
	}
	fmAllFields, err := util.BuildFieldMaskFromMessage(&allFields)
	require.NoError(t, err, "Failed to create fieldmask for all Fields")

	testcases := map[string]struct {
		in           *os_v1.OperatingSystemResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdatePatchAllFields": {
			in:           &allFields,
			resourceID:   os1.ResourceId,
			fieldMask:    fmAllFields,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableProfileName": {
			in: &os_v1.OperatingSystemResource{
				ProfileName: "Another test OS profile name 1",
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldProfileName,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableProfileName2": {
			in: &os_v1.OperatingSystemResource{
				ProfileName: "Test OS profile name 2",
			},
			resourceID: os1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldProfileName,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableSHA256": {
			in: &os_v1.OperatingSystemResource{
				Sha256: inv_testing.GenerateRandomSha256(),
			},
			resourceID: osResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldSha256,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableSHA2562": {
			in: &os_v1.OperatingSystemResource{
				Sha256: inv_testing.RandomSha256v2,
			},
			resourceID: os1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldSha256,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableOSType": {
			in: &os_v1.OperatingSystemResource{
				OsType: os_v1.OsType_OS_TYPE_IMMUTABLE,
			},
			resourceID: os1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldOsType,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableImageUrl": {
			in: &os_v1.OperatingSystemResource{
				ImageUrl: "some new URL",
			},
			resourceID: os1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldImageURL,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateImmutableProfileVersion": {
			in: &os_v1.OperatingSystemResource{
				ProfileVersion: "2.0.0",
			},
			resourceID: os1.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					oss.FieldProfileVersion,
				},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Os{Os: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, tc.resourceID,
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
			getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, tc.resourceID)
			require.NoError(t, err, "GetResource() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

// Test_StrongRelations_On_Delete_Ou_Os validates if an OS cannot be deleted as
// long as an instance has an edge relationship with it.
func Test_StrongRelations_On_Delete_Ou_Os(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create the Desired OS to test on.
	os := &os_v1.OperatingSystemResource{
		Name:          "Test Os 1",
		UpdateSources: []string{"source 1"},
		ImageUrl:      "test repo url",
		Sha256:        inv_testing.RandomSha256v1,
		ProfileName:   "Test OS profile name",
		OsType:        os_v1.OsType_OS_TYPE_MUTABLE,
		OsProvider:    os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
	}
	resp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx,
		&inv_v1.Resource{Resource: &inv_v1.Resource_Os{Os: os}})
	require.NoError(t, err)
	os1ResID := inv_testing.GetResourceIDOrFail(t, resp)
	os.ResourceId = os1ResID

	// Create an instance pointing to that OS. Not using the helper functions,
	// as we don't want the auto-cleanup.
	ins := &computev1.InstanceResource{
		Kind:         computev1.InstanceKind_INSTANCE_KIND_VM,
		Name:         "test instance",
		DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
		Host:         nil,
		DesiredOs:    os,
	}
	resp, err = inv_testing.TestClients[inv_testing.APIClient].Create(ctx,
		&inv_v1.Resource{Resource: &inv_v1.Resource_Instance{Instance: ins}})
	require.NoError(t, err)
	os2ResID := inv_testing.GetResourceIDOrFail(t, resp)
	ins.ResourceId = os2ResID

	// Try to delete the OS, this should fail because the instance points to it.
	_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, os.ResourceId)
	require.Error(t, err, "DeleteOs() should fail")
	assertStrongRelationError(t, err, "violates foreign key constraint")
	// Delete the instance and try to delete the OS again. This should work.
	inv_testing.HardDeleteInstance(t, ins.ResourceId)
	_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, os.ResourceId)
	require.NoError(t, err)
}

func Test_Create_Get_Delete_Update_Os_Install_Packages(t *testing.T) {
	testcases := map[string]struct {
		in    *os_v1.OperatingSystemResource
		valid bool
	}{
		"CreateOswithInstallPackages": {
			in: &os_v1.OperatingSystemResource{
				Name:              "Test Os 1",
				UpdateSources:     []string{"test entry1", "test entry2"},
				ImageUrl:          "Repo test entry",
				Sha256:            inv_testing.RandomSha256v1,
				ProfileName:       "Test OS profile name",
				InstalledPackages: "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero",
				OsType:            os_v1.OsType_OS_TYPE_MUTABLE,
				OsProvider:        os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
			valid: true,
		},
		"CreateOswithoutInstallPackages": {
			in: &os_v1.OperatingSystemResource{
				Name:          "Test Os 1",
				UpdateSources: []string{"test entry1", "test entry2"},
				ImageUrl:      "Repo test entry",
				Sha256:        inv_testing.RandomSha256v1,
				ProfileName:   "Test OS profile name",
				OsType:        os_v1.OsType_OS_TYPE_MUTABLE,
				OsProvider:    os_v1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
			},
			valid: true,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Os{Os: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create
			cupdatesourceResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			osResID := cupdatesourceResp.GetOs().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateOs() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = osResID // Update with created resource ID.
				tc.in.CreatedAt = cupdatesourceResp.GetOs().GetCreatedAt()
				tc.in.UpdatedAt = cupdatesourceResp.GetOs().GetUpdatedAt()
				assertSameResource(t, createresreq, cupdatesourceResp, nil)
				if !tc.valid {
					t.Errorf("CreateOs() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, osResID)
				require.NoError(t, err, "GetOs() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetOs()); !eq {
					t.Errorf("GetOs() data not equal: %v", diff)
				}

				// update
				updateresreq := &inv_v1.Resource{
					Resource: &inv_v1.Resource_Os{
						Os: &os_v1.OperatingSystemResource{
							Name:              "Updated Name",
							InstalledPackages: "intel-opencl-icd-updated\nintel-level-zero-gpu-updated\nlevel-zero-updated",
						},
					},
				}

				fieldMask := &fieldmaskpb.FieldMask{
					Paths: []string{oss.FieldName, oss.FieldInstalledPackages},
				}

				upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(
					ctx,
					tc.in.ResourceId,
					fieldMask,
					updateresreq,
				)
				require.NoError(t, err)

				// Validate returned resource
				assertSameResource(t, updateresreq, upRes, fieldMask)

				// delete
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(
					ctx,
					osResID,
				)
				require.NoError(t, err)

				_, err = inv_testing.TestClients[inv_testing.APIClient].Get(ctx, osResID)
				require.Error(t, err, "Failure - OS was not deleted, but should be deleted")
			}
		})
	}
}

func Test_OsEnumStateMap(t *testing.T) {
	v, err := store.OsEnumStateMap("invalid_input", int32(os_v1.SecurityFeature_SECURITY_FEATURE_NONE))
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestOperatingSystemMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				os := dao.CreateOs(t, tenantID)
				res, err := util.WrapResource(os)
				require.NoError(t, err)
				return os.GetResourceId(), res
			},
		},
	})
}

func TestDeleteResources_OSes(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				return tenantID, len(
					[]any{
						dao.CreateOsNoCleanup(t, tenantID),
						dao.CreateOsNoCleanup(t, tenantID),
						dao.CreateOsNoCleanup(t, tenantID),
					},
				)
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_OS,
		},
	})
}
