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

var (
	showRegions            = true
	showSites              = true
	emptyParentID          = ""
	commonSiteRegionSuffix = "-12345678"
	commonManySuffix       = "23456"

	regionPrefixName    = "state"
	subRegionPrefixName = "city"
	sitePrefixName      = "building"
	maxRegions          = 10
	maxSubRegions       = 10
	maxSites            = 10

	regionKind = api.RESOURCEKINDREGION
	siteKind   = api.RESOURCEKINDSITE
)

type testCase struct {
	name            string
	params          *api.LocationServiceListLocationsParams
	expected        []api.ListLocationsResponseLocationNode
	listedElements  int  // the expected length of the Nodes array inside the response
	totalElements   int  // the expected response value of TotalElements
	outputElements  int  // the expected response value of outputElements
	allowMoreListed bool // allow extra ancestors when ordering changes
}

//nolint:gocritic // more than 5 return value to return the whole hierarchy.
func setupRegionSiteHierarchy(
	ctx context.Context,
	t *testing.T,
	apiClient *api.ClientWithResponses,
) (reg1, reg2, reg3 *api.RegionResource, site1, site2, site3 *api.SiteResource) {
	t.Helper()

	r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)

	utils.Region2Request.ParentId = r1.JSON200.ResourceId
	r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)
	utils.Region2Request.ParentId = nil

	utils.Region3Request.ParentId = r2.JSON200.ResourceId
	r3 := CreateRegion(ctx, t, apiClient, utils.Region3Request)
	utils.Region3Request.ParentId = nil

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Site1Request.RegionId = nil

	utils.Site2Request.RegionId = r2.JSON200.ResourceId
	s2 := CreateSite(ctx, t, apiClient, utils.Site2Request)
	utils.Site2Request.Region = nil

	utils.Site2Request.RegionId = r2.JSON200.ResourceId
	s3 := CreateSite(ctx, t, apiClient, utils.Site3Request)
	utils.Site2Request.Region = nil

	return r1.JSON200, r2.JSON200, r3.JSON200, s1.JSON200, s2.JSON200, s3.JSON200
}

func setupRegionSiteLargeHierarchy(
	ctx context.Context,
	t *testing.T,
	apiClient *api.ClientWithResponses,
) {
	t.Helper()
	for r := 0; r < maxRegions; r++ {
		regName := fmt.Sprintf("%s-%d", regionPrefixName, r)
		utils.Region1Request.Name = &regName
		utils.Region1Request.ParentId = nil
		r1 := CreateRegion(ctx, t, apiClient, utils.Region1Request)
		utils.Region1Request.Name = &utils.Region1Name

		for sr := 0; sr < maxSubRegions; sr++ {
			subregName := fmt.Sprintf("%s-%d-%d", subRegionPrefixName, r, sr)
			utils.Region2Request.Name = &subregName
			utils.Region2Request.ParentId = r1.JSON200.ResourceId
			r2 := CreateRegion(ctx, t, apiClient, utils.Region2Request)
			utils.Region2Request.ParentId = nil
			utils.Region2Request.Name = &utils.Region2Name

			for si := 0; si < maxSites; si++ {
				siteName := fmt.Sprintf("%s-%s-%d", subRegionPrefixName, sitePrefixName, si)
				utils.Site2Request.Name = &siteName
				utils.Site2Request.RegionId = r2.JSON200.ResourceId
				CreateSite(ctx, t, apiClient, utils.Site2Request)
				utils.Site2Request.Region = nil
				utils.Site2Request.Name = &utils.Site2Name
			}
		}
	}
}

