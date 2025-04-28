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

// OpenAPIRepeatedSchedToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs SchedTemplate defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPIRepeatedSchedToProto = map[string]string{
	schedulev1.RepeatedScheduleResourceFieldName:            inv_schedulev1.RepeatedScheduleResourceFieldName,
	schedulev1.RepeatedScheduleResourceFieldTargetSiteId:    inv_schedulev1.RepeatedScheduleResourceEdgeTargetSite,
	schedulev1.RepeatedScheduleResourceFieldTargetHostId:    inv_schedulev1.RepeatedScheduleResourceEdgeTargetHost,
	schedulev1.RepeatedScheduleResourceFieldTargetRegionId:  inv_schedulev1.RepeatedScheduleResourceEdgeTargetRegion,
	schedulev1.RepeatedScheduleResourceFieldDurationSeconds: inv_schedulev1.RepeatedScheduleResourceFieldDurationSeconds,
	schedulev1.RepeatedScheduleResourceFieldCronMinutes:     inv_schedulev1.RepeatedScheduleResourceFieldCronMinutes,
	schedulev1.RepeatedScheduleResourceFieldCronHours:       inv_schedulev1.RepeatedScheduleResourceFieldCronHours,
	schedulev1.RepeatedScheduleResourceFieldCronDayMonth:    inv_schedulev1.RepeatedScheduleResourceFieldCronDayMonth,
	schedulev1.RepeatedScheduleResourceFieldCronMonth:       inv_schedulev1.RepeatedScheduleResourceFieldCronMonth,
	schedulev1.RepeatedScheduleResourceFieldCronDayWeek:     inv_schedulev1.RepeatedScheduleResourceFieldCronDayWeek,
	schedulev1.RepeatedScheduleResourceFieldScheduleStatus:  inv_schedulev1.RepeatedScheduleResourceFieldScheduleStatus,
}

var OpenAPIScheduleRepeatedObjectsNames = map[string]struct{}{
	"Relation": {},
}

func createRSRTargetRegion(targetRegionID string) *inv_schedulev1.RepeatedScheduleResource_TargetRegion {
	return &inv_schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: &inv_locationv1.RegionResource{
			ResourceId: targetRegionID,
		},
	}
}

func createRSRTargetHost(targetHostID string) *inv_schedulev1.RepeatedScheduleResource_TargetHost {
	return &inv_schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: &inv_computev1.HostResource{
			ResourceId: targetHostID,
		},
	}
}

func createRSRTargetSite(targetSiteID string) *inv_schedulev1.RepeatedScheduleResource_TargetSite {
	return &inv_schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: &inv_locationv1.SiteResource{
			ResourceId: targetSiteID,
		},
	}
}

