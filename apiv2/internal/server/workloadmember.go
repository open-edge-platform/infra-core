// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIWorkloadMemberToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs WorkloadMember defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPIWorkloadMemberToProto = map[string]string{
	computev1.WorkloadMemberFieldKind:       inv_computev1.WorkloadMemberFieldKind,
	computev1.WorkloadMemberFieldInstanceId: inv_computev1.WorkloadMemberEdgeInstance,
	computev1.WorkloadMemberFieldWorkloadId: inv_computev1.WorkloadMemberEdgeWorkload,
}

func toInvWorkloadMember(workloadMember *computev1.WorkloadMember) (*inv_computev1.WorkloadMember, error) {
	if workloadMember == nil {
		return &inv_computev1.WorkloadMember{}, nil
	}

	invWorkloadMember := &inv_computev1.WorkloadMember{
		Kind: inv_computev1.WorkloadMemberKind(workloadMember.GetKind()),
	}

	workloadID := workloadMember.GetWorkloadId()
	if isSet(&workloadID) {
		invWorkloadMember.Workload = &inv_computev1.WorkloadResource{
			ResourceId: workloadID,
		}
	}

	instanceID := workloadMember.GetInstanceId()
	if isSet(&instanceID) {
		invWorkloadMember.Instance = &inv_computev1.InstanceResource{
			ResourceId: instanceID,
		}
	}

	err := validator.ValidateMessage(invWorkloadMember)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invWorkloadMember, nil
}

func fromInvWorkloadMember(invWorkloadMember *inv_computev1.WorkloadMember) (*computev1.WorkloadMember, error) {
	if invWorkloadMember == nil {
		return &computev1.WorkloadMember{}, nil
	}
	workload, err := fromInvWorkload(invWorkloadMember.GetWorkload())
	if err != nil {
		return nil, err
	}

	instance, err := fromInvInstance(invWorkloadMember.GetInstance())
	if err != nil {
		return nil, err
	}

	workloadMember := &computev1.WorkloadMember{
		ResourceId:       invWorkloadMember.GetResourceId(),
		WorkloadMemberId: invWorkloadMember.GetResourceId(),
		Kind:             computev1.WorkloadMemberKind(invWorkloadMember.GetKind()),
		Workload:         workload,
		Instance:         instance,
		InstanceId:       instance.GetResourceId(),
		Member:           instance,
		WorkloadId:       workload.GetResourceId(),
		Timestamps:       GrpcToOpenAPITimestamps(invWorkloadMember),
	}

	return workloadMember, nil
}

func (is *InventorygRPCServer) CreateWorkloadMember(
	ctx context.Context,
	req *restv1.CreateWorkloadMemberRequest,
) (*computev1.WorkloadMember, error) {
	zlog.Debug().Msg("CreateWorkloadMember")

	workloadMember := req.GetWorkloadMember()
	invWorkloadMember, err := toInvWorkloadMember(workloadMember)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory workload member")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_WorkloadMember{
			WorkloadMember: invWorkloadMember,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create workload member in inventory")
		return nil, errors.Wrap(err)
	}

	workloadMemberCreated, err := fromInvWorkloadMember(invResp.GetWorkloadMember())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory workload member")
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Created %s", workloadMemberCreated)
	return workloadMemberCreated, nil
}

// Get a list of workloadMembers.
func (is *InventorygRPCServer) ListWorkloadMembers(
	ctx context.Context,
	req *restv1.ListWorkloadMembersRequest,
) (*restv1.ListWorkloadMembersResponse, error) {
	zlog.Debug().Msg("ListWorkloadMembers")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, errors.Wrap(err)
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{
			Resource: &inventory.Resource_WorkloadMember{WorkloadMember: &inv_computev1.WorkloadMember{}},
		},
		Offset:  offset,
		Limit:   limit,
		OrderBy: req.GetOrderBy(),
		Filter:  req.GetFilter(),
	}
	if err = validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list workload members from inventory")
		return nil, errors.Wrap(err)
	}

	workloadMembers := []*computev1.WorkloadMember{}
	for _, invRes := range invResp.GetResources() {
		workloadMember, err := fromInvWorkloadMember(invRes.GetResource().GetWorkloadMember())
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to convert from inventory workload member")
			return nil, errors.Wrap(err)
		}
		workloadMembers = append(workloadMembers, workloadMember)
	}

	resp := &restv1.ListWorkloadMembersResponse{
		WorkloadMembers: workloadMembers,
		TotalElements:   invResp.GetTotalElements(),
		HasNext:         invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific workloadMember.
func (is *InventorygRPCServer) GetWorkloadMember(
	ctx context.Context,
	req *restv1.GetWorkloadMemberRequest,
) (*computev1.WorkloadMember, error) {
	zlog.Debug().Msg("GetWorkloadMember")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get workload member from inventory")
		return nil, errors.Wrap(err)
	}

	invWorkloadMember := invResp.GetResource().GetWorkloadMember()
	workloadMember, err := fromInvWorkloadMember(invWorkloadMember)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory workload member")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Got %s", workloadMember)
	return workloadMember, nil
}

// Delete a workloadMember.
func (is *InventorygRPCServer) DeleteWorkloadMember(
	ctx context.Context,
	req *restv1.DeleteWorkloadMemberRequest,
) (*restv1.DeleteWorkloadMemberResponse, error) {
	zlog.Debug().Msg("DeleteWorkloadMember")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete workload member from inventory")
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteWorkloadMemberResponse{}, nil
}
