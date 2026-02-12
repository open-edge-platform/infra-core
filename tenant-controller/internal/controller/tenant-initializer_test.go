// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/flags"
	inv_util "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	testutils "github.com/open-edge-platform/infra-core/tenant-controller/internal/testing"
)

const (
	lenovoResourceDefinitionFile = "../../configuration/default/resources-lenovo.json"
	resourceDefinitionFile       = "../../configuration/default/resources.json"
	tenant1ID                    = "11111111-1111-1111-1111-111111111111"
	tenant2ID                    = "22222222-2222-2222-2222-222222222222"
)

func TestNewInitResourcesProvider(t *testing.T) {
	*flags.FlagDisableCredentialsManagement = true

	configLoader, err := configuration.NewInitResourcesProvider(resourceDefinitionFile)
	require.NoError(t, err)
	lenovoconfigLoader, err := configuration.NewLenovoInitResourcesDefinitionLoader(lenovoResourceDefinitionFile)
	require.NoError(t, err)
	loaders := []configuration.InitResourcesProvider{
		configLoader, lenovoconfigLoader,
	}
	ic := testutils.CreateInvClient(t)
	tc := NewTenantInitializationController(loaders, ic, nil, false)
	require.Equal(t, ic, tc.ic)
	require.Equal(t, configLoader, tc.resourceDefinitionLoader[0])
	require.Equal(t, lenovoconfigLoader, tc.resourceDefinitionLoader[1])
}

func TestInitializeTenant_TenantAlreadyExist(t *testing.T) {
	icMock := new(invClientMock)
	icMock.On("GetTenantResource", mock.Anything, mock.AnythingOfType("string")).
		Return("tid", "rid", nil)
	tic := NewTenantInitializationController(nil, icMock, nil, false)

	sut := tic.InitializeTenant
	err := sut(context.TODO(), ProjectConfig{})
	require.NoError(t, err)
}

func TestInitializeTenant_FailOnGetTenantResource(t *testing.T) {
	expectedError := fmt.Errorf("cannot get tenant resource")
	icMock := new(invClientMock)
	icMock.On("GetTenantResource", mock.Anything, mock.AnythingOfType("string")).
		Return("", "", expectedError)
	tic := NewTenantInitializationController(nil, icMock, nil, false)

	sut := tic.InitializeTenant
	err := sut(context.TODO(), ProjectConfig{})
	require.Error(t, err)
	require.ErrorIs(t, err, expectedError)
}

type testResourcesProvider struct {
	resources []*inv_v1.Resource
}

func (g testResourcesProvider) Get() []*inv_v1.Resource {
	return g.resources
}

func TestInitializeTenant_FailOnUnsupportedResourceKind(t *testing.T) {
	icMock := new(invClientMock)
	icMock.On("GetTenantResource", mock.Anything, mock.AnythingOfType("string")).
		Return("", "", errors.Errorfc(codes.NotFound, "tenant does not exist"))

	resourceProvider := testResourcesProvider{
		resources: []*inv_v1.Resource{
			{}, // empty resource
		},
	}
	tic := NewTenantInitializationController([]configuration.InitResourcesProvider{resourceProvider}, icMock, nil, false)

	sut := tic.InitializeTenant
	err := sut(context.TODO(), ProjectConfig{})
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported resource kind")
}

func TestInitializeTenant_FailOnListAll(t *testing.T) {
	icMock := new(invClientMock)
	icMock.On("GetTenantResource", mock.Anything, mock.AnythingOfType("string")).
		Return("", "", errors.Errorfc(codes.NotFound, "tenant does not exist"))
	icMock.On("ListAll", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("cannot list all resources"))

	resourceProvider := testResourcesProvider{
		resources: []*inv_v1.Resource{
			{
				Resource: &inv_v1.Resource_Provider{Provider: &providerv1.ProviderResource{}},
			},
		},
	}
	tic := NewTenantInitializationController([]configuration.InitResourcesProvider{resourceProvider}, icMock, nil, false)

	sut := tic.InitializeTenant
	err := sut(context.TODO(), ProjectConfig{})
	require.Error(t, err)
	require.ErrorContains(t, err, "cannot list all")
}

