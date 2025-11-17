// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIInstanceToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs Instance defined in edge-infrastructure-manager-openapi-types.gen.go.
var OpenAPIInstanceToProto = map[string]string{
	"name":           computev1.InstanceResourceFieldName,
	"kind":           computev1.InstanceResourceFieldKind,
	"osID":           computev1.InstanceResourceEdgeOs,
	"hostID":         computev1.InstanceResourceEdgeHost,
	"localAccountID": computev1.InstanceResourceEdgeLocalaccount,
}

// OpenAPIInstanceToProtoExcluded defines exclusion rules as there are some fields
// defined in the OpenAPI spec that are not currently mapped to the proto
// fields.
var OpenAPIInstanceToProtoExcluded = map[string]struct{}{
	"resourceId":                        {}, // instanceID must not be set from the API
	"instanceID":                        {}, // instanceID must not be set from the API
	"currentState":                      {}, // currentState must not be set from the API
	"desiredState":                      {}, // desiredState must not be set from the API
	"os":                                {}, // os must not be set from the API
	"desiredOs":                         {}, // os must not be set from the API
	"currentOs":                         {}, // currentOs must not be set from the API
	"host":                              {}, // host must not be set from the API
	"instanceStatus":                    {}, // read-only field
	"instanceStatusIndicator":           {}, // read-only field
	"instanceStatusTimestamp":           {}, // read-only field
	"provisioningStatus":                {}, // read-only field
	"provisioningStatusIndicator":       {}, // read-only field
	"provisioningStatusTimestamp":       {}, // read-only field
	"updateStatus":                      {}, // read-only field
	"updateStatusIndicator":             {}, // read-only field
	"updateStatusTimestamp":             {}, // read-only field
	"trustedAttestationStatus":          {}, // read-only field
	"trustedAttestationStatusIndicator": {}, // read-only field
	"trustedAttestationStatusTimestamp": {}, // read-only field
	"workloadMembers":                   {}, // workload members must not be set from the API
	"securityFeature":                   {}, // immutable field
	"timestamps":                        {}, // read-only field
	"instanceStatusDetail":              {}, // read-only field
	"localAccount":                      {}, // localaccount must not be set from the API
}

func NewInstanceHandler(invClient *clients.InventoryClientHandler) InventoryResource {
	return &instanceHandler{invClient: invClient}
}

type instanceHandler struct {
	invClient *clients.InventoryClientHandler
}

func (h *instanceHandler) Create(job *types.Job) (*types.Payload, error) {
	body, err := castInstanceAPI(&job.Payload)
	if err != nil {
		return nil, err
	}

	instance, err := openapiToGrpcInstance(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: instance,
		},
	}

	invResp, err := h.invClient.InvClient.Create(job.Context, req)
	if err != nil {
		return nil, err
	}
	createdInstance := invResp.GetInstance()

	obj := GrpcToOpenAPIInstance(createdInstance)

	return &types.Payload{Data: obj}, err
}

func (h *instanceHandler) Get(job *types.Job) (*types.Payload, error) {
	req, err := instanceResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Get(job.Context, req)
	if err != nil {
		return nil, err
	}

	instance, err := castToInstance(invResp)
	if err != nil {
		return nil, err
	}

	obj := GrpcToOpenAPIInstance(instance)

	return &types.Payload{Data: obj}, nil
}

func instanceAction(payload *types.Payload) (types.Action, error) {
	params, ok := payload.Params.(InstanceURLParams)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument, "InstanceURLParams incorrectly formatted: %T",
			payload.Data)
		log.InfraErr(err).Msg("could not parse job payload params")
		return types.ActionUnspecified, err
	}
	return params.Action, nil
}

func (h *instanceHandler) doInstanceInvalidate(job *types.Job, instanceID string) (*types.Payload, error) {
	res := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: &computev1.InstanceResource{
				DesiredState: computev1.InstanceState_INSTANCE_STATE_UNTRUSTED,
			},
		},
	}

	fm, err := fieldmaskpb.New(res.GetInstance(), computev1.InstanceResourceFieldDesiredState)
	if err != nil {
		return nil, err
	}

	_, err = h.invClient.InvClient.Update(job.Context, instanceID, fm, res)
	if err != nil {
		return nil, err
	}

	return &types.Payload{}, nil
}

func (h *instanceHandler) Update(job *types.Job) (*types.Payload, error) {
	resID, err := instanceResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	action, err := instanceAction(&job.Payload)
	if err != nil {
		return nil, err
	}

	switch action {
	case types.InstanceActionInvalidate:
		return h.doInstanceInvalidate(job, resID)
	default:
		return h.doInstanceUpdate(job)
	}
}

