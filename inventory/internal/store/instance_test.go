// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/localaccountresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/operatingsystemresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/providerresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/workloadmember"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	res_status "github.com/open-edge-platform/infra-core/inventory/v2/pkg/status"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
)

func Test_StrongRelations_On_Delete_Instance(t *testing.T) {
	t.Run("Instance_WorkloadMember", func(t *testing.T) {
		os := inv_testing.CreateOs(t)
		instance := inv_testing.CreateInstance(t, nil, os)
		workload := inv_testing.CreateWorkload(t)
		inv_testing.CreateWorkloadMember(t, workload, instance)

		err := inv_testing.HardDeleteInstanceAndReturnError(t, instance.ResourceId)
		assertStrongRelationError(t, err, "violates foreign key constraint")
	})
	t.Run("Instance_Os", func(t *testing.T) {
		os := inv_testing.CreateOs(t)
		inv_testing.CreateInstance(t, nil, os)

		err := inv_testing.DeleteResourceAndReturnError(t, os.ResourceId)
		assertStrongRelationError(t, err, "violates foreign key constraint")
	})
}

func Test_Create_Get_Delete_Instance(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	tenantID := uuid.NewString()
	host := dao.CreateHost(t, tenantID)
	os := dao.CreateOs(t, tenantID)
	provider := dao.CreateProvider(t, tenantID,
		"Test Provider1", inv_testing.ProviderKind(providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL))
	localaccount := dao.CreateLocalAccount(t, tenantID,
		"test-user",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ "+
			"test-user@example.com")
	oup := dao.CreateOSUpdatePolicy(t, tenantID,
		inv_testing.OsUpdatePolicyName("OsUpdatePolicy"), inv_testing.OSUpdatePolicyLatest())

	testcases := map[string]struct {
		in    *computev1.InstanceResource
		valid bool
	}{
		"CreateGoodVmInstance": {
			in: &computev1.InstanceResource{
				TenantId:        tenantID,
				Kind:            computev1.InstanceKind_INSTANCE_KIND_VM,
				DesiredState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
				DesiredOs:       os,
				CurrentOs:       os,
				VmMemoryBytes:   2 * util.Gigabyte,
				VmCpuCores:      4,
				VmStorageBytes:  16 * util.Gigabyte,
				SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
			},
			valid: true,
		},
		"CreateGoodMetalInstance": {
			in: &computev1.InstanceResource{
				TenantId:     tenantID,
				Kind:         computev1.InstanceKind_INSTANCE_KIND_METAL,
				DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:         host,
				DesiredOs:    os,
				CurrentOs:    os,
				Os:           os,
				ExistingCves: `[
{
  "cve_id": "CVE-000-000",
  "priority": "critical",
  "affected_packages": [
    "test-package-0.0.0",
    "test-2\test3"
  ]
}]`,
				RuntimePackages: `{
 "Repo": [
  {
   "Name": "openssl",
   "Version": "3.3.3-1",
   "Architecture": "x86_64",
   "Distribution": "emt3",
   "URL": "http://www.openssl.org/",
   "License": "Apache-2.0\"
  }
]
}`,
				OsUpdateAvailable: "Update to version 1.2.3 available!",
			},
			valid: true,
		},
		"CreateDiscoveredInstance": {
			in: &computev1.InstanceResource{
				TenantId:  tenantID,
				Kind:      computev1.InstanceKind_INSTANCE_KIND_METAL,
				Host:      host,
				DesiredOs: os,
			},
			valid: true,
		},
		"CreateDiscoveredInstanceWithProvider": {
			in: &computev1.InstanceResource{
				TenantId:  tenantID,
				Kind:      computev1.InstanceKind_INSTANCE_KIND_METAL,
				Host:      host,
				Provider:  provider,
				DesiredOs: os,
			},
			valid: true,
		},
		"CreateGoodMetalInstanceWithLocalAccount": {
			in: &computev1.InstanceResource{
				TenantId:     tenantID,
				Kind:         computev1.InstanceKind_INSTANCE_KIND_METAL,
				DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:         host,
				DesiredOs:    os,
				CurrentOs:    os,
				Localaccount: localaccount,
			},
			valid: true,
		},
		"CreateInstanceWithOsUpdatePolicy": {
			in: &computev1.InstanceResource{
				TenantId:       tenantID,
				Kind:           computev1.InstanceKind_INSTANCE_KIND_METAL,
				DesiredState:   computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:           host,
				DesiredOs:      os,
				CurrentOs:      os,
				OsUpdatePolicy: oup,
			},
			valid: true,
		},
		"CreateBadInstanceWithInvalidResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.InstanceResource{
				TenantId:       tenantID,
				ResourceId:     "instance-12345678",
				Kind:           computev1.InstanceKind_INSTANCE_KIND_VM,
				DesiredState:   computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:           host,
				DesiredOs:      os,
				VmMemoryBytes:  2 * util.Gigabyte,
				VmCpuCores:     4,
				VmStorageBytes: 16 * util.Gigabyte,
			},
			valid: false,
		},
		"CreateBadInstanceWithResourceIdSet": {
			// This tests case verifies that create requests with a resource ID
			// already set are rejected.
			in: &computev1.InstanceResource{
				TenantId:       tenantID,
				ResourceId:     "inst-12345678",
				Kind:           computev1.InstanceKind_INSTANCE_KIND_VM,
				DesiredState:   computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:           host,
				DesiredOs:      os,
				VmMemoryBytes:  2 * util.Gigabyte,
				VmCpuCores:     4,
				VmStorageBytes: 16 * util.Gigabyte,
			},
			valid: false,
		},
		"CreateBadInstanceWithoutOs": {
			in: &computev1.InstanceResource{
				TenantId:       tenantID,
				Kind:           computev1.InstanceKind_INSTANCE_KIND_VM,
				DesiredState:   computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:           host,
				VmMemoryBytes:  2 * util.Gigabyte,
				VmCpuCores:     4,
				VmStorageBytes: 16 * util.Gigabyte,
			},
			valid: false,
		},
		"CreateBadMetalInstanceWithVmFields": {
			in: &computev1.InstanceResource{
				TenantId:      tenantID,
				Kind:          computev1.InstanceKind_INSTANCE_KIND_METAL,
				DesiredState:  computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:          host,
				VmMemoryBytes: 2 * util.Gigabyte,
			},
			valid: false,
		},
		"CreateBadVMInstanceWithHost": {
			in: &computev1.InstanceResource{
				TenantId:      tenantID,
				Kind:          computev1.InstanceKind_INSTANCE_KIND_VM,
				DesiredState:  computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Host:          host,
				VmMemoryBytes: 2 * util.Gigabyte,
			},
			valid: false,
		},
		"CreateBadInstanceInvalidExistingCVEs": {
			in: &computev1.InstanceResource{
				Kind:      computev1.InstanceKind_INSTANCE_KIND_METAL,
				Host:      host,
				DesiredOs: os,
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: tc.in},
			}
			invClient := inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient()

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create instance
			cInstResp, err := invClient.Create(ctx, tenantID, createresreq)
			instanceResID := cInstResp.GetInstance().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateInstance() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = instanceResID // Update with created resource ID.
				tc.in.InstanceStatus = res_status.DefaultInstanceStatus
				tc.in.ProvisioningStatus = res_status.DefaultProvisioningStatus
				tc.in.UpdateStatus = res_status.DefaultUpdateStatus
				tc.in.TrustedAttestationStatus = res_status.DefaultTrustedAttestationStatus
				tc.in.CreatedAt = cInstResp.GetInstance().GetCreatedAt()
				tc.in.UpdatedAt = cInstResp.GetInstance().GetUpdatedAt()
				assertSameResource(t, createresreq, cInstResp, nil)
				if !tc.valid {
					t.Errorf("CreateInstance() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := invClient.Get(ctx, tenantID, "inst-12345678")
				require.Error(t, err)

				// get instance
				getresp, err := invClient.Get(ctx, tenantID, instanceResID)
				require.NoError(t, err, "GetInstance() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetInstance()); !eq {
					t.Errorf("GetInstance() data not equal: %v", diff)
				}

				// delete non-existent first
				err = dao.DeleteResourceAndReturnError(t, tenantID, "inst-12345678")
				require.Error(t, err)

				// Remove instance.
				dao.HardDeleteInstance(t, tenantID, instanceResID)

				// get after complete Delete of instance, should fail as Instance is 2-phase deleted
				_, err = invClient.Get(ctx, tenantID, instanceResID)
				require.Error(t, err, "Failure - Instance was not deleted, but should be deleted")
			}
		})
	}
}

