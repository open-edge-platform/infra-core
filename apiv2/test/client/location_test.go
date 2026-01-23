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

func setupRegionHierarchy(
	ctx context.Context,
	t *testing.T,
	apiClient *api.ClientWithResponses,
) (reg1, reg2, reg3 *api.RegionResource) {
	t.Helper()

	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)

	utils.Region2Request.ParentId = r1.JSON200.ResourceId
	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)
	utils.Region2Request.ParentId = nil

	utils.Region3Request.ParentId = r2.JSON200.ResourceId
	r3 := CreateRegion(ctx, t, apiClient, utils.Region3Request)
	utils.Region3Request.ParentId = nil

	return r1.JSON200, r2.JSON200, r3.JSON200
}

func TestLocation_Metadata(t *testing.T) {
	log.Info().Msgf("Begin Location Metadata Validation OK/NOK tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	region, err := apiClient.RegionServiceCreateRegionWithResponse(
		ctx, projectName, utils.Region1RequestMetadataNOK, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	require.NotNil(t, region)
	assert.Equal(t, http.StatusBadRequest, region.StatusCode())

	region, err = apiClient.RegionServiceCreateRegionWithResponse(
		ctx, projectName, utils.Region1RequestMetadataOK, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	require.NotNil(t, region)
	assert.Equal(t, http.StatusOK, region.StatusCode())

	t.Cleanup(func() { DeleteRegion(context.Background(), t, apiClient, *region.JSON200.ResourceId) })
}

func TestLocation_MetadataInheritance(t *testing.T) {
	log.Info().Msgf("Begin Location Meta Inheritance tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1, r2, r3 := setupRegionHierarchy(ctx, t, apiClient)

	utils.Site1Request.RegionId = r3.ResourceId
	s1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Site1Request.RegionId = nil

	utils.Site2Request.RegionId = r2.ResourceId
	s2 := CreateSite(ctx, t, apiClient, utils.Site2Request)
	utils.Site2Request.RegionId = nil

	projectName := getProjectID(t)

	getr1, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx, projectName, *r1.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getr1.StatusCode())
	assert.Empty(t, getr1.JSON200.ParentRegion.ResourceId)
	assert.Equal(t, utils.MetadataR1, *getr1.JSON200.Metadata)
	assert.Equal(t, []api.MetadataItem{}, *getr1.JSON200.InheritedMetadata)

	getr2, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx, projectName, *r2.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getr2.StatusCode())
	assert.Equal(t, *r1.ResourceId, *getr2.JSON200.ParentRegion.ResourceId)
	assert.Equal(t, utils.MetadataR2, *getr2.JSON200.Metadata)
	assert.Equal(t, []api.MetadataItem{}, *getr2.JSON200.InheritedMetadata)

	getr3, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx, projectName, *r3.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getr3.StatusCode())
	assert.Equal(t, *r2.ResourceId, *getr3.JSON200.ParentId)
	assert.Equal(t, utils.MetadataR3, *getr3.JSON200.Metadata)
	assert.Equal(t, utils.MetadataR3Inherited, *getr3.JSON200.InheritedMetadata)

	gets1, err := apiClient.SiteServiceGetSiteWithResponse(
		ctx, projectName, *s1.JSON200.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, gets1.StatusCode())
	assert.Equal(t, *r3.ResourceId, *gets1.JSON200.Region.ResourceId)
	assert.Equal(t, 2, len(*gets1.JSON200.InheritedMetadata))
	assert.True(
		t,
		ListMetadataContains(*gets1.JSON200.InheritedMetadata, "examplekey2", "r2"),
	)
	assert.True(
		t,
		ListMetadataContains(*gets1.JSON200.InheritedMetadata, "examplekey", "r3"),
	)

	gets2, err := apiClient.SiteServiceGetSiteWithResponse(
		ctx, projectName, *s2.JSON200.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, gets2.StatusCode())
	assert.Equal(t, *r2.ResourceId, *gets2.JSON200.Region.ResourceId)
	assert.Equal(
		t,
		[]api.MetadataItem{{Key: "examplekey", Value: "r2"}},
		*gets2.JSON200.InheritedMetadata,
	)
}

