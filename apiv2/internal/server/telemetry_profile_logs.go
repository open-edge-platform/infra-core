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
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPITelemetryLogsProfileToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs TelemetryLogsProfile defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPITelemetryLogsProfileToProto = map[string]string{
	telemetryv1.TelemetryLogsProfileResourceFieldTargetRegion:   inv_telemetryv1.TelemetryProfileEdgeRegion,
	telemetryv1.TelemetryLogsProfileResourceFieldTargetSite:     inv_telemetryv1.TelemetryProfileEdgeSite,
	telemetryv1.TelemetryLogsProfileResourceFieldTargetInstance: inv_telemetryv1.TelemetryProfileEdgeInstance,
	telemetryv1.TelemetryLogsProfileResourceFieldLogLevel:       inv_telemetryv1.TelemetryProfileFieldLogLevel,
	telemetryv1.TelemetryLogsProfileResourceFieldLogsGroupId:    inv_telemetryv1.TelemetryProfileEdgeGroup,
}

func TelemetryLogsProfileResourcetoAPI(
	telemetryProfile *inv_telemetryv1.TelemetryProfile,
) *telemetryv1.TelemetryLogsProfileResource {
	if telemetryProfile == nil {
		return nil
	}

	telemetryLogsProfile := &telemetryv1.TelemetryLogsProfileResource{
		ResourceId: telemetryProfile.GetResourceId(),
		ProfileId:  telemetryProfile.GetResourceId(),
		LogLevel:   telemetryv1.SeverityLevel(*telemetryProfile.GetLogLevel().Enum()),
		Timestamps: GrpcToOpenAPITimestamps(telemetryProfile),
	}

	if telemetryProfile.GetInstance() != nil {
		resInstID := telemetryProfile.GetInstance().GetResourceId()
		telemetryLogsProfile.TargetInstance = resInstID
	}

	if telemetryProfile.GetSite() != nil {
		resSiteID := telemetryProfile.GetSite().GetResourceId()
		telemetryLogsProfile.TargetSite = resSiteID
	}

	if telemetryProfile.GetRegion() != nil {
		resRegionID := telemetryProfile.GetRegion().GetResourceId()
		telemetryLogsProfile.TargetRegion = resRegionID
	}

	if telemetryProfile.GetGroup() != nil {
		telemetryLogsProfile.LogsGroup = TelemetryLogsGroupResourcetoAPI(telemetryProfile.GetGroup())
		telemetryLogsProfile.LogsGroupId = telemetryProfile.GetGroup().GetResourceId()
	}

	return telemetryLogsProfile
}

func TelemetryLogsProfileResourcetoGRPC(
	telemetryLogsProfile *telemetryv1.TelemetryLogsProfileResource,
) (*inv_telemetryv1.TelemetryProfile, error) {
	if telemetryLogsProfile == nil {
		return &inv_telemetryv1.TelemetryProfile{}, nil
	}
	telemetryProfile := &inv_telemetryv1.TelemetryProfile{
		Kind: inv_telemetryv1.TelemetryResourceKind(*telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS.Enum()),
		Group: &inv_telemetryv1.TelemetryGroupResource{
			ResourceId: telemetryLogsProfile.GetLogsGroupId(),
		},
		LogLevel: inv_telemetryv1.SeverityLevel(*telemetryLogsProfile.GetLogLevel().Enum()),
	}

	err := validateAndSetTelemetryProfileRelations(
		telemetryProfile,
		telemetryLogsProfile.GetTargetInstance(),
		telemetryLogsProfile.GetTargetSite(),
		telemetryLogsProfile.GetTargetRegion(),
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

func (is *InventorygRPCServer) CreateTelemetryLogsProfile(
	ctx context.Context,
	req *restv1.CreateTelemetryLogsProfileRequest,
) (*telemetryv1.TelemetryLogsProfileResource, error) {
	zlog.Debug().Msg("CreateTelemetryLogsProfile")

	telemetryLogsProfile := req.GetTelemetryLogsProfile()
	telemetryProfile, err := TelemetryLogsProfileResourcetoGRPC(telemetryLogsProfile)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory telemetry logs profile")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryProfile{
			TelemetryProfile: telemetryProfile,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create telemetry logs profile in inventory")
		return nil, errors.Wrap(err)
	}

	telemetryLogsProfileCreated := TelemetryLogsProfileResourcetoAPI(invResp.GetTelemetryProfile())
	zlog.Debug().Msgf("Created %s", telemetryLogsProfileCreated)
	return telemetryLogsProfileCreated, nil
}

// Get a list of telemetryLogsProfiles.
func (is *InventorygRPCServer) ListTelemetryLogsProfiles(
	ctx context.Context,
	req *restv1.ListTelemetryLogsProfilesRequest,
) (*restv1.ListTelemetryLogsProfilesResponse, error) {
	zlog.Debug().Msg("ListTelemetryLogsProfiles")

	hasNext := false
	var totalElems int32
	telemetryLogsProfiles := []*telemetryv1.TelemetryLogsProfileResource{}

	filter := telemetryProfileFilter(
		telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		req.GetInstanceId(),
		req.GetSiteId(),
		req.GetRegionId())

	invReq := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_TelemetryProfile{},
		},
		Offset:  req.GetOffset(),
		Limit:   req.GetPageSize(),
		OrderBy: req.GetOrderBy(),
		Filter:  filter,
	}
	if err := validator.ValidateMessage(invReq); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	if !req.GetShowInherited() {
		resp, err := is.InvClient.List(ctx, invReq)
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to list telemetry logs profiles from inventory")
			return nil, errors.Wrap(err)
		}

		for _, res := range resp.GetResources() {
			telemetryProfile := res.GetResource().GetTelemetryProfile()
			telemetryLogsProfile := TelemetryLogsProfileResourcetoAPI(telemetryProfile)
			telemetryLogsProfiles = append(telemetryLogsProfiles, telemetryLogsProfile)
		}
		hasNext = resp.GetHasNext()
		totalElems = resp.GetTotalElements()
	} else {
		var telProfiles []*inv_telemetryv1.TelemetryProfile
		var err error
		telProfiles, totalElems, hasNext, err = is.listInheritedTelemetryLogs(ctx, req)
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to list inherited telemetry logs profiles")
			return nil, err
		}
		for _, telemetryProfile := range telProfiles {
			telemetryLogsProfile := TelemetryLogsProfileResourcetoAPI(telemetryProfile)
			telemetryLogsProfiles = append(telemetryLogsProfiles, telemetryLogsProfile)
		}
	}

	resp := &restv1.ListTelemetryLogsProfilesResponse{
		TelemetryLogsProfiles: telemetryLogsProfiles,
		TotalElements:         totalElems,
		HasNext:               hasNext,
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific telemetryLogsProfile.
func (is *InventorygRPCServer) GetTelemetryLogsProfile(
	ctx context.Context,
	req *restv1.GetTelemetryLogsProfileRequest,
) (*telemetryv1.TelemetryLogsProfileResource, error) {
	zlog.Debug().Msg("GetTelemetryLogsProfile")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get telemetry logs profile from inventory")
		return nil, errors.Wrap(err)
	}

	telemetryProfile := invResp.GetResource().GetTelemetryProfile()
	telemetryLogsProfile := TelemetryLogsProfileResourcetoAPI(telemetryProfile)
	zlog.Debug().Msgf("Got %s", telemetryLogsProfile)
	return telemetryLogsProfile, nil
}

