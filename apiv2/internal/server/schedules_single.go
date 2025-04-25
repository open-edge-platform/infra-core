// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	schedulev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/schedule/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	invcollections "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPISingleSchedToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs SingleSchedTemplate defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPISingleSchedToProto = map[string]string{
	schedulev1.SingleScheduleResourceFieldName:           inv_schedulev1.SingleScheduleResourceFieldName,
	schedulev1.SingleScheduleResourceFieldTargetRegionId: inv_schedulev1.SingleScheduleResourceEdgeTargetRegion,
	schedulev1.SingleScheduleResourceFieldTargetSiteId:   inv_schedulev1.SingleScheduleResourceEdgeTargetSite,
	schedulev1.SingleScheduleResourceFieldTargetHostId:   inv_schedulev1.SingleScheduleResourceEdgeTargetHost,
	schedulev1.SingleScheduleResourceFieldStartSeconds:   inv_schedulev1.SingleScheduleResourceFieldStartSeconds,
	schedulev1.SingleScheduleResourceFieldEndSeconds:     inv_schedulev1.SingleScheduleResourceFieldEndSeconds,
	schedulev1.SingleScheduleResourceFieldScheduleStatus: inv_schedulev1.SingleScheduleResourceFieldScheduleStatus,
}

func toInvSingleschedule(singleSchedule *schedulev1.SingleScheduleResource) (*inv_schedulev1.SingleScheduleResource, error) {
	if singleSchedule == nil {
		return &inv_schedulev1.SingleScheduleResource{}, nil
	}

	requestedTargets := invcollections.Filter(
		[]*string{&singleSchedule.TargetHostId, &singleSchedule.TargetSiteId, &singleSchedule.TargetRegionId},
		isSet)
	if len(requestedTargets) > 1 {
		err := errors.Errorfc(
			codes.InvalidArgument,
			"only site, host or region target must be provided for schedule resource")
		zlog.InfraErr(err).Msg("Failed parsing schedule resource")
		return nil, err
	}

	invSingleschedule := &inv_schedulev1.SingleScheduleResource{
		ScheduleStatus: inv_schedulev1.ScheduleStatus(singleSchedule.GetScheduleStatus()),
		Name:           singleSchedule.GetName(),
		StartSeconds:   uint64(singleSchedule.GetStartSeconds()),
		EndSeconds:     uint64(singleSchedule.GetEndSeconds()),
	}

	if isSet(&singleSchedule.TargetRegionId) {
		invSingleschedule.Relation = &inv_schedulev1.SingleScheduleResource_TargetRegion{
			TargetRegion: &inv_locationv1.RegionResource{
				ResourceId: singleSchedule.TargetRegionId,
			},
		}
	}

	if isSet(&singleSchedule.TargetHostId) {
		invSingleschedule.Relation = &inv_schedulev1.SingleScheduleResource_TargetHost{
			TargetHost: &inv_computev1.HostResource{
				ResourceId: singleSchedule.TargetHostId,
			},
		}
	}

	if isSet(&singleSchedule.TargetSiteId) {
		invSingleschedule.Relation = &inv_schedulev1.SingleScheduleResource_TargetSite{
			TargetSite: &inv_locationv1.SiteResource{
				ResourceId: singleSchedule.TargetSiteId,
			},
		}
	}

	err := validator.ValidateMessage(invSingleschedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invSingleschedule, nil
}

func fromInvSingleschedule(
	invSingleschedule *inv_schedulev1.SingleScheduleResource,
) (*schedulev1.SingleScheduleResource, error) {
	if invSingleschedule == nil {
		return &schedulev1.SingleScheduleResource{}, nil
	}
	startSec, err := SafeUint64ToUint32(invSingleschedule.GetStartSeconds())
	if err != nil {
		return nil, err
	}
	endSec, err := SafeUint64ToUint32(invSingleschedule.GetEndSeconds())
	if err != nil {
		return nil, err
	}

	singleSchedule := &schedulev1.SingleScheduleResource{
		ResourceId:       invSingleschedule.GetResourceId(),
		SingleScheduleId: invSingleschedule.GetResourceId(),
		ScheduleStatus:   schedulev1.ScheduleStatus(invSingleschedule.GetScheduleStatus()),
		Name:             invSingleschedule.GetName(),
		StartSeconds:     startSec,
		EndSeconds:       endSec,
		CreatedAt:        invSingleschedule.GetCreatedAt(),
		UpdatedAt:        invSingleschedule.GetUpdatedAt(),
	}

	switch relation := invSingleschedule.GetRelation().(type) {
	case *inv_schedulev1.SingleScheduleResource_TargetSite:
		targetSite, err := fromInvSite(relation.TargetSite, nil)
		if err != nil {
			return nil, err
		}
		singleSchedule.TargetSiteId = relation.TargetSite.GetResourceId()
		singleSchedule.Relation = &schedulev1.SingleScheduleResource_TargetSite{
			TargetSite: targetSite,
		}
	case *inv_schedulev1.SingleScheduleResource_TargetHost:
		targetHost, err := fromInvHost(relation.TargetHost, nil)
		if err != nil {
			return nil, err
		}
		singleSchedule.TargetHostId = relation.TargetHost.GetResourceId()
		singleSchedule.Relation = &schedulev1.SingleScheduleResource_TargetHost{
			TargetHost: targetHost,
		}
	case *inv_schedulev1.SingleScheduleResource_TargetRegion:
		targetRegion, err := fromInvRegion(relation.TargetRegion, nil)
		if err != nil {
			return nil, err
		}
		singleSchedule.TargetRegionId = relation.TargetRegion.GetResourceId()
		singleSchedule.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
			TargetRegion: targetRegion,
		}
	}
	return singleSchedule, nil
}