func toInvRepeatedSchedule(
	repeatedSchedule *schedulev1.RepeatedScheduleResource,
) (*inv_schedulev1.RepeatedScheduleResource, error) {
	if repeatedSchedule == nil {
		return &inv_schedulev1.RepeatedScheduleResource{}, nil
	}
	requestedTargets := invcollections.Filter(
		[]*string{&repeatedSchedule.TargetHostId, &repeatedSchedule.TargetSiteId, &repeatedSchedule.TargetRegionId},
		isSet)
	if len(requestedTargets) > 1 {
		err := errors.Errorfc(
			codes.InvalidArgument,
			"only site, host or region target must be provided for schedule resource")
		zlog.InfraErr(err).Msg("Failed parsing schedule resource")
		return nil, err
	}

	invRepeatedSchedule := &inv_schedulev1.RepeatedScheduleResource{
		ScheduleStatus:  inv_schedulev1.ScheduleStatus(repeatedSchedule.GetScheduleStatus()),
		Name:            repeatedSchedule.GetName(),
		DurationSeconds: repeatedSchedule.GetDurationSeconds(),
		CronMinutes:     repeatedSchedule.GetCronMinutes(),
		CronHours:       repeatedSchedule.GetCronHours(),
		CronDayMonth:    repeatedSchedule.GetCronDayMonth(),
		CronMonth:       repeatedSchedule.GetCronMonth(),
		CronDayWeek:     repeatedSchedule.GetCronDayWeek(),
	}

	regionID := repeatedSchedule.GetTargetRegionId()
	hostID := repeatedSchedule.GetTargetHostId()
	siteID := repeatedSchedule.GetTargetSiteId()
	if isSet(&regionID) {
		invRepeatedSchedule.Relation = createRSRTargetRegion(regionID)
	}
	if isSet(&hostID) {
		invRepeatedSchedule.Relation = createRSRTargetHost(hostID)
	}
	if isSet(&siteID) {
		invRepeatedSchedule.Relation = createRSRTargetSite(siteID)
	}

	err := validator.ValidateMessage(invRepeatedSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invRepeatedSchedule, nil
}

func fromInvRepeatedSchedule(
	invRepeatedSchedule *inv_schedulev1.RepeatedScheduleResource,
) (*schedulev1.RepeatedScheduleResource, error) {
	if invRepeatedSchedule == nil {
		return &schedulev1.RepeatedScheduleResource{}, nil
	}
	repeatedSchedule := &schedulev1.RepeatedScheduleResource{
		ResourceId:         invRepeatedSchedule.GetResourceId(),
		RepeatedScheduleId: invRepeatedSchedule.GetResourceId(),
		ScheduleStatus:     schedulev1.ScheduleStatus(invRepeatedSchedule.GetScheduleStatus()),
		Name:               invRepeatedSchedule.GetName(),
		DurationSeconds:    invRepeatedSchedule.GetDurationSeconds(),
		CronMinutes:        invRepeatedSchedule.GetCronMinutes(),
		CronHours:          invRepeatedSchedule.GetCronHours(),
		CronDayMonth:       invRepeatedSchedule.GetCronDayMonth(),
		CronMonth:          invRepeatedSchedule.GetCronMonth(),
		CronDayWeek:        invRepeatedSchedule.GetCronDayWeek(),
		Timestamps:         GrpcToOpenAPITimestamps(invRepeatedSchedule),
	}

	switch relation := invRepeatedSchedule.GetRelation().(type) {
	case *inv_schedulev1.RepeatedScheduleResource_TargetSite:
		targetSite, err := fromInvSite(relation.TargetSite, nil)
		if err != nil {
			return nil, err
		}
		repeatedSchedule.TargetSiteId = relation.TargetSite.GetResourceId()
		repeatedSchedule.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
			TargetSite: targetSite,
		}
	case *inv_schedulev1.RepeatedScheduleResource_TargetHost:
		targetHost, err := fromInvHost(relation.TargetHost, nil)
		if err != nil {
			return nil, err
		}
		repeatedSchedule.TargetHostId = relation.TargetHost.GetResourceId()
		repeatedSchedule.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
			TargetHost: targetHost,
		}
	case *inv_schedulev1.RepeatedScheduleResource_TargetRegion:
		targetRegion, err := fromInvRegion(relation.TargetRegion, nil)
		if err != nil {
			return nil, err
		}
		repeatedSchedule.TargetRegionId = relation.TargetRegion.GetResourceId()
		repeatedSchedule.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
			TargetRegion: targetRegion,
		}
	}
	return repeatedSchedule, nil
}

func (is *InventorygRPCServer) CreateRepeatedSchedule(
	ctx context.Context,
	req *restv1.CreateRepeatedScheduleRequest,
) (*schedulev1.RepeatedScheduleResource, error) {
	zlog.Debug().Msg("CreateRepeatedSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}
	repeatedSchedule := req.GetRepeatedSchedule()
	invRepeatedSchedule, err := toInvRepeatedSchedule(repeatedSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory repeated schedule")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Repeatedschedule{
			Repeatedschedule: invRepeatedSchedule,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create repeated schedule in inventory")
		return nil, err
	}

	createdRSched := invResp.GetRepeatedschedule()
	is.InvHCacheClient.InvalidateCache(
		tenantID, createdRSched.GetResourceId(), inventory.SubscribeEventsResponse_EVENT_KIND_CREATED)

	repeatedScheduleCreated, err := fromInvRepeatedSchedule(createdRSched)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory repeated schedule")
		return nil, err
	}
	zlog.Debug().Msgf("Created %s", repeatedScheduleCreated)
	return repeatedScheduleCreated, nil
}

