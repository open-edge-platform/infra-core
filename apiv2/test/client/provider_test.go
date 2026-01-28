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

func TestProvider_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin CreateGetDelete Provider tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	provider1 := CreateProvider(ctx, t, apiClient, utils.Provider1Request)
	provider2 := CreateProvider(ctx, t, apiClient, utils.Provider2Request)
	provider3 := CreateProvider(ctx, t, apiClient, utils.Provider3Request)

	get1, err := apiClient.ProviderServiceGetProviderWithResponse(
		ctx,
		projectName,
		*provider1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.ProviderName1, get1.JSON200.Name)
	assert.Equal(t, *utils.Provider1Request.Config, *get1.JSON200.Config)

	get2, err := apiClient.ProviderServiceGetProviderWithResponse(
		ctx,
		projectName,
		*provider2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, utils.ProviderName2, get2.JSON200.Name)

	get3, err := apiClient.ProviderServiceGetProviderWithResponse(
		ctx,
		projectName,
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

	projectName := getProjectID(t)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_NoKind_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderNoKind,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Post_NoName_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderNoName,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Post_NoApiEndpoint_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderNoAPIEndpoint,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Post_BadApiCredentials_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceCreateProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderBadCredentials,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Get_UnexistID_NotFound", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceGetProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, provider.StatusCode())
	})

	t.Run("Delete_UnexistID_NotFound", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceDeleteProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, provider.StatusCode())
	})

	t.Run("Get_WrongID_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceGetProviderWithResponse(
			ctx,
			projectName,
			utils.ProviderWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, provider.StatusCode())
	})

	t.Run("Delete_WrongID_BadRequest", func(t *testing.T) {
		provider, err := apiClient.ProviderServiceDeleteProviderWithResponse(
			ctx,
			projectName,
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

	projectName := getProjectID(t)

	resList, err := apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		projectName,
		&api.ProviderServiceListProvidersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	existingProviders := len(resList.JSON200.Providers)

	totalItems := 10
	var offset int
	pageSize := 4

	name := "provider"
	for id := 0; id < totalItems; id++ {
		// Generate sequentialnames
		nameID := fmt.Sprintf("%s%d", name, id)
		utils.Provider1Request.Name = nameID
		CreateProvider(ctx, t, apiClient, utils.Provider1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err = apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		projectName,
		&api.ProviderServiceListProvidersParams{
			Offset:   &offset,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.Providers), pageSize)
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		projectName,
		&api.ProviderServiceListProvidersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	// Adds existing providers
	totalItemsExistent := totalItems + existingProviders
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

	projectName := getProjectID(t)

	resList, err := apiClient.ProviderServiceListProvidersWithResponse(
		ctx,
		projectName,
		&api.ProviderServiceListProvidersParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	// Checks existing pre-populated provider
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, len(resList.JSON200.Providers), 1)
}
