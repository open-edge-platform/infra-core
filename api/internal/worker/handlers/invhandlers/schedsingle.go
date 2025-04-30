// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers

import (
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/pkg/utils"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	schedv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	schedule_cache "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache/schedule"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	invcollections "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/function"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPISingleSchedToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs SingleSchedTemplate defined in edge-infrastructure-manager-openapi-types.gen.go.
var OpenAPISingleSchedToProto = map[string]string{
	"name":           schedv1.SingleScheduleResourceFieldName,
	"targetRegionId": schedv1.SingleScheduleResourceEdgeTargetRegion,
	"targetSiteId":   schedv1.SingleScheduleResourceEdgeTargetSite,
	"targetHostId":   schedv1.SingleScheduleResourceEdgeTargetHost,
	"startSeconds":   schedv1.SingleScheduleResourceFieldStartSeconds,
	"endSeconds":     schedv1.SingleScheduleResourceFieldEndSeconds,
	"scheduleStatus": schedv1.SingleScheduleResourceFieldScheduleStatus,
}

// OpenAPISingleSchedToProtoExcluded defines exclusion rules as there are some fields
// defined in the OpenAPI spec that are not currently mapped to the proto
// fields.
var OpenAPISingleSchedToProtoExcluded = map[string]struct{}{
	"singleScheduleID": {}, // singleScheduleID must not be set from the API
	"resourceId":       {}, // resourceId must not be set from the API
	"targetHost":       {}, // targetHost must not be set from the API
	"targetSite":       {}, // targetSite must not be set from the API
	"targetRegion":     {}, // targetRegion must not be set from the API
	"timestamps":       {}, // read-only field
}

func NewSingleSchedHandler(
	invClient *clients.InventoryClientHandler,
	hScheduleCache *schedule_cache.HScheduleCacheClient,
) InventoryResource {
	return &singleSchedHandler{
		invClient:      invClient,
		hScheduleCache: hScheduleCache,
	}
}

type singleSchedHandler struct {
	invClient      *clients.InventoryClientHandler
	hScheduleCache *schedule_cache.HScheduleCacheClient
}

func (h *singleSchedHandler) Create(job *types.Job) (*types.Payload, error) {
	tenantID, exists := tenant.GetTenantIDFromContext(job.Context)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		log.InfraSec().InfraErr(err).Msg("Create single schedule is not authenticated")
		return nil, err
	}

	body, err := castSingleSched(&job.Payload)
	if err != nil {
		return nil, err
	}

	singleSched, err := openapiToGrpcSingleSched(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Singleschedule{
			Singleschedule: singleSched,
		},
	}

	invResp, err := h.invClient.InvClient.Create(job.Context, req)
	if err != nil {
		return nil, err
	}

	createdSSched := invResp.GetSingleschedule()
	h.hScheduleCache.InvalidateCache(
		tenantID, createdSSched.GetResourceId(), inventory.SubscribeEventsResponse_EVENT_KIND_CREATED)

	obj, err := grpcToOpenAPISingleSched(createdSSched, nil)
	if err != nil {
		log.InfraErr(err).Msgf("Failed to parse schedule")
		return nil, err
	}
	return &types.Payload{Data: obj}, err
}

func (h *singleSchedHandler) Get(job *types.Job) (*types.Payload, error) {
	tenantID, exists := tenant.GetTenantIDFromContext(job.Context)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		log.InfraSec().InfraErr(err).Msg("Get single schedule is not authenticated")
		return nil, err
	}

	req, err := singleSchedResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	singleSched, err := h.hScheduleCache.GetSingleSchedule(tenantID, req)
	if err != nil {
		return nil, err
	}

	obj, err := grpcToOpenAPISingleSched(singleSched, nil)
	if err != nil {
		log.InfraErr(err).Msgf("Failed to parse schedule")
		return nil, err
	}
	return &types.Payload{Data: obj}, nil
}

func (h *singleSchedHandler) Update(job *types.Job) (*types.Payload, error) {
	tenantID, exists := tenant.GetTenantIDFromContext(job.Context)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		log.InfraSec().InfraErr(err).Msg("Update single schedule is not authenticated")
		return nil, err
	}

	resID, err := singleSchedResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	fm, err := singleSchedFieldMask(&job.Payload, job.Operation)
	if err != nil {
		return nil, err
	}

	res, err := singleSchedResource(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Update(job.Context, resID, fm, res)
	if err != nil {
		return nil, err
	}

	h.hScheduleCache.InvalidateCache(tenantID, resID, inventory.SubscribeEventsResponse_EVENT_KIND_UPDATED)

	updatedSingleSched := invResp.GetSingleschedule()
	obj, err := grpcToOpenAPISingleSched(updatedSingleSched, nil)
	if err != nil {
		log.InfraErr(err).Msgf("Failed to parse schedule")
		return nil, err
	}
	obj.SingleScheduleID = &resID // to be removed
	obj.ResourceId = &resID
	return &types.Payload{Data: obj}, nil
}