func (h *instanceHandler) doInstanceUpdate(job *types.Job) (*types.Payload, error) {
	if job.Operation != types.Patch {
		return nil, errors.Errorfc(codes.Unimplemented, "invalid operation %v", job.Operation)
	}

	resID, err := instanceResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	fm, err := instanceFieldMask(&job.Payload)
	if err != nil {
		return nil, err
	}

	res, err := instanceResource(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Update(job.Context, resID, fm, res)
	if err != nil {
		return nil, err
	}

	updatedInst := invResp.GetInstance()
	obj := GrpcToOpenAPIInstance(updatedInst)
	obj.InstanceID = &resID // to be removed
	obj.ResourceId = &resID
	return &types.Payload{Data: obj}, nil
}

func (h *instanceHandler) Delete(job *types.Job) error {
	req, err := instanceResourceID(&job.Payload)
	if err != nil {
		return err
	}

	_, err = h.invClient.InvClient.Delete(job.Context, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *instanceHandler) List(job *types.Job) (*types.Payload, error) {
	filter, err := instanceFilter(&job.Payload)
	if err != nil {
		return nil, err
	}

	resp, err := h.invClient.InvClient.List(job.Context, filter)
	if err != nil {
		return nil, err
	}

	instanceResource := make([]api.Instance, 0, len(resp.GetResources()))
	for _, res := range resp.GetResources() {
		instance, err := castToInstance(res)
		if err != nil {
			return nil, err
		}
		obj := GrpcToOpenAPIInstance(instance)
		instanceResource = append(instanceResource, *obj)
	}

	hasNext := resp.GetHasNext()
	totalElems := int(resp.GetTotalElements())
	instanceResourceList := api.InstanceList{
		Instances:     &instanceResource,
		HasNext:       &hasNext,
		TotalElements: &totalElems,
	}

	payload := &types.Payload{Data: instanceResourceList}
	return payload, nil
}

func castInstanceAPI(payload *types.Payload) (*api.Instance, error) {
	body, ok := payload.Data.(*api.Instance)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not Instance: %T",
			payload.Data,
		)
		log.InfraErr(err).Msgf("")
		return nil, err
	}
	return body, nil
}

func instanceResource(payload *types.Payload) (*inventory.Resource, error) {
	body, err := castInstanceAPI(payload)
	if err != nil {
		return nil, err
	}

	ins, err := openapiToGrpcInstance(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: ins,
		},
	}
	return req, nil
}

func instanceFieldMask(payload *types.Payload) (*fieldmaskpb.FieldMask, error) {
	body, ok := payload.Data.(*api.Instance)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not Instance: %T",
			payload.Data,
		)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	fieldmask, err := getInstanceFieldmask(*body)
	if err != nil {
		log.InfraErr(err).Msgf("could not create fieldmask")
		return nil, errors.Wrap(err)
	}

	return fieldmask, nil
}

func getInstanceFieldmask(body api.Instance) (*fieldmaskpb.FieldMask, error) {
	fieldList := getProtoFieldListFromOpenapiValue(body, OpenAPIInstanceToProto)
	log.Debug().Msgf("Proto Valid Fields: %s", fieldList)
	return fieldmaskpb.New(&computev1.InstanceResource{}, fieldList...)
}

func instanceFilter(payload *types.Payload) (*inventory.ResourceFilter, error) {
	req := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Instance{Instance: &computev1.InstanceResource{}}},
	}
	if payload.Data != nil {
		query, ok := payload.Data.(api.GetInstancesParams)
		if !ok {
			err := errors.Errorfc(codes.InvalidArgument,
				"GetInstancesParams incorrectly formatted: %T",
				payload.Data,
			)
			log.InfraErr(err).Msg("list operation")
			return nil, err
		}
		err := castInstanceQueryList(&query, req)
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

func instanceResourceID(payload *types.Payload) (string, error) {
	params, ok := payload.Params.(InstanceURLParams)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument, "InstanceURLParams incorrectly formatted: %T",
			payload.Data)
		log.InfraErr(err).Msg("could not parse job payload params")
		return "", err
	}
	return params.InstanceID, nil
}

