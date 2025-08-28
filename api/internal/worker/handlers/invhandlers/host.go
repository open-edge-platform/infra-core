// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

const (
	two = 2
)

// OpenAPIHostToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs HostTemplate and HostBmManagementInfo defined in
// edge-infrastructure-manager-openapi-types.gen.go.
// Here we should have only fields that are writable from the API.
var OpenAPIHostToProto = map[string]string{
	"desiredPowerState": computev1.HostResourceFieldDesiredPowerState,
	"name":              computev1.HostResourceFieldName,
	"siteId":            computev1.HostResourceEdgeSite,
	"metadata":          computev1.HostResourceFieldMetadata,
}

// OpenAPIHostToProtoExcluded defines exclusion rules as there are some fields
// defined in the OpenAPI spec that are not currently mapped to the proto
// fields. The key is derived from the json property respectively of the
// HostTemplate and HostBmManagementInfo structs defined in the generated types.
var OpenAPIHostToProtoExcluded = map[string]struct{}{
	"resourceId":                  {}, // read-only field
	"hostStatus":                  {}, // read-only field
	"hostStatusIndicator":         {}, // read-only field
	"hostStatusTimestamp":         {}, // read-only field
	"serialNumber":                {}, // read-only field
	"memoryBytes":                 {}, // read-only field
	"cpuModel":                    {}, // read-only field
	"cpuSockets":                  {}, // read-only field
	"cpuCores":                    {}, // read-only field
	"cpuCapabilities":             {}, // read-only field
	"cpuArchitecture":             {}, // read-only field
	"cpuThreads":                  {}, // read-only field
	"cpuTopology":                 {}, // read-only field
	"bmcKind":                     {}, // read-only field
	"bmcIp":                       {}, // read-only field
	"hostname":                    {}, // read-only field
	"productName":                 {}, // read-only field
	"biosVersion":                 {}, // read-only field
	"biosReleaseDate":             {}, // read-only field
	"biosVendor":                  {}, // read-only field
	"currentPowerState":           {}, // read-only field
	"hostID":                      {}, // read-only field
	"hostStorages":                {}, // read-only field
	"hostNics":                    {}, // read-only field
	"hostUsbs":                    {}, // read-only field
	"hostGpus":                    {}, // read-only field
	"inheritedMetadata":           {}, // read-only field
	"instance":                    {}, // read-only field
	"provider":                    {}, // read-only field
	"note":                        {}, // read-only field
	"site":                        {}, // read-only field
	"onboardingStatus":            {}, // read-only field
	"onboardingStatusIndicator":   {}, // read-only field
	"onboardingStatusTimestamp":   {}, // read-only field
	"registrationStatus":          {}, // read-only field
	"registrationStatusIndicator": {}, // read-only field
	"registrationStatusTimestamp": {}, // read-only field
	"uuid":                        {}, // immutable field
	"currentState":                {}, // read-only field
	"desiredState":                {}, // read-only field
	"timestamps":                  {}, // read-only field
}

func NewHostHandler(invClient *clients.InventoryClientHandler) InventoryResource {
	return &hostHandler{
		invClient: invClient,
	}
}

type hostHandler struct {
	invClient *clients.InventoryClientHandler
}

func (h *hostHandler) Create(job *types.Job) (*types.Payload, error) {
	var host *computev1.HostResource

	action, err := hostAction(&job.Payload)
	if err != nil {
		return nil, err
	}

	if action == types.HostActionRegister {
		body, err1 := castHostRegisterInfoAPI(&job.Payload)
		if err1 != nil {
			return nil, err1
		}

		host, err1 = openapiToGrpcHostRegister(body)
		if err1 != nil {
			return nil, err1
		}

		err1 = validateRegisterRequest(host)
		if err1 != nil {
			return nil, err1
		}
	} else {
		body, err1 := castHostAPI(&job.Payload)
		if err1 != nil {
			return nil, err1
		}

		host, err1 = openapiToGrpcHost(body)
		if err1 != nil {
			return nil, err1
		}
		// On create the host is onboarded.
		host.DesiredState = computev1.HostState_HOST_STATE_ONBOARDED
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: host,
		},
	}

	invResp, err := h.invClient.InvClient.Create(job.Context, req)
	if err != nil {
		return nil, err
	}
	createdHost := invResp.GetHost()

	// do not retrieve ipaddresses associated to the nics on create
	nicToIPAddresses, err := h.getInterfaceToIPAddresses(job.Context, createdHost, false)
	if err != nil {
		return nil, err
	}

	obj := GrpcToOpenAPIHost(createdHost, nil, nicToIPAddresses)

	return &types.Payload{Data: obj}, err
}

func (h *hostHandler) Get(job *types.Job) (*types.Payload, error) {
	req, err := hostResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Get(job.Context, req)
	if err != nil {
		return nil, err
	}

	host, meta, err := castToHost(invResp)
	if err != nil {
		return nil, err
	}

	// always retrieve ipaddresses associated to the nics on get
	nicToIPAddresses, err := h.getInterfaceToIPAddresses(job.Context, host, true)
	if err != nil {
		return nil, err
	}

	obj := GrpcToOpenAPIHost(host, meta, nicToIPAddresses)
	return &types.Payload{Data: obj}, nil
}