func (h *singleSchedHandler) Delete(job *types.Job) error {
	tenantID, exists := tenant.GetTenantIDFromContext(job.Context)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		log.InfraSec().InfraErr(err).Msg("Delete single schedule is not authenticated")
		return err
	}

	req, err := singleSchedResourceID(&job.Payload)
	if err != nil {
		return err
	}

	_, err = h.invClient.InvClient.Delete(job.Context, req)
	if err != nil {
		return err
	}

	h.hScheduleCache.InvalidateCache(tenantID, req, inventory.SubscribeEventsResponse_EVENT_KIND_DELETED)

	return nil
}

func (h *singleSchedHandler) List(job *types.Job) (*types.Payload, error) {
	tenantID, exists := tenant.GetTenantIDFromContext(job.Context)
	if !exists {
		// This should never happen! Interceptor should either fail or set it!
		err := errors.Errorfc(codes.Unauthenticated, "Tenant ID is not present in context")
		log.InfraSec().InfraErr(err).Msg("List single schedule is not authenticated")
		return nil, err
	}

	filters, ulimit, uoffset, err := singleSchedFilter(&job.Payload)
	if err != nil {
		return nil, err
	}

	var offset, limit int
	offset, err = util.Uint32ToInt(uoffset)
	if err != nil {
		return nil, err
	}
	limit, err = util.Uint32ToInt(ulimit)
	if err != nil {
		return nil, err
	}

	singleSchedules, hasNext, totalElems, err := h.hScheduleCache.GetSingleSchedules(
		job.Context, tenantID, offset, limit, filters)
	if err != nil {
		return nil, err
	}

	schedulesAPI := make([]api.SingleSchedule, 0, len(singleSchedules))
	for _, sSched := range singleSchedules {
		obj, err := grpcToOpenAPISingleSched(sSched, &inventory.GetResourceResponse_ResourceMetadata{})
		if err != nil {
			log.InfraErr(err).Msgf("Failed to parse schedule")
			return nil, err
		}
		schedulesAPI = append(schedulesAPI, *obj)
	}

	schedsList := api.SingleSchedulesList{
		SingleSchedules: &schedulesAPI,
		HasNext:         &hasNext,
		TotalElements:   &totalElems,
	}

	payload := &types.Payload{Data: schedsList}
	return payload, nil
}

func castSingleSched(payload *types.Payload) (*api.SingleSchedule, error) {
	body, ok := payload.Data.(*api.SingleSchedule)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not SingleSchedule: %T",
			payload.Data,
		)
		log.InfraErr(err).Msgf("")
		return nil, err
	}
	return body, nil
}

func singleSchedResource(payload *types.Payload) (*inventory.Resource, error) {
	body, err := castSingleSched(payload)
	if err != nil {
		return nil, err
	}

	singleSched, err := openapiToGrpcSingleSched(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Singleschedule{
			Singleschedule: singleSched,
		},
	}
	return req, nil
}

func singleSchedResourceID(payload *types.Payload) (string, error) {
	params, ok := payload.Params.(SingleSchedURLParams)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument, "SingleSchedURLParams incorrectly formatted: %T",
			payload.Data)
		log.InfraErr(err).Msg("could not parse job payload params")
		return "", err
	}
	return params.SingleSchedID, nil
}

func singleSchedFilter(
	payload *types.Payload,
) (filters *schedule_cache.Filters, limit, offset uint32, err error) {
	if payload.Data != nil {
		query, ok := payload.Data.(api.GetSchedulesSingleParams)
		if !ok {
			errInt := errors.Errorfc(codes.InvalidArgument,
				"GetSchedulesSingleParams incorrectly formatted: %T",
				payload.Data,
			)
			log.InfraErr(errInt).Msg("list operation")
			return nil, 0, 0, errInt
		}
		filters, limit, offset, err = castSingleSchedQueryList(&query)
		if err != nil {
			log.Debug().Msgf("error parsing query parameters in list operation: %s",
				err.Error())
			return nil, 0, 0, err
		}
		filters.Add(schedule_cache.FilterByTS(query.UnixEpoch))
	}

	return filters, limit, offset, nil
}

