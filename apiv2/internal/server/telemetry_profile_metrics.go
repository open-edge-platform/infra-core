// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	telemetryv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/telemetry/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPITelemetryMetricsProfileToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs TelemetryLogsProfile defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPITelemetryMetricsProfileToProto = map[string]string{
	"targetRegion":    inv_telemetryv1.TelemetryProfileEdgeRegion,
	"targetSite":      inv_telemetryv1.TelemetryProfileEdgeSite,
	"targetInstance":  inv_telemetryv1.TelemetryProfileEdgeInstance,
	"metricsInterval": inv_telemetryv1.TelemetryProfileFieldMetricsInterval,
	"metricsGroupId":  inv_telemetryv1.TelemetryProfileEdgeGroup,
}

func TelemetryMetricsProfileResourcetoAPI(
	telemetryProfile *inv_telemetryv1.TelemetryProfile,
) *telemetryv1.TelemetryMetricsProfileResource {
	if telemetryProfile == nil {
		return nil
	}

	telemetryMetricsProfile := &telemetryv1.TelemetryMetricsProfileResource{
		ResourceId:      telemetryProfile.GetResourceId(),
		ProfileId:       telemetryProfile.GetResourceId(),
		MetricsInterval: telemetryProfile.GetMetricsInterval(),
	}

	if telemetryProfile.GetInstance() != nil {
		resInstID := telemetryProfile.GetInstance().GetResourceId()
		telemetryMetricsProfile.TargetInstance = resInstID
	}

	if telemetryProfile.GetSite() != nil {
		resSiteID := telemetryProfile.GetSite().GetResourceId()
		telemetryMetricsProfile.TargetSite = resSiteID
	}

	if telemetryProfile.GetRegion() != nil {
		resRegionID := telemetryProfile.GetRegion().GetResourceId()
		telemetryMetricsProfile.TargetRegion = resRegionID
	}

	if telemetryProfile.GetGroup() != nil {
		telemetryMetricsProfile.MetricsGroup = TelemetryMetricsGroupResourcetoAPI(telemetryProfile.GetGroup())
		telemetryMetricsProfile.MetricsGroupId = telemetryProfile.GetGroup().GetResourceId()
	}

	return telemetryMetricsProfile
}

func TelemetryMetricsProfileResourcetoGRPC(
	telemetryMetricsProfile *telemetryv1.TelemetryMetricsProfileResource,
) (*inv_telemetryv1.TelemetryProfile, error) {
	if telemetryMetricsProfile == nil {
		return &inv_telemetryv1.TelemetryProfile{}, nil
	}
	telemetryProfile := &inv_telemetryv1.TelemetryProfile{
		Kind: inv_telemetryv1.TelemetryResourceKind(*telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS.Enum()),
		Group: &inv_telemetryv1.TelemetryGroupResource{
			ResourceId: telemetryMetricsProfile.GetMetricsGroupId(),
		},
		MetricsInterval: telemetryMetricsProfile.GetMetricsInterval(),
	}

	err := validateAndSetTelemetryProfileRelations(
		telemetryProfile,
		telemetryMetricsProfile.GetTargetInstance(),
		telemetryMetricsProfile.GetTargetSite(),
		telemetryMetricsProfile.GetTargetRegion(),
	)
	if err != nil {
		return nil, err
	}

	err = validator.ValidateMessage(telemetryProfile)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return telemetryProfile, nil
}

func (is *InventorygRPCServer) CreateTelemetryMetricsProfile(
	ctx context.Context,
	req *restv1.CreateTelemetryMetricsProfileRequest,
) (*telemetryv1.TelemetryMetricsProfileResource, error) {
	zlog.Debug().Msg("CreateTelemetryMetricsProfile")

	telemetryMetricsProfile := req.GetTelemetryMetricsProfile()
	telemetryProfile, err := TelemetryMetricsProfileResourcetoGRPC(telemetryMetricsProfile)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory telemetry metrics profile")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryProfile{
			TelemetryProfile: telemetryProfile,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create telemetry metrics profile in inventory")
		return nil, err
	}

	telemetryMetricsProfileCreated := TelemetryMetricsProfileResourcetoAPI(invResp.GetTelemetryProfile())
	zlog.Debug().Msgf("Created %s", telemetryMetricsProfileCreated)
	return telemetryMetricsProfileCreated, nil
}