func Test_CreateInstanceFromRM(t *testing.T) {
	host := inv_testing.CreateHost(t, nil, nil)
	os := inv_testing.CreateOs(t)

	instRes := computev1.InstanceResource{
		Kind:            computev1.InstanceKind_INSTANCE_KIND_METAL,
		DesiredState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
		DesiredOs:       os,
		CurrentOs:       os,
		Host:            host,
		SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
	}

	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &instRes,
		},
	}
	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// create instance from RM client
	cInstResp, err := inv_testing.TestClients[inv_testing.RMClient].Create(ctx, createresreq)
	require.NoError(t, err)
	instCreateResp := cInstResp.GetInstance()
	instanceResID := instCreateResp.GetResourceId()
	instRes.ResourceId = instanceResID

	// update should fail when setting the desired state from RM
	upInst := inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &computev1.InstanceResource{
				DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
			},
		},
	}
	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{
			instanceresource.FieldDesiredState,
		},
	}
	upRes, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instanceResID, fieldMask, &upInst)
	require.Error(t, err)
	assert.Nil(t, upRes)

	// read instance from RM client
	getresp, err := inv_testing.TestClients[inv_testing.RMClient].Get(ctx, instanceResID)
	require.NoError(t, err, "GetInstance() failed")
	if eq, diff := inv_testing.ProtoEqualOrDiff(instCreateResp, getresp.GetResource().GetInstance()); !eq {
		t.Errorf("GetInstance() data not equal: %v", diff)
	}

	// read instance from RM client
	getresp, err = inv_testing.TestClients[inv_testing.APIClient].Get(ctx, instanceResID)
	require.NoError(t, err, "GetInstance() failed")
	if eq, diff := inv_testing.ProtoEqualOrDiff(instCreateResp, getresp.GetResource().GetInstance()); !eq {
		t.Errorf("GetInstance() data not equal: %v", diff)
	}

	// Remove instance.
	inv_testing.HardDeleteInstance(t, instanceResID)
}

