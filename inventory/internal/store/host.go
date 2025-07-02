// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

// host.go  store information for Host objects

import (
	"context"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	hosts "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/utils/booleans"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var hostResourceCreationValidators = []resourceValidator[*computev1.HostResource]{
	protoValidator[*computev1.HostResource],
	doNotAcceptResourceID[*computev1.HostResource],
}

// enum state mapping.
func HostEnumStateMap(fname string, eint int32) (ent.Value, error) {
	stateMap := map[string]ent.Value{
		hosts.FieldDesiredState:                hosts.DesiredState(computev1.HostState_name[eint]),
		hosts.FieldCurrentState:                hosts.CurrentState(computev1.HostState_name[eint]),
		hosts.FieldBmcKind:                     hosts.BmcKind(computev1.BaremetalControllerKind_name[eint]),
		hosts.FieldDesiredPowerState:           hosts.DesiredPowerState(computev1.PowerState_name[eint]),
		hosts.FieldCurrentPowerState:           hosts.CurrentPowerState(computev1.PowerState_name[eint]),
		hosts.FieldHostStatusIndicator:         hosts.HostStatusIndicator(statusv1.StatusIndication_name[eint]),
		hosts.FieldRegistrationStatusIndicator: hosts.RegistrationStatusIndicator(statusv1.StatusIndication_name[eint]),
		hosts.FieldOnboardingStatusIndicator:   hosts.OnboardingStatusIndicator(statusv1.StatusIndication_name[eint]),
		hosts.FieldPowerCommandPolicy:          hosts.PowerCommandPolicy(computev1.PowerCommandPolicy_name[eint]),
		hosts.FieldDesiredAmtState:             hosts.DesiredAmtState(computev1.AmtState_name[eint]),
		hosts.FieldCurrentAmtState:             hosts.CurrentAmtState(computev1.AmtState_name[eint]),
		hosts.FieldPowerStatusIndicator:        hosts.PowerStatusIndicator(statusv1.StatusIndication_name[eint]),
		hosts.FieldAmtStatusIndicator:          hosts.AmtStatusIndicator(statusv1.StatusIndication_name[eint]),
	}

	if v, ok := stateMap[fname]; ok {
		return v, nil
	}

	zlog.InfraSec().InfraError("unknown Enum field %s", fname).Msg("")
	return nil, errors.Errorfc(codes.InvalidArgument, "unknown Enum field %s", fname)
}

func (is *InvStore) CreateHost(ctx context.Context, in *computev1.HostResource) (*inv_v1.Resource, error) {
	if err := validate(in, hostResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, hostResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Host Created: %s, %s", res.GetHost().GetResourceId(), res)
	return res, nil
}

func setHostStateFieldsOnCreate(in *computev1.HostResource, mut *ent.HostResourceMutation) {
	// Ensure the host state fields are never set to NULL in the DB.
	if in.GetCurrentState() == computev1.HostState_HOST_STATE_UNSPECIFIED {
		mut.SetCurrentState(hosts.CurrentStateHOST_STATE_UNSPECIFIED)
	}
	if in.GetDesiredState() == computev1.HostState_HOST_STATE_UNSPECIFIED {
		mut.SetDesiredState(hosts.DesiredStateHOST_STATE_UNSPECIFIED)
	}
	if in.GetCurrentPowerState() == computev1.PowerState_POWER_STATE_UNSPECIFIED {
		mut.SetCurrentPowerState(hosts.CurrentPowerStatePOWER_STATE_UNSPECIFIED)
	}
	if in.GetDesiredPowerState() == computev1.PowerState_POWER_STATE_UNSPECIFIED {
		mut.SetDesiredPowerState(hosts.DesiredPowerStatePOWER_STATE_UNSPECIFIED)
	}
	if in.GetPowerCommandPolicy() == computev1.PowerCommandPolicy_POWER_COMMAND_POLICY_UNSPECIFIED {
		mut.SetPowerCommandPolicy(hosts.PowerCommandPolicyPOWER_COMMAND_POLICY_UNSPECIFIED)
	}
	if in.GetDesiredAmtState() == computev1.AmtState_AMT_STATE_UNSPECIFIED {
		mut.SetDesiredAmtState(hosts.DesiredAmtStateAMT_STATE_UNSPECIFIED)
	}
	if in.GetCurrentAmtState() == computev1.AmtState_AMT_STATE_UNSPECIFIED {
		mut.SetCurrentAmtState(hosts.CurrentAmtStateAMT_STATE_UNSPECIFIED)
	}
}

func hostResourceCreator(in *computev1.HostResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_HOST)
		zlog.Debug().Msgf("CreateHost: %s", id)

		newEntity := tx.HostResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, HostEnumStateMap, nil); err != nil {
			return nil, err
		}

		// Look up the optional site ID for this host.
		if err := setEdgeSiteIDForMut(ctx, tx.Client(), mut, in.GetSite()); err != nil {
			return nil, err
		}
		// Look up the optional provider ID for this host.
		if err := setEdgeProviderIDForMut(ctx, tx.Client(), mut, in.GetProvider()); err != nil {
			return nil, err
		}

		// Set the host state fields to their default values if not set.
		setHostStateFieldsOnCreate(in, mut)

		// Set the resource_id field last.
		if err := mut.SetField(hosts.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, _, err := getHostQuery(ctx, tx, in.GetTenantId(), id, false, false)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entHostResourceToProtoHostResource(res))
	}
}

