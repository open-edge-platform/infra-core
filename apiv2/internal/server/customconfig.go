// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"google.golang.org/grpc/codes"
	//	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
)

func (is *InventorygRPCServer) CreateCustomConfig(
	ctx context.Context,
	req *restv1.CreateInstanceRequest,
) (*computev1.CustomConfigResource, error) {
	zlog.Debug().Msg("CreateCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "CreateCustomConfig not implemented")
}

func (is *InventorygRPCServer) ListCustomConfigs(
	ctx context.Context,
	req *restv1.ListCustomConfigsRequest,
) (*restv1.ListCustomConfigsResponse, error) {
	zlog.Debug().Msg("ListCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "ListCustomConfigs not implemented")
}

func (is *InventorygRPCServer) GetCustomConfig(
	ctx context.Context,
	req *restv1.GetCustomConfigRequest,
) (*computev1.CustomConfigResource, error) {
	zlog.Debug().Msg("GetCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "GetCustomConfig not implemented")
}

func (is *InventorygRPCServer) DeleteCustomConfig(
	ctx context.Context,
	req *restv1.DeleteCustomConfigRequest,
) (*restv1.DeleteCustomConfigResponse, error) {
	zlog.Debug().Msg("DeleteCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "DeleteCustomConfig not implemented")

}
