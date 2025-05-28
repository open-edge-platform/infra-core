// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin CreateGetDelete Provider tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	provider1 := CreateProvider(t, ctx, apiClient, utils.Provider1Request)
	provider2 := CreateProvider(t, ctx, apiClient, utils.Provider2Request)
	provider3 := CreateProvider(t, ctx, apiClient, utils.Provider3Request)

	get1, err := apiClient.ProviderServiceGetProviderWithResponse(
		ctx,
		*provider1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.ProviderName1, get1.JSON200.Name)
	assert.Equal(t, *utils.Provider1Request.Config, *get1.JSON200.Config)

	get2, err := apiClient.ProviderServiceGetProviderWithResponse(
		ctx,
		*provider2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, utils.ProviderName2, get2.JSON200.Name)

	get3, err := apiClient.ProviderServiceGetProviderWithResponse(
		ctx,
		*provider3.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get3.StatusCode())
	assert.Equal(t, utils.ProviderName3, get3.JSON200.Name)

	log.Info().Msgf("End CreateGetDelete Provider tests")
}

func TestProvider_Errors(t *testing.T) {
	log.Info().Msgf("Begin Errors Provider tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_NoKind_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			utils.ProviderNoKind,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Post_NoName_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			utils.ProviderNoName,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Post_NoApiEndpoint_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			utils.ProviderNoApiEndpoint,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Post_BadApiCredentials_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			utils.ProviderBadCredentials,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Get_UnexistID_NotFound", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceGetProviderWithResponse(
			ctx,
			utils.ProviderUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, provider.StatusCode())
	})

	t.Run("Delete_UnexistID_NotFound", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceDeleteProviderWithResponse(
			ctx,
			utils.ProviderUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, provider.StatusCode())
	})

	t.Run("Get_WrongID_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceGetProviderWithResponse(
			ctx,
			utils.ProviderWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Delete_WrongID_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceDeleteProviderWithResponse(
			ctx,
			utils.ProviderWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})
	log.Info().Msgf("End Errors Provider tests")
}

func TestProviderList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	var offset int
	var pageSize int = 4

	name := "provider"
	for id := 0; id < totalItems; id++ {
		// Generate sequentialnames
		nameId := fmt.Sprintf("%s%d", name, id)
		utils.Provider1Request.Name = nameId
		CreateProvider(t, ctx, apiClient, utils.Provider1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		&api.ProviderServiceListProvidersParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Providers), int(pageSize))
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		&api.ProviderServiceListProvidersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	// Adds existing pre-populated provider
	totalItemsExistent := totalItems + 1
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItemsExistent, len(resList.JSON200.Providers))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestProviderList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		&api.ProviderServiceListProvidersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	// Checks existing pre-populated provider
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, 1, len(resList.JSON200.Providers))
}
