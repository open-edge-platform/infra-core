// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
)

func TestComputeSummary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	utils.Site1Request.RegionId = nil
	s1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Site2Request.RegionId = nil
	s2 := CreateSite(ctx, t, apiClient, utils.Site2Request)

	// Baseline summaries (environment may already contain hosts)
	res, err := apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName, &api.HostServiceGetHostsSummaryParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	baselineTotalHost := *res.JSON200.Total
	baselineUnallocatedHost := *res.JSON200.Unallocated

	filterMetaInherited := fmt.Sprintf("metadata='{"+"\""+"key"+"\""+":%q,"+"\""+"value"+"\""+":%q}'",
		utils.MetadataHost2[0].Key, utils.MetadataHost2[0].Value)
	assert.Equal(t, `metadata='{"key":"examplekey1","value":"host2"}'`, filterMetaInherited)
	resMetaInherited, err := apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName,
		&api.HostServiceGetHostsSummaryParams{Filter: &filterMetaInherited},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resMetaInherited.StatusCode())
	baselineMetaInheritedTotal := *resMetaInherited.JSON200.Total
	baselineMetaInheritedUnallocated := *resMetaInherited.JSON200.Unallocated
	baselineMetaInheritedError := *resMetaInherited.JSON200.Error
	baselineMetaInheritedRunning := *resMetaInherited.JSON200.Running

	filterMetaStandalone := fmt.Sprintf("metadata='{"+"\""+"key"+"\""+":%q,"+"\""+"value"+"\""+":%q}'",
		utils.MetadataHost2[0].Key, utils.MetadataHost1[0].Value)
	assert.Equal(t, `metadata='{"key":"examplekey1","value":"host1"}'`, filterMetaStandalone)
	resMetaStandalone, err := apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName,
		&api.HostServiceGetHostsSummaryParams{Filter: &filterMetaStandalone},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resMetaStandalone.StatusCode())
	baselineMetaStandaloneTotal := *resMetaStandalone.JSON200.Total
	baselineMetaStandaloneUnallocated := *resMetaStandalone.JSON200.Unallocated
	baselineMetaStandaloneError := *resMetaStandalone.JSON200.Error
	baselineMetaStandaloneRunning := *resMetaStandalone.JSON200.Running

	filterSite1 := fmt.Sprintf("site.resourceId=%q", *s1.JSON200.SiteID)
	resSite1, err := apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName, &api.HostServiceGetHostsSummaryParams{Filter: &filterSite1}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSite1.StatusCode())
	baselineSite1Total := *resSite1.JSON200.Total
	baselineSite1Unallocated := *resSite1.JSON200.Unallocated
	baselineSite1Error := *resSite1.JSON200.Error
	baselineSite1Running := *resSite1.JSON200.Running

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
		CreateHost(ctx, t, apiClient, hostRequest)
	}

	// Hosts with Meta
	for i := 0; i < hostsWithoutSiteWithMeta; i++ {
		expectedTotalHost++
		expectedUnallocatedHost++
		hostRequest := GetHostRequestWithRandomUUID()
		hostRequest.Metadata = &utils.MetadataHost1
		CreateHost(ctx, t, apiClient, hostRequest)
	}

	// Hosts with site
	for i := 0; i < hostsWithSiteFromSite1; i++ {
		expectedTotalHost++
		hostRequest := GetHostRequestWithRandomUUID()
		hostRequest.SiteId = s1.JSON200.SiteID
		CreateHost(ctx, t, apiClient, hostRequest)
	}

	// Hosts with site and meta from site
	for i := 0; i < hostsWithSiteAndMetaFromSite2; i++ {
		expectedTotalHost++
		hostRequest := GetHostRequestWithRandomUUID()
		hostRequest.SiteId = s2.JSON200.SiteID
		hostRequest.Metadata = &utils.MetadataHost2
		CreateHost(ctx, t, apiClient, hostRequest)
	}

	// Total (all hosts)
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName, &api.HostServiceGetHostsSummaryParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	t.Logf("DEBUG: expectedTotalHost=%d, actual Total=%d", expectedTotalHost+baselineTotalHost, *res.JSON200.Total)
	t.Logf("DEBUG: expectedUnallocatedHost=%d, actual Unallocated=%d",
		expectedUnallocatedHost+baselineUnallocatedHost, *res.JSON200.Unallocated)
	assert.Equal(t, expectedTotalHost+baselineTotalHost, *res.JSON200.Total)
	assert.Equal(t, expectedUnallocatedHost+baselineUnallocatedHost, *res.JSON200.Unallocated)

	// Filter by metadata (inherited) `metadata='{"key":"examplekey3","value":"host2"}'`
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName,
		&api.HostServiceGetHostsSummaryParams{Filter: &filterMetaInherited},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, baselineMetaInheritedTotal+hostsWithSiteAndMetaFromSite2, *res.JSON200.Total)
	assert.Equal(t, baselineMetaInheritedUnallocated, *res.JSON200.Unallocated)
	assert.Equal(t, baselineMetaInheritedError, *res.JSON200.Error)
	assert.Equal(t, baselineMetaInheritedRunning, *res.JSON200.Running)

	// Filter by metadata (standalone) `metadata='{"key":"examplekey3","value":"host2"}'`
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName,
		&api.HostServiceGetHostsSummaryParams{Filter: &filterMetaStandalone},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, baselineMetaStandaloneTotal+hostsWithoutSiteWithMeta, *res.JSON200.Total)
	assert.Equal(t, baselineMetaStandaloneUnallocated+hostsWithoutSiteWithMeta, *res.JSON200.Unallocated)
	assert.Equal(t, baselineMetaStandaloneError, *res.JSON200.Error)
	assert.Equal(t, baselineMetaStandaloneRunning, *res.JSON200.Running)

	// Filter by host's site-id
	res, err = apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx, projectName, &api.HostServiceGetHostsSummaryParams{Filter: &filterSite1}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, baselineSite1Total+hostsWithSiteFromSite1, *res.JSON200.Total)
	assert.Equal(t, baselineSite1Unallocated, *res.JSON200.Unallocated)
	assert.Equal(t, baselineSite1Error, *res.JSON200.Error)
	assert.Equal(t, baselineSite1Running, *res.JSON200.Running)
	// Cleanup done in create helper functions
}
