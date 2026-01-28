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

func assertSameMemberIDs(t *testing.T, expectedMembers, actualMembers []api.WorkloadMember) {
	t.Helper()

	assert.Equal(t, len(expectedMembers), len(actualMembers))

	expectedIDs := make([]string, 0, len(expectedMembers))
	for _, em := range expectedMembers {
		if em.Member != nil && em.Member.ResourceId != nil {
			expectedIDs = append(expectedIDs, *em.Member.ResourceId)
		}
	}

	actualIDs := make([]string, 0, len(actualMembers))
	for _, am := range actualMembers {
		if am.Member != nil && am.Member.ResourceId != nil {
			actualIDs = append(actualIDs, *am.Member.ResourceId)
		}
	}

	assert.ElementsMatch(t, expectedIDs, actualIDs)
}

func TestWorkload_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin workload CRUD tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	h1 := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())
	h2 := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())
	h3 := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())
	os := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = h1.JSON200.ResourceId
	utils.Instance1Request.OsUpdatePolicyID = nil // Clear any OS update policy from previous tests
	i1 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)
	require.NotNil(t, i1.JSON200, "Instance i1 creation returned nil JSON200")
	require.NotNil(t, i1.JSON200.ResourceId, "Instance i1 creation returned nil ResourceId")
	i1ID := *i1.JSON200.ResourceId

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = h2.JSON200.ResourceId
	utils.Instance1Request.OsUpdatePolicyID = nil // Clear any OS update policy from previous tests
	i2 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)
	require.NotNil(t, i2.JSON200, "Instance i2 creation returned nil JSON200")
	require.NotNil(t, i2.JSON200.ResourceId, "Instance i2 creation returned nil ResourceId")
	i2ID := *i2.JSON200.ResourceId

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = h3.JSON200.ResourceId
	utils.Instance1Request.OsUpdatePolicyID = nil // Clear any OS update policy from previous tests
	i3 := CreateInstance(ctx, t, apiClient, utils.Instance1Request)
	require.NotNil(t, i3.JSON200, "Instance i3 creation returned nil JSON200")
	require.NotNil(t, i3.JSON200.ResourceId, "Instance i3 creation returned nil ResourceId")
	i3ID := *i3.JSON200.ResourceId

	w1 := CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster1Request)
	w1ID := *w1.JSON200.ResourceId
	w2 := CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster2Request)
	w2ID := *w2.JSON200.ResourceId

	// Create workload member (associate workload to hosts)
	wmKind := api.WORKLOADMEMBERKINDCLUSTERNODE
	m1w1 := CreateWorkloadMember(ctx, t, apiClient, api.WorkloadMember{
		InstanceId: &i1ID,
		WorkloadId: &w1ID,
		Kind:       wmKind,
	})
	m2w1 := CreateWorkloadMember(ctx, t, apiClient, api.WorkloadMember{
		InstanceId: &i2ID,
		WorkloadId: &w1ID,
		Kind:       wmKind,
	})
	m1w2 := CreateWorkloadMember(ctx, t, apiClient, api.WorkloadMember{
		InstanceId: &i3ID,
		WorkloadId: &w2ID,
		Kind:       wmKind,
	})

	// Assert presence of workload with expected members
	getw1, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
		ctx,
		projectName,
		w1ID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getw1.StatusCode())
	assert.Equal(t, utils.WorkloadName1, *getw1.JSON200.Name)
	assert.Equal(t, utils.WorkloadCluster1Request.Kind, getw1.JSON200.Kind)
	assert.NotNil(t, getw1.JSON200.Members)
	assertSameMemberIDs(t, *getw1.JSON200.Members, []api.WorkloadMember{*m1w1.JSON200, *m2w1.JSON200})

	getw2, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
		ctx,
		projectName,
		w2ID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getw2.StatusCode())
	assert.Equal(t, utils.WorkloadName2, *getw2.JSON200.Name)
	assert.Equal(t, utils.WorkloadCluster2Request.Kind, getw2.JSON200.Kind)
	assert.NotNil(t, getw2.JSON200.Members)
	assertSameMemberIDs(t, *getw2.JSON200.Members, []api.WorkloadMember{*m1w2.JSON200})

	// Assert presence of workload members with expected instance and workload
	getm1w1, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
		ctx,
		projectName,
		*m1w1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getm1w1.StatusCode())
	assert.Equal(t, w1ID, *getm1w1.JSON200.Workload.ResourceId)
	assert.Equal(t, i1ID, *getm1w1.JSON200.Member.InstanceID)

	getm2w1, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
		ctx,
		projectName,
		*m2w1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getm1w1.StatusCode())
	assert.Equal(t, w1ID, *getm2w1.JSON200.Workload.ResourceId)
	assert.Equal(t, i2ID, *getm2w1.JSON200.Member.InstanceID)

	getm1w2, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
		ctx,
		projectName,
		*m1w2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getm1w2.StatusCode())
	assert.Equal(t, w2ID, *getm1w2.JSON200.Workload.ResourceId)
	assert.Equal(t, i3ID, *getm1w2.JSON200.Member.InstanceID)

	clearInstanceIDs()

	log.Info().Msgf("End workload CRUD tests")
}

