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

func TestLocalAccount_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin CreateGetDelete LocalAccount tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Create LocalAccounts
	account1 := CreateLocalAccount(ctx, t, apiClient, utils.LocalAccount1Request)
	account2 := CreateLocalAccount(ctx, t, apiClient, utils.LocalAccount2Request)

	// Get LocalAccount 1
	get1, err := apiClient.LocalAccountServiceGetLocalAccountWithResponse(
		ctx,
		*account1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.LocalAccount1Request.Username, get1.JSON200.Username)
	assert.Equal(t, utils.LocalAccount1Request.SshKey, get1.JSON200.SshKey)

	// Get LocalAccount 2
	get2, err := apiClient.LocalAccountServiceGetLocalAccountWithResponse(
		ctx,
		*account2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, utils.LocalAccount2Request.Username, get2.JSON200.Username)
	assert.Equal(t, utils.LocalAccount2Request.SshKey, get2.JSON200.SshKey)

	log.Info().Msgf("End CreateGetDelete LocalAccount tests")
}

func TestLocalAccount_Errors(t *testing.T) {
	log.Info().Msgf("Begin Errors LocalAccount tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	t.Run("Post_NoUsername_BadRequest", func(t *testing.T) {
		account, err := apiClient.LocalAccountServiceCreateLocalAccountWithResponse(
			ctx,
			utils.LocalAccountNoName,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, account.StatusCode())
	})

	t.Run("Get_UnexistID_NotFound", func(t *testing.T) {
		account, err := apiClient.LocalAccountServiceGetLocalAccountWithResponse(
			ctx,
			utils.LocalAccountUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, account.StatusCode())
	})

	t.Run("Delete_UnexistID_NotFound", func(t *testing.T) {
		account, err := apiClient.LocalAccountServiceDeleteLocalAccountWithResponse(
			ctx,
			utils.LocalAccountUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, account.StatusCode())
	})

	log.Info().Msgf("End Errors LocalAccount tests")
}

func TestLocalAccountList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 5
	var offset int
	pageSize := 2

	for id := 0; id < totalItems; id++ {
		// Generate sequential usernames
		username := fmt.Sprintf("user%d", id)
		utils.LocalAccount1Request.Username = username
		CreateLocalAccount(ctx, t, apiClient, utils.LocalAccount1Request)
	}

	// Check if list resources return expected number of entries
	resList, err := apiClient.LocalAccountServiceListLocalAccountsWithResponse(
		ctx,
		&api.LocalAccountServiceListLocalAccountsParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, pageSize, len(resList.JSON200.LocalAccounts))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.LocalAccountServiceListLocalAccountsWithResponse(
		ctx,
		&api.LocalAccountServiceListLocalAccountsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	// Adds existing pre-populated local accounts
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems, len(resList.JSON200.LocalAccounts))
	assert.Equal(t, false, resList.JSON200.HasNext)
}
