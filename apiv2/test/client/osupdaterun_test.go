// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
)

func TestOSUpdateRun_GetListNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	osUpdateRunNonexistResourceID := "osupdaterun-111111"

	// Get OSUpdateRun
	getResp, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, osUpdateRunNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode())

	// List OSUpdateRuns should not be found
	listResp, err := apiClient.OSUpdateRunListOSUpdateRunWithResponse(
		ctx, &api.OSUpdateRunListOSUpdateRunParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	// The returned list is empty
	assert.NotNil(t, listResp.JSON200)
	assert.Empty(t, listResp.JSON200.OsUpdateRuns, "Expected OSUpdateRuns list to be empty")
}
