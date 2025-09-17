// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	customconfigv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/customconfig/v1"
	localaccountv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/localaccount/v1"
	osv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/os/v1"
	statusv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/status/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	inv_osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIInstanceToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs Instance defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPIInstanceToProto = map[string]string{
	computev1.InstanceResourceFieldName:             inv_computev1.InstanceResourceFieldName,
	computev1.InstanceResourceFieldKind:             inv_computev1.InstanceResourceFieldKind,
	computev1.InstanceResourceFieldOsID:             inv_computev1.InstanceResourceEdgeDesiredOs,
	computev1.InstanceResourceFieldHostID:           inv_computev1.InstanceResourceEdgeHost,
	computev1.InstanceResourceFieldOsUpdatePolicyID: inv_computev1.InstanceResourceEdgeOsUpdatePolicy,
}

func toInvInstance(instance *computev1.InstanceResource) (*inv_computev1.InstanceResource, error) {
	if instance == nil {
		return &inv_computev1.InstanceResource{}, nil
	}

	invInstance := &inv_computev1.InstanceResource{
		Kind:            inv_computev1.InstanceKind(instance.GetKind()),
		Name:            instance.GetName(),
		DesiredState:    inv_computev1.InstanceState_INSTANCE_STATE_RUNNING,
		SecurityFeature: inv_osv1.SecurityFeature(instance.GetSecurityFeature()),
		ExistingCves:    instance.GetExistingCves(),
	}

	hostID := instance.GetHostID()
	if isSet(&hostID) {
		invInstance.Host = &inv_computev1.HostResource{
			ResourceId: hostID,
		}
	}

	osID := instance.GetOsID()
	if isSet(&osID) {
		invInstance.DesiredOs = &inv_osv1.OperatingSystemResource{
			ResourceId: osID,
		}
		invInstance.Os = &inv_osv1.OperatingSystemResource{
			ResourceId: osID,
		}
	}

	laID := instance.GetLocalAccountID()
	if isSet(&laID) {
		invInstance.Localaccount = &inv_localaccountv1.LocalAccountResource{
			ResourceId: laID,
		}
	}

	ccIDs := instance.GetCustomConfigID()
	for _, ccID := range ccIDs {
		invInstance.CustomConfig = append(invInstance.CustomConfig,
			&inv_computev1.CustomConfigResource{ResourceId: ccID})
	}

	oupID := instance.GetOsUpdatePolicyID()
	if isSet(&oupID) {
		invInstance.OsUpdatePolicy = &inv_computev1.OSUpdatePolicyResource{
			ResourceId: oupID,
		}
	}

	err := validator.ValidateMessage(invInstance)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	return invInstance, nil
}

func fromInvInstanceStatus(
	invInstance *inv_computev1.InstanceResource,
	instance *computev1.InstanceResource,
) {
	instanceStatus := invInstance.GetInstanceStatus()
	instanceStatusDetail := invInstance.GetInstanceStatusDetail()
	instanceStatusIndicator := statusv1.StatusIndication(invInstance.GetInstanceStatusIndicator())
	instanceStatusTimestamp := TruncateUint64ToUint32(invInstance.GetInstanceStatusTimestamp())

	provisioningStatus := invInstance.GetProvisioningStatus()
	provisioningStatusIndicator := statusv1.StatusIndication(invInstance.GetProvisioningStatusIndicator())
	provisioningStatusTimestamp := TruncateUint64ToUint32(invInstance.GetProvisioningStatusTimestamp())

	updateStatus := invInstance.GetUpdateStatus()
	updateStatusIndicator := statusv1.StatusIndication(invInstance.GetUpdateStatusIndicator())
	updateStatusTimestamp := TruncateUint64ToUint32(invInstance.GetUpdateStatusTimestamp())

	attestationStatus := invInstance.GetTrustedAttestationStatus()
	attestationStatusIndicator := statusv1.StatusIndication(invInstance.GetTrustedAttestationStatusIndicator())
	attestationStatusTimestamp := TruncateUint64ToUint32(invInstance.GetTrustedAttestationStatusTimestamp())

	instance.InstanceStatus = instanceStatus
	instance.InstanceStatusDetail = instanceStatusDetail
	instance.InstanceStatusIndicator = instanceStatusIndicator
	instance.InstanceStatusTimestamp = instanceStatusTimestamp
	instance.ProvisioningStatus = provisioningStatus
	instance.ProvisioningStatusIndicator = provisioningStatusIndicator
	instance.ProvisioningStatusTimestamp = provisioningStatusTimestamp
	instance.TrustedAttestationStatus = attestationStatus
	instance.TrustedAttestationStatusIndicator = attestationStatusIndicator
	instance.TrustedAttestationStatusTimestamp = attestationStatusTimestamp
	instance.UpdateStatus = updateStatus
	instance.UpdateStatusIndicator = updateStatusIndicator
	instance.UpdateStatusTimestamp = updateStatusTimestamp
}

