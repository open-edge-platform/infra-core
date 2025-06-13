// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	statusv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/status/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

func fromInvOSUpdateRunResource(invOSUpdateRunResource *inv_computev1.OSUpdateRunResource) *computev1.OSUpdateRun {
	parseTimestamp := func(ts string) *timestamppb.Timestamp {
		parsedTime, err := time.Parse(ISO8601TimeFormat, ts)
		if err != nil {
			zlog.Warn().Err(err).Msgf("Failed to parse timestamp: %s", ts)
			return nil
		}
		return timestamppb.New(parsedTime)
	}

	if invOSUpdateRunResource == nil {
		return &computev1.OSUpdateRun{}
	}
	instance, err := fromInvInstance(invOSUpdateRunResource.GetInstance())
	if err != nil {
		zlog.Warn().Err(err).Msgf("Failed to get the inventory instance edge from OS Update Run resource")
		return nil
	}

	osUpdateRunResource := &computev1.OSUpdateRun{
		ResourceId:      invOSUpdateRunResource.GetResourceId(),
		Name:            invOSUpdateRunResource.GetName(),
		Description:     invOSUpdateRunResource.GetDescription(),
		AppliedPolicy:   fromInvOSUpdatePolicy(invOSUpdateRunResource.GetAppliedPolicy()),
		Instance:        instance,
		StatusIndicator: statusv1.StatusIndication(invOSUpdateRunResource.GetStatusIndicator()),
		Status:          invOSUpdateRunResource.GetStatus(),
		StatusDetails:   invOSUpdateRunResource.GetStatusDetails(),
		StatusTimestamp: parseTimestamp(invOSUpdateRunResource.GetStatusTimestamp()),
		StartTime:       parseTimestamp(invOSUpdateRunResource.GetStartTime()),
		EndTime:         parseTimestamp(invOSUpdateRunResource.GetEndTime()),
		Timestamps:      GrpcToOpenAPITimestamps(invOSUpdateRunResource),
	}
	return osUpdateRunResource
}

func (is *InventorygRPCServer) ListOSUpdateRun(ctx context.Context, req *restv1.ListOSUpdateRunRequest) (
	*restv1.ListOSUpdateRunResponse, error,
) {
	zlog.Debug().Msg("ListOSUpdateRunResources")

	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_OsUpdateRun{
				OsUpdateRun: &inv_computev1.OSUpdateRunResource{},
			},
		},
		Offset:  req.GetOffset(),
		Limit:   req.GetPageSize(),
		OrderBy: req.GetOrderBy(),
		Filter:  req.GetFilter(),
	}
	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list OS resources from inventory")
		return nil, errors.Wrap(err)
	}

	osUpdateRunResources := []*computev1.OSUpdateRun{}
	for _, invRes := range invResp.GetResources() {
		osUpdateRunResource := fromInvOSUpdateRunResource(invRes.GetResource().GetOsUpdateRun())
		osUpdateRunResources = append(osUpdateRunResources, osUpdateRunResource)
	}

	resp := &restv1.ListOSUpdateRunResponse{
		OsUpdateRuns:  osUpdateRunResources,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

func (is *InventorygRPCServer) GetOSUpdateRun(ctx context.Context, req *restv1.GetOSUpdateRunRequest) (
	*computev1.OSUpdateRun, error,
) {
	zlog.Debug().Msg("GetOSUpdateRunResource")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get OS Update Run resource from inventory")
		return nil, errors.Wrap(err)
	}

	invOSUpdateRunResource := invResp.GetResource().GetOsUpdateRun()
	osUpdateRunResource := fromInvOSUpdateRunResource(invOSUpdateRunResource)

	zlog.Debug().Msgf("Got %s", osUpdateRunResource)
	return osUpdateRunResource, nil
}

func (is *InventorygRPCServer) DeleteOSUpdateRun(ctx context.Context, req *restv1.DeleteOSUpdateRunRequest) (
	*restv1.DeleteOSUpdateRunResponse, error,
) {
	zlog.Debug().Msg("DeleteOSUpdateRunResource")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete OS Update Run resource from inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteOSUpdateRunResponse{}, nil
}