func singleSchedFieldMask(payload *types.Payload, operation types.Operation) (*fieldmaskpb.FieldMask, error) {
	body, ok := payload.Data.(*api.SingleSchedule)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not SingleSchedule: %T",
			payload.Data,
		)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	singleSchedRes, err := singleSchedResource(payload)
	if err != nil {
		return nil, err
	}
	var fieldmask *fieldmaskpb.FieldMask
	if operation == types.Patch {
		fieldmask, err = getSingleSchedFieldmask(body)
	} else {
		fieldmask, err = fieldmaskpb.New(singleSchedRes.GetSingleschedule(),
			maps.Values(OpenAPISingleSchedToProto)...)
	}
	if err != nil {
		log.InfraErr(err).Msgf("could not create fieldmask")
		return nil, errors.Wrap(err)
	}

	return fieldmask, nil
}

func castSingleSchedQueryList(
	query *api.GetSchedulesSingleParams,
) (schedFilters *schedule_cache.Filters, limit, offset uint32, err error) {
	limit, offset, err = parsePagination(
		query.PageSize,
		query.Offset,
	)
	if err != nil {
		return nil, 0, 0, err
	}

	schedFilters = new(schedule_cache.Filters)
	nonEmptyFilters := 0
	if function.IsEmptyNullCase(query.RegionID) {
		schedFilters.Add(schedule_cache.HasNoRegion())
	} else if function.IsNotEmptyNullCase(query.RegionID) {
		schedFilters.Add(schedule_cache.HasRegionID(query.RegionID))
		nonEmptyFilters++
	}

	if function.IsEmptyNullCase(query.SiteID) {
		schedFilters.Add(schedule_cache.HasNoSite())
	} else if function.IsNotEmptyNullCase(query.SiteID) {
		schedFilters.Add(schedule_cache.HasSiteID(query.SiteID))
		nonEmptyFilters++
	}

	if function.IsEmptyNullCase(query.HostID) {
		schedFilters.Add(schedule_cache.HasNoHost())
	} else if function.IsNotEmptyNullCase(query.HostID) {
		schedFilters.Add(schedule_cache.HasHostID(query.HostID))
		nonEmptyFilters++
	}
	// Cannot apply multiple filters
	if nonEmptyFilters > 1 {
		return nil, 0, 0,
			errors.Errorfc(codes.InvalidArgument, "Only one of TargetHost, TargetSite and TargetRegion can be specified")
	}
	// No filters
	if schedFilters.Size() == 0 {
		schedFilters.Add(schedule_cache.NewStandardFilter(nil, "All"))
	}

	return schedFilters, limit, offset, nil
}

func getSingleSchedFieldmask(body *api.SingleSchedule) (*fieldmaskpb.FieldMask, error) {
	fieldList := getProtoFieldListFromOpenapiValue(*body, OpenAPISingleSchedToProto)
	log.Debug().Msgf("Proto Valid Fields: %s", fieldList)
	return fieldmaskpb.New(&schedv1.SingleScheduleResource{}, fieldList...)
}

func OpenAPISingleSchedStatusTogrpcSingleSchedStatus(s api.ScheduleStatus) schedv1.ScheduleStatus {
	mapStatus := map[api.ScheduleStatus]schedv1.ScheduleStatus{
		api.SCHEDULESTATUSMAINTENANCE: schedv1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		api.SCHEDULESTATUSOSUPDATE:    schedv1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		api.SCHEDULESTATUSUNSPECIFIED: schedv1.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED,
	}

	state, ok := mapStatus[s]
	if !ok {
		return schedv1.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED
	}
	return state
}

func openapiToGrpcSingleSchedTarget(body *api.SingleSchedule, sched *schedv1.SingleScheduleResource) error {
	requestedTargets := invcollections.Filter([]*string{body.TargetHostId, body.TargetSiteId, body.TargetRegionId}, isSet)
	if len(requestedTargets) > 1 {
		err := errors.Errorfc(
			codes.InvalidArgument,
			"only site, host or region target must be provided for schedule resource")
		log.InfraErr(err).Msg("error in parsing schedule resource")
		return err
	}

	if isSet(body.TargetRegionId) {
		sched.Relation = &schedv1.SingleScheduleResource_TargetRegion{
			TargetRegion: &locationv1.RegionResource{
				ResourceId: *body.TargetRegionId,
			},
		}
	}

	if isSet(body.TargetHostId) {
		sched.Relation = &schedv1.SingleScheduleResource_TargetHost{
			TargetHost: &computev1.HostResource{
				ResourceId: *body.TargetHostId,
			},
		}
	}

	if isSet(body.TargetSiteId) {
		sched.Relation = &schedv1.SingleScheduleResource_TargetSite{
			TargetSite: &locationv1.SiteResource{
				ResourceId: *body.TargetSiteId,
			},
		}
	}

	return nil
}