// Delete a telemetryLogsProfile.
func (is *InventorygRPCServer) DeleteTelemetryLogsProfile(
	ctx context.Context,
	req *restv1.DeleteTelemetryLogsProfileRequest,
) (*restv1.DeleteTelemetryLogsProfileResponse, error) {
	zlog.Debug().Msg("DeleteTelemetryLogsProfile")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete telemetry logs profile from inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteTelemetryLogsProfileResponse{}, nil
}

// Update a telemetryLogsProfile. (PUT).
func (is *InventorygRPCServer) UpdateTelemetryLogsProfile(
	ctx context.Context,
	req *restv1.UpdateTelemetryLogsProfileRequest,
) (*telemetryv1.TelemetryLogsProfileResource, error) {
	zlog.Debug().Msg("UpdateTelemetryLogsProfile")

	telemetryLogsProfile := req.GetTelemetryLogsProfile()
	telemetryProfile, err := TelemetryLogsProfileResourcetoGRPC(telemetryLogsProfile)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory telemetry logs profile")
		return nil, errors.Wrap(err)
	}

	fieldmask, err := fieldmaskpb.New(telemetryProfile, maps.Values(OpenAPITelemetryLogsProfileToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryProfile{
			TelemetryProfile: telemetryProfile,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, errors.Wrap(err)
	}
	invUp := upRes.GetTelemetryProfile()
	invUpRes := TelemetryLogsProfileResourcetoAPI(invUp)
	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a telemetryLogsProfile. (PATCH).
func (is *InventorygRPCServer) PatchTelemetryLogsProfile(
	ctx context.Context,
	req *restv1.PatchTelemetryLogsProfileRequest,
) (*telemetryv1.TelemetryLogsProfileResource, error) {
	zlog.Debug().Msg("PatchTelemetryLogsProfile")

	telemetryLogsProfile := req.GetTelemetryLogsProfile()
	telemetryProfile, err := TelemetryLogsProfileResourcetoGRPC(telemetryLogsProfile)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory telemetry logs profile")
		return nil, errors.Wrap(err)
	}

	fieldmask, err := parseFielmask(telemetryProfile, req.GetFieldMask(), OpenAPITelemetryLogsProfileToProto)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryProfile{
			TelemetryProfile: telemetryProfile,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, errors.Wrap(err)
	}
	invUp := upRes.GetTelemetryProfile()
	invUpRes := TelemetryLogsProfileResourcetoAPI(invUp)
	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Should be called only when requesting inherited telemetry logs.
func (is *InventorygRPCServer) listInheritedTelemetryLogs(
	ctx context.Context,
	req *restv1.ListTelemetryLogsProfilesRequest,
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
	resp, err := is.InvClient.ListInheritedTelemetryProfiles(
		ctx,
		&inheritBy,
		fmt.Sprintf("%s = %s",
			inv_telemetryv1.TelemetryProfileFieldKind,
			inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS),
		req.GetOrderBy(), req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list inherited telemetry logs profiles from inventory")
		return nil, 0, false, errors.Wrap(err)
	}
	telProfiles = resp.GetTelemetryProfiles()
	totalElements = resp.GetTotalElements()
	// Safe to cast to offset to int since it comes from an int already.
	more = int(req.GetOffset())+len(resp.GetTelemetryProfiles()) < int(totalElements)

	return telProfiles, totalElements, more, nil
}
