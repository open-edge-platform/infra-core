// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	osv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/os/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

func (is *InventorygRPCServer) CreateOSUpdatePolicy(ctx context.Context, req *restv1.CreateOSUpdatePolicyRequest) (
	*computev1.OSUpdatePolicy, error,
) {
	zlog.Debug().Msg("CreateOsUpdatePolicy")

	osUpdatePolicy := req.GetOsUpdatePolicy()
	invOSUpdatePolicy, err := toInvOSUpdatePolicy(osUpdatePolicy)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory instance")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_OsUpdatePolicy{
			OsUpdatePolicy: invOSUpdatePolicy,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create instance in inventory")
		return nil, errors.Wrap(err)
	}

	osPolicyCreated := fromInvOSUpdatePolicy(invResp.GetOsUpdatePolicy())
	zlog.Debug().Msgf("Created %s", osPolicyCreated)

	return osPolicyCreated, nil
}

func (is *InventorygRPCServer) ListOSUpdatePolicy(ctx context.Context, req *restv1.ListOSUpdatePolicyRequest) (
	*restv1.ListOSUpdatePolicyResponse, error,
) {
	zlog.Debug().Msg("ListOSUpdatePolicy")

	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_OsUpdatePolicy{OsUpdatePolicy: &inv_computev1.OSUpdatePolicyResource{}},
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
		zlog.InfraErr(err).Msg("Failed to list OSUpdatePolicies from inventory")
		return nil, errors.Wrap(err)
	}

	invResources := invResp.GetResources()
	osUpdatePolicies := make([]*computev1.OSUpdatePolicy, 0, len(invResources))
	collections.ForEach(invResources, func(invRes *inventory.GetResourceResponse) {
		osUpdatePolicies = append(osUpdatePolicies, fromInvOSUpdatePolicy(invRes.GetResource().GetOsUpdatePolicy()))
	})
	resp := &restv1.ListOSUpdatePolicyResponse{
		OsUpdatePolicies: osUpdatePolicies,
		TotalElements:    invResp.GetTotalElements(),
		HasNext:          invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

func (is *InventorygRPCServer) GetOSUpdatePolicy(ctx context.Context, req *restv1.GetOSUpdatePolicyRequest) (
	*computev1.OSUpdatePolicy, error,
) {
	zlog.Debug().Msg("GetOSUpdatePolicy")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get OSUpdatePolicy from inventory")
		return nil, errors.Wrap(err)
	}

	invOSUpdatePolicy := invResp.GetResource().GetOsUpdatePolicy()
	osUpPolicy := fromInvOSUpdatePolicy(invOSUpdatePolicy)
	zlog.Debug().Msgf("Got %s", osUpPolicy)
	return osUpPolicy, nil
}

func (is *InventorygRPCServer) DeleteOSUpdatePolicy(ctx context.Context, req *restv1.DeleteOSUpdatePolicyRequest) (
	*restv1.DeleteOSUpdatePolicyResponse, error,
) {
	zlog.Debug().Msg("DeleteOSUpdatePolicy")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete OSUpdatePolicy from inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteOSUpdatePolicyResponse{}, nil
}

func fromInvOSUpdatePolicy(invOSUpdatePolicy *inv_computev1.OSUpdatePolicyResource) *computev1.OSUpdatePolicy {
	if invOSUpdatePolicy == nil {
		return &computev1.OSUpdatePolicy{}
	}
	var targetOs *osv1.OperatingSystemResource

	if invOSUpdatePolicy.GetTargetOs() != nil {
		targetOs = fromInvOSResource(invOSUpdatePolicy.GetTargetOs())
	}
	osUpdatePolicy := &computev1.OSUpdatePolicy{
		ResourceId:      invOSUpdatePolicy.GetResourceId(),
		Name:            invOSUpdatePolicy.GetName(),
		Description:     invOSUpdatePolicy.GetDescription(),
		InstallPackages: invOSUpdatePolicy.GetInstallPackages(),
		UpdateSources:   invOSUpdatePolicy.GetUpdateSources(),
		KernelCommand:   invOSUpdatePolicy.GetKernelCommand(),
		TargetOs:        targetOs,
		UpdatePolicy:    computev1.UpdatePolicy(invOSUpdatePolicy.GetUpdatePolicy()),
		Timestamps:      GrpcToOpenAPITimestamps(invOSUpdatePolicy),
	}
	return osUpdatePolicy
}

func toInvOSUpdatePolicy(osUpdatePolicy *computev1.OSUpdatePolicy) (*inv_computev1.OSUpdatePolicyResource, error) {
	if osUpdatePolicy == nil {
		return &inv_computev1.OSUpdatePolicyResource{}, nil
	}

	invOSUpdatePolicy := &inv_computev1.OSUpdatePolicyResource{
		Name:            osUpdatePolicy.GetName(),
		Description:     osUpdatePolicy.GetDescription(),
		InstallPackages: osUpdatePolicy.GetInstallPackages(),
		UpdateSources:   osUpdatePolicy.GetUpdateSources(),
		KernelCommand:   osUpdatePolicy.GetKernelCommand(),
		UpdatePolicy:    inv_computev1.UpdatePolicy(osUpdatePolicy.GetUpdatePolicy()),
	}

	targetOSID := osUpdatePolicy.GetTargetOsId()
	if isSet(&targetOSID) {
		invOSUpdatePolicy.TargetOs = &inv_osv1.OperatingSystemResource{
			ResourceId: targetOSID,
		}
	}

	err := validator.ValidateMessage(invOSUpdatePolicy)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	return invOSUpdatePolicy, nil
}
