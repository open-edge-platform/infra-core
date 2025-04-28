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
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPISiteToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs SiteTemplate defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPISiteToProto = map[string]string{
	locationv1.SiteResourceFieldSiteLat:  inv_locationv1.SiteResourceFieldSiteLat,
	locationv1.SiteResourceFieldSiteLng:  inv_locationv1.SiteResourceFieldSiteLng,
	locationv1.SiteResourceEdgeMetadata:  inv_locationv1.SiteResourceFieldMetadata,
	locationv1.SiteResourceFieldName:     inv_locationv1.SiteResourceFieldName,
	locationv1.SiteResourceFieldRegionId: inv_locationv1.SiteResourceEdgeRegion,
}

func toInvSite(site *locationv1.SiteResource) (*inv_locationv1.SiteResource, error) {
	var invSite *inv_locationv1.SiteResource
	if site == nil {
		return invSite, nil
	}
	metadata, err := toInvMetadata(site.GetMetadata())
	if err != nil {
		return nil, err
	}

	invSite = &inv_locationv1.SiteResource{
		ResourceId: site.GetResourceId(),
		Name:       site.GetName(),
		SiteLat:    site.GetSiteLat(),
		SiteLng:    site.GetSiteLng(),
		Metadata:   metadata,
	}

	regionID := site.GetRegionId()
	if isSet(&regionID) {
		invSite.Region = &inv_locationv1.RegionResource{
			ResourceId: regionID,
		}
	}

	err = validator.ValidateMessage(invSite)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	return invSite, nil
}

func fromInvSite(invSite *inv_locationv1.SiteResource,
	resMeta *inventory.GetResourceResponse_ResourceMetadata,
) (*locationv1.SiteResource, error) {
	if invSite == nil {
		return &locationv1.SiteResource{}, nil
	}

	region, err := fromInvRegion(invSite.GetRegion(), nil)
	if err != nil {
		return nil, err
	}

	metadata, err := fromInvMetadata(invSite.GetMetadata())
	if err != nil {
		return nil, err
	}

	site := &locationv1.SiteResource{
		ResourceId:        invSite.GetResourceId(),
		SiteId:            invSite.GetResourceId(),
		Name:              invSite.GetName(),
		Region:            region,
		RegionId:          region.GetResourceId(),
		SiteLat:           invSite.GetSiteLat(),
		SiteLng:           invSite.GetSiteLng(),
		Metadata:          metadata,
		InheritedMetadata: []*commonv1.MetadataItem{},
		Timestamps:        GrpcToOpenAPITimestamps(invSite),
	}

	if resMeta != nil {
		inheritedMetadata, err := fromInvMetadata(resMeta.GetPhyMetadata())
		if err != nil {
			return nil, err
		}
		site.InheritedMetadata = inheritedMetadata
	}
	return site, nil
}

func (is *InventorygRPCServer) CreateSite(
	ctx context.Context,
	req *restv1.CreateSiteRequest,
) (*locationv1.SiteResource, error) {
	zlog.Debug().Msg("CreateSite")

	site := req.GetSite()
	invSite, err := toInvSite(site)
	if err != nil {
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Site{
			Site: invSite,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to create inventory resource %s", invRes)
		return nil, err
	}

	siteCreated, err := fromInvSite(invResp.GetSite(), nil)
	if err != nil {
		return nil, err
	}
	zlog.Debug().Msgf("Created %s", siteCreated)
	return siteCreated, nil
}

// Get a list of sites.
func (is *InventorygRPCServer) ListSites(
	ctx context.Context,
	req *restv1.ListSitesRequest,
) (*restv1.ListSitesResponse, error) {
	zlog.Debug().Msg("ListSites")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Site{Site: &inv_locationv1.SiteResource{}}},
		Offset:   offset,
		Limit:    limit,
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to list inventory resources %s", filter)
		return nil, err
	}

	sites := []*locationv1.SiteResource{}
	for _, invRes := range invResp.GetResources() {
		site, err := fromInvSite(invRes.GetResource().GetSite(), invRes.GetRenderedMetadata())
		if err != nil {
			return nil, err
		}
		sites = append(sites, site)
	}

	resp := &restv1.ListSitesResponse{
		Sites:         sites,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific site.
func (is *InventorygRPCServer) GetSite(ctx context.Context, req *restv1.GetSiteRequest) (*locationv1.SiteResource, error) {
	zlog.Debug().Msg("GetSite")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to get inventory resource %s", req.GetResourceId())
		return nil, err
	}

	invSite := invResp.GetResource().GetSite()
	site, err := fromInvSite(invSite, invResp.GetRenderedMetadata())
	if err != nil {
		return nil, err
	}
	zlog.Debug().Msgf("Got %s", site)
	return site, nil
}

// Update a site. (PUT).
func (is *InventorygRPCServer) UpdateSite(
	ctx context.Context,
	req *restv1.UpdateSiteRequest,
) (*locationv1.SiteResource, error) {
	zlog.Debug().Msg("UpdateSite")

	site := req.GetSite()
	invSite, err := toInvSite(site)
	if err != nil {
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invSite, maps.Values(OpenAPISiteToProto)...)
	if err != nil {
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Site{
			Site: invSite,
		},
	}

	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetSite()
	invUpRes, err := fromInvSite(invUp, nil)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a site. (PATCH).
func (is *InventorygRPCServer) PatchSite(
	ctx context.Context,
	req *restv1.PatchSiteRequest,
) (*locationv1.SiteResource, error) {
	zlog.Debug().Msg("PatchSite")

	site := req.GetSite()
	invSite, err := toInvSite(site)
	if err != nil {
		return nil, err
	}

	fieldmask, err := parseFielmask(invSite, req.GetFieldMask(), OpenAPISiteToProto)
	if err != nil {
		return nil, err
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Site{
			Site: invSite,
		},
	}

	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetSite()
	invUpRes, err := fromInvSite(invUp, nil)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a site.
func (is *InventorygRPCServer) DeleteSite(
	ctx context.Context,
	req *restv1.DeleteSiteRequest,
) (*restv1.DeleteSiteResponse, error) {
	zlog.Debug().Msg("DeleteSite")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to delete inventory resource %s", req.GetResourceId())
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteSiteResponse{}, nil
}
