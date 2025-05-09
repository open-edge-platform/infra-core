package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/status"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

func Test_DefaultHostResourceStatus(t *testing.T) {
	t.Run("DefaultStatusesAreSetCorrectly", func(t *testing.T) {
		dao := inv_testing.NewInvResourceDAOOrFail(t)
		host := dao.CreateHost(t, tenantIDZero)

		require.Equal(t, status.DefaultHostStatus, host.GetHostStatus())
		require.Equal(t, status.DefaultOnboardingStatus, host.GetOnboardingStatus())
		require.Equal(t, status.DefaultRegistrationStatus, host.GetRegistrationStatus())
	})

	t.Run("CustomStatusesAreNotOverwritten", func(t *testing.T) {
		customStatus := "Test Status"

		dao := inv_testing.NewInvResourceDAOOrFail(t)
		host := dao.CreateHostWithOpts(t, tenantIDZero, true, func(c *computev1.HostResource) {
			c.HostStatus = customStatus
			c.OnboardingStatus = customStatus
		})

		require.Equal(t, customStatus, host.GetHostStatus())
		require.Equal(t, customStatus, host.GetOnboardingStatus())
		require.Equal(t, status.DefaultRegistrationStatus, host.GetRegistrationStatus())
	})
}

func Test_DefaultInstanceResourceStatus(t *testing.T) {
	t.Run("DefaultStatusesAreSetCorrectly", func(t *testing.T) {
		dao := inv_testing.NewInvResourceDAOOrFail(t)
		host := dao.CreateHost(t, tenantIDZero)
		os := dao.CreateOs(t, tenantIDZero)
		instance := dao.CreateInstance(t, tenantIDZero, host, os)

		require.Equal(t, status.DefaultInstanceStatus, instance.GetInstanceStatus())
		require.Equal(t, status.DefaultProvisioningStatus, instance.GetProvisioningStatus())
		require.Equal(t, status.DefaultUpdateStatus, instance.GetUpdateStatus())
		require.Equal(t, status.DefaultTrustedAttestationStatus, instance.GetTrustedAttestationStatus())
	})

	t.Run("UpdatedStatusesAreNotOverwritten", func(t *testing.T) {
		dao := inv_testing.NewInvResourceDAOOrFail(t)
		host := dao.CreateHost(t, tenantIDZero)
		os := dao.CreateOs(t, tenantIDZero)
		instance := dao.CreateInstance(t, tenantIDZero, host, os)

		// expect the default statuses to be set
		require.Equal(t, status.DefaultInstanceStatus, instance.GetInstanceStatus())
		require.Equal(t, status.DefaultProvisioningStatus, instance.GetProvisioningStatus())
		require.Equal(t, status.DefaultUpdateStatus, instance.GetUpdateStatus())
		require.Equal(t, status.DefaultTrustedAttestationStatus, instance.GetTrustedAttestationStatus())

		// update the instance status
		newStatus := "Updated Status"
		instance.InstanceStatus = newStatus

		// build a context for gRPC
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		upRes, err := inv_testing.TestClients[inv_testing.APIClient].Update(
			ctx,
			instance.GetResourceId(),
			&fieldmaskpb.FieldMask{
				Paths: []string{instanceresource.FieldInstanceStatus},
			},
			&inv_v1.Resource{
				Resource: &inv_v1.Resource_Instance{Instance: instance},
			},
		)

		require.NoError(t, err, "failed to update instance resource")
		// expect the updated instance status instead of the default status
		require.Equal(t, newStatus, upRes.GetInstance().GetInstanceStatus())
		require.Equal(t, status.DefaultProvisioningStatus, instance.GetProvisioningStatus())
		require.Equal(t, status.DefaultUpdateStatus, instance.GetUpdateStatus())
		require.Equal(t, status.DefaultTrustedAttestationStatus, instance.GetTrustedAttestationStatus())

	})

}
