// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"google.golang.org/grpc/codes"

	customconfig_v1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/customconfig/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
)

func (is *InventorygRPCServer) CreateCustomConfig(
	_ context.Context,
	_ *restv1.CreateCustomConfigRequest,
) (*customconfig_v1.CustomConfigResource, error) {
	zlog.Debug().Msg("CreateCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "CreateCustomConfig not implemented")
}

func (is *InventorygRPCServer) ListCustomConfigs(
	_ context.Context,
	_ *restv1.ListCustomConfigsRequest,
) (*restv1.ListCustomConfigsResponse, error) {
	zlog.Debug().Msg("ListCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "ListCustomConfigs not implemented")
}

func (is *InventorygRPCServer) GetCustomConfig(
	_ context.Context,
	_ *restv1.GetCustomConfigRequest,
) (*customconfig_v1.CustomConfigResource, error) {
	zlog.Debug().Msg("GetCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "GetCustomConfig not implemented")
}

func (is *InventorygRPCServer) DeleteCustomConfig(
	_ context.Context,
	_ *restv1.DeleteCustomConfigRequest,
) (*restv1.DeleteCustomConfigResponse, error) {
	zlog.Debug().Msg("DeleteCustomConfig")
	return nil, errors.Errorfc(codes.Unimplemented, "DeleteCustomConfig not implemented")
}
