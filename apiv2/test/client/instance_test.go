// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
)

func clearInstanceIDs() {
	utils.Instance1Request.HostID = nil
	utils.Instance2Request.HostID = nil
	utils.Instance1Request.OsID = nil
	utils.Instance2Request.OsID = nil
	utils.Host1Request.SiteId = nil
	utils.Host2Request.SiteId = nil
	utils.Host3Request.SiteId = nil
	utils.Host4Request.SiteId = nil
}

func TestInstance_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin Instance tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	utils.Site1Request.RegionId = nil
	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Host3Request.SiteId = site1.JSON200.SiteID
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host3Request)
	hostCreated2 := CreateHost(ctx, t, apiClient, utils.Host4Request)
	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance2Request.HostID = hostCreated2.JSON200.ResourceId

	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID
	utils.Instance2Request.OsID = osCreated1.JSON200.OsResourceID

	inst1 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)
	inst2 := CreateInstance(ctx, t, apiClient, utils.Instance2Request)

	get1, err := apiClient.InstanceServiceGetInstanceWithResponse(
		ctx,
		projectName,
		*inst1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, *utils.Instance1Request.Name, *get1.JSON200.Name)
	assert.Equal(t, api.INSTANCESTATERUNNING, *get1.JSON200.DesiredState)

	get2, err := apiClient.InstanceServiceGetInstanceWithResponse(
		ctx,
		projectName,
		*inst2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, *utils.Instance2Request.Name, *get2.JSON200.Name)
	assert.Equal(t, *utils.Instance2Request.SecurityFeature, *get2.JSON200.SecurityFeature)

	clearInstanceIDs()
	log.Info().Msgf("End Instance tests")
}

func TestInstance_Update(t *testing.T) {
	log.Info().Msgf("Begin Instance Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	utils.Host1Request.SiteId = nil
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host1Request)
	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID

	inst1 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)
	assert.Equal(t, utils.Inst1Name, *inst1.JSON200.Name)

	newName := utils.Inst1Name + "-mod"
	inst1Mod := api.InstanceResource{
		Name: &newName,
	}
	inst1Up, err := apiClient.InstanceServicePatchInstanceWithResponse(
		ctx,
		projectName,
		*inst1.JSON200.ResourceId,
		nil,
		inst1Mod,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, inst1Up.StatusCode())
	assert.Equal(t, newName, *inst1Up.JSON200.Name)

	inst1Get, err := apiClient.InstanceServiceGetInstanceWithResponse(ctx,
		projectName,
		*inst1.JSON200.ResourceId, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, inst1Get.StatusCode())
	assert.Equal(t, newName, *inst1Get.JSON200.Name)
	assert.Equal(t, *inst1.JSON200.OsID, *inst1Get.JSON200.OsID)

	log.Info().Msgf("End Instance Update tests")
}

