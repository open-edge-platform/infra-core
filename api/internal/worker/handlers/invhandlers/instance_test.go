// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/internal/worker/handlers"
	inv_handlers "github.com/open-edge-platform/infra-core/api/internal/worker/handlers/invhandlers"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/test/utils"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var (
	instName        = "instName"
	instHostID      = "host-12345678"
	instOSID        = "os-12345678"
	instKind        = api.INSTANCEKINDMETAL
	securityFeature = api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION
	Username        = "test-user"
	SSHKey          = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ test-user@example.com"
)

func verifyInstanceStatusFields(t *testing.T, instance *api.Instance,
	expected *computev1.InstanceResource,
) {
	t.Helper()

	require.NotNil(t, instance.InstanceStatusIndicator)
	assert.Equal(t, expected.GetInstanceStatus(), *instance.InstanceStatus)
	assert.Equal(t, expected.GetInstanceStatusTimestamp(), *instance.InstanceStatusTimestamp)
	assert.Equal(t, *inv_handlers.GrpcToOpenAPIStatusIndicator(expected.GetInstanceStatusIndicator()),
		*instance.InstanceStatusIndicator)
	assert.Equal(t, expected.GetInstanceStatusDetail(), *instance.InstanceStatusDetail)

	require.NotNil(t, instance.ProvisioningStatusIndicator)
	assert.Equal(t, expected.GetProvisioningStatus(), *instance.ProvisioningStatus)
	assert.Equal(t, expected.GetProvisioningStatusTimestamp(), *instance.ProvisioningStatusTimestamp)
	assert.Equal(t, *inv_handlers.GrpcToOpenAPIStatusIndicator(expected.GetProvisioningStatusIndicator()),
		*instance.ProvisioningStatusIndicator)

	require.NotNil(t, instance.UpdateStatusIndicator)
	require.NotNil(t, instance.UpdateStatusDetail)
	assert.Equal(t, expected.GetUpdateStatusDetail(), *instance.UpdateStatusDetail)

	require.NotNil(t, instance.TrustedAttestationStatusIndicator)
	assert.Equal(t, expected.GetTrustedAttestationStatus(), *instance.TrustedAttestationStatus)
	assert.Equal(t, expected.GetTrustedAttestationStatusTimestamp(), *instance.TrustedAttestationStatusTimestamp)
	assert.Equal(t, *inv_handlers.GrpcToOpenAPIStatusIndicator(expected.GetTrustedAttestationStatusIndicator()),
		*instance.TrustedAttestationStatusIndicator)
}

func Test_InstanceHandler_Job_Error(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	job := types.NewJob(ctx, BadOperation, types.Instance, nil, inv_handlers.InstanceURLParams{})
	_, err := h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotImplemented, errors.ErrorToHTTPStatus(err))
}

