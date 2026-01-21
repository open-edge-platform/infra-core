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
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
)

func TestOSUpdatePolicy_CreateGetListDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Create OSUpdatePolicy
	policy1 := CreateOsUpdatePolicy(ctx, t, apiClient, utils.OsUpdatePolicyRequest1)
	policy2 := CreateOsUpdatePolicy(ctx, t, apiClient, utils.OsUpdatePolicyRequest2)

	// Get OSUpdatePolicy
	getResp1, err := apiClient.OSUpdatePolicyGetOSUpdatePolicyWithResponse(
		ctx, *policy1.JSON200.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp1.StatusCode())
	assert.Equal(t, utils.OsUpdatePolicyName1, getResp1.JSON200.Name)

	getResp2, err := apiClient.OSUpdatePolicyGetOSUpdatePolicyWithResponse(
		ctx, *policy2.JSON200.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp2.StatusCode())
	assert.Equal(t, utils.OsUpdatePolicyName2, getResp2.JSON200.Name)

	// List OSUpdatePolicies
	listResp, err := apiClient.OSUpdatePolicyListOSUpdatePolicyWithResponse(
		ctx, &api.OSUpdatePolicyListOSUpdatePolicyParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	// Verify both policies are in the list
	require.NotNil(t, listResp.JSON200)
	require.GreaterOrEqual(t, len(listResp.JSON200.OsUpdatePolicies), 2, "Expected at least 2 policies in the list")

	// Collect all resource IDs from the list
	resourceIDs := make([]string, 0, len(listResp.JSON200.OsUpdatePolicies))
	for _, policy := range listResp.JSON200.OsUpdatePolicies {
		resourceIDs = append(resourceIDs, *policy.ResourceId)
	}
	assert.Contains(t, resourceIDs, *policy1.JSON200.ResourceId)
	assert.Contains(t, resourceIDs, *policy2.JSON200.ResourceId)
}

func TestOSUpdatePolicy_GetListNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Clean up any existing policies to ensure test isolation
	DeleteAllOSUpdatePolicies(ctx, t, apiClient)

	osUpdatePolicyNonexistResourceID := "osupdatepolicy-111111"

	// Get OSUpdatePolicy - should return 404 for non-existent ID
	getResp, err := apiClient.OSUpdatePolicyGetOSUpdatePolicyWithResponse(
		ctx, osUpdatePolicyNonexistResourceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode())

	// List OSUpdatePolicies should return empty list
	listResp, err := apiClient.OSUpdatePolicyListOSUpdatePolicyWithResponse(
		ctx, &api.OSUpdatePolicyListOSUpdatePolicyParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)
	assert.Empty(t, listResp.JSON200.OsUpdatePolicies, "Expected OSUpdatePolicies list to be empty")
}