func TestWorkload_UpdatePut(t *testing.T) {
	log.Info().Msgf("Begin Workload Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	w1 := CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster1Request)

	w1Update, err := apiClient.WorkloadServiceUpdateWorkloadWithResponse(
		ctx,
		projectName,
		*w1.JSON200.ResourceId,
		utils.WorkloadCluster2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w1Update.StatusCode())

	w1GetUp, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
		ctx,
		projectName,
		*w1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w1GetUp.StatusCode())
	assert.Equal(t, *utils.WorkloadCluster2Request.Name, *w1GetUp.JSON200.Name)
	assert.Equal(t, utils.WorkloadCluster2Request.Status, w1GetUp.JSON200.Status)

	log.Info().Msgf("End Workload Update tests")
}

func TestWorkload_Errors(t *testing.T) {
	log.Info().Msgf("Begin Workload Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_NoKind_BadRequest", func(t *testing.T) {
		w1Up, err := apiClient.WorkloadServiceCreateWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadNoKind,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w1Up.StatusCode())
	})

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		w1Up, err := apiClient.WorkloadServiceUpdateWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadUnexistID,
			utils.WorkloadCluster1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		w1res, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelW, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelW.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusNotFound", func(t *testing.T) {
		w1Up, err := apiClient.WorkloadServiceUpdateWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadWrongID,
			utils.WorkloadCluster1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusNotFound", func(t *testing.T) {
		w1res, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelW, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
			ctx,
			projectName,
			utils.WorkloadWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelW.StatusCode())
	})
	log.Info().Msgf("End Workload Error tests")
}