func (is *InventorygRPCServer) CreateSingleSchedule(
	ctx context.Context,
	req *restv1.CreateSingleScheduleRequest,
) (*schedulev1.SingleScheduleResource, error) {
	zlog.Debug().Msg("CreateSingleschedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	singleSchedule := req.GetSingleSchedule()
	invSingleschedule, err := toInvSingleschedule(singleSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory single schedule")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Singleschedule{
			Singleschedule: invSingleschedule,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create single schedule in inventory")
		return nil, err
	}
	createdSSched := invResp.GetSingleschedule()
	is.InvHCacheClient.InvalidateCache(
		tenantID, createdSSched.GetResourceId(), inventory.SubscribeEventsResponse_EVENT_KIND_CREATED)

	invSinglescheduleCreated, err := fromInvSingleschedule(createdSSched)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory single schedule")
		return nil, err
	}
	zlog.Debug().Msgf("Created %s", invSinglescheduleCreated)
	return invSinglescheduleCreated, nil
}

// Get a list of singleSchedules.
func (is *InventorygRPCServer) ListSingleSchedules(
	ctx context.Context,
	req *restv1.ListSingleSchedulesRequest,
) (*restv1.ListSingleSchedulesResponse, error) {
	zlog.Debug().Msg("ListSingleSchedules")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	hostID, siteID, regionID, epoch := req.GetHostId(), req.GetSiteId(), req.GetRegionId(), req.GetUnixEpoch()
	schedFilters, err := parseSchedulesFilter(&hostID, &siteID, &regionID, &epoch)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to parse schedules filter")
		return nil, err
	}
	var offset, limit int
	offset, err = util.Int32ToInt(req.GetOffset())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert offset")
		return nil, err
	}
	limit, err = util.Int32ToInt(req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert page size")
		return nil, err
	}
	repeatedSchedules, hasNext, totalElems, err := is.InvHCacheClient.GetSingleSchedules(
		ctx, tenantID, offset, limit, schedFilters)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get single schedules from inventory")
		return nil, err
	}

	singleSchedules := []*schedulev1.SingleScheduleResource{}
	for _, invRes := range repeatedSchedules {
		singleSchedule, errConv := fromInvSingleschedule(invRes)
		if errConv != nil {
			zlog.InfraErr(errConv).Msg("Failed to convert from inventory single schedule")
			return nil, errConv
		}
		singleSchedules = append(singleSchedules, singleSchedule)
	}
	totalElements, err := SafeIntToInt32(totalElems)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert total elements to int32")
		return nil, err
	}

	resp := &restv1.ListSingleSchedulesResponse{
		SingleSchedules: singleSchedules,
		TotalElements:   totalElements,
		HasNext:         hasNext,
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific singleSchedule.
func (is *InventorygRPCServer) GetSingleSchedule(
	ctx context.Context,
	req *restv1.GetSingleScheduleRequest,
) (*schedulev1.SingleScheduleResource, error) {
	zlog.Debug().Msg("GetSingleSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	invSingleschedule, err := is.InvHCacheClient.GetSingleSchedule(tenantID, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get single schedule from inventory")
		return nil, err
	}

	singleSchedule, err := fromInvSingleschedule(invSingleschedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory single schedule")
		return nil, err
	}
	zlog.Debug().Msgf("Got %s", singleSchedule)
	return singleSchedule, nil
}

// Update a singleSchedule. (PUT).
func (is *InventorygRPCServer) UpdateSingleSchedule(
	ctx context.Context,
	req *restv1.UpdateSingleScheduleRequest,
) (*schedulev1.SingleScheduleResource, error) {
	zlog.Debug().Msg("UpdateSingleschedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	singleSchedule := req.GetSingleSchedule()
	invSingleschedule, err := toInvSingleschedule(singleSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory single schedule")
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invSingleschedule, maps.Values(OpenAPISingleSchedToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Singleschedule{
			Singleschedule: invSingleschedule,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	is.InvHCacheClient.InvalidateCache(
		tenantID,
		req.GetResourceId(),
		inventory.SubscribeEventsResponse_EVENT_KIND_UPDATED,
	)

	invUp := upRes.GetSingleschedule()
	invUpRes, err := fromInvSingleschedule(invUp)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a singleSchedule. (PATCH).
func (is *InventorygRPCServer) PatchSingleSchedule(
	ctx context.Context,
	req *restv1.PatchSingleScheduleRequest,
) (*schedulev1.SingleScheduleResource, error) {
	zlog.Debug().Msg("PatchSingleSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	singleSchedule := req.GetSingleSchedule()
	invSingleschedule, err := toInvSingleschedule(singleSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory single schedule")
		return nil, err
	}

	fieldmask, err := parseFielmask(invSingleschedule, req.GetFieldMask(), OpenAPISingleSchedToProto)
	if err != nil {
		return nil, err
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Singleschedule{
			Singleschedule: invSingleschedule,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	is.InvHCacheClient.InvalidateCache(
		tenantID,
		req.GetResourceId(),
		inventory.SubscribeEventsResponse_EVENT_KIND_UPDATED,
	)

	invUp := upRes.GetSingleschedule()
	invUpRes, err := fromInvSingleschedule(invUp)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a singleSchedule.
func (is *InventorygRPCServer) DeleteSingleSchedule(
	ctx context.Context,
	req *restv1.DeleteSingleScheduleRequest,
) (*restv1.DeleteSingleScheduleResponse, error) {
	zlog.Debug().Msg("DeleteSingleschedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}
	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete single schedule from inventory")
		return nil, err
	}
	is.InvHCacheClient.InvalidateCache(
		tenantID,
		req.GetResourceId(),
		inventory.SubscribeEventsResponse_EVENT_KIND_DELETED,
	)

	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteSingleScheduleResponse{}, nil
}
