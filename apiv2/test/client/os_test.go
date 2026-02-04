// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

const (
	NumPreloadedOSResources = 4
)

func shortOSSuffix() string {
	trimmed := strings.ReplaceAll(uuid.New().String(), "-", "")
	if len(trimmed) > 8 {
		return trimmed[:8]
	}
	return trimmed
}

func TestOS_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin os tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	os1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)
	os2 := CreateOS(ctx, t, apiClient, utils.OSResource2Request)

	get1, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.OSName1, *get1.JSON200.Name)

	get2, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os2.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get2.StatusCode())
	assert.Equal(t, utils.OSName2, *get2.JSON200.Name)
	assert.Equal(t, utils.OSSecurityFeature2, *get2.JSON200.SecurityFeature)

	log.Info().Msgf("End OSResource tests")
}

func TestOS_UpdatePut(t *testing.T) {
	log.Info().Msgf("Begin OSResource Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// This OS request contains OS Profile Name
	os1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	OSResource1Get, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1Get.StatusCode())
	assert.Equal(t, utils.OSName1, *OSResource1Get.JSON200.Name)

	// Create update request based on existing resource
	// Only update mutable fields: architecture
	// All other fields (name, sha256, profile_name, security_feature, os_type, etc.) are immutable
	arch := "x86"
	updateRequest := api.OperatingSystemResource{
		Architecture: &arch,
		// Sha256 is required field
		Sha256: OSResource1Get.JSON200.Sha256,
	}
	os1Update, err := apiClient.OperatingSystemServiceUpdateOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.ResourceId,
		updateRequest,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, os1Update.StatusCode())

	OSResource1GetUp, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1GetUp.StatusCode())
	// Verify mutable fields were updated
	assert.Equal(t, *updateRequest.Architecture, *OSResource1GetUp.JSON200.Architecture)
	assert.Equal(t, OSResource1Get.JSON200.Sha256, OSResource1GetUp.JSON200.Sha256)
	assert.Empty(t, *OSResource1GetUp.JSON200.Name)
	// Verify other immutable fields remain unchanged
	assert.Equal(t, *OSResource1Get.JSON200.SecurityFeature, *OSResource1GetUp.JSON200.SecurityFeature)
	assert.Equal(t, *OSResource1Get.JSON200.OsType, *OSResource1GetUp.JSON200.OsType)

	log.Info().Msgf("End OSResource Update tests")
}

func TestOS_Errors(t *testing.T) {
	log.Info().Msgf("Begin OSResource Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_InvalidSha_Status_BadRequest", func(t *testing.T) {
		os1Up, err := apiClient.OperatingSystemServiceCreateOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceRequestInvalidSha256,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, os1Up.StatusCode())
	})

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		os1Up, err := apiClient.OperatingSystemServiceUpdateOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceUnexistID,
			utils.OSResource1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, os1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.OperatingSystemServiceDeleteOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceUnexistID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_NotFoundError", func(t *testing.T) {
		os1Up, err := apiClient.OperatingSystemServiceUpdateOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceWrongID,
			utils.OSResource1Request,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, os1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusNotFound", func(t *testing.T) {
		s1res, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusNotFound", func(t *testing.T) {
		resDelSite, err := apiClient.OperatingSystemServiceDeleteOperatingSystemWithResponse(
			ctx,
			projectName,
			utils.OSResourceWrongID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})
	log.Info().Msgf("End OSResource Error tests")
}

