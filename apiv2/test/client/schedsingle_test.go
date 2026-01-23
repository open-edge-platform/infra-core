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

func TestSchedSingle_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin SingleSched tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	siteCreated1 := CreateSite(ctx, t, apiClient, utils.Site1Request)

	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	singleSched1 := CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule1Request)

	utils.SingleSchedule2Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	singleSched2 := CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule2Request)

	get1, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.SschedName1, *get1.JSON200.Name)

	get2, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, utils.SschedName2, *get2.JSON200.Name)
	log.Info().Msgf("End SingleSchedule tests")
}

func TestSchedSingle_CreateError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	siteCreated1 := CreateSite(ctx, t, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = siteCreated1.JSON200.ResourceId
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host1Request)

	// Expected BadRequest Error because of target site and host are set in Schedule
	utils.SingleScheduleError.TargetSiteId = siteCreated1.JSON200.ResourceId
	utils.SingleScheduleError.TargetHostId = hostCreated1.JSON200.ResourceId

	sched, err := apiClient.ScheduleServiceCreateSingleScheduleWithResponse(
		ctx,
		projectName,
		utils.SingleScheduleError,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, sched.StatusCode())

	utils.SingleScheduleErrorSeconds.TargetSiteId = siteCreated1.JSON200.ResourceId
	sched, err = apiClient.ScheduleServiceCreateSingleScheduleWithResponse(
		ctx,
		projectName,
		utils.SingleScheduleErrorSeconds,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, sched.StatusCode())
}

func TestSchedSingle_UpdatePut(t *testing.T) {
	log.Info().Msgf("Begin SingleSchedule Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	siteCreated1 := CreateSite(ctx, t, apiClient, utils.Site1Request)

	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	singleSched1 := CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule1Request)

	SingleSchedule1Get, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1Get.StatusCode())
	assert.Equal(t, utils.SschedName1, *SingleSchedule1Get.JSON200.Name)

	siteCreated2 := CreateSite(ctx, t, apiClient, utils.SiteListRequest1)

	utils.SingleSchedule2Request.TargetSiteId = siteCreated2.JSON200.ResourceId
	singleSched1Update, err := apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, singleSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *singleSched1Update.JSON200.Name)

	SingleSchedule1GetUp, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1GetUp.StatusCode())
	assert.Equal(t, utils.SschedName2, *SingleSchedule1GetUp.JSON200.Name)
	assert.Equal(t, *siteCreated2.JSON200.ResourceId, *SingleSchedule1GetUp.JSON200.TargetSite.ResourceId)
	assert.Equal(t, utils.SingleSchedule2Request.ScheduleStatus, SingleSchedule1GetUp.JSON200.ScheduleStatus)

	// Uses PUT to set empty string to TargetSite and verifies it
	utils.SingleSchedule2Request.TargetSiteId = &emptyString
	singleSched1Update, err = apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, singleSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *singleSched1Update.JSON200.Name)

	SingleSchedule1GetUp, err = apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1GetUp.StatusCode())
	assert.Equal(t, utils.SschedName2, *SingleSchedule1GetUp.JSON200.Name)
	assert.Empty(t, SingleSchedule1GetUp.JSON200.TargetSite)
	assert.Equal(t, utils.SingleSchedule2Request.ScheduleStatus, SingleSchedule1GetUp.JSON200.ScheduleStatus)

	// Uses PUT to set wrong empty string to TargetSite and verifies its BadRequest error
	utils.SingleSchedule2Request.TargetSiteId = &emptyStringWrong
	singleSched1Update, err = apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.ResourceId,
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, singleSched1Update.StatusCode())

	log.Info().Msgf("End SingleSchedule Update tests")
}