func (is *InvStore) GetHost(
	ctx context.Context, id, tenantID string,
) (*inv_v1.Resource, *inv_v1.GetResourceResponse_ResourceMetadata, error) {
	res, resMeta, err := ExecuteInRoTxAndReturnDouble[ent.HostResource, inv_v1.GetResourceResponse_ResourceMetadata](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.HostResource, *inv_v1.GetResourceResponse_ResourceMetadata, error) {
			return getHostQuery(ctx, tx, tenantID, id, true, true)
		})
	if err != nil {
		return nil, nil, err
	}

	apiResource := entHostResourceToProtoHostResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_Host{Host: apiResource}}, resMeta, nil
}

func getHostQuery(ctx context.Context, tx *ent.Tx, tenantID, resourceID string, loadMetadata, nestedLoad bool) (
	*ent.HostResource, *inv_v1.GetResourceResponse_ResourceMetadata, error,
) {
	query := tx.HostResource.Query().
		Where(hosts.ResourceID(resourceID)).
		WithSite().
		WithProvider().
		WithHostStorages().
		WithHostNics().
		WithHostUsbs().
		WithHostGpus()
	if nestedLoad {
		query.WithInstance(func(query *ent.InstanceResourceQuery) {
			query.WithDesiredOs().
				WithCurrentOs().
				WithOs()
		})
	} else {
		query.WithInstance()
	}
	entity, err := query.Only(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	if !loadMetadata {
		// Avoid loading inherited metadata
		return entity, nil, nil
	}

	// Build metadata hierarchy
	var phyMeta, logiMeta map[int]map[string]string
	phyMeta, logiMeta, err = getHostsInheritedMeta(ctx, tx.Client(), []int{entity.ID}, tenantID)
	if err != nil {
		return nil, nil, err
	}

	resMeta := BuildResourceMeta(phyMeta[entity.ID], logiMeta[entity.ID])

	return entity, resMeta, nil
}

func (is *InvStore) UpdateHost(
	ctx context.Context, id string, in *computev1.HostResource, fieldmask *fieldmaskpb.FieldMask, tenantID string,
) (*inv_v1.Resource, bool, error) {
	zlog.Debug().Msgf("UpdateHost (%s): %v, fm: %v", id, in, fieldmask)

	res, hardDelete, err := ExecuteInTxAndReturnDouble[inv_v1.Resource, bool](is)(
		ctx, hostResourceUpdater(id, in, fieldmask, tenantID))
	if err != nil {
		return nil, false, err
	}

	return res, *hardDelete, err
}

//nolint:cyclop // high cyclomatic complexity due to multiple fieldmask checks and state fields
func setHostStateFieldsOnUpdate(in *computev1.HostResource, mut *ent.HostResourceMutation,
	fieldmask *fieldmaskpb.FieldMask,
) {
	// Ensure the host state fields are never set to NULL in the DB.
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldCurrentState) &&
		in.GetCurrentState() == computev1.HostState_HOST_STATE_UNSPECIFIED {
		mut.ResetCurrentState()
		mut.SetCurrentState(hosts.CurrentStateHOST_STATE_UNSPECIFIED)
	}
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldDesiredState) &&
		in.GetDesiredState() == computev1.HostState_HOST_STATE_UNSPECIFIED {
		mut.ResetDesiredState()
		mut.SetDesiredState(hosts.DesiredStateHOST_STATE_UNSPECIFIED)
	}
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldCurrentPowerState) &&
		in.GetCurrentPowerState() == computev1.PowerState_POWER_STATE_UNSPECIFIED {
		mut.ResetCurrentPowerState()
		mut.SetCurrentPowerState(hosts.CurrentPowerStatePOWER_STATE_UNSPECIFIED)
	}
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldDesiredPowerState) &&
		in.GetDesiredPowerState() == computev1.PowerState_POWER_STATE_UNSPECIFIED {
		mut.ResetDesiredPowerState()
		mut.SetDesiredPowerState(hosts.DesiredPowerStatePOWER_STATE_UNSPECIFIED)
	}
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldPowerCommandPolicy) &&
		in.GetPowerCommandPolicy() == computev1.PowerCommandPolicy_POWER_COMMAND_POLICY_UNSPECIFIED {
		mut.ResetPowerCommandPolicy()
		mut.SetPowerCommandPolicy(hosts.PowerCommandPolicyPOWER_COMMAND_POLICY_UNSPECIFIED)
	}
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldCurrentAmtState) &&
		in.GetCurrentAmtState() == computev1.AmtState_AMT_STATE_UNSPECIFIED {
		mut.ResetCurrentAmtState()
		mut.SetCurrentAmtState(hosts.CurrentAmtStateAMT_STATE_UNSPECIFIED)
	}
	if slices.Contains(fieldmask.GetPaths(), hosts.FieldDesiredAmtState) &&
		in.GetDesiredAmtState() == computev1.AmtState_AMT_STATE_UNSPECIFIED {
		mut.ResetDesiredAmtState()
		mut.SetDesiredAmtState(hosts.DesiredAmtStateAMT_STATE_UNSPECIFIED)
	}
}