func openapiToGrpcSingleSchedSeconds(body *api.SingleSchedule, sched *schedv1.SingleScheduleResource) error {
	var err error
	if body.StartSeconds != 0 {
		sched.StartSeconds, err = utils.SafeIntToUint64(body.StartSeconds)
		if err != nil {
			log.InfraErr(err).Msgf("Failed to parse schedule start secs")
			return err
		}
	}

	if body.EndSeconds != nil {
		sched.EndSeconds, err = utils.SafeIntToUint64(*body.EndSeconds)
		if err != nil {
			log.InfraErr(err).Msgf("Failed to parse schedule end secs")
			return err
		}
	}

	if body.EndSeconds != nil && body.StartSeconds != 0 {
		if sched.EndSeconds < sched.StartSeconds {
			err := errors.Errorfc(codes.InvalidArgument,
				"The schedule end time must be greater than the start time")
			log.InfraErr(err).Msg("error in specified values of end_seconds and start_seconds")
			return err
		}
	}
	return nil
}

func openapiToGrpcSingleSched(body *api.SingleSchedule) (*schedv1.SingleScheduleResource, error) {
	status := OpenAPISingleSchedStatusTogrpcSingleSchedStatus(body.ScheduleStatus)

	sched := &schedv1.SingleScheduleResource{
		ScheduleStatus: status,
	}

	if body.Name != nil {
		sched.Name = *body.Name
	}

	err := openapiToGrpcSingleSchedSeconds(body, sched)
	if err != nil {
		return nil, err
	}

	err = openapiToGrpcSingleSchedTarget(body, sched)
	if err != nil {
		return nil, err
	}

	err = validator.ValidateMessage(sched)
	if err != nil {
		log.InfraSec().InfraErr(err).Msg("could not validate inventory resource")
		return nil, errors.Wrap(err)
	}

	return sched, nil
}

func grpcSingleSchedStatusToOpenAPISingleSchedStatus(s schedv1.ScheduleStatus) api.ScheduleStatus {
	mapStatus := map[schedv1.ScheduleStatus]api.ScheduleStatus{
		schedv1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE: api.SCHEDULESTATUSMAINTENANCE,
		schedv1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE:   api.SCHEDULESTATUSOSUPDATE,
		schedv1.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED: api.SCHEDULESTATUSUNSPECIFIED,
	}

	state, ok := mapStatus[s]
	if !ok {
		return api.SCHEDULESTATUSUNSPECIFIED
	}
	return state
}

func grpcToOpenAPISingleSched(
	sched *schedv1.SingleScheduleResource,
	_ *inventory.GetResourceResponse_ResourceMetadata,
) (*api.SingleSchedule, error) {
	resID := sched.GetResourceId()
	resName := sched.GetName()

	resStart, err := utils.SafeUint64ToInt(sched.GetStartSeconds())
	if err != nil {
		log.InfraErr(err).Msgf("Failed to parse schedule start secs")
		return nil, err
	}
	resEnd, err := utils.SafeUint64ToInt(sched.GetEndSeconds())
	if err != nil {
		log.InfraErr(err).Msgf("Failed to parse schedule end secs")
		return nil, err
	}

	status := grpcSingleSchedStatusToOpenAPISingleSchedStatus(sched.GetScheduleStatus())
	obj := api.SingleSchedule{
		SingleScheduleID: &resID,
		Name:             &resName,
		StartSeconds:     resStart,
		EndSeconds:       &resEnd,
		ScheduleStatus:   status,
		ResourceId:       &resID,
		Timestamps:       GrpcToOpenAPITimestamps(sched),
	}

	if sched.GetTargetHost() != nil {
		resHostID := sched.GetTargetHost().GetResourceId()
		obj.TargetHostId = &resHostID
		obj.TargetHost = grpcToOpenAPIHostStatus(sched.GetTargetHost(), nil, nil, nil, nil, nil, nil)
	}

	if sched.GetTargetSite() != nil {
		resSiteID := sched.GetTargetSite().GetResourceId()
		obj.TargetSiteId = &resSiteID
		obj.TargetSite = grpcToOpenAPISite(sched.GetTargetSite(), nil)
	}

	if sched.GetTargetRegion() != nil {
		id := sched.GetTargetRegion().GetResourceId()
		obj.TargetRegionId = &id
		obj.TargetRegion = grpcToOpenAPIRegion(sched.GetTargetRegion(), nil)
	}

	return &obj, nil
}
