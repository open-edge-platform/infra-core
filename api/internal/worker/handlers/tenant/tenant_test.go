// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/internal/worker/handlers"
	inv_handlers "github.com/open-edge-platform/infra-core/api/internal/worker/handlers/invhandlers"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	os_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
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

var (
	pgSize   = 10
	pgOffset = 0
)

// Runs all CRULD job operations with tenant ID added to ctx
// inv_testing.APIClient is enabled with interceptor to extract ctx
// and add it to messages/calls.
//
//nolint:funlen // it is a test
func Test_OSHandler_CRUD_Tenant_OK(t *testing.T) {
	ctx := tenant.AddTenantIDToContext(context.TODO(), uuid.New().String())
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	// Validate Create
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
		OsProvider:        &OsProvider,
	}
	job := types.NewJob(
		ctx, types.Post, types.OSResource,
		&body, inv_handlers.OSResourceURLParams{},
	)
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	osResource, ok := r.Payload.Data.(*api.OperatingSystemResource)
	require.True(t, ok)

	// Validate Get
	job = types.NewJob(
		ctx,
		types.Get,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: *osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate List
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
	listResources, ok := r.Payload.Data.(api.OperatingSystemResourceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.OperatingSystemResources))

	filter := fmt.Sprintf("%s = %q", os_v1.OperatingSystemResourceFieldResourceId, *osResource.ResourceId)
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

	// Validate Put
	bodyUpdate := api.OperatingSystemResource{
		UpdateSources:     UpdateSources,
		KernelCommand:     &KernelCommand,
		Architecture:      &Arch,
		InstalledPackages: &InstalledPackages,
	}

	job = types.NewJob(
		ctx,
		types.Put,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: *osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Patch
	bodyUpdate = api.OperatingSystemResource{
		RepoUrl:       &ImageURL,
		UpdateSources: UpdateSources,
	}

	job = types.NewJob(
		ctx,
		types.Patch,
		types.OSResource,
		&bodyUpdate,
		inv_handlers.OSResourceURLParams{OSResourceID: *osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Delete
	job = types.NewJob(
		ctx,
		types.Delete,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: *osResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, r.Status)

	// Validate Delete was done
	job = types.NewJob(
		ctx,
		types.Get,
		types.OSResource,
		nil,
		inv_handlers.OSResourceURLParams{OSResourceID: *osResource.ResourceId},
	)
	_, err = h.Do(job)
	require.Error(t, err)
}