func Test_UpdateInstance(t *testing.T) {
	site := inv_testing.CreateSite(t, nil, nil)
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host := inv_testing.CreateHost(t, site, provider)
	host.Site = site
	host.Provider = provider
	host2 := inv_testing.CreateHost(t, site, provider)
	host2.Site = site
	host2.Provider = provider
	os := inv_testing.CreateOs(t)
	os2 := inv_testing.CreateOs(t)
	localaccount := inv_testing.CreateLocalAccount(t,
		"test-user",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ test-user1@example.com")

	// create Instance to update
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &computev1.InstanceResource{
				Host:            host,
				DesiredOs:       os,
				VmMemoryBytes:   2 * util.Gigabyte,
				VmCpuCores:      4,
				VmStorageBytes:  16 * util.Gigabyte,
				DesiredState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
				SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_NONE,
			},
		},
	}

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cInstResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
	require.NoError(t, err)
	instanceResID := inv_testing.GetResourceIDOrFail(t, cInstResp)
	t.Cleanup(func() { inv_testing.HardDeleteInstance(t, instanceResID) })

	testcases := map[string]struct {
		in           *computev1.InstanceResource
		resourceID   string
		fieldMask    *fieldmaskpb.FieldMask
		valid        bool
		expErrorCode codes.Code
	}{
		"UpdateInstance1": {
			in: &computev1.InstanceResource{
				VmCpuCores:   8,
				CurrentState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				DesiredOs:    os,
				CurrentOs:    os,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldVMCPUCores,
					instanceresource.FieldCurrentState,
					instanceresource.EdgeDesiredOs,
					instanceresource.EdgeCurrentOs,
				},
			},
			valid: true,
		},
		"UpdateInstance2": {
			in: &computev1.InstanceResource{
				VmCpuCores:   8,
				CurrentState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldVMCPUCores,
					instanceresource.FieldCurrentState,
				},
			},
			valid: true,
		},
		"UpdateInstance3": {
			in:         &computev1.InstanceResource{},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.FieldCurrentState},
			},
			valid: true,
		},
		"UpdateInstance4": {
			in: &computev1.InstanceResource{
				Provider: provider,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.EdgeProvider},
			},
			valid: true,
		},
		"UpdateInstanceOsNoop": {
			in: &computev1.InstanceResource{
				DesiredOs: os,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.EdgeDesiredOs},
			},
			valid: true,
		},
		"UpdateInstanceOs2": {
			in: &computev1.InstanceResource{
				DesiredOs: os2,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.EdgeDesiredOs},
			},
			valid: true,
		},
		"UpdateInstanceClearHost": {
			in: &computev1.InstanceResource{
				Host: nil,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.EdgeHost},
			},
			valid: true,
		},
		"UpdateInstanceHost2": {
			in: &computev1.InstanceResource{
				Host: host2,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.EdgeHost},
			},
			valid: true,
		},
		"UpdateInstanceHostNoop": {
			in: &computev1.InstanceResource{
				Host: host,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.EdgeHost},
			},
			valid: true,
		},
		"UpdateInstanceNoFieldMask": {
			in: &computev1.InstanceResource{
				VmCpuCores:   8,
				CurrentState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				DesiredOs:    os,
			},
			resourceID:   instanceResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateMetalInstanceStatus": {
			in: &computev1.InstanceResource{
				InstanceStatus:          "Some instance status",
				InstanceStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS,
				InstanceStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldInstanceStatus,
					instanceresource.FieldInstanceStatusIndicator,
					instanceresource.FieldInstanceStatusTimestamp,
				},
			},
			valid: true,
		},
		"UpdateMetalInstanceUpdateStatus": {
			in: &computev1.InstanceResource{
				UpdateStatus:          "Some update status",
				UpdateStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				UpdateStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
				UpdateStatusDetail:    "Some update status detail",
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldUpdateStatus,
					instanceresource.FieldUpdateStatusIndicator,
					instanceresource.FieldUpdateStatusTimestamp,
					instanceresource.FieldUpdateStatusDetail,
				},
			},
			valid: true,
		},
		"UpdateMetalInstanceProvisioningStatus": {
			in: &computev1.InstanceResource{
				ProvisioningStatus:          "Some provisioning status",
				ProvisioningStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				ProvisioningStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldProvisioningStatus,
					instanceresource.FieldProvisioningStatusIndicator,
					instanceresource.FieldProvisioningStatusTimestamp,
				},
			},
			valid: true,
		},
		"UpdateMetalInstanceTrustedAttestationStatus": {
			in: &computev1.InstanceResource{
				TrustedAttestationStatus:          "AttestationVerified",
				TrustedAttestationStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
				TrustedAttestationStatusTimestamp: uint64(time.Now().Unix()), //nolint:gosec // This is a test
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldTrustedAttestationStatus,
					instanceresource.FieldTrustedAttestationStatusIndicator,
					instanceresource.FieldTrustedAttestationStatusTimestamp,
				},
			},
			valid: true,
		},
		"UpdateInstanceInvalidFieldMask": {
			in: &computev1.InstanceResource{
				VmCpuCores: 8,
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_FIELD"},
			},
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateVmInstanceHost": {
			in: &computev1.InstanceResource{
				Kind: computev1.InstanceKind_INSTANCE_KIND_VM,
				Host: host,
			},
			resourceID:   instanceResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateMetalInstanceVmFields": {
			in: &computev1.InstanceResource{
				Kind:          computev1.InstanceKind_INSTANCE_KIND_METAL,
				VmMemoryBytes: 2 * util.Gigabyte,
			},
			resourceID:   instanceResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateSecurityFeatureNotValid": {
			in: &computev1.InstanceResource{
				SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
			},
			resourceID:   instanceResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateResourceIDNotFound": {
			in: &computev1.InstanceResource{
				CurrentState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
			},
			resourceID: "inst-12345678",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldCurrentState,
				},
			},
			valid:        false,
			expErrorCode: codes.NotFound,
		},
		"UpdateInstanceLocalAccountFail": {
			in: &computev1.InstanceResource{
				CurrentState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				Localaccount: localaccount,
			},
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldCurrentState,
					instanceresource.EdgeLocalaccount,
				},
			},
			resourceID:   instanceResID,
			valid:        false,
			expErrorCode: codes.InvalidArgument,
		},
		"UpdateInstanceDetailStatus": {
			in: &computev1.InstanceResource{
				InstanceStatusDetail: "2 of 5 components Running",
			},
			resourceID: instanceResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					instanceresource.FieldInstanceStatusDetail,
				},
			},
			valid: true,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := inv_testing.TestClients[inv_testing.RMClient].Update(
				ctx,
				tc.resourceID,
				tc.fieldMask,
				updateresreq,
			)

			if !tc.valid {
				require.Errorf(t, err, "UpdateResource() succeeded but should have failed")
				assert.Equal(t, tc.expErrorCode, status.Code(err))
				assert.Nil(t, upRes)
				return
			}
			require.NoErrorf(t, err, "UpdateResource() failed: %s", err)

			assert.Equal(t, tc.resourceID, upRes.GetInstance().GetResourceId())

			// validate update via a get
			getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, tc.resourceID)
			require.NoError(t, err, "GetResource() failed")

			assertSameResource(t, updateresreq, getresp.GetResource(), tc.fieldMask)
		})
	}
}