func TestInitializeTenant_ExistingResourceShallBeSkip(t *testing.T) {
	projectConfig := ProjectConfig{
		TenantID: "anyTenant",
	}

	existingResource := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Provider{
			Provider: &providerv1.ProviderResource{
				Name:     "existingResource",
				TenantId: projectConfig.TenantID,
			},
		},
	}

	icMock := new(invClientMock)
	icMock.On("GetTenantResource", mock.Anything, mock.AnythingOfType("string")).
		Return("", "", errors.Errorfc(codes.NotFound, "tenant does not exist"))
	icMock.On("ListAll", mock.Anything, mock.Anything).
		Return([]*inv_v1.Resource{existingResource}, nil)
	icMock.On("CreateTenantResource", mock.Anything, mock.Anything).Return(nil, nil)

	resourceProvider := testResourcesProvider{
		resources: []*inv_v1.Resource{existingResource},
	}
	tic := NewTenantInitializationController([]configuration.InitResourcesProvider{resourceProvider}, icMock, nil, false)

	sut := tic.InitializeTenant

	err := sut(context.TODO(), projectConfig)
	require.NoError(t, err)
	icMock.AssertNotCalled(t, "CreateResource", mock.Anything, mock.Anything, mock.Anything)
	icMock.AssertExpectations(t)
}

func TestInitializeTenant_FailOnCreateResource(t *testing.T) {
	projectConfig := ProjectConfig{
		TenantID: "anyTenant",
	}

	existingResource := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Provider{
			Provider: &providerv1.ProviderResource{
				Name:     "existingResource",
				TenantId: projectConfig.TenantID,
			},
		},
	}

	icMock := new(invClientMock)
	icMock.On("GetTenantResource", mock.Anything, mock.AnythingOfType("string")).
		Return("", "", errors.Errorfc(codes.NotFound, "tenant does not exist"))
	icMock.On("ListAll", mock.Anything, mock.Anything).
		Return([]*inv_v1.Resource{}, nil)
	icMock.On("CreateResource", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("cannot create resource"))

	resourceProvider := testResourcesProvider{
		resources: []*inv_v1.Resource{existingResource},
	}
	tic := NewTenantInitializationController([]configuration.InitResourcesProvider{resourceProvider}, icMock, nil, false)

	sut := tic.InitializeTenant

	err := sut(context.TODO(), projectConfig)
	require.Error(t, err)
	icMock.AssertExpectations(t)
	icMock.AssertNotCalled(t, "CreateTenantResource", mock.Anything, mock.Anything)
}

func TestInitializeTenant(t *testing.T) {
	*flags.FlagDisableCredentialsManagement = true

	ic := testutils.CreateInvClient(t)
	cl, err := configuration.NewInitResourcesProvider(resourceDefinitionFile)
	require.NoError(t, err)
	lcl, err := configuration.NewLenovoInitResourcesDefinitionLoader(lenovoResourceDefinitionFile)
	require.NoError(t, err)
	ls := []configuration.InitResourcesProvider{
		cl, lcl,
	}

	sut := NewTenantInitializationController(ls, ic, nil, false)

	require.NoError(t, sut.InitializeTenant(context.TODO(), ProjectConfig{
		TenantID: tenant1ID,
	}))

	provider, err := inv_util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER)
	require.NoError(t, err)
	providers, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
		Resource: provider,
		Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", tenant1ID)).Build(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, providers)

	craftedFilter := fmt.Sprintf("%s = %q AND %s=%s AND %s=%s",
		providerv1.ProviderResourceFieldTenantId, tenant1ID,
		providerv1.ProviderResourceFieldProviderKind, providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL.String(),
		providerv1.ProviderResourceFieldProviderVendor, providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA.String(),
	)
	filter := &inv_v1.ResourceFilter{
		Resource: provider,
		Filter:   craftedFilter,
	}
	lenovoproviders, err := ic.ListAll(context.TODO(), filter)
	require.NoError(t, err)
	require.NotEmpty(t, lenovoproviders)

	telemetryGroup, err := inv_util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP)
	require.NoError(t, err)
	tgs, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
		Resource: telemetryGroup,
		Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", tenant1ID)).Build(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, tgs)

	tenant, err := inv_util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_TENANT)
	require.NoError(t, err)
	tenants, err := ic.ListAll(context.TODO(), &inv_v1.ResourceFilter{
		Resource: tenant,
		Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", tenant1ID)).Build(),
	})
	require.NoError(t, err)
	require.Len(t, tenants, 1)
}

