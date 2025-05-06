// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
)

var (
	collectorKindHostLogs    = api.TelemetryLogsGroupResourceCollectorKindCOLLECTORKINDHOST
	collectorKindHostMetrics = api.TelemetryMetricsGroupResourceCollectorKindCOLLECTORKINDHOST
)

func clearIDs() {
	utils.Instance1Request.HostID = nil
	utils.Instance1Request.OsID = nil
	utils.Site1Request.Region = nil
	utils.Host1Request.Site = nil
}

func TestTelemetryGroup_CreateGetDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	allLogsGroups, err := apiClient.TelemetryLogsGroupServiceListTelemetryLogsGroupsWithResponse(
		ctx,
		&api.TelemetryLogsGroupServiceListTelemetryLogsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, allLogsGroups.StatusCode())
	for _, logsGroups := range allLogsGroups.JSON200.TelemetryLogsGroups {
		DeleteTelemetryLogsGroup(t, context.Background(), apiClient, *logsGroups.TelemetryLogsGroupId)
	}

	allMetricsGroups, err := apiClient.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsWithResponse(
		ctx,
		&api.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, allMetricsGroups.StatusCode())
	for _, metricsGroups := range allMetricsGroups.JSON200.TelemetryMetricsGroups {
		DeleteTelemetryMetricsGroup(t, context.Background(), apiClient, *metricsGroups.TelemetryMetricsGroupId)
	}

	res1 := CreateTelemetryLogsGroup(t, ctx, apiClient, utils.TelemetryLogsGroup1Request)
	res2 := CreateTelemetryMetricsGroup(t, ctx, apiClient, utils.TelemetryMetricsGroup1Request)

	// Assert presence of telemetry resources
	allLogsGroups, err = apiClient.TelemetryLogsGroupServiceListTelemetryLogsGroupsWithResponse(
		ctx,
		&api.TelemetryLogsGroupServiceListTelemetryLogsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, allLogsGroups.StatusCode())
	assert.Len(t, allLogsGroups.JSON200.TelemetryLogsGroups, 1)

	allMetricsGroups, err = apiClient.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsWithResponse(
		ctx,
		&api.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, allMetricsGroups.StatusCode())
	assert.Len(t, allMetricsGroups.JSON200.TelemetryMetricsGroups, 1)

	logsGroup, err := apiClient.TelemetryLogsGroupServiceGetTelemetryLogsGroupWithResponse(
		ctx,
		*res1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, logsGroup.StatusCode())
	assert.Equal(t, res1.JSON200.Name, logsGroup.JSON200.Name)
	assert.Equal(t, res1.JSON200.Groups, logsGroup.JSON200.Groups)
	assert.Equal(t, res1.JSON200.CollectorKind, logsGroup.JSON200.CollectorKind)

	metricsGroup, err := apiClient.TelemetryMetricsGroupServiceGetTelemetryMetricsGroupWithResponse(
		ctx,
		*res2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, metricsGroup.StatusCode())
	assert.Equal(t, res2.JSON200.Name, metricsGroup.JSON200.Name)
	assert.Equal(t, res2.JSON200.Groups, metricsGroup.JSON200.Groups)
	assert.Equal(t, res2.JSON200.CollectorKind, metricsGroup.JSON200.CollectorKind)

	// delete with auto-cleanup
}

func TestTelemetryLogsGroup_PostErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	testCases := map[string]struct {
		in                 api.TelemetryLogsGroupResource
		expectedHTTPStatus int
		valid              bool
	}{
		"Post_NoName_Status_BadRequest": {
			in: api.TelemetryLogsGroupResource{
				CollectorKind: collectorKindHostLogs,
				Groups:        []string{"test group"},
			},
			expectedHTTPStatus: http.StatusBadRequest,
		},
		"Post_NoCollectorKind_Status_BadRequest": {
			in: api.TelemetryLogsGroupResource{
				Name:   "Test Name",
				Groups: []string{"test group"},
			},
			expectedHTTPStatus: http.StatusBadRequest,
		},
		"Post_NoGroups_Status_BadRequest": {
			in: api.TelemetryLogsGroupResource{
				Name:          "Test Name",
				CollectorKind: collectorKindHostLogs,
			},
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			resp, reqErr := apiClient.TelemetryLogsGroupServiceCreateTelemetryLogsGroupWithResponse(
				ctx,
				tc.in,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, reqErr)
			assert.Equal(t, tc.expectedHTTPStatus, resp.StatusCode())
		})
	}
}

func TestTelemetryMetricsGroup_PostErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	testCases := map[string]struct {
		in                 api.TelemetryMetricsGroupResource
		expectedHTTPStatus int
		valid              bool
	}{
		"Post_NoName_Status_BadRequest": {
			in: api.TelemetryMetricsGroupResource{
				CollectorKind: collectorKindHostMetrics,
				Groups:        []string{"test group"},
			},
			expectedHTTPStatus: http.StatusBadRequest,
		},
		"Post_NoCollectorKind_Status_BadRequest": {
			in: api.TelemetryMetricsGroupResource{
				Name:   "Test Name",
				Groups: []string{"test group"},
			},
			expectedHTTPStatus: http.StatusBadRequest,
		},
		"Post_NoGroups_Status_BadRequest": {
			in: api.TelemetryMetricsGroupResource{
				Name:          "Test Name",
				CollectorKind: collectorKindHostMetrics,
			},
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			resp, reqErr := apiClient.TelemetryMetricsGroupServiceCreateTelemetryMetricsGroupWithResponse(
				ctx,
				tc.in,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, reqErr)
			assert.Equal(t, tc.expectedHTTPStatus, resp.StatusCode())
		})
	}
}

func TestTelemetryGroup_GetDeleteErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	testCases := map[string]struct {
		ID                 string
		expectedHTTPStatus int
		valid              bool
	}{
		"UnexistingID_Status_NotFound": {
			ID:                 "telemetrygroup-00000000",
			expectedHTTPStatus: http.StatusNotFound,
		},
		"InvalidID_Status_NotFound": {
			ID:                 "telemetrygroup-XXXXXX",
			expectedHTTPStatus: http.StatusNotFound,
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			resp1, reqErr := apiClient.TelemetryLogsGroupServiceGetTelemetryLogsGroupWithResponse(
				ctx,
				tc.ID,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, reqErr)
			assert.Equal(t, tc.expectedHTTPStatus, resp1.StatusCode())

			resp2, reqErr := apiClient.TelemetryMetricsGroupServiceGetTelemetryMetricsGroupWithResponse(
				ctx,
				tc.ID,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, reqErr)
			assert.Equal(t, tc.expectedHTTPStatus, resp2.StatusCode())

			respDel1, reqErr := apiClient.TelemetryLogsGroupServiceDeleteTelemetryLogsGroupWithResponse(
				ctx,
				tc.ID,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, reqErr)
			assert.Equal(t, tc.expectedHTTPStatus, respDel1.StatusCode())

			respDel2, reqErr := apiClient.TelemetryMetricsGroupServiceDeleteTelemetryMetricsGroupWithResponse(
				ctx,
				tc.ID,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, reqErr)
			assert.Equal(t, tc.expectedHTTPStatus, respDel2.StatusCode())
		})
	}
}

