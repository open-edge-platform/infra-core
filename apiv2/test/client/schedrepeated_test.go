// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedRepeated_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin RepeatedSched tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.RegionId = nil
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.RepeatedSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	RepeatedSched1 := CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)

	utils.RepeatedSchedule2Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	RepeatedSched2 := CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)

	get1, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.SschedName1, *get1.JSON200.Name)

	get2, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, utils.SschedName1, *get2.JSON200.Name)

	log.Info().Msgf("End RepeatedSchedule tests")
}

func TestSchedRepeated_CreateError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.Region = nil
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = siteCreated1.JSON200.ResourceId
	hostCreated1 := CreateHost(t, ctx, apiClient, utils.Host1Request)

	// Expected StatusBadRequest Error because of target site and host are set in Schedule
	utils.RepeatedScheduleError.TargetSiteId = siteCreated1.JSON200.ResourceId
	utils.RepeatedScheduleError.TargetHostId = hostCreated1.JSON200.ResourceId

	sched, err := apiClient.ScheduleServiceCreateRepeatedScheduleWithResponse(
		ctx,
		utils.RepeatedScheduleError,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, sched.StatusCode())
}

func TestSchedRepeated_UpdatePut(t *testing.T) {
	log.Info().Msgf("Begin RepeatedSchedule Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.Region = nil

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.RepeatedSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	RepeatedSched1 := CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)

	RepeatedSchedule1Get, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, RepeatedSchedule1Get.StatusCode())
	assert.Equal(t, utils.SschedName1, *RepeatedSchedule1Get.JSON200.Name)

	utils.SiteListRequest1.Region = nil

	siteCreated2 := CreateSite(t, ctx, apiClient, utils.SiteListRequest1)
	utils.RepeatedSchedule2Request.TargetSiteId = siteCreated2.JSON200.ResourceId

	RepeatedSched1Update, err := apiClient.ScheduleServiceUpdateRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		utils.RepeatedSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, RepeatedSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *RepeatedSched1Update.JSON200.Name)

	RepeatedSchedule1GetUp, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, RepeatedSchedule1GetUp.StatusCode())
	assert.Equal(t, utils.SschedName2, *RepeatedSchedule1GetUp.JSON200.Name)
	assert.Equal(t, *siteCreated2.JSON200.ResourceId, *RepeatedSchedule1GetUp.JSON200.TargetSite.ResourceId)
	assert.Equal(
		t,
		utils.RepeatedSchedule2Request.CronDayMonth,
		RepeatedSchedule1GetUp.JSON200.CronDayMonth,
	)

	// Uses PUT to set empty string to TargetSite and verifies it
	utils.RepeatedSchedule2Request.TargetSiteId = &emptyString
	RepeatedSched1Update, err = apiClient.ScheduleServiceUpdateRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		utils.RepeatedSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, RepeatedSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *RepeatedSched1Update.JSON200.Name)

	RepeatedSchedule1GetUp, err = apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, RepeatedSchedule1GetUp.StatusCode())
	assert.Equal(t, utils.SschedName2, *RepeatedSchedule1GetUp.JSON200.Name)
	assert.Empty(t, RepeatedSchedule1GetUp.JSON200.TargetSite)

	// Uses PUT to set wrong empty string to TargetSite and verifies its BadRequest error
	utils.RepeatedSchedule2Request.TargetSiteId = &emptyStringWrong
	RepeatedSched1Update, err = apiClient.ScheduleServiceUpdateRepeatedScheduleWithResponse(
		ctx,
		*RepeatedSched1.JSON200.ResourceId,
		utils.RepeatedSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, RepeatedSched1Update.StatusCode())

	utils.RepeatedSchedule2Request.TargetSite = nil

	log.Info().Msgf("End RepeatedSchedule Update tests")
}

