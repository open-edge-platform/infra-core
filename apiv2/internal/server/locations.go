// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

const (
	defaultOffset = 0
	defaultLimit  = 50
)

func (is *InventorygRPCServer) validateQueryParams(req *restv1.ListLocationsRequest) error {
	err := errors.Errorfc(codes.InvalidArgument, "GetLocationsParams incorrectly provided")
	if req.GetName() == "" {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate GetLocationsParams, missing name parameter")
		return err
	}

	if !req.GetShowRegions() && !req.GetShowSites() {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate GetLocationsParams, " +
			"at least one of showSites or showRegions must be provided")
		return err
	}

	return nil
}

func (is *InventorygRPCServer) getLocationIDs(
	ctx context.Context,
	req *restv1.ListLocationsRequest,
	resType inventory.ResourceKind,
) (locationIDs []string, totalElements int32, err error) {
	var resource *inventory.Resource
	switch resType {
	case inventory.ResourceKind_RESOURCE_KIND_REGION:
		resource = &inventory.Resource{Resource: &inventory.Resource_Region{Region: &inv_locationv1.RegionResource{}}}
	case inventory.ResourceKind_RESOURCE_KIND_SITE:
		resource = &inventory.Resource{Resource: &inventory.Resource_Site{Site: &inv_locationv1.SiteResource{}}}
	default:
		err = errors.Errorfc(codes.Internal, "invalid resource type %v", resType)
		zlog.InfraErr(err).Msg("invalid resource type")
		return nil, 0, errors.Wrap(err)
	}

	filterbyName := fmt.Sprintf(`%s = %q`, "name", req.GetName())
	filter := &inventory.ResourceFilter{
		Resource: resource,
		Filter:   filterbyName,
		Offset:   defaultOffset,
		Limit:    defaultLimit,
		OrderBy:  "name asc",
	}
	if err = validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, 0, errors.Wrap(err)
	}

	findResp, err := is.InvClient.Find(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to find locations in inventory")
		return nil, 0, errors.Wrap(err)
	}

	locationIDs = collections.MapSlice[*client.ResourceTenantIDCarrier, string](findResp.GetResources(),
		func(carrier *client.ResourceTenantIDCarrier) string {
			return carrier.GetResourceId()
		})
	totalElements = findResp.GetTotalElements()
	return locationIDs, totalElements, nil
}

func getLocationType(invKind inventory.ResourceKind) (restv1.ListLocationsResponse_ResourceKind, error) {
	locMap := map[inventory.ResourceKind]restv1.ListLocationsResponse_ResourceKind{
		inventory.ResourceKind_RESOURCE_KIND_REGION: restv1.ListLocationsResponse_RESOURCE_KIND_REGION,
		inventory.ResourceKind_RESOURCE_KIND_SITE:   restv1.ListLocationsResponse_RESOURCE_KIND_SITE,
	}

	apiType, ok := locMap[invKind]
	if !ok {
		err := errors.Errorfc(codes.Internal, "invalid location type %v", invKind)
		zlog.InfraErr(err).Msg("failed to find locations type")
		return restv1.ListLocationsResponse_RESOURCE_KIND_UNSPECIFIED, errors.Wrap(err)
	}
	return apiType, nil
}

func getParentID(locNode *inventory.GetTreeHierarchyResponse_TreeNode) string {
	// Filters parentIDs (removes possible OU resources)
	parentIDs := []string{}
	for _, parentNode := range locNode.GetParentNodes() {
		if parentNode.ResourceKind == inventory.ResourceKind_RESOURCE_KIND_REGION ||
			parentNode.ResourceKind == inventory.ResourceKind_RESOURCE_KIND_SITE {
			parentIDs = append(parentIDs, parentNode.GetResourceId())
		}
	}
	var parentID string
	// Region and Site have only one parentID, when existent.
	if len(parentIDs) > 0 {
		parentID = parentIDs[0]
	} else if len(parentIDs) > 1 {
		zlog.Warn().Msgf("Invalid location parent IDs resource %s"+
			"has more than one parent of type region or site %v",
			locNode.CurrentNode.GetResourceId(), parentIDs)
	}
	return parentID
}