func TestLocation_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin Location RegionSite tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)

	utils.Site1Request.RegionId = nil
	s1 := CreateSite(ctx, t, apiClient, utils.Site1Request)

	utils.Site2Request.RegionId = nil
	s2 := CreateSite(ctx, t, apiClient, utils.Site2Request)

	projectName := getProjectID(t)

	sites1, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sites1.StatusCode())

	sites2, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*r2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sites2.StatusCode())

	s1res, err := apiClient.SiteServiceGetSiteWithResponse(
		ctx,
		projectName,
		*s1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, s1res.StatusCode())

	s2res, err := apiClient.SiteServiceGetSiteWithResponse(
		ctx,
		projectName,
		*s2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, s2res.StatusCode())

	log.Info().Msgf("End Location RegionSite tests")
}

func TestLocation_RegionUpdate(t *testing.T) {
	log.Info().Msgf("Begin Location Region Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)
	assert.Equal(t, utils.Region2Name, *r2.JSON200.Name)

	projectName := getProjectID(t)

	region1Get, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, region1Get.StatusCode())
	assert.Equal(t, utils.Region1Name, *region1Get.JSON200.Name)

	r1Update, err := apiClient.RegionServiceUpdateRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		utils.Region2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r1Update.StatusCode())
	assert.Equal(t, utils.Region2Name, *r1Update.JSON200.Name)

	region1GetUp, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, region1GetUp.StatusCode())
	assert.Equal(t, utils.Region2Name, *region1GetUp.JSON200.Name)

	// Updates using Put r1 Parent with r2 regionID
	utils.Region1Request.ParentId = r2.JSON200.ResourceId
	r1Update, err = apiClient.RegionServiceUpdateRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		utils.Region1Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r1Update.StatusCode())

	// Gets r1 and checks Parent equals to r2 regionID
	region1GetUp, err = apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, region1GetUp.StatusCode())
	assert.Equal(t, utils.Region1Name, *region1GetUp.JSON200.Name)
	assert.Equal(t, *r2.JSON200.ResourceId, *region1GetUp.JSON200.ParentId)

	// Updates using Put r1 Parent with empty string
	utils.Region1Request.ParentId = &emptyString
	r1Update, err = apiClient.RegionServiceUpdateRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		utils.Region1Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r1Update.StatusCode())

	// Gets r1 and checks Parent equals to empty string
	region1GetUp, err = apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, region1GetUp.StatusCode())
	assert.Equal(t, utils.Region1Name, *region1GetUp.JSON200.Name)
	assert.Equal(t, "", *region1GetUp.JSON200.ParentId)

	// Check for BadReqeuest error in case Parent contains empty character in Put
	utils.Region1Request.ParentId = &emptyStringWrong
	r1Update, err = apiClient.RegionServiceUpdateRegionWithResponse(
		ctx,
		projectName,
		*r1.JSON200.ResourceId,
		utils.Region1Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, r1Update.StatusCode())

	// Cleanup
	utils.Region1Request.ParentId = nil
	log.Info().Msgf("End Location Region Update tests")
}

