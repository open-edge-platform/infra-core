// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	inv_util "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/invclient"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/nexus"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/util"
)

var (
	log = logging.GetLogger("tenant-controller")

	supportedResources = []typeInfo{{
		v:             providerv1.ProviderResource{},
		ignoredFields: []string{},
	}, {
		v:             telemetryv1.TelemetryGroupResource{},
		ignoredFields: []string{"Profiles"},
	}}
)

type typeInfo struct {
	v             any
	ignoredFields []string
}

type ProjectConfig struct{ TenantID string }

func NewTenantInitializationController(
	rdl []configuration.InitResourcesProvider,
	ic InventoryClient,
	nxc *nexus.Client,
) *TenantInitializationController {
	cmpOpts := createResourceComparisonOptions(supportedResources)
	return &TenantInitializationController{
		ic:                        ic,
		nxc:                       nxc,
		resourceDefinitionLoader:  rdl,
		resourceComparisonOptions: cmpOpts,
	}
}

type TenantInitializationController struct {
	ic                        InventoryClient
	resourceDefinitionLoader  []configuration.InitResourcesProvider
	resourceComparisonOptions []cmp.Option
	nxc                       *nexus.Client
}

func (tc *TenantInitializationController) InitializeTenant(ctx context.Context, config ProjectConfig) error {
	log.Info().Msgf("Initializing new tenant(%s)", config.TenantID)

	_, _, err := tc.ic.GetTenantResource(ctx, config.TenantID)
	if err == nil {
		log.Info().Msgf("Requested tenant(%s) already exists, skiping tenant initialization", config.TenantID)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	log.Debug().Msgf("No tenant(%s) resource found in Inventory, creating...", config.TenantID)

	cachingInvClient := invclient.NewInventoryClientCache(tc.ic)

	for _, resourceDefinitionLoader := range tc.resourceDefinitionLoader {
		for _, toBeCreated := range resourceDefinitionLoader.Get() {
			// prepare 'empty' resource to be used as a part of filter
			zero, err := inv_util.GetResourceFromKind(inv_util.GetResourceKindFromResource(toBeCreated))
			if err != nil {
				log.Err(err).Msgf("Cannot determine resource kind for resource %s", toBeCreated)
				return err
			}
			// get all resources of given type (tenant scope)
			existing, err := cachingInvClient.ListAll(
				ctx,
				&inv_v1.ResourceFilter{Resource: zero, Filter: tenantFilter(config.TenantID)},
			)
			if err != nil {
				log.Err(err).Msgf("Cannot ListAll resources(%v) for tenantID(%s)", zero, config.TenantID)
				return err
			}
			// ignore all requested resources which are already existing on inventory side
			if tc.contains(existing, toBeCreated) {
				log.Info().Msgf("Requested resource(%v) already exists, skipping creation ", toBeCreated)
				continue
			}
			// create if requested resources does not exist
			if err := configureResource(toBeCreated, config); err != nil {
				log.Err(err).Msgf("Cannot configure resource(%v)", toBeCreated)
				return err
			}
			if _, err := tc.ic.CreateResource(ctx, config.TenantID, toBeCreated); err != nil {
				log.Err(err).Msgf("Cannot create resource(%v)", toBeCreated)
				return err
			}
		}
	}
	// create tenant instance
	return second(tc.ic.CreateTenantResource(ctx, config.TenantID))
}

func (tc *TenantInitializationController) HandleEvent(we *client.WatchEvents) {
	log.Info().Msgf("[HANDLE-EVENT] HandleEvent received event: %v", we.Event)

	tenant := we.Event.GetResource().GetTenant()
	if tenant == nil {
		// Shall never happen
		log.Debug().Msgf("[HANDLE-EVENT-EXIT1] Given envelope doesn't contain tenant: %v", we.Event.GetResource())
		return
	}

	log.Info().Msgf("[HANDLE-EVENT] tenant=%s, WatcherOsmanager=%v, DesiredState=%v",
		tenant.GetTenantId(), tenant.WatcherOsmanager, tenant.DesiredState)

	if !tenant.WatcherOsmanager {
		log.Warn().Msgf("[HANDLE-EVENT-EXIT2] OS Resource Manager is not yet done with %s (WatcherOsmanager=false)",
			tenant.GetTenantId())
		return
	}
	if tenant.DesiredState != tenantv1.TenantState_TENANT_STATE_CREATED {
		log.Debug().Msgf("[HANDLE-EVENT-EXIT3] tenant=%s DesiredState=%v (not CREATED)",
			tenant.GetTenantId(), tenant.DesiredState)
		return
	}

	log.Info().Msgf("[HANDLE-EVENT] All conditions passed for tenant=%s, calling handleTenantDesiredStateCreated",
		tenant.GetTenantId())
	if err := tc.handleTenantDesiredStateCreated(tenant); err != nil {
		log.Err(err).Msgf("[HANDLE-EVENT] couldn't handle update event for Tenant(%s)", tenant.GetTenantId())
	}
}

func (tc *TenantInitializationController) handleTenantDesiredStateCreated(tenant *tenantv1.Tenant) error {
	log.Info().Msgf("[ACK-SEND] handleTenantDesiredStateCreated called for tenant=%s (CurrentState=%v)",
		tenant.GetTenantId(), tenant.CurrentState)

	if tenant.CurrentState != tenantv1.TenantState_TENANT_STATE_CREATED {
		log.Info().Msgf("[ACK-SEND] CurrentState is not CREATED, calling setCurrentStatusCreated for tenant=%s",
			tenant.GetTenantId())
		if err := tc.setCurrentStatusCreated(tenant); err != nil {
			log.Error().Err(err).Msgf("[ACK-SEND] setCurrentStatusCreated FAILED for tenant=%s", tenant.GetTenantId())
			return err
		}
		log.Info().Msgf("[ACK-SEND] setCurrentStatusCreated succeeded for tenant=%s", tenant.GetTenantId())
	}

	if tenant.CurrentState == tenantv1.TenantState_TENANT_STATE_CREATED {
		log.Info().Msgf("[ACK-SEND] Sending ACK via TryToSetActiveWatcherStatusIdle for tenant=%s",
			tenant.GetTenantId())
		if err := tc.nxc.TryToSetActiveWatcherStatusIdle(tenant.GetTenantId()); err != nil {
			log.Error().Err(err).Msgf("[ACK-SEND] TryToSetActiveWatcherStatusIdle FAILED for tenant=%s", tenant.GetTenantId())
			return err
		}
		log.Info().Msgf("[ACK-SEND] Successfully sent ACK for tenant=%s", tenant.GetTenantId())
	}

	return nil
}

func (tc *TenantInitializationController) setCurrentStatusCreated(tenant *tenantv1.Tenant) error {
	tenant.CurrentState = tenantv1.TenantState_TENANT_STATE_CREATED

	_, err := tc.ic.UpdateTenantResource(
		context.TODO(),
		&fieldmaskpb.FieldMask{Paths: []string{tenantv1.TenantFieldCurrentState}},
		tenant,
	)
	if err != nil {
		return errors.Errorf("cannot update tenant(%s)", tenant.TenantId)
	}
	return nil
}

func (tc *TenantInitializationController) contains(all []*inv_v1.Resource, requested *inv_v1.Resource) bool {
	for _, existing := range all {
		if cmp.Equal(existing, requested, tc.resourceComparisonOptions...) {
			return true
		}
		log.Debug().Msgf("resources are different: %s", cmp.Diff(existing, requested, tc.resourceComparisonOptions...))
	}
	return false
}

func createResourceComparisonOptions(supportedResources []typeInfo) []cmp.Option {
	resourceComparisonOptions := collections.MapSlice[typeInfo, cmp.Option](supportedResources, cmpOptionForType)
	return append(resourceComparisonOptions, cmpopts.IgnoreUnexported(inv_v1.Resource{}))
}

func cmpOptionForType(ti typeInfo) cmp.Option {
	return cmpopts.IgnoreFields(
		ti.v,
		append(ti.ignoredFields, "TenantId", "ResourceId", "unknownFields", "state", "sizeCache")...)
}

func tenantFilter(tenantID string) string {
	return filters.NewBuilderWith(filters.ValEq("tenant_id", tenantID)).Build()
}

func second[FIRST any, SECOND any](_ FIRST, s SECOND) SECOND {
	return s
}

type fieldConfiguration func() (fieldName string, fieldValue any)

type fieldConfigurations []fieldConfiguration

func configureResource(resource *inv_v1.Resource, config ProjectConfig) error {
	for _, fc := range prepareFieldConfiguration(config) {
		fieldName, fieldValue := fc()
		if err := util.SetResourceValue(resource, fieldName, fieldValue); err != nil {
			return err
		}
	}
	return nil
}

func prepareFieldConfiguration(config ProjectConfig) fieldConfigurations {
	return fieldConfigurations{tenantIDProvider(config.TenantID)}
}

func tenantIDProvider(tenantID string) fieldConfiguration {
	return func() (k string, v any) {
		return "tenant_id", tenantID
	}
}

type InventoryClient interface {
	CreateResource(ctx context.Context, tenantID string, resource *inv_v1.Resource) (*inv_v1.Resource, error)
	CreateTenantResource(ctx context.Context, tenantID string) (*inv_v1.Resource, error)
	Delete(ctx context.Context, tenantID, id string) (*inv_v1.DeleteResourceResponse, error)
	DeleteAllResources(ctx context.Context, tenantID string, kind inv_v1.ResourceKind, enforce bool) error
	GetTenantResource(ctx context.Context, tenantID string) (tid, rid string, err error)
	GetTenantResourceInstance(ctx context.Context, tenantID string) (*tenantv1.Tenant, error)
	FindAll(context.Context, *inv_v1.ResourceFilter) ([]*client.ResourceTenantIDCarrier, error)
	HardDeleteTenantResource(ctx context.Context, tenantID, resourceID string) error
	ListAll(context.Context, *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error)
	UpdateTenantResource(ctx context.Context, fm *fieldmaskpb.FieldMask, tenant *tenantv1.Tenant) (*inv_v1.Resource, error)
}
