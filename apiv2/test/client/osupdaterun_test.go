// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
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

//nolint:gosec // uint64 conversions are safe for testing
func TestOSUpdateRun_CreateGetDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Create OSUpdateRun (mimic infrastructure southbound API call)
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	host := dao.CreateHost(t, tenantID)
	os := dao.CreateOs(t, tenantID)
	instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
	osUpdatePolicy := dao.CreateOSUpdatePolicy(
		t, tenantID,
		inv_testing.OsUpdatePolicyName("test-policy"),
		inv_testing.OSUpdatePolicyLatest())

	osUpdateRun := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test-run"),
		inv_testing.OsUpdateRunDescription("Test OS update run"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())))

	// Test GET OSUpdateRun
	getResp, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, osUpdateRun.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.NotNil(t, getResp.JSON200)
	assert.Equal(t, osUpdateRun.ResourceId, *getResp.JSON200.ResourceId)
	assert.Equal(t, osUpdateRun.Name, *getResp.JSON200.Name)
	assert.Equal(t, osUpdateRun.Description, *getResp.JSON200.Description)

	// Test LIST OSUpdateRuns
	listResp, err := apiClient.OSUpdateRunListOSUpdateRunWithResponse(
		ctx, &api.OSUpdateRunListOSUpdateRunParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)
	assert.Len(t, listResp.JSON200.OsUpdateRuns, 1)
	assert.Equal(t, osUpdateRun.ResourceId, *listResp.JSON200.OsUpdateRuns[0].ResourceId)

	// Test DELETE OSUpdateRun
	deleteResp, err := apiClient.OSUpdateRunDeleteOSUpdateRunWithResponse(
		ctx, osUpdateRun.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode())

	// Verify resource is deleted
	getAfterDelete, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, osUpdateRun.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getAfterDelete.StatusCode())
}

//nolint:gosec // uint64 conversions are safe for testing
func TestOSUpdateRun_ListMultiple(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Mimic use of SB API to create multiple OSUpdateRuns
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	host := dao.CreateHost(t, tenantID)
	os := dao.CreateOs(t, tenantID)
	instance := dao.CreateInstanceWithOpts(t, tenantID, host, os, true)
	osUpdatePolicy := dao.CreateOSUpdatePolicy(
		t, tenantID,
		inv_testing.OsUpdatePolicyName("test-policy"),
		inv_testing.OSUpdatePolicyLatest())

	run1 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test-run-1"),
		inv_testing.OsUpdateRunDescription("First test run"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())))

	run2 := dao.CreateOSUpdateRun(t, tenantID,
		inv_testing.OsUpdateRunName("test-run-2"),
		inv_testing.OsUpdateRunDescription("Second test run"),
		inv_testing.OSUpdateRunAppliedPolicy(osUpdatePolicy),
		inv_testing.OSUpdateRunInstance(instance),
		inv_testing.OSUpdateRunStartTime(uint64(time.Now().Unix())))

	// Test LIST OSUpdateRuns returns both
	listResp, err := apiClient.OSUpdateRunListOSUpdateRunWithResponse(
		ctx, &api.OSUpdateRunListOSUpdateRunParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)
	assert.Len(t, listResp.JSON200.OsUpdateRuns, 2)

	// Verify both OSUpdateRuns are in the list
	resourceIDs := []string{*listResp.JSON200.OsUpdateRuns[0].ResourceId, *listResp.JSON200.OsUpdateRuns[1].ResourceId}
	assert.Contains(t, resourceIDs, run1.ResourceId)
	assert.Contains(t, resourceIDs, run2.ResourceId)
}

func TestOSUpdateRun_DeleteNonExistent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	osUpdateRunNonexistResourceID := "osupdaterun-111111"

	// Test DELETE non-existent OSUpdateRun
	deleteResp, err := apiClient.OSUpdateRunDeleteOSUpdateRunWithResponse(
		ctx, osUpdateRunNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode())
}

func TestOSUpdateRun_InvalidResourceID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	invalidResourceID := "invalid-resource-id"

	// Test GET with invalid resource ID format
	getResp, err := apiClient.OSUpdateRunGetOSUpdateRunWithResponse(
		ctx, invalidResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, getResp.StatusCode())

	// Test DELETE with invalid resource ID format
	deleteResp, err := apiClient.OSUpdateRunDeleteOSUpdateRunWithResponse(
		ctx, invalidResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, deleteResp.StatusCode())
}