func TestSchedRepeated_Errors(t *testing.T) {
	log.Info().Msgf("Begin RepeatedSchedule Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	utils.Site1Request.Region = nil
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.RepeatedSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		RepeatedSched1Up, err := apiClient.ScheduleServiceUpdateRepeatedScheduleWithResponse(
			ctx,
			utils.RepeatedScheduleUnexistID,
			utils.RepeatedSchedule1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, RepeatedSched1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
			ctx,
			utils.RepeatedScheduleUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.ScheduleServiceDeleteRepeatedScheduleWithResponse(
			ctx,
			utils.RepeatedScheduleUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_NotFoundError", func(t *testing.T) {
		RepeatedSched1Up, err := apiClient.ScheduleServiceUpdateRepeatedScheduleWithResponse(
			ctx,
			utils.RepeatedScheduleWrongID,
			utils.RepeatedSchedule1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, RepeatedSched1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
			ctx,
			utils.RepeatedScheduleWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.ScheduleServiceDeleteRepeatedScheduleWithResponse(
			ctx,
			utils.RepeatedScheduleWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	// Verify partial updates
	utils.RepeatedSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	RepeatedSched1 := CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)

	t.Run("Put_WrongCron_StatusBadRequest", func(t *testing.T) {
		RepeatedSched1Up, err := apiClient.ScheduleServiceUpdateRepeatedScheduleWithResponse(
			ctx,
			*RepeatedSched1.JSON200.ResourceId,
			utils.RepeatedMissingRequest,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, RepeatedSched1Up.StatusCode())
	})

	log.Info().Msgf("End RepeatedSchedule Error tests")
}

func TestSchedRepeatedList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.Region = nil
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.RepeatedSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId

	totalItems := 10
	var pageId int32 = 1
	var pageSize int32 = 4

	for id := 0; id < totalItems; id++ {
		CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{
			Offset:   &pageId,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, int(pageSize), len(resList.JSON200.RepeatedSchedules))
	assert.Equal(t, true, resList.JSON200.HasNext)

	// Checks if list resources return expected number of entries
	resList, err = apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{
			Offset:   &pageId,
			PageSize: &pageSize,
			SiteId:   siteCreated1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, int(pageSize), len(resList.JSON200.RepeatedSchedules))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{
			SiteId: siteCreated1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.RepeatedSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestSchedRepeatedListQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.Region = nil
	postRespSite1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Site2Request.Region = nil
	postRespSite2 := CreateSite(t, ctx, apiClient, utils.Site2Request)

	utils.RepeatedSchedule1Request.TargetSiteId = postRespSite1.JSON200.ResourceId
	CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)

	utils.RepeatedSchedule2Request.TargetSiteId = postRespSite2.JSON200.ResourceId
	CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule2Request)

	utils.RepeatedSchedule3Request.TargetSiteId = postRespSite2.JSON200.ResourceId
	CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule3Request)

	// Checks list of RepeatedSchedules with siteID 1
	resList, err := apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{
			SiteId: postRespSite1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, len(resList.JSON200.RepeatedSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Checks list of all RepeatedSchedules
	resList, err = apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3, len(resList.JSON200.RepeatedSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Checks list of RepeatedSchedules with SiteId 2
	resList, err = apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{
			SiteId: postRespSite2.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 2, len(resList.JSON200.RepeatedSchedules))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestSchedRepeatedMaintenanceQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	region1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	region2 := CreateRegion(t, ctx, apiClient, utils.Region1Request)

	utils.Site1Request.RegionId = region1.JSON200.ResourceId
	site1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Site1Request.RegionId = nil

	utils.Site1Request.RegionId = region2.JSON200.ResourceId
	site2 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.Site1Request.RegionId = nil

	utils.Host1Request.SiteId = site1.JSON200.ResourceId
	host1 := CreateHost(t, ctx, apiClient, utils.Host1Request)
	utils.Host1Request.SiteId = nil

	host2 := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())
	host3 := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())

	utils.Host2Request.SiteId = site2.JSON200.ResourceId
	host4 := CreateHost(t, ctx, apiClient, utils.Host2Request)
	utils.Host2Request.SiteId = nil

	utils.RepeatedScheduleAlwaysRequest.TargetSiteId = site1.JSON200.ResourceId
	CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedScheduleAlwaysRequest)
	utils.RepeatedScheduleAlwaysRequest.TargetSiteId = nil

	utils.RepeatedScheduleAlwaysRequest.TargetHostId = host2.JSON200.ResourceId
	CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedScheduleAlwaysRequest)
	utils.RepeatedScheduleAlwaysRequest.TargetHostId = nil

	utils.RepeatedScheduleAlwaysRequest.TargetRegionId = region2.JSON200.ResourceId
	CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedScheduleAlwaysRequest)
	utils.RepeatedScheduleAlwaysRequest.TargetRegionId = nil

	// Host1 should be in maintenance (it's in Site1, and we have maintenance window for it)
	AssertInMaintenance(t, ctx, apiClient, host1.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 1, true)
	AssertInMaintenance(t, ctx, apiClient, nil, site1.JSON200.ResourceId, nil, utils.FutureEpoch, 1, true)

	// Host2 should be in maintenance (it's directly in maintenance)
	AssertInMaintenance(t, ctx, apiClient, host2.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 1, true)

	// Host3 should not be in maintenance
	AssertInMaintenance(t, ctx, apiClient, host3.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 0, false)

	// Host4 should be in maintenance (it's in Region2, and we have maintenance window for it)
	AssertInMaintenance(t, ctx, apiClient, host4.JSON200.ResourceId, nil, nil, utils.FutureEpoch, 1, true)
	AssertInMaintenance(t, ctx, apiClient, nil, site2.JSON200.ResourceId, nil, utils.FutureEpoch, 1, true)
	AssertInMaintenance(t, ctx, apiClient, nil, nil, region2.JSON200.ResourceId, utils.FutureEpoch, 1, true)
}