func (h *hostHandler) doHostUpdate(job *types.Job, hostID string) (*types.Payload, error) {
	fm, err := hostFieldMask(&job.Payload, job.Operation)
	if err != nil {
		return nil, err
	}

	res, err := hostResource(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Update(job.Context, hostID, fm, res)
	if err != nil {
		return nil, err
	}

	updatedHost := invResp.GetHost()
	// do not retrieve ipaddresses associated to the nics on update
	nicToIPAddresses, err := h.getInterfaceToIPAddresses(job.Context, updatedHost, false)
	if err != nil {
		return nil, err
	}

	obj := GrpcToOpenAPIHost(updatedHost, nil, nicToIPAddresses)
	obj.ResourceId = &hostID

	return &types.Payload{Data: obj}, nil
}

func (h *hostHandler) doHostUpdateRegister(job *types.Job, hostID string) (*types.Payload, error) {
	fm, err := hostRegisterFieldMask(&job.Payload)
	if err != nil {
		return nil, err
	}

	body, err := castHostRegisterInfoAPI(&job.Payload)
	if err != nil {
		return nil, err
	}

	host, err := openapiToGrpcHostRegister(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: host,
		},
	}

	_, err = h.invClient.InvClient.Update(job.Context, hostID, fm, req)
	if err != nil {
		return nil, err
	}

	return &types.Payload{}, nil
}

func (h *hostHandler) doHostOnboard(job *types.Job, hostID string) (*types.Payload, error) {
	res := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: &computev1.HostResource{
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
		},
	}

	fm, err := fieldmaskpb.New(res.GetHost(), computev1.HostResourceFieldDesiredState)
	if err != nil {
		return nil, err
	}

	_, err = h.invClient.InvClient.Update(job.Context, hostID, fm, res)
	if err != nil {
		return nil, err
	}

	return &types.Payload{}, nil
}

func (h *hostHandler) doHostInvalidate(job *types.Job, hostID string) (*types.Payload, error) {
	note, err := hostNote(&job.Payload)
	if err != nil {
		return nil, err
	}

	res := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: &computev1.HostResource{
				DesiredState: computev1.HostState_HOST_STATE_UNTRUSTED,
				Note:         note,
			},
		},
	}

	fm, err := fieldmaskpb.New(res.GetHost(), computev1.HostResourceFieldDesiredState, computev1.HostResourceFieldNote)
	if err != nil {
		return nil, err
	}

	_, err = h.invClient.InvClient.Update(job.Context, hostID, fm, res)
	if err != nil {
		return nil, err
	}

	return &types.Payload{}, nil
}

func (h *hostHandler) Update(job *types.Job) (*types.Payload, error) {
	resID, err := hostResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	action, err := hostAction(&job.Payload)
	if err != nil {
		return nil, err
	}

	switch action {
	case types.HostActionRegister:
		return h.doHostUpdateRegister(job, resID)
	case types.HostActionOnboard:
		return h.doHostOnboard(job, resID)
	case types.HostActionInvalidate:
		return h.doHostInvalidate(job, resID)
	default:
		return h.doHostUpdate(job, resID)
	}
}

// doHostUpdateNote updates Host object in Inventory with a note.
// It will always override the previous note content, so if empty note is provided, it clears the note field.
func (h *hostHandler) doHostUpdateNote(job *types.Job, hostID string) error {
	note, err := hostNote(&job.Payload)
	if err != nil {
		return err
	}

	res := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: &computev1.HostResource{
				Note: note,
			},
		},
	}

	fm, err := fieldmaskpb.New(res.GetHost(), computev1.HostResourceFieldNote)
	if err != nil {
		return err
	}

	_, err = h.invClient.InvClient.Update(job.Context, hostID, fm, res)
	if err != nil {
		return err
	}

	return nil
}

func (h *hostHandler) Delete(job *types.Job) error {
	hostID, err := hostResourceID(&job.Payload)
	if err != nil {
		return err
	}

	err = h.doHostUpdateNote(job, hostID)
	if err != nil {
		// do not fail, just warn
		log.Warn().Err(err).Msg("Cannot set deletion note for Host")
	}

	_, err = h.invClient.InvClient.Delete(job.Context, hostID)
	if err != nil {
		return err
	}

	return nil
}

func (h *hostHandler) List(job *types.Job) (*types.Payload, error) {
	filter, err := hostFilter(&job.Payload)
	if err != nil {
		return nil, err
	}

	detail, err := hostDetail(&job.Payload)
	if err != nil {
		return nil, err
	}

	resp, err := h.invClient.InvClient.List(job.Context, filter)
	if err != nil {
		return nil, err
	}

	hosts := make([]api.Host, 0, len(resp.GetResources())) // pre-allocate proper length
	for _, hostResp := range resp.GetResources() {
		host, meta, err := castToHost(hostResp)
		if err != nil {
			return nil, err
		}
		// retrieve ipaddresses associated to the nics depending on detail
		nicToIPAddresses, err := h.getInterfaceToIPAddresses(job.Context, host, detail)
		if err != nil {
			return nil, err
		}

		obj := GrpcToOpenAPIHost(host, meta, nicToIPAddresses)
		hosts = append(hosts, *obj)
	}

	hasNext := resp.GetHasNext()
	totalElems := int(resp.GetTotalElements())
	hostsList := api.HostsList{
		Hosts:         &hosts,
		HasNext:       &hasNext,
		TotalElements: &totalElems,
	}

	payload := &types.Payload{Data: hostsList}
	return payload, nil
}

