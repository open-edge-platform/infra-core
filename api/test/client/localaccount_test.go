// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalAccount_CreateGetDelete(t *testing.T) {

	log.Info().Msgf("Begin LocalAccount tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	localAccount1 := CreateLocalAccount(t, ctx, apiClient, utils.LocalAccount1Request)
	localAccount2 := CreateLocalAccount(t, ctx, apiClient, utils.LocalAccount2Request)
	get1, err := apiClient.GetLocalAccountsLocalAccountIDWithResponse(
		ctx,
		*localAccount1.JSON201.LocalAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, get1.StatusCode())
	require.Equal(t, utils.LocalAccountName1, get1.JSON200.Username)
	require.Equal(t, utils.LocalAccount1Request.SshKey, get1.JSON200.SshKey)
	get2, err := apiClient.GetLocalAccountsLocalAccountIDWithResponse(
		ctx,
		*localAccount2.JSON201.LocalAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, get2.StatusCode())
	require.Equal(t, utils.LocalAccountName2, get2.JSON200.Username)
	require.Equal(t, utils.LocalAccount2Request.SshKey, get2.JSON200.SshKey)
	// Delete the local accounts
	delete1, err := apiClient.DeleteLocalAccountsLocalAccountIDWithResponse(
		ctx,
		*localAccount1.JSON201.LocalAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, delete1.StatusCode())
	delete2, err := apiClient.DeleteLocalAccountsLocalAccountIDWithResponse(
		ctx,
		*localAccount2.JSON201.LocalAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, delete2.StatusCode())
	// Verify the local accounts are deleted
	get1, err = apiClient.GetLocalAccountsLocalAccountIDWithResponse(
		ctx,
		*localAccount1.JSON201.LocalAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, get1.StatusCode())
	get2, err = apiClient.GetLocalAccountsLocalAccountIDWithResponse(
		ctx,
		*localAccount2.JSON201.LocalAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, get2.StatusCode())
	log.Info().Msgf("End LocalAccount tests")
}

func TestLocalAccount_Errors(t *testing.T) {
	log.Info().Msgf("Begin Errors LocalAccount tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_NoUsername_BadRequest", func(t *testing.T) {
		localAccount, err := apiClient.PostLocalAccountsWithResponse(
			ctx,
			utils.LocalAccountNoUsername,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		log.Info().Msgf("Error Username %s", localAccount.Body)
		assert.Equal(t, http.StatusBadRequest, localAccount.StatusCode())
	})

	t.Run("Post_NoSshKey_BadRequest", func(t *testing.T) {
		localAccount, err := apiClient.PostLocalAccountsWithResponse(
			ctx,
			utils.LocalAccountNoSshKey,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		log.Info().Msgf("Error SshKey %s", localAccount.Body)
		assert.Equal(t, http.StatusBadRequest, localAccount.StatusCode())
	})

	t.Run("Get_UnexistID_NotFound", func(t *testing.T) {
		localAccount, err := apiClient.GetLocalAccountsLocalAccountIDWithResponse(
			ctx,
			utils.LocalAccountUnexistID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, localAccount.StatusCode())
	})

	t.Run("Delete_UnexistID_NotFound", func(t *testing.T) {
		localAccount, err := apiClient.DeleteLocalAccountsLocalAccountIDWithResponse(
			ctx,
			utils.LocalAccountUnexistID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, localAccount.StatusCode())
	})

	t.Run("Get_WrongID_BadRequest", func(t *testing.T) {
		localAccount, err := apiClient.GetLocalAccountsLocalAccountIDWithResponse(
			ctx,
			utils.LocalAccountWrongID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, localAccount.StatusCode())
	})

	t.Run("Delete_WrongID_BadRequest", func(t *testing.T) {
		localAccount, err := apiClient.DeleteLocalAccountsLocalAccountIDWithResponse(
			ctx,
			utils.LocalAccountWrongID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, localAccount.StatusCode())
	})
	log.Info().Msgf("End Errors LocalAccount tests")
}

func TestLocalAccountList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	pageId := 1
	pageSize := 4

	for id := 0; id < totalItems; id++ {
		CreateLocalAccount(t, ctx, apiClient, utils.LocalAccount1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.GetLocalAccountsWithResponse(
		ctx,
		&api.GetLocalAccountsParams{
			Offset:   &pageId,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(*resList.JSON200.LocalAccounts), pageSize)
	assert.Equal(t, true, *resList.JSON200.HasNext)

	resList, err = apiClient.GetLocalAccountsWithResponse(
		ctx,
		&api.GetLocalAccountsParams{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(*resList.JSON200.LocalAccounts))
	assert.Equal(t, false, *resList.JSON200.HasNext)
}

func TestList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.GetLocalAccountsWithResponse(
		ctx,
		&api.GetLocalAccountsParams{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 0, len(*resList.JSON200.LocalAccounts))
}
