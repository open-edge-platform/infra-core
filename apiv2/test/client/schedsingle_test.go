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

func TestSchedSingle_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin SingleSched tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	singleSched1 := CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule1Request)

	utils.SingleSchedule2Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	singleSched2 := CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule2Request)

	get1, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		*singleSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.SschedName1, *get1.JSON200.Name)

	get2, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
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

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = siteCreated1.JSON200.ResourceId
	hostCreated1 := CreateHost(t, ctx, apiClient, utils.Host1Request)

	// Expected BadRequest Error because of target site and host are set in Schedule
	utils.SingleScheduleError.TargetSiteId = siteCreated1.JSON200.ResourceId
	utils.SingleScheduleError.TargetHostId = hostCreated1.JSON200.ResourceId

	sched, err := apiClient.ScheduleServiceCreateSingleScheduleWithResponse(
		ctx,
		utils.SingleScheduleError,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, sched.StatusCode())

	utils.SingleScheduleErrorSeconds.TargetSiteId = siteCreated1.JSON200.ResourceId
	sched, err = apiClient.ScheduleServiceCreateSingleScheduleWithResponse(
		ctx,
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

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId
	singleSched1 := CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule1Request)

	SingleSchedule1Get, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		*singleSched1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, SingleSchedule1Get.StatusCode())
	assert.Equal(t, utils.SschedName1, *SingleSchedule1Get.JSON200.Name)

	siteCreated2 := CreateSite(t, ctx, apiClient, utils.SiteListRequest1)

	utils.SingleSchedule2Request.TargetSiteId = siteCreated2.JSON200.ResourceId
	singleSched1Update, err := apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
		ctx,
		*singleSched1.JSON200.ResourceId,
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, singleSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *singleSched1Update.JSON200.Name)

	SingleSchedule1GetUp, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
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
		*singleSched1.JSON200.ResourceId,
		utils.SingleSchedule2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, singleSched1Update.StatusCode())
	assert.Equal(t, utils.SschedName2, *singleSched1Update.JSON200.Name)

	SingleSchedule1GetUp, err = apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
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
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		singleSched1Up, err := apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
			ctx,
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
			utils.SingleScheduleUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.ScheduleServiceDeleteSingleScheduleWithResponse(
			ctx,
			utils.SingleScheduleUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusNotFound", func(t *testing.T) {
		singleSched1Up, err := apiClient.ScheduleServiceUpdateSingleScheduleWithResponse(
			ctx,
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
			utils.SingleScheduleWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelSite, err := apiClient.ScheduleServiceDeleteSingleScheduleWithResponse(
			ctx,
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

	siteCreated1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	utils.SingleSchedule1Request.TargetSiteId = siteCreated1.JSON200.ResourceId

	totalItems := 10
	var pageId int32 = 1
	var pageSize int32 = 4

	for id := 0; id < totalItems; id++ {
		CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListSingleSchedulesParams{
			Offset:   &pageId,
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
		&api.ScheduleServiceListSingleSchedulesParams{
			Offset:   &pageId,
			PageSize: &pageSize,
			SiteId:   siteCreated1.JSON200.ResourceId,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, int(pageSize), len(resList.JSON200.SingleSchedules))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
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

	postRespSite1 := CreateSite(t, ctx, apiClient, utils.Site1Request)
	postRespSite2 := CreateSite(t, ctx, apiClient, utils.Site2Request)

	utils.SingleSchedule1Request.TargetSiteId = postRespSite1.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule1Request)

	utils.SingleSchedule2Request.TargetSiteId = postRespSite2.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule2Request)

	utils.SingleSchedule3Request.TargetSiteId = postRespSite2.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule3Request)

	// Checks list of SingleSchedules with noo siteID
	resList, err := apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
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

	utils.SingleScheduleAlwaysRequest.TargetSiteId = site1.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleScheduleAlwaysRequest)
	utils.SingleScheduleAlwaysRequest.TargetSiteId = nil

	utils.SingleScheduleAlwaysRequest.TargetHostId = host2.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleScheduleAlwaysRequest)
	utils.SingleScheduleAlwaysRequest.TargetHostId = nil

	utils.SingleScheduleNever.TargetHostId = host3.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleScheduleNever)
	utils.SingleScheduleNever.TargetHostId = nil

	utils.SingleScheduleAlwaysRequest.TargetRegionId = region2.JSON200.ResourceId
	CreateSchedSingle(t, ctx, apiClient, utils.SingleScheduleAlwaysRequest)
	utils.SingleScheduleAlwaysRequest.TargetRegionId = nil

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

func TestSchedSingleList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.ScheduleServiceListSingleSchedulesWithResponse(
		ctx,
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

	resList, err := apiClient.ScheduleServiceListSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListSchedulesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.NotNil(t, resList.JSON200.SingleSchedules)
	assert.NotNil(t, resList.JSON200.RepeatedSchedules)
}

func TestSingleSchedule_Patch(t *testing.T) {
	log.Info().Msgf("Begin SingleSchedule Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Create a SingleSchedule
	schedule := CreateSchedSingle(t, ctx, apiClient, utils.SingleSchedule1Request)
	assert.Equal(t, utils.SingleSchedule1Request.ScheduleStatus, schedule.JSON200.ScheduleStatus)
	assert.Equal(t, utils.SingleSchedule1Request.StartSeconds, schedule.JSON200.StartSeconds)

	// Modify fields for patching
	newScheduleStatus := api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSOSUPDATE
	var newStartSeconds uint32 = 1234567890
	patchRequest := api.SingleScheduleResource{
		ScheduleStatus: newScheduleStatus,
		StartSeconds:   newStartSeconds,
	}

	// Perform the Patch operation
	updatedSchedule, err := apiClient.ScheduleServicePatchSingleScheduleWithResponse(
		ctx,
		*schedule.JSON200.ResourceId,
		nil,
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedSchedule.StatusCode())
	assert.Equal(t, newScheduleStatus, updatedSchedule.JSON200.ScheduleStatus)
	assert.Equal(t, newStartSeconds, updatedSchedule.JSON200.StartSeconds)

	// Verify the changes with a Get operation
	getSchedule, err := apiClient.ScheduleServiceGetSingleScheduleWithResponse(
		ctx,
		*schedule.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getSchedule.StatusCode())
	assert.Equal(t, newScheduleStatus, getSchedule.JSON200.ScheduleStatus)
	assert.Equal(t, newStartSeconds, getSchedule.JSON200.StartSeconds)

	log.Info().Msgf("End SingleSchedule Patch tests")
}