func hostResourceUpdater(
	id string, in *computev1.HostResource, fieldmask *fieldmaskpb.FieldMask, tenantID string,
) func(context.Context, *ent.Tx) (*inv_v1.Resource, *bool, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, *bool, error) {
		entity, err := tx.HostResource.Query().
			Where(hosts.ResourceID(id)).
			WithInstance().
			Only(ctx)
		if err != nil {
			return nil, booleans.Pointer(false), errors.Wrap(err)
		}

		// hard delete - if both Desired and Current state are Deleted, remove
		var updatedHost *ent.HostResource
		hardDelete := false
		if isHostHardDelete(fieldmask, entity, in) {
			zlog.Debug().Msgf("UpdateHost Hard Delete: %s", in.GetResourceId())
			err = hardDeleteHost(ctx, tx.Client(), entity)
			if err != nil {
				return nil, booleans.Pointer(false), err
			}
			hardDelete = true
			// Set current state to be consistent on the returned value on events and upon update.
			entity.CurrentState = hosts.CurrentStateHOST_STATE_DELETED
			updatedHost = entity
		} else {
			if isInValidHostTransition(fieldmask, entity, in) {
				zlog.InfraSec().InfraError("%s from %s to %s is not allowed",
					id, entity.CurrentState, in.DesiredState).Msgf("UpdateHost")
				return nil, booleans.Pointer(false),
					errors.Errorfc(codes.InvalidArgument,
						"UpdateHost %s from %s to %s is not allowed", id, entity.CurrentState, in.DesiredState)
			}

			updateBuilder := tx.Client().HostResource.UpdateOneID(entity.ID)
			mut := updateBuilder.Mutation()

			// Look up the (new) referenced edges for this host.
			err = setRelationsForHostMutIfNeeded(ctx, tx.Client(), mut, in, fieldmask)
			if err != nil {
				return nil, booleans.Pointer(false), err
			}

			err = buildEntMutate(in, mut, HostEnumStateMap, fieldmask.GetPaths())
			if err != nil {
				return nil, booleans.Pointer(false), err
			}

			// Set the host state fields to their default values if not set.
			setHostStateFieldsOnUpdate(in, mut, fieldmask)

			// save the UpdateOne
			_, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, booleans.Pointer(false), errors.Wrap(err)
			}

			updatedHost, _, err = getHostQuery(ctx, tx, tenantID, id, false, false)
			if err != nil {
				return nil, booleans.Pointer(false), err
			}
		}

		toBeReturned, err := util.WrapResource(entHostResourceToProtoHostResource(updatedHost))

		return toBeReturned, &hardDelete, err
	}
}

