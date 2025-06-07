// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/clientreg"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy/rbac"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var zlog = logging.GetLogger("InfraInvInvgRPC")

type InventorygRPCServer struct {
	inv_v1.UnimplementedInventoryServiceServer
	CR                   *clientreg.ClientReg
	INVPOLICY            *policy.Policy
	RBAC                 *rbac.Policy
	AuthorizationEnabled bool
	IS                   *store.InvStore
}

func NewInventoryServer(dbURLWriter, dbURLReader, policyFile string, enableTracing, enableAuth bool) *InventorygRPCServer {
	cr := clientreg.NewClientReg(enableTracing)
	invstore := store.NewStore(dbURLWriter, dbURLReader)

	// initialize policy agent
	invPolicy, err := policy.New(policyFile)
	if err != nil {
		zlog.InfraSec().Fatal().Err(err).Msg("Failed to initialize policy agent")
	}

	rbacPolicy, err := rbac.New(policyFile)
	if err != nil {
		zlog.InfraSec().Fatal().Err(err).Msg("Failed to start OPA for RBAC")
	}

	zlog.InfraSec().Info().Msgf("OPA agent successfully initialized.....")
	zlog.InfraSec().Info().Msgf("Authorization enabled is %v", enableAuth)

	iserv := InventorygRPCServer{
		IS:                   invstore,
		CR:                   cr,
		INVPOLICY:            invPolicy,
		RBAC:                 rbacPolicy,
		AuthorizationEnabled: enableAuth,
	}

	return &iserv
}

func (srv *InventorygRPCServer) Authorize(ctx context.Context, request interface{}) error {
	// ToDo - remove these lines when Authentication is enabled E2E
	if !srv.AuthorizationEnabled {
		// if authorization is disabled, just return nil
		return nil
	}

	// Create the input data map from the request context
	ctxClaims := metautils.ExtractIncoming(ctx)

	var err error
	switch req := request.(type) {
	case *inv_v1.CreateResourceRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.CreateKey)
	case *inv_v1.ListResourcesRequest, *inv_v1.ListInheritedTelemetryProfilesRequest, *inv_v1.GetTreeHierarchyRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.ListKey)
	case *inv_v1.FindResourcesRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.FindKey)
	case *inv_v1.GetResourceRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.GetKey)
	case *inv_v1.UpdateResourceRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.UpdateKey)
	case *inv_v1.DeleteResourceRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.DeleteKey)
	case *inv_v1.DeleteAllResourcesRequest:
		err = srv.RBAC.Verify(ctxClaims, rbac.DeleteKey)
	default:
		zlog.InfraSec().InfraError("unspecified request type %v", req).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unspecified request type %v", req)
	}

	if err != nil {
		return err
	}

	zlog.Debug().Msgf("Call is authorized")
	return nil
}

func (srv *InventorygRPCServer) SubscribeEvents(
	in *inv_v1.SubscribeEventsRequest,
	stream inv_v1.InventoryService_SubscribeEventsServer,
) error {
	zlog.Info().Msgf("SubscribeEvents from client: %v", in)

	// Register the new client.
	clientUUID, err := srv.CR.RegisterClient(clientreg.ClientInfo{
		Name:          in.GetName(),
		Version:       in.GetVersion(),
		ClientKind:    in.GetClientKind(),
		ResourceKinds: in.GetSubscribedResourceKinds(),
		Stream:        stream,
	})
	if err != nil {
		return err
	}
	defer srv.CR.ExitClient(clientUUID)

	// Send the assigned UUID back to the client.
	if err = stream.Send(&inv_v1.SubscribeEventsResponse{ClientUuid: clientUUID}); err != nil {
		zlog.Warn().Msgf("Problem streaming response to: %s", clientUUID)
		return errors.Wrap(err)
	}

	// Block until exit.
	<-stream.Context().Done()
	zlog.InfraSec().Info().Msgf("SubscribeEvents stream disconnect client: %v", clientUUID)
	return nil
}

func (srv *InventorygRPCServer) ChangeSubscribeEvents(
	ctx context.Context,
	in *inv_v1.ChangeSubscribeEventsRequest,
) (*inv_v1.ChangeSubscribeEventsResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("ChangeSubscribeEvents from client: %v", in)
	err := srv.CR.UpdateClient(in.GetClientUuid(), in.GetSubscribedResourceKinds())
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &inv_v1.ChangeSubscribeEventsResponse{}, nil
}

