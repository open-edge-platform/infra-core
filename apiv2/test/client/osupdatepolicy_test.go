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
	policy2 := CreateOsUpdatePolicy(ctx, t, apiClient, utils.OsUpdatePolicyRequest1)

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
	found1 := false
	found2 := false
	for _, p := range listResp.JSON200.OsUpdatePolicies {
		if p.Name == policy1.JSON200.Name {
			found1 = true
		}
		if p.Name == policy2.JSON200.Name {
			found2 = true
		}
	}
	assert.True(t, found1, "First created policy should be in the list")
	assert.True(t, found2, "Second created policy should be in the list")
}