func TestLocation_SiteUpdate(t *testing.T) {
	log.Info().Msgf("Begin Location Site Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)
	assert.Equal(t, utils.Region2Name, *r2.JSON200.Name)

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(ctx, t, apiClient, utils.Site1Request)

	// Updates site using Put, sets Region to r1 regionID and verifies it
	utils.Site1RequestUpdate.RegionId = r1.JSON200.ResourceId
	projectName := getProjectID(t)
	s1Up, err := apiClient.SiteServiceUpdateSiteWithResponse(
		ctx,
		projectName,
		*s1.JSON200.ResourceId,
		utils.Site1RequestUpdate,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, s1Up.StatusCode())

	s1res, err := apiClient.SiteServiceGetSiteWithResponse(
		ctx,
		projectName,
		*s1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, s1res.StatusCode())
	assert.Equal(t, *r1.JSON200.ResourceId, *s1res.JSON200.Region.ResourceId)

	// Updates site using Put, sets Region to emptyString and verifies it
	utils.Site1RequestUpdate.RegionId = &emptyString
	s1Up, err = apiClient.SiteServiceUpdateSiteWithResponse(
		ctx,
		projectName,
		*s1.JSON200.ResourceId,
		utils.Site1RequestUpdate,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, s1Up.StatusCode())

	s1res, err = apiClient.SiteServiceGetSiteWithResponse(
		ctx,
		projectName,
		*s1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, s1res.StatusCode())
	assert.Empty(t, s1res.JSON200.Region.ResourceId)

	// Updates site using Put and Patch, sets Region to wrong emptyString and verifies
	// expected error BadRequest
	utils.Site1RequestUpdate.RegionId = &emptyStringWrong
	s1Up, err = apiClient.SiteServiceUpdateSiteWithResponse(
		ctx,
		projectName,
		*s1.JSON200.ResourceId,
		utils.Site1RequestUpdate,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, s1Up.StatusCode())

	// Sets region in Update resources OU to nil
	utils.Site1RequestUpdatePatch.RegionId = nil
	utils.Site1RequestUpdate.RegionId = nil

	log.Info().Msgf("End Location Site Update tests")
}

func TestLocation_RegionErrors(t *testing.T) {
	log.Info().Msgf("Begin Location Region Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	projectName := getProjectID(t)

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		r1Up, err := apiClient.RegionServiceUpdateRegionWithResponse(
			ctx,
			projectName,
			utils.RegionUnexistID,
			utils.Region1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, r1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.RegionServiceGetRegionWithResponse(
			ctx,
			projectName,
			utils.RegionUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.RegionServiceDeleteRegionWithResponse(
			ctx,
			projectName,
			utils.RegionUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusUnprocessableEntity", func(t *testing.T) {
		r1Up, err := apiClient.RegionServiceUpdateRegionWithResponse(
			ctx,
			projectName,
			utils.RegionWrongID,
			utils.Region1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, r1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusUnprocessableEntity", func(t *testing.T) {
		s1res, err := apiClient.RegionServiceGetRegionWithResponse(
			ctx,
			projectName,
			utils.RegionWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusUnprocessableEntity", func(t *testing.T) {
		resDelSite, err := apiClient.RegionServiceDeleteRegionWithResponse(
			ctx,
			projectName,
			utils.RegionWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resDelSite.StatusCode())
	})

	log.Info().Msgf("End Location Region Error tests")
}

func TestLocation_SiteErrors(t *testing.T) {
	log.Info().Msgf("Begin Location Site Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)
	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1Up, err := apiClient.SiteServiceUpdateSiteWithResponse(
			ctx,
			projectName,
			utils.SiteUnexistID,
			utils.Site1RequestUpdate,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.SiteServiceGetSiteWithResponse(
			ctx,
			projectName,
			utils.SiteUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.SiteServiceDeleteSiteWithResponse(
			ctx,
			projectName,
			utils.SiteUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusUnprocessableEntity", func(t *testing.T) {
		s1Up, err := apiClient.SiteServiceUpdateSiteWithResponse(
			ctx,
			projectName,
			utils.SiteWrongID,
			utils.Site1RequestUpdate,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, s1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusUnprocessableEntity", func(t *testing.T) {
		s1res, err := apiClient.SiteServiceGetSiteWithResponse(
			ctx,
			projectName,
			utils.SiteWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusUnprocessableEntity", func(t *testing.T) {
		resDelSite, err := apiClient.SiteServiceDeleteSiteWithResponse(
			ctx,
			projectName,
			utils.SiteWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resDelSite.StatusCode())
	})
	log.Info().Msgf("End Location Site Error tests")
}

func TestRegionList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	baselineList, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineList.StatusCode())
	baselineTotalItems := len(baselineList.JSON200.Regions)

	totalItems := 10
	pageID := 1
	pageSize := 4

	for id := 0; id < totalItems; id++ {
		CreateRegion(ctx, t, apiClient, utils.Region1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{
			Offset:   &pageID,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Regions), pageSize)
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems+baselineTotalItems, len(resList.JSON200.Regions))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.NotNil(t, resList)
}

func TestLocation_RegionListQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	baselineAllList, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineAllList.StatusCode())
	baselineAllRegions := len(baselineAllList.JSON200.Regions)

	baselineRootList, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{Filter: &FilterRegionNotHasParent},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineRootList.StatusCode())
	baselineRootRegions := len(baselineRootList.JSON200.Regions)

	postResp1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)

	utils.Region2Request.ParentId = postResp1.JSON200.ResourceId
	CreateRegion(ctx, t, apiClient, utils.Region2Request)

	utils.Region3Request.ParentId = postResp1.JSON200.ResourceId
	CreateRegion(ctx, t, apiClient, utils.Region3Request)

	// Checks Regions with Parent Region ID
	filterByParentRegionID := fmt.Sprintf(FilterRegionParentID, *postResp1.JSON200.ResourceId)
	resList, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{
			Filter: &filterByParentRegionID,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 2, len(resList.JSON200.Regions))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Checks all Regions
	resList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 3+baselineAllRegions, len(resList.JSON200.Regions))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Checks Regions without Parent Region ID
	// emptyParent := "null"
	resList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{
			Filter: &FilterRegionNotHasParent,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1+baselineRootRegions, len(resList.JSON200.Regions))
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, false, resList.JSON200.HasNext)

	resList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.NotNil(t, resList)
}

func TestLocation_SiteList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	baselineList, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineList.StatusCode())
	baselineTotalItems := len(baselineList.JSON200.Sites)

	totalItems := 10
	pageID := 1
	pageSize := 4

	for id := 0; id < totalItems; id++ {
		CreateSite(ctx, t, apiClient, utils.SiteListRequest)
	}

	// Checks if list resources return expected number of entries
	resSiteList, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{
			Offset:   &pageID,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, len(resSiteList.JSON200.Sites), pageSize)
	assert.Equal(t, true, resSiteList.JSON200.HasNext)

	resSiteList, err = apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, totalItems+baselineTotalItems, len(resSiteList.JSON200.Sites))
	assert.Equal(t, false, resSiteList.JSON200.HasNext)
}