func TestTelemetryProfile_CreateGetDelete(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	site1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = site1.JSON200.ResourceId
	hostCreated1 := CreateHost(t, ctx, apiClient, utils.Host1Request)
	osCreated1 := CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance1Request.OsID = osCreated1.JSON200.ResourceId
	inst1 := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	telemetryGroupMetrics1 := utils.TelemetryMetricsGroup1Request
	telemetryGroupMetrics2 := api.TelemetryMetricsGroupResource{
		Name:          "CPU Usage",
		CollectorKind: collectorKindHostMetrics,
		Groups: []string{
			"cpu",
		},
	}

	logsGroup := CreateTelemetryLogsGroup(t, ctx, apiClient, utils.TelemetryLogsGroup1Request)
	metricsGroup1 := CreateTelemetryMetricsGroup(t, ctx, apiClient, telemetryGroupMetrics1)
	metricsGroup2 := CreateTelemetryMetricsGroup(t, ctx, apiClient, telemetryGroupMetrics2)

	TelemetryLogsProfilePerInstance := api.TelemetryLogsProfileResource{
		LogLevel:       api.SEVERITYLEVELDEBUG,
		TargetInstance: inst1.JSON200.ResourceId,
		LogsGroupId:    *logsGroup.JSON200.ResourceId,
	}
	TelemetryMetricsProfilePerSite := api.TelemetryMetricsProfileResource{
		MetricsInterval: 300,
		TargetSite:      site1.JSON200.ResourceId,
		MetricsGroupId:  *metricsGroup1.JSON200.ResourceId,
	}
	TelemetryMetricsProfilePerRegion := api.TelemetryMetricsProfileResource{
		MetricsInterval: 300,
		TargetRegion:    r1.JSON200.ResourceId,
		MetricsGroupId:  *metricsGroup2.JSON200.ResourceId,
	}

	res1 := CreateTelemetryLogsProfile(t, ctx, apiClient, TelemetryLogsProfilePerInstance)
	res1.JSON200.LogsGroup = logsGroup.JSON200
	res2 := CreateTelemetryMetricsProfile(t, ctx, apiClient, TelemetryMetricsProfilePerSite)
	res2.JSON200.MetricsGroup = metricsGroup1.JSON200
	res3 := CreateTelemetryMetricsProfile(t, ctx, apiClient, TelemetryMetricsProfilePerRegion)
	res3.JSON200.MetricsGroup = metricsGroup2.JSON200

	// Assert presence of telemetry resources
	allLogsProfiles, err := apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, allLogsProfiles.StatusCode())
	assert.Len(t, allLogsProfiles.JSON200.TelemetryLogsProfiles, 1)

	allMetricsProfiles, err := apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, allMetricsProfiles.StatusCode())
	assert.Len(t, allMetricsProfiles.JSON200.TelemetryMetricsProfiles, 2)

	res, err := apiClient.TelemetryLogsProfileServiceGetTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, res1.JSON200.ProfileId, res.JSON200.ProfileId)
	assert.Equal(t, res1.JSON200.TargetInstance, res.JSON200.TargetInstance)
	assert.Equal(t, res1.JSON200.TargetSite, res.JSON200.TargetSite)
	assert.Equal(t, res1.JSON200.TargetRegion, res.JSON200.TargetRegion)
	assert.Equal(t, res1.JSON200.LogsGroupId, res.JSON200.LogsGroupId)
	assert.Equal(t, res1.JSON200.LogsGroup.TelemetryLogsGroupId, res.JSON200.LogsGroup.TelemetryLogsGroupId)
	assert.Equal(t, res1.JSON200.LogsGroup.Name, res.JSON200.LogsGroup.Name)
	assert.Equal(t, res1.JSON200.LogLevel, res.JSON200.LogLevel)

	for _, profile := range []*api.TelemetryMetricsProfileResource{res2.JSON200, res3.JSON200} {
		resp, err := apiClient.TelemetryMetricsProfileServiceGetTelemetryMetricsProfileWithResponse(
			ctx,
			*profile.ProfileId,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		assert.Equal(t, profile.ProfileId, resp.JSON200.ProfileId)
		assert.Equal(t, res1.JSON200.TargetInstance, res.JSON200.TargetInstance)
		assert.Equal(t, res1.JSON200.TargetSite, res.JSON200.TargetSite)
		assert.Equal(t, res1.JSON200.TargetRegion, res.JSON200.TargetRegion)
		assert.Equal(t, profile.MetricsGroupId, resp.JSON200.MetricsGroupId)
		assert.Equal(t, profile.MetricsGroup.TelemetryMetricsGroupId, resp.JSON200.MetricsGroup.TelemetryMetricsGroupId)
		assert.Equal(t, profile.MetricsGroup.Name, resp.JSON200.MetricsGroup.Name)
		assert.Equal(t, profile.MetricsInterval, resp.JSON200.MetricsInterval)
	}
}

