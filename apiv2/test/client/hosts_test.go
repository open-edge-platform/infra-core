// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostCustom(t *testing.T) {
	log.Info().Msgf("Begin compute host tests")

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Site2Request.RegionId = r1.JSON200.ResourceId
	s2 := CreateSite(t, ctx, apiClient, utils.Site2Request)

	utils.Host1Request.SiteId = s1.JSON200.ResourceId
	utils.Host2Request.SiteId = s1.JSON200.ResourceId

	h1 := CreateHost(t, ctx, apiClient, utils.Host1Request)
	CreateHost(t, ctx, apiClient, utils.Host2Request)

	resHostH1, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostH1.StatusCode())
	assert.Equal(t, utils.Host1Name, resHostH1.JSON200.Name)

	// Change site of Host1 via PUT
	utils.Host1RequestUpdate.SiteId = s2.JSON200.ResourceId
	h1Up, err := apiClient.HostServiceUpdateHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		utils.Host1RequestUpdate,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, h1Up.StatusCode())

	// now check the filter, 1 host in site1, 1 host in site2
	filterBySite := fmt.Sprintf(FilterSiteId, *s1.JSON200.ResourceId)
	res, err := apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{Filter: &filterBySite},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, 1, len(res.JSON200.Hosts))

	filterBySite = fmt.Sprintf(FilterSiteId, *s2.JSON200.ResourceId)
	res, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{Filter: &filterBySite},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, 1, len(res.JSON200.Hosts))

	resHostH1Up, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, utils.Host2Name, resHostH1Up.JSON200.Name)

	// Uses Puts to update Host site with empty string
	utils.Host1RequestUpdate.SiteId = &emptyString
	h1Up, err = apiClient.HostServiceUpdateHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		utils.Host1RequestUpdate,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, h1Up.StatusCode())

	// now check the filter
	filterBySite = fmt.Sprintf(FilterSiteId, *s2.JSON200.ResourceId)
	res, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{Filter: &filterBySite},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, 0, len(res.JSON200.Hosts))

	resHostH1Up, err = apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, utils.Host2Name, resHostH1Up.JSON200.Name)
	assert.Nil(t, resHostH1Up.JSON200.Site)

	// Uses Patch to update host1 site with s2 siteID
	h1PatchRequest := api.HostResource{
		Name:   utils.Host3Name,
		SiteId: s2.JSON200.ResourceId,
	}
	h1Patch, err := apiClient.HostServicePatchHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		nil,
		h1PatchRequest,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, h1Patch.StatusCode())

	resHostH1Patched, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, utils.Host3Name, resHostH1Patched.JSON200.Name)
	assert.Equal(t, *s2.JSON200.ResourceId, *resHostH1Patched.JSON200.Site.ResourceId)

	// Uses Patch to update host1 site with s2 siteID
	h1PatchRequest.SiteId = &emptyString
	h1Patch, err = apiClient.HostServicePatchHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		nil,
		h1PatchRequest,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, h1Patch.StatusCode())

	resHostH1Patched, err = apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, utils.Host3Name, resHostH1Patched.JSON200.Name)
	assert.Empty(t, resHostH1Patched.JSON200.Site)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEONBOARDED, *resHostH1Patched.JSON200.DesiredState)

	// Expect BadRequest errors in Patch/Put with emptyString wrong
	utils.Host1RequestUpdate.SiteId = &emptyStringWrong
	h1Up, err = apiClient.HostServiceUpdateHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		utils.Host1RequestUpdate,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, h1Up.StatusCode())

	// cleanup
	utils.Host1RequestPatch.Site = nil
	utils.Host1RequestUpdate.Site = nil

	log.Info().Msgf("End compute host tests")
}