func hardDeleteHost(ctx context.Context, client *ent.Client, host *ent.HostResource) error {
	if host.Edges.Instance != nil {
		zlog.InfraSec().InfraError("the host has a relation with Instance and cannot be deleted").Msg("")
		return errors.Errorfc(codes.FailedPrecondition, "the host has a relation with Instance and cannot be deleted")
	}

	// should be nil on success
	return errors.Wrap(client.HostResource.DeleteOneID(host.ID).Exec(ctx))
}

func (is *InvStore) DeleteHost(ctx context.Context, id, tenantID string) (*inv_v1.Resource, bool, error) {
	// this is a "Soft Delete" - it only sets the Desired State to Deleted
	// Hard delete happens in Update, when both Desired and Current state are
	// both Deleted.
	zlog.Debug().Msgf("DeleteHost Soft Delete: %s", id)

	res, isSoftDelete, err := ExecuteInTxAndReturnDouble[inv_v1.Resource, bool](is)(ctx, deleteHost(tenantID, id))
	if err != nil {
		return nil, false, err
	}

	return res, *isSoftDelete, err
}

func deleteHost(tenantID, id string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, *bool, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, *bool, error) {
		entity, err := tx.HostResource.Query().
			Where(hosts.ResourceID(id)).
			Only(ctx)
		if err != nil {
			return nil, booleans.Pointer(false), errors.Wrap(err)
		}

		_, err = tx.HostResource.UpdateOneID(entity.ID).
			SetDesiredState(hosts.DesiredStateHOST_STATE_DELETED).
			Save(ctx)
		if err != nil {
			return nil, booleans.Pointer(false), errors.Wrap(err)
		}

		res, _, err := getHostQuery(ctx, tx, tenantID, id, false, false)
		if err != nil {
			return nil, booleans.Pointer(false), err
		}
		toBeReturned, err := util.WrapResource(entHostResourceToProtoHostResource(res))
		if err != nil {
			return nil, booleans.Pointer(false), err
		}

		return toBeReturned, booleans.Pointer(true), nil
	}
}

func (is *InvStore) DeleteHosts(
	ctx context.Context, tenantID string, enforce bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	deletionStrategies := map[bool]func(ctx context.Context, tx *ent.Tx, tenantID string) (DeletionKind, int, error){
		true: func(ctx context.Context, tx *ent.Tx, tenantID string) (DeletionKind, int, error) {
			i, e := tx.HostResource.Delete().Where(hosts.TenantID(tenantID)).Exec(ctx)
			return HARD, i, e
		},
		false: func(ctx context.Context, tx *ent.Tx, tenantID string) (DeletionKind, int, error) {
			i, e := tx.HostResource.Update().
				Where(hosts.TenantID(tenantID)).
				SetDesiredState(hosts.DesiredStateHOST_STATE_DELETED).
				Save(ctx)
			return SOFT, i, e
		},
	}
	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]

	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			collection, err := tx.HostResource.Query().Where(hosts.TenantID(tenantID)).All(ctx)
			if err != nil {
				return err
			}
			dk, noOfDeleted, err := deletionStrategies[enforce](ctx, tx, tenantID)
			if err != nil {
				return err
			}
			if noOfDeleted != len(collection) {
				return errors.Errorf(
					"Returned number of updated/delete hosts(%d) is different that number of retrieved hosts(%d)",
					noOfDeleted,
					len(collection))
			}
			if dk == SOFT {
				// because of performance reasons we do not want to fetch updated instance from DB,
				// and in the same time we want to have updated resource reported by the event.
				collections.ForEach(collection, func(i *ent.HostResource) {
					i.DesiredState = hosts.DesiredStateHOST_STATE_DELETED
				})
			}
			for _, element := range collection {
				res, err := util.WrapResource(entHostResourceToProtoHostResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(dk, res))
			}

			return nil
		})
	return deleted, txErr
}

