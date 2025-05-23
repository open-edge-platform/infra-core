// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeSummary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.RegionId = nil
	s1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Site2Request.RegionId = nil
	s2 := CreateSite(t, ctx, apiClient, utils.Site2Request)

	expectedTotalHost := 0
	expectedUnallocatedHost := 0

	hostsWithSiteAndMetaFromSite2 := 31
	hostsWithoutSiteWithMeta := 15
	hostsWithSiteFromSite1 := 20

	// Hosts without site
	for i := 1; i < 15; i++ {
		expectedTotalHost++
		expectedUnallocatedHost++
		hostRequest := GetHostRequestWithRandomUUID()
		CreateHost(t, ctx, apiClient, hostRequest)
	}

	// Hosts with Meta
	for i := 0; i < hostsWithoutSiteWithMeta; i++ {
		expectedTotalHost++
		expectedUnallocatedHost++
		hostRequest := GetHostRequestWithRandomUUID()
		hostRequest.Metadata = &utils.MetadataHost1
		CreateHost(t, ctx, apiClient, hostRequest)
	}

	// Hosts with site
	for i := 0; i < hostsWithSiteFromSite1; i++ {
		expectedTotalHost++
		hostRequest := GetHostRequestWithRandomUUID()
		hostRequest.SiteId = s1.JSON200.SiteID
		CreateHost(t, ctx, apiClient, hostRequest)
	}

	// Hosts with site and meta from site
	for i := 0; i < hostsWithSiteAndMetaFromSite2; i++ {
		expectedTotalHost++
		hostRequest := GetHostRequestWithRandomUUID()
		hostRequest.SiteId = s2.JSON200.SiteID
		hostRequest.Metadata = &utils.MetadataHost2
		CreateHost(t, ctx, apiClient, hostRequest)
	}

	// Total (all hosts)
	res, err := apiClient.HostServiceGetHostsSummaryWithResponse(ctx, &api.HostServiceGetHostsSummaryParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, expectedTotalHost, int(*res.JSON200.Total))
	assert.Equal(t, expectedUnallocatedHost, int(*res.JSON200.Unallocated))

	// Filter by metadata (inherited) `metadata='{"key":"examplekey3","value":"host2"}'`
	filter := fmt.Sprintf("metadata='{\"key\":\"%s\",\"value\":\"%s\"}'",
		utils.MetadataHost2[0].Key, utils.MetadataHost2[0].Value)
	assert.Equal(t, `metadata='{"key":"examplekey1","value":"host2"}'`, filter)
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(ctx, &api.HostServiceGetHostsSummaryParams{Filter: &filter}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, hostsWithSiteAndMetaFromSite2, int(*res.JSON200.Total))
	assert.Zero(t, *res.JSON200.Unallocated)
	assert.Zero(t, *res.JSON200.Error)
	assert.Zero(t, *res.JSON200.Running)

	// Filter by metadata (standalone) `metadata='{"key":"examplekey3","value":"host2"}'`
	filter = fmt.Sprintf("metadata='{\"key\":\"%s\",\"value\":\"%s\"}'",
		utils.MetadataHost2[0].Key, utils.MetadataHost1[0].Value)
	assert.Equal(t, `metadata='{"key":"examplekey1","value":"host1"}'`, filter)
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(ctx, &api.HostServiceGetHostsSummaryParams{Filter: &filter}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, hostsWithoutSiteWithMeta, int(*res.JSON200.Total))
	assert.Equal(t, hostsWithoutSiteWithMeta, int(*res.JSON200.Unallocated))
	assert.Zero(t, *res.JSON200.Error)
	assert.Zero(t, *res.JSON200.Running)

	// Filter by host's site-id
	filter = fmt.Sprintf("site.resourceId=\"%s\"", *s1.JSON200.SiteID)
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(ctx, &api.HostServiceGetHostsSummaryParams{Filter: &filter}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, hostsWithSiteFromSite1, int(*res.JSON200.Total))
	assert.Zero(t, *res.JSON200.Unallocated)
	assert.Zero(t, *res.JSON200.Error)
	assert.Zero(t, *res.JSON200.Running)
	// Cleanup done in create helper functions
}