func TestWorkloadMember_Errors(t *testing.T) {
	log.Info().Msgf("Begin WorkloadMember Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	h1 := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())
	h1ID := *h1.JSON200.ResourceId
	w1 := CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster1Request)
	w1ID := *w1.JSON200.ResourceId
	wmKind := api.WORKLOADMEMBERKINDCLUSTERNODE

	t.Run("Post_NoKind_BadRequest", func(t *testing.T) {
		mUp, err := apiClient.WorkloadMemberServiceCreateWorkloadMemberWithResponse(
			ctx,
			projectName,
			api.WorkloadMember{
				WorkloadId: &w1ID,
				InstanceId: &h1ID,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, mUp.StatusCode())
	})

	t.Run("Post_NoWorkloadID_BadRequest", func(t *testing.T) {
		mUp, err := apiClient.WorkloadMemberServiceCreateWorkloadMemberWithResponse(
			ctx,
			projectName,
			api.WorkloadMember{
				Kind:       wmKind,
				InstanceId: &h1ID,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, mUp.StatusCode())
	})

	t.Run("Post_NoHostID_BadRequest", func(t *testing.T) {
		mUp, err := apiClient.WorkloadMemberServiceCreateWorkloadMemberWithResponse(
			ctx,
			projectName,
			api.WorkloadMember{
				WorkloadId: &w1ID,
				Kind:       wmKind,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, mUp.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		mRes, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
			ctx,
			projectName,
			utils.WorkloadMemberUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, mRes.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelM, err := apiClient.WorkloadMemberServiceDeleteWorkloadMemberWithResponse(
			ctx,
			projectName,
			utils.WorkloadMemberUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelM.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusNotFound", func(t *testing.T) {
		mRes, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
			ctx,
			projectName,
			utils.WorkloadMemberWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, mRes.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelM, err := apiClient.WorkloadMemberServiceDeleteWorkloadMemberWithResponse(
			ctx,
			projectName,
			utils.WorkloadMemberWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelM.StatusCode())
	})
	log.Info().Msgf("End Workload Member Error tests")
}

func TestWorkloadList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resList, err := apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		projectName,
		&api.WorkloadServiceListWorkloadsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	existingWorkloads := len(resList.JSON200.Workloads)

	totalItems := 10
	pageID := 1
	pageSize := 4

	for id := 0; id < totalItems; id++ {
		CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster2Request)
	}

	// Checks if list resources return expected number of entries
	resList, err = apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		projectName,
		&api.WorkloadServiceListWorkloadsParams{
			Offset:   &pageID,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Workloads), pageSize)
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		projectName,
		&api.WorkloadServiceListWorkloadsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems+existingWorkloads, len(resList.JSON200.Workloads))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestWorkloadMemberList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	totalItems := 10
	pageID := 1
	pageSize := 4

	workload := CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster1Request)
	os := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	resList, err := apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		projectName,
		&api.WorkloadMemberServiceListWorkloadMembersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	existingMembers := len(resList.JSON200.WorkloadMembers)

	for id := 0; id < totalItems; id++ {
		host := CreateHost(ctx, t, apiClient, GetHostRequestWithRandomUUID())

		utils.Instance1Request.OsID = os.JSON200.ResourceId
		utils.Instance1Request.HostID = host.JSON200.ResourceId
		utils.Instance1Request.OsUpdatePolicyID = nil // Clear any OS update policy from previous tests
		instance := CreateInstance(ctx, t, apiClient, utils.Instance1Request)
		require.NotNil(t, instance.JSON200, "Instance creation returned nil JSON200")
		require.NotNil(t, instance.JSON200.ResourceId, "Instance creation returned nil ResourceId")

		wmKind := api.WORKLOADMEMBERKINDCLUSTERNODE
		CreateWorkloadMember(ctx, t, apiClient, api.WorkloadMember{
			InstanceId: instance.JSON200.ResourceId,
			WorkloadId: workload.JSON200.ResourceId,
			Kind:       wmKind,
		})
	}

	// Checks if list resources return expected number of entries
	resList, err = apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		projectName,
		&api.WorkloadMemberServiceListWorkloadMembersParams{
			Offset:   &pageID,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.WorkloadMembers), pageSize)
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		projectName,
		&api.WorkloadMemberServiceListWorkloadMembersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems+existingMembers, len(resList.JSON200.WorkloadMembers))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestWorkloadList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resList, err := apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		projectName,
		&api.WorkloadServiceListWorkloadsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Workloads), 0)
}

func TestWorkloadMemberList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	resList, err := apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		projectName,
		&api.WorkloadMemberServiceListWorkloadMembersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.WorkloadMembers), 0)
}

func TestWorkload_Patch(t *testing.T) {
	log.Info().Msgf("Begin Workload Patch tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	// Create a Workload
	workload := CreateWorkload(ctx, t, apiClient, utils.WorkloadCluster1Request)
	assert.Equal(t, utils.WorkloadName1, *workload.JSON200.Name)

	// Modify fields for patching
	newName := utils.WorkloadName1 + "-updated"
	patchRequest := api.WorkloadResource{
		Name:   &newName,
		Kind:   api.WORKLOADKINDCLUSTER,
		Status: &utils.WorkloadStatus3,
	}

	// Perform the Patch operation
	updatedWorkload, err := apiClient.WorkloadServicePatchWorkloadWithResponse(
		ctx,
		projectName,
		*workload.JSON200.ResourceId,
		&api.WorkloadServicePatchWorkloadParams{},
		patchRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatedWorkload.StatusCode())
	assert.Equal(t, newName, *updatedWorkload.JSON200.Name)
	assert.Equal(t, utils.WorkloadStatus3, *updatedWorkload.JSON200.Status)

	// Verify the changes with a Get operation
	getWorkload, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
		ctx,
		projectName,
		*workload.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getWorkload.StatusCode())
	assert.Equal(t, newName, *getWorkload.JSON200.Name)

	log.Info().Msgf("End Workload Patch tests")
}