func TestHostSites(t *testing.T) {
	log.Info().Msgf("Begin compute host tests")

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = s1.JSON200.ResourceId
	utils.Host2Request.SiteId = nil

	h1 := CreateHost(t, ctx, apiClient, utils.Host1Request)
	CreateHost(t, ctx, apiClient, utils.Host2Request)

	// now check the filter
	filterBySite := fmt.Sprintf(FilterSiteId, *s1.JSON200.ResourceId)
	res, err := apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{Filter: &filterBySite},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, 1, len(res.JSON200.Hosts))

	res, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &FilterNotHasSite,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, 1, len(res.JSON200.Hosts))

	resHostH1, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostH1.StatusCode())
	assert.Equal(t, utils.Host1Name, resHostH1.JSON200.Name)

	log.Info().Msgf("End compute host tests")
}

func TestHostErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Get_UnexistID_Status_StatusNotFound", func(t *testing.T) {
		resHost, err := apiClient.HostServiceGetHostWithResponse(
			ctx,
			utils.HostUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resHost.StatusCode())
	})

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resHost, err := apiClient.HostServiceUpdateHostWithResponse(
			ctx,
			utils.HostUnexistID,
			utils.Host1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resHost.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resHost, err := apiClient.HostServiceDeleteHostWithResponse(
			ctx,
			utils.HostUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resHost.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		resHost, err := apiClient.HostServiceGetHostWithResponse(
			ctx,
			utils.HostWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resHost.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		resHost, err := apiClient.HostServiceUpdateHostWithResponse(
			ctx,
			utils.HostWrongID,
			utils.Host1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resHost.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		resHost, err := apiClient.HostServiceDeleteHostWithResponse(
			ctx,
			utils.HostWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resHost.StatusCode())
	})

	t.Run("Post_NonPrintable_Status_StatusBadRequest", func(t *testing.T) {
		resHost, err := apiClient.HostServiceCreateHostWithResponse(
			ctx,
			utils.HostNonPrintable,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resHost.StatusCode())
	})
}

func TestHostList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Host1Request.SiteId = s1.JSON200.ResourceId

	totalItems := 10
	var pageId int32 = 1
	var pageSize int32 = 4

	for id := 0; id < totalItems; id++ {
		h := GetHostRequestWithRandomUUID()
		CreateHost(t, ctx, apiClient, h)
	}

	resList, err := apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Offset:   &pageId,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Hosts), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	// Use default page size (20)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func BenchmarkHostList(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	b.Cleanup(func() { cancel() })

	apiClient, err := GetAPIClient()
	assert.NoError(b, err)

	r1 := CreateRegion(b, ctx, apiClient, utils.Region1Request)
	assert.Equal(b, utils.Region1Name, *r1.JSON200.Name)
	b.Cleanup(func() { DeleteRegion(b, ctx, apiClient, *r1.JSON200.ResourceId) })

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(b, ctx, apiClient, utils.Site1Request)
	b.Cleanup(func() { DeleteSite(b, ctx, apiClient, *s1.JSON200.ResourceId) })

	utils.Host1Request.SiteId = s1.JSON200.ResourceId

	// this is the shakeup run
	benchmarkHosts(b, 5, apiClient, ctx)

	// Loop for different number of hosts.
	for _, i := range []int{10, 50, 100, 250} {
		b.Run(fmt.Sprintf("Hosts%d", i), func(b *testing.B) {
			benchmarkHosts(b, i, apiClient, ctx)
		})
	}
}