func TestLocation_SiteListQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	baselineAllList, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineAllList.StatusCode())
	baselineAllSites := len(baselineAllList.JSON200.Sites)

	baselineNoRegionList, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{Filter: &FilterSiteNotHasRegion},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineNoRegionList.StatusCode())
	baselineNoRegionSites := len(baselineNoRegionList.JSON200.Sites)

	postRespRegion := CreateRegion(ctx, t, apiClient, utils.Region1Request)

	CreateSite(ctx, t, apiClient, utils.SiteListRequest1)

	CreateSite(ctx, t, apiClient, utils.SiteListRequest2)

	utils.SiteListRequest3.RegionId = postRespRegion.JSON200.ResourceId
	CreateSite(ctx, t, apiClient, utils.SiteListRequest3)

	// Checks query to all sites
	resSiteList, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, 3+baselineAllSites, len(resSiteList.JSON200.Sites))
	assert.Equal(t, false, resSiteList.JSON200.HasNext)

	// Checks query to sites with region ID
	filterByRegionID := fmt.Sprintf(FilterSiteRegionID, *postRespRegion.JSON200.ResourceId)
	resSiteList, err = apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{
			Filter: &filterByRegionID,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, 1, len(resSiteList.JSON200.Sites))
	assert.Equal(t, false, resSiteList.JSON200.HasNext)

	// Checks query to sites without region ID
	// emptyRegion := "null"
	resSiteList, err = apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{
			Filter: &FilterSiteNotHasRegion,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, 2+baselineNoRegionSites, len(resSiteList.JSON200.Sites))
	assert.Equal(t, false, resSiteList.JSON200.HasNext)

	resSiteList, err = apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, false, resSiteList.JSON200.HasNext)
}

func TestLocation_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resRegionList, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resRegionList.StatusCode())
	baselineRegions := len(resRegionList.JSON200.Regions)

	resRegionList, err = apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resRegionList.StatusCode())
	assert.Equal(t, baselineRegions, len(resRegionList.JSON200.Regions))

	resSiteList, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	baselineSites := len(resSiteList.JSON200.Sites)

	resSiteList, err = apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSiteList.StatusCode())
	assert.Equal(t, baselineSites, len(resSiteList.JSON200.Sites))
}