func TestInstance_Errors(t *testing.T) {
	log.Info().Msgf("Begin InstanceResource Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}
	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Host3Request.SiteId = site1.JSON200.ResourceId
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host3Request)
	hostCreated2 := CreateHost(ctx, t, apiClient, utils.Host4Request)
	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance2Request.HostID = hostCreated2.JSON200.ResourceId

	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID
	utils.Instance2Request.OsID = osCreated1.JSON200.OsResourceID

	t.Run("Post_NoUpdateSources_Status_BadRequest", func(t *testing.T) {
		utils.InstanceRequestNoOSID.HostID = utils.Instance1Request.HostID // host ID must be provided
		inst1Up, err := apiClient.InstanceServiceCreateInstanceWithResponse(
			ctx,
			projectName,
			utils.InstanceRequestNoOSID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		utils.InstanceRequestNoOSID.HostID = nil // setting Host ID back to original state (see common.go)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, inst1Up.StatusCode())
	})

	t.Run("Post_NoHostL_Status_PreconditionFailed", func(t *testing.T) {
		utils.InstanceRequestNoHostID.HostID = utils.Instance1Request.HostID
		inst1Up, err := apiClient.InstanceServiceCreateInstanceWithResponse(
			ctx,
			projectName,
			utils.InstanceRequestNoHostID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		utils.InstanceRequestNoHostID.HostID = nil
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, inst1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.InstanceServiceGetInstanceWithResponse(
			ctx,
			projectName,
			utils.InstanceUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.InstanceServiceDeleteInstanceWithResponse(
			ctx,
			projectName,
			utils.InstanceUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Get_WrongID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.InstanceServiceGetInstanceWithResponse(
			ctx,
			projectName,
			utils.InstanceWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelSite, err := apiClient.InstanceServiceDeleteInstanceWithResponse(
			ctx,
			projectName,
			utils.InstanceWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})
	clearInstanceIDs()
	log.Info().Msgf("End Instance Error tests")
}

func TestInstanceList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	totalItems := 5
	var offset int
	pageSize := 4

	baselineList, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineList.StatusCode())
	baselineTotalItems := len(baselineList.JSON200.Instances)

	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = site1.JSON200.SiteID
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host1Request)
	hostCreated2 := CreateHost(ctx, t, apiClient, utils.Host2Request)
	host3Name := "Host-Three"
	hostCreated3 := CreateHost(ctx, t, apiClient, api.HostResource{
		Name: host3Name,
		Metadata: &[]api.MetadataItem{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Uuid: &utils.Host3UUID,
	})
	host4Name := "Host-Four"
	hostCreated4 := CreateHost(ctx, t, apiClient, api.HostResource{
		Name: host4Name,
		Metadata: &[]api.MetadataItem{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Uuid: &utils.Host4UUID1,
	})
	host5Name := "Host-Five"
	hostCreated5 := CreateHost(ctx, t, apiClient, api.HostResource{
		Name: host5Name,
		Metadata: &[]api.MetadataItem{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Uuid: &utils.Host5UUID,
	})
	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)
	osCreated2 := CreateOS(ctx, t, apiClient, utils.OSResource2Request)

	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID
	// creating 1st Instance
	CreateInstance(ctx, t, apiClient, utils.Instance1Request)

	// composing request to create 2nd Instance
	utils.Instance2Request.HostID = hostCreated2.JSON200.ResourceId
	utils.Instance2Request.OsID = osCreated1.JSON200.OsResourceID
	// creating 2nd Instance
	CreateInstance(ctx, t, apiClient, utils.Instance2Request)

	// composing request to create 3rd Instance
	utils.Instance2Request.HostID = hostCreated3.JSON200.ResourceId
	utils.Instance2Request.OsID = osCreated2.JSON200.OsResourceID
	// creating 3rd Instance
	CreateInstance(ctx, t, apiClient, utils.Instance2Request)

	// composing request to create 4th Instance
	utils.Instance2Request.HostID = hostCreated4.JSON200.ResourceId
	utils.Instance2Request.OsID = osCreated2.JSON200.OsResourceID
	// creating 4th Instance
	CreateInstance(ctx, t, apiClient, utils.Instance2Request)

	// composing request to create 5th Instance
	utils.Instance2Request.HostID = hostCreated5.JSON200.ResourceId
	utils.Instance2Request.OsID = osCreated2.JSON200.OsResourceID
	// creating 5th Instance
	CreateInstance(ctx, t, apiClient, utils.Instance2Request)

	// Checks if list resources return expected number of entries
	resList, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Instances), pageSize)
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems+baselineTotalItems, len(resList.JSON200.Instances))
	assert.Equal(t, false, resList.JSON200.HasNext)

	clearInstanceIDs()
}

func TestInstanceList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resList, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Instances), 0)
}

func TestInstance_Filter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	baselineWorkloadMemberList, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{
			Filter: &FilterHasWorkloadMember,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineWorkloadMemberList.StatusCode())
	baselineWorkloadMemberTotal := int(baselineWorkloadMemberList.JSON200.TotalElements)

	utils.Site1Request.Region = nil
	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = site1.JSON200.SiteID
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host1Request)
	hostCreated2 := CreateHost(ctx, t, apiClient, utils.Host2Request)

	baselineHostFilter := fmt.Sprintf("host.resourceId=%q", *hostCreated1.JSON200.ResourceId)
	baselineHostList, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{Filter: &baselineHostFilter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineHostList.StatusCode())
	baselineHostTotal := int(baselineHostList.JSON200.TotalElements)

	baselineSiteFilter := fmt.Sprintf("host.site.resourceId=%q", *site1.JSON200.SiteID)
	baselineSiteList, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{Filter: &baselineSiteFilter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, baselineSiteList.StatusCode())
	baselineSiteTotal := int(baselineSiteList.JSON200.TotalElements)

	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID
	inst1 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)

	utils.Instance1Request.HostID = hostCreated2.JSON200.ResourceId
	_ = CreateInstance(ctx, t, apiClient, utils.Instance1Request)

	// filter on Instance->Host->resourceId (host.resourceId="hostId")
	filter := fmt.Sprintf("host.resourceId=%q", *inst1.JSON200.Host.ResourceId)
	assert.Equal(t, *hostCreated1.JSON200.ResourceId, *inst1.JSON200.Host.ResourceId)
	get1, err := apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{Filter: &filter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, baselineHostTotal+1, int(get1.JSON200.TotalElements))

	// filter on Instance->Host->Site->resourceId (host.site.resourceId="siteId")
	filter = fmt.Sprintf("host.site.resourceId=%q", *site1.JSON200.SiteID)
	assert.Equal(t, *hostCreated1.JSON200.Site.ResourceId, *site1.JSON200.SiteID)
	get1, err = apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{Filter: &filter},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, baselineSiteTotal+1, int(get1.JSON200.TotalElements))

	// filter all instances having workload members
	// workloadmemberID := ""
	get1, err = apiClient.InstanceServiceListInstancesWithResponse(
		ctx,
		projectName,
		&api.InstanceServiceListInstancesParams{
			Filter: &FilterHasWorkloadMember,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, baselineWorkloadMemberTotal, int(get1.JSON200.TotalElements))

	// TODO: Re-enable workload member filtering once mock server supports workload members
	//nolint:gocritic // intentionally commented for future re-enabling
	/*
		// filter workloadMember=created ones

		byWorkloadMemberIDFilter := fmt.Sprintf(FilterByWorkloadMemberID, *workloadMember.JSON200.ResourceId)
		get1, err = apiClient.InstanceServiceListInstancesWithResponse(
			ctx,
			projectName,
			&api.InstanceServiceListInstancesParams{Filter: &byWorkloadMemberIDFilter},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, get1.StatusCode())
		assert.Equal(t, 1, int(get1.JSON200.TotalElements))

		// filter workloadMember=
		get1, err = apiClient.InstanceServiceListInstancesWithResponse(
			ctx,
			projectName,
			&api.InstanceServiceListInstancesParams{
				Filter: &FilterHasWorkloadMember,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, get1.StatusCode())
		assert.Equal(t, 1, int(get1.JSON200.TotalElements))

		// filter workloadMember=null
		// workloadmemberID = "null"
		get1, err = apiClient.InstanceServiceListInstancesWithResponse(
			ctx,
			projectName,
			&api.InstanceServiceListInstancesParams{
				Filter: &FilterNotHasWorkloadMember,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, get1.StatusCode())
		assert.Equal(t, 1, int(get1.JSON200.TotalElements))
	*/
}