func castHostAPI(payload *types.Payload) (*api.Host, error) {
	body, ok := payload.Data.(*api.Host)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument, "body format is not HostRequest: %T",
			payload.Data)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	return body, nil
}

func castHostRegisterInfoAPI(payload *types.Payload) (*api.HostRegisterInfo, error) {
	body, ok := payload.Data.(*api.HostRegisterInfo)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not HostRegister: %T",
			payload.Data)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	return body, nil
}

func hostFilter(payload *types.Payload) (*inventory.ResourceFilter, error) {
	req := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Host{Host: &computev1.HostResource{}}},
	}

	if payload.Data != nil {
		query, ok := payload.Data.(api.GetComputeHostsParams)
		if !ok {
			err := errors.Errorfc(codes.InvalidArgument,
				"GetComputeHostsParams incorrectly formatted: %T",
				payload.Data)
			log.InfraErr(err).Msg("list operation")
			return nil, err
		}
		err := castHostQueryList(&query, req)
		if err != nil {
			log.Debug().Msgf("error parsing query parameters in list operation: %s",
				err.Error())
			return nil, err
		}
	}

	if err := validator.ValidateMessage(req); err != nil {
		log.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	return req, nil
}

func hostDetail(payload *types.Payload) (bool, error) {
	var detail bool
	if payload.Data != nil {
		query, ok := payload.Data.(api.GetComputeHostsParams)
		if !ok {
			err := errors.Errorfc(codes.InvalidArgument,
				"GetComputeHostsParams incorrectly formatted: %T",
				payload.Data)
			log.InfraErr(err).Msg("list operation")
			return false, err
		}
		if query.Detail != nil {
			detail = *query.Detail
		}
	}

	return detail, nil
}

func hostResourceID(payload *types.Payload) (string, error) {
	params, ok := payload.Params.(HostURLParams)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"HostURLParams incorrectly formatted: %T",
			payload.Data)
		log.InfraErr(err).Msg("could not parse job payload params")
		return "", err
	}
	return params.HostID, nil
}

func hostAction(payload *types.Payload) (types.Action, error) {
	params, ok := payload.Params.(HostURLParams)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"HostURLParams incorrectly formatted: %T",
			payload.Params)
		log.InfraErr(err).Msg("could not parse job payload params")
		return types.ActionUnspecified, err
	}
	return params.Action, nil
}

func hostNote(payload *types.Payload) (string, error) {
	if payload.Data == nil {
		return "", nil
	}

	data, ok := payload.Data.(*api.HostOperationWithNote)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"HostOperationWithNote incorrectly formatted: %T",
			payload.Data)
		log.InfraErr(err).Msg("could not parse job payload params")
		return "", err
	}

	return data.Note, nil
}

func hostResource(payload *types.Payload) (*inventory.Resource, error) {
	body, err := castHostAPI(payload)
	if err != nil {
		return nil, err
	}

	host, err := openapiToGrpcHost(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: host,
		},
	}
	return req, nil
}

func hostFieldMask(payload *types.Payload, operation types.Operation) (*fieldmaskpb.FieldMask, error) {
	body, ok := payload.Data.(*api.Host)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not HostRequest: %T",
			payload.Data)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	host, err := hostResource(payload)
	if err != nil {
		return nil, err
	}

	// casting message to correct format
	castedMsg, err := castToInventoryResource(host)
	if err != nil {
		return nil, err
	}
	var fieldmask *fieldmaskpb.FieldMask
	if operation == types.Patch {
		fieldmask, err = getHostFieldmask(body)
	} else {
		fieldmask, err = fieldmaskpb.New(castedMsg.GetHost(), maps.Values(OpenAPIHostToProto)...)
	}
	if err != nil {
		log.InfraErr(err).Msgf("could not create fieldmask")
		return nil, errors.Wrap(err)
	}

	return fieldmask, nil
}

func hostRegisterFieldMask(payload *types.Payload) (*fieldmaskpb.FieldMask, error) {
	body, ok := payload.Data.(*api.HostRegisterInfo)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not HostRegisterInfo: %T",
			payload.Data)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	fieldmask, err := getHostRegisterFieldmask(body)
	if err != nil {
		log.InfraErr(err).Msgf("could not create fieldmask")
		return nil, errors.Wrap(err)
	}

	return fieldmask, nil
}