func TestLocation_Hierarchy(t *testing.T) {
	log.Info().Msgf("Begin TestLocation_Hierarchy")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1, r2, r3, s1, s2, s3 := setupRegionSiteHierarchy(ctx, t, apiClient)

	testCases := []testCase{
		{
			name: "Test root regions",
			params: &api.LocationServiceListLocationsParams{
				Name:        &utils.Region1Name,
				ShowRegions: &showRegions,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  1,
			outputElements: 1,
		},
		{
			name: "Test mid regions tree: looks for r2 -> gets [r1,r2]",
			params: &api.LocationServiceListLocationsParams{
				Name:        &utils.Region2Name,
				ShowRegions: &showRegions,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *r2.ResourceId,
					Name:       *r2.Name,
					ParentId:   *r1.ResourceId,
					Type:       regionKind,
				},
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  1,
			outputElements: 1,
		},
		{
			name: "Test mid regions tree: looks for r3 -> gets [r1,r2,r3]",
			params: &api.LocationServiceListLocationsParams{
				Name:        &utils.Region3Name,
				ShowRegions: &showRegions,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *r3.ResourceId,
					Name:       *r3.Name,
					ParentId:   *r2.ResourceId,
					Type:       regionKind,
				},
				{
					ResourceId: *r2.ResourceId,
					Name:       *r2.Name,
					ParentId:   *r1.ResourceId,
					Type:       regionKind,
				},
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  1,
			outputElements: 1,
		},
		{
			name: "Test mid sites tree: looks for s1 -> gets [r1,s1]",
			params: &api.LocationServiceListLocationsParams{
				Name:      &utils.Site1Name,
				ShowSites: &showSites,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *s1.ResourceId,
					Name:       *s1.Name,
					ParentId:   *r1.ResourceId,
					Type:       siteKind,
				},
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  1,
			outputElements: 1,
		},
		{
			name: "Test mid sites tree: looks for s2 -> gets [r1,r2,s2]",
			params: &api.LocationServiceListLocationsParams{
				Name:      &utils.Site2Name,
				ShowSites: &showSites,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *s2.ResourceId,
					Name:       *s2.Name,
					ParentId:   *r2.ResourceId,
					Type:       siteKind,
				},
				{
					ResourceId: *r2.ResourceId,
					Name:       *r2.Name,
					ParentId:   *r1.ResourceId,
					Type:       regionKind,
				},
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  1,
			outputElements: 1,
		},
		{
			name: "Test site and region tree: looks for common name -> gets [r1,s1] and totalElements 2",
			params: &api.LocationServiceListLocationsParams{
				Name:        &commonSiteRegionSuffix,
				ShowSites:   &showSites,
				ShowRegions: &showRegions,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *s1.ResourceId,
					Name:       *s1.Name,
					ParentId:   *r1.ResourceId,
					Type:       siteKind,
				},
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  2,
			outputElements: 2,
		},
		{
			name: "Test site and region tree: looks for common name -> gets [r1,r2,s1,s2,s3] and totalElements 5",
			params: &api.LocationServiceListLocationsParams{
				Name:        &commonManySuffix,
				ShowSites:   &showSites,
				ShowRegions: &showRegions,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *s1.ResourceId,
					Name:       *s1.Name,
					ParentId:   *r1.ResourceId,
					Type:       siteKind,
				},
				{
					ResourceId: *s2.ResourceId,
					Name:       *s2.Name,
					ParentId:   *r2.ResourceId,
					Type:       siteKind,
				},
				{
					ResourceId: *s3.ResourceId,
					Name:       *s3.Name,
					ParentId:   emptyParentID,
					Type:       siteKind,
				},
				{
					ResourceId: *r2.ResourceId,
					Name:       *r2.Name,
					ParentId:   *r1.ResourceId,
					Type:       regionKind,
				},
				{
					ResourceId: *r1.ResourceId,
					Name:       *r1.Name,
					ParentId:   emptyParentID,
					Type:       regionKind,
				},
			},
			totalElements:  5,
			outputElements: 5,
		},
		{
			name: "Test leaf sites",
			params: &api.LocationServiceListLocationsParams{
				Name:      &utils.Site3Name,
				ShowSites: &showSites,
			},
			expected: []api.ListLocationsResponseLocationNode{
				{
					ResourceId: *s3.ResourceId,
					Name:       *s3.Name,
					ParentId:   emptyParentID,
					Type:       siteKind,
				},
			},
			totalElements:  1,
			outputElements: 1,
		},
		{
			name: "Test empty/unknown site",
			params: &api.LocationServiceListLocationsParams{
				Name:      &utils.SiteUnexistID,
				ShowSites: &showSites,
			},
			expected:       []api.ListLocationsResponseLocationNode{},
			totalElements:  0,
			outputElements: 0,
		},
		{
			name: "Test empty/unknown region",
			params: &api.LocationServiceListLocationsParams{
				Name:        &utils.RegionUnexistID,
				ShowRegions: &showRegions,
			},
			expected:       []api.ListLocationsResponseLocationNode{},
			totalElements:  0,
			outputElements: 0,
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			getlocResponse, err := apiClient.LocationServiceListLocationsWithResponse(
				ctx, projectName, tcase.params, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
			require.NoError(t, err)
			respStatusCode := getlocResponse.StatusCode()

			require.Equal(t, http.StatusOK, respStatusCode)
			assert.EqualValues(t, tcase.expected, getlocResponse.JSON200.Nodes)
			assert.EqualValues(t, tcase.totalElements, *getlocResponse.JSON200.TotalElements)
			assert.EqualValues(t, tcase.outputElements, *getlocResponse.JSON200.OutputElements)
		})
	}
}