func TestSchedSingle_Errors(t *testing.T) {
	log.Info().Msgf("Begin SingleSchedule Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	siteCreated1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		singleSched1Up, err := apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
			ctx,
			projectName,
			utils.SingleScheduleUnexistID,
			utils.SingleSchedule1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, singleSched1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
			ctx,
			projectName,
			utils.SingleScheduleUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.ScheduleServiceDeleteSingleScheduleWithResponse(
			ctx,
			projectName,
			utils.SingleScheduleUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusNotFound", func(t *testing.T) {
		singleSched1Up, err := apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
			ctx,
			projectName,
			utils.SingleScheduleWrongID,
			utils.SingleSchedule1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, singleSched1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusNotFound", func(t *testing.T) {
		s1res, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
			ctx,
			projectName,
			utils.SingleScheduleWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelSite, err := apiClient.ScheduleServiceDeleteSingleScheduleWithResponse(
			ctx,
			projectName,
			utils.SingleScheduleWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	log.Info().Msgf("End SingleSchedule Error tests")
}

func TestSchedSingleList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	siteCreated1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId

	totalItems := 10
	pageID := 1
	pageSize := 4

	for id := 0; id < totalItems; id++ {
		CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{
			Offset:   &pageID,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 4, len(resList.JSON200.SingleSchedules))
	assert.Equal(t, true, resList.JSON200.HasNext)

	// Checks if list resources return expected number of entries
	resList, err = apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{
			Offset:   &pageID,
			PageSize: &pageSize,
			SiteId:   siteCreated1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, pageSize, len(resList.JSON200.SingleSchedules))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{
			SiteId: siteCreated1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.SingleSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestSchedSingleListQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	postRespSite1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	postRespSite2 := CreateSite(ctx, t, apiClient, utils.Site2Request)

	utils.SingleSchedule1Request.TargetSiteId = postRespSite1.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule1Request)

	utils.SingleSchedule2Request.TargetSiteId = postRespSite2.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule2Request)

	utils.SingleSchedule3Request.TargetSiteId = postRespSite2.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule3Request)

	// Checks list of SingleSchedules with noo siteID
	resList, err := apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, len(resList.JSON200.SingleSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Checks list of SingleSchedules with siteID 1
	resList, err = apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{
			SiteId: postRespSite1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, len(resList.JSON200.SingleSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Checks list of SingleSchedules with SiteId 2
	resList, err = apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{
			SiteId: postRespSite2.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 2, len(resList.JSON200.SingleSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestSchedSingleMaintenanceQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	region1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	region2 := CreateRegion(ctx, t, apiClient, utils.Region1Request)

	utils.Site1Request.RegionId = region1.JSON200.ResourceId
	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Site1Request.RegionId = nil

	utils.Site1Request.RegionId = region2.JSON200.ResourceId
	site2 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Site1Request.RegionId = nil

	utils.Host1Request.SiteId = site1.JSON200.ResourceId
	host1 := CreateHost(ctx, t, apiClient, utils.Host1Request)
	utils.Host1Request.SiteId = nil

	host2 := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())
	host3 := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())

	utils.Host2Request.SiteId = site2.JSON200.ResourceId
	host4 := CreateHost(ctx, t, apiClient, utils.Host2Request)
	utils.Host2Request.SiteId = nil

	utils.SingleScheduleAlwaysRequest.TargetSiteId = site1.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleScheduleAlwaysRequest)
	utils.SingleScheduleAlwaysRequest.TargetSiteId = nil

	utils.SingleScheduleAlwaysRequest.TargetHostId = host2.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleScheduleAlwaysRequest)
	utils.SingleScheduleAlwaysRequest.TargetHostId = nil

	utils.SingleScheduleNever.TargetHostId = host3.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleScheduleNever)
	utils.SingleScheduleNever.TargetHostId = nil

	utils.SingleScheduleAlwaysRequest.TargetRegionId = region2.JSON200.ResourceId
	CreateSchedSingle(ctx, t, apiClient, utils.SingleScheduleAlwaysRequest)
	utils.SingleScheduleAlwaysRequest.TargetRegionId = nil

	// Host1 should be in maintenance (it's in Site1, and we have maintenance window for it)
	AssertInMaintenance(ctx, t, apiClient, host1.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 1, true)
	AssertInMaintenance(ctx, t, apiClient, nil, site1.JSON200.ResourceId, nil, utils.FutureEpoch, 1, true)

	// Host2 should be in maintenance (it's directly in maintenance)
	AssertInMaintenance(ctx, t, apiClient, host2.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 1, true)

	// Host3 should not be in maintenance
	AssertInMaintenance(ctx, t, apiClient, host3.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 0, false)

	// Host4 should be in maintenance (it's in Region2, and we have maintenance window for it)
	AssertInMaintenance(ctx, t, apiClient, host4.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 1, true)
	AssertInMaintenance(ctx, t, apiClient, nil, site2.JSON200.ResourceId, nil, utils.FutureEpoch, 1, true)
	AssertInMaintenance(ctx, t, apiClient, nil, nil, region2.JSON200.ResourceId, utils.FutureEpoch, 1, true)
}

func TestSchedSingleList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resList, err := apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSingleSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.NotNil(t, resList.JSON200.SingleSchedules)
	assert.Equal(t, 0, len(resList.JSON200.SingleSchedules))
}