func TestLocation_Filter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	baselineFilter := fmt.Sprintf(`NOT has(%s)`, "parent_region")
	baselineRegionsWithSites, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{
			ShowTotalSites: &showSites,
			Filter:         &baselineFilter,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineRegionsWithSites.StatusCode())
	baselineNoParentRegions := len(baselineRegionsWithSites.JSON200.Regions)

	// create regions
	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	utils.Region2Request.ParentId = nil
	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)

	// create sites with region
	s1req := utils.Site1Request
	s1req.RegionId = r1.JSON200.ResourceId
	s1req.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(ctx, t, apiClient, s1req)

	s2req := utils.Site2Request
	s2req.RegionId = r2.JSON200.ResourceId
	s2 := CreateSite(ctx, t, apiClient, s2req)

	// filter- site->region->resource-id
	filter := fmt.Sprintf("region.resourceId=%q", *r1.JSON200.ResourceId)
	sites1, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{Filter: &filter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sites1.StatusCode())
	assert.Equal(t, 1, int(sites1.JSON200.TotalElements))
	assert.Equal(t, *s1.JSON200.Region.ResourceId, *r1.JSON200.ResourceId)

	// filter- site->region->resource-id
	filter = fmt.Sprintf("region.resourceId=%q", *r2.JSON200.ResourceId)
	sites2, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{Filter: &filter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sites2.StatusCode())
	assert.Equal(t, 1, int(sites2.JSON200.TotalElements))
	assert.Equal(t, *s2.JSON200.Region.ResourceId, *r2.JSON200.ResourceId)

	// filter- region with ShotTotalSites: region1 and region2 has not parent and 1 site each
	filter = fmt.Sprintf(`NOT has(%s)`, "parent_region")
	regions, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{
			ShowTotalSites: &showSites,
			Filter:         &filter,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sites2.StatusCode())
	assert.Equal(t, 2+baselineNoParentRegions, int(regions.JSON200.TotalElements))
	assert.Equal(t, 2+baselineNoParentRegions, len(regions.JSON200.Regions))

	var foundRegion1, foundRegion2 bool
	for _, region := range regions.JSON200.Regions {
		if region.ResourceId != nil && *region.ResourceId == *r1.JSON200.ResourceId {
			assert.Equal(t, 1, int(*region.TotalSites))
			foundRegion1 = true
		}
		if region.ResourceId != nil && *region.ResourceId == *r2.JSON200.ResourceId {
			assert.Equal(t, 1, int(*region.TotalSites))
			foundRegion2 = true
		}
	}
	assert.True(t, foundRegion1)
	assert.True(t, foundRegion2)
}