// Get a list of repeatedSchedules.
func (is *InventorygRPCServer) ListRepeatedSchedules(
	ctx context.Context,
	req *restv1.ListRepeatedSchedulesRequest,
) (*restv1.ListRepeatedSchedulesResponse, error) {
	zlog.Debug().Msg("ListRepeatedSchedules")
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

	invRepeatedSchedules, hasNext, totalElems, err := is.InvHCacheClient.GetRepeatedSchedules(
		ctx, tenantID, offset, limit, schedFilters)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get repeated schedules from inventory")
		return nil, err
	}

	repeatedSchedules := []*schedulev1.RepeatedScheduleResource{}
	for _, invRes := range invRepeatedSchedules {
		repeatedSchedule, errConv := fromInvRepeatedSchedule(invRes)
		if errConv != nil {
			zlog.InfraErr(errConv).Msg("Failed to convert from inventory repeated schedule")
			return nil, errConv
		}
		repeatedSchedules = append(repeatedSchedules, repeatedSchedule)
	}

	totalElements, err := SafeIntToInt32(totalElems)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert total elements to int32")
		return nil, err
	}

	resp := &restv1.ListRepeatedSchedulesResponse{
		RepeatedSchedules: repeatedSchedules,
		TotalElements:     totalElements,
		HasNext:           hasNext,
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific repeatedSchedule.
func (is *InventorygRPCServer) GetRepeatedSchedule(
	ctx context.Context,
	req *restv1.GetRepeatedScheduleRequest,
) (*schedulev1.RepeatedScheduleResource, error) {
	zlog.Debug().Msg("GetRepeatedSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	invRepeatedSchedule, err := is.InvHCacheClient.GetRepeatedSchedule(tenantID, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get repeated schedule from inventory")
		return nil, err
	}

	repeatedSchedule, err := fromInvRepeatedSchedule(invRepeatedSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory repeated schedule")
		return nil, err
	}
	zlog.Debug().Msgf("Got %s", repeatedSchedule)
	return repeatedSchedule, nil
}

// Update a repeatedSchedule. (PUT).
func (is *InventorygRPCServer) UpdateRepeatedSchedule(
	ctx context.Context,
	req *restv1.UpdateRepeatedScheduleRequest,
) (*schedulev1.RepeatedScheduleResource, error) {
	zlog.Debug().Msg("UpdateRepeatedSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}
	repeatedSchedule := req.GetRepeatedSchedule()
	invRepeatedSchedule, err := toInvRepeatedSchedule(repeatedSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory repeated schedule")
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invRepeatedSchedule, maps.Values(OpenAPIRepeatedSchedToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Repeatedschedule{
			Repeatedschedule: invRepeatedSchedule,
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
	invUp := upRes.GetRepeatedschedule()
	invUpRes, err := fromInvRepeatedSchedule(invUp)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a repeatedSchedule. (PATCH).
func (is *InventorygRPCServer) PatchRepeatedSchedule(
	ctx context.Context,
	req *restv1.PatchRepeatedScheduleRequest,
) (*schedulev1.RepeatedScheduleResource, error) {
	zlog.Debug().Msg("PatchRepeatedSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}
	repeatedSchedule := req.GetRepeatedSchedule()
	invRepeatedSchedule, err := toInvRepeatedSchedule(repeatedSchedule)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory repeated schedule")
		return nil, err
	}

	fieldmask, err := parseFielmask(invRepeatedSchedule, req.GetFieldMask(), OpenAPIRepeatedSchedToProto)
	if err != nil {
		return nil, err
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Repeatedschedule{
			Repeatedschedule: invRepeatedSchedule,
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
	invUp := upRes.GetRepeatedschedule()
	invUpRes, err := fromInvRepeatedSchedule(invUp)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a repeatedSchedule.
func (is *InventorygRPCServer) DeleteRepeatedSchedule(
	ctx context.Context,
	req *restv1.DeleteRepeatedScheduleRequest,
) (*restv1.DeleteRepeatedScheduleResponse, error) {
	zlog.Debug().Msg("DeleteRepeatedSchedule")
	tenantID, exists := tenant.GetTenantIDFromContext(ctx)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		zlog.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}
	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete repeated schedule from inventory")
		return nil, err
	}
	is.InvHCacheClient.InvalidateCache(
		tenantID,
		req.GetResourceId(),
		inventory.SubscribeEventsResponse_EVENT_KIND_DELETED,
	)
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteRepeatedScheduleResponse{}, nil
}
