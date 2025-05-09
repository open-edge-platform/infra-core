// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	telemetryv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/telemetry/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

func TelemetryLogsGroupResourcetoAPI(
	telemetryGroup *inv_telemetryv1.TelemetryGroupResource,
) *telemetryv1.TelemetryLogsGroupResource {
	if telemetryGroup == nil {
		return nil
	}
	telemetryLogsGroup := &telemetryv1.TelemetryLogsGroupResource{
		ResourceId:           telemetryGroup.GetResourceId(),
		TelemetryLogsGroupId: telemetryGroup.GetResourceId(),
		Name:                 telemetryGroup.GetName(),
		CollectorKind:        telemetryv1.TelemetryCollectorKind(*telemetryGroup.GetCollectorKind().Enum()),
		Groups:               telemetryGroup.GetGroups(),
		Timestamps:           GrpcToOpenAPITimestamps(telemetryGroup),
	}
	return telemetryLogsGroup
}

func (is *InventorygRPCServer) CreateTelemetryLogsGroup(
	ctx context.Context,
	req *restv1.CreateTelemetryLogsGroupRequest,
) (*telemetryv1.TelemetryLogsGroupResource, error) {
	zlog.Debug().Msg("CreateTelemetryLogsGroup")

	telemetryLogsGroup := req.GetTelemetryLogsGroup()

	telemetryGroup := &inv_telemetryv1.TelemetryGroupResource{
		Name:   telemetryLogsGroup.GetName(),
		Groups: telemetryLogsGroup.GetGroups(),
		Kind: inv_telemetryv1.TelemetryResourceKind(
			*telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS.Enum(),
		),
		CollectorKind: inv_telemetryv1.CollectorKind(*telemetryLogsGroup.GetCollectorKind().Enum()),
	}

	err := validator.ValidateMessage(telemetryGroup)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryGroup{
			TelemetryGroup: telemetryGroup,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create telemetry logs group in inventory")
		return nil, errors.Wrap(err)
	}
	telemetryLogsGroupCreated := TelemetryLogsGroupResourcetoAPI(invResp.GetTelemetryGroup())
	zlog.Debug().Msgf("Created %s", telemetryLogsGroupCreated)
	return telemetryLogsGroupCreated, nil
}

// Get a list of telemetryLogsGroups.
func (is *InventorygRPCServer) ListTelemetryLogsGroups(
	ctx context.Context,
	req *restv1.ListTelemetryLogsGroupsRequest,
) (*restv1.ListTelemetryLogsGroupsResponse, error) {
	zlog.Debug().Msg("ListTelemetryLogsGroups")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, errors.Wrap(err)
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_TelemetryGroup{
				TelemetryGroup: &inv_telemetryv1.TelemetryGroupResource{},
			},
		},
		Offset:  offset,
		Limit:   limit,
		OrderBy: req.GetOrderBy(),
		Filter: fmt.Sprintf("%s = %s", inv_telemetryv1.TelemetryGroupResourceFieldKind,
			inv_telemetryv1.TelemetryResourceKind_name[int32(
				inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS)]),
	}
	if err = validator.ValidateMessage(filter); err != nil {
		zlog.InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list telemetry logs groups from inventory")
		return nil, errors.Wrap(err)
	}

	telemetryLogsGroups := []*telemetryv1.TelemetryLogsGroupResource{}
	for _, invRes := range invResp.GetResources() {
		telemetryGroup := invRes.GetResource().GetTelemetryGroup()
		telemetryLogsGroup := TelemetryLogsGroupResourcetoAPI(telemetryGroup)
		telemetryLogsGroups = append(telemetryLogsGroups, telemetryLogsGroup)
	}

	resp := &restv1.ListTelemetryLogsGroupsResponse{
		TelemetryLogsGroups: telemetryLogsGroups,
		TotalElements:       invResp.GetTotalElements(),
		HasNext:             invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific telemetryLogsGroup.
func (is *InventorygRPCServer) GetTelemetryLogsGroup(
	ctx context.Context,
	req *restv1.GetTelemetryLogsGroupRequest,
) (*telemetryv1.TelemetryLogsGroupResource, error) {
	zlog.Debug().Msg("GetTelemetryLogsGroup")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get telemetry logs group from inventory")
		return nil, errors.Wrap(err)
	}

	telemetryGroup := invResp.GetResource().GetTelemetryGroup()
	telemetryLogsGroup := TelemetryLogsGroupResourcetoAPI(telemetryGroup)
	zlog.Debug().Msgf("Got %s", telemetryLogsGroup)
	return telemetryLogsGroup, nil
}

// Delete a telemetryLogsGroup.
func (is *InventorygRPCServer) DeleteTelemetryLogsGroup(
	ctx context.Context,
	req *restv1.DeleteTelemetryLogsGroupRequest,
) (*restv1.DeleteTelemetryLogsGroupResponse, error) {
	zlog.Debug().Msg("DeleteTelemetryLogsGroup")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete telemetry logs group from inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteTelemetryLogsGroupResponse{}, nil
}