func TestLocation_FilterSites(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	// create regions
	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	utils.Region2Request.ParentId = nil
	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)

	// create sites with region
	s1req := utils.Site1Request
	s1req.RegionId = r1.JSON200.ResourceId
	s1req.RegionId = r1.JSON200.ResourceId
	CreateSite(ctx, t, apiClient, s1req)

	s2req := utils.Site2Request
	s2req.RegionId = r2.JSON200.ResourceId
	CreateSite(ctx, t, apiClient, s2req)

	orderByResourceID := "resource_id asc"
	orderByRegion := "site_resource_region desc, resource_id"
	orderByWrong := "resource_?"

	type testCase struct {
		name            string
		filter          string
		orderby         *string
		amountResources int
		fail            bool
	}

	testCasesSites := []testCase{
		{
			name:            "test sites: no resource_id",
			filter:          fmt.Sprintf(`%s = ""`, "resource_id"),
			amountResources: 0,
			fail:            false,
		},
		{
			name:            "test sites: no region with parent_region",
			filter:          fmt.Sprintf(`has(%s.%s)`, "region", "parent_region"),
			amountResources: 0,
			fail:            false,
		},
		{
			name:            "test sites: sites with no region",
			filter:          fmt.Sprintf(`NOT has(%s)`, "region"),
			amountResources: 0,
			fail:            false,
		},
		{
			name:            "test sites: sites with non existing metadata",
			filter:          fmt.Sprintf(`%s = '%s'`, "metadata", `{"key":"cluster-name","value":""}`),
			amountResources: 0,
			fail:            false,
		},
		{
			name:            "test sites: sites with existing metadata - site2",
			filter:          fmt.Sprintf(`%s = '%s'`, "metadata", `{"key":"examplekey2","value":"site1"}`),
			orderby:         &orderByResourceID,
			amountResources: 1,
			fail:            false,
		},
		{
			name:            "test sites: sites with existing metadata - site2",
			filter:          fmt.Sprintf(`%s = '%s'`, "metadata", `{"key":"examplekey2","value":"site1"}`),
			orderby:         &orderByRegion,
			amountResources: 1,
			fail:            false,
		},
		{
			name:            "test sites: sites with bad metadata value",
			filter:          fmt.Sprintf(`%s = '%s'`, "metadata", `{"key":"??","value":"site1"}`),
			amountResources: 0,
			fail:            true,
		},
		{
			name:            "test sites: sites with bad orderby value",
			filter:          fmt.Sprintf(`%s = '%s'`, "metadata", `{"key":"examplekey2","value":"site1"}`),
			orderby:         &orderByWrong,
			amountResources: 0,
			fail:            true,
		},
	}

	for _, tc := range testCasesSites {
		t.Run(tc.name, func(t *testing.T) {
			sites, err := apiClient.SiteServiceListSitesWithResponse(
				ctx,
				projectName,
				&api.SiteServiceListSitesParams{
					Filter:  &tc.filter,
					OrderBy: tc.orderby,
				},
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			require.NoError(t, err)

			if !tc.fail {
				require.Equal(t, http.StatusOK, sites.StatusCode())
				assert.Equal(t, tc.amountResources, int(sites.JSON200.TotalElements))
				assert.Equal(t, tc.amountResources, len(sites.JSON200.Sites))
			} else {
				require.NotEqual(t, http.StatusOK, sites.StatusCode())
			}
		})
	}
}

func TestRegion_Patch(t *testing.T) {
	log.Info().Msgf("Begin Region Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	// Create a Region
	region := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *region.JSON200.Name)

	// Modify fields for patching
	newName := utils.Region1Name + "-updated"
	patchRequest := api.RegionResource{
		Name: &newName,
	}

	// Perform the Patch operation
	updatedRegion, err := apiClient.RegionServicePatchRegionWithResponse(
		ctx,
		projectName,
		*region.JSON200.ResourceId,
		&api.RegionServicePatchRegionParams{},
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedRegion.StatusCode())
	assert.Equal(t, newName, *updatedRegion.JSON200.Name)

	// Verify the changes with a Get operation
	getRegion, err := apiClient.RegionServiceGetRegionWithResponse(
		ctx,
		projectName,
		*region.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getRegion.StatusCode())
	assert.Equal(t, newName, *getRegion.JSON200.Name)

	log.Info().Msgf("End Region Patch tests")
}

func TestSite_Patch(t *testing.T) {
	log.Info().Msgf("Begin Site Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	// Create a Site
	region := CreateRegion(ctx, t, apiClient, utils.Region1Request)
	utils.Site1Request.RegionId = region.JSON200.ResourceId
	site := CreateSite(ctx, t, apiClient, utils.Site1Request)
	assert.Equal(t, utils.Site1Name, *site.JSON200.Name)

	// Modify fields for patching
	newName := utils.Site1Name + "-updated"
	patchRequest := api.SiteResource{
		Name: &newName,
	}

	// Perform the Patch operation
	updatedSite, err := apiClient.SiteServicePatchSiteWithResponse(
		ctx,
		projectName,
		*site.JSON200.ResourceId,
		&api.SiteServicePatchSiteParams{},
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedSite.StatusCode())
	assert.Equal(t, newName, *updatedSite.JSON200.Name)

	// Verify the changes with a Get operation
	getSite, err := apiClient.SiteServiceGetSiteWithResponse(
		ctx,
		projectName,
		*site.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getSite.StatusCode())
	assert.Equal(t, newName, *getSite.JSON200.Name)

	log.Info().Msgf("End Site Patch tests")
}
