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
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
)

func Test_Metadata_Inheritance_Region(t *testing.T) {
	// Create required Regions
	region1 := inv_testing.CreateRegionWithMeta(t, metaR1, nil)
	region2 := inv_testing.CreateRegionWithMeta(t, metaR2, region1)
	region3 := inv_testing.CreateRegionWithMeta(t, metaR3, region2)
	region4 := inv_testing.CreateRegion(t, region3)
	region5 := inv_testing.CreateRegionWithMeta(t, metaR5, region3)

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	testcases := map[string]struct {
		in                string
		expPhyMeta        string
		expStandaloneMeta string
	}{
		"NoMetadataFromParent": {
			in:                region4.ResourceId,
			expPhyMeta:        expPhyMeta1,
			expStandaloneMeta: "",
		},
		"PhyMetadataFromParent": {
			in:                region5.ResourceId,
			expPhyMeta:        expPhyMeta2,
			expStandaloneMeta: `[{"key":"key1-test", "value":"region_key1_lvl4-test"}]`,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			t.Run("Verify with Get", func(t *testing.T) {
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, tc.in)
				require.NoError(t, err, "GetRegion() failed")

				if getresp.RenderedMetadata == nil {
					t.Errorf("Get rendered Metadata failed")
					t.FailNow()
				}
				if !CompareMetadata(t, getresp.RenderedMetadata.PhyMetadata, tc.expPhyMeta) {
					t.Errorf("Physical Metadata data not equal - want: %s, got: %s",
						tc.expPhyMeta, getresp.RenderedMetadata.PhyMetadata,
					)
				}
				if !CompareMetadata(t, getresp.GetResource().GetRegion().Metadata, tc.expStandaloneMeta) {
					t.Errorf("Standalone Metadata data not equal - want: %s, got: %s",
						tc.expStandaloneMeta, getresp.GetResource().GetRegion().Metadata,
					)
				}
				if getresp.RenderedMetadata.LogiMetadata != "" {
					t.Errorf("Regions should not have Physical Metadata")
				}
			})
			t.Run("Verify with LIST", func(t *testing.T) {
				listResp, err := inv_testing.TestClients[inv_testing.APIClient].
					List(ctx, &inv_v1.ResourceFilter{
						Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Region{}},
					})

				require.NoError(t, err, "GetRegion() failed")
				require.NotEmpty(t, listResp.GetResources())
				resource := collections.Filter(listResp.GetResources(), func(v *inv_v1.GetResourceResponse) bool {
					return v.GetResource().GetRegion().GetResourceId() == tc.in
				})[0]

				if resource.RenderedMetadata == nil {
					t.Errorf("Get rendered Metadata failed")
					t.FailNow()
				}
				if !CompareMetadata(t, resource.RenderedMetadata.PhyMetadata, tc.expPhyMeta) {
					t.Errorf("Physical Metadata data not equal - want: %s, got: %s",
						tc.expPhyMeta, resource.RenderedMetadata.PhyMetadata,
					)
				}
				if !CompareMetadata(t, resource.GetResource().GetRegion().Metadata, tc.expStandaloneMeta) {
					t.Errorf("Standalone Metadata data not equal - want: %s, got: %s",
						tc.expStandaloneMeta, resource.GetResource().GetRegion().Metadata,
					)
				}
				if resource.RenderedMetadata.LogiMetadata != "" {
					t.Errorf("Regions should not have Physical Metadata")
				}
			})
		})
	}
}

func Test_StrongRelations_On_Delete_Region(t *testing.T) {
	region1 := inv_testing.CreateRegion(t, nil)
	region2 := inv_testing.CreateRegion(t, region1)
	_ = inv_testing.CreateSite(t, region2, nil)
	// Adding this region to verify delete works fine for region without children.
	// Delete is done on cleanup automatically
	_ = inv_testing.CreateRegion(t, region1)

	err := inv_testing.DeleteResourceAndReturnError(t, region1.ResourceId)
	assertStrongRelationError(t, err, "the region has relations with region and cannot be deleted")

	err = inv_testing.DeleteResourceAndReturnError(t, region2.ResourceId)
	assertStrongRelationError(t, err, "the region has relations with site and cannot be deleted")
}