func benchmarkHosts(b *testing.B, nHosts int,
	apiClient *api.ClientWithResponses, ctx context.Context,
) {
	b.Helper()

	// Emulate the request of the GUI
	var pageId int32 = 1
	var pageSize int32 = 100

	for id := 0; id < nHosts; id++ {
		postResp := CreateHost(b, ctx, apiClient, utils.Host1Request)
		b.Cleanup(func() { SoftDeleteHost(b, ctx, apiClient, postResp.JSON200) })
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resList, err := apiClient.HostServiceListHostsWithResponse(
			ctx,
			&api.HostServiceListHostsParams{
				Offset:   &pageId,
				PageSize: &pageSize,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		assert.NoError(b, err)
		assert.Equal(b, http.StatusOK, resList.StatusCode())
	}
	b.StopTimer()
}

func TestHostListFilterMetadata(t *testing.T) {
	// initializing context
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// creating host #1
	postResp := CreateHost(t, ctx, apiClient, utils.HostReqFilterMeta1)
	hID1 := *postResp.JSON200.ResourceId

	// creating host #2
	postResp = CreateHost(t, ctx, apiClient, utils.HostReqFilterMeta2)
	hID2 := *postResp.JSON200.ResourceId

	// obtaining host with Metadata Key=filtermetakey1 and Value=filtermetavalue1
	filterMetadata := fmt.Sprintf(FilterByMetadata, `{"key":"filtermetakey1","value":"filtermetavalue1"}`)
	resList, err := apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &filterMetadata,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Hosts), 2)
	assert.True(t, hostsContainsId(resList.JSON200.Hosts, hID1))
	assert.True(t, hostsContainsId(resList.JSON200.Hosts, hID2))

	// obtaining host with Metadata Key=filtermetakey2 and Value=filtermetavalue2
	filterMetadata = fmt.Sprintf(FilterByMetadata, `{"key":"filtermetakey2","value":"filtermetavalue2"}`)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &filterMetadata,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Hosts), 1)
	assert.True(t, hostsContainsId(resList.JSON200.Hosts, hID1))
	assert.False(t, hostsContainsId(resList.JSON200.Hosts, hID2))

	// obtaining host with Metadata Key=filtermetakey2 and Value=filtermetavalue2_mod
	filterMetadata = fmt.Sprintf(FilterByMetadata, `{"key":"filtermetakey2","value":"filtermetavalue2_mod"}`)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &filterMetadata,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Hosts), 1)
	assert.False(t, hostsContainsId(resList.JSON200.Hosts, hID1))
	assert.True(t, hostsContainsId(resList.JSON200.Hosts, hID2))

	// obtaining host with Metadata from Host1
	// reqMetadataJoin := []string{"filtermetakey1=filtermetavalue1", "filtermetakey2=filtermetavalue2"}
	filterMetadata = fmt.Sprintf(FilterByMetadata,
		`{"key":"filtermetakey1","value":"filtermetavalue1"},{"key":"filtermetakey2","value":"filtermetavalue2"}`)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &filterMetadata,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Hosts), 1)
	assert.True(t, hostsContainsId(resList.JSON200.Hosts, hID1))

	// obtaining host with Metadata from Host2
	// reqMetadataJoin = []string{"filtermetakey1=filtermetavalue1", "filtermetakey2=filtermetavalue2_mod"}
	filterMetadata = fmt.Sprintf(FilterByMetadata,
		`{"key":"filtermetakey1","value":"filtermetavalue1"},{"key":"filtermetakey2","value":"filtermetavalue2_mod"}`)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &filterMetadata,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Hosts), 1)
	assert.True(t, hostsContainsId(resList.JSON200.Hosts, hID2))

	// Look for a host with wrong metadata
	// reqMetadata4 := []string{"randomKey=randomValue"}
	filterMetadata = fmt.Sprintf(FilterByMetadata, `{"key":"randomKey","value":"randomValue"}`)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &filterMetadata,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Empty(t, resList.JSON200.Hosts)
}