func metadataListToJSON(metadata []string) (string, error) {
	// Metadata struct representing the JSON metadata.
	type Metadata struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	metaList := make([]Metadata, len(metadata))
	for i, meta := range metadata {
		kv := strings.Split(meta, "=")
		if len(kv) != two {
			// This should be already enforced by the pattern validation
			log.InfraError("invalid metadata parameter").
				Str("metadata", meta).
				Msg("query metadata")
			return "", errors.Errorfc(codes.InvalidArgument, "invalid metadata parameter")
		}
		metaList[i] = Metadata{
			Key:   kv[0],
			Value: kv[1],
		}
	}
	metaString, err := json.Marshal(metaList)
	if err != nil {
		log.InfraSec().InfraErr(err).Msgf("Error while marshaling the metadata")
		return "", errors.Wrap(err)
	}
	return string(metaString), nil
}

//nolint:cyclop // legacy filter field handling
func castHostQueryList(
	query *api.GetComputeHostsParams,
	req *inventory.ResourceFilter,
) error {
	host := &computev1.HostResource{}
	err := error(nil)
	req.Limit, req.Offset, err = parsePagination(
		query.PageSize,
		query.Offset,
	)
	if err != nil {
		return err
	}
	if query.OrderBy != nil {
		req.OrderBy = *query.OrderBy
	}
	if query.Filter != nil {
		req.Filter = *query.Filter
	} else {
		var clauses []string
		if query.SiteID != nil {
			if *query.SiteID != emptyNullCase {
				clauses = append(clauses, fmt.Sprintf("%s.%s = %q", computev1.HostResourceEdgeSite,
					locationv1.SiteResourceFieldResourceId, *query.SiteID))
			} else {
				clauses = append(clauses, fmt.Sprintf("NOT has(%s)", computev1.HostResourceEdgeSite))
			}
		}
		if query.Uuid != nil {
			clauses = append(clauses, fmt.Sprintf("%s = %q", computev1.HostResourceFieldUuid, *query.Uuid))
		}
		if query.InstanceID != nil {
			if *query.InstanceID != emptyNullCase {
				clauses = append(clauses, fmt.Sprintf("%s.%s = %q", computev1.HostResourceEdgeInstance,
					computev1.InstanceResourceFieldResourceId, *query.InstanceID))
			} else {
				clauses = append(clauses, fmt.Sprintf("NOT has(%s)", computev1.HostResourceEdgeInstance))
			}
		}
		req.Filter = strings.Join(clauses, " AND ")
	}
	if query.Metadata != nil {
		// Marshal Metadata to JSON
		var metaJSON string
		metaJSON, err = metadataListToJSON(*query.Metadata)
		if err != nil {
			return err
		}
		host.Metadata = metaJSON
	}

	req.Resource.Resource = &inventory.Resource_Host{
		Host: host,
	}
	return nil
}

// helpers method to convert between API formats.
func castToHost(resp *inventory.GetResourceResponse) (
	*computev1.HostResource, *inventory.GetResourceResponse_ResourceMetadata, error,
) {
	if resp.GetResource().GetHost() != nil {
		return resp.GetResource().GetHost(), resp.GetRenderedMetadata(), nil
	}
	err := errors.Errorfc(codes.Internal, "%s is not a HostResource", resp.GetResource())
	log.InfraErr(err).Msgf("could not cast inventory resource")
	return nil, nil, err
}

// By default, sets host power state to ON in case it is not specified.
func grpcHostPowerStateToOpenAPIPowerState(s computev1.PowerState) api.HostPowerState {
	switch s {
	case computev1.PowerState_POWER_STATE_UNSPECIFIED:
		return api.POWERSTATEUNSPECIFIED
	case computev1.PowerState_POWER_STATE_ON:
		return api.POWERSTATEON
	case computev1.PowerState_POWER_STATE_OFF:
		return api.POWERSTATEOFF
	default:
		return api.POWERSTATEON
	}
}

func grpcHostStateToOpenAPIState(s computev1.HostState) api.HostState {
	switch s {
	case computev1.HostState_HOST_STATE_UNSPECIFIED:
		return api.HOSTSTATEUNSPECIFIED
	case computev1.HostState_HOST_STATE_DELETED:
		return api.HOSTSTATEDELETED
	case computev1.HostState_HOST_STATE_ONBOARDED:
		return api.HOSTSTATEONBOARDED
	case computev1.HostState_HOST_STATE_REGISTERED:
		return api.HOSTSTATEREGISTERED
	case computev1.HostState_HOST_STATE_UNTRUSTED:
		return api.HOSTSTATEUNTRUSTED
	default:
		return api.HOSTSTATEUNSPECIFIED
	}
}

// By default sets host power state to ON in case it is not specified.
func OpenAPIPowerStateTogrpcHostPowerState(s *api.HostPowerState) computev1.PowerState {
	if s != nil {
		switch *s {
		case api.POWERSTATEUNSPECIFIED:
			return computev1.PowerState_POWER_STATE_ON
		case api.POWERSTATEON:
			return computev1.PowerState_POWER_STATE_ON
		case api.POWERSTATEOFF:
			return computev1.PowerState_POWER_STATE_OFF
		default:
			return computev1.PowerState_POWER_STATE_ON
		}
	}
	return computev1.PowerState_POWER_STATE_ON
}