func inventoryTreeToAPITree(
	treeResp []*inventory.GetTreeHierarchyResponse_TreeNode,
) ([]*restv1.ListLocationsResponse_LocationNode, error) {
	// Filters tree response (removes possible OU or Host resources in the response)
	locationNodes := []*inventory.GetTreeHierarchyResponse_TreeNode{}
	for _, node := range treeResp {
		if node.CurrentNode.ResourceKind == inventory.ResourceKind_RESOURCE_KIND_REGION ||
			node.CurrentNode.ResourceKind == inventory.ResourceKind_RESOURCE_KIND_SITE {
			locationNodes = append(locationNodes, node)
		}
	}
	apiLocNodes := make([]*restv1.ListLocationsResponse_LocationNode, 0, len(locationNodes))
	for _, locNode := range locationNodes {
		resID := locNode.GetCurrentNode().GetResourceId()
		resName := locNode.GetName()
		resType, err := getLocationType(locNode.GetCurrentNode().GetResourceKind())
		if err != nil {
			return nil, err
		}

		parentID := getParentID(locNode)
		locationNode := &restv1.ListLocationsResponse_LocationNode{
			ResourceId: resID,
			Name:       resName,
			ParentId:   parentID,
			Type:       resType,
		}
		apiLocNodes = append(apiLocNodes, locationNode)
	}

	return apiLocNodes, nil
}

func (is *InventorygRPCServer) getLocationTree(
	ctx context.Context,
	locationIDs []string,
) (*restv1.ListLocationsResponse, error) {
	zlog.Debug().Msg("getLocationTree")

	apiLocList := &restv1.ListLocationsResponse{
		Nodes: []*restv1.ListLocationsResponse_LocationNode{},
	}

	if len(locationIDs) == 0 {
		return apiLocList, nil
	}

	request := &inventory.GetTreeHierarchyRequest{
		Filter:     locationIDs,
		Descending: true,
	}
	treeResp, err := is.InvClient.GetTreeHierarchy(ctx, request)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to get tree hierarchy from inventory")
		return apiLocList, errors.Wrap(err)
	}

	apiLocNodes, err := inventoryTreeToAPITree(treeResp)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to convert inventory tree to API tree")
		return apiLocList, err
	}

	apiLocList.Nodes = apiLocNodes
	zlog.Debug().Msgf("LocationTree %s", apiLocList)
	return apiLocList, nil
}

func (is *InventorygRPCServer) getAllLocationIDs(
	ctx context.Context,
	req *restv1.ListLocationsRequest,
) (allIDs []string, totalElements int32, outputElements int, err error) {
	allIDs = []string{}
	outputElements = 0
	if req.GetShowRegions() {
		regionIDs, totalRegions, errReg := is.getLocationIDs(ctx, req, inventory.ResourceKind_RESOURCE_KIND_REGION)
		if errReg != nil {
			return nil, 0, 0, errReg
		}

		allIDs = append(allIDs, regionIDs...)
		totalElements += totalRegions
		outputElements += len(regionIDs)
	}

	if req.GetShowSites() {
		siteIDs, totalSites, errSit := is.getLocationIDs(ctx, req, inventory.ResourceKind_RESOURCE_KIND_SITE)
		if errSit != nil {
			return nil, 0, 0, errSit
		}
		allIDs = append(allIDs, siteIDs...)
		totalElements += totalSites
		outputElements += len(siteIDs)
	}

	return allIDs, totalElements, outputElements, nil
}

// Get a list of regions.
func (is *InventorygRPCServer) ListLocations(
	ctx context.Context,
	req *restv1.ListLocationsRequest,
) (*restv1.ListLocationsResponse, error) {
	zlog.Debug().Msg("ListLocations")

	err := is.validateQueryParams(req)
	if err != nil {
		return nil, err
	}

	allIDs, totalElements, outputElements, err := is.getAllLocationIDs(ctx, req)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to get all location IDs")
		return nil, err
	}

	resp, err := is.getLocationTree(ctx, allIDs)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to get location tree")
		return nil, err
	}

	resp.TotalElements = totalElements
	outElements, err := util.IntToInt32(outputElements)
	if err != nil {
		zlog.InfraErr(err).Msg("failed to convert output elements to int32")
		return nil, errors.Wrap(err)
	}
	resp.OutputElements = outElements
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}