//nolint:funlen // table-driven tests
func TestTenantController_contains(t *testing.T) {
	provider := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: &providerv1.ProviderResource{
			ResourceId:     "provider-12345678",
			ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
			ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
			Name:           "name",
			ApiEndpoint:    "ep",
			ApiCredentials: []string{"a", "b"},
			Config:         "foo/bar",
			TenantId:       tenant1ID,
		},
	}}

	sameProvider := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: &providerv1.ProviderResource{
			ResourceId:     "provider-12345678",
			ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
			ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
			Name:           "name",
			ApiEndpoint:    "ep",
			ApiCredentials: []string{"a", "b"},
			Config:         "foo/bar",
			TenantId:       tenant1ID,
		},
	}}

	sameProviderWithDifferentIDs := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: &providerv1.ProviderResource{
			ResourceId:     "provider-87654321",
			ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
			ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
			Name:           "name",
			ApiEndpoint:    "ep",
			ApiCredentials: []string{"a", "b"},
			Config:         "foo/bar",
			TenantId:       tenant2ID,
		},
	}}

	sameProviderNoTenantID := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: &providerv1.ProviderResource{
			ResourceId:     "provider-12345678",
			ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
			ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
			Name:           "name",
			ApiEndpoint:    "ep",
			ApiCredentials: []string{"a", "b"},
			Config:         "foo/bar",
		},
	}}

	sameProviderNoResourceID := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: &providerv1.ProviderResource{
			ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
			ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
			Name:           "name",
			ApiEndpoint:    "ep",
			ApiCredentials: []string{"a", "b"},
			Config:         "foo/bar",
			TenantId:       tenant1ID,
		},
	}}

	sameProviderNoIDs := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: &providerv1.ProviderResource{
			ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
			ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
			Name:           "name",
			ApiEndpoint:    "ep",
			ApiCredentials: []string{"a", "b"},
			Config:         "foo/bar",
		},
	}}

	tg := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-12345678",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant1ID,
		},
	}}

	sameTg := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-12345678",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant1ID,
		},
	}}

	anotherTg := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-12345678",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_HOST,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant1ID,
		},
	}}

	sameTgWithDifferentIDs := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-87654321",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant2ID,
		},
	}}

	sameTgNoTenantID := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-12345678",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
		},
	}}

	sameTgNoResourceID := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant1ID,
		},
	}}

	sameTgNoIDs := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-12345678",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant1ID,
		},
	}}

	sameTgButHasProfile := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: &telemetryv1.TelemetryGroupResource{
			ResourceId:    "telemetrygroup-12345678",
			Name:          "name",
			Kind:          telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
			CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
			Groups:        []string{"foo", "bar"},
			TenantId:      tenant1ID,
			Profiles: []*telemetryv1.TelemetryProfile{{
				ResourceId: "rid",
				TenantId:   tenant1ID,
			}},
		},
	}}

	tcs := []struct {
		name         string
		r1, r2       *inv_v1.Resource
		shallBeEqual bool
	}{
		{
			name:         "exactly same providers shall be equal",
			r1:           provider,
			r2:           sameProvider,
			shallBeEqual: true,
		},
		{
			name:         "same provider but second has not tenantID",
			r1:           provider,
			r2:           sameProviderNoTenantID,
			shallBeEqual: true,
		},
		{
			name:         "same provider but second has not resourceID",
			r1:           provider,
			r2:           sameProviderNoResourceID,
			shallBeEqual: true,
		},
		{
			name:         "same provider but second has IDs",
			r1:           provider,
			r2:           sameProviderNoIDs,
			shallBeEqual: true,
		},
		{
			name:         "same providers but different IDs",
			r1:           provider,
			r2:           sameProviderWithDifferentIDs,
			shallBeEqual: true,
		},
		// TGs
		{
			name:         "exactly same telemetry groups shall be equal",
			r1:           tg,
			r2:           sameTg,
			shallBeEqual: true,
		},
		{
			name:         "same telemetry groups but second has not tenantID",
			r1:           tg,
			r2:           sameTgNoTenantID,
			shallBeEqual: true,
		},
		{
			name:         "same telemetry groups but second has not resourceID",
			r1:           tg,
			r2:           sameTgNoResourceID,
			shallBeEqual: true,
		},
		{
			name:         "same telemetry groups but second has no IDs",
			r1:           tg,
			r2:           sameTgNoIDs,
			shallBeEqual: true,
		},
		{
			name:         "same telemetry groups but different IDs",
			r1:           tg,
			r2:           sameTgWithDifferentIDs,
			shallBeEqual: true,
		},
		{
			name:         "different telemetry groups",
			r1:           tg,
			r2:           anotherTg,
			shallBeEqual: false,
		},
		{
			name:         "same telemetry groups but one have backreferences",
			r1:           tg,
			r2:           sameTgButHasProfile,
			shallBeEqual: true,
		},
	}

	sut := NewTenantInitializationController(nil, nil, nil, false).contains

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			areEqual := sut([]*inv_v1.Resource{tc.r2}, tc.r1)
			assert.Equal(t, tc.shallBeEqual, areEqual)
		})
	}
}