func getHostFieldmask(body *api.Host) (*fieldmaskpb.FieldMask, error) {
	var fieldList []string
	fieldList = append(
		fieldList,
		getProtoFieldListFromOpenapiPointer(body, OpenAPIHostToProto)...)
	log.Debug().Msgf("Proto Valid Fields: %s", fieldList)
	return fieldmaskpb.New(&computev1.HostResource{}, fieldList...)
}

func getHostRegisterFieldmask(body *api.HostRegisterInfo) (*fieldmaskpb.FieldMask, error) {
	var fieldList []string
	// Manually mapping the two mutable fields: Name & AutoOnboard, instead of using map table
	// AutoOnboard is translated to DesiredState field
	if body.Name != nil {
		fieldList = append(fieldList, computev1.HostResourceFieldName)
	}
	if body.AutoOnboard != nil {
		fieldList = append(fieldList, computev1.HostResourceFieldDesiredState)
	}
	return fieldmaskpb.New(&computev1.HostResource{}, fieldList...)
}

func openapiToGrpcHost(body *api.Host) (*computev1.HostResource, error) {
	var hostMAC string
	var hostSerial string
	var hostKind string
	var hostUUID string

	if body.Uuid != nil {
		hostUUID = body.Uuid.String()
	}

	metadata, metaErr := marshalMetadata(body.Metadata)
	if metaErr != nil {
		log.Debug().Msgf("marshal host metadata error: %s", metaErr.Error())
	}

	host := &computev1.HostResource{
		DesiredPowerState: OpenAPIPowerStateTogrpcHostPowerState(body.DesiredPowerState),
		Name:              body.Name,
		SerialNumber:      hostSerial,
		PxeMac:            hostMAC,
		Kind:              hostKind,
		Uuid:              hostUUID,
		Metadata:          metadata,
	}

	if !isUnset(body.SiteId) {
		siteID := *body.SiteId
		hostSite := &locationv1.SiteResource{
			ResourceId: siteID,
		}
		host.Site = hostSite
	}

	err := validator.ValidateMessage(host)
	if err != nil {
		log.InfraSec().InfraErr(err).Msg("could not validate inventory resource")
		return nil, errors.Wrap(err)
	}

	return host, nil
}

func openapiToGrpcHostRegister(body *api.HostRegisterInfo) (*computev1.HostResource, error) {
	var hostName string
	var hostKind string
	var hostSerial string
	var hostUUID string

	if body.Uuid != nil && body.Uuid.String() != "" {
		hostUUID = body.Uuid.String()
	}

	if body.SerialNumber != nil && *body.SerialNumber != "" {
		hostSerial = *body.SerialNumber
	}

	if body.Name != nil && *body.Name != "" {
		hostName = *body.Name
	}

	hostDesiredState := computev1.HostState_HOST_STATE_REGISTERED
	if body.AutoOnboard != nil && *body.AutoOnboard {
		hostDesiredState = computev1.HostState_HOST_STATE_ONBOARDED
	}

	host := &computev1.HostResource{
		DesiredState: hostDesiredState,
		Name:         hostName,
		Kind:         hostKind,
		SerialNumber: hostSerial,
		Uuid:         hostUUID,
	}

	err := validator.ValidateMessage(host)
	if err != nil {
		log.InfraSec().InfraErr(err).Msg("could not validate inventory resource")
		return nil, errors.Wrap(err)
	}

	return host, nil
}

func validateRegisterRequest(host *computev1.HostResource) error {
	if host.Uuid == "" && host.SerialNumber == "" {
		err := errors.Errorfc(codes.InvalidArgument, "Serial Number or UUID needs to be provided for host registration")
		log.InfraErr(err).Msgf("")
		return err
	}
	return nil
}

func grpcToOpenAPIHostUSB(
	hostUSBs []*computev1.HostusbResource,
) *[]api.HostResourcesUSB {
	USBs := []api.HostResourcesUSB{}

	for _, hostUsb := range hostUSBs {
		usbDeviceName := hostUsb.GetDeviceName()
		usbClass := hostUsb.GetClass()
		usbSerial := hostUsb.GetSerial()
		usbVendorID := hostUsb.GetIdvendor()
		usbProductID := hostUsb.GetIdproduct()
		usbBus := strconv.FormatUint(uint64(hostUsb.GetBus()), 10)
		usbAddress := strconv.FormatUint(uint64(hostUsb.GetAddr()), 10)

		usb := api.HostResourcesUSB{
			DeviceName: &usbDeviceName,
			Class:      &usbClass,
			Serial:     &usbSerial,
			IdVendor:   &usbVendorID,
			IdProduct:  &usbProductID,
			Bus:        &usbBus,
			Addr:       &usbAddress,
		}
		USBs = append(USBs, usb)
	}

	return &USBs
}