func Test_InstanceStateTransitionFromUntrusted(t *testing.T) {
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host1 := inv_testing.CreateHost(t, nil, provider)
	os1 := inv_testing.CreateOs(t)
	instance1 := inv_testing.CreateInstance(t, host1, os1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// make a instance untrusted
	upRes, err := inv_testing.GetClient(t, inv_testing.RMClient).Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					CurrentState: computev1.InstanceState_INSTANCE_STATE_UNTRUSTED,
				},
			},
		},
	)
	require.NoError(t, err, "UpdateInstance() failed")
	assert.Equal(t, computev1.InstanceState_INSTANCE_STATE_UNTRUSTED, upRes.GetInstance().GetCurrentState())

	// try to update other fields, this should be allowed
	upRes, err = inv_testing.TestClients[inv_testing.APIClient].Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.FieldName}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					Name: "test",
				},
			},
		},
	)
	require.NoError(t, err, "UpdateInstance() failed")
	assert.Equal(t, "test", upRes.GetInstance().GetName())

	// After the instance state is set to untrusted, the only allowed next state is deleted.
	// This test tries to change the instance state from untrusted to running, and expects an error.
	upRes, err = inv_testing.TestClients[inv_testing.APIClient].Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.FieldDesiredState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				},
			},
		},
	)
	require.Error(t, err, "UpdateInstance() to RUNNING should fail")
	assert.Nil(t, upRes)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// move instance to DELETED state (allowed transition from UNTRUSTED)
	upRes, err = inv_testing.TestClients[inv_testing.APIClient].Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.FieldDesiredState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					DesiredState: computev1.InstanceState_INSTANCE_STATE_DELETED,
				},
			},
		},
	)
	require.NoError(t, err, "UpdateInstance() to DELETED should not fail")
	assert.Equal(t, computev1.InstanceState_INSTANCE_STATE_DELETED, upRes.GetInstance().GetDesiredState())
}