func Test_Create_Get_Delete_Update_Region(t *testing.T) {
	parentRegion := inv_testing.CreateRegion(t, nil)

	testcases := map[string]struct {
		in    *location_v1.RegionResource
		valid bool
	}{
		"CreateGoodRegion": {
			in: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
			},
			valid: true,
		},
		"CreateRegionWithInvalidMetadataKey": {
			in: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name-1_","value":"test"}]`,
			},
			valid: false,
		},
		"CreateRegionWithInvalidMetadataValue": {
			in: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"Test*"}]`,
			},
			valid: false,
		},
		"CreateRegionWithInvalidMetadataKeyLength": {
			in: &location_v1.RegionResource{
				Name: "Test Region 1",
				Metadata: `[
					{"key":"example.com/invalidkeylengthinvalidkeylengthinvalidkeylengthinvalidkeylength","value":"test"}
				]`,
			},
			valid: false,
		},
		"CreateRegionWithInvalidMetadataValueLength": {
			in: &location_v1.RegionResource{
				Name: "Test Region 1",
				Metadata: `[
					{"key":"cluster-name","value":"invalidvaluelengthinvalidvaluelengthinvalidvaluelengthinvalidval"}
					]`,
			},
			valid: false,
		},
		"CreateBadRegionWithMetadata": {
			in: &location_v1.RegionResource{
				Name:     "Wrong Region",
				Metadata: `{"key":"cluster-name","value":"test"}]`,
			},
			valid: false,
		},
		"CreateBadRegionWithInvalidMetadata": {
			in: &location_v1.RegionResource{
				Name:     "Wrong Region Metadata (duplicate keys)",
				Metadata: metaDuplicatedKeys,
			},
			valid: false,
		},
		"CreateBadRegionWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &location_v1.RegionResource{
				ResourceId: "region-12345678",
				Name:       "Test Region 2",
			},
			valid: false,
		},
		"CreateBadRegionWithInvalidResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &location_v1.RegionResource{
				ResourceId: "region-test-12345678",
				Name:       "Test Region 2",
			},
			valid: false,
		},
		"CreateGoodRegionWithParentRegion": {
			in: &location_v1.RegionResource{
				Name:         "Test Region 1",
				ParentRegion: parentRegion,
			},
			valid: true,
		},
		"CreateBadRegionWithNonExistingParentRegion": {
			in: &location_v1.RegionResource{
				Name: "Test Region 1",
				ParentRegion: &location_v1.RegionResource{
					ResourceId: "region-12345678",
				},
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{Region: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create
			cregResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			regionResID := cregResp.GetRegion().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateRegion() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = regionResID // Update with created resource ID.
				tc.in.CreatedAt = cregResp.GetRegion().CreatedAt
				tc.in.UpdatedAt = cregResp.GetRegion().UpdatedAt
				assertSameResource(t, createresreq, cregResp, nil)
				if !tc.valid {
					t.Errorf("CreateRegion() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "region-12345678")
				require.Error(t, err)

				// get
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, regionResID)
				require.NoError(t, err, "GetRegion() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetRegion()); !eq {
					t.Errorf("GetRegion() data not equal: %v", diff)
				}

				// update
				updateresreq := &inv_v1.Resource{
					Resource: &inv_v1.Resource_Region{
						Region: &location_v1.RegionResource{
							Name: "Updated Name",
						},
					},
				}

				fieldMask := &fieldmaskpb.FieldMask{Paths: []string{regionresource.FieldName}}
				upRes, err := inv_testing.TestClients[inv_testing.RMClient].Update(
					ctx,
					regionResID,
					fieldMask,
					updateresreq,
				)
				if err != nil {
					t.Errorf("UpdateRegion() failed: %s", err)
				}

				// Validate returned resource
				assertSameResource(t, updateresreq, upRes, fieldMask)

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "region-12345678")
				require.Error(t, err)

				// delete
				_, err = inv_testing.TestClients[inv_testing.RMClient].Delete(
					ctx,
					regionResID,
				)
				if err != nil {
					t.Errorf("DeleteRegion() failed %s", err)
				}
			}
		})
	}
}

func Test_RegionDepthDetection(t *testing.T) {
	fmUpdateParentRegion := &fieldmaskpb.FieldMask{Paths: []string{regionresource.EdgeParentRegion}}

	t.Run("DepthDetectionOnCreate", func(t *testing.T) {
		// Create a linear graph of 5 regions.
		region1 := inv_testing.CreateRegion(t, nil)
		region2 := inv_testing.CreateRegion(t, region1)
		region3 := inv_testing.CreateRegion(t, region2)
		region4 := inv_testing.CreateRegion(t, region3)
		region5 := inv_testing.CreateRegion(t, region4)

		// Link a 6th region to the graph. This should fail for exceeding the depth.
		createresreq := &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{
				Region: &location_v1.RegionResource{
					Name:         "Test Region 6",
					ParentRegion: region5,
				},
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
		assert.Error(t, err)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, errors.ErrorToStringWithDetails(err), "depth")
	})

	t.Run("CycleDetectionOnUpdate", func(t *testing.T) {
		region1 := inv_testing.CreateRegion(t, nil)
		region2 := inv_testing.CreateRegion(t, region1)

		// Try to update region1 to point to region2.
		region2.ParentRegion = nil // Delete ref to prevent proto cycle. Not committed to the DB.
		region1.ParentRegion = region2
		updateresreq := &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{
				Region: region1,
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := inv_testing.TestClients[inv_testing.APIClient].
			Update(ctx, region1.ResourceId, fmUpdateParentRegion, updateresreq)
		assert.Error(t, err)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, errors.ErrorToStringWithDetails(err), "depth")
	})

	t.Run("CycleDetectionSelfOnUpdate", func(t *testing.T) {
		region1 := inv_testing.CreateRegion(t, nil)
		region1.ParentRegion = &location_v1.RegionResource{ResourceId: region1.ResourceId}
		updateresreq := &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{
				Region: region1,
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := inv_testing.TestClients[inv_testing.APIClient].
			Update(ctx, region1.ResourceId, fmUpdateParentRegion, updateresreq)
		assert.Error(t, err)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, errors.ErrorToStringWithDetails(err), "depth")
	})

	t.Run("DepthDetectionOnUpdateHead", func(t *testing.T) {
		// Create a head region, not connected to others yet.
		region0 := inv_testing.CreateRegion(t, nil)
		// Create a linear graph of 5 regions.
		region1 := inv_testing.CreateRegion(t, nil)
		region2 := inv_testing.CreateRegion(t, region1)
		region3 := inv_testing.CreateRegion(t, region2)
		region4 := inv_testing.CreateRegion(t, region3)
		inv_testing.CreateRegion(t, region4)

		// Link the regions together by updating the current head. This should fail for exceeding the depth.
		region1.ParentRegion = region0
		updateresreq := &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{
				Region: region1,
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := inv_testing.TestClients[inv_testing.APIClient].
			Update(ctx, region1.ResourceId, fmUpdateParentRegion, updateresreq)
		assert.Error(t, err)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, errors.ErrorToStringWithDetails(err), "depth")
	})

	t.Run("DepthDetectionOnMergeLinearTrees", func(t *testing.T) {
		// Create two linear region trees, both with 3 members each.
		region1a := inv_testing.CreateRegion(t, nil)
		region2a := inv_testing.CreateRegion(t, region1a)
		region3a := inv_testing.CreateRegion(t, region2a)

		region1b := inv_testing.CreateRegion(t, nil)
		region2b := inv_testing.CreateRegion(t, region1b)
		inv_testing.CreateRegion(t, region2b)

		// Link the regions together by pointing one head to the tail of the other.
		// This should fail for exceeding the depth.
		region1b.ParentRegion = region3a
		updateresreq := &inv_v1.Resource{
			Resource: &inv_v1.Resource_Region{
				Region: region1b,
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := inv_testing.TestClients[inv_testing.APIClient].
			Update(ctx, region1b.ResourceId, fmUpdateParentRegion, updateresreq)
		assert.Error(t, err)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, errors.ErrorToStringWithDetails(err), "depth")
	})
}

func benchmarkInvStoreUpdateRegionManyChildren(b *testing.B, numChildren int) {
	b.Helper()

	// Create a head region, not connected to others yet.
	region0 := inv_testing.CreateRegion(b, nil)
	// Create a linear graph of 4 regions.
	region1 := inv_testing.CreateRegion(b, nil)
	region2 := inv_testing.CreateRegion(b, region1)
	region3 := inv_testing.CreateRegion(b, region2)
	region4 := inv_testing.CreateRegion(b, region3)
	// Add lots of child regions to the tail.
	for i := 0; i < numChildren; i++ {
		inv_testing.CreateRegion(b, region4)
	}

	// Link the regions together by updating the current head. This should fail for exceeding the depth.
	region1.ParentRegion = region0
	updateresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{Region: region1},
	}
	fmUpdate := &fieldmaskpb.FieldMask{Paths: []string{regionresource.EdgeParentRegion}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, region1.ResourceId, fmUpdate, updateresreq)
		assert.Error(b, err)
		s, ok := status.FromError(err)
		assert.True(b, ok)
		assert.Equal(b, codes.InvalidArgument, s.Code())
		assert.Contains(b, errors.ErrorToStringWithDetails(err), "depth")
	}
	b.StopTimer() // Stop the timer so deferred cleanup is not measured.
}

func BenchmarkInvStore_UpdateRegionManyChildren(b *testing.B) {
	l := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(l)
	b.Run("1", func(b *testing.B) {
		benchmarkInvStoreUpdateRegionManyChildren(b, 1)
	})
	b.Run("10", func(b *testing.B) {
		benchmarkInvStoreUpdateRegionManyChildren(b, 10)
	})
	b.Run("100", func(b *testing.B) {
		benchmarkInvStoreUpdateRegionManyChildren(b, 100)
	})
	// Higher numbers take very long to complete.
	b.Run("200", func(b *testing.B) {
		benchmarkInvStoreUpdateRegionManyChildren(b, 200)
	})
}

func Test_FilterRegions(t *testing.T) {
	parentRegion := inv_testing.CreateRegion(t, nil)
	cregResp1 := inv_testing.CreateRegion(t, nil)
	cregResp2 := inv_testing.CreateRegion(t, parentRegion)
	// Setting again the edge nilled by the helper
	cregResp2.ParentRegion = parentRegion

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*location_v1.RegionResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*location_v1.RegionResource{cregResp1, cregResp2, parentRegion},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: regionresource.FieldResourceID,
			},
			resources: []*location_v1.RegionResource{cregResp1, cregResp2, parentRegion},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, regionresource.FieldResourceID),
			},
			valid: true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, regionresource.FieldResourceID, cregResp2.ResourceId),
			},
			resources: []*location_v1.RegionResource{cregResp2},
			valid:     true,
		},
		"FilterParentRegion": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, regionresource.EdgeParentRegion,
					regionresource.FieldResourceID, parentRegion.GetResourceId()),
			},
			resources: []*location_v1.RegionResource{cregResp2},
			valid:     true,
		},
		"FilterByHasParentRegion": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, regionresource.EdgeParentRegion),
			},
			resources: []*location_v1.RegionResource{cregResp2},
			valid:     true,
		},
		"FilterByHasParentRegionHasParentRegion": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s.%s)`, regionresource.EdgeParentRegion, regionresource.EdgeParentRegion),
			},
			resources: []*location_v1.RegionResource{},
			valid:     true,
		},
		"FilterByParentRegionJSON": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`parentRegion.resourceId = %q`,
					cregResp2.ParentRegion.ResourceId),
			},
			resources: []*location_v1.RegionResource{cregResp2},
			valid:     true,
		},
		"FilterByChildrenID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, regionresource.EdgeChildren, regionresource.FieldResourceID,
					cregResp2.GetResourceId()),
			},
			resources: []*location_v1.RegionResource{parentRegion},
			valid:     true,
		},
		"FilterByParentRegionByParentRegionJSON": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`parentRegion.parentRegion.resourceId = %q`,
					cregResp2.ParentRegion.ResourceId),
			},
			resources: []*location_v1.RegionResource{},
			valid:     true,
		},
		"FilterByNotHasParentRegion": {
			in: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: &inv_v1.Resource_Region{}},
				Filter:   fmt.Sprintf(`NOT has(%s)`, regionresource.EdgeParentRegion),
			},
			resources: []*location_v1.RegionResource{cregResp1, parentRegion},
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
			resources: []*location_v1.RegionResource{cregResp1, cregResp2, parentRegion},
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_Region{}} // Set the resource kind
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterRegions() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterRegions() succeeded but should have failed")
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
						"FilterRegions() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListRegions() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListRegions() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*location_v1.RegionResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetRegion())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					expCopy := *expected //nolint:govet // it's ok to copy the mutex
					regionEdgesOnlyResourceID(&expCopy)
					regionEdgesOnlyResourceID(resources[i])

					// Compare metadata separately
					assert.True(t, CompareMetadata(t, resources[i].Metadata, expCopy.Metadata))
					expCopy.Metadata = ""
					resources[i].Metadata = ""
					if eq, diff := inv_testing.ProtoEqualOrDiff(&expCopy, resources[i]); !eq {
						t.Errorf("ListRegions() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func Test_UpdateRegion(t *testing.T) {
	region1 := inv_testing.CreateRegion(t, nil)
	region2 := inv_testing.CreateRegion(t, nil)
	region3 := inv_testing.CreateRegion(t, nil)
	region4 := inv_testing.CreateRegion(t, region1)
	region5 := inv_testing.CreateRegion(t, region1)

	putRegion := location_v1.RegionResource{
		ResourceId:   region5.ResourceId,
		Name:         "TEST",
		RegionKind:   "TEST",
		ParentRegion: region3,
		Metadata:     metaR1,
	}

	testcases := map[string]struct {
		in           *location_v1.RegionResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdatePut": {
			in:           &putRegion,
			resourceID:   region5.ResourceId,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateName": {
			in: &location_v1.RegionResource{
				Name: "Updated Name",
			},
			resourceID: region2.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{regionresource.FieldName},
			},
			valid: true,
		},
		"UpdateResetName": {
			in:         &location_v1.RegionResource{},
			resourceID: region2.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{regionresource.FieldName},
			},
			valid: true,
		},
		"UpdateParentRegion": {
			in: &location_v1.RegionResource{
				ParentRegion: region1,
			},
			resourceID: region2.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{regionresource.EdgeParentRegion},
			},
			valid: true,
		},
		"UpdateResetParentRegion": {
			in:         &location_v1.RegionResource{},
			resourceID: region4.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{regionresource.EdgeParentRegion},
			},
			valid: true,
		},
		"UpdateNoFieldMask": {
			in: &location_v1.RegionResource{
				Name: "Updated Name 2",
			},
			resourceID:   region2.ResourceId,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateInvalidFieldMask1": {
			in: &location_v1.RegionResource{
				Name: "Updated Name 3",
			},
			resourceID: region2.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateInvalidMetadata": {
			in: &location_v1.RegionResource{
				Metadata: metaDuplicatedKeys,
			},
			resourceID: region2.ResourceId,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{regionresource.FieldMetadata},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateResourceIDNotFound": {
			in: &location_v1.RegionResource{
				Name: "Updated Name",
			},
			resourceID: "region-12345678",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{regionresource.FieldName},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{Region: tc.in},
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

func regionEdgesOnlyResourceID(expected *location_v1.RegionResource) {
	if expected.ParentRegion != nil {
		expected.ParentRegion = &location_v1.RegionResource{ResourceId: expected.ParentRegion.ResourceId}
	}
}

func Test_NestedFilterRegion(t *testing.T) {
	parentRegion := inv_testing.CreateRegion(t, nil)
	region1 := inv_testing.CreateRegion(t, parentRegion)
	region2 := inv_testing.CreateRegion(t, region1)
	cRegion1 := *region1 //nolint:govet // copying locks in test
	region1.ParentRegion = parentRegion
	region2.ParentRegion = &cRegion1

	testcases := map[string]struct {
		in                *inv_v1.ResourceFilter
		resources         []*location_v1.RegionResource
		valid             bool
		expectedCodeError codes.Code
	}{
		"FilterByParentRegionID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, regionresource.EdgeParentRegion,
					regionresource.EdgeParentRegion, regionresource.FieldResourceID, parentRegion.GetResourceId()),
			},
			resources: []*location_v1.RegionResource{region2},
			valid:     true,
		},
		"FilterByEmptyRegionID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s.%s)`, regionresource.EdgeParentRegion, regionresource.EdgeParentRegion),
			},
			resources: []*location_v1.RegionResource{parentRegion, region1},
			valid:     true,
		},
		"FilterByChildRegionID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, regionresource.EdgeChildren,
					regionresource.FieldResourceID, region1.GetResourceId()),
			},
			resources: []*location_v1.RegionResource{parentRegion},
			valid:     true,
		},
		"FilterByHasChildren": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, regionresource.EdgeChildren),
			},
			resources: []*location_v1.RegionResource{parentRegion, region1},
			valid:     true,
		},
		"FilterByNotHasChildren": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, regionresource.EdgeChildren),
			},
			resources: []*location_v1.RegionResource{region2},
			valid:     true,
		},
		"FailTooDeep": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s.%s.%s = %q`, regionresource.EdgeParentRegion,
					regionresource.EdgeParentRegion, regionresource.EdgeParentRegion, regionresource.EdgeParentRegion,
					regionresource.EdgeParentRegion, regionresource.FieldResourceID, parentRegion.GetResourceId()),
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

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_Region{}} // Set the resource kind

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
						"FilterRegion() failed - want: %s, got: %s",
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

				resources := make([]*location_v1.RegionResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetRegion())
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

func TestMultitenancySanity_GetTreeHierarchy(t *testing.T) {
	T1 := uuid.NewString()
	T2 := uuid.NewString()
	res1Tenant1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
				TenantId: T1,
			},
		},
	}
	res2Tenant1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
				TenantId: T1,
			},
		},
	}

	res1Tenant2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
				TenantId: T2,
			},
		},
	}
	res2Tenant2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Region{
			Region: &location_v1.RegionResource{
				Name:     "Test Region 1",
				Metadata: `[{"key":"cluster-name","value":"test"},{"key":"app-id","value":"test2-value"}]`,
				TenantId: T2,
			},
		},
	}

	dao := inv_testing.NewInvResourceDAOOrFail(t)

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// create
	r1t1, err := dao.GetAPIClient().Create(ctx, T1, res1Tenant1)
	require.NoError(t, err, "CreateRegion() failed")
	r1t1ID := inv_testing.GetResourceIDOrFail(t, r1t1)
	res2Tenant1.GetRegion().ParentRegion = &location_v1.RegionResource{
		ResourceId: r1t1ID,
	}
	t.Cleanup(func() {
		_, derr := dao.GetAPIClient().Delete(context.TODO(), T1, r1t1ID)
		require.NoError(t, derr)
	})
	r2t1, err := dao.GetAPIClient().Create(ctx, T1, res2Tenant1)
	require.NoError(t, err, "CreateRegion() failed")
	r2t1ID := inv_testing.GetResourceIDOrFail(t, r2t1)
	t.Cleanup(func() {
		_, derr := dao.GetAPIClient().Delete(context.TODO(), T1, r2t1ID)
		require.NoError(t, derr)
	})
	r1t2, err := dao.GetAPIClient().Create(ctx, T2, res1Tenant2)
	require.NoError(t, err, "CreateRegion() failed")
	r1t2ID := inv_testing.GetResourceIDOrFail(t, r1t2)
	t.Cleanup(func() {
		_, derr := dao.GetAPIClient().Delete(context.TODO(), T2, r1t2ID)
		require.NoError(t, derr)
	})
	res2Tenant2.GetRegion().ParentRegion = &location_v1.RegionResource{
		ResourceId: r1t2ID,
	}
	r2t2, err := dao.GetAPIClient().Create(ctx, T2, res2Tenant2)
	require.NoError(t, err, "CreateRegion() failed")
	r2t2ID := inv_testing.GetResourceIDOrFail(t, r2t2)
	t.Cleanup(func() {
		_, derr := dao.GetAPIClient().Delete(context.TODO(), T2, r2t2ID)
		require.NoError(t, derr)
	})

	treeReqT1 := &inv_v1.GetTreeHierarchyRequest{
		Filter:   []string{r2t1ID},
		TenantId: T1,
	}
	resp, err := dao.GetAPIClient().GetTreeHierarchy(ctx, treeReqT1)
	require.NoError(t, err, "GetTreeHierarchy() failed")
	require.Len(t, resp, 2)

	treeReqT2 := &inv_v1.GetTreeHierarchyRequest{
		Filter:   []string{r2t2ID},
		TenantId: T2,
	}
	resp, err = dao.GetAPIClient().GetTreeHierarchy(ctx, treeReqT2)
	require.NoError(t, err, "GetTreeHierarchy() failed")
	require.Len(t, resp, 2)
}

func TestRegionMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)

	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				region := dao.CreateRegion(t, tenantID)
				res, err := util.WrapResource(region)
				require.NoError(t, err)
				return region.GetResourceId(), res
			},
		},
	})
}

func TestCreationOfResourceWithEdge(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	region1 := dao.CreateRegion(t, tenantIDZero)

	regionOwnedByT2CreationResp, err := dao.GetAPIClient().
		Create(
			context.TODO(),
			tenantIDOne, &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{
					Region: &location_v1.RegionResource{
						ParentRegion: region1,
						TenantId:     tenantIDOne,
					},
				},
			},
		)

	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err), "not found error expected")
	require.Nil(t, regionOwnedByT2CreationResp)
}

func TestDeleteResources_Regions(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				reg1 := dao.CreateRegionNoCleanup(t, tenantID)
				reg2 := dao.CreateRegionNoCleanup(t, tenantID)
				reg3 := dao.CreateRegionNoCleanup(t, tenantID, inv_testing.RegionParentRegion(reg2))
				reg4 := dao.CreateRegionNoCleanup(t, tenantID, inv_testing.RegionParentRegion(reg2))
				reg5 := dao.CreateRegionNoCleanup(t, tenantID, inv_testing.RegionParentRegion(reg3))
				return tenantID, len([]any{reg1, reg2, reg3, reg4, reg5})
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_REGION,
		},
	})
}
