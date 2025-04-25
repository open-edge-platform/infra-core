// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	locationv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/location/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_utils "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var OpenAPIRegionToProto = map[string]string{
	"Name":     inv_locationv1.RegionResourceFieldName,
	"ParentId": inv_locationv1.RegionResourceEdgeParentRegion,
	// "Metadata": inv_locationv1.RegionResourceFieldMetadata,
}

var OpenAPIRegionObjectsNames = map[string]struct{}{
	"Metadata":          {},
	"InheritedMetadata": {},
	"ParentRegion":      {},
}

func toInvRegion(region *locationv1.RegionResource) (*inv_locationv1.RegionResource, error) {
	if region == nil {
		return &inv_locationv1.RegionResource{}, nil
	}
	var err error
	var metadata string
	if region.GetMetadata() != nil {
		metadata, err = toInvMetadata(region.GetMetadata())
		if err != nil {
			return nil, err
		}
	}
	invRegion := &inv_locationv1.RegionResource{
		ResourceId: region.GetResourceId(),
		Name:       region.GetName(),
		Metadata:   metadata,
	}

	parentID := region.GetParentId()
	if isSet(&parentID) {
		invRegion.ParentRegion = &inv_locationv1.RegionResource{
			ResourceId: parentID,
		}
	}

	err = validator.ValidateMessage(invRegion)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	return invRegion, nil
}

func fromInvRegion(
	invRegion *inv_locationv1.RegionResource,
	resMeta *inventory.GetResourceResponse_ResourceMetadata,
) (*locationv1.RegionResource, error) {
	if invRegion == nil {
		return &locationv1.RegionResource{}, nil
	}

	parentRegion, err := fromInvRegion(invRegion.GetParentRegion(), nil)
	if err != nil {
		return nil, err
	}
	metadata, err := fromInvMetadata(invRegion.GetMetadata())
	if err != nil {
		return nil, err
	}

	region := &locationv1.RegionResource{
		ResourceId:        invRegion.GetResourceId(),
		RegionId:          invRegion.GetResourceId(),
		Name:              invRegion.GetName(),
		ParentRegion:      parentRegion,
		ParentId:          parentRegion.GetResourceId(),
		Metadata:          metadata,
		InheritedMetadata: []*commonv1.MetadataItem{},
		CreatedAt:         invRegion.GetCreatedAt(),
		UpdatedAt:         invRegion.GetUpdatedAt(),
	}

	if resMeta != nil {
		inheritedMetadata, err := fromInvMetadata(resMeta.GetPhyMetadata())
		if err != nil {
			return nil, err
		}
		region.InheritedMetadata = inheritedMetadata
	}
	return region, nil
}

func (is *InventorygRPCServer) CreateRegion(
	ctx context.Context,
	req *restv1.CreateRegionRequest,
) (*locationv1.RegionResource, error) {
	zlog.Debug().Msg("CreateRegion")

	region := req.GetRegion()
	invRegion, err := toInvRegion(region)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory region")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Region{
			Region: invRegion,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create region in inventory")
		return nil, err
	}

	regionCreated, err := fromInvRegion(invResp.GetRegion(), nil)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory region")
		return nil, err
	}
	zlog.Debug().Msgf("Created %s", regionCreated)
	return regionCreated, nil
}

func (is *InventorygRPCServer) getSitesPerRegion(
	ctx context.Context,
	showTotalSites bool,
	resp *inventory.ListResourcesResponse,
) (map[string]int32, error) {
	sitesPerRegions := make(map[string]int32, len(resp.GetResources()))

	if !showTotalSites || len(resp.GetResources()) == 0 {
		return sitesPerRegions, nil
	}

	regionIDs := []string{}
	for _, res := range resp.GetResources() {
		resID, errGet := inv_utils.GetResourceIDFromResource(res.GetResource())
		if errGet != nil {
			zlog.InfraErr(errGet).Msgf("resource %v has no resourceID", res)
			continue
		}
		regionIDs = append(regionIDs, resID)
	}

	request := &inventory.GetSitesPerRegionRequest{
		Filter: regionIDs,
	}

	response, err := is.InvClient.GetSitesPerRegion(ctx, request)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get sites per region from inventory")
		return nil, err
	}

	collections.MapSlice[*inventory.GetSitesPerRegionResponse_Node, int32](
		response.GetRegions(),
		func(tN *inventory.GetSitesPerRegionResponse_Node) int32 {
			sitesPerRegions[tN.GetResourceId()] = tN.GetChildSites()
			return tN.GetChildSites()
		},
	)

	return sitesPerRegions, nil
}

// Get a list of regions.
func (is *InventorygRPCServer) ListRegions(
	ctx context.Context,
	req *restv1.ListRegionsRequest,
) (*restv1.ListRegionsResponse, error) {
	zlog.Debug().Msg("ListRegions")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Region{Region: &inv_locationv1.RegionResource{}}},
		Offset:   offset,
		Limit:    limit,
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list regions from inventory")
		return nil, err
	}

	sitesPerRegions, err := is.getSitesPerRegion(ctx, req.GetShowTotalSites(), invResp)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get sites per region")
		return nil, err
	}

	regions := []*locationv1.RegionResource{}
	for _, invRes := range invResp.GetResources() {
		region, err := fromInvRegion(invRes.GetResource().GetRegion(), invRes.GetRenderedMetadata())
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to convert from inventory region")
			return nil, err
		}

		if totalSites, hasTotalSites := sitesPerRegions[region.GetResourceId()]; hasTotalSites {
			region.TotalSites = totalSites
		}
		regions = append(regions, region)
	}

	resp := &restv1.ListRegionsResponse{
		Regions:       regions,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific region.
func (is *InventorygRPCServer) GetRegion(
	ctx context.Context,
	req *restv1.GetRegionRequest,
) (*locationv1.RegionResource, error) {
	zlog.Debug().Msg("GetRegion")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get region from inventory")
		return nil, err
	}

	invRegion := invResp.GetResource().GetRegion()
	region, err := fromInvRegion(invRegion, invResp.GetRenderedMetadata())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory region")
		return nil, err
	}
	zlog.Debug().Msgf("Got %s", region)
	return region, nil
}

// Update a region. (PUT).
func (is *InventorygRPCServer) UpdateRegion(
	ctx context.Context,
	req *restv1.UpdateRegionRequest,
) (*locationv1.RegionResource, error) {
	zlog.Debug().Msg("UpdateRegion")

	region := req.GetRegion()
	invRegion, err := toInvRegion(region)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory region")
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invRegion, maps.Values(OpenAPIRegionToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Region{
			Region: invRegion,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetRegion()
	invUpRes, err := fromInvRegion(invUp, nil)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a region. (PATCH).
func (is *InventorygRPCServer) PatchRegion(
	ctx context.Context,
	req *restv1.PatchRegionRequest,
) (*locationv1.RegionResource, error) {
	zlog.Debug().Msg("PatchRegion")

	region := req.GetRegion()
	invRegion, err := toInvRegion(region)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory region")
		return nil, err
	}

	fieldmask := req.GetFieldMask()
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Region{
			Region: invRegion,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetRegion()
	invUpRes, err := fromInvRegion(invUp, nil)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a region.
func (is *InventorygRPCServer) DeleteRegion(
	ctx context.Context,
	req *restv1.DeleteRegionRequest,
) (*restv1.DeleteRegionResponse, error) {
	zlog.Debug().Msg("DeleteRegion")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete region from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteRegionResponse{}, nil
}