func Test_InstanceLocalAccountUpdate(t *testing.T) {
	host1 := inv_testing.CreateHost(t, nil, nil)
	os1 := inv_testing.CreateOs(t)
	localaccount := inv_testing.CreateLocalAccount(t,
		"test-user",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ test-user1@example.com",
	)
	instance1 := inv_testing.CreateInstanceWithLocalAccount(t, host1, os1, localaccount)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// make progress on the instance
	upRes, err := inv_testing.GetClient(t, inv_testing.RMClient).Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.FieldCurrentState}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					CurrentState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				},
			},
		},
	)
	require.NoError(t, err, "UpdateInstance() failed")
	assert.Equal(t, computev1.InstanceState_INSTANCE_STATE_RUNNING, upRes.GetInstance().GetCurrentState())

	// After the instance is no longer in the initial state, we cannot touch anymore
	// local account.
	upRes, err = inv_testing.TestClients[inv_testing.APIClient].Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.EdgeLocalaccount}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					Localaccount: localaccount,
				},
			},
		},
	)
	require.Error(t, err, "UpdateInstance() with LocalAccount should fail")
	assert.Nil(t, upRes)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Nilling the local account should also fail
	upRes, err = inv_testing.TestClients[inv_testing.APIClient].Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{instanceresource.EdgeLocalaccount}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{},
			},
		},
	)
	require.Error(t, err, "UpdateInstance() with LocalAccount should fail")
	assert.Nil(t, upRes)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	curRes, err := inv_testing.TestClients[inv_testing.APIClient].Get(
		ctx,
		instance1.ResourceId,
	)
	require.NoError(t, err, "GetInstance() with LocalAccount should not fail")
	assert.NotNil(t, curRes)
	require.NotNil(t, curRes.GetResource().GetInstance().GetLocalaccount())

	// Should not fail and should not take effect
	upRes, err = inv_testing.TestClients[inv_testing.APIClient].Update(
		ctx,
		instance1.ResourceId,
		&fieldmaskpb.FieldMask{Paths: []string{}},
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{
				Instance: &computev1.InstanceResource{
					Localaccount: localaccount,
				},
			},
		},
	)
	require.NoError(t, err, "UpdateInstance() with LocalAccount should not fail")
	assert.NotNil(t, upRes)
	assert.NotNil(t, upRes.GetInstance().GetLocalaccount())
}

