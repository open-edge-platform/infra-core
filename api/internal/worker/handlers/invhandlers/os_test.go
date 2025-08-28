// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/internal/worker/handlers"
	inv_handlers "github.com/open-edge-platform/infra-core/api/internal/worker/handlers/invhandlers"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/test/utils"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	os_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var (
	osResName         = "osName"
	ImageURL          = "repo"
	ImageID           = "some OS-specific version"
	UpdateSources     = []string{"sources"}
	KernelCommand     = "cmd"
	Arch              = "x86_64"
	Sha256            = "0425b2a513f6b391850f9c308cf9716b6fb13e43eb0e891b63e63ccc47c85ec8"
	ProfileName       = "OS profile name test"
	ProfileVersion    = "1.0.0"
	InstalledPackages = "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero"
	SecurityFeature   = api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION
	OsType            = api.OPERATINGSYSTEMTYPEIMMUTABLE
	OsProvider        = api.OPERATINGSYSTEMPROVIDERINFRA
)

//nolint:cyclop // it's a test helper
func BuildFmFromOSRequest(body api.OperatingSystemResource) []string {
	fm := []string{}
	fm = append(fm, "name")
	if body.RepoUrl != nil {
		fm = append(fm, "image_url")
	}
	if body.ImageUrl != nil {
		fm = append(fm, "image_url")
	}
	if body.ImageId != nil {
		fm = append(fm, "image_id")
	}
	if body.Architecture != nil {
		fm = append(fm, "architecture")
	}
	if body.KernelCommand != nil {
		fm = append(fm, "kernel_command")
	}
	if body.UpdateSources != nil {
		fm = append(fm, "update_sources")
	}
	if body.Sha256 != "" {
		fm = append(fm, "sha256")
	}
	if body.ProfileName != nil {
		fm = append(fm, "profile_name")
	}
	if body.ProfileVersion != nil {
		fm = append(fm, "profile_version")
	}
	if body.InstalledPackages != nil {
		fm = append(fm, "installed_packages")
	}
	if body.SecurityFeature != nil {
		fm = append(fm, "security_feature")
	}
	if body.OsType != nil {
		fm = append(fm, "os_type")
	}
	if body.OsProvider != nil {
		fm = append(fm, "os_provider")
	}

	return fm
}

func Test_OSHandler_Job_Error(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	job := types.NewJob(ctx, BadOperation, types.OSResource, nil, inv_handlers.OSResourceURLParams{})
	_, err := h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotImplemented, errors.ErrorToHTTPStatus(err))
}