func TestHostListFilterUUID(t *testing.T) {
	// initializing context
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Site2Request.RegionId = r1.JSON200.ResourceId
	s2 := CreateSite(t, ctx, apiClient, utils.Site2Request)

	utils.Host1Request.SiteId = s1.JSON200.ResourceId
	utils.Host2Request.SiteId = s1.JSON200.ResourceId

	// creating host #1
	CreateHost(t, ctx, apiClient, utils.Host1Request)

	// creating host #2
	CreateHost(t, ctx, apiClient, utils.Host2Request)

	metadata := []api.MetadataItem{
		{Key: "k", Value: "v"},
	}
	// creating host #3
	CreateHost(t, ctx, apiClient, api.HostResource{
		Name:     utils.Host3Name,
		SiteId:   s2.JSON200.ResourceId,
		Metadata: &metadata,
		Uuid:     &utils.Host3UUID,
	})

	// obtaining host with Device GUID#1
	guidFind1 := utils.Host1UUID1
	byUUIDFilter := fmt.Sprintf(FilterUUID, guidFind1)
	resList, err := apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &byUUIDFilter,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, len(resList.JSON200.Hosts))
	assert.Equal(t, false, resList.JSON200.HasNext)

	// obtaining host with Device GUID #2
	guidFind2 := utils.Host2UUID
	byUUIDFilter = fmt.Sprintf(FilterUUID, guidFind2)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &byUUIDFilter,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Hosts), 1)
	assert.Equal(t, false, resList.JSON200.HasNext)

	var largePageSize int32 = 100
	// Look for all hosts
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			PageSize: &largePageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Hosts), 3)
	assert.Equal(t, false, resList.JSON200.HasNext)

	// Look for an unexistent host
	guidFindUnexists := utils.HostUUIDUnexists
	byUUIDFilter = fmt.Sprintf(FilterUUID, guidFindUnexists)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &byUUIDFilter,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Empty(t, resList.JSON200.Hosts)

	// Look for a host with wrong UUID - utils.HostUUIDError
	byUUIDFilter = fmt.Sprintf(FilterUUID, utils.HostUUIDError)
	resList, err = apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{
			Filter: &byUUIDFilter,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Empty(t, resList.JSON200.Hosts)
}

func TestHostInvalidate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	_, _, h4 := loadZTPTest(t, ctx, apiClient, &utils.Region1Request,
		&utils.Site1Request, &utils.Host4Request)
	require.NoError(t, err)

	_, err = apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h4.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)

	note := "host is lost"
	_, err = apiClient.HostServiceInvalidateHostWithResponse(
		ctx,
		*h4.JSON200.ResourceId,
		&api.HostServiceInvalidateHostParams{
			Note: &note,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)

	// TODO: wait for condition instead of sleep()
	time.Sleep(3 * time.Second)

	res, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h4.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, api.HostResourceCurrentStateHOSTSTATEUNTRUSTED, *res.JSON200.CurrentState)
	assert.Equal(t, "Invalidated", *res.JSON200.HostStatus)
	assert.Equal(t, note, *res.JSON200.Note)
}

func loadZTPTest(t *testing.T, ctx context.Context, apiClient *api.ClientWithResponses,
	regionRequest *api.RegionResource, siteRequest *api.SiteResource, hostRequest *api.HostResource) (
	*api.RegionServiceCreateRegionResponse, *api.SiteServiceCreateSiteResponse, *api.HostServiceCreateHostResponse,
) {
	reg := CreateRegion(t, ctx, apiClient, *regionRequest)
	assert.Equal(t, regionRequest.Name, reg.JSON200.Name)

	siteRequest.RegionId = reg.JSON200.ResourceId
	sit := CreateSite(t, ctx, apiClient, *siteRequest)
	assert.Equal(t, siteRequest.Name, sit.JSON200.Name)

	// No site defined
	hos := CreateHost(t, ctx, apiClient, *hostRequest)

	resH, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*hos.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resH.StatusCode())
	assert.Equal(t, hostRequest.Name, resH.JSON200.Name)
	assert.Equal(t, *hostRequest.Uuid, *resH.JSON200.Uuid)

	return reg, sit, hos
}

// Test main workflow for ZTP using PUT
func TestHostZTPWithPut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	_, s1, h4 := loadZTPTest(t, ctx, apiClient, &utils.Region1Request,
		&utils.Site1Request, &utils.Host4Request)
	require.NoError(t, err)

	// Simulate ZTP with PUT
	utils.Host4RequestPut.SiteId = s1.JSON200.ResourceId
	h4Put, err := apiClient.HostServiceUpdateHostWithResponse(
		ctx,
		*h4.JSON200.ResourceId,
		utils.Host4RequestPut,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, h4Put.StatusCode())

	// now check the filter
	UUID := utils.Host4UUID1
	byUUIDFilter := fmt.Sprintf(FilterUUID, UUID)
	res, err := apiClient.HostServiceListHostsWithResponse(
		ctx,
		&api.HostServiceListHostsParams{Filter: &byUUIDFilter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, 1, len(res.JSON200.Hosts))

	resHostH4Put, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h4.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, utils.Host4UUID1, *resHostH4Put.JSON200.Uuid)
}