func TestConfigureProvider(t *testing.T) {
	provider := new(providerv1.ProviderResource)
	provider.ProviderKind = providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL
	resource := &inv_v1.Resource{Resource: &inv_v1.Resource_Provider{
		Provider: provider,
	}}

	projectConfig := ProjectConfig{TenantID: tenant1ID}

	require.NoError(t, configureResource(resource, projectConfig))
	require.Equal(t, projectConfig.TenantID, resource.GetProvider().TenantId)
}

func TestConfigureTelemetryGroup(t *testing.T) {
	k, v := tenantIDProvider(tenant1ID)()
	require.Equal(t, "tenant_id", k)
	require.Equal(t, v, tenant1ID)

	tg := &inv_v1.Resource{Resource: &inv_v1.Resource_TelemetryGroup{
		TelemetryGroup: new(telemetryv1.TelemetryGroupResource),
	}}

	projectConfig := ProjectConfig{TenantID: tenant1ID}

	require.NoError(t, configureResource(tg, projectConfig))
	require.Equal(t, projectConfig.TenantID, tg.GetTelemetryGroup().TenantId)
}

type invClientMock struct {
	mock.Mock
}

func (i *invClientMock) Delete(ctx context.Context, tenantID, id string) (*inv_v1.DeleteResourceResponse, error) {
	args := i.Called(ctx, tenantID, id)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*inv_v1.DeleteResourceResponse)
	if !ok {
		return nil, errors.Errorf("unexpected type for DeleteResourceResponse: %T", args.Get(0))
	}
	return resp, args.Error(1)
}

func (i *invClientMock) DeleteAllResources(ctx context.Context, tenantID string, kind inv_v1.ResourceKind, enforce bool) error {
	args := i.Called(ctx, tenantID, kind, enforce)
	return args.Error(0)
}

func (i *invClientMock) GetTenantResourceInstance(ctx context.Context, tenantID string) (*tenantv1.Tenant, error) {
	args := i.Called(ctx, tenantID)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	tenant, ok := args.Get(0).(*tenantv1.Tenant)
	if !ok {
		return nil, errors.Errorf("unexpected type for Tenant: %T", args.Get(0))
	}
	return tenant, args.Error(1)
}

func (i *invClientMock) FindAll(ctx context.Context, filter *inv_v1.ResourceFilter) ([]*client.ResourceTenantIDCarrier, error) {
	args := i.Called(ctx, filter)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	tenantIDCarrier, ok := args.Get(0).([]*client.ResourceTenantIDCarrier)
	if !ok {
		return nil, errors.Errorf("unexpected type for []*ResourceTenantIDCarrier: %T", args.Get(0))
	}
	return tenantIDCarrier, args.Error(1)
}

func (i *invClientMock) HardDeleteTenantResource(ctx context.Context, tenantID, resourceID string) error {
	args := i.Called(ctx, tenantID, resourceID)
	return args.Error(0)
}

func (i *invClientMock) CreateResource(ctx context.Context, tenantID string, res *inv_v1.Resource) (*inv_v1.Resource, error) {
	args := i.Called(ctx, tenantID, res)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	resource, ok := args.Get(0).(*inv_v1.Resource)
	if !ok {
		return nil, errors.Errorf("unexpected type for Resource: %T", args.Get(0))
	}
	return resource, args.Error(1)
}

func (i *invClientMock) CreateTenantResource(ctx context.Context, tenantID string) (*inv_v1.Resource, error) {
	args := i.Called(ctx, tenantID)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	resource, ok := args.Get(0).(*inv_v1.Resource)
	if !ok {
		return nil, errors.Errorf("unexpected type for Resource: %T", args.Get(0))
	}
	return resource, args.Error(1)
}

func (i *invClientMock) GetTenantResource(ctx context.Context, tenantID string) (tid, rid string, err error) {
	args := i.Called(ctx, tenantID)
	return args.String(0), args.String(1), args.Error(2)
}

func (i *invClientMock) ListAll(ctx context.Context, filter *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error) {
	args := i.Called(ctx, filter)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	resources, ok := args.Get(0).([]*inv_v1.Resource)
	if !ok {
		return nil, errors.Errorf("unexpected type for []*Resource: %T", args.Get(0))
	}
	return resources, args.Error(1)
}

func (i *invClientMock) UpdateTenantResource(
	ctx context.Context, fm *fieldmaskpb.FieldMask, tenant *tenantv1.Tenant,
) (*inv_v1.Resource, error) {
	args := i.Called(ctx, fm, tenant)
	if args[0] == nil {
		return nil, args.Error(1)
	}
	resource, ok := args.Get(0).(*inv_v1.Resource)
	if !ok {
		return nil, errors.Errorf("unexpected type for Resource: %T", args.Get(0))
	}
	return resource, args.Error(1)
}
