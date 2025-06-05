// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	invserver "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
)

func TestCustomConfig_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.CreateCustomConfig(context.Background(), &restv1.CreateCustomConfigRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "CreateCustomConfig not implemented")
}

func TestCustomConfig_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.GetCustomConfig(context.Background(), &restv1.GetCustomConfigRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "GetCustomConfig not implemented")
}

func TestCustomConfig_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.ListCustomConfigs(context.Background(), &restv1.ListCustomConfigsRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "ListCustomConfigs not implemented")
}

func TestCustomConfig_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.DeleteCustomConfig(context.Background(), &restv1.DeleteCustomConfigRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "DeleteCustomConfig not implemented")
}
