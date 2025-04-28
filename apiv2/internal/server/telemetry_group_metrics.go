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
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

func TelemetryMetricsGroupResourcetoAPI(
	telemetryGroup *inv_telemetryv1.TelemetryGroupResource,
) *telemetryv1.TelemetryMetricsGroupResource {
	if telemetryGroup == nil {
		return nil
	}
	telemetryMetricsGroup := &telemetryv1.TelemetryMetricsGroupResource{
		ResourceId:              telemetryGroup.GetResourceId(),
		TelemetryMetricsGroupId: telemetryGroup.GetResourceId(),
		Name:                    telemetryGroup.GetName(),
		CollectorKind:           telemetryv1.CollectorKind(*telemetryGroup.GetCollectorKind().Enum()),
		Groups:                  telemetryGroup.GetGroups(),
		Timestamps:              GrpcToOpenAPITimestamps(telemetryGroup),
	}
	return telemetryMetricsGroup
}

func (is *InventorygRPCServer) CreateTelemetryMetricsGroup(
	ctx context.Context,
	req *restv1.CreateTelemetryMetricsGroupRequest,
) (*telemetryv1.TelemetryMetricsGroupResource, error) {
	zlog.Debug().Msg("CreateTelemetryMetricsGroup")

	telemetryMetricsGroup := req.GetTelemetryMetricsGroup()

	telemetryGroup := &inv_telemetryv1.TelemetryGroupResource{
		Name:   telemetryMetricsGroup.GetName(),
		Groups: telemetryMetricsGroup.GetGroups(),
		Kind: inv_telemetryv1.TelemetryResourceKind(
			*telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS.Enum(),
		),
		CollectorKind: inv_telemetryv1.CollectorKind(*telemetryMetricsGroup.GetCollectorKind().Enum()),
	}

	err := validator.ValidateMessage(telemetryGroup)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_TelemetryGroup{
			TelemetryGroup: telemetryGroup,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create telemetry metrics group in inventory")
		return nil, err
	}

	telemetryMetricsGroupCreated := TelemetryMetricsGroupResourcetoAPI(invResp.GetTelemetryGroup())
	zlog.Debug().Msgf("Created %s", telemetryMetricsGroupCreated)
	return telemetryMetricsGroupCreated, nil
}

// Get a list of telemetryMetricsGroups.
func (is *InventorygRPCServer) ListTelemetryMetricsGroups(
	ctx context.Context,
	req *restv1.ListTelemetryMetricsGroupsRequest,
) (*restv1.ListTelemetryMetricsGroupsResponse, error) {
	zlog.Debug().Msg("ListTelemetryMetricsGroups")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
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
				inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS)]),
	}
	if err = validator.ValidateMessage(filter); err != nil {
		zlog.InfraErr(err).Msg("failed to validate query params")
		return nil, err
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list telemetry metrics groups from inventory")
		return nil, err
	}

	telemetryMetricsGroups := []*telemetryv1.TelemetryMetricsGroupResource{}
	for _, invRes := range invResp.GetResources() {
		telemetryGroup := invRes.GetResource().GetTelemetryGroup()
		telemetryMetricsGroup := TelemetryMetricsGroupResourcetoAPI(telemetryGroup)
		telemetryMetricsGroups = append(telemetryMetricsGroups, telemetryMetricsGroup)
	}

	resp := &restv1.ListTelemetryMetricsGroupsResponse{
		TelemetryMetricsGroups: telemetryMetricsGroups,
		TotalElements:          invResp.GetTotalElements(),
		HasNext:                invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific telemetryMetricsGroup.
func (is *InventorygRPCServer) GetTelemetryMetricsGroup(
	ctx context.Context,
	req *restv1.GetTelemetryMetricsGroupRequest,
) (*telemetryv1.TelemetryMetricsGroupResource, error) {
	zlog.Debug().Msg("GetTelemetryMetricsGroup")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get telemetry metrics group from inventory")
		return nil, err
	}

	telemetryGroup := invResp.GetResource().GetTelemetryGroup()
	telemetryMetricsGroup := TelemetryMetricsGroupResourcetoAPI(telemetryGroup)
	zlog.Debug().Msgf("Got %s", telemetryMetricsGroup)
	return telemetryMetricsGroup, nil
}

// Delete a telemetryMetricsGroup.
func (is *InventorygRPCServer) DeleteTelemetryMetricsGroup(
	ctx context.Context,
	req *restv1.DeleteTelemetryMetricsGroupRequest,
) (*restv1.DeleteTelemetryMetricsGroupResponse, error) {
	zlog.Debug().Msg("DeleteTelemetryMetricsGroup")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete telemetry metrics group from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteTelemetryMetricsGroupResponse{}, nil
}