func TestOS_List(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// Checks if list resources return expected number of entries
	resList, err := apiClient.OperatingSystemServiceListOperatingSystemsWithResponse(
		ctx,
		projectName,
		&api.OperatingSystemServiceListOperatingSystemsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())

	ExistingOSs := len(resList.JSON200.OperatingSystemResources)

	totalItems := 10
	pageID := 1
	pageSize := 4

	CreateOS(ctx, t, apiClient, utils.OSResource1Request)
	CreateOS(ctx, t, apiClient, utils.OSResource2Request)

	// Checks if list resources return expected number of entries
	resList, err = apiClient.OperatingSystemServiceListOperatingSystemsWithResponse(
		ctx,
		projectName,
		&api.OperatingSystemServiceListOperatingSystemsParams{
			Offset:   &pageID,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(resList.JSON200.OperatingSystemResources), pageSize)
	assert.Equal(t, true, resList.JSON200.HasNext)

	resList, err = apiClient.OperatingSystemServiceListOperatingSystemsWithResponse(
		ctx,
		projectName,
		&api.OperatingSystemServiceListOperatingSystemsParams{},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.GreaterOrEqual(t, totalItems+ExistingOSs, len(resList.JSON200.OperatingSystemResources))
	assert.Equal(t, false, resList.JSON200.HasNext)
}

func TestOS_CreatewithInstallPackage(t *testing.T) {
	log.Info().Msgf("Begin OSResource create with install packages")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	os := CreateOS(ctx, t, apiClient, utils.OSResource1ReqwithInstallPackages)

	get, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get.StatusCode())
	assert.Equal(t, utils.OSName1, *get.JSON200.Name)
	log.Info().Msgf("End OSResource create test")
}

func TestOS_GetWithInstalledPackages(t *testing.T) {
	log.Info().Msgf("Begin OSResource get with installed packages test")

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	osList, err := apiClient.OperatingSystemServiceListOperatingSystemsWithResponse(
		ctx,
		projectName,
		&api.OperatingSystemServiceListOperatingSystemsParams{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, osList.StatusCode())
	assert.GreaterOrEqual(t, len(osList.JSON200.OperatingSystemResources), NumPreloadedOSResources)

	for _, osRes := range osList.JSON200.OperatingSystemResources {
		if *osRes.OsType == api.OSTYPEMUTABLE {
			// Skip if OS is MUTABLE, as it does not have InstalledPackages as the pkg manifest
			continue
		}
		// InstalledPackages shall be JSON-encoded string for IMMUTABLE OS
		// InstalledPackages is empty string for MUTABLE OS
		assert.NotEqual(t, "", *osRes.InstalledPackages)
		var osPackages struct {
			Repo []struct {
				Name    *string `json:"name"`
				Version *string `json:"version"`
			} `json:"repo"`
		}
		// validate that the obtained InstalledPackages is truly unmarshal-able JSON string
		err := json.Unmarshal([]byte(*osRes.InstalledPackages), &osPackages)
		require.NoError(t, err)
		assert.NotEmpty(t, osPackages.Repo)
		assert.NotNil(t, osPackages.Repo[0].Name)
		assert.NotNil(t, osPackages.Repo[0].Version)
	}
	log.Info().Msgf("End OSResource get with installed packages test")
}

func TestOS_CreatewithCustom(t *testing.T) {
	log.Info().Msgf("Begin OSResource create with custom fields")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	suffix := shortOSSuffix()
	OSName1 := "Ubuntu 22.04 LTS generic EXT (24.08.0-n20240816) " + suffix
	OSProfileName1 := "ubuntu-22.04-lts-generic-ext:1.0.2-" + suffix
	OSArch1 := "x86_64"
	OSRepo1 := "http://test.com/test-" + suffix + ".raw.gz"
	metadata := fmt.Sprintf(`{"createdby":"int-test","testrun":%q}`, suffix)

	OSSecFeat := api.SECURITYFEATURENONE
	randSHA := inv_testing.GenerateRandomSha256()
	OSResource1ReqwithCustom := api.OperatingSystemResource{
		Name:            &OSName1,
		ProfileName:     &OSProfileName1,
		Architecture:    &OSArch1,
		RepoUrl:         &OSRepo1,
		Sha256:          randSHA,
		SecurityFeature: &OSSecFeat,
		OsType:          &utils.OsTypeMutable,
		OsProvider:      &utils.OSProvider,
		Metadata:        &metadata,
	}

	// Create OS without installedPackages (it's a read-only field)
	os := CreateOS(ctx, t, apiClient, OSResource1ReqwithCustom)

	get, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os.JSON200.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get.StatusCode())
	assert.Equal(t, OSName1, *get.JSON200.Name)
	log.Info().Msgf("End OSResource create test")
}

func TestOS_UpdatePatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	projectName := getProjectID(t)

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	os1 := CreateOS(ctx, t, apiClient, utils.OSResource1Request)

	OSResource1Get, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1Get.StatusCode())
	assert.Equal(t, utils.OSName1, *OSResource1Get.JSON200.Name)

	// PATCH only mutable fields (Architecture)
	// Cannot patch immutable fields like Name, SecurityFeature, OsType, etc.
	newArch := "arm64"
	patchRequest := api.OperatingSystemResource{
		Architecture: &newArch,
		Sha256:       OSResource1Get.JSON200.Sha256,
	}

	os1Update, err := apiClient.OperatingSystemServicePatchOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.OsResourceID,
		&api.OperatingSystemServicePatchOperatingSystemParams{},
		patchRequest,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, os1Update.StatusCode())
	assert.Equal(t, utils.OSName1, *os1Update.JSON200.Name)

	OSResource1GetUp, err := apiClient.OperatingSystemServiceGetOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1GetUp.StatusCode())
	assert.Equal(t, *utils.OSResource1Request.Name, *OSResource1GetUp.JSON200.Name)
	assert.Equal(
		t,
		patchRequest.Architecture,
		OSResource1GetUp.JSON200.Architecture,
	)
	// Security Feature is immutable
	assert.Equal(t, *utils.OSResource1Request.SecurityFeature, *OSResource1GetUp.JSON200.SecurityFeature)

	osTypeImmutable := api.OSTYPEIMMUTABLE
	osProviderInfra := api.OSPROVIDERKINDINFRA
	immutableUpdate, err := apiClient.OperatingSystemServicePatchOperatingSystemWithResponse(
		ctx,
		projectName,
		*os1.JSON200.OsResourceID,
		&api.OperatingSystemServicePatchOperatingSystemParams{},
		api.OperatingSystemResource{
			OsType:     &osTypeImmutable,
			OsProvider: &osProviderInfra,
		},
		AddJWTtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, immutableUpdate.StatusCode())
}