func Test_FilterInstances(t *testing.T) {
	provider := inv_testing.CreateProvider(t, "Test Provider1")
	host1 := inv_testing.CreateHost(t, nil, provider)
	host2 := inv_testing.CreateHost(t, nil, provider)
	os1 := inv_testing.CreateOs(t)
	os2 := inv_testing.CreateOs(t)
	os3 := inv_testing.CreateOs(t)

	// create Instances to find
	createresreq1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &computev1.InstanceResource{
				Host:            host1,
				DesiredOs:       os1,
				CurrentOs:       os1,
				VmMemoryBytes:   2 * util.Gigabyte,
				VmCpuCores:      4,
				VmStorageBytes:  16 * util.Gigabyte,
				DesiredState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
				SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
			},
		},
	}

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &computev1.InstanceResource{
				Host:           host2,
				DesiredOs:      os2,
				VmMemoryBytes:  2 * util.Gigabyte,
				VmCpuCores:     4,
				VmStorageBytes: 16 * util.Gigabyte,
				DesiredState:   computev1.InstanceState_INSTANCE_STATE_RUNNING,
			},
		},
	}

	createresreqEmpty := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &computev1.InstanceResource{
				VmCpuCores:   4,
				DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
				DesiredOs:    os3,
				Provider:     provider,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cInstResp1, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq1)
	require.NoError(t, err)
	instance1ResID := inv_testing.GetResourceIDOrFail(t, cInstResp1)
	cInstResp2, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	require.NoError(t, err)
	instance2ResID := inv_testing.GetResourceIDOrFail(t, cInstResp2)
	cInstRespEmpty, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreqEmpty)
	require.NoError(t, err)
	instanceEmptyResID := inv_testing.GetResourceIDOrFail(t, cInstRespEmpty)
	t.Cleanup(func() { inv_testing.HardDeleteInstance(t, instance1ResID) })
	t.Cleanup(func() { inv_testing.HardDeleteInstance(t, instance2ResID) })
	t.Cleanup(func() { inv_testing.HardDeleteInstance(t, instanceEmptyResID) })

	instExp1 := cInstResp1.GetInstance()
	instExp1.ResourceId = instance1ResID
	workload1 := inv_testing.CreateWorkload(t)
	workloadMember1 := inv_testing.CreateWorkloadMember(t, workload1, instExp1)
	workloadMember1.Workload = workload1
	instExp1.WorkloadMembers = append(instExp1.WorkloadMembers, workloadMember1)

	instExp2 := cInstResp2.GetInstance()
	instExp2.ResourceId = instance2ResID

	instExpEmpty := cInstRespEmpty.GetInstance()
	instExpEmpty.ResourceId = instanceEmptyResID

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.InstanceResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.InstanceResource{instExp1, instExp2, instExpEmpty},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: instanceresource.FieldResourceID,
			},
			resources: []*computev1.InstanceResource{instExp1, instExp2, instExpEmpty},
			valid:     true,
		},
		"FilterByEmptyResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = ""`, instanceresource.FieldResourceID),
			},
			valid: true,
		},
		"FilterByResourceIdEq": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, instanceresource.FieldResourceID, instExp2.ResourceId),
			},
			resources: []*computev1.InstanceResource{instExp2},
			valid:     true,
		},
		"FilterHost": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeHost,
					hostresource.FieldResourceID, host1.GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
		"FilterByHasHost": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, instanceresource.EdgeHost),
			},
			resources: []*computev1.InstanceResource{instExp1, instExp2},
			valid:     true,
		},
		"FilterByProviderID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeProvider, providerresource.FieldResourceID,
					instExpEmpty.GetProvider().GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExpEmpty},
			valid:     true,
		},
		"FilterByWorkloadMemberID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeWorkloadMembers, workloadmember.FieldResourceID,
					instExp1.GetWorkloadMembers()[0].GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
		"FilterByDesiredOsID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeDesiredOs, operatingsystemresource.FieldResourceID,
					instExp1.GetDesiredOs().GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
		"FilterByInstalledOsID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeCurrentOs, operatingsystemresource.FieldResourceID,
					instExp1.GetCurrentOs().GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
		"FilterHostEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, instanceresource.EdgeHost),
			},
			resources: []*computev1.InstanceResource{instExpEmpty},
			valid:     true,
		},
		"FilterDesiredOSEmpty": {
			// OS cannot be empty
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, instanceresource.EdgeDesiredOs),
			},
			valid: true,
		},
		"FilterBySecurityFeatures": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %s`, instanceresource.FieldSecurityFeature, instExp1.SecurityFeature),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
		"FilterWorkloadMembers": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeWorkloadMembers,
					workloadmember.FieldResourceID, workloadMember1.GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
		"FilterWorkloadMembersEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, instanceresource.EdgeWorkloadMembers),
			},
			resources: []*computev1.InstanceResource{instExp2, instExpEmpty},
			valid:     true,
		},
		"FilterProvider": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeProvider,
					providerresource.FieldResourceID, provider.GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instExpEmpty},
			valid:     true,
		},
		"FilterProviderEmpty": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, instanceresource.EdgeProvider),
			},
			resources: []*computev1.InstanceResource{instExp1, instExp2},
			valid:     true,
		},
		"FilterLimit": {
			in: &inv_v1.ResourceFilter{
				Limit: 3,
			},
			resources: []*computev1.InstanceResource{instExp1, instExp2, instExpEmpty},
			valid:     true,
		},
		"FilterWithOffsetLimit1": {
			in: &inv_v1.ResourceFilter{
				Offset: 5,
				Limit:  0,
			},
			valid: true,
		},
		"FilterWithOffsetLimit2": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  5,
			},
			resources: []*computev1.InstanceResource{instExp1, instExp2, instExpEmpty},
			valid:     true,
		},
		"FilterInvalidEdge": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, "invalid_edge"),
			},
			valid: false,
		},
		"FilterInvalidField": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, "invalid_field", "some-value"),
			},
			valid: false,
		},
		"FilterByHasWorkloadMemebers": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, instanceresource.EdgeWorkloadMembers),
			},
			resources: []*computev1.InstanceResource{instExp1},
			valid:     true,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_Instance{}} // Set the resource kind
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterInstances() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterInstances() succeeded but should have failed")
				}
			}

			// only get/delete if valid test with non-zero returned response and hasn't failed, otherwise may segfault
			if !t.Failed() && tc.valid {
				if len(findres.Resources) != len(tc.resources) {
					t.Errorf("Expected to obtain %d Resource IDs, but obtained back %d Resource IDs",
						len(tc.resources), len(findres.Resources))
				}

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findres.Resources)

				if !reflect.DeepEqual(resIDs, findres.Resources) {
					t.Errorf(
						"FilterInstances() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListInstances() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListInstances() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.InstanceResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetInstance())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					instanceEdgesOnlyResourceID(expected)
					instanceEdgesOnlyResourceID(resources[i])

					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListInstances() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func instanceEdgesOnlyResourceID(expected *computev1.InstanceResource) {
	if expected.Host != nil {
		expected.Host = &computev1.HostResource{ResourceId: expected.Host.ResourceId}
	}
	if expected.DesiredOs != nil {
		expected.DesiredOs = &osv1.OperatingSystemResource{ResourceId: expected.DesiredOs.ResourceId}
	}
	if expected.CurrentOs != nil {
		expected.CurrentOs = &osv1.OperatingSystemResource{ResourceId: expected.CurrentOs.ResourceId}
	}
}

func Test_NestedFilterInstances(t *testing.T) {
	site := inv_testing.CreateSite(t, nil, nil)
	host1 := inv_testing.CreateHost(t, site, nil)
	host1.Site = site
	host2 := inv_testing.CreateHost(t, nil, nil)
	host3 := inv_testing.CreateHost(t, nil, nil)
	host4 := inv_testing.CreateHost(t, nil, nil)
	host5 := inv_testing.CreateHost(t, nil, nil)
	host6 := inv_testing.CreateHost(t, nil, nil)
	os1 := inv_testing.CreateOs(t)
	os2 := inv_testing.CreateOs(t)

	instance1 := inv_testing.CreateInstance(t, host1, os1)
	workload1 := inv_testing.CreateWorkload(t)
	workloadMember1 := inv_testing.CreateWorkloadMember(t, workload1, instance1)
	workloadMember1.Workload = workload1
	instance1.WorkloadMembers = append(instance1.WorkloadMembers, workloadMember1)
	instance1.Host = host1
	instance1.DesiredOs = os1
	instance1.CurrentOs = os1

	instance2 := inv_testing.CreateInstance(t, host2, os1)
	workload2 := inv_testing.CreateWorkload(t)
	workloadMember2 := inv_testing.CreateWorkloadMember(t, workload2, instance2)
	workloadMember2.Workload = workload2
	instance4 := inv_testing.CreateInstance(t, host5, os2)
	workloadMember3 := inv_testing.CreateWorkloadMember(t, workload1, instance4)
	workloadMember3.Workload = workload1
	instance4.WorkloadMembers = append(instance4.WorkloadMembers, workloadMember3)
	instance4.Host = host5
	instance4.DesiredOs = os2
	instance4.CurrentOs = os2
	instance2.WorkloadMembers = append(instance2.WorkloadMembers, workloadMember2)
	instance2.Host = host2
	instance2.DesiredOs = os1
	instance2.CurrentOs = os1

	instance3 := inv_testing.CreateInstance(t, host3, os2)
	instance3.Host = host3
	instance3.DesiredOs = os2
	instance3.CurrentOs = os2

	provider := inv_testing.CreateProvider(t, "Test Provider1")
	instanceWithProvider := inv_testing.CreateInstanceWithProvider(t, host4, os2, provider)
	instanceWithProvider.Host = host4
	instanceWithProvider.DesiredOs = os2
	instanceWithProvider.CurrentOs = os2
	instanceWithProvider.Provider = provider

	localaccount := inv_testing.CreateLocalAccount(t,
		"test-user",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILtu+7Pdtj6ihyFynecnd+155AdxqvHhMRxvxdcQ8/D/ test-user1@example.com",
	)
	instanceWithLocalAccount := inv_testing.CreateInstanceWithLocalAccount(t, host6, os2, localaccount)
	instanceWithLocalAccount.Host = host6
	instanceWithLocalAccount.DesiredOs = os2
	instanceWithLocalAccount.CurrentOs = os2
	instanceWithLocalAccount.Localaccount = localaccount

	testcases := map[string]struct {
		in                *inv_v1.ResourceFilter
		resources         []*computev1.InstanceResource
		valid             bool
		expectedCodeError codes.Code
	}{
		"FilterByHostUuid": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeHost,
					hostresource.FieldUUID, host1.GetUuid()),
			},
			resources: []*computev1.InstanceResource{instance1},
			valid:     true,
		},
		"FilterByLocalaccount": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeLocalaccount,
					localaccountresource.FieldResourceID, localaccount.GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instanceWithLocalAccount},
			valid:     true,
		},
		"FilterByLocalaccountUsername": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeLocalaccount,
					localaccountresource.FieldUsername, localaccount.GetUsername()),
			},
			resources: []*computev1.InstanceResource{instanceWithLocalAccount},
			valid:     true,
		},
		"FilterByEmptySite": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s.%s)`, instanceresource.EdgeHost, hostresource.EdgeSite),
			},
			resources: []*computev1.InstanceResource{
				instance2, instance3,
				instance4, instanceWithProvider, instanceWithLocalAccount,
			},
			valid: true,
		},
		"FilterBySite": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s = %q`, instanceresource.EdgeHost, hostresource.EdgeSite,
					siteresource.FieldResourceID, site.GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instance1},
			valid:     true,
		},
		"FilterByOsName": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`,
					instanceresource.EdgeDesiredOs, operatingsystemresource.FieldName, os1.GetName()),
			},
			resources: []*computev1.InstanceResource{
				instance1, instance2, instance3, instance4,
				instanceWithProvider, instanceWithLocalAccount,
			},
			valid: true,
		},
		"FilterByWorkloadMemberID": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %q`, instanceresource.EdgeWorkloadMembers,
					workloadmember.FieldResourceID, workloadMember1.GetResourceId()),
			},
			resources: []*computev1.InstanceResource{instance1},
			valid:     true,
		},
		"FilterByHasWorkloadMembers": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`has(%s)`, instanceresource.EdgeWorkloadMembers),
			},
			resources: []*computev1.InstanceResource{instance1, instance2, instance4},
			valid:     true,
		},
		"FilterByNotHasWorkloadMembers": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`NOT has(%s)`, instanceresource.EdgeWorkloadMembers),
			},
			resources: []*computev1.InstanceResource{instanceWithLocalAccount, instance3, instanceWithProvider},
			valid:     true,
		},
		"FilterByProviderVendor": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s = %s`, instanceresource.EdgeProvider,
					providerresource.FieldProviderVendor, providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LXCA),
			},
			resources: []*computev1.InstanceResource{instanceWithProvider},
			valid:     true,
		},
		"FailTooDeep": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s.%s.%s.%s.%s.%s = %q`, instanceresource.EdgeHost,
					hostresource.EdgeSite, siteresource.EdgeRegion, regionresource.EdgeParentRegion,
					regionresource.EdgeParentRegion, regionresource.FieldResourceID, ""),
			},
			valid:             false,
			expectedCodeError: codes.InvalidArgument,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_Instance{}} // Set the resource kind

			// Test FIND
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expectedCodeError, status.Code(err))
			} else {
				require.NoError(t, err)

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findres.Resources)

				if !reflect.DeepEqual(resIDs, findres.Resources) {
					t.Errorf(
						"FilterInstances() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			// Test LIST
			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expectedCodeError, status.Code(err))
			} else {
				require.NoError(t, err)
				require.Len(t, listres.Resources, len(tc.resources))

				resources := make([]*computev1.InstanceResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetInstance())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListInstances() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func Test_One2One_Relation_InstanceHost(t *testing.T) {
	host1 := inv_testing.CreateHost(t, nil, nil)
	os := inv_testing.CreateOs(t)
	inv_testing.CreateInstance(t, host1, os)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	inst2Req := computev1.InstanceResource{
		Kind:         computev1.InstanceKind_INSTANCE_KIND_METAL,
		DesiredState: computev1.InstanceState_INSTANCE_STATE_RUNNING,
		DesiredOs:    os,
		Host:         host1,
	}
	_, err := inv_testing.GetClient(t, inv_testing.APIClient).Create(ctx,
		&inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{Instance: &inst2Req},
		})
	require.Error(t, err)
}

// Test_Instance_NestedEagerLoading verifies that the eager loading of the workload associated to the member is provided
// when querying instances (via Get or List). The same is validated for Site and Provider associated to the Host.
func Test_Instance_NestedEagerLoading(t *testing.T) {
	site := inv_testing.CreateSite(t, nil, nil)
	provider := inv_testing.CreateProvider(t, inv_testing.DummyProviderName)
	host := inv_testing.CreateHost(t, site, provider)
	os := inv_testing.CreateOs(t)
	instance := inv_testing.CreateInstance(t, host, os)
	workload := inv_testing.CreateWorkload(t)
	inv_testing.CreateWorkloadMember(t, workload, instance)

	// build a context for gRPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	getRes, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, instance.ResourceId)
	require.NoError(t, err)

	getInst := getRes.GetResource().GetInstance()
	if eq, diff := inv_testing.ProtoEqualOrDiff(workload, getInst.GetWorkloadMembers()[0].Workload); !eq {
		t.Errorf("Workload nested eager loading via get: %v", diff)
	}
	if eq, diff := inv_testing.ProtoEqualOrDiff(site, getInst.GetHost().Site); !eq {
		t.Errorf("Workload site eager loading via get:  %v", diff)
	}
	if eq, diff := inv_testing.ProtoEqualOrDiff(provider, getInst.GetHost().Provider); !eq {
		t.Errorf("Workload provider eager loading via get:  %v", diff)
	}

	listRes, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Instance{},
		},
	})
	require.NoError(t, err)
	require.Len(t, listRes.GetResources(), 1)

	respInstance := listRes.Resources[0].GetResource().GetInstance()
	if eq, diff := inv_testing.ProtoEqualOrDiff(workload, respInstance.GetWorkloadMembers()[0].Workload); !eq {
		t.Errorf("Workload nested eager loading via list: %v", diff)
	}

	if eq, diff := inv_testing.ProtoEqualOrDiff(site, respInstance.GetHost().Site); !eq {
		t.Errorf("Workload site eager loading via list:  %v", diff)
	}
	if eq, diff := inv_testing.ProtoEqualOrDiff(provider, respInstance.GetHost().Provider); !eq {
		t.Errorf("Workload provider eager loading via list:  %v", diff)
	}
}

func Test_InstanceEnumStateMap(t *testing.T) {
	v, err := store.InstanceEnumStateMap("invalid_input",
		int32(computev1.InstanceState_INSTANCE_STATE_RUNNING))
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestInstanceMTSanity(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	suite.Run(t, &struct{ mt }{
		mt: mt{
			createResource: func(tenantID string) (string, *inv_v1.Resource) {
				host := dao.CreateHost(t, tenantID)
				os := dao.CreateOs(t, tenantID)
				instance := dao.CreateInstance(t, tenantID, host, os)
				res, err := util.WrapResource(instance)
				require.NoError(t, err)
				return instance.GetResourceId(), res
			},
		},
	})
}

func TestSoftDeleteResources_Instances(t *testing.T) {
	suite.Run(t, &softDeleteAllResourcesSuite{
		createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
			tenantID := uuid.NewString()
			host := dao.CreateHost(t, tenantID)
			os := dao.CreateOs(t, tenantID)
			return tenantID, len([]any{
				dao.CreateInstance(t, tenantID, host, os),
			})
		},
		resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		deletedClause: filters.ValEq(
			computev1.InstanceResourceFieldDesiredState, computev1.InstanceState_INSTANCE_STATE_DELETED),
		notDeletedClause: filters.ValNotEq(
			computev1.InstanceResourceFieldDesiredState, computev1.InstanceState_INSTANCE_STATE_DELETED),
	})
}

func TestHardDeleteResources_Instances(t *testing.T) {
	suite.Run(t, &hardDeleteAllResourcesSuite{
		createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
			tenantID := uuid.NewString()
			host := dao.CreateHost(t, tenantID)
			os := dao.CreateOs(t, tenantID)
			return tenantID, len([]any{
				dao.CreateInstanceNoCleanup(t, tenantID, host, os),
			})
		},
		resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
	})
}
