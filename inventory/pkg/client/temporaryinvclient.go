// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache"
)

var FakeTenantID = "00000000-0000-0000-0000-000000000000"

// InventoryClient defines all the methods that inventoryClient must implement.
type InventoryClient interface {
	// Close unregisters the client from the inventory server and terminates the
	// gRPC connection. The client cannot be reused after this call. It is safe
	// to call this multiple times and from multiple goroutines.
	Close() error
	// List looks for inventory resources based on a filter definition
	// returning their objects. If no resources are found, an empty slice (of length 0) is returned.
	List(context.Context, *inv_v1.ResourceFilter) (*inv_v1.ListResourcesResponse, error)
	// ListAll looks for inventory resources based on the given filter and fieldMask
	// returning all objects that matches the filter. If no resources are found, an empty slice (of length 0) is returned.
	// Offset and limit set in the resource filter are ignored.
	ListAll(context.Context, *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error)
	// Find looks for inventory resources based on a filter definition
	// returning their IDs. If no resources are found, an empty slice (of length 0) is returned.
	Find(context.Context, *inv_v1.ResourceFilter) (*inv_v1.FindResourcesResponse, error)
	// FindAll looks for inventory resources based on the given filter and fieldMask
	// returning all the ID that matches the filter. If no resources are found, an empty slice (of length 0) is returned.
	// Offset and limit set in the resource filter are ignored.
	FindAll(context.Context, *inv_v1.ResourceFilter) ([]*ResourceTenantIDCarrier, error)
	// Get retrieves a resource from inventory based on its ID.
	Get(ctx context.Context, id string) (*inv_v1.GetResourceResponse, error)
	// Create creates a resource in inventory, providing its newly created ID in the response.
	Create(ctx context.Context, res *inv_v1.Resource) (*inv_v1.Resource, error)
	// Update updates a resource in inventory, given the resource ID, the fieldmask
	// to be applied on the resource fields, and the resource instance.
	Update(ctx context.Context, id string, fm *fieldmaskpb.FieldMask,
		res *inv_v1.Resource) (*inv_v1.Resource, error)
	// Delete deletes a resource from inventory based on its ID.
	Delete(ctx context.Context, id string) (*inv_v1.DeleteResourceResponse, error)
	// UpdateSubscriptions sets the resource kinds this clients will receive events for.
	UpdateSubscriptions(ctx context.Context, kinds []inv_v1.ResourceKind) error
	// ListInheritedTelemetryProfiles lists inherited telemetry profiles given the inheritBy parameter.
	// The given filter parameter can then be added to filter the list of inherited telemetry.
	// orderBy can be specified to order result by a given field.
	// limit and offset parameters are used to paginate results.
	ListInheritedTelemetryProfiles(
		ctx context.Context,
		inheritBy *inv_v1.ListInheritedTelemetryProfilesRequest_InheritBy,
		filter string,
		orderBy string,
		limit, offset uint32,
	) (*inv_v1.ListInheritedTelemetryProfilesResponse, error)
	GetHostByUUID(ctx context.Context, uuid string) (*computev1.HostResource, error)
	GetTreeHierarchy(context.Context, *inv_v1.GetTreeHierarchyRequest) ([]*inv_v1.GetTreeHierarchyResponse_TreeNode, error)
	GetSitesPerRegion(context.Context, *inv_v1.GetSitesPerRegionRequest) (*inv_v1.GetSitesPerRegionResponse, error)
	// TestingOnlySetClient allows to set the internal inventory service client API for testing purposes only.
	TestingOnlySetClient(inv_v1.InventoryServiceClient)
	// TestGetClientCache allows access to client cache for testing cache content.
	TestGetClientCache() *cache.InventoryCache
	// TestGetClientCacheUUID allows access to client cache for testing cache content.
	TestGetClientCacheUUID() *cache.InventoryCache

	GetTenantAwareInventoryClient() TenantAwareInventoryClient
}

func NewInventoryClient(ctx context.Context, cfg InventoryClientConfig) (InventoryClient, error) {
	tenantAwareClient, err := NewTenantAwareInventoryClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &temporaryInventoryClient{
		ic:       tenantAwareClient,
		tenantID: FakeTenantID,
	}, nil
}

type temporaryInventoryClient struct {
	ic       TenantAwareInventoryClient
	tenantID string
}

func (t *temporaryInventoryClient) GetTenantAwareInventoryClient() TenantAwareInventoryClient {
	return t.ic
}

func (t *temporaryInventoryClient) Close() error {
	return t.ic.Close()
}

func (t *temporaryInventoryClient) List(
	ctx context.Context, filter *inv_v1.ResourceFilter,
) (*inv_v1.ListResourcesResponse, error) {
	return t.ic.List(ctx, filter)
}

func (t *temporaryInventoryClient) ListAll(ctx context.Context, filter *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error) {
	return t.ic.ListAll(ctx, filter)
}

func (t *temporaryInventoryClient) Find(
	ctx context.Context, filter *inv_v1.ResourceFilter,
) (*inv_v1.FindResourcesResponse, error) {
	return t.ic.Find(ctx, filter)
}

func (t *temporaryInventoryClient) FindAll(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*ResourceTenantIDCarrier, error,
) {
	return t.ic.FindAll(ctx, filter)
}

func (t *temporaryInventoryClient) Get(ctx context.Context, id string) (*inv_v1.GetResourceResponse, error) {
	return t.ic.Get(ctx, FakeTenantID, id)
}