// check that we pass the expected filters to the inventory.
//
//nolint:funlen // it is a test
func Test_InstanceHandler_List(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	// test List
	job := types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		}, nil,
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
	listResources, ok := r.Payload.Data.(api.InstanceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 0, len(*listResources.Instances))

	hostResource := inv_testing.CreateHost(t, nil, nil)
	osResource := inv_testing.CreateOs(t)
	instance1 := inv_testing.CreateInstance(t, hostResource, osResource)

	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		}, nil,
	)
	r, err = h.Do(job)
	assert.Equal(t, http.StatusOK, r.Status)
	assert.NoError(t, err)
	listResources, ok = r.Payload.Data.(api.InstanceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.Instances))

	// test List with workloadMemberID filter
	workloadMemberID := "workloadmember-12345678"
	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:           &pgOffset,
			PageSize:         &pgSize,
			WorkloadMemberID: &workloadMemberID,
		}, nil,
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	listResources, ok = r.Payload.Data.(api.InstanceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 0, len(*listResources.Instances))

	// test List by hasWorkloadMember
	workloadMemberID = ""
	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:           &pgOffset,
			PageSize:         &pgSize,
			WorkloadMemberID: &workloadMemberID,
		}, nil,
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	listResources, ok = r.Payload.Data.(api.InstanceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 0, len(*listResources.Instances))

	// test List with hostID filter
	hostID := hostResource.ResourceId
	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
			HostID:   &hostID,
		}, nil,
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// test List with siteID filter
	siteID := "site-12345678"
	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
			SiteID:   &siteID,
		}, nil,
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// test List with generic filter
	filter := fmt.Sprintf("%s = %q", computev1.InstanceResourceFieldResourceId, instance1.GetResourceId())
	orderBy := computev1.InstanceResourceFieldResourceId
	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
			Filter:   &filter,
			OrderBy:  &orderBy,
		}, nil,
	)
	r, err = h.Do(job)
	assert.Equal(t, http.StatusOK, r.Status)
	assert.NoError(t, err)
	listResources, ok = r.Payload.Data.(api.InstanceList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.Instances))

	// test List error - wrong params
	job = types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetComputeHostsParams{}, nil,
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_InstanceHandler_Post(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	hostResource := inv_testing.CreateHost(t, nil, nil)
	osResource := inv_testing.CreateOs(t)
	localaccount := inv_testing.CreateLocalAccount(t,
		"test-user",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ test-user@example.com")
	localaccountResID := localaccount.GetResourceId()
	body := api.Instance{
		HostID:          &hostResource.ResourceId,
		OsID:            &osResource.ResourceId,
		Kind:            &instKind,
		Name:            &instName,
		SecurityFeature: &securityFeature,
		LocalAccountID:  &localaccountResID,
	}
	job := types.NewJob(
		ctx, types.Post, types.Instance,
		&body, inv_handlers.InstanceURLParams{},
	)
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok := r.Payload.Data.(*api.Instance)
	require.True(t, ok)

	// Validate Post changes
	job = types.NewJob(ctx, types.Get, types.Instance, nil, inv_handlers.InstanceURLParams{
		InstanceID: *gotRes.InstanceID,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Instance)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.Name, gotRes.Name)

	// Post error - wrong body request format
	job = types.NewJob(
		context.TODO(), types.Post, types.Instance,
		&api.Host{}, inv_handlers.InstanceURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// Make sure to delete Instance, so Host/OS can be deleted.
	inv_testing.HardDeleteInstance(t, *gotRes.InstanceID)
}

func Test_InstanceHandler_Put(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	// Update - error not implemented
	job := types.NewJob(
		context.TODO(),
		types.Put,
		types.Instance,
		&api.Instance{},
		inv_handlers.InstanceURLParams{InstanceID: "inst-12345678"},
	)
	_, err := h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotImplemented, errors.ErrorToHTTPStatus(err))
}

func Test_InstanceHandler_Patch(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	osResource1 := inv_testing.CreateOs(t)
	osResource2 := inv_testing.CreateOs(t)
	localaccount := inv_testing.CreateLocalAccount(t, Username, SSHKey)
	insResource := inv_testing.CreateInstance(t, nil, osResource1)
	ctx := context.TODO()
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	bodyUpdate := api.Instance{
		Name: &instName,
		OsID: &osResource2.ResourceId,
	}
	job := types.NewJob(
		ctx,
		types.Patch,
		types.Instance,
		&bodyUpdate,
		inv_handlers.InstanceURLParams{InstanceID: insResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Patch changes
	job = types.NewJob(ctx, types.Get, types.Instance, nil, inv_handlers.InstanceURLParams{
		InstanceID: insResource.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok := r.Payload.Data.(*api.Instance)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, instName, *gotRes.Name)
	assert.Equal(t, osResource2.GetResourceId(), *gotRes.Os.OsResourceID)

	bodyUpdate = api.Instance{
		Name:            &instName,
		SecurityFeature: &securityFeature,
	}
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Instance,
		&bodyUpdate,
		inv_handlers.InstanceURLParams{InstanceID: insResource.ResourceId},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	bodyUpdate = api.Instance{
		Name:           &instName,
		LocalAccountID: &localaccount.ResourceId,
	}
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Instance,
		&bodyUpdate,
		inv_handlers.InstanceURLParams{InstanceID: insResource.ResourceId},
	)
	// LocalAccountID cannot be updated if current state is anything other than unspecified
	r, err = h.Do(job)
	require.Error(t, err)
	assert.Nil(t, r)
}

func Test_InstanceHandler_Invalidate(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()

	hostResource := inv_testing.CreateHost(t, nil, nil)
	osResource := inv_testing.CreateOs(t)
	instanceResource := inv_testing.CreateInstance(t, hostResource, osResource)

	// test Put /invalidate
	job := types.NewJob(
		ctx,
		types.Put,
		types.Instance,
		&api.PutInstancesInstanceIDInvalidateResponse{},
		inv_handlers.InstanceURLParams{
			InstanceID: instanceResource.ResourceId,
			Action:     types.InstanceActionInvalidate,
		},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// test Put /invalidate Error - wrong params
	job = types.NewJob(
		ctx,
		types.Put,
		types.Instance,
		nil,
		inv_handlers.OUURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_InstanceHandler_Get(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}

	hostResource := inv_testing.CreateHost(t, nil, nil)
	osResource := inv_testing.CreateOs(t)
	instanceResource := inv_testing.CreateInstance(t, hostResource, osResource)

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	// Note that we must update InstanceResource to verify status fields, we may skip updating once
	// inv_testing.CreateHost will set status fields by default.
	//nolint:gosec // uint64 conversions are safe for testing
	updatedInstance := &computev1.InstanceResource{
		ProvisioningStatus:                "some provisioning status",
		ProvisioningStatusIndicator:       statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		ProvisioningStatusTimestamp:       uint64(time.Now().Unix()),
		InstanceStatus:                    "some instance status",
		InstanceStatusIndicator:           statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS,
		InstanceStatusTimestamp:           uint64(time.Now().Unix()),
		InstanceStatusDetail:              "5 of 5 components are Running",
		UpdateStatus:                      "some update status",
		UpdateStatusIndicator:             statusv1.StatusIndication_STATUS_INDICATION_ERROR,
		UpdateStatusTimestamp:             uint64(time.Now().Unix()),
		UpdateStatusDetail:                "{\"some_json\":[]}",
		TrustedAttestationStatus:          "some trusted attestation status",
		TrustedAttestationStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		TrustedAttestationStatusTimestamp: uint64(time.Now().Unix()),
	}

	_, err := inv_testing.TestClients[inv_testing.APIClient].Update(context.TODO(), instanceResource.GetResourceId(),
		&fieldmaskpb.FieldMask{Paths: []string{
			computev1.InstanceResourceFieldProvisioningStatus,
			computev1.InstanceResourceFieldProvisioningStatusIndicator,
			computev1.InstanceResourceFieldProvisioningStatusTimestamp,
			computev1.InstanceResourceFieldInstanceStatus,
			computev1.InstanceResourceFieldInstanceStatusIndicator,
			computev1.InstanceResourceFieldInstanceStatusTimestamp,
			computev1.InstanceResourceFieldInstanceStatusDetail,
			computev1.InstanceResourceFieldUpdateStatus,
			computev1.InstanceResourceFieldUpdateStatusIndicator,
			computev1.InstanceResourceFieldUpdateStatusTimestamp,
			computev1.InstanceResourceFieldUpdateStatusDetail,
			computev1.InstanceResourceFieldTrustedAttestationStatus,
			computev1.InstanceResourceFieldTrustedAttestationStatusIndicator,
			computev1.InstanceResourceFieldTrustedAttestationStatusTimestamp,
		}}, &inventory.Resource{
			Resource: &inventory.Resource_Instance{
				Instance: updatedInstance,
			},
		})
	require.NoError(t, err)

	job := types.NewJob(
		context.TODO(),
		types.Get,
		types.Instance,
		nil,
		inv_handlers.InstanceURLParams{InstanceID: instanceResource.ResourceId},
	)
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok := r.Payload.Data.(*api.Instance)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, instanceResource.Name, *gotRes.Name)
	assert.Equal(t, api.INSTANCESTATERUNNING, *gotRes.DesiredState)

	verifyInstanceStatusFields(t, gotRes, updatedInstance)

	// Get error - wrong params
	job = types.NewJob(
		context.TODO(),
		types.Get,
		types.Instance,
		nil,
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_InstanceHandler_Delete(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}

	hostResource := inv_testing.CreateHost(t, nil, nil)
	osResource := inv_testing.CreateOs(t)
	instanceResource := inv_testing.CreateInstanceNoCleanup(t, hostResource, osResource)

	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	job := types.NewJob(
		context.TODO(),
		types.Delete,
		types.Instance,
		nil,
		inv_handlers.InstanceURLParams{InstanceID: instanceResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, r.Status)

	// Delete error - wrong params
	job = types.NewJob(
		context.TODO(),
		types.Delete,
		types.Instance,
		nil,
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// Make sure to delete Instance, so Host/OS can be deleted.
	inv_testing.HardDeleteInstance(t, instanceResource.ResourceId)
}

func Test_Inventory_Instance_Integration(t *testing.T) {
	// verify the projection of the constants to Proto first;
	// we build a map using the field names of the proto stored in the
	// ProtoInstance* slices in internal/work/handlers/instance.go. Elements must
	// have a mapping key otherwise we throw an error if there is no
	// alignment with Instance proto in Inventory. Make sure to update these
	// two slices in internal/work/handlers/instance.go
	instanceResource := &computev1.InstanceResource{}
	validateInventoryIntegration(
		t,
		instanceResource,
		api.Instance{},
		inv_handlers.OpenAPIInstanceToProto,
		inv_handlers.OpenAPIInstanceToProtoExcluded,
		maps.Values(inv_handlers.OpenAPIInstanceToProto),
		true,
	)
}

// Test_InstanceHandler_InvMockClient_Errors evaluates all
// Instance handler methods with mock inventory client
// that returns errors.
func Test_InstanceHandler_InvMockClient_Errors(t *testing.T) {
	mockClient := utils.NewMockInventoryServiceClientError()
	client := &clients.InventoryClientHandler{
		InvClient: mockClient,
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	job := types.NewJob(
		context.TODO(), types.List, types.Instance,
		api.GetInstancesParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		}, nil,
	)
	_, err := h.Do(job)
	assert.Error(t, err)

	body := api.Instance{
		HostID:          &instHostID,
		OsID:            &instOSID,
		Kind:            &instKind,
		Name:            &instName,
		SecurityFeature: &securityFeature,
	}
	job = types.NewJob(
		context.TODO(), types.Post, types.Instance,
		&body, inv_handlers.InstanceURLParams{},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	bodyUpdate := api.Instance{
		Name: &instName,
	}
	job = types.NewJob(
		context.TODO(),
		types.Put,
		types.Instance,
		&bodyUpdate,
		inv_handlers.InstanceURLParams{InstanceID: "inst-12345678"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		context.TODO(),
		types.Get,
		types.Instance,
		nil,
		inv_handlers.InstanceURLParams{InstanceID: "inst-12345678"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		context.TODO(),
		types.Delete,
		types.Instance,
		nil,
		inv_handlers.InstanceURLParams{InstanceID: "inst-1234"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)
}
