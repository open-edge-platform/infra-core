// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"google.golang.org/grpc/codes"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
)

func (is *InventorygRPCServer) CreateOSUpdatePolicy(_ context.Context, _ *restv1.CreateOSUpdatePolicyRequest) (
	*computev1.OSUpdatePolicy, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "CreateOSUpdatePolicy not implemented")
}

func (is *InventorygRPCServer) ListOSUpdatePolicy(_ context.Context, _ *restv1.ListOSUpdatePolicyRequest) (
	*restv1.ListOSUpdatePolicyResponse, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "ListOSUpdatePolicy not implemented")
}

func (is *InventorygRPCServer) GetOSUpdatePolicy(_ context.Context, _ *restv1.GetOSUpdatePolicyRequest) (
	*computev1.OSUpdatePolicy, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "GetOSUpdatePolicy not implemented")
}

func (is *InventorygRPCServer) DeleteOSUpdatePolicy(_ context.Context, _ *restv1.DeleteOSUpdatePolicyRequest) (
	*computev1.OSUpdatePolicy, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "DeleteOsUpdatePolicy not implemented")
}
