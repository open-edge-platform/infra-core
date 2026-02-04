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

	projectName := getProjectID(t)

	// Get OSUpdateRun
	getResp, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, projectName, osUpdateRunNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode())

	// List OSUpdateRuns should not be found
	listResp, err := apiClient.OSUpdateRunListOSUpdateRunWithResponse(
		ctx, projectName, &api.OSUpdateRunListOSUpdateRunParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	// The returned list is empty
	assert.NotNil(t, listResp.JSON200)
	assert.Empty(t, listResp.JSON200.OsUpdateRuns, "Expected OSUpdateRuns list to be empty")
}

func TestOSUpdateRun_CreateGetDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Note: OSUpdateRun resources cannot be created via the northbound API
	// They are created automatically by the infrastructure when OS updates are applied
	// This test verifies the API behavior when working with non-existent resources

	osUpdateRunNonexistResourceID := "osupdaterun-999999"

	projectName := getProjectID(t)

	// Test GET non-existent OSUpdateRun - should return 404
	getResp, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, projectName, osUpdateRunNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode())

	// Test LIST OSUpdateRuns - should return empty list or existing runs from infrastructure
	listResp, err := apiClient.OSUpdateRunListOSUpdateRunWithResponse(
		ctx, projectName, &api.OSUpdateRunListOSUpdateRunParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)
	// Don't assert specific count as there may be existing runs from infrastructure

	// Test DELETE non-existent OSUpdateRun - should return 404
	deleteResp, err := apiClient.OSUpdateRunDeleteOSUpdateRunWithResponse(
		ctx, projectName, osUpdateRunNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode())
}

func TestOSUpdateRun_ListMultiple(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Note: OSUpdateRun resources are created by infrastructure when OS updates occur
	// This test verifies the LIST operation works correctly

	projectName := getProjectID(t)

	// Test LIST OSUpdateRuns - should return successfully even if empty
	listResp, err := apiClient.OSUpdateRunListOSUpdateRunWithResponse(
		ctx, projectName, &api.OSUpdateRunListOSUpdateRunParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)

	// Verify the response structure is correct
	assert.NotNil(t, listResp.JSON200.OsUpdateRuns)

	// If there are any runs, verify they have the required fields
	if len(listResp.JSON200.OsUpdateRuns) > 0 {
		for _, run := range listResp.JSON200.OsUpdateRuns {
			assert.NotNil(t, run.ResourceId)
			assert.NotEmpty(t, *run.ResourceId)
		}
	}
}

func TestOSUpdateRun_DeleteNonExistent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	osUpdateRunNonexistResourceID := "osupdaterun-111111"

	projectName := getProjectID(t)

	// Test DELETE non-existent OSUpdateRun
	deleteResp, err := apiClient.OSUpdateRunDeleteOSUpdateRunWithResponse(
		ctx, projectName, osUpdateRunNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode())
}

func TestOSUpdateRun_InvalidResourceID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	invalidResourceID := "invalid-resource-id"

	projectName := getProjectID(t)

	// Test GET with invalid resource ID format
	getResp, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, projectName, invalidResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, getResp.StatusCode())

	// Test DELETE with invalid resource ID format
	deleteResp, err := apiClient.OSUpdateRunDeleteOSUpdateRunWithResponse(
		ctx, projectName, invalidResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, deleteResp.StatusCode())
}