func TestInstanceInvalidate(t *testing.T) {
	log.Info().Msg("TestInstanceInvalidate Started")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	utils.Site1Request.RegionId = nil
	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Host1Request.SiteId = site1.JSON200.SiteID
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host1Request)
	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID

	inst1 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)

	get1, err := apiClient.InstanceServiceGetInstanceWithResponse(
		ctx,
		projectName,
		*inst1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, *utils.Instance1Request.Name, *get1.JSON200.Name)
	assert.Equal(t, api.INSTANCESTATERUNNING, *get1.JSON200.DesiredState)

	log.Info().Msg("PutInstancesInstanceIDInvalidateWithResponse")
	resp, err := apiClient.InstanceServiceInvalidateInstanceWithResponse(
		ctx,
		projectName,
		*inst1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	if err != nil {
		log.Error().Err(err).Msgf("failed PutInstancesInstanceIDInvalidateWithResponse")
	}
	assert.NoError(t, err)
	if resp != nil {
		log.Info().Msgf("Invalidate response status: %d", resp.StatusCode())
	}

	// TODO: wait for condition instead of sleep()
	time.Sleep(3 * time.Second)

	get2, err := apiClient.InstanceServiceGetInstanceWithResponse(
		ctx,
		projectName,
		*inst1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, *utils.Instance1Request.Name, *get2.JSON200.Name)
	assert.Equal(t, api.INSTANCESTATEUNTRUSTED, *get2.JSON200.DesiredState)
	clearInstanceIDs()

	log.Info().Msg("TestInstanceInvalidate Finished")
}

func TestInstance_CreateWithOSUpdatePolicy(t *testing.T) {
	log.Info().Msgf("Begin Instance with OSUpdatePolicy test")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	utils.Site1Request.RegionId = nil
	site1 := CreateSite(ctx, t, apiClient, utils.Site1Request)
	utils.Host3Request.SiteId = site1.JSON200.SiteID
	hostCreated1 := CreateHost(ctx, t, apiClient, utils.Host3Request)
	osCreated1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	// Create OSUpdatePolicy
	osUpdatePolicy := CreateOsUpdatePolicy(ctx, t, apiClient, utils.OsUpdatePolicyRequest1)

	// Create instance with OSUpdatePolicy assigned
	utils.Instance1Request.HostID = hostCreated1.JSON200.ResourceId
	utils.Instance1Request.OsID = osCreated1.JSON200.OsResourceID
	utils.Instance1Request.OsUpdatePolicyID = osUpdatePolicy.JSON200.ResourceId

	inst1 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)

	// Get the created instance and verify OSUpdatePolicy is assigned
	get1, err := apiClient.InstanceServiceGetInstanceWithResponse(
		ctx,
		projectName,
		*inst1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, *utils.Instance1Request.Name, *get1.JSON200.Name)

	// Check if the OSUpdatePolicy is the one assigned earlier
	require.NotNil(t, get1.JSON200.UpdatePolicy, "OSUpdatePolicy should be assigned")
	assert.Equal(t,
		*osUpdatePolicy.JSON200.ResourceId, *get1.JSON200.UpdatePolicy.ResourceId, "OSUpdatePolicy ResourceId should match")
	assert.Equal(t,
		osUpdatePolicy.JSON200.Name, get1.JSON200.UpdatePolicy.Name, "OSUpdatePolicy Name should match")
	assert.Equal(t,
		osUpdatePolicy.JSON200.Description, get1.JSON200.UpdatePolicy.Description, "OSUpdatePolicy Description should match")

	clearInstanceIDs()
}