func TestHostsSummary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	r1 := CreateRegion(t, ctx, apiClient, utils.Region1Request)
	assert.Equal(t, utils.Region1Name, *r1.JSON200.Name)

	utils.Site1Request.RegionId = r1.JSON200.ResourceId
	s1 := CreateSite(t, ctx, apiClient, utils.Site1Request)

	utils.Site2Request.RegionId = r1.JSON200.ResourceId
	CreateSite(t, ctx, apiClient, utils.Site2Request)

	utils.Host1Request.SiteId = s1.JSON200.ResourceId
	utils.Host2Request.SiteId = nil

	h1 := CreateHost(t, ctx, apiClient, utils.Host1Request)
	CreateHost(t, ctx, apiClient, utils.Host2Request)

	resHostH1, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*h1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostH1.StatusCode())
	assert.Equal(t, utils.Host1Name, resHostH1.JSON200.Name)

	resHostSummary, err := apiClient.HostServiceGetHostsSummaryWithResponse(
		ctx,
		&api.HostServiceGetHostsSummaryParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostSummary.StatusCode())
	assert.GreaterOrEqual(t, int(*resHostSummary.JSON200.Total), 2)
	if resHostSummary.JSON200.Error != nil {
		assert.GreaterOrEqual(t, *resHostSummary.JSON200.Error, uint32(0))
	}
	if resHostSummary.JSON200.Running != nil {
		assert.GreaterOrEqual(t, *resHostSummary.JSON200.Running, uint32(0))
	}
	assert.GreaterOrEqual(t, *resHostSummary.JSON200.Unallocated, uint32(1))
}