//nolint:cyclop // high cyclomatic complexity due to filtering by metadata in go code
func filterHosts(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) ([]hostWithInheritedMeta, int, error) {
	host := filter.GetResource().GetHost()

	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_HOST, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[hosts.OrderOption](filter.GetOrderBy(), hosts.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// FIXME: ITEP-774 better define the behavior of host.Metadata with len == 0
	filterMeta, err := ParseMetadata(host.Metadata)
	if err != nil {
		return nil, 0, err
	}
	isMetadataSet := len(filterMeta) != 0

	var total int

	// perform query - And together all the predicates
	query := client.HostResource.Query().
		WithSite().
		WithProvider().
		WithHostStorages().
		WithHostNics().
		WithHostUsbs().
		WithHostGpus().
		WithInstance(func(query *ent.InstanceResourceQuery) {
			query.WithDesiredOs().
				WithCurrentOs().
				WithOs()
		}).
		Where(pred).
		Order(orderOpts...)
	// since metadata filter is applied explicitly in go, rather than via ent query, we need to query all resources,
	// filter them and apply the offset and limit later.
	if !isMetadataSet {
		// Count total number of item without applying pagination limits, order, or loading edges.
		total, err = client.HostResource.Query().
			Where(pred).
			Count(ctx)
		if err != nil {
			return nil, 0, err
		}

		// Limits number of query results if existent
		if limit != 0 {
			query = query.Limit(limit)
		}
		query = query.Offset(offset)
	}

	hostList, err := query.All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	hostIDMap := getHostIDToHostMap(hostList)
	phyMeta, logiMeta, err := getHostsInheritedMeta(ctx, client, maps.Keys(hostIDMap))
	if err != nil {
		return nil, 0, err
	}

	if isMetadataSet {
		hostList = filterHostsByMetadata(hostIDMap, phyMeta, logiMeta, filterMeta)
		total = len(hostList)
		// Apply offset and limit
		switch {
		case limit != 0 && offset+limit <= len(hostList):
			hostList = hostList[offset : offset+limit]
		case offset <= len(hostList):
			hostList = hostList[offset:]
		default:
			hostList = make([]*ent.HostResource, 0)
		}
	}

	hostWithMetaList := collections.MapSlice[*ent.HostResource, hostWithInheritedMeta](
		hostList,
		func(resource *ent.HostResource) hostWithInheritedMeta {
			return createHostWithInheritedMeta(resource, phyMeta, logiMeta)
		})
	return hostWithMetaList, total, nil
}

func createHostWithInheritedMeta(
	hostResource *ent.HostResource, physicalMeta, logicalMeta map[int]map[string]string,
) hostWithInheritedMeta {
	res := hostWithInheritedMeta{resource: hostResource}
	pm, ok := physicalMeta[hostResource.ID]
	if ok {
		res.meta.physical = pm
	}
	lm, ok := logicalMeta[hostResource.ID]
	if ok {
		res.meta.logical = lm
	}
	return res
}

func (is *InvStore) ListHosts(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]hostWithInheritedMeta, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]hostWithInheritedMeta, *int, error) {
			resources, total, err := filterHosts(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &resources, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[hostWithInheritedMeta, *inv_v1.GetResourceResponse](*resources,
		func(res hostWithInheritedMeta) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: entHostResourceToProtoHostResource(res.resource),
					},
				},
				RenderedMetadata: BuildResourceMeta(res.meta.physical, res.meta.logical),
			}
		})
	if err := collections.FirstError[*inv_v1.GetResourceResponse](resps, validateProto[*inv_v1.GetResourceResponse]); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, 0, errors.Wrap(err)
	}

	return resps, *total, nil
}

