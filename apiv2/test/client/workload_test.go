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

	h1 := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())
	h2 := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())
	h3 := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = h1.JSON200.ResourceId
	i1 := CreateInstance(t, ctx, apiClient, utils.Instance1Request)
	i1ID := *i1.JSON200.ResourceId

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = h2.JSON200.ResourceId
	i2 := CreateInstance(t, ctx, apiClient, utils.Instance1Request)
	i2ID := *i2.JSON200.ResourceId

	utils.Instance1Request.OsID = os.JSON200.ResourceId
	utils.Instance1Request.HostID = h3.JSON200.ResourceId
	i3 := CreateInstance(t, ctx, apiClient, utils.Instance1Request)
	i3ID := *i3.JSON200.ResourceId

	w1 := CreateWorkload(t, ctx, apiClient, utils.WorkloadCluster1Request)
	w1ID := *w1.JSON200.ResourceId
	w2 := CreateWorkload(t, ctx, apiClient, utils.WorkloadCluster2Request)
	w2ID := *w2.JSON200.ResourceId

	// Create workload member (associate workload to hosts)
	wmKind := api.WORKLOADMEMBERKINDCLUSTERNODE
	m1w1 := CreateWorkloadMember(t, ctx, apiClient, api.WorkloadMember{
		InstanceId: &i1ID,
		WorkloadId: &w1ID,
		Kind:       wmKind,
	})
	m2w1 := CreateWorkloadMember(t, ctx, apiClient, api.WorkloadMember{
		InstanceId: &i2ID,
		WorkloadId: &w1ID,
		Kind:       wmKind,
	})
	m1w2 := CreateWorkloadMember(t, ctx, apiClient, api.WorkloadMember{
		InstanceId: &i3ID,
		WorkloadId: &w2ID,
		Kind:       wmKind,
	})

	// Assert presence of workload with expected members
	getw1, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
		ctx,
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
		*m1w1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getm1w1.StatusCode())
	assert.Equal(t, w1ID, *getm1w1.JSON200.Workload.ResourceId)
	assert.Equal(t, i1ID, *getm1w1.JSON200.Member.InstanceID)

	getm2w1, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
		ctx,
		*m2w1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getm1w1.StatusCode())
	assert.Equal(t, w1ID, *getm2w1.JSON200.Workload.ResourceId)
	assert.Equal(t, i2ID, *getm2w1.JSON200.Member.InstanceID)

	getm1w2, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
		ctx,
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

	w1 := CreateWorkload(t, ctx, apiClient, utils.WorkloadCluster1Request)

	w1Update, err := apiClient.WorkloadServiceUpdateWorkloadWithResponse(
		ctx,
		*w1.JSON200.ResourceId,
		utils.WorkloadCluster2Request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w1Update.StatusCode())

	w1GetUp, err := apiClient.WorkloadServiceGetWorkloadWithResponse(
		ctx,
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
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_NoKind_BadRequest", func(t *testing.T) {
		w1Up, err := apiClient.WorkloadServiceCreateWorkloadWithResponse(
			ctx,
			utils.WorkloadNoKind,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w1Up.StatusCode())
	})

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		w1Up, err := apiClient.WorkloadServiceUpdateWorkloadWithResponse(
			ctx,
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
			utils.WorkloadUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelW, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
			ctx,
			utils.WorkloadUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelW.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusNotFound", func(t *testing.T) {
		w1Up, err := apiClient.WorkloadServiceUpdateWorkloadWithResponse(
			ctx,
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
			utils.WorkloadWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelW, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
			ctx,
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
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	h1 := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())
	h1ID := *h1.JSON200.ResourceId
	w1 := CreateWorkload(t, ctx, apiClient, utils.WorkloadCluster1Request)
	w1ID := *w1.JSON200.ResourceId
	wmKind := api.WORKLOADMEMBERKINDCLUSTERNODE

	t.Run("Post_NoKind_BadRequest", func(t *testing.T) {
		mUp, err := apiClient.WorkloadMemberServiceCreateWorkloadMemberWithResponse(
			ctx,
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
			utils.WorkloadMemberUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, mRes.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelM, err := apiClient.WorkloadMemberServiceDeleteWorkloadMemberWithResponse(
			ctx,
			utils.WorkloadMemberUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelM.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusNotFound", func(t *testing.T) {
		mRes, err := apiClient.WorkloadMemberServiceGetWorkloadMemberWithResponse(
			ctx,
			utils.WorkloadMemberWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, mRes.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelM, err := apiClient.WorkloadMemberServiceDeleteWorkloadMemberWithResponse(
			ctx,
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

	totalItems := 10
	var pageId int32 = 1
	var pageSize int32 = 4

	for id := 0; id < totalItems; id++ {
		CreateWorkload(t, ctx, apiClient, utils.WorkloadCluster2Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		&api.WorkloadServiceListWorkloadsParams{
			Offset:   &pageId,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Workloads), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		&api.WorkloadServiceListWorkloadsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.Workloads))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestWorkloadMemberList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	var pageId int32 = 1
	var pageSize int32 = 4

	workload := CreateWorkload(t, ctx, apiClient, utils.WorkloadCluster1Request)
	os := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	for id := 0; id < totalItems; id++ {
		host := CreateHost(t, ctx, apiClient, GetHostRequestWithRandomUUID())

		utils.Instance1Request.OsID = os.JSON200.ResourceId
		utils.Instance1Request.HostID = host.JSON200.ResourceId
		instance := CreateInstance(t, ctx, apiClient, utils.Instance1Request)

		wmKind := api.WORKLOADMEMBERKINDCLUSTERNODE
		CreateWorkloadMember(t, ctx, apiClient, api.WorkloadMember{
			InstanceId: instance.JSON200.ResourceId,
			WorkloadId: workload.JSON200.ResourceId,
			Kind:       wmKind,
		})
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		&api.WorkloadMemberServiceListWorkloadMembersParams{
			Offset:   &pageId,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.WorkloadMembers), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		&api.WorkloadMemberServiceListWorkloadMembersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.WorkloadMembers))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestWorkloadList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.WorkloadServiceListWorkloadsWithResponse(
		ctx,
		&api.WorkloadServiceListWorkloadsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Empty(t, resList.JSON200.Workloads)
}

func TestWorkloadMemberList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.WorkloadMemberServiceListWorkloadMembersWithResponse(
		ctx,
		&api.WorkloadMemberServiceListWorkloadMembersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Empty(t, resList.JSON200.WorkloadMembers)
}