func setNonQueryFilter(query *api.GetInstancesParams, req *inventory.ResourceFilter) error {
	var clauses []string
	if query.WorkloadMemberID != nil {
		//nolint:gocritic // switch/case will worsen the readability of the code
		if *query.WorkloadMemberID == emptyCase {
			clauses = append(clauses, fmt.Sprintf("has(%s)", computev1.InstanceResourceEdgeWorkloadMembers))
		} else if *query.WorkloadMemberID != emptyNullCase {
			clauses = append(clauses, fmt.Sprintf("%s.%s = %q", computev1.InstanceResourceEdgeWorkloadMembers,
				computev1.WorkloadMemberFieldResourceId, *query.WorkloadMemberID))
		} else {
			clauses = append(clauses, fmt.Sprintf("NOT has(%s)", computev1.InstanceResourceEdgeWorkloadMembers))
		}
	}
	if query.HostID != nil {
		clauses = append(clauses, fmt.Sprintf("%s.%s = %q", computev1.InstanceResourceEdgeHost,
			computev1.HostResourceFieldResourceId, *query.HostID))
	}
	if query.SiteID != nil {
		clauses = append(clauses, fmt.Sprintf("%s.%s.%s = %q", computev1.InstanceResourceEdgeHost,
			computev1.HostResourceEdgeSite, locationv1.SiteResourceFieldResourceId, *query.SiteID))
	}

	req.Filter = strings.Join(clauses, " AND ")

	return nil
}

func castInstanceQueryList(
	query *api.GetInstancesParams,
	req *inventory.ResourceFilter,
) error {
	instance := &computev1.InstanceResource{}

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
		if err := setNonQueryFilter(query, req); err != nil {
			return err
		}
	}
	req.Resource.Resource = &inventory.Resource_Instance{
		Instance: instance,
	}
	return nil
}

// helpers method to convert between API formats.
func castToInstance(resp *inventory.GetResourceResponse) (
	*computev1.InstanceResource, error,
) {
	if resp.GetResource().GetInstance() != nil {
		return resp.GetResource().GetInstance(), nil
	}
	err := errors.Errorfc(codes.InvalidArgument, "%s is not a Instance", resp.GetResource())
	log.InfraErr(err).Msgf("could not cast inventory resource")
	return nil, err
}

func instanceKindAPItoGRPC(apiKind api.InstanceKind) computev1.InstanceKind {
	kindMap := map[api.InstanceKind]computev1.InstanceKind{
		api.INSTANCEKINDMETAL: computev1.InstanceKind_INSTANCE_KIND_METAL,
		// api.INSTANCEKINDUNSPECIFIED: instancev1.InstanceKind_INSTANCE_KIND_UNSPECIFIED,
		// : instancev1.InstanceKind_INSTANCE_KIND_VM,
	}

	grpcKind, ok := kindMap[apiKind]
	if !ok {
		return computev1.InstanceKind_INSTANCE_KIND_UNSPECIFIED
	}
	return grpcKind
}

func instanceKindGRPCtoAPI(grpcKind computev1.InstanceKind) api.InstanceKind {
	kindMap := map[computev1.InstanceKind]api.InstanceKind{
		computev1.InstanceKind_INSTANCE_KIND_UNSPECIFIED: api.INSTANCEKINDUNSPECIFIED,
		computev1.InstanceKind_INSTANCE_KIND_METAL:       api.INSTANCEKINDMETAL,
		// : instancev1.InstanceKind_INSTANCE_KIND_VM,
	}

	apiKind, ok := kindMap[grpcKind]
	if !ok {
		return api.INSTANCEKINDUNSPECIFIED
	}
	return apiKind
}

func instanceStateGRPCtoAPI(grpcState computev1.InstanceState) api.InstanceState {
	stateMap := map[computev1.InstanceState]api.InstanceState{
		computev1.InstanceState_INSTANCE_STATE_DELETED:     api.INSTANCESTATEDELETED,
		computev1.InstanceState_INSTANCE_STATE_RUNNING:     api.INSTANCESTATERUNNING,
		computev1.InstanceState_INSTANCE_STATE_UNTRUSTED:   api.INSTANCESTATEUNTRUSTED,
		computev1.InstanceState_INSTANCE_STATE_UNSPECIFIED: api.INSTANCESTATEUNSPECIFIED,
	}

	apiState, ok := stateMap[grpcState]
	if !ok {
		return api.INSTANCESTATEUNSPECIFIED
	}
	return apiState
}

func openapiToGrpcInstance(body *api.Instance) (*computev1.InstanceResource, error) {
	instance := &computev1.InstanceResource{}

	if body.Name != nil {
		instance.Name = *body.Name
	}

	if body.OsID != nil {
		instance.Os = &osv1.OperatingSystemResource{
			ResourceId: *body.OsID,
		}
	}
	if body.HostID != nil {
		instance.Host = &computev1.HostResource{
			ResourceId: *body.HostID,
		}
	}
	if body.SecurityFeature != nil {
		instance.SecurityFeature = openAPISecurityFeatureTogrpcSecurityFeature(body.SecurityFeature)
	}

	if body.Kind != nil {
		instance.Kind = instanceKindAPItoGRPC(*body.Kind)
	}
	instance.DesiredState = computev1.InstanceState_INSTANCE_STATE_RUNNING // Sets default desired state.

	if body.LocalAccountID != nil {
		instance.Localaccount = &localaccountv1.LocalAccountResource{
			ResourceId: *body.LocalAccountID,
		}
	}

	err := validator.ValidateMessage(instance)
	if err != nil {
		log.InfraSec().InfraErr(err).Msg("could not validate inventory resource")
		return nil, errors.Wrap(err)
	}

	return instance, nil
}