// CRUD functions

func (srv *InventorygRPCServer) CreateResource(
	ctx context.Context,
	in *inv_v1.CreateResourceRequest,
) (*inv_v1.Resource, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("CreateResource for UUID %v", in.ClientUuid)

	err := error(nil)

	// authorize call first
	err = srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	// fetch client_kind info from client registration map
	// and validate the information
	clientKind, err := srv.extractClientKind(in.ClientUuid)
	if err != nil {
		return nil, err
	}

	// policy evaluation
	err = srv.INVPOLICY.Verify(clientKind, in)
	// error handling for policy violation
	if err != nil {
		return nil, err
	}

	var res *inv_v1.Resource

	switch in.GetResource().GetResource().(type) {
	// location.proto
	case *inv_v1.Resource_Region:
		res, err = srv.IS.CreateRegion(ctx, in.GetResource().GetRegion())
	case *inv_v1.Resource_Site:
		res, err = srv.IS.CreateSite(ctx, in.GetResource().GetSite())

	// ou.proto
	case *inv_v1.Resource_Ou:
		res, err = srv.IS.CreateOu(ctx, in.GetResource().GetOu())

	// instance.proto
	case *inv_v1.Resource_Instance:
		res, err = srv.IS.CreateInstance(ctx, in.GetResource().GetInstance())

	// host.proto
	case *inv_v1.Resource_Host:
		res, err = srv.IS.CreateHost(ctx, in.GetResource().GetHost())

	case *inv_v1.Resource_Hoststorage:
		res, err = srv.IS.CreateHoststorage(ctx, in.GetResource().GetHoststorage())
	case *inv_v1.Resource_Hostnic:
		res, err = srv.IS.CreateHostnic(ctx, in.GetResource().GetHostnic())
	case *inv_v1.Resource_Hostusb:
		res, err = srv.IS.CreateHostusb(ctx, in.GetResource().GetHostusb())
	case *inv_v1.Resource_Hostgpu:
		res, err = srv.IS.CreateHostgpu(ctx, in.GetResource().GetHostgpu())

	// network.proto
	case *inv_v1.Resource_NetworkSegment:
		res, err = srv.IS.CreateNetworkSegment(ctx, in.GetResource().GetNetworkSegment())
	case *inv_v1.Resource_Netlink:
		res, err = srv.IS.CreateNetlink(ctx, in.GetResource().GetNetlink())
	case *inv_v1.Resource_Endpoint:
		res, err = srv.IS.CreateEndpoint(ctx, in.GetResource().GetEndpoint())
	case *inv_v1.Resource_Ipaddress:
		res, err = srv.IS.CreateIPAddress(ctx, in.GetResource().GetIpaddress())

	// provider.proto
	case *inv_v1.Resource_Provider:
		res, err = srv.IS.CreateProvider(ctx, in.GetResource().GetProvider())

	// os.proto
	case *inv_v1.Resource_Os:
		res, err = srv.IS.CreateOs(ctx, in.GetResource().GetOs())

	// schedule.proto
	case *inv_v1.Resource_Singleschedule:
		res, err = srv.IS.CreateSingleSchedule(ctx, in.GetResource().GetSingleschedule())
	case *inv_v1.Resource_Repeatedschedule:
		res, err = srv.IS.CreateRepeatedSchedule(ctx, in.GetResource().GetRepeatedschedule())

	// telemetry.proto
	case *inv_v1.Resource_TelemetryGroup:
		res, err = srv.IS.CreateTelemetryGroup(ctx, in.GetResource().GetTelemetryGroup())
	case *inv_v1.Resource_TelemetryProfile:
		res, err = srv.IS.CreateTelemetryProfile(ctx, in.GetResource().GetTelemetryProfile())

	// workload.proto
	case *inv_v1.Resource_Workload:
		res, err = srv.IS.CreateWorkload(ctx, in.GetResource().GetWorkload())
	case *inv_v1.Resource_WorkloadMember:
		res, err = srv.IS.CreateWorkloadMember(ctx, in.GetResource().GetWorkloadMember())

	case *inv_v1.Resource_RemoteAccess:
		res, err = srv.IS.CreateRemoteAccessConfig(ctx, in.GetResource().GetRemoteAccess())

	case *inv_v1.Resource_Tenant:
		res, err = srv.IS.CreateTenant(ctx, in.GetResource().GetTenant())
	// localaccount.proto
	case *inv_v1.Resource_LocalAccount:
		res, err = srv.IS.CreateLocalAccount(ctx, in.GetResource().GetLocalAccount())

	case *inv_v1.Resource_OsUpdatePolicy:
		res, err = srv.IS.CreateOSUpdatePolicy(ctx, in.GetResource().GetOsUpdatePolicy())

	case *inv_v1.Resource_OsUpdateRun:
		res, err = srv.IS.CreateOSUpdateRun(ctx, in.GetResource().GetOsUpdateRun())
	default:
		zlog.InfraSec().InfraError("unknown Resource Kind: %T", in.Resource).Msg("create resource error")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Resource Kind: %T", in.Resource)
	}

	if err != nil {
		return nil, err
	}

	// notify others
	srv.CR.StreamNotify(
		ctx,
		inv_v1.SubscribeEventsResponse_EVENT_KIND_CREATED,
		res,
		in.ClientUuid,
	)

	return res, err
}

