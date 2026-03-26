// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	hostamtconfig "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostamtconfigresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var hostamtconfigResourceCreationValidators = []resourceValidator[*computev1.HostamtconfigResource]{
	protoValidator[*computev1.HostamtconfigResource],
	doNotAcceptResourceID[*computev1.HostamtconfigResource],
}

func (is *InvStore) CreateHostamtconfig(ctx context.Context, in *computev1.HostamtconfigResource) (*inv_v1.Resource, error) {
	if err := validate(in, hostamtconfigResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, hostamtconfigResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("HostAmtconfig Created: %s, %s", res.GetHostAmtconfig().GetResourceId(), res)
	return res, nil
}

func hostamtconfigResourceCreator(in *computev1.HostamtconfigResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_HOSTAMTCONFIG)
		zlog.Debug().Msgf("CreateHostamtconfig: %s", id)

		newEntity := tx.HostamtconfigResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, EmptyEnumStateMap, nil); err != nil {
			return nil, err
		}
		// Look up the optional host ID for this amtconfig.
		if err := setEdgeHostIDForMut(ctx, tx.Client(), mut, in.GetHost()); err != nil {
			return nil, err
		}

		// Set the resource_id field last.
		if err := mut.SetField(hostamtconfig.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getHostamtconfigQuery(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entHostAmtconfigResourceToProtoHostAmtconfigResource(res))
	}
}

func (is *InvStore) GetHostamtconfig(ctx context.Context, id string) (*inv_v1.Resource, error) {
	res, err := ExecuteInRoTxAndReturnSingle[ent.HostamtconfigResource](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.HostamtconfigResource, error) {
			return getHostamtconfigQuery(ctx, tx, id)
		})
	if err != nil {
		return nil, err
	}

	apiResource := entHostAmtconfigResourceToProtoHostAmtconfigResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_HostAmtconfig{HostAmtconfig: apiResource}}, nil
}

func getHostamtconfigQuery(ctx context.Context, tx *ent.Tx, resourceID string) (*ent.HostamtconfigResource, error) {
	entity, err := tx.HostamtconfigResource.Query().
		Where(hostamtconfig.ResourceID(resourceID)).
		WithHost().
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return entity, nil
}

func (is *InvStore) UpdateHostamtconfig(
	ctx context.Context, id string, in *computev1.HostamtconfigResource, fieldmask *fieldmaskpb.FieldMask,
) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("UpdateHostamtconfig (%s): %v, fm: %v", id, in, fieldmask)

	return ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx,
		func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
			entity, err := tx.HostamtconfigResource.Query().
				Select(hostamtconfig.FieldID).
				Where(hostamtconfig.ResourceID(id)).
				Only(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			updateBuilder := tx.HostamtconfigResource.UpdateOneID(entity.ID)
			mut := updateBuilder.Mutation()

			// Look up the (new) referenced edges for this amtconfig.
			err = setRelationsForHostamtconfigMutIfNeeded(ctx, tx.Client(), mut, in, fieldmask)
			if err != nil {
				return nil, err
			}

			err = buildEntMutate(in, mut, EmptyEnumStateMap, fieldmask.GetPaths())
			if err != nil {
				return nil, err
			}

			_, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			res, err := getHostamtconfigQuery(ctx, tx, id)
			if err != nil {
				return nil, err
			}
			toBeReturned, err := util.WrapResource(entHostAmtconfigResourceToProtoHostAmtconfigResource(res))

			return toBeReturned, errors.Wrap(err)
		},
	)
}

func (is *InvStore) DeleteHostamtconfig(ctx context.Context, id string) (*inv_v1.Resource, error) {
	// this is a "Hard Delete" as Hostamtconfigs don't have state
	zlog.Debug().Msgf("DeleteHostamtconfig Hard Delete: %s", id)

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, deleteHostAmtconfig(id))

	return res, err
}

func deleteHostAmtconfig(id string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		entity, err := tx.HostamtconfigResource.Query().
			Where(hostamtconfig.ResourceID(id)).
			Only(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = tx.HostamtconfigResource.DeleteOneID(entity.ID).Exec(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		return util.WrapResource(entHostAmtconfigResourceToProtoHostAmtconfigResource(entity))
	}
}

func (is *InvStore) DeleteHostAmtconfigs(
	ctx context.Context, tenantID string, _ bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]
	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			all, err := tx.HostamtconfigResource.Query().Where(hostamtconfig.TenantID(tenantID)).All(ctx)
			if err != nil {
				return err
			}
			if _, err := tx.HostamtconfigResource.Delete().Where(hostamtconfig.TenantID(tenantID)).Exec(ctx); err != nil {
				return err
			}
			for _, element := range all {
				res, err := util.WrapResource(entHostAmtconfigResourceToProtoHostAmtconfigResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(HARD, res))
			}
			return nil
		})
	return deleted, txErr
}

func filterHostamtconfigs(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) (
	[]*ent.HostamtconfigResource,
	int,
	error,
) {
	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_HOSTAMTCONFIG, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[hostamtconfig.OrderOption](filter.GetOrderBy(), hostamtconfig.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// perform query - And together all the predicates
	query := client.HostamtconfigResource.Query().
		WithHost().
		Where(pred).
		Order(orderOpts...).
		Offset(offset)

	// Limits number of query results if existent
	if limit != 0 {
		query = query.Limit(limit)
	}

	hostamtconfigList, err := query.All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	// Count total number of item without applying pagination limits, order, or loading edges.
	total, err := client.HostamtconfigResource.Query().
		Where(pred).
		Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return hostamtconfigList, total, nil
}

func (is *InvStore) ListHostamtconfig(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.HostamtconfigResource, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]*ent.HostamtconfigResource, *int, error) {
			filtered, total, err := filterHostamtconfigs(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[*ent.HostamtconfigResource, *inv_v1.GetResourceResponse](*resources,
		func(res *ent.HostamtconfigResource) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_HostAmtconfig{
						HostAmtconfig: entHostAmtconfigResourceToProtoHostAmtconfigResource(res),
					},
				},
			}
		})
	if err := collections.FirstError[*inv_v1.GetResourceResponse](resps, validateProto[*inv_v1.GetResourceResponse]); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, 0, errors.Wrap(err)
	}

	return resps, *total, nil
}

func (is *InvStore) FilterHostamtconfig(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*cl.ResourceTenantIDCarrier, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.HostamtconfigResource, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]*ent.HostamtconfigResource, *int, error) {
			filtered, total, err := filterHostamtconfigs(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}

	ids := collections.MapSlice[*ent.HostamtconfigResource, *cl.ResourceTenantIDCarrier](
		*resources, func(c *ent.HostamtconfigResource) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.TenantID, ResourceId: c.ResourceID}
		})

	return ids, *total, err
}

func setRelationsForHostamtconfigMutIfNeeded(
	ctx context.Context,
	client *ent.Client,
	mut *ent.HostamtconfigResourceMutation,
	in *computev1.HostamtconfigResource,
	fieldmask *fieldmaskpb.FieldMask,
) error {
	mut.ResetHost()
	if slices.Contains(fieldmask.GetPaths(), hostamtconfig.EdgeHost) {
		if err := setEdgeHostIDForMut(ctx, client, mut, in.GetHost()); err != nil {
			return err
		}
	}
	return nil
}