func GrpcToOpenAPIInstance(
	instance *computev1.InstanceResource,
) *api.Instance {
	resID := instance.GetResourceId()
	resName := instance.GetName()
	resCurrState := instance.GetCurrentState()
	resDesState := instance.GetDesiredState()
	apiCurrState := instanceStateGRPCtoAPI(resCurrState)
	apiDesState := instanceStateGRPCtoAPI(resDesState)
	workloadMembers := instance.GetWorkloadMembers()

	var apiDesiredOS *api.OperatingSystemResource
	resDesiredOS := instance.GetOs()
	if resDesiredOS != nil {
		apiDesiredOS = grpcToOpenAPIOSResource(resDesiredOS, nil)
	}

	var apiCurrentOS *api.OperatingSystemResource
	resCurrentOS := instance.GetOs()
	if resCurrentOS != nil {
		apiCurrentOS = grpcToOpenAPIOSResource(resCurrentOS, nil)
	}

	var apiHost *api.Host
	resHost := instance.GetHost()
	if resHost != nil {
		apiHost = GrpcToOpenAPIHost(resHost, nil, map[string][]*networkv1.IPAddressResource{})
	}

	resKind := instance.GetKind()
	resAPIKind := instanceKindGRPCtoAPI(resKind)
	securityFeature := grpcSecurityFeatureToOpenAPISecurityFeature(instance.SecurityFeature)

	provisioningStatus := instance.GetProvisioningStatus()
	provisioningStatusTimestamp := instance.GetProvisioningStatusTimestamp()
	provisioningStatusIndicator := GrpcToOpenAPIStatusIndicator(instance.GetProvisioningStatusIndicator())

	instanceStatus := instance.GetInstanceStatus()
	instanceStatusTimestamp := instance.GetInstanceStatusTimestamp()
	instanceStatusIndicator := GrpcToOpenAPIStatusIndicator(instance.GetInstanceStatusIndicator())

	updateStatus := instance.GetUpdateStatus()
	updateStatusTimestamp := instance.GetUpdateStatusTimestamp()
	updateStatusIndicator := GrpcToOpenAPIStatusIndicator(instance.GetUpdateStatusIndicator())

	trustedAttestationStatus := instance.GetTrustedAttestationStatus()
	trustedAttestationStatusTimestamp := instance.GetTrustedAttestationStatusTimestamp()
	trustedAttestationStatusIndicator := GrpcToOpenAPIStatusIndicator(instance.GetTrustedAttestationStatusIndicator())

	localAccount := GrpcLocalAccountToOpenAPIcreatedLocalAccount(instance.GetLocalaccount())

	instanceStatusDetail := instance.GetInstanceStatusDetail()
	obj := api.Instance{
		InstanceID:      &resID,
		Name:            &resName,
		CurrentState:    &apiCurrState,
		DesiredState:    &apiDesState,
		Os:              apiDesiredOS,
		DesiredOs:       apiDesiredOS,
		CurrentOs:       apiCurrentOS,
		Host:            apiHost,
		Kind:            &resAPIKind,
		SecurityFeature: &securityFeature,

		InstanceStatusIndicator: instanceStatusIndicator,
		InstanceStatus:          &instanceStatus,
		InstanceStatusTimestamp: &instanceStatusTimestamp,

		ProvisioningStatusIndicator: provisioningStatusIndicator,
		ProvisioningStatus:          &provisioningStatus,
		ProvisioningStatusTimestamp: &provisioningStatusTimestamp,

		UpdateStatusIndicator: updateStatusIndicator,
		UpdateStatus:          &updateStatus,
		UpdateStatusTimestamp: &updateStatusTimestamp,

		TrustedAttestationStatusIndicator: trustedAttestationStatusIndicator,
		TrustedAttestationStatus:          &trustedAttestationStatus,
		TrustedAttestationStatusTimestamp: &trustedAttestationStatusTimestamp,
		LocalAccount:                      localAccount,

		ResourceId:           &resID,
		Timestamps:           GrpcToOpenAPITimestamps(instance),
		InstanceStatusDetail: &instanceStatusDetail,
	}
	if workloadMembers != nil {
		var members []api.WorkloadMember
		for _, m := range workloadMembers {
			members = append(members, *grpcToOpenAPIWorkloadMember(m))
		}
		obj.WorkloadMembers = &members
	}

	return &obj
}