func (srv *InventorygRPCServer) ListResources(
	ctx context.Context,
	in *inv_v1.ListResourcesRequest,
) (*inv_v1.ListResourcesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("ListResources for UUID %v", in.ClientUuid)

	// authorize call first
	err := srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	err = validator.ValidateMessage(in)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	resp, total, err := srv.IS.ListResources(ctx, in.Filter)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		zlog.Debug().Msgf("no resources found for filter: %s", in.Filter)
	}
	totalInt, err := util.IntToInt32(total)
	if err != nil {
		return nil, err
	}
	return &inv_v1.ListResourcesResponse{
		Resources:     resp,
		HasNext:       len(resp)+int(in.Filter.GetOffset()) < total,
		TotalElements: totalInt,
	}, nil
}

func (srv *InventorygRPCServer) FindResources(
	ctx context.Context,
	in *inv_v1.FindResourcesRequest,
) (*inv_v1.FindResourcesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("FindResources for UUID %v", in.ClientUuid)

	// authorize call first
	err := srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	err = validator.ValidateMessage(in)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	tenResIDs, total, err := srv.IS.FindResources(ctx, in.Filter)
	if err != nil {
		return nil, err
	}

	if len(tenResIDs) == 0 {
		zlog.Debug().Msgf("no resources found for filter: %s", in.Filter)
		tenResIDs = make([]*client.ResourceTenantIDCarrier, 0)
	}
	totalInt, err := util.IntToInt32(total)
	if err != nil {
		return nil, err
	}
	return &inv_v1.FindResourcesResponse{
		Resources:     tenResIDs,
		HasNext:       len(tenResIDs)+int(in.Filter.GetOffset()) < total,
		TotalElements: totalInt,
	}, nil
}

