// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	localaccountv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/localaccount/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

func toInvLocalAccount(localaccount *localaccountv1.LocalAccountResource) (*inv_localaccountv1.LocalAccountResource, error) {
	if localaccount == nil {
		return &inv_localaccountv1.LocalAccountResource{}, nil
	}
	invLocalAccount := &inv_localaccountv1.LocalAccountResource{
		Username: localaccount.GetUsername(),
		SshKey:   localaccount.GetSshKey(),
	}

	err := validator.ValidateMessage(invLocalAccount)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invLocalAccount, nil
}

func fromInvLocalAccount(invLocalAccount *inv_localaccountv1.LocalAccountResource) *localaccountv1.LocalAccountResource {
	if invLocalAccount == nil {
		return &localaccountv1.LocalAccountResource{}
	}
	localaccount := &localaccountv1.LocalAccountResource{
		ResourceId: invLocalAccount.GetResourceId(),
		Username:   invLocalAccount.GetUsername(),
		SshKey:     invLocalAccount.GetSshKey(),
		Timestamps: GrpcToOpenAPITimestamps(invLocalAccount),
	}

	return localaccount
}

func (is *InventorygRPCServer) CreateLocalAccount(
	ctx context.Context,
	req *restv1.CreateLocalAccountRequest,
) (*localaccountv1.LocalAccountResource, error) {
	zlog.Debug().Msg("CreateLocalAccount")

	localaccount := req.GetLocalAccount()
	invLocalAccount, err := toInvLocalAccount(localaccount)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory localaccount")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_LocalAccount{
			LocalAccount: invLocalAccount,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create localaccount in inventory")
		return nil, err
	}

	localaccountCreated := fromInvLocalAccount(invResp.GetLocalAccount())
	zlog.Debug().Msgf("Created %s", localaccountCreated)
	return localaccountCreated, nil
}

// Get a list of localaccounts.
func (is *InventorygRPCServer) ListLocalAccounts(
	ctx context.Context,
	req *restv1.ListLocalAccountsRequest,
) (*restv1.ListLocalAccountsResponse, error) {
	zlog.Debug().Msg("ListLocalAccounts")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_LocalAccount{
			LocalAccount: &inv_localaccountv1.LocalAccountResource{},
		}},
		Offset:  offset,
		Limit:   limit,
		OrderBy: req.GetOrderBy(),
		Filter:  req.GetFilter(),
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list localaccounts from inventory")
		return nil, err
	}

	localaccounts := []*localaccountv1.LocalAccountResource{}
	for _, invRes := range invResp.GetResources() {
		localaccount := fromInvLocalAccount(invRes.GetResource().GetLocalAccount())
		localaccounts = append(localaccounts, localaccount)
	}

	resp := &restv1.ListLocalAccountsResponse{
		LocalAccounts: localaccounts,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific localaccount.
func (is *InventorygRPCServer) GetLocalAccount(
	ctx context.Context,
	req *restv1.GetLocalAccountRequest,
) (*localaccountv1.LocalAccountResource, error) {
	zlog.Debug().Msg("GetLocalAccount")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get localaccount from inventory")
		return nil, err
	}

	invLocalAccount := invResp.GetResource().GetLocalAccount()
	localaccount := fromInvLocalAccount(invLocalAccount)
	zlog.Debug().Msgf("Got %s", localaccount)
	return localaccount, nil
}

// Delete a localaccount.
func (is *InventorygRPCServer) DeleteLocalAccount(
	ctx context.Context,
	req *restv1.DeleteLocalAccountRequest,
) (*restv1.DeleteLocalAccountResponse, error) {
	zlog.Debug().Msg("DeleteLocalAccount")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete localaccount from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteLocalAccountResponse{}, nil
}