func TestHostRegister(t *testing.T) {
	log.Info().Msgf("Begin compute host register tests")
	var registeredHosts []*api.HostServiceRegisterHostResponse

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// register host using UUID & SN
	registeredHost1, err := apiClient.HostServiceRegisterHostWithResponse(
		ctx,
		nil,
		utils.HostRegister,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	registeredHosts = append(registeredHosts, registeredHost1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, registeredHost1.StatusCode())
	assert.Equal(t, *utils.HostRegister.Uuid, *registeredHost1.JSON200.Uuid)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEREGISTERED, *registeredHost1.JSON200.DesiredState)

	// change registered host name - via Patch
	resHostRegisterPatch, err := apiClient.HostServiceRegisterUpdateHostWithResponse(
		ctx,
		*registeredHost1.JSON200.ResourceId,
		utils.HostRegisterPatch,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostRegisterPatch.StatusCode())

	// get the patched host and verify name is updated
	resHostGet, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*registeredHost1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, *utils.HostRegisterPatch.Name, resHostGet.JSON200.Name)

	// change name & autoOnboard=true for registered host - via Patch
	resHostRegisterPatch, err = apiClient.HostServiceRegisterUpdateHostWithResponse(
		ctx,
		*registeredHost1.JSON200.ResourceId,
		api.HostRegister{
			Name:        &utils.Host2bName,
			AutoOnboard: &utils.AutoOnboardTrue,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostRegisterPatch.StatusCode())

	// get the patched host and verify desiredState is updated
	resHostGet, err = apiClient.HostServiceGetHostWithResponse(
		ctx,
		*registeredHost1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostGet.StatusCode())
	assert.Equal(t, utils.Host2bName, resHostGet.JSON200.Name)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEONBOARDED, *resHostGet.JSON200.DesiredState)

	// change autoOnboard=false only for registered host - via Patch
	resHostRegisterPatch, err = apiClient.HostServiceRegisterUpdateHostWithResponse(
		ctx,
		*registeredHost1.JSON200.ResourceId,
		api.HostRegister{
			AutoOnboard: &utils.AutoOnboardFalse,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostRegisterPatch.StatusCode())

	// get the patched host and verify desiredState is updated
	resHostGet, err = apiClient.HostServiceGetHostWithResponse(
		ctx,
		*registeredHost1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostGet.StatusCode())
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEREGISTERED, *resHostGet.JSON200.DesiredState)

	// register host with autoOnboard=true
	registeredHost2, err := apiClient.HostServiceRegisterHostWithResponse(
		ctx,
		nil,
		utils.HostRegisterAutoOnboard,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	registeredHosts = append(registeredHosts, registeredHost2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, registeredHost2.StatusCode())
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEONBOARDED, *registeredHost2.JSON200.DesiredState)

	// change name & autoOnboard=false for registered host - via Patch
	resHostRegisterPatch, err = apiClient.HostServiceRegisterUpdateHostWithResponse(
		ctx,
		*registeredHost2.JSON200.ResourceId,
		api.HostRegister{
			Name:        &utils.Host1Name,
			AutoOnboard: &utils.AutoOnboardFalse,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostRegisterPatch.StatusCode())

	// get the patched host and verify desiredState is updated
	resHostGet, err = apiClient.HostServiceGetHostWithResponse(
		ctx,
		*registeredHost2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHostGet.StatusCode())
	assert.Equal(t, utils.Host1Name, resHostGet.JSON200.Name)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEREGISTERED, *resHostGet.JSON200.DesiredState)

	// register host using UUID only
	registeredHost3, err := apiClient.HostServiceRegisterHostWithResponse(
		ctx,
		nil,
		api.HostRegister{
			Uuid: &utils.Host5UUID,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	registeredHosts = append(registeredHosts, registeredHost3)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, registeredHost3.StatusCode())
	assert.Equal(t, utils.Host5UUID, *registeredHost3.JSON200.Uuid)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEREGISTERED, *registeredHost3.JSON200.DesiredState)

	// register host using SN only
	registeredHost4, err := apiClient.HostServiceRegisterHostWithResponse(
		ctx,
		nil,
		api.HostRegister{
			SerialNumber: &utils.HostSerialNumber3,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	registeredHosts = append(registeredHosts, registeredHost4)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, registeredHost4.StatusCode())
	assert.Equal(t, utils.HostSerialNumber3, *registeredHost4.JSON200.SerialNumber)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEREGISTERED, *registeredHost4.JSON200.DesiredState)

	// invalid register command - no UUID, no SN
	resHostRegisterInv, err := apiClient.HostServiceRegisterHostWithResponse(
		ctx,
		nil,
		api.HostRegister{Name: &utils.Host4Name},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resHostRegisterInv.StatusCode())

	// delete the registered hosts
	for _, host := range registeredHosts {
		resHost, err := apiClient.HostServiceDeleteHostWithResponse(
			ctx,
			*host.JSON200.ResourceId,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resHost.StatusCode())
	}

	log.Info().Msgf("End compute host register tests")
}

func TestHostOnboard(t *testing.T) {
	log.Info().Msgf("Begin compute host onboard tests")

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// register host using UUID & SN
	registeredHost, err := apiClient.HostServiceRegisterHostWithResponse(
		ctx,
		nil,
		utils.HostRegister,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, registeredHost.StatusCode())
	assert.Equal(t, *utils.HostRegister.Uuid, *registeredHost.JSON200.Uuid)
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEREGISTERED, *registeredHost.JSON200.DesiredState)

	// onboard host
	resOnboard, err := apiClient.HostServiceOnboardHostWithResponse(
		ctx,
		*registeredHost.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resOnboard.StatusCode())

	// get the onboarded host and verify the desiredState is updated
	onboardedHost, err := apiClient.HostServiceGetHostWithResponse(
		ctx,
		*registeredHost.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, onboardedHost.StatusCode())
	assert.Equal(t, api.HostResourceDesiredStateHOSTSTATEONBOARDED, *onboardedHost.JSON200.DesiredState)

	// delete the onboarded host
	resHost, err := apiClient.HostServiceDeleteHostWithResponse(
		ctx,
		*registeredHost.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resHost.StatusCode())

	log.Info().Msgf("End compute host onboard tests")
}
