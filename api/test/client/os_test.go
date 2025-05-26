// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/test/utils"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

const (
	NumPreloadedOSResources = 4
)

func TestOS_CreateGetDelete(t *testing.T) {
	log.Info().Msgf("Begin os tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	os1 := CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	os2 := CreateOS(t, ctx, apiClient, utils.OSResource2Request)

	get1, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get1.StatusCode())
	assert.Equal(t, utils.OSName1, *get1.JSON200.Name)

	get2, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os2.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
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

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	// This OS request contains OS Profile Name
	os1 := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	OSResource1Get, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1Get.StatusCode())
	assert.Equal(t, utils.OSName1, *OSResource1Get.JSON200.Name)

	// this OS request does not contain Profile Name, but we need to set SHA256 checksum
	// and Profile Name to be equal to what we had in the first request
	utils.OSResource2Request.Sha256 = utils.OSResource1Request.Sha256
	utils.OSResource2Request.ProfileName = utils.OSResource1Request.ProfileName
	utils.OSResource2Request.SecurityFeature = utils.OSResource1Request.SecurityFeature
	utils.OSResource2Request.OsType = nil
	utils.OSResource2Request.OsProvider = nil
	os1Update, err := apiClient.PutOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		utils.OSResource2Request,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, os1Update.StatusCode())

	OSResource1GetUp, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1GetUp.StatusCode())
	assert.Equal(t, *utils.OSResource2Request.Name, *OSResource1GetUp.JSON200.Name)
	assert.Equal(t, *utils.OSResource2Request.Architecture, *OSResource1GetUp.JSON200.Architecture)
	assert.Equal(t, "", *OSResource1GetUp.JSON200.KernelCommand)
	// Security Feature is immutable
	assert.Equal(t, *utils.OSResource1Request.SecurityFeature, *OSResource1GetUp.JSON200.SecurityFeature)

	log.Info().Msgf("End OSResource Update tests")
}

func TestOS_UpdatePatch(t *testing.T) {
	log.Info().Msgf("Begin OSResource Update tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	os1 := CreateOS(t, ctx, apiClient, utils.OSResource1Request)

	OSResource1Get, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1Get.StatusCode())
	assert.Equal(t, utils.OSName1, *OSResource1Get.JSON200.Name)

	os1Update, err := apiClient.PatchOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		utils.OSResource2Request,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, os1Update.StatusCode())
	assert.Equal(t, utils.OSName2, *os1Update.JSON200.Name)

	OSResource1GetUp, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, OSResource1GetUp.StatusCode())
	assert.Equal(t, *utils.OSResource1Request.KernelCommand, *OSResource1GetUp.JSON200.KernelCommand)
	assert.Equal(t, *utils.OSResource2Request.Name, *OSResource1GetUp.JSON200.Name)
	assert.Equal(
		t,
		utils.OSResource2Request.Architecture,
		OSResource1GetUp.JSON200.Architecture,
	)
	// Security Feature is immutable
	assert.Equal(t, *utils.OSResource1Request.SecurityFeature, *OSResource1GetUp.JSON200.SecurityFeature)

	osTypeImmutable := api.OPERATINGSYSTEMTYPEIMMUTABLE
	osProviderInfra := api.OPERATINGSYSTEMPROVIDERINFRA
	immutableUpdate, err := apiClient.PatchOSResourcesOSResourceIDWithResponse(
		ctx,
		*os1.JSON201.OsResourceID,
		api.OperatingSystemResource{
			OsType:     &osTypeImmutable,
			OsProvider: &osProviderInfra,
		},
		AddJWTtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, immutableUpdate.StatusCode())

	log.Info().Msgf("End OSResource Update tests")
}