func TestSchedList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resList, err := apiClient.ScheduleServiceListSchedulesWithResponse(
		ctx,
		projectName,
		&api.ScheduleServiceListSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.NotNil(t, resList.JSON200.SingleSchedules)
	assert.NotNil(t, resList.JSON200.RepeatedSchedules)
}

func TestSchedSingle_UpdatePatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	siteCreated1 := CreateSite(ctx, t, apiClient, utils.Site1Request)

	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.SiteID
	singleSched1 := CreateSchedSingle(ctx, t, apiClient, utils.SingleSchedule1Request)

	SingleSchedule1Get, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.SingleScheduleID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1Get.StatusCode())
	assert.Equal(t, utils.SschedName1, *SingleSchedule1Get.JSON200.Name)

	siteCreated2 := CreateSite(ctx, t, apiClient, utils.SiteListRequest1)
	utils.SingleSchedule2Request.TargetSiteId = siteCreated2.JSON200.SiteID

	singleSched1Update, err := apiClient.ScheduleServicePatchSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.SingleScheduleID,
		&api.ScheduleServicePatchSingleScheduleParams{},
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, singleSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *singleSched1Update.JSON200.Name)

	SingleSchedule1GetUp, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.SingleScheduleID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1GetUp.StatusCode())
	assert.Equal(t, utils.SschedName2, *SingleSchedule1GetUp.JSON200.Name)
	assert.Equal(t, *siteCreated2.JSON200.SiteID, *SingleSchedule1GetUp.JSON200.TargetSite.ResourceId)
	assert.Equal(
		t,
		*utils.SingleSchedule1Request.EndSeconds,
		*SingleSchedule1GetUp.JSON200.EndSeconds,
	)
	assert.Equal(t, utils.SingleSchedule2Request.ScheduleStatus, SingleSchedule1GetUp.JSON200.ScheduleStatus)

	// Uses PATCH to set empty string to TargetSite and verifies it
	utils.SingleSchedule2Request.TargetSiteId = &emptyString
	singleSched1Update, err = apiClient.ScheduleServicePatchSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.SingleScheduleID,
		&api.ScheduleServicePatchSingleScheduleParams{},
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, singleSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *singleSched1Update.JSON200.Name)

	SingleSchedule1GetUp, err = apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.SingleScheduleID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1GetUp.StatusCode())
	assert.Equal(t, utils.SschedName2, *SingleSchedule1GetUp.JSON200.Name)
	assert.Empty(t, SingleSchedule1GetUp.JSON200.TargetSite)
	assert.Equal(
		t,
		*utils.SingleSchedule1Request.EndSeconds,
		*SingleSchedule1GetUp.JSON200.EndSeconds,
	)
	assert.Equal(t, utils.SingleSchedule2Request.ScheduleStatus, SingleSchedule1GetUp.JSON200.ScheduleStatus)

	// Uses PATCH to set wrong empty string to TargetSite and verifies its BadRequest error
	utils.SingleSchedule2Request.TargetSiteId = &emptyStringWrong
	singleSched1Update, err = apiClient.ScheduleServicePatchSingleScheduleWithResponse(
		ctx,
		projectName,
		*singleSched1.JSON200.SingleScheduleID,
		&api.ScheduleServicePatchSingleScheduleParams{},
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, singleSched1Update.StatusCode())
}