// Get a list of telemetryMetricsProfiles.
func (is *InventorygRPCServer) ListTelemetryMetricsProfiles(
	ctx context.Context,
	req *restv1.ListTelemetryMetricsProfilesRequest,
) (*restv1.ListTelemetryMetricsProfilesResponse, error) {
	zlog.Debug().Msg("ListTelemetryMetricsProfiles")

	hasNext := false
	var totalElems int32
	telemetryMetricsProfiles := []*telemetryv1.TelemetryMetricsProfileResource{}
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := telemetryProfileFilter(
		telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		req.GetInstanceId(),
		req.GetSiteId(),
		req.GetRegionId())

	invReq := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_TelemetryProfile{},
		},
		Offset:  offset,
		Limit:   limit,
		OrderBy: req.GetOrderBy(),
		Filter:  filter,
	}

	if !req.GetShowInherited() {
		resp, err := is.InvClient.List(ctx, invReq)
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to list telemetry metrics profiles from inventory")
			return nil, err
		}

		for _, res := range resp.GetResources() {
			telemetryProfile := res.GetResource().GetTelemetryProfile()
			telemetryMetricsProfile := TelemetryMetricsProfileResourcetoAPI(telemetryProfile)
			telemetryMetricsProfiles = append(telemetryMetricsProfiles, telemetryMetricsProfile)
		}
		hasNext = resp.GetHasNext()
		totalElems = resp.GetTotalElements()
	} else {
		var telProfiles []*inv_telemetryv1.TelemetryProfile
		var err error
		telProfiles, totalElems, hasNext, err = is.listInheritedTelemetryMetrics(ctx, req)
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to list inherited telemetry metrics profiles")
			return nil, err
		}
		for _, telemetryProfile := range telProfiles {
			telemetryMetricsProfile := TelemetryMetricsProfileResourcetoAPI(telemetryProfile)
			telemetryMetricsProfiles = append(telemetryMetricsProfiles, telemetryMetricsProfile)
		}
	}

	resp := &restv1.ListTelemetryMetricsProfilesResponse{
		TelemetryMetricsProfiles: telemetryMetricsProfiles,
		TotalElements:            totalElems,
		HasNext:                  hasNext,
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific telemetryMetricsProfile.
func (is *InventorygRPCServer) GetTelemetryMetricsProfile(
	ctx context.Context,
	req *restv1.GetTelemetryMetricsProfileRequest,
) (*telemetryv1.TelemetryMetricsProfileResource, error) {
	zlog.Debug().Msg("GetTelemetryMetricsProfile")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get telemetry metrics profile from inventory")
		return nil, err
	}

	telemetryProfile := invResp.GetResource().GetTelemetryProfile()
	telemetryMetricsProfile := TelemetryMetricsProfileResourcetoAPI(telemetryProfile)
	zlog.Debug().Msgf("Got %s", telemetryMetricsProfile)
	return telemetryMetricsProfile, nil
}

// Delete a telemetryMetricsProfile.
func (is *InventorygRPCServer) DeleteTelemetryMetricsProfile(
	ctx context.Context,
	req *restv1.DeleteTelemetryMetricsProfileRequest,
) (*restv1.DeleteTelemetryMetricsProfileResponse, error) {
	zlog.Debug().Msg("DeleteTelemetryMetricsProfile")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete telemetry metrics profile from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteTelemetryMetricsProfileResponse{}, nil
}

// Update a telemetryMetricsProfile. (PUT).
func (is *InventorygRPCServer) UpdateTelemetryMetricsProfile(
	ctx context.Context,
	req *restv1.UpdateTelemetryMetricsProfileRequest,
) (*telemetryv1.TelemetryMetricsProfileResource, error) {
	zlog.Debug().Msg("UpdateTelemetryMetricsProfile")

	telemetryMetricsProfile := req.GetTelemetryMetricsProfile()
	telemetryProfile, err := TelemetryMetricsProfileResourcetoGRPC(telemetryMetricsProfile)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory telemetry metrics profile")
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(
		telemetryProfile,
		maps.Values(OpenAPITelemetryMetricsProfileToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryProfile{
			TelemetryProfile: telemetryProfile,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetTelemetryProfile()
	invUpRes := TelemetryMetricsProfileResourcetoAPI(invUp)
	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Should be called only when requesting inherited telemetry logs.
func (is *InventorygRPCServer) listInheritedTelemetryMetrics(
	ctx context.Context,
	req *restv1.ListTelemetryMetricsProfilesRequest,
) (telProfiles []*inv_telemetryv1.TelemetryProfile, totalElements int32, more bool, err error) {
	err = validateTelemetryProfileRelations(req.GetInstanceId(), req.GetSiteId(), req.GetRegionId(), false)
	if err != nil {
		return nil, 0, false, err
	}
	var inheritBy inventory.ListInheritedTelemetryProfilesRequest_InheritBy
	switch {
	case req.GetInstanceId() != "":
		inheritBy = inventory.ListInheritedTelemetryProfilesRequest_InheritBy{
			Id: &inventory.ListInheritedTelemetryProfilesRequest_InheritBy_InstanceId{
				InstanceId: req.GetInstanceId(),
			},
		}
	case req.GetSiteId() != "":
		inheritBy = inventory.ListInheritedTelemetryProfilesRequest_InheritBy{
			Id: &inventory.ListInheritedTelemetryProfilesRequest_InheritBy_SiteId{
				SiteId: req.GetSiteId(),
			},
		}
	case req.GetRegionId() != "":
		inheritBy = inventory.ListInheritedTelemetryProfilesRequest_InheritBy{
			Id: &inventory.ListInheritedTelemetryProfilesRequest_InheritBy_RegionId{
				RegionId: req.GetRegionId(),
			},
		}
	}
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, 0, false, err
	}
	resp, err := is.InvClient.ListInheritedTelemetryProfiles(
		ctx,
		&inheritBy,
		fmt.Sprintf("%s = %s",
			inv_telemetryv1.TelemetryProfileFieldKind,
			inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS),
		req.GetOrderBy(), offset, limit)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list inherited telemetry metrics profiles from inventory")
		return nil, 0, false, err
	}
	telProfiles = resp.GetTelemetryProfiles()
	totalElements = resp.GetTotalElements()
	// Safe to cast to offset to int since it comes from an int already.
	more = int(req.GetOffset())+len(resp.GetTelemetryProfiles()) < int(totalElements)

	return telProfiles, totalElements, more, nil
}