// check that we pass the expected filters to the inventory.
func Test_OSHandler_List(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	// test List
	job := types.NewJob(
		ctx, types.List, types.OSResource,
		api.GetOSResourcesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		}, nil,
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
	listResources, ok := r.Payload.Data.(api.OperatingSystemResourceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 0, len(*listResources.OperatingSystemResources))

	os1 := inv_testing.CreateOs(t)

	job = types.NewJob(
		ctx, types.List, types.OSResource,
		api.GetOSResourcesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		}, nil,
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
	listResources, ok = r.Payload.Data.(api.OperatingSystemResourceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.OperatingSystemResources))

	filter := fmt.Sprintf("%s = %q", os_v1.OperatingSystemResourceFieldResourceId, os1.GetResourceId())
	orderBy := os_v1.OperatingSystemResourceFieldResourceId
	job = types.NewJob(
		ctx, types.List, types.OSResource,
		api.GetOSResourcesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
			Filter:   &filter,
			OrderBy:  &orderBy,
		}, nil,
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
	listResources, ok = r.Payload.Data.(api.OperatingSystemResourceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.OperatingSystemResources))

	// test List error - wrong params
	job = types.NewJob(
		context.TODO(), types.List, types.OSResource,
		api.GetComputeHostsParams{}, nil,
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_OSHandler_Post(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	ctx := context.TODO()

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)
	body := api.OperatingSystemResource{
		Name:              &osResName,
		RepoUrl:           &ImageURL,
		ImageId:           &ImageID,
		UpdateSources:     UpdateSources,
		KernelCommand:     &KernelCommand,
		Architecture:      &Arch,
		Sha256:            Sha256,
		ProfileName:       &ProfileName,
		ProfileVersion:    &ProfileVersion,
		InstalledPackages: &InstalledPackages,
		SecurityFeature:   &SecurityFeature,
		OsType:            &OsType,
		// no OsProvider, required
	}

	// create without required field
	job := types.NewJob(
		ctx, types.Post, types.OSResource,
		&body, inv_handlers.OSResourceURLParams{},
	)
	_, err := h.Do(job)
	require.Error(t, err)

	body.OsProvider = &OsProvider
	job = types.NewJob(
		ctx, types.Post, types.OSResource,
		&body, inv_handlers.OSResourceURLParams{},
	)
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok := r.Payload.Data.(*api.OperatingSystemResource)
	require.True(t, ok)

	// Validate Post changes
	job = types.NewJob(ctx, types.Get, types.OSResource, nil, inv_handlers.OSResourceURLParams{
		OSResourceID: *gotRes.OsResourceID,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.OperatingSystemResource)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, osResName, *gotRes.Name)
	assert.Equal(t, api.OPERATINGSYSTEMTYPEIMMUTABLE, *gotRes.OsType)
	assert.Equal(t, api.OPERATINGSYSTEMPROVIDERINFRA, *gotRes.OsProvider)

	// Post error - wrong body request format
	job = types.NewJob(
		ctx, types.Post, types.OSResource,
		&api.Host{}, inv_handlers.OSResourceURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

//nolint:funlen // it is an unit test
func Test_OSHandler_Put(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	ctx := context.TODO()
	osResource := inv_testing.CreateOs(t)

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	osKernel := osResource.GetKernelCommand()
	osArch := osResource.GetArchitecture()
	osProfileName := osResource.GetProfileName()
	osProfileVersion := osResource.GetProfileVersion()
	osInstalledPackages := osResource.GetInstalledPackages()
	osSecurityFeature := api.SECURITYFEATUREUNSPECIFIED
	osSHA := osResource.GetSha256()
	osImageURL := osResource.GetImageUrl()
	osImageID := osResource.GetImageId()
	osUpdateSource := osResource.GetUpdateSources()
	osType := api.OPERATINGSYSTEMTYPEUNSPECIFIED

	bodyUpdate := api.OperatingSystemResource{
		UpdateSources:     osUpdateSource,
		KernelCommand:     &osKernel,
		Architecture:      &osArch,
		InstalledPackages: &osInstalledPackages,
	}

	job := types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Put changes
	job = types.NewJob(ctx, types.Get, types.OSResource, nil, inv_handlers.OSResourceURLParams{
		OSResourceID: osResource.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok := r.Payload.Data.(*api.OperatingSystemResource)
	require.True(t, ok)
	assert.NotNil(t, gotRes)

	// Update error - wrong body format
	job = types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&api.Host{},
		inv_handlers.OSResourceURLParams{OSResourceID: "os-12345678"},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// Update error - wrong params
	job = types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// immutable fields are not changing
	bodyUpdate = api.OperatingSystemResource{
		ProfileName:    &osProfileName,
		ProfileVersion: &osProfileVersion,
		ImageId:        &osImageID,
		RepoUrl:        &osImageURL,
		ImageUrl:       &osImageURL,
	}

	job = types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	osProviderLenovo := api.OPERATINGSYSTEMPROVIDERLENOVO
	bodyUpdate = api.OperatingSystemResource{
		SecurityFeature: &osSecurityFeature,
		OsType:          &osType,
		OsProvider:      &osProviderLenovo,
	}

	job = types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	bodyUpdate = api.OperatingSystemResource{
		Sha256: osSHA,
	}

	job = types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// validate all immutable fields unchanged
	job = types.NewJob(ctx, types.Get, types.OSResource, nil, inv_handlers.OSResourceURLParams{
		OSResourceID: *gotRes.OsResourceID,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.OperatingSystemResource)
	require.True(t, ok)
	// CreateOs creates OS with Infra provider by default, immutable field shouldn't be updated
	assert.Equal(t, api.OPERATINGSYSTEMPROVIDERINFRA, *gotRes.OsProvider)
}

func Test_OSHandler_Patch(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}

	ctx := context.TODO()
	osResource := inv_testing.CreateOs(t)

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	osImageURL := osResource.GetImageUrl()
	osUpdateSource := osResource.GetUpdateSources()

	bodyUpdate := api.OperatingSystemResource{
		RepoUrl:       &osImageURL,
		UpdateSources: osUpdateSource,
	}

	job := types.NewJob(
		ctx,
		types.Patch,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Patch changes
	job = types.NewJob(ctx, types.Get, types.OSResource, nil, inv_handlers.OSResourceURLParams{
		OSResourceID: osResource.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok := r.Payload.Data.(*api.OperatingSystemResource)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, osResource.GetName(), *gotRes.Name)

	// Patch ok - immutable field are discarded
	bodyUpdate = api.OperatingSystemResource{
		RepoUrl: &ImageURL,
		Sha256:  Sha256, // This field is changing
	}

	job = types.NewJob(
		ctx,
		types.Patch,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	_, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	bodyUpdate = api.OperatingSystemResource{
		RepoUrl:         &ImageURL,
		SecurityFeature: &SecurityFeature, // This field is changing
	}

	job = types.NewJob(
		ctx,
		types.Patch,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	_, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	bodyUpdate = api.OperatingSystemResource{
		RepoUrl:     &ImageURL,
		ProfileName: &ProfileName, // This field is changing
	}

	job = types.NewJob(
		ctx,
		types.Patch,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	_, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
}

func Test_OSHandler_PatchFieldMask(t *testing.T) {
	mockClient := utils.NewMockInventoryServiceClient(
		utils.MockResponses{
			ListResourcesResponse: &inventory.ListResourcesResponse{
				Resources: []*inventory.GetResourceResponse{},
			},
			GetResourceResponse:    &inventory.GetResourceResponse{},
			UpdateResourceResponse: &inventory.Resource{},
		},
	)
	client := &clients.InventoryClientHandler{
		InvClient: mockClient,
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	bodyUpdate := api.OperatingSystemResource{
		Name:              &osResName,
		UpdateSources:     UpdateSources,
		KernelCommand:     &KernelCommand,
		Architecture:      &Arch,
		InstalledPackages: &InstalledPackages,
	}

	job := types.NewJob(
		context.TODO(),
		types.Patch,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: "os-1234"},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// test Patch FieldMask
	expectedPatchFieldMask := BuildFmFromOSRequest(bodyUpdate)
	ou := &os_v1.OperatingSystemResource{}
	expectedFieldMask, err := fieldmaskpb.New(ou, expectedPatchFieldMask...)
	require.NoError(t, err)

	if mockClient.LastUpdateResourceRequestFieldMask != nil {
		mockClient.LastUpdateResourceRequestFieldMask.Normalize()
		expectedFieldMask.Normalize()
		if !proto.Equal(expectedFieldMask, mockClient.LastUpdateResourceRequestFieldMask) {
			err = fmt.Errorf(
				"FieldMask is incorrectly constructed, expected: %s got: %s",
				expectedFieldMask.Paths,
				mockClient.LastUpdateResourceRequestFieldMask.Paths,
			)
		}
	} else {
		err = fmt.Errorf("no request in Mock Inventory")
	}
	assert.NoError(t, err)
}

func Test_OSHandler_Get(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}

	ctx := context.TODO()
	osResource := inv_testing.CreateOs(t)

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	job := types.NewJob(
		ctx,
		types.Get,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Get error - wrong params
	job = types.NewJob(
		ctx,
		types.Get,
		types.OSResource,
		nil,
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_OSHandler_Delete(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	ctx := context.TODO()
	osResource := inv_testing.CreateOsNoCleanup(t)

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	job := types.NewJob(
		ctx,
		types.Delete,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, r.Status)

	// Validate Delete
	job = types.NewJob(
		ctx,
		types.Get,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: osResource.ResourceId},
	)
	_, err = h.Do(job)
	require.Error(t, err)

	// Delete error - wrong params
	job = types.NewJob(
		ctx,
		types.Delete,
		types.OSResource,
		nil,
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_Inventory_OS_Integration(t *testing.T) {
	// verify the projection of the constants to Proto first;
	// we build a map using the field names of the proto stored in the
	// ProtoOu* slices in internal/work/handlers/os.go. Elements must
	// have a mapping key otherwise we throw an error if there is no
	// alignment with OU proto in Inventory. Make sure to update these
	// two slices in internal/work/handlers/os.go
	osResource := &os_v1.OperatingSystemResource{}
	validateInventoryIntegration(
		t,
		osResource,
		api.OperatingSystemResource{},
		inv_handlers.OpenAPIOSResourceToProto,
		inv_handlers.OpenAPIOSToProtoExcluded,
		maps.Values(inv_handlers.OpenAPIOSResourceToProto),
		true,
	)
}

// Test_OSHandler_InvMockClient_Errors evaluates all
// OS handler methods with mock inventory client
// that returns errors.
func Test_OSHandler_InvMockClient_Errors(t *testing.T) {
	mockClient := utils.NewMockInventoryServiceClientError()
	client := &clients.InventoryClientHandler{
		InvClient: mockClient,
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	job := types.NewJob(
		context.TODO(), types.List, types.OSResource,
		api.GetOSResourcesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		}, nil,
	)
	_, err := h.Do(job)
	assert.Error(t, err)

	body := api.OperatingSystemResource{
		RepoUrl:       &ImageURL,
		UpdateSources: UpdateSources,
		KernelCommand: &KernelCommand,
		Architecture:  &Arch,
	}
	job = types.NewJob(
		context.TODO(), types.Post, types.OSResource,
		&body, inv_handlers.OSResourceURLParams{},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		context.TODO(),
		types.Put,
		types.OSResource,
		&body,
		inv_handlers.OSResourceURLParams{OSResourceID: "os-12345678"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		context.TODO(),
		types.Get,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: "os-12345678"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		context.TODO(),
		types.Delete,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: "os-1234"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)
}
