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

func (is *InventorygRPCServer) ListOSUpdateRun(_ context.Context, _ *restv1.ListOSUpdateRunRequest) (
	*restv1.ListOSUpdateRunResponse, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "ListOSUpdateRun not implemented")
}

func (is *InventorygRPCServer) GetOSUpdateRun(_ context.Context, _ *restv1.GetOSUpdateRunRequest) (
	*computev1.OSUpdateRun, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "GetOSUpdateRun not implemented")
}

func (is *InventorygRPCServer) DeleteOSUpdateRun(_ context.Context, _ *restv1.DeleteOSUpdateRunRequest) (
	*restv1.DeleteOSUpdateRunResponse, error,
) {
	// TODO implement me
	return nil, errors.Errorfc(codes.Unimplemented, "DeleteOSUpdateRun not implemented")
}