func TestOS_Errors(t *testing.T) {
	log.Info().Msgf("Begin OSResource Error tests")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)
	if err != nil {
		t.Fatalf("new API client error %s", err.Error())
	}

	t.Run("Post_NoUpdateSources_Status_BadRequest", func(t *testing.T) {
		os1Up, err := apiClient.PostOSResourcesWithResponse(
			ctx,
			utils.OSResourceRequestNoUpdateSources,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		log.Info().Msgf("Error UpSources %s", os1Up.Body)
		assert.Equal(t, http.StatusBadRequest, os1Up.StatusCode())
	})

	t.Run("Post_NoSha_Status_BadRequest", func(t *testing.T) {
		os1Up, err := apiClient.PostOSResourcesWithResponse(
			ctx,
			utils.OSResourceRequestNoSha,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		log.Info().Msgf("Error RepoURL %s", os1Up.Body)
		assert.Equal(t, http.StatusBadRequest, os1Up.StatusCode())
	})

	t.Run("Post_InvalidSha_Status_BadRequest", func(t *testing.T) {
		os1Up, err := apiClient.PostOSResourcesWithResponse(
			ctx,
			utils.OSResourceRequestInvalidSha256,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		log.Info().Msgf("Error RepoURL %s", os1Up.Body)
		assert.Equal(t, http.StatusBadRequest, os1Up.StatusCode())
	})

	t.Run("Put_UnexistID_Status_NotFoundError", func(t *testing.T) {
		os1Up, err := apiClient.PutOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceUnexistID,
			utils.OSResource2Request,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		log.Info().Msgf("Error OSResourceUnexistID %s", os1Up.Body)
		assert.Equal(t, http.StatusNotFound, os1Up.StatusCode())
	})

	t.Run("Patch_UnexistID_Status_NotFoundError", func(t *testing.T) {
		os1Up, err := apiClient.PatchOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceUnexistID,
			utils.OSResource2Request,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, os1Up.StatusCode())
	})

	t.Run("Get_UnexistID_Status_NotFoundError", func(t *testing.T) {
		s1res, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceUnexistID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, s1res.StatusCode())
	})

	t.Run("Delete_UnexistID_Status_NotFoundError", func(t *testing.T) {
		resDelSite, err := apiClient.DeleteOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceUnexistID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resDelSite.StatusCode())
	})

	t.Run("Put_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		os1Up, err := apiClient.PutOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceWrongID,
			utils.OSResource1Request,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, os1Up.StatusCode())
	})

	t.Run("Patch_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		os1Up, err := apiClient.PatchOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceWrongID,
			utils.OSResource1Request,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, os1Up.StatusCode())
	})

	t.Run("Get_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		s1res, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceWrongID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, s1res.StatusCode())
	})

	t.Run("Delete_WrongID_Status_StatusBadRequest", func(t *testing.T) {
		resDelSite, err := apiClient.DeleteOSResourcesOSResourceIDWithResponse(
			ctx,
			utils.OSResourceWrongID,
			AddJWTtoTheHeader,
			AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resDelSite.StatusCode())
	})
	log.Info().Msgf("End OSResource Error tests")
}

func TestOSList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	totalItems := 10
	pageId := 1
	pageSize := 4

	for id := 0; id < totalItems; id++ {
		// Re-generate the random sha for each new OS resource being created
		utils.OSResource1Request.Sha256 = inv_testing.GenerateRandomSha256()
		profileName := inv_testing.GenerateRandomProfileName()
		utils.OSResource1Request.ProfileName = &profileName
		CreateOS(t, ctx, apiClient, utils.OSResource1Request)
	}

	// Checks if list resources return expected number of entries
	resList, err := apiClient.GetOSResourcesWithResponse(
		ctx,
		&api.GetOSResourcesParams{
			Offset:   &pageId,
			PageSize: &pageSize,
		},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, len(*resList.JSON200.OperatingSystemResources), pageSize)
	assert.Equal(t, true, *resList.JSON200.HasNext)

	resList, err = apiClient.GetOSResourcesWithResponse(
		ctx,
		&api.GetOSResourcesParams{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, totalItems+NumPreloadedOSResources, len(*resList.JSON200.OperatingSystemResources))
	assert.Equal(t, false, *resList.JSON200.HasNext)
}

func TestOSList_ListEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	resList, err := apiClient.GetOSResourcesWithResponse(
		ctx,
		&api.GetOSResourcesParams{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resList.StatusCode())
	assert.Equal(t, NumPreloadedOSResources, len(*resList.JSON200.OperatingSystemResources))
}