func grpcToOpenAPIHostGPU(
	hostGPUs []*computev1.HostgpuResource,
) *[]api.HostResourcesGPU {
	GPUs := []api.HostResourcesGPU{}

	for _, hostGpu := range hostGPUs {
		gpuDescription := hostGpu.GetDescription()
		gpuVendor := hostGpu.GetVendor()
		gpuModel := hostGpu.GetProduct()
		gpuPciID := hostGpu.GetPciId()
		gpuName := hostGpu.GetDeviceName()
		gpuCapabilities := strings.Split(hostGpu.GetFeatures(), ",")

		GPUs = append(GPUs, api.HostResourcesGPU{
			DeviceName:   &gpuName,
			Description:  &gpuDescription,
			Product:      &gpuModel,
			PciId:        &gpuPciID,
			Vendor:       &gpuVendor,
			Capabilities: &gpuCapabilities,
		})
	}

	return &GPUs
}

func grpcToOpenAPIIPConfigMode(cm networkv1.IPAddressConfigMethod) api.IPAddressConfigMethod {
	mapIPAddressConfigMode := map[networkv1.IPAddressConfigMethod]api.IPAddressConfigMethod{
		networkv1.IPAddressConfigMethod_IP_ADDRESS_CONFIG_METHOD_STATIC:  api.IPADDRESSCONFIGMODESTATIC,
		networkv1.IPAddressConfigMethod_IP_ADDRESS_CONFIG_METHOD_DYNAMIC: api.IPADDRESSCONFIGMODEDYNAMIC,
	}

	state, ok := mapIPAddressConfigMode[cm]
	if !ok {
		return api.IPADDRESSCONFIGMODEUNSPECIFIED
	}
	return state
}

func grpcToOpenAPIIPStatus(st networkv1.IPAddressStatus) api.IPAddressStatus {
	mapIPAddressStatus := map[networkv1.IPAddressStatus]api.IPAddressStatus{
		networkv1.IPAddressStatus_IP_ADDRESS_STATUS_ASSIGNMENT_ERROR:    api.IPADDRESSSTATUSASSIGNMENTERROR,
		networkv1.IPAddressStatus_IP_ADDRESS_STATUS_ASSIGNED:            api.IPADDRESSSTATUSASSIGNED,
		networkv1.IPAddressStatus_IP_ADDRESS_STATUS_CONFIGURATION_ERROR: api.IPADDRESSSTATUSCONFIGURATIONERROR,
		networkv1.IPAddressStatus_IP_ADDRESS_STATUS_CONFIGURED:          api.IPADDRESSSTATUSCONFIGURED,
		networkv1.IPAddressStatus_IP_ADDRESS_STATUS_RELEASED:            api.IPADDRESSSTATUSRELEASED,
		networkv1.IPAddressStatus_IP_ADDRESS_STATUS_ERROR:               api.IPADDRESSSTATUSERROR,
	}

	status, ok := mapIPAddressStatus[st]
	if !ok {
		return api.IPADDRESSSTATUSUNSPECIFIED
	}
	return status
}

func grpcToOpenAPILinkState(ls computev1.NetworkInterfaceLinkState) api.LinkStateType {
	mapLinkState := map[computev1.NetworkInterfaceLinkState]api.LinkStateType{
		computev1.NetworkInterfaceLinkState_NETWORK_INTERFACE_LINK_STATE_DOWN: api.LINKSTATEDOWN,
		computev1.NetworkInterfaceLinkState_NETWORK_INTERFACE_LINK_STATE_UP:   api.LINKSTATEUP,
	}

	state, ok := mapLinkState[ls]
	if !ok {
		return api.LINKSTATEUNSPECIFIED
	}
	return state
}

func grpcToOpenAPIIPAddresses(
	nicID string,
	nicToIPAddresses map[string][]*networkv1.IPAddressResource,
) *[]api.IPAddress {
	IPAddresses := []api.IPAddress{}
	invIPAddresses, ok := nicToIPAddresses[nicID]
	if ok {
		for _, invIPAddress := range invIPAddresses {
			configMode := grpcToOpenAPIIPConfigMode(invIPAddress.ConfigMethod)
			status := grpcToOpenAPIIPStatus(invIPAddress.Status)
			cidrAddress := strfmt.CIDR(invIPAddress.Address)
			statusDetail := invIPAddress.StatusDetail
			ipAddress := api.IPAddress{
				ConfigMethod: &configMode,
				Address:      &cidrAddress,
				Status:       &status,
				StatusDetail: &statusDetail,
			}
			IPAddresses = append(IPAddresses, ipAddress)
		}
	}
	return &IPAddresses
}