func (is *InvStore) FilterHosts(ctx context.Context, filter *inv_v1.ResourceFilter) ([]*cl.ResourceTenantIDCarrier, int, error) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]hostWithInheritedMeta, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]hostWithInheritedMeta, *int, error) {
			filteredOus, total, err := filterHosts(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filteredOus, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}
	ids := collections.MapSlice[hostWithInheritedMeta, *cl.ResourceTenantIDCarrier](
		*resources, func(c hostWithInheritedMeta) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.resource.TenantID, ResourceId: c.resource.ResourceID}
		})

	return ids, *total, err
}

func setRelationsForHostMutIfNeeded(
	ctx context.Context,
	client *ent.Client,
	mut *ent.HostResourceMutation,
	in *computev1.HostResource,
	fieldmask *fieldmaskpb.FieldMask,
) error {
	mut.ResetSite()
	if slices.Contains(fieldmask.GetPaths(), hosts.EdgeSite) {
		if err := setEdgeSiteIDForMut(ctx, client, mut, in.GetSite()); err != nil {
			return err
		}
	}
	mut.ResetProvider()
	if slices.Contains(fieldmask.GetPaths(), hosts.EdgeProvider) {
		if err := setEdgeProviderIDForMut(ctx, client, mut, in.GetProvider()); err != nil {
			return err
		}
	}
	return nil
}

func getHostIDFromResourceID(ctx context.Context, client *ent.Client, hostRes *computev1.HostResource) (int, error) {
	host, qerr := client.HostResource.Query().
		Where(hosts.ResourceID(hostRes.ResourceId)).
		Only(ctx)
	if qerr != nil {
		return 0, errors.Wrap(qerr)
	}
	return host.ID, nil
}

func isHostHardDelete(fieldmask *fieldmaskpb.FieldMask, hostq *ent.HostResource, in *computev1.HostResource) bool {
	return slices.Contains(fieldmask.GetPaths(), hosts.FieldCurrentState) &&
		hostq.DesiredState == hosts.DesiredStateHOST_STATE_DELETED &&
		in.CurrentState == computev1.HostState_HOST_STATE_DELETED
}

func isInValidHostTransition(fieldmask *fieldmaskpb.FieldMask, hostq *ent.HostResource, in *computev1.HostResource) bool {
	// This transition table maps the possible desired state transitions that can happen
	// from a current state. Note that all current states can transition to Deleted state.
	// This must conform to the state machine captured in the documentation.
	currentToDesiredStateTransitionTable := map[hosts.CurrentState][]computev1.HostState{
		hosts.CurrentStateHOST_STATE_UNTRUSTED: {
			computev1.HostState_HOST_STATE_DELETED,
			computev1.HostState_HOST_STATE_UNTRUSTED,
		},
		hosts.CurrentStateHOST_STATE_UNSPECIFIED: {
			computev1.HostState_HOST_STATE_DELETED,
			computev1.HostState_HOST_STATE_UNSPECIFIED,
			computev1.HostState_HOST_STATE_REGISTERED,
			computev1.HostState_HOST_STATE_ONBOARDED,
			computev1.HostState_HOST_STATE_UNTRUSTED,
		},
		hosts.CurrentStateHOST_STATE_REGISTERED: {
			computev1.HostState_HOST_STATE_DELETED,
			computev1.HostState_HOST_STATE_REGISTERED,
			computev1.HostState_HOST_STATE_ONBOARDED,
		},
		hosts.CurrentStateHOST_STATE_ONBOARDED: {
			computev1.HostState_HOST_STATE_DELETED,
			computev1.HostState_HOST_STATE_ONBOARDED,
			computev1.HostState_HOST_STATE_UNTRUSTED,
		},
		hosts.CurrentStateHOST_STATE_DELETED: {computev1.HostState_HOST_STATE_DELETED},
	}
	return slices.Contains(fieldmask.GetPaths(), hosts.FieldDesiredState) &&
		!slices.Contains(currentToDesiredStateTransitionTable[hostq.CurrentState], in.DesiredState)
}
