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

func TestListOSUpdateRun(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.ListOSUpdateRun(context.Background(), &restv1.ListOSUpdateRunRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.OK, status.Code(err))
}

func TestGetOSUpdateRun(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.GetOSUpdateRun(context.Background(), &restv1.GetOSUpdateRunRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.OK, status.Code(err))
}

func TestDeleteOSUpdateRun(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	_, err := server.DeleteOSUpdateRun(context.Background(), &restv1.DeleteOSUpdateRunRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.OK, status.Code(err))
}