func (srv *InventorygRPCServer) GetResource(
	ctx context.Context,
	in *inv_v1.GetResourceRequest,
) (*inv_v1.GetResourceResponse, error) {
	var err error

	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("GetResource %s for UUID %s", in.ResourceId, in.ClientUuid)

	// authorize call first
	err = srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	kind, err := util.GetResourceKindFromResourceID(in.ResourceId)
	if err != nil {
		return nil, err
	}

	// response (empty, filled in switch below)
	gresresp := &inv_v1.GetResourceResponse{}

	switch kind {
	// location.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_REGION:
		gresresp.Resource, gresresp.RenderedMetadata, err = srv.IS.GetRegion(ctx, in.ResourceId, in.GetTenantId())
	case inv_v1.ResourceKind_RESOURCE_KIND_SITE:
		gresresp.Resource, gresresp.RenderedMetadata, err = srv.IS.GetSite(ctx, in.ResourceId, in.GetTenantId())

	// ou.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_OU:
		gresresp.Resource, gresresp.RenderedMetadata, err = srv.IS.GetOu(ctx, in.ResourceId, in.GetTenantId())

	// instance.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:
		gresresp.Resource, err = srv.IS.GetInstance(ctx, in.ResourceId)

	// host.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_HOST:
		gresresp.Resource, gresresp.RenderedMetadata, err = srv.IS.GetHost(ctx, in.ResourceId, in.GetTenantId())
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:
		gresresp.Resource, err = srv.IS.GetHoststorage(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:
		gresresp.Resource, err = srv.IS.GetHostnic(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:
		gresresp.Resource, err = srv.IS.GetHostusb(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:
		gresresp.Resource, err = srv.IS.GetHostgpu(ctx, in.ResourceId)

	// network.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:
		gresresp.Resource, err = srv.IS.GetNetworkSegment(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:
		gresresp.Resource, err = srv.IS.GetNetlink(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:
		gresresp.Resource, err = srv.IS.GetEndpoint(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:
		gresresp.Resource, err = srv.IS.GetIPAddress(ctx, in.ResourceId)

	// provider.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:
		gresresp.Resource, err = srv.IS.GetProvider(ctx, in.ResourceId)

	// os.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_OS:
		gresresp.Resource, err = srv.IS.GetOs(ctx, in.ResourceId)

	// schedule.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:
		gresresp.Resource, err = srv.IS.GetSingleSchedule(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:
		gresresp.Resource, err = srv.IS.GetRepeatedSchedule(ctx, in.ResourceId)

	// telemetry.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:
		gresresp.Resource, err = srv.IS.GetTelemetryGroup(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE:
		gresresp.Resource, err = srv.IS.GetTelemetryProfile(ctx, in.ResourceId)

	// workload.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:
		gresresp.Resource, err = srv.IS.GetWorkload(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:
		gresresp.Resource, err = srv.IS.GetWorkloadMember(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF:
		gresresp.Resource, err = srv.IS.GetRemoteAccessConfig(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_TENANT:
		gresresp.Resource, err = srv.IS.GetTenant(ctx, in.ResourceId)

	// localaccount.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:
		gresresp.Resource, err = srv.IS.GetLocalAccount(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:
		gresresp.Resource, err = srv.IS.GetOSUpdatePolicy(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:
		gresresp.Resource, err = srv.IS.GetOSUpdateRun(ctx, in.ResourceId)
	default:
		zlog.InfraSec().InfraError("unknown Resource Kind: %s", kind).Msg("get resource parse error")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Resource Kind: %s", kind)
	}

	return gresresp, err
}

func (srv *InventorygRPCServer) doUpdateResource(
	ctx context.Context,
	kind inv_v1.ResourceKind,
	in *inv_v1.UpdateResourceRequest,
) (*inv_v1.Resource, bool, error) {
	var hardDelete bool
	var err error
	var res *inv_v1.Resource

	switch kind {
	// location.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_REGION:
		res, err = srv.IS.UpdateRegion(ctx, in.ResourceId, in.GetResource().GetRegion(), in.GetFieldMask(), in.GetTenantId())
	case inv_v1.ResourceKind_RESOURCE_KIND_SITE:
		res, err = srv.IS.UpdateSite(ctx, in.ResourceId, in.GetResource().GetSite(), in.GetFieldMask(), in.GetTenantId())

	// ou.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_OU:
		res, err = srv.IS.UpdateOu(ctx, in.ResourceId, in.GetResource().GetOu(), in.GetFieldMask(), in.GetTenantId())

	// network.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:
		res, err = srv.IS.UpdateNetworkSegment(ctx, in.ResourceId, in.GetResource().GetNetworkSegment(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:
		res, hardDelete, err = srv.IS.UpdateNetlink(ctx, in.ResourceId, in.GetResource().GetNetlink(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:
		res, err = srv.IS.UpdateEndpoint(ctx, in.ResourceId, in.GetResource().GetEndpoint(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:
		res, hardDelete, err = srv.IS.UpdateIPAddress(ctx, in.ResourceId, in.GetResource().GetIpaddress(), in.GetFieldMask())

	// host.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_HOST:
		res, hardDelete, err = srv.IS.UpdateHost(
			ctx, in.ResourceId, in.GetResource().GetHost(), in.GetFieldMask(), in.GetTenantId())
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:
		res, err = srv.IS.UpdateHoststorage(ctx, in.ResourceId, in.GetResource().GetHoststorage(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:
		res, err = srv.IS.UpdateHostnic(ctx, in.ResourceId, in.GetResource().GetHostnic(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:
		res, err = srv.IS.UpdateHostusb(ctx, in.ResourceId, in.GetResource().GetHostusb(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:
		res, err = srv.IS.UpdateHostgpu(ctx, in.ResourceId, in.GetResource().GetHostgpu(), in.GetFieldMask())
	// instance.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:
		res, hardDelete, err = srv.IS.UpdateInstance(ctx, in.ResourceId, in.GetResource().GetInstance(), in.GetFieldMask())

	// provider.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:
		res, err = srv.IS.UpdateProvider(ctx, in.ResourceId, in.GetResource().GetProvider(), in.GetFieldMask())

	// os.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_OS:
		res, err = srv.IS.UpdateOs(ctx, in.ResourceId, in.GetResource().GetOs(), in.GetFieldMask())

	// schedule.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:
		res, err = srv.IS.UpdateSingleSchedule(ctx, in.ResourceId, in.GetResource().GetSingleschedule(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:
		res, err = srv.IS.UpdateRepeatedSchedule(ctx, in.ResourceId, in.GetResource().GetRepeatedschedule(), in.GetFieldMask())

	// telemetry.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:
		res, err = srv.IS.UpdateTelemetryGroup(ctx, in.ResourceId, in.GetResource().GetTelemetryGroup(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE:
		res, err = srv.IS.UpdateTelemetryProfile(ctx, in.ResourceId, in.GetResource().GetTelemetryProfile(), in.GetFieldMask())

	// workload.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:
		res, hardDelete, err = srv.IS.UpdateWorkload(ctx, in.ResourceId, in.GetResource().GetWorkload(), in.GetFieldMask())
	case inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:
		res, err = srv.IS.UpdateWorkloadMember(ctx, in.ResourceId, in.GetResource().GetWorkloadMember(), in.GetFieldMask())

	case inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF:
		res, hardDelete, err = srv.IS.UpdateRemoteAccessConfig(ctx, in.ResourceId,
			in.GetResource().GetRemoteAccess(), in.GetFieldMask())

	case inv_v1.ResourceKind_RESOURCE_KIND_TENANT:
		res, hardDelete, err = srv.IS.UpdateTenant(ctx, in.ResourceId, in.GetResource().GetTenant(), in.GetFieldMask())

	case inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:
		res, err = srv.IS.UpdateOSUpdatePolicy(ctx, in.ResourceId, in.GetResource().GetOsUpdatePolicy(), in.GetFieldMask())

	case inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:
		res, err = srv.IS.UpdateOSUpdateRun(ctx, in.ResourceId, in.GetResource().GetOsUpdateRun(), in.GetFieldMask())

	default:
		zlog.InfraSec().InfraError("unknown Resource Kind: %s", kind).Msg("update resource parse error")
		return nil, false, errors.Errorfc(codes.InvalidArgument, "unknown Resource Kind: %s", kind)
	}

	return res, hardDelete, err
}

func (srv *InventorygRPCServer) UpdateResource(
	ctx context.Context,
	in *inv_v1.UpdateResourceRequest,
) (*inv_v1.Resource, error) {
	zlog.Info().Msgf("UpdateResource for UUID %v", in.ClientUuid)

	err := error(nil)

	// authorize call first
	err = srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	// validate input
	err = validator.ValidateMessage(in)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	// fetch client_kind info from client registration map
	// and validate the information
	clientKind, err := srv.extractClientKind(in.ClientUuid)
	if err != nil {
		return nil, err
	}

	// policy evaluation
	err = srv.INVPOLICY.Verify(clientKind, in)
	// error handling for policy violation
	if err != nil {
		return nil, err
	}

	// Validate fieldmask against the message content
	res, err := util.UnwrapResource[proto.Message](in.GetResource())
	if err != nil {
		return nil, err
	}
	err = util.ValidateMaskAndFilterMessage(res, in.GetFieldMask(), true)
	if err != nil {
		return nil, err
	}

	kind, err := util.GetResourceKindFromResourceID(in.ResourceId)
	if err != nil {
		return nil, err
	}

	updatedRes, hardDelete, err := srv.doUpdateResource(ctx, kind, in)
	if err != nil {
		return nil, err
	}

	if hardDelete {
		srv.CR.StreamNotify(
			ctx,
			inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED,
			updatedRes,
			in.ClientUuid,
		)
	} else {
		srv.CR.StreamNotify(
			ctx,
			inv_v1.SubscribeEventsResponse_EVENT_KIND_UPDATED,
			updatedRes,
			in.ClientUuid,
		)
	}
	return updatedRes, nil
}

func (srv *InventorygRPCServer) doDeleteResource(
	ctx context.Context,
	kind inv_v1.ResourceKind,
	in *inv_v1.DeleteResourceRequest,
) (*inv_v1.Resource, bool, error) {
	var res *inv_v1.Resource
	var softDelete bool
	var err error

	// These methods have a common interface and this switch statement
	// can be replaced with a map[string]func(IS *store, param string)
	switch kind {
	// location.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_REGION:
		res, err = srv.IS.DeleteRegion(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_SITE:
		res, err = srv.IS.DeleteSite(ctx, in.ResourceId)

	// ou.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_OU:
		res, err = srv.IS.DeleteOu(ctx, in.ResourceId)

	// network.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:
		res, err = srv.IS.DeleteNetworkSegment(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:
		res, softDelete, err = srv.IS.DeleteNetlink(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:
		res, err = srv.IS.DeleteEndpoint(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:
		res, softDelete, err = srv.IS.DeleteIPAddress(ctx, in.ResourceId)

	// host.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_HOST:
		res, softDelete, err = srv.IS.DeleteHost(ctx, in.ResourceId, in.GetTenantId())
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:
		res, err = srv.IS.DeleteHoststorage(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:
		res, err = srv.IS.DeleteHostnic(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:
		res, err = srv.IS.DeleteHostusb(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:
		res, err = srv.IS.DeleteHostgpu(ctx, in.ResourceId)

	// instance.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:
		res, softDelete, err = srv.IS.DeleteInstance(ctx, in.ResourceId)

	// provider.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:
		res, err = srv.IS.DeleteProvider(ctx, in.ResourceId)

	// os.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_OS:
		res, err = srv.IS.DeleteOs(ctx, in.ResourceId)

	// schedule.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:
		res, err = srv.IS.DeleteSingleSchedule(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:
		res, err = srv.IS.DeleteRepeatedSchedule(ctx, in.ResourceId)

	// telemetry.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:
		res, err = srv.IS.DeleteTelemetryGroup(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE:
		res, err = srv.IS.DeleteTelemetryProfile(ctx, in.ResourceId)

	// workload.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:
		res, softDelete, err = srv.IS.DeleteWorkload(ctx, in.ResourceId)
	case inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:
		res, err = srv.IS.DeleteWorkloadMember(ctx, in.ResourceId)

	// remote access
	case inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF:
		res, err = srv.IS.SoftDeleteRemoteAccessConfig(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_TENANT:
		softDelete = true
		res, err = srv.IS.SoftDeleteTenant(ctx, in.ResourceId)

	// localaccount.proto
	case inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:
		res, err = srv.IS.DeleteLocalAccount(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:
		res, err = srv.IS.DeleteOSUpdatePolicy(ctx, in.ResourceId)

	case inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:
		res, err = srv.IS.DeleteOSUpdateRun(ctx, in.ResourceId)

	default:
		zlog.InfraSec().InfraError("unknown Resource Kind: %s", kind).Msg("delete resource parse error")
		return nil, softDelete, errors.Errorfc(codes.InvalidArgument, "unknown resource kind: %s", kind)
	}

	return res, softDelete, err
}

func (srv *InventorygRPCServer) DeleteResource(
	ctx context.Context,
	in *inv_v1.DeleteResourceRequest,
) (*inv_v1.DeleteResourceResponse, error) {
	// authorize call first
	err := srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	// fetch client_kind info from client registration map
	// and validate the information
	clientKind, err := srv.extractClientKind(in.ClientUuid)
	if err != nil {
		return nil, err
	}

	// policy evaluation
	err = srv.INVPOLICY.Verify(clientKind, in)
	// error handling for policy violation
	if err != nil {
		return nil, err
	}

	kind, err := util.GetResourceKindFromResourceID(in.ResourceId)
	if err != nil {
		return nil, err
	}

	deletedRes, softDelete, err := srv.doDeleteResource(ctx, kind, in)
	if err != nil {
		return nil, err
	}

	if softDelete {
		srv.CR.StreamNotify(
			ctx,
			inv_v1.SubscribeEventsResponse_EVENT_KIND_UPDATED,
			deletedRes,
			in.ClientUuid,
		)
	} else {
		srv.CR.StreamNotify(
			ctx,
			inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED,
			deletedRes,
			in.ClientUuid,
		)
	}
	return &inv_v1.DeleteResourceResponse{}, nil
}

func (srv *InventorygRPCServer) extractClientKind(uuid string) (string, error) {
	clientMapping := srv.CR.ClientRegistrationMap()
	clientInfo, ok := clientMapping.Load(uuid)
	if !ok {
		zlog.InfraSec().InfraError("failed to fetch client information to apply permissions: %s", uuid).Msg("")
		return "", errors.Errorfr(errors.Reason_UNKNOWN_CLIENT,
			"failed to fetch client information to apply permissions: %s", uuid,
		)
	}

	clientInfoTyped, ok := clientInfo.(clientreg.ClientInfo)
	if !ok {
		zlog.InfraSec().InfraError("unexpected type for ClientInfo").Msg("")
		return "", errors.Errorf("unexpected type for ClientInfo: %T", clientInfo)
	}
	return clientInfoTyped.ClientKind.String(), nil
}

func (srv *InventorygRPCServer) ListInheritedTelemetryProfiles(
	ctx context.Context,
	in *inv_v1.ListInheritedTelemetryProfilesRequest,
) (*inv_v1.ListInheritedTelemetryProfilesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("ListInheritedTelemetryProfiles: client_uuid=%v", in.ClientUuid)
	zlog.Debug().Msgf("ListInheritedTelemetryProfiles: request=%v", in)

	// authorize call first
	err := srv.Authorize(ctx, in)
	if err != nil {
		return nil, err
	}

	err = validator.ValidateMessage(in)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}

	telProfiles, totElems, err := srv.IS.ListInheritedTelemetryProfile(ctx, in)
	if err != nil {
		return nil, err
	}
	// Safely convert int to int32
	totElemsInt32, err := util.IntToInt32(totElems)
	if err != nil {
		return nil, err
	}
	return &inv_v1.ListInheritedTelemetryProfilesResponse{
		TelemetryProfiles: telProfiles,
		TotalElements:     totElemsInt32,
	}, nil
}

func (srv *InventorygRPCServer) GetTreeHierarchy(ctx context.Context, req *inv_v1.GetTreeHierarchyRequest) (
	*inv_v1.GetTreeHierarchyResponse,
	error,
) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("GetHierarchy: ")
	zlog.Debug().Msgf("GetHierarchy: request=%v", req)

	// authorize call first
	if err := srv.Authorize(ctx, req); err != nil {
		return nil, err
	}

	treeH, err := srv.IS.GetTreeHierarchy(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = validator.ValidateMessage(treeH); err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}
	return treeH, err
}

func (srv *InventorygRPCServer) GetSitesPerRegion(ctx context.Context, req *inv_v1.GetSitesPerRegionRequest) (
	*inv_v1.GetSitesPerRegionResponse,
	error,
) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("GetSitesPerRegion: ")
	zlog.Debug().Msgf("GetSitesPerRegion: request=%v", req)

	// authorize call first
	if err := srv.Authorize(ctx, req); err != nil {
		return nil, err
	}

	siterPerRegion, err := srv.IS.GetSitesPerRegion(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = validator.ValidateMessage(siterPerRegion); err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}
	return siterPerRegion, err
}

type deleteResourcesHandler func(ctx context.Context, tenantID string, enforce bool) (
	[]*util.Tuple[store.DeletionKind, *inv_v1.Resource], error,
)

type deleteResourcesHandlerProvider func(*store.InvStore) deleteResourcesHandler

//nolint:lll // oneliners give clean and compact look
var deleteResourcesHandlers = map[inv_v1.ResourceKind]deleteResourcesHandlerProvider{
	inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:          func(is *store.InvStore) deleteResourcesHandler { return is.DeleteEndpoints },
	inv_v1.ResourceKind_RESOURCE_KIND_HOST:              func(is *store.InvStore) deleteResourcesHandler { return is.DeleteHosts },
	inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:           func(is *store.InvStore) deleteResourcesHandler { return is.DeleteHostNICs },
	inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:           func(is *store.InvStore) deleteResourcesHandler { return is.DeleteHostGPUs },
	inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:       func(is *store.InvStore) deleteResourcesHandler { return is.DeleteHostStorages },
	inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:           func(is *store.InvStore) deleteResourcesHandler { return is.DeleteHostUSBs },
	inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:          func(is *store.InvStore) deleteResourcesHandler { return is.DeleteInstances },
	inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:         func(is *store.InvStore) deleteResourcesHandler { return is.DeleteIPAddresses },
	inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:           func(is *store.InvStore) deleteResourcesHandler { return is.DeleteNetLinks },
	inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:    func(is *store.InvStore) deleteResourcesHandler { return is.DeleteNetworkSegments },
	inv_v1.ResourceKind_RESOURCE_KIND_OS:                func(is *store.InvStore) deleteResourcesHandler { return is.DeleteOSes },
	inv_v1.ResourceKind_RESOURCE_KIND_OU:                func(is *store.InvStore) deleteResourcesHandler { return is.DeleteOus },
	inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:          func(is *store.InvStore) deleteResourcesHandler { return is.DeleteProviders },
	inv_v1.ResourceKind_RESOURCE_KIND_REGION:            func(is *store.InvStore) deleteResourcesHandler { return is.DeleteRegions },
	inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:  func(is *store.InvStore) deleteResourcesHandler { return is.DeleteRepeatedSchedules },
	inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:    func(is *store.InvStore) deleteResourcesHandler { return is.DeleteSingleSchedules },
	inv_v1.ResourceKind_RESOURCE_KIND_SITE:              func(is *store.InvStore) deleteResourcesHandler { return is.DeleteSites },
	inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:   func(is *store.InvStore) deleteResourcesHandler { return is.DeleteTelemetryGroups },
	inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE: func(is *store.InvStore) deleteResourcesHandler { return is.DeleteTelemetryProfiles },
	inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:          func(is *store.InvStore) deleteResourcesHandler { return is.DeleteWorkloads },
	inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:   func(is *store.InvStore) deleteResourcesHandler { return is.DeleteWorkloadMembers },
	inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:      func(is *store.InvStore) deleteResourcesHandler { return is.DeleteLocalAccounts },
	inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:    func(is *store.InvStore) deleteResourcesHandler { return is.DeleteOSUpdatePolicies },
	inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:       func(is *store.InvStore) deleteResourcesHandler { return is.DeleteOSUpdateRuns },
}

func (srv *InventorygRPCServer) DeleteAllResources(
	ctx context.Context, in *inv_v1.DeleteAllResourcesRequest,
) (*inv_v1.DeleteAllResourcesResponse, error) {
	zlog := zlog.TraceCtx(ctx)
	zlog.Info().Msgf("DeleteAllResources: request=%v", in)

	// authorize call first
	if aerr := srv.Authorize(ctx, in); aerr != nil {
		return nil, aerr
	}

	// fetch client_kind info from client registration map
	// and validate the information
	clientKind, err := srv.extractClientKind(in.ClientUuid)
	if err != nil {
		return nil, err
	}

	// policy evaluation
	if perr := srv.INVPOLICY.Verify(clientKind, in); perr != nil {
		return nil, perr
	}

	if verr := validator.ValidateMessage(in); verr != nil {
		zlog.InfraSec().InfraErr(verr).Send()
		return nil, errors.Wrap(verr)
	}

	handler, ok := deleteResourcesHandlers[in.GetResourceKind()]
	if !ok {
		zlog.InfraSec().InfraError("DeleteAllResources for Resource Kind: %s is not implemented", in.GetResourceKind())
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown resource kind: %s", in.GetResourceKind())
	}

	deletionInfo, err := handler(srv.IS)(ctx, in.TenantId, in.Enforce)
	if err != nil {
		return nil, err
	}

	for _, di := range deletionInfo {
		switch di.A {
		case store.SOFT:
			srv.CR.StreamNotify(ctx, inv_v1.SubscribeEventsResponse_EVENT_KIND_UPDATED, di.B, in.ClientUuid)
		case store.HARD:
			srv.CR.StreamNotify(ctx, inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED, di.B, in.ClientUuid)
		}
	}

	return new(inv_v1.DeleteAllResourcesResponse), nil
}