func grpcToOpenAPIHostInterfaces(
	hostInterfaces []*computev1.HostnicResource,
	nicToIPAddresses map[string][]*networkv1.IPAddressResource,
) *[]api.HostResourcesInterface {
	interfaces := []api.HostResourcesInterface{}

	for _, hostInterface := range hostInterfaces {
		MacAddr := hostInterface.GetMacAddr()
		DeviceName := hostInterface.GetDeviceName()
		PciIdentifier := hostInterface.GetPciIdentifier()
		SriovEnabled := hostInterface.GetSriovEnabled()
		SriovVfsNum := int(hostInterface.GetSriovVfsNum())
		SriovVfsTotal := int(hostInterface.GetSriovVfsTotal())
		IPAddresses := grpcToOpenAPIIPAddresses(hostInterface.GetResourceId(), nicToIPAddresses)
		Mtu := strconv.FormatUint(uint64(hostInterface.GetMtu()), 10)
		LinkState := grpcToOpenAPILinkState(hostInterface.GetLinkState())
		Bmc := hostInterface.GetBmcInterface()

		intf := api.HostResourcesInterface{
			MacAddr:       &MacAddr,
			DeviceName:    &DeviceName,
			PciIdentifier: &PciIdentifier,
			SriovEnabled:  &SriovEnabled,
			SriovVfsNum:   &SriovVfsNum,
			SriovVfsTotal: &SriovVfsTotal,
			Ipaddresses:   IPAddresses,
			Mtu:           &Mtu,
			LinkState: &api.LinkState{
				Type: &LinkState,
			},
			BmcInterface: &Bmc,
		}

		interfaces = append(interfaces, intf)
	}

	return &interfaces
}

func grpcToOpenAPIHostStorages(
	hostStorages []*computev1.HoststorageResource,
) *[]api.HostResourcesStorage {
	storages := []api.HostResourcesStorage{}

	for _, hostStorage := range hostStorages {
		deviceName := hostStorage.GetDeviceName()
		Capacity := strconv.FormatUint(hostStorage.GetCapacityBytes(), 10)
		Model := hostStorage.GetModel()
		Serial := hostStorage.GetSerial()
		Vendor := hostStorage.GetVendor()
		Wwid := hostStorage.GetWwid()

		storage := api.HostResourcesStorage{
			DeviceName:    &deviceName,
			CapacityBytes: &Capacity,
			Model:         &Model,
			Serial:        &Serial,
			Vendor:        &Vendor,
			Wwid:          &Wwid,
		}

		storages = append(storages, storage)
	}
	return &storages
}

func grpcToOpenAPIHostStatus(
	host *computev1.HostResource,
	hostInterfaces []*computev1.HostnicResource,
	hostStorages []*computev1.HoststorageResource,
	hostUSBs []*computev1.HostusbResource,
	hostGPUs []*computev1.HostgpuResource,
	meta *inventory.GetResourceResponse_ResourceMetadata,
	nicToIPAddresses map[string][]*networkv1.IPAddressResource,
) *api.Host {
	cores := host.GetCpuCores()
	sockets := host.GetCpuSockets()
	threads := host.GetCpuThreads()
	architecture := host.GetCpuArchitecture()
	cpuModel := host.GetCpuModel()
	capabilities := host.GetCpuCapabilities()
	cpuTopology := host.GetCpuTopology()
	storages := grpcToOpenAPIHostStorages(hostStorages)
	memoryCapacity := strconv.FormatUint(host.GetMemoryBytes(), 10)
	interfaces := grpcToOpenAPIHostInterfaces(hostInterfaces, nicToIPAddresses)
	USBs := grpcToOpenAPIHostUSB(hostUSBs)
	GPUs := grpcToOpenAPIHostGPU(hostGPUs)

	var instance *api.Instance
	if host.GetInstance() != nil {
		instance = GrpcToOpenAPIInstance(host.GetInstance())
	}

	var provider *api.Provider
	if host.GetProvider() != nil {
		provider = GrpcProviderToOpenAPIProvider(host.GetProvider())
	}

	hostCurrentState := grpcHostStateToOpenAPIState(host.GetCurrentState())
	hostDesiredState := grpcHostStateToOpenAPIState(host.GetDesiredState())
	currentPowerState := grpcHostPowerStateToOpenAPIPowerState(host.GetCurrentPowerState())
	note := host.GetNote()
	status := &api.Host{
		CurrentState:      &hostCurrentState,
		DesiredState:      &hostDesiredState,
		CurrentPowerState: &currentPowerState,
		Note:              &note,
		BiosReleaseDate:   &host.BiosReleaseDate,
		BiosVendor:        &host.BiosVendor,
		BiosVersion:       &host.BiosVersion,
		Hostname:          &host.Hostname,
		ProductName:       &host.ProductName,
		SerialNumber:      &host.SerialNumber,
		CpuCores:          &cores,
		CpuModel:          &cpuModel,
		CpuSockets:        &sockets,
		CpuThreads:        &threads,
		CpuArchitecture:   &architecture,
		CpuCapabilities:   &capabilities,
		CpuTopology:       &cpuTopology,
		MemoryBytes:       &memoryCapacity,
		HostNics:          interfaces,
		HostStorages:      storages,
		HostUsbs:          USBs,
		HostGpus:          GPUs,
		Instance:          instance,
		Provider:          provider,
	}

	if meta != nil {
		var err error
		status.InheritedMetadata = &api.MetadataJoin{}
		status.InheritedMetadata.Location, err = unmarshalMetadata(meta.GetPhyMetadata())
		if err != nil {
			log.Debug().Msgf("unmarshal rendered location metadata error")
		}

		status.InheritedMetadata.Ou, err = unmarshalMetadata(meta.GetLogiMetadata())
		if err != nil {
			log.Debug().Msgf("unmarshal rendered OU metadata error")
		}
	}

	return status
}