func (t *temporaryInventoryClient) Create(ctx context.Context, res *inv_v1.Resource) (*inv_v1.Resource, error) {
	if err := setTenantID(res, FakeTenantID); err != nil {
		return nil, err
	}
	return t.ic.Create(ctx, FakeTenantID, res)
}

func (t *temporaryInventoryClient) Update(ctx context.Context,
	id string,
	fm *fieldmaskpb.FieldMask,
	res *inv_v1.Resource,
) (*inv_v1.Resource, error) {
	return t.ic.Update(ctx, FakeTenantID, id, fm, res)
}

func (t *temporaryInventoryClient) Delete(ctx context.Context, id string) (*inv_v1.DeleteResourceResponse, error) {
	return t.ic.Delete(ctx, FakeTenantID, id)
}

func (t *temporaryInventoryClient) UpdateSubscriptions(ctx context.Context, kinds []inv_v1.ResourceKind) error {
	return t.ic.UpdateSubscriptions(ctx, FakeTenantID, kinds)
}

func (t *temporaryInventoryClient) ListInheritedTelemetryProfiles(
	ctx context.Context,
	inheritBy *inv_v1.ListInheritedTelemetryProfilesRequest_InheritBy,
	filter string,
	orderBy string,
	limit,
	offset uint32,
) (*inv_v1.ListInheritedTelemetryProfilesResponse, error) {
	return t.ic.ListInheritedTelemetryProfiles(ctx, FakeTenantID, inheritBy, filter, orderBy, limit, offset)
}

func (t *temporaryInventoryClient) GetHostByUUID(ctx context.Context, uuid string) (*computev1.HostResource, error) {
	return t.ic.GetHostByUUID(ctx, FakeTenantID, uuid)
}

func (t *temporaryInventoryClient) GetTreeHierarchy(
	ctx context.Context,
	request *inv_v1.GetTreeHierarchyRequest,
) ([]*inv_v1.GetTreeHierarchyResponse_TreeNode, error) {
	request.TenantId = FakeTenantID
	return t.ic.GetTreeHierarchy(ctx, request)
}

func (t *temporaryInventoryClient) GetSitesPerRegion(
	ctx context.Context,
	request *inv_v1.GetSitesPerRegionRequest,
) (*inv_v1.GetSitesPerRegionResponse, error) {
	request.TenantId = FakeTenantID
	return t.ic.GetSitesPerRegion(ctx, request)
}

func (t *temporaryInventoryClient) TestingOnlySetClient(client inv_v1.InventoryServiceClient) {
	t.ic.TestingOnlySetClient(client)
}

func (t *temporaryInventoryClient) TestGetClientCache() *cache.InventoryCache {
	return t.ic.TestGetClientCache()
}

func (t *temporaryInventoryClient) TestGetClientCacheUUID() *cache.InventoryCache {
	return t.ic.TestGetClientCacheUUID()
}

// setTenantID sets tenantID for any requested resource
// TODO: code below is temporary, ot allows inventory clients.
//
//nolint:funlen,cyclop // this code is temporary
func setTenantID(resource *inv_v1.Resource, tenantID string) error {
	var message proto.Message
	switch resource.GetResource().(type) {
	case *inv_v1.Resource_Region:
		message = resource.GetRegion()
	case *inv_v1.Resource_Site:
		message = resource.GetSite()
	case *inv_v1.Resource_Ou:
		message = resource.GetOu()
	case *inv_v1.Resource_Instance:
		message = resource.GetInstance()
	case *inv_v1.Resource_Host:
		message = resource.GetHost()
	case *inv_v1.Resource_Hoststorage:
		message = resource.GetHoststorage()
	case *inv_v1.Resource_Hostnic:
		message = resource.GetHostnic()
	case *inv_v1.Resource_Hostusb:
		message = resource.GetHostusb()
	case *inv_v1.Resource_Hostgpu:
		message = resource.GetHostgpu()
	case *inv_v1.Resource_NetworkSegment:
		message = resource.GetNetworkSegment()
	case *inv_v1.Resource_Netlink:
		message = resource.GetNetlink()
	case *inv_v1.Resource_Endpoint:
		message = resource.GetEndpoint()
	case *inv_v1.Resource_Ipaddress:
		message = resource.GetIpaddress()
	case *inv_v1.Resource_Provider:
		message = resource.GetProvider()
	case *inv_v1.Resource_Os:
		message = resource.GetOs()
	case *inv_v1.Resource_Singleschedule:
		message = resource.GetSingleschedule()
	case *inv_v1.Resource_Repeatedschedule:
		message = resource.GetRepeatedschedule()
	case *inv_v1.Resource_TelemetryGroup:
		message = resource.GetTelemetryGroup()
	case *inv_v1.Resource_TelemetryProfile:
		message = resource.GetTelemetryProfile()
	case *inv_v1.Resource_Workload:
		message = resource.GetWorkload()
	case *inv_v1.Resource_WorkloadMember:
		message = resource.GetWorkloadMember()
	case *inv_v1.Resource_RemoteAccess:
		message = resource.GetRemoteAccess()
	case *inv_v1.Resource_LocalAccount:
		message = resource.GetLocalAccount()
	case *inv_v1.Resource_CustomConfig:
		message = resource.GetCustomConfig()
	default:
		return fmt.Errorf("unknown resource type: %v", resource.GetResource())
	}

	if carrier, ok := message.(interface{ GetTenantId() string }); ok {
		if carrier.GetTenantId() != "" {
			return nil
		}
		refValue := reflect.ValueOf(message).Elem()
		tenantIDField := refValue.FieldByName("TenantId")
		tenantIDField.SetString(tenantID)
	}
	return nil
}