func TestTelemetryLogsProfile_UpdatePUT(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	logsGroup1 := CreateTelemetryLogsGroup(t, ctx, apiClient, utils.TelemetryLogsGroup1Request)
	logsGroup2 := CreateTelemetryLogsGroup(t, ctx, apiClient, api.TelemetryLogsGroupResource{
		Name:          "Kernel logs",
		CollectorKind: collectorKindHostLogs,
		Groups: []string{
			"kern",
		},
	})

	regionCreated1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = nil
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	TelemetryLogsProfile := api.TelemetryLogsProfileResource{
		LogLevel:    api.SEVERITYLEVELDEBUG,
		TargetSite:  siteCreated1.JSON200.ResourceId,
		LogsGroupId: *logsGroup1.JSON200.ResourceId,
	}
	res1 := CreateTelemetryLogsProfile(t, ctx, apiClient, TelemetryLogsProfile)
	res1.JSON200.LogsGroup = logsGroup1.JSON200

	// Assert presence of the telemetry profile
	TelemetryProfile1Get, err := apiClient.TelemetryLogsProfileServiceGetTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, TelemetryLogsProfile.LogLevel, TelemetryProfile1Get.JSON200.LogLevel)

	// re-assign telemetry profile from Site to Region
	TelemetryLogsProfile.TargetSite = &emptyString
	TelemetryLogsProfile.TargetRegion = regionCreated1.JSON200.ResourceId
	telemetryLogsProfile1Update, err := apiClient.TelemetryLogsProfileServiceUpdateTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryLogsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, telemetryLogsProfile1Update.StatusCode())
	assert.Equal(t, *TelemetryLogsProfile.TargetRegion, *telemetryLogsProfile1Update.JSON200.TargetRegion)

	TelemetryProfile1Get, err = apiClient.TelemetryLogsProfileServiceGetTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, TelemetryLogsProfile.LogLevel, TelemetryProfile1Get.JSON200.LogLevel)
	assert.Equal(t, TelemetryLogsProfile.LogsGroupId, TelemetryProfile1Get.JSON200.LogsGroupId)
	assert.Empty(t, TelemetryProfile1Get.JSON200.TargetSite)
	assert.Equal(t, *regionCreated1.JSON200.ResourceId, *TelemetryProfile1Get.JSON200.TargetRegion)

	// change log level
	TelemetryLogsProfile.LogLevel = api.SEVERITYLEVELINFO
	telemetryLogsProfile1Update, err = apiClient.TelemetryLogsProfileServiceUpdateTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryLogsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, telemetryLogsProfile1Update.StatusCode())
	assert.Equal(t, api.SEVERITYLEVELINFO, telemetryLogsProfile1Update.JSON200.LogLevel)

	TelemetryProfile1Get, err = apiClient.TelemetryLogsProfileServiceGetTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, api.SEVERITYLEVELINFO, telemetryLogsProfile1Update.JSON200.LogLevel)

	// change the telemetry group
	TelemetryLogsProfile.LogsGroupId = *logsGroup2.JSON200.ResourceId
	telemetryLogsProfile1Update, err = apiClient.TelemetryLogsProfileServiceUpdateTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryLogsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, telemetryLogsProfile1Update.StatusCode())
	assert.Equal(t, *logsGroup2.JSON200.ResourceId, telemetryLogsProfile1Update.JSON200.LogsGroupId)

	TelemetryProfile1Get, err = apiClient.TelemetryLogsProfileServiceGetTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, *logsGroup2.JSON200.ResourceId, telemetryLogsProfile1Update.JSON200.LogsGroupId)

	// PUT with empty target relation
	TelemetryLogsProfile.TargetRegion = &emptyString
	telemetryLogsProfile1Update, err = apiClient.TelemetryLogsProfileServiceUpdateTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryLogsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, telemetryLogsProfile1Update.StatusCode())

	// update to wrong type of telemetry group (logs profile cannot be associated with metrics group)
	metricsGroup := CreateTelemetryMetricsGroup(t, ctx, apiClient, utils.TelemetryMetricsGroup1Request)
	TelemetryLogsProfile.TargetRegion = regionCreated1.JSON200.ResourceId
	TelemetryLogsProfile.LogsGroupId = *metricsGroup.JSON200.ResourceId
	telemetryLogsProfile1Update, err = apiClient.TelemetryLogsProfileServiceUpdateTelemetryLogsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryLogsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, telemetryLogsProfile1Update.StatusCode())
}

func TestTelemetryMetricsProfile_UpdatePUT(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	metricsGroup1 := CreateTelemetryMetricsGroup(t, ctx, apiClient, utils.TelemetryMetricsGroup1Request)
	metricsGroup2 := CreateTelemetryMetricsGroup(t, ctx, apiClient, api.TelemetryMetricsGroupResource{
		Name:          "NW Usage",
		CollectorKind: collectorKindHostMetrics,
		Groups: []string{
			"net",
		},
	})

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	regionCreated1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)

	TelemetryMetricsProfile := api.TelemetryMetricsProfileResource{
		MetricsInterval: 300,
		TargetSite:      siteCreated1.JSON200.ResourceId,
		MetricsGroupId:  *metricsGroup1.JSON200.ResourceId,
	}
	res1 := CreateTelemetryMetricsProfile(t, ctx, apiClient, TelemetryMetricsProfile)
	res1.JSON200.MetricsGroup = metricsGroup1.JSON200

	// Assert presence of the telemetry profile
	TelemetryProfile1Get, err := apiClient.TelemetryMetricsProfileServiceGetTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, TelemetryMetricsProfile.MetricsInterval, TelemetryProfile1Get.JSON200.MetricsInterval)

	// re-assign telemetry profile from Site to Region
	TelemetryMetricsProfile.TargetSite = &emptyString
	TelemetryMetricsProfile.TargetRegion = regionCreated1.JSON200.ResourceId
	telemetryMetricsProfile1Update, err := apiClient.TelemetryMetricsProfileServiceUpdateTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryMetricsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, telemetryMetricsProfile1Update.StatusCode())
	assert.Equal(t, *TelemetryMetricsProfile.TargetRegion, *telemetryMetricsProfile1Update.JSON200.TargetRegion)

	TelemetryProfile1Get, err = apiClient.TelemetryMetricsProfileServiceGetTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, TelemetryMetricsProfile.MetricsInterval, TelemetryProfile1Get.JSON200.MetricsInterval)
	assert.Equal(t, TelemetryMetricsProfile.MetricsGroupId, TelemetryProfile1Get.JSON200.MetricsGroupId)
	assert.Empty(t, TelemetryProfile1Get.JSON200.TargetSite)
	assert.Equal(t, *regionCreated1.JSON200.ResourceId, *TelemetryProfile1Get.JSON200.TargetRegion)

	// change log level
	TelemetryMetricsProfile.MetricsInterval = 5
	telemetryMetricsProfile1Update, err = apiClient.TelemetryMetricsProfileServiceUpdateTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryMetricsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, telemetryMetricsProfile1Update.StatusCode())
	assert.Equal(t, 5, int(telemetryMetricsProfile1Update.JSON200.MetricsInterval))

	TelemetryProfile1Get, err = apiClient.TelemetryMetricsProfileServiceGetTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, 5, int(telemetryMetricsProfile1Update.JSON200.MetricsInterval))

	// change the telemetry group
	TelemetryMetricsProfile.MetricsGroupId = *metricsGroup2.JSON200.ResourceId
	telemetryMetricsProfile1Update, err = apiClient.TelemetryMetricsProfileServiceUpdateTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryMetricsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, telemetryMetricsProfile1Update.StatusCode())
	assert.Equal(t, *metricsGroup2.JSON200.ResourceId, telemetryMetricsProfile1Update.JSON200.MetricsGroupId)

	TelemetryProfile1Get, err = apiClient.TelemetryMetricsProfileServiceGetTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, TelemetryProfile1Get.StatusCode())
	assert.Equal(t, *metricsGroup2.JSON200.ResourceId, telemetryMetricsProfile1Update.JSON200.MetricsGroupId)

	// PUT with empty target relation
	TelemetryMetricsProfile.TargetRegion = &emptyString
	telemetryMetricsProfile1Update, err = apiClient.TelemetryMetricsProfileServiceUpdateTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryMetricsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, telemetryMetricsProfile1Update.StatusCode())

	// update to wrong type of telemetry group (logs profile cannot be associated with metrics group)
	logsGroup := CreateTelemetryLogsGroup(t, ctx, apiClient, utils.TelemetryLogsGroup1Request)
	TelemetryMetricsProfile.TargetRegion = regionCreated1.JSON200.ResourceId
	TelemetryMetricsProfile.MetricsGroupId = *logsGroup.JSON200.ResourceId
	telemetryMetricsProfile1Update, err = apiClient.TelemetryMetricsProfileServiceUpdateTelemetryMetricsProfileWithResponse(
		ctx,
		*res1.JSON200.ProfileId,
		TelemetryMetricsProfile,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, telemetryMetricsProfile1Update.StatusCode())
}