func TestSchedRepeatedList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.ScheduleServiceListRepeatedSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListRepeatedSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.NotNil(t, resList.JSON200.RepeatedSchedules)
	assert.Equal(t, 0, len(resList.JSON200.RepeatedSchedules))
}

func TestSchedRepeated_cronjobValidationError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	utils.Site1Request.Region = nil
	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.RepeatedScheduleCronReqErr.TargetSiteId = siteCreated1.JSON200.ResourceId

	sched, err := apiClient.ScheduleServiceCreateRepeatedScheduleWithResponse(
		ctx,
		utils.RepeatedScheduleCronReqErr,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	assert.Equal(t, http.StatusBadRequest, sched.StatusCode())
	require.NoError(t, err)
}

func TestRepeatedSchedule_Patch(t *testing.T) {
	log.Info().Msgf("Begin RepeatedSchedule Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Create a RepeatedSchedule
	schedule := CreateSchedRepeated(t, ctx, apiClient, utils.RepeatedSchedule1Request)
	assert.Equal(t, utils.RepeatedSchedule1Request.ScheduleStatus, schedule.JSON200.ScheduleStatus)
	assert.Equal(t, utils.RepeatedSchedule1Request.CronDayWeek, schedule.JSON200.CronDayWeek)

	// Modify fields for patching
	newScheduleStatus := api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSOSUPDATE
	newCronDayWeek := "0 12 * * 1-5"
	patchRequest := api.RepeatedScheduleResource{
		ScheduleStatus: newScheduleStatus,
		CronDayWeek:    newCronDayWeek,
	}

	// Perform the Patch operation
	updatedSchedule, err := apiClient.ScheduleServicePatchRepeatedScheduleWithResponse(
		ctx,
		*schedule.JSON200.ResourceId,
		nil,
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedSchedule.StatusCode())
	assert.Equal(t, newScheduleStatus, updatedSchedule.JSON200.ScheduleStatus)
	assert.Equal(t, newCronDayWeek, updatedSchedule.JSON200.CronDayWeek)

	// Verify the changes with a Get operation
	getSchedule, err := apiClient.ScheduleServiceGetRepeatedScheduleWithResponse(
		ctx,
		*schedule.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getSchedule.StatusCode())
	assert.Equal(t, newScheduleStatus, getSchedule.JSON200.ScheduleStatus)
	assert.Equal(t, newCronDayWeek, getSchedule.JSON200.CronDayWeek)

	log.Info().Msgf("End RepeatedSchedule Patch tests")
}