//nolint:cyclop // it is a conversion function
func fromInvInstance(invInstance *inv_computev1.InstanceResource) (*computev1.InstanceResource, error) {
	if invInstance == nil {
		return &computev1.InstanceResource{}, nil
	}

	var err error
	var desiredOs *osv1.OperatingSystemResource
	var currentOs *osv1.OperatingSystemResource
	var os *osv1.OperatingSystemResource
	var host *computev1.HostResource
	var la *localaccountv1.LocalAccountResource
	var oup *computev1.OSUpdatePolicy
	if invInstance.GetDesiredOs() != nil {
		desiredOs = fromInvOSResource(invInstance.GetDesiredOs())
	}
	if invInstance.GetCurrentOs() != nil {
		currentOs = fromInvOSResource(invInstance.GetCurrentOs())
	}
	if invInstance.GetOs() != nil {
		os = fromInvOSResource(invInstance.GetOs())
	}

	if invInstance.GetHost() != nil {
		host, err = fromInvHost(invInstance.GetHost(), nil, nil)
		if err != nil {
			return nil, err
		}
	}
	if invInstance.GetLocalaccount() != nil {
		la = fromInvLocalAccount(invInstance.GetLocalaccount())
	}
	if invInstance.GetOsUpdatePolicy() != nil {
		oup = fromInvOSUpdatePolicy(invInstance.GetOsUpdatePolicy())
	}

	workloadMembers := []*computev1.WorkloadMember{}
	for _, instWM := range invInstance.GetWorkloadMembers() {
		workloadMember, errWM := fromInvWorkloadMember(instWM)
		if errWM != nil {
			return nil, errWM
		}
		workloadMembers = append(workloadMembers, workloadMember)
	}

	customConfigs := []*customconfigv1.CustomConfigResource{}
	customConfigIDs := []string{}
	for _, cc := range invInstance.GetCustomConfig() {
		customConfigs = append(customConfigs, fromInvCustomConfig(cc))
		customConfigIDs = append(customConfigIDs, cc.GetResourceId())
	}

	instance := &computev1.InstanceResource{
		ResourceId:         invInstance.GetResourceId(),
		InstanceID:         invInstance.GetResourceId(),
		Kind:               computev1.InstanceKind(invInstance.GetKind()),
		Name:               invInstance.GetName(),
		DesiredState:       computev1.InstanceState(invInstance.GetDesiredState()),
		CurrentState:       computev1.InstanceState(invInstance.GetCurrentState()),
		Host:               host,
		HostID:             host.GetResourceId(),
		Os:                 os,
		DesiredOs:          desiredOs,
		CurrentOs:          currentOs,
		OsID:               currentOs.GetResourceId(),
		SecurityFeature:    osv1.SecurityFeature(invInstance.GetSecurityFeature()),
		Localaccount:       la,
		LocalAccountID:     la.GetResourceId(),
		UpdateStatusDetail: invInstance.GetUpdateStatusDetail(),
		WorkloadMembers:    workloadMembers,
		UpdatePolicy:       oup,
		Timestamps:         GrpcToOpenAPITimestamps(invInstance),
		ExistingCves:       invInstance.GetExistingCves(),
		RuntimePackages:    invInstance.GetRuntimePackages(),
		OsUpdateAvailable:  invInstance.GetOsUpdateAvailable(),
		CustomConfig:       customConfigs,
		CustomConfigID:     customConfigIDs,
	}

	fromInvInstanceStatus(invInstance, instance)
	return instance, nil
}

func (is *InventorygRPCServer) CreateInstance(
	ctx context.Context,
	req *restv1.CreateInstanceRequest,
) (*computev1.InstanceResource, error) {
	zlog.Debug().Msg("CreateInstance")

	instance := req.GetInstance()
	invInstance, err := toInvInstance(instance)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory instance")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: invInstance,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create instance in inventory")
		return nil, errors.Wrap(err)
	}

	instanceCreated, err := fromInvInstance(invResp.GetInstance())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory instance")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Created %s", instanceCreated)
	return instanceCreated, nil
}

// Get a list of instances.
func (is *InventorygRPCServer) ListInstances(
	ctx context.Context,
	req *restv1.ListInstancesRequest,
) (*restv1.ListInstancesResponse, error) {
	zlog.Debug().Msg("ListInstances")

	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Instance{Instance: &inv_computev1.InstanceResource{}}},
		Offset:   req.GetOffset(),
		Limit:    req.GetPageSize(),
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}
	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list instances from inventory")
		return nil, errors.Wrap(err)
	}

	invResources := invResp.GetResources()
	instances := make([]*computev1.InstanceResource, 0, len(invResources))
	for _, invRes := range invResources {
		instance, err := fromInvInstance(invRes.GetResource().GetInstance())
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to convert from inventory instance")
			return nil, errors.Wrap(err)
		}
		instances = append(instances, instance)
	}

	resp := &restv1.ListInstancesResponse{
		Instances:     instances,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific instance.
func (is *InventorygRPCServer) GetInstance(
	ctx context.Context,
	req *restv1.GetInstanceRequest,
) (*computev1.InstanceResource, error) {
	zlog.Debug().Msg("GetInstance")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get instance from inventory")
		return nil, errors.Wrap(err)
	}

	invInstance := invResp.GetResource().GetInstance()
	instance, err := fromInvInstance(invInstance)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory instance")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Got %s", instance)
	return instance, nil
}