func TestTelemetryGroupList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList1, err := apiClient.TelemetryLogsGroupServiceListTelemetryLogsGroupsWithResponse(
		ctx,
		&api.TelemetryLogsGroupServiceListTelemetryLogsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList1.StatusCode())
	assert.Empty(t, resList1.JSON200.TelemetryLogsGroups)

	resList2, err := apiClient.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsWithResponse(
		ctx,
		&api.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList2.StatusCode())
	assert.Empty(t, resList2.JSON200.TelemetryMetricsGroups)
}

func TestTelemetryProfileList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList1, err := apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList1.StatusCode())
	assert.Empty(t, resList1.JSON200.TelemetryLogsProfiles)

	resList2, err := apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList2.StatusCode())
	assert.Empty(t, resList2.JSON200.TelemetryMetricsProfiles)
}

func TestTelemetryLogsGroupList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	var offset int32 = 1
	var pageSize int32 = 4

	for id := 0; id < totalItems; id++ {
		CreateTelemetryLogsGroup(t, ctx, apiClient, api.TelemetryLogsGroupResource{
			CollectorKind: api.TelemetryLogsGroupResourceCollectorKindCOLLECTORKINDCLUSTER,
			Groups:        []string{"test"},
			Name:          "Test Name",
		})
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.TelemetryLogsGroupServiceListTelemetryLogsGroupsWithResponse(
		ctx,
		&api.TelemetryLogsGroupServiceListTelemetryLogsGroupsParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryLogsGroups), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsGroupServiceListTelemetryLogsGroupsWithResponse(
		ctx,
		&api.TelemetryLogsGroupServiceListTelemetryLogsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.TelemetryLogsGroups))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsGroupList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	var offset int32 = 1
	var pageSize int32 = 4

	for id := 0; id < totalItems; id++ {
		CreateTelemetryMetricsGroup(t, ctx, apiClient, api.TelemetryMetricsGroupResource{
			CollectorKind: api.TelemetryMetricsGroupResourceCollectorKindCOLLECTORKINDCLUSTER,
			Groups:        []string{"test"},
			Name:          "Test Name",
		})
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsWithResponse(
		ctx,
		&api.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryMetricsGroups), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsWithResponse(
		ctx,
		&api.TelemetryMetricsGroupServiceListTelemetryMetricsGroupsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.TelemetryMetricsGroups))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryLogsProfileList(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	var offset int32 = 1
	var pageSize int32 = 4

	group := CreateTelemetryLogsGroup(t, ctx, apiClient, api.TelemetryLogsGroupResource{
		CollectorKind: collectorKindHostLogs,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	region1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = region1.JSON200.ResourceId
	site1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = site1.JSON200.ResourceId
	host := CreateHost(t, ctx, apiClient, utils.Host1Request)
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = host.JSON200.ResourceId
	instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	for id := 0; id < totalItems; id++ {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId:    *group.JSON200.ResourceId,
			LogLevel:       api.SEVERITYLEVELWARN,
			TargetInstance: instance.JSON200.ResourceId,
		})
	}

	for id := 0; id < totalItems; id++ {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId: *group.JSON200.ResourceId,
			LogLevel:    api.SEVERITYLEVELWARN,
			TargetSite:  site1.JSON200.ResourceId,
		})
	}

	for id := 0; id < totalItems; id++ {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId:  *group.JSON200.ResourceId,
			LogLevel:     api.SEVERITYLEVELWARN,
			TargetRegion: region1.JSON200.ResourceId,
		})
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryLogsProfiles), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	var allPageSize int32 = 30
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems*3, len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// check filters
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			InstanceId: instance.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryLogsProfiles), totalItems)
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			SiteId: site1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryLogsProfiles), totalItems)
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			RegionId: region1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryLogsProfiles), totalItems)
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsProfileList(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	var offset int32 = 1
	var pageSize int32 = 4

	group := CreateTelemetryMetricsGroup(t, ctx, apiClient, api.TelemetryMetricsGroupResource{
		CollectorKind: collectorKindHostMetrics,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	region1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = region1.JSON200.ResourceId
	site1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = site1.JSON200.ResourceId
	hostUUID := uuid.New().String()
	host := CreateHost(t, ctx, apiClient, api.HostResource{
		Name:     utils.Host1Request.Name,
		Metadata: utils.Host1Request.Metadata,
		Uuid:     &hostUUID,
	})

	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = host.JSON200.ResourceId
	instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	for id := 0; id < totalItems; id++ {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetInstance:  instance.JSON200.ResourceId,
		})
	}

	for id := 0; id < totalItems; id++ {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetSite:      site1.JSON200.ResourceId,
		})
	}

	for id := 0; id < totalItems; id++ {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetRegion:    region1.JSON200.ResourceId,
		})
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryMetricsProfiles), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	var allPageSize int32 = 30
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems*3, len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// check filters
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			InstanceId: instance.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryMetricsProfiles), totalItems)
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			SiteId: site1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryMetricsProfiles), totalItems)
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			RegionId: region1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryMetricsProfiles), totalItems)
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsProfileListInherited(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	group := CreateTelemetryMetricsGroup(t, ctx, apiClient, api.TelemetryMetricsGroupResource{
		CollectorKind: collectorKindHostMetrics,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	region1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = region1.JSON200.ResourceId
	site1Region1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Site2Request.RegionId = region1.JSON200.ResourceId
	site2Region1 := CreateSite(t, ctx, apiClient, utils.Site2Request)
	parentRegion2Name := "Parent Region 2"
	parentRegion2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &parentRegion2Name,
	})
	region2Name := "Region 2"
	region2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &region2Name,
		ParentId: parentRegion2.JSON200.ResourceId,
	})
	site1Region2Name := "Site 1 Region 2"
	site1Region2 := CreateSite(t, ctx, apiClient, api.SiteResource{
		Name:     &site1Region2Name,
		RegionId: region2.JSON200.ResourceId,
	})
	// 3 Instances in Site 1 of Region 1
	site1Region1Instances := make([]*api.InstanceResource, 0)
	kindMetal := api.INSTANCEKINDMETAL
	for i := 0; i < 3; i++ {
		hostUuid := uuid.New().String()
		name := fmt.Sprintf("Host %d S1R1", i)
		host := CreateHost(t, ctx, apiClient, api.HostResource{
			Name:   name,
			SiteId: site1Region1.JSON200.ResourceId,
			Uuid:   &hostUuid,
		})
		instName := fmt.Sprintf("Site 1 Region 1 - Instance %d", i)
		inst := CreateInstance(t, ctx, apiClient, api.InstanceResource{
			HostID: host.JSON200.ResourceId,
			OsID:   os.JSON200.ResourceId,
			Kind:   &kindMetal,
			Name:   &instName,
		})
		site1Region1Instances = append(site1Region1Instances, inst.JSON200)
	}

	// 3 Instances in Site 2 of Region 1
	site2Region1Instances := make([]*api.InstanceResource, 0)
	for i := 0; i < 3; i++ {
		hostUuid := uuid.New().String()
		host := CreateHost(t, ctx, apiClient, api.HostResource{
			Name:   fmt.Sprintf("Host %d S2R1", i),
			SiteId: site2Region1.JSON200.ResourceId,
			Uuid:   &hostUuid,
		})
		instName := fmt.Sprintf("Site 2 Region 1 - Instance %d", i)
		inst := CreateInstance(t, ctx, apiClient, api.InstanceResource{
			HostID: host.JSON200.ResourceId,
			OsID:   os.JSON200.ResourceId,
			Kind:   &kindMetal,
			Name:   &instName,
		})
		site2Region1Instances = append(site2Region1Instances, inst.JSON200)
	}

	// 1 Instance in Site 1 of Region 2
	site1Region2Instances := make([]*api.InstanceResource, 0)
	for i := 0; i < 1; i++ {
		hostUuid := uuid.New().String()
		host := CreateHost(t, ctx, apiClient, api.HostResource{
			Name:   fmt.Sprintf("Host %d S1R2", i),
			SiteId: site1Region2.JSON200.ResourceId,
			Uuid:   &hostUuid,
		})
		instName := fmt.Sprintf("Site 1 Region 2 - Instance %d", i)
		inst := CreateInstance(t, ctx, apiClient, api.InstanceResource{
			HostID: host.JSON200.ResourceId,
			OsID:   os.JSON200.ResourceId,
			Kind:   &kindMetal,
			Name:   &instName,
		})
		site1Region2Instances = append(site1Region2Instances, inst.JSON200)
	}

	// Region 1 - 3 Telemetry Metrics Profiles
	for id := 0; id < 3; id++ {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetRegion:    region1.JSON200.ResourceId,
		})
	}

	// Region 2 - 1 Telemetry Metrics Profile
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetRegion:    region2.JSON200.ResourceId,
	})

	// Parent Region 2 - 2 Telemetry Metrics Profiles
	for id := 0; id < 2; id++ {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetRegion:    parentRegion2.JSON200.ResourceId,
		})
	}

	// Site 1 Region 1 - no Telemetry Metrics Profile

	// Site 2 Region 1 - 2 Telemetry Metrics Profiles
	for id := 0; id < 2; id++ {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetSite:      site2Region1.JSON200.ResourceId,
		})
	}

	// Site 1 Region 2 - 1 Telemetry Metrics Profile
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetSite:      site1Region2.JSON200.ResourceId,
	})

	// Site 1 Region 1 - 1 Telemetry Profile per Instance
	for _, inst := range site1Region1Instances {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetInstance:  inst.ResourceId,
		})
	}

	// Site 2 Region 1 - No Telemetry Profiles for any Instance

	// Site 1 Region 2 - 1 Telemetry Profile per Instance
	for _, inst := range site1Region2Instances {
		CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
			MetricsGroupId:  *group.JSON200.ResourceId,
			MetricsInterval: 300,
			TargetInstance:  inst.ResourceId,
		})
	}

	var offset int32 = 1
	var pageSize int32 = 4

	// list all telemetry profiles (no filtering)
	resList, err := apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryMetricsProfiles), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	var allPageSize int32 = 100
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 13, len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	showInherited := true
	// render for Instances in Site 1 Region 1
	for _, inst := range site1Region1Instances {
		resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
			ctx,
			&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
				InstanceId:    inst.ResourceId,
				ShowInherited: &showInherited,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 4, // 1 for Instance + 0 for Site + 3 for Region 1 (no parent regions)
			len(resList.JSON200.TelemetryMetricsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)

		// no inheritence
		resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
			ctx,
			&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
				InstanceId: inst.ResourceId,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())

		assert.Equal(t, 1, // 1 for Instance
			len(resList.JSON200.TelemetryMetricsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)
	}

	// render for Instances in Site 2 Region 1
	for _, inst := range site2Region1Instances {
		resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
			ctx,
			&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
				InstanceId:    inst.ResourceId,
				ShowInherited: &showInherited,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		expectedItems := 5 // 0 for Instance + 2 for Site + 3 for Region (no parent regions)
		assert.Equal(t, expectedItems, len(resList.JSON200.TelemetryMetricsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)

		// no inheritence
		resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
			ctx,
			&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
				InstanceId: inst.ResourceId,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 0, // 0 for Instance
			len(resList.JSON200.TelemetryMetricsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)
	}

	// render for Instances in Site 1 Region 2
	for _, inst := range site1Region2Instances {
		resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
			ctx,
			&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
				InstanceId:    inst.ResourceId,
				ShowInherited: &showInherited,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 5, // 1 for Instance + 1 for Site + 1 for Region + 2 from Parent Region 2
			len(resList.JSON200.TelemetryMetricsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)

		// no inheritence
		resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
			ctx,
			&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
				InstanceId: inst.ResourceId,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 1, // 1 for Instance
			len(resList.JSON200.TelemetryMetricsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)
	}

	// render for Site 1 Region 1
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			SiteId:        site1Region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, // 0 for Site + 3 for Region 1 (no parent regions)
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Site 2 Region 1
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			SiteId:        site2Region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 5, // 2 for Site + 3 for Region 1 (no parent regions)
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Site 1 Region 2
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			SiteId:        site1Region2.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, // 1 for Site + 1 for Region 2 + 2 for parent region
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Region 1
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			RegionId:      region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, //  3 for Region 1 (no parent regions)
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Region 2
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			RegionId:      region2.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, //  1 for Region 2 + 2 for parent region
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsProfileListInheritedNestingLimit(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	group := CreateTelemetryMetricsGroup(t, ctx, apiClient, api.TelemetryMetricsGroupResource{
		CollectorKind: collectorKindHostMetrics,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	regionLevel5Name := "Region 5"
	regionLevel5 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &regionLevel5Name,
	})

	regionLevel4Name := "Region 4"
	regionLevel4 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel4Name,
		ParentId: regionLevel5.JSON200.ResourceId,
	})

	regionLevel3Name := "Region 3"
	regionLevel3 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel3Name,
		ParentId: regionLevel4.JSON200.ResourceId,
	})

	regionLevel2Name := "Region 2"
	regionLevel2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel2Name,
		ParentId: regionLevel3.JSON200.ResourceId,
	})

	regionLevel1Name := "Region 1"
	regionLevel1 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel1Name,
		ParentId: regionLevel2.JSON200.ResourceId,
	})

	utils.Site1Request.RegionId = regionLevel1.JSON200.ResourceId
	site := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = site.JSON200.ResourceId
	host := CreateHost(t, ctx, apiClient, utils.Host1Request)

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = host.JSON200.ResourceId
	instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	// profile per instance
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetInstance:  instance.JSON200.ResourceId,
	})
	// profile per site
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetSite:      site.JSON200.ResourceId,
	})
	// profile per region level 1
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetRegion:    regionLevel1.JSON200.ResourceId,
	})
	// profile per region level 3
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetRegion:    regionLevel3.JSON200.ResourceId,
	})
	// profile per region level 5
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetRegion:    regionLevel5.JSON200.ResourceId,
	})

	var allPageSize int32 = 100
	resList, err := apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 5, len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	showInherited := true
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			InstanceId:    instance.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 5, // 1 for Instance + 1 for Site + 1 for Region Level 1 + 1 for Region Level 3 + 1 for Region Level 5
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			SiteId:        site.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, // 1 for Site + 1 for Region Level 1 + 1 for Region Level 3 + 1 for Region Level 5
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			RegionId:      regionLevel1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, // 1 for Region Level 1 + 1 for Region Level 3 + 1 for Region Level 5
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			RegionId:      regionLevel4.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Region Level 5
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsProfileListInheritedNoParents(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	group := CreateTelemetryMetricsGroup(t, ctx, apiClient, api.TelemetryMetricsGroupResource{
		CollectorKind: collectorKindHostMetrics,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	region2Name := "Region 2"
	region2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &region2Name,
	})

	region1Name := "Region 1"
	region1 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &region1Name,
	})

	utils.Site1Request.RegionId = nil
	site := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = nil
	host := CreateHost(t, ctx, apiClient, utils.Host1Request)

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = host.JSON200.ResourceId
	instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	// profile per instance
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetInstance:  instance.JSON200.ResourceId,
	})
	// profile per site
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetSite:      site.JSON200.ResourceId,
	})
	// profile per region 1
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetRegion:    region1.JSON200.ResourceId,
	})
	// profile per region 2
	CreateTelemetryMetricsProfile(t, ctx, apiClient, api.TelemetryMetricsProfileResource{
		MetricsGroupId:  *group.JSON200.ResourceId,
		MetricsInterval: 300,
		TargetRegion:    region2.JSON200.ResourceId,
	})

	var allPageSize int32 = 100
	resList, err := apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	showInherited := true
	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			InstanceId:    instance.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Instance, no parent relations
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			SiteId:        site.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Site, no parent relations
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesWithResponse(
		ctx,
		&api.TelemetryMetricsProfileServiceListTelemetryMetricsProfilesParams{
			RegionId:      region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Region, no parents
		len(resList.JSON200.TelemetryMetricsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryLogsProfileListInherited(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	group := CreateTelemetryLogsGroup(t, ctx, apiClient, api.TelemetryLogsGroupResource{
		CollectorKind: collectorKindHostLogs,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	region1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = region1.JSON200.ResourceId
	site1Region1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Site2Request.RegionId = region1.JSON200.ResourceId
	site2Region1 := CreateSite(t, ctx, apiClient, utils.Site2Request)
	parentRegion2Name := "Parent Region 2"
	parentRegion2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &parentRegion2Name,
	})
	region2Name := "Region 2"
	region2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &region2Name,
		ParentId: parentRegion2.JSON200.ResourceId,
	})
	site1Region2Name := "Site 1 Region 2"
	site1Region2 := CreateSite(t, ctx, apiClient, api.SiteResource{
		Name:     &site1Region2Name,
		RegionId: region2.JSON200.ResourceId,
	})
	// 3 Instances in Site 1 of Region 1
	site1Region1Instances := make([]*api.InstanceResource, 0)
	kindMetal := api.INSTANCEKINDMETAL
	for i := 0; i < 3; i++ {
		hostUuid := uuid.New().String()
		host := CreateHost(t, ctx, apiClient, api.HostResource{
			Name:   fmt.Sprintf("Host %d S1R1", i),
			SiteId: site1Region1.JSON200.ResourceId,
			Uuid:   &hostUuid,
		})
		instName := fmt.Sprintf("Site 1 Region 1 - Instance %d", i)
		inst := CreateInstance(t, ctx, apiClient, api.InstanceResource{
			HostID: host.JSON200.ResourceId,
			OsID:   os.JSON200.ResourceId,
			Kind:   &kindMetal,
			Name:   &instName,
		})
		site1Region1Instances = append(site1Region1Instances, inst.JSON200)
	}

	// 3 Instances in Site 2 of Region 1
	site2Region1Instances := make([]*api.InstanceResource, 0)
	for i := 0; i < 3; i++ {
		hostUuid := uuid.New().String()
		host := CreateHost(t, ctx, apiClient, api.HostResource{
			Name:   fmt.Sprintf("Host %d S2R1", i),
			SiteId: site2Region1.JSON200.ResourceId,
			Uuid:   &hostUuid,
		})
		instName := fmt.Sprintf("Site 2 Region 1 - Instance %d", i)
		inst := CreateInstance(t, ctx, apiClient, api.InstanceResource{
			HostID: host.JSON200.ResourceId,
			OsID:   os.JSON200.ResourceId,
			Kind:   &kindMetal,
			Name:   &instName,
		})
		site2Region1Instances = append(site2Region1Instances, inst.JSON200)
	}

	// 1 Instance in Site 1 of Region 2
	site1Region2Instances := make([]*api.InstanceResource, 0)
	for i := 0; i < 1; i++ {
		hostUuid := uuid.New().String()
		host := CreateHost(t, ctx, apiClient, api.HostResource{
			Name:   fmt.Sprintf("Host %d S1R2", i),
			SiteId: site1Region2.JSON200.ResourceId,
			Uuid:   &hostUuid,
		})
		instName := fmt.Sprintf("Site 1 Region 2 - Instance %d", i)
		inst := CreateInstance(t, ctx, apiClient, api.InstanceResource{
			HostID: host.JSON200.ResourceId,
			OsID:   os.JSON200.ResourceId,
			Kind:   &kindMetal,
			Name:   &instName,
		})
		site1Region2Instances = append(site1Region2Instances, inst.JSON200)
	}

	// Region 1 - 3 Telemetry Logs Profiles
	for id := 0; id < 3; id++ {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId:  *group.JSON200.ResourceId,
			LogLevel:     api.SEVERITYLEVELWARN,
			TargetRegion: region1.JSON200.ResourceId,
		})
	}

	// Region 2 - 1 Telemetry Logs Profile
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:  *group.JSON200.ResourceId,
		LogLevel:     api.SEVERITYLEVELWARN,
		TargetRegion: region2.JSON200.ResourceId,
	})

	// Parent Region 2 - 2 Telemetry Logs Profiles
	for id := 0; id < 2; id++ {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId:  *group.JSON200.ResourceId,
			LogLevel:     api.SEVERITYLEVELWARN,
			TargetRegion: parentRegion2.JSON200.ResourceId,
		})
	}

	// Site 1 Region 1 - no Telemetry Logs Profile

	// Site 2 Region 1 - 2 Telemetry Logs Profiles
	for id := 0; id < 2; id++ {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId: *group.JSON200.ResourceId,
			LogLevel:    api.SEVERITYLEVELWARN,
			TargetSite:  site2Region1.JSON200.ResourceId,
		})
	}

	// Site 1 Region 2 - 1 Telemetry Logs Profile
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId: *group.JSON200.ResourceId,
		LogLevel:    api.SEVERITYLEVELWARN,
		TargetSite:  site1Region2.JSON200.ResourceId,
	})

	// Site 1 Region 1 - 1 Telemetry Profile per Instance
	for _, inst := range site1Region1Instances {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId:    *group.JSON200.ResourceId,
			LogLevel:       api.SEVERITYLEVELWARN,
			TargetInstance: inst.ResourceId,
		})
	}

	// Site 2 Region 1 - No Telemetry Profiles for any Instance

	// Site 1 Region 2 - 1 Telemetry Profile per Instance
	for _, inst := range site1Region2Instances {
		CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
			LogsGroupId:    *group.JSON200.ResourceId,
			LogLevel:       api.SEVERITYLEVELWARN,
			TargetInstance: inst.ResourceId,
		})
	}

	var offset int32 = 1
	var pageSize int32 = 4

	// list all telemetry profiles (no filtering)
	resList, err := apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.TelemetryLogsProfiles), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	var allPageSize int32 = 100
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 13, len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	showInherited := true
	// render for Instances in Site 1 Region 1
	for _, inst := range site1Region1Instances {
		resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
			ctx,
			&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
				InstanceId:    inst.ResourceId,
				ShowInherited: &showInherited,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 4, // 1 for Instance + 0 for Site + 3 for Region 1 (no parent regions)
			len(resList.JSON200.TelemetryLogsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)

		// no inheritence
		resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
			ctx,
			&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
				InstanceId: inst.ResourceId,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())

		assert.Equal(t, 1, // 1 for Instance
			len(resList.JSON200.TelemetryLogsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)
	}

	// render for Instances in Site 2 Region 1
	for _, inst := range site2Region1Instances {
		resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
			ctx,
			&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
				InstanceId:    inst.ResourceId,
				ShowInherited: &showInherited,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		expectedItems := 5 // 0 for Instance + 2 for Site + 3 for Region (no parent regions)
		assert.Equal(t, expectedItems, len(resList.JSON200.TelemetryLogsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)

		// no inheritence
		resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
			ctx,
			&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
				InstanceId: inst.ResourceId,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 0, // 0 for Instance
			len(resList.JSON200.TelemetryLogsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)
	}

	// render for Instances in Site 1 Region 2
	for _, inst := range site1Region2Instances {
		resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
			ctx,
			&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
				InstanceId:    inst.ResourceId,
				ShowInherited: &showInherited,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 5, // 1 for Instance + 1 for Site + 1 for Region + 2 from Parent Region 2
			len(resList.JSON200.TelemetryLogsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)

		// no inheritence
		resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
			ctx,
			&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
				InstanceId: inst.ResourceId,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resList.StatusCode())
		assert.Equal(t, 1, // 1 for Instance
			len(resList.JSON200.TelemetryLogsProfiles))
		assert.Equal(t, false, resList.JSON200.HasNext)
	}

	// render for Site 1 Region 1
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			SiteId:        site1Region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, // 0 for Site + 3 for Region 1 (no parent regions)
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Site 2 Region 1
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			SiteId:        site2Region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 5, // 2 for Site + 3 for Region 1 (no parent regions)
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Site 1 Region 2
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			SiteId:        site1Region2.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, // 1 for Site + 1 for Region 2 + 2 for parent region
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Region 1
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			RegionId:      region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, //  3 for Region 1 (no parent regions)
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// render for Region 2
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			RegionId:      region2.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, //  1 for Region 2 + 2 for parent region
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsLogsListInheritedNestingLimit(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	group := CreateTelemetryLogsGroup(t, ctx, apiClient, api.TelemetryLogsGroupResource{
		CollectorKind: collectorKindHostLogs,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	regionLevel5Name := "Region 5"
	regionLevel5 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &regionLevel5Name,
	})

	regionLevel4Name := "Region 4"
	regionLevel4 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel4Name,
		ParentId: regionLevel5.JSON200.ResourceId,
	})

	regionLevel3Name := "Region 3"
	regionLevel3 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel3Name,
		ParentId: regionLevel4.JSON200.ResourceId,
	})

	regionLevel2Name := "Region 2"
	regionLevel2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel2Name,
		ParentId: regionLevel3.JSON200.ResourceId,
	})

	regionLevel1Name := "Region 1"
	regionLevel1 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name:     &regionLevel1Name,
		ParentId: regionLevel2.JSON200.ResourceId,
	})

	utils.Site1Request.RegionId = regionLevel1.JSON200.ResourceId
	site := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = site.JSON200.ResourceId
	host := CreateHost(t, ctx, apiClient, utils.Host1Request)

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = host.JSON200.ResourceId
	instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	// profile per instance
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:    *group.JSON200.ResourceId,
		LogLevel:       api.SEVERITYLEVELWARN,
		TargetInstance: instance.JSON200.ResourceId,
	})
	// profile per site
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId: *group.JSON200.ResourceId,
		LogLevel:    api.SEVERITYLEVELWARN,
		TargetSite:  site.JSON200.ResourceId,
	})
	// profile per region level 1
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:  *group.JSON200.ResourceId,
		LogLevel:     api.SEVERITYLEVELWARN,
		TargetRegion: regionLevel1.JSON200.ResourceId,
	})
	// profile per region level 3
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:  *group.JSON200.ResourceId,
		LogLevel:     api.SEVERITYLEVELWARN,
		TargetRegion: regionLevel3.JSON200.ResourceId,
	})
	// profile per region level 5
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:  *group.JSON200.ResourceId,
		LogLevel:     api.SEVERITYLEVELWARN,
		TargetRegion: regionLevel5.JSON200.ResourceId,
	})

	var allPageSize int32 = 100
	resList, err := apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 5, len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	showInherited := true
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			InstanceId:    instance.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 5, // 1 for Instance + 1 for Site + 1 for Region Level 1 + 1 for Region Level 3 + 1 for Region Level 5
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			SiteId:        site.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, // 1 for Site + 1 for Region Level 1 + 1 for Region Level 3 + 1 for Region Level 5
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			RegionId:      regionLevel1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, // 1 for Region Level 1 + 1 for Region Level 3 + 1 for Region Level 5
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			RegionId:      regionLevel4.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Region Level 5
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryLogsProfileListInheritedNoParents(t *testing.T) {
	defer clearIDs()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	group := CreateTelemetryLogsGroup(t, ctx, apiClient, api.TelemetryLogsGroupResource{
		CollectorKind: collectorKindHostLogs,
		Groups:        []string{"test"},
		Name:          "Test Name",
	})
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	region2Name := "Region 2"
	region2 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &region2Name,
	})

	region1Name := "Region 1"
	region1 := CreateRegion(t, ctx, apiClient, api.RegionResource{
		Name: &region1Name,
	})

	utils.Site1Request.RegionId = nil
	site := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = nil
	host := CreateHost(t, ctx, apiClient, utils.Host1Request)

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = host.JSON200.ResourceId
	instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

	// profile per instance
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:    *group.JSON200.ResourceId,
		LogLevel:       api.SEVERITYLEVELWARN,
		TargetInstance: instance.JSON200.ResourceId,
	})
	// profile per site
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId: *group.JSON200.ResourceId,
		LogLevel:    api.SEVERITYLEVELWARN,
		TargetSite:  site.JSON200.ResourceId,
	})
	// profile per region 1
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:  *group.JSON200.ResourceId,
		LogLevel:     api.SEVERITYLEVELWARN,
		TargetRegion: region1.JSON200.ResourceId,
	})
	// profile per region 2
	CreateTelemetryLogsProfile(t, ctx, apiClient, api.TelemetryLogsProfileResource{
		LogsGroupId:  *group.JSON200.ResourceId,
		LogLevel:     api.SEVERITYLEVELWARN,
		TargetRegion: region2.JSON200.ResourceId,
	})

	var allPageSize int32 = 100
	resList, err := apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			PageSize: &allPageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	showInherited := true
	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			InstanceId:    instance.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Instance, no parent relations
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			SiteId:        site.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Site, no parent relations
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.TelemetryLogsProfileServiceListTelemetryLogsProfilesWithResponse(
		ctx,
		&api.TelemetryLogsProfileServiceListTelemetryLogsProfilesParams{
			RegionId:      region1.JSON200.ResourceId,
			ShowInherited: &showInherited,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, // 1 for Region, no parents
		len(resList.JSON200.TelemetryLogsProfiles))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestTelemetryMetricsProfile_Patch(t *testing.T) {
	log.Info().Msgf("Begin TelemetryMetricsProfile Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	metricsGroup1 := CreateTelemetryMetricsGroup(t, ctx, apiClient, utils.TelemetryMetricsGroup1Request)
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	telemetryMetricsProfile := api.TelemetryMetricsProfileResource{
		MetricsInterval: 300,
		TargetSite:      siteCreated1.JSON200.ResourceId,
		MetricsGroupId:  *metricsGroup1.JSON200.ResourceId,
	}
	metricsProfile := CreateTelemetryMetricsProfile(t, ctx, apiClient, telemetryMetricsProfile)
	assert.Equal(t, 100, metricsProfile.JSON200.MetricsInterval)

	// Modify fields for patching
	var newMetricsInterval int32 = 200
	patchRequest := api.TelemetryMetricsProfileResource{
		MetricsInterval: newMetricsInterval,
	}

	// Perform the Patch operation
	updatedMetricsGroup, err := apiClient.TelemetryMetricsProfileServicePatchTelemetryMetricsProfileWithResponse(
		ctx,
		*metricsProfile.JSON200.ResourceId,
		nil,
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedMetricsGroup.StatusCode())
	assert.Equal(t, newMetricsInterval, updatedMetricsGroup.JSON200.MetricsInterval)

	// Verify the changes with a Get operation
	getMetricsGroup, err := apiClient.TelemetryMetricsProfileServiceGetTelemetryMetricsProfileWithResponse(
		ctx,
		*metricsProfile.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getMetricsGroup.StatusCode())
	assert.Equal(t, newMetricsInterval, getMetricsGroup.JSON200.MetricsInterval)

	log.Info().Msgf("End TelemetryMetricsProfile Patch tests")
}

func TestTelemetryLogsProfile_Patch(t *testing.T) {
	log.Info().Msgf("Begin TelemetryLogsProfile Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	metricsGroup1 := CreateTelemetryLogsGroup(t, ctx, apiClient, utils.TelemetryLogsGroup1Request)
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	TelemetryLogsProfile := api.TelemetryLogsProfileResource{
		LogLevel:    api.SEVERITYLEVELDEBUG,
		TargetSite:  siteCreated1.JSON200.ResourceId,
		LogsGroupId: *metricsGroup1.JSON200.ResourceId,
	}

	logsProfile := CreateTelemetryLogsProfile(t, ctx, apiClient, TelemetryLogsProfile)
	assert.Equal(t, api.SEVERITYLEVELWARN, logsProfile.JSON200.LogLevel)

	// Modify fields for patching
	newLogLevel := api.SEVERITYLEVELCRITICAL
	patchRequest := api.TelemetryLogsProfileResource{
		LogLevel: newLogLevel,
	}

	// Perform the Patch operation
	updatedLogsGroup, err := apiClient.TelemetryLogsProfileServicePatchTelemetryLogsProfileWithResponse(
		ctx,
		*logsProfile.JSON200.ResourceId,
		nil,
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedLogsGroup.StatusCode())
	assert.Equal(t, newLogLevel, updatedLogsGroup.JSON200.LogLevel)

	// Verify the changes with a Get operation
	getLogsGroup, err := apiClient.TelemetryLogsProfileServiceGetTelemetryLogsProfileWithResponse(
		ctx,
		*logsProfile.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getLogsGroup.StatusCode())
	assert.Equal(t, newLogLevel, getLogsGroup.JSON200.LogLevel)

	log.Info().Msgf("End TelemetryLogsProfile Patch tests")
}
