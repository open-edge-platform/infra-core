// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	invserver "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestCreateOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.CreateOSUpdatePolicy(context.Background(), &restv1.CreateOSUpdatePolicyRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "CreateOSUpdatePolicy not implemented")
}

func TestListOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.ListOSUpdatePolicy(context.Background(), &restv1.ListOSUpdatePolicyRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "ListOSUpdatePolicy not implemented")
}

func TestGetOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.GetOSUpdatePolicy(context.Background(), &restv1.GetOSUpdatePolicyRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "GetOSUpdatePolicy not implemented")
}

func TestDeleteOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.DeleteOSUpdatePolicy(context.Background(), &restv1.DeleteOSUpdatePolicyRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Unimplemented, status.Code(err))
	assert.Contains(t, err.Error(), "DeleteOsUpdatePolicy not implemented")
}