func TestLocation_LargeHierarchy(t *testing.T) {
	log.Info().Msgf("Begin TestLocation_Hierarchy")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout*5)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	testCases := []testCase{
		{
			name: "Test root regions",
			params: &api.LocationServiceListLocationsParams{
				Name:        &regionPrefixName,
				ShowRegions: &showRegions,
			},

			totalElements:  10,
			outputElements: 10,
			listedElements: 10,
		},
		{
			name: "Test sub regions",
			params: &api.LocationServiceListLocationsParams{
				Name:        &subRegionPrefixName,
				ShowRegions: &showRegions,
			},

			totalElements:  100,
			outputElements: 50,
			listedElements: 55,
		},
		{
			name: "Test sites",
			params: &api.LocationServiceListLocationsParams{
				Name:      &sitePrefixName,
				ShowSites: &showSites,
			},

			totalElements:   1000,
			outputElements:  50,
			listedElements:  105, // 5 root regions, 50 sub regions (10/root), 1 site/subregion
			allowMoreListed: true,
		},
		{
			name: "Test subregions and sites - contain the same prefix",
			params: &api.LocationServiceListLocationsParams{
				Name:        &subRegionPrefixName,
				ShowSites:   &showSites,
				ShowRegions: &showRegions,
			},

			totalElements:   1100,
			outputElements:  100,
			listedElements:  105, // 5 root regions, 50 sub regions (10/root), 1 site/subregion
			allowMoreListed: true,
		},
	}

	baselineCounts := map[string]struct {
		total  int
		output int
		listed int
	}{}
	for _, tcase := range testCases {
		getlocResponse, err := apiClient.LocationServiceListLocationsWithResponse(
			ctx, projectName, tcase.params, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, getlocResponse.StatusCode())
		baselineCounts[tcase.name] = struct {
			total  int
			output int
			listed int
		}{
			total:  int(*getlocResponse.JSON200.TotalElements),
			output: int(*getlocResponse.JSON200.OutputElements),
			listed: len(getlocResponse.JSON200.Nodes),
		}
	}

	setupRegionSiteLargeHierarchy(ctx, t, apiClient)
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			getlocResponse, err := apiClient.LocationServiceListLocationsWithResponse(
				ctx, projectName, tcase.params, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
			require.NoError(t, err)
			respStatusCode := getlocResponse.StatusCode()
			require.Equal(t, http.StatusOK, respStatusCode)
			baseline := baselineCounts[tcase.name]
			assert.EqualValues(t, tcase.totalElements+baseline.total, *getlocResponse.JSON200.TotalElements)
			assert.EqualValues(t, tcase.outputElements+baseline.output, *getlocResponse.JSON200.OutputElements)
			if tcase.allowMoreListed {
				assert.GreaterOrEqual(t, len(getlocResponse.JSON200.Nodes), tcase.listedElements+baseline.listed)
			} else {
				assert.Equal(t, tcase.listedElements+baseline.listed, len(getlocResponse.JSON200.Nodes))
			}
		})
	}
}

func TestLocation_Cleanup(t *testing.T) {
	log.Info().Msgf("Begin TestLocation_Cleanup")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout*5)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	pgSize := 100
	regions, err := apiClient.RegionServiceListRegionsWithResponse(
		ctx,
		projectName,
		&api.RegionServiceListRegionsParams{
			PageSize: &pgSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, regions.StatusCode())

	for _, region := range regions.JSON200.Regions {
		_, err = apiClient.RegionServiceDeleteRegionWithResponse(
			ctx,
			projectName,
			*region.ResourceId,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
	}

	sites, err := apiClient.SiteServiceListSitesWithResponse(
		ctx,
		projectName,
		&api.SiteServiceListSitesParams{
			PageSize: &pgSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, regions.StatusCode())

	for _, site := range sites.JSON200.Sites {
		_, err := apiClient.SiteServiceDeleteSiteWithResponse(
			ctx,
			projectName,
			*site.ResourceId,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, regions.StatusCode())
	}
}