func GrpcToOpenAPIHost(
	host *computev1.HostResource,
	meta *inventory.GetResourceResponse_ResourceMetadata,
	nicToIPAddresses map[string][]*networkv1.IPAddressResource,
) *api.Host {
	hostID := host.GetResourceId()

	var UUID *uuid.UUID
	UUIDTmp, err := uuid.Parse(host.GetUuid())
	if err == nil {
		UUID = &UUIDTmp
	}

	var siteID *string
	hostSite := host.GetSite()
	if hostSite != nil {
		siteID = getPtr(hostSite.GetResourceId())
	}

	hostDesiredPowerState := grpcHostPowerStateToOpenAPIPowerState(host.GetDesiredPowerState())
	hostCurrentPowerState := grpcHostPowerStateToOpenAPIPowerState(host.GetCurrentPowerState())

	hostInterfaces := host.GetHostNics()
	hostStorages := host.GetHostStorages()
	hostUSBs := host.GetHostUsbs()
	hostGPUs := host.GetHostGpus()

	hostObj := grpcToOpenAPIHostStatus(
		host,
		hostInterfaces,
		hostStorages,
		hostUSBs,
		hostGPUs,
		meta,
		nicToIPAddresses,
	)

	metadata, metaErr := unmarshalMetadata(host.GetMetadata())
	if metaErr != nil {
		log.Debug().Msgf("unmarshal host metadata error: %s", metaErr.Error())
	}

	if hostSite != nil {
		hostObj.Site = grpcToOpenAPISite(hostSite, nil)
	}

	hostObj.ResourceId = &hostID
	hostObj.Name = host.GetName()
	hostObj.CurrentPowerState = &hostCurrentPowerState
	hostObj.DesiredPowerState = &hostDesiredPowerState
	hostObj.SiteId = siteID
	hostObj.Uuid = UUID
	hostObj.Metadata = metadata

	onboardingStatus := host.GetOnboardingStatus()
	onboardingStatusTimestamp := host.GetOnboardingStatusTimestamp()
	hostObj.OnboardingStatusIndicator = GrpcToOpenAPIStatusIndicator(host.GetOnboardingStatusIndicator())
	hostObj.OnboardingStatus = &onboardingStatus
	hostObj.OnboardingStatusTimestamp = &onboardingStatusTimestamp

	registrationStatus := host.GetRegistrationStatus()
	registrationStatusTimestamp := host.GetRegistrationStatusTimestamp()
	hostObj.RegistrationStatusIndicator = GrpcToOpenAPIStatusIndicator(host.GetRegistrationStatusIndicator())
	hostObj.RegistrationStatus = &registrationStatus
	hostObj.RegistrationStatusTimestamp = &registrationStatusTimestamp

	hostStatus := host.GetHostStatus()
	hostStatusTimestamp := host.GetHostStatusTimestamp()
	hostObj.HostStatusIndicator = GrpcToOpenAPIStatusIndicator(host.GetHostStatusIndicator())
	hostObj.HostStatus = &hostStatus
	hostObj.HostStatusTimestamp = &hostStatusTimestamp

	hostObj.Timestamps = GrpcToOpenAPITimestamps(host)

	return hostObj
}

func (h *hostHandler) getInterfaceToIPAddresses(
	ctx context.Context,
	host *computev1.HostResource,
	detail bool,
) (map[string][]*networkv1.IPAddressResource, error) {
	nicToIPAddresses := make(map[string][]*networkv1.IPAddressResource)
	if detail {
		hostInterfaces := host.GetHostNics()
		for _, hostInterface := range hostInterfaces {
			ipAddresses := make([]*networkv1.IPAddressResource, 0)
			req := &inventory.ResourceFilter{
				Resource: &inventory.Resource{Resource: &inventory.Resource_Ipaddress{}},
				Filter: fmt.Sprintf("%s.%s = %q", networkv1.IPAddressResourceEdgeNic,
					computev1.HostnicResourceFieldResourceId, hostInterface.GetResourceId()),
			}
			inventoryRes, err := h.invClient.InvClient.List(ctx, req)
			if errors.IsNotFound(err) {
				// resp is nil but we can continue in this case
				nicToIPAddresses[hostInterface.GetResourceId()] = ipAddresses
				continue
			}
			if err != nil {
				return nil, err
			}
			for _, ipResp := range inventoryRes.Resources {
				ipAddress, err := castToIPAddress(ipResp)
				if err != nil {
					return nil, err
				}
				ipAddresses = append(ipAddresses, ipAddress)
			}
			nicToIPAddresses[hostInterface.GetResourceId()] = ipAddresses
		}
	}
	return nicToIPAddresses, nil
}

func castToIPAddress(resp *inventory.GetResourceResponse) (*networkv1.IPAddressResource, error) {
	if resp.GetResource().GetIpaddress() != nil {
		return resp.GetResource().GetIpaddress(), nil
	}
	err := errors.Errorfc(codes.Internal, "%s is not an IPAddress", resp.GetResource())
	log.InfraErr(err).Msgf("could not cast inventory resource")
	return nil, err
}