// Update an instance. (PUT).
func (is *InventorygRPCServer) UpdateInstance(
	ctx context.Context,
	req *restv1.UpdateInstanceRequest,
) (*computev1.InstanceResource, error) {
	zlog.Debug().Msg("UpdateInstance")

	instance := req.GetInstance()
	invInstance, err := toInvInstance(instance)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory instance")
		return nil, errors.Wrap(err)
	}

	fieldmask, err := fieldmaskpb.New(invInstance, maps.Values(OpenAPIInstanceToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: invInstance,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, errors.Wrap(err)
	}
	invUp := upRes.GetInstance()
	invUpRes, err := fromInvInstance(invUp)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update an instance. (PATCH).
func (is *InventorygRPCServer) PatchInstance(
	ctx context.Context,
	req *restv1.PatchInstanceRequest,
) (*computev1.InstanceResource, error) {
	zlog.Debug().Msg("PatchInstance")

	instance := req.GetInstance()
	invInstance, err := toInvInstance(instance)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory instance")
		return nil, errors.Wrap(err)
	}

	fieldmask, err := parseFielmask(invInstance, req.GetFieldMask(), OpenAPIInstanceToProto)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: invInstance,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, errors.Wrap(err)
	}
	invUp := upRes.GetInstance()
	invUpRes, err := fromInvInstance(invUp)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Invalidate an instance.
func (is *InventorygRPCServer) InvalidateInstance(
	ctx context.Context,
	req *restv1.InvalidateInstanceRequest,
) (*restv1.InvalidateInstanceResponse, error) {
	zlog.Debug().Msg("InvalidateInstance")
	res := &inventory.Resource{
		Resource: &inventory.Resource_Instance{
			Instance: &inv_computev1.InstanceResource{
				DesiredState: inv_computev1.InstanceState_INSTANCE_STATE_UNTRUSTED,
			},
		},
	}

	fm, err := fieldmaskpb.New(res.GetInstance(), inv_computev1.InstanceResourceFieldDesiredState)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, errors.Wrap(err)
	}

	_, err = is.InvClient.Update(ctx, req.GetResourceId(), fm, res)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to invalidate instance in inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Invalidated %s", req.GetResourceId())
	return &restv1.InvalidateInstanceResponse{}, nil
}

// deleteOSUpdateRunsForInstance deletes all OSUpdateRun resources linked to the given instance ID.
func (is *InventorygRPCServer) deleteOSUpdateRunsForInstance(ctx context.Context, instanceID string) error {
	zlog.Debug().Msgf("Deleting OSUpdateRuns for instance %s", instanceID)

	// List all OSUpdateRuns that reference this instance
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_OsUpdateRun{
				OsUpdateRun: &inv_computev1.OSUpdateRunResource{},
			},
		},
		Filter: fmt.Sprintf("%s.%s = %q",
			inv_computev1.OSUpdateRunResourceEdgeInstance,
			inv_computev1.InstanceResourceFieldResourceId,
			instanceID),
	}

	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate OSUpdateRun filter")
		return errors.Wrap(err)
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msgf("Failed to list OSUpdateRuns for instance %s", instanceID)
		return errors.Wrap(err)
	}

	// Delete each OSUpdateRun
	for _, invRes := range invResp.GetResources() {
		osUpdateRun := invRes.GetResource().GetOsUpdateRun()
		if osUpdateRun != nil {
			runID := osUpdateRun.GetResourceId()
			zlog.Debug().Msgf("Deleting OSUpdateRun %s for instance %s", runID, instanceID)

			_, err := is.InvClient.Delete(ctx, runID)
			if err != nil {
				zlog.InfraErr(err).Msgf("Failed to delete OSUpdateRun %s for instance %s", runID, instanceID)
				return errors.Wrap(err)
			}
		}
	}

	zlog.Debug().Msgf("Successfully deleted %d OSUpdateRuns for instance %s", len(invResp.GetResources()), instanceID)
	return nil
}

// Delete an instance.
func (is *InventorygRPCServer) DeleteInstance(
	ctx context.Context,
	req *restv1.DeleteInstanceRequest,
) (*restv1.DeleteInstanceResponse, error) {
	zlog.Debug().Msg("DeleteInstance")

	instanceID := req.GetResourceId()

	// First, delete all OSUpdateRuns linked to this instance
	if err := is.deleteOSUpdateRunsForInstance(ctx, instanceID); err != nil {
		zlog.InfraErr(err).Msgf("Failed to delete OSUpdateRuns for instance %s", instanceID)
		return nil, errors.Wrap(err)
	}

	// Then delete the instance itself
	_, err := is.InvClient.Delete(ctx, instanceID)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete instance from inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted instance %s", instanceID)
	return &restv1.DeleteInstanceResponse{}, nil
}