func TestOS_CreatewithInstallPackage(t *testing.T) {
	log.Info().Msgf("Begin OSResource create with install packages")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	os := CreateOS(t, ctx, apiClient, utils.OSResource1ReqwithInstallPackages)

	get, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
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

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	osList, err := apiClient.GetOSResourcesWithResponse(
		ctx,
		&api.GetOSResourcesParams{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, osList.StatusCode())
	assert.Equal(t, NumPreloadedOSResources, len(*osList.JSON200.OperatingSystemResources))

	for _, osRes := range *osList.JSON200.OperatingSystemResources {
		// InstalledPackages shall be JSON-encoded string for IMMUTABLE OS
		// InstalledPackages is empty string for MUTABLE OS
		if *osRes.OsType == api.OPERATINGSYSTEMTYPEIMMUTABLE {
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
		} else if *osRes.OsType == api.OPERATINGSYSTEMTYPEMUTABLE {
			assert.Equal(t, "", *osRes.InstalledPackages)
		}
	}
	log.Info().Msgf("End OSResource get with installed packages test")
}

func TestOS_CreatewithCustom(t *testing.T) {
	log.Info().Msgf("Begin OSResource create with install custom fields")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	OSName1 := "Ubuntu 22.04 LTS generic EXT (24.08.0-n20240816)"
	OSProfileName1 := "ubuntu-22.04-lts-generic-ext:1.0.2 TestName#724"
	OSKernel1 := "kvmgt vfio-iommu-type1 vfio-mdev i915.enable_gvt=1 kvm.ignore_msrs=1 intel_iommu=on iommu=pt drm.debug=0"
	OSArch1 := "x86"
	OSRepo1 := "http://test.com/test.raw.gz"
	OSInstalledPackages := "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero"
	OSSecFeat := api.SECURITYFEATUREUNSPECIFIED
	OSUpdateSources := `#ReleaseService\nTypes: deb\nURIs:
https://files-rs.internal.intel.com/repository\nSuites:
24.08\nComponents: release\nSigned-By:\n -----BEGIN PGP PUBLIC KEY BLOCK-----\n .\n mQINBGXE3tkBEAD85hzXnrq6rPnOXxwns35NfLaT595jJ3r5J17U/heOymT+K18D\n A6ewAwQgyHEWemW87xW6iqzRI4jB5m/ #### FAKE ### tboh57AZ40JFRlzz4\n dKybtByZ2ntW/sYvXwR818/sUd2PjtRHekBq+bprw2JR2OwPhfAswBs9UzWNiSqd\n rA3NksCeuj/j6sSaqpXn123ZtlliZttviM+bvbSps5qJ5TbxHtSwr4H5gYSlHVT/\n IwqUfFrYNoQVDejlGkVgyjQYonEqk8eX\n =w4R+\n -----END PGP PUBLIC KEY BLOCK-----`
	OSTypeMutable := api.OPERATINGSYSTEMTYPEMUTABLE
	OSProviderLenovo := api.OPERATINGSYSTEMPROVIDERLENOVO

	OSResource1ReqwithCustom := api.OperatingSystemResource{
		Name:              &OSName1,
		ProfileName:       &OSProfileName1,
		KernelCommand:     &OSKernel1,
		Architecture:      &OSArch1,
		UpdateSources:     []string{OSUpdateSources},
		RepoUrl:           &OSRepo1,
		Sha256:            inv_testing.GenerateRandomSha256(),
		InstalledPackages: &OSInstalledPackages,
		SecurityFeature:   &OSSecFeat,
		OsType:            &OSTypeMutable,
		OsProvider:        &OSProviderLenovo,
	}

	os := CreateOS(t, ctx, apiClient, OSResource1ReqwithCustom)

	get, err := apiClient.GetOSResourcesOSResourceIDWithResponse(
		ctx,
		*os.JSON201.OsResourceID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get.StatusCode())
	assert.Equal(t, OSName1, *get.JSON200.Name)
	log.Info().Msgf("End OSResource create test")
}
