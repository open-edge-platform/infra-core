// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

// osupdaterun.go  store information for OSUpdateRun objects

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	our "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdaterunresource"
	compute_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	status_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var osUpRunResourceCreationValidators = []resourceValidator[*compute_v1.OSUpdateRunResource]{
	protoValidator[*compute_v1.OSUpdateRunResource],
	doNotAcceptResourceID[*compute_v1.OSUpdateRunResource],
}

// OSUpdateRunEnumStateMap maps proto enum fields to their Ent equivalents.
func OSUpdateRunEnumStateMap(fname string, eint int32) (ent.Value, error) {
	switch fname {
	case our.FieldStatusIndicator:
		return our.StatusIndicator(status_v1.StatusIndication_name[eint]), nil
	default:
		zlog.InfraSec().InfraError("unknown Enum field %s", fname).Msg("")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Enum field %s", fname)
	}
}

func (is *InvStore) CreateOSUpdateRun(ctx context.Context, in *compute_v1.OSUpdateRunResource) (*inv_v1.Resource, error) {
	if err := validate(in, osUpRunResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, osUpdateRunResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("OS Update Run Created: %s, %s", res.GetOsUpdateRun().GetResourceId(), res)
	return res, nil
}

func osUpdateRunResourceCreator(in *compute_v1.OSUpdateRunResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN)
		zlog.Debug().Msgf("CreateOs: %s", id)

		newEntity := tx.OSUpdateRunResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, OSUpdateRunEnumStateMap, nil); err != nil {
			return nil, err
		}

		// Look up the edges
		if err := setEdgeAppliedPolicyIDForMut(ctx, tx.Client(), mut, in.GetAppliedPolicy()); err != nil {
			return nil, err
		}
		if err := setEdgeInstanceIDForMut(ctx, tx.Client(), mut, in.GetInstance()); err != nil {
			return nil, err
		}

		if err := mut.SetField(our.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getOSUpdateRunQuery(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entOSUpdateRunResourceToProtoOSUpdateRunResource(res))
	}
}

func (is *InvStore) GetOSUpdateRun(ctx context.Context, id string) (*inv_v1.Resource, error) {
	res, err := ExecuteInRoTxAndReturnSingle[ent.OSUpdateRunResource](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.OSUpdateRunResource, error) {
			return getOSUpdateRunQuery(ctx, tx, id)
		})
	if err != nil {
		return nil, err
	}

	apiResource := entOSUpdateRunResourceToProtoOSUpdateRunResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_OsUpdateRun{OsUpdateRun: apiResource}}, nil
}

func getOSUpdateRunQuery(ctx context.Context, tx *ent.Tx, resourceID string) (*ent.OSUpdateRunResource, error) {
	entity, err := tx.OSUpdateRunResource.Query().
		Where(our.ResourceID(resourceID)).
		WithAppliedPolicy().
		WithInstance().
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return entity, nil
}

func (is *InvStore) UpdateOSUpdateRun(
	ctx context.Context,
	id string,
	in *compute_v1.OSUpdateRunResource,
	fieldmask *fieldmaskpb.FieldMask,
) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("UpdateosUpdateRun (%s): %v, fm: %v", id, in, fieldmask)
	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx,
		func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
			entity, err := tx.OSUpdateRunResource.Query().
				// we need to also retrieve immutable edges to do the check
				Select(our.FieldID).
				Where(our.ResourceID(id)).
				WithAppliedPolicy().
				WithInstance().
				Only(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			// Applied Policy is immutable.
			if in.GetAppliedPolicy() != nil && entity.Edges.AppliedPolicy.ResourceID != in.GetAppliedPolicy().GetResourceId() {
				return nil, errors.Errorfc(codes.InvalidArgument, "Cannot change Applied Policy for an existing OS Update Run")
			}

			// Instance is immutable.
			if in.GetInstance() != nil && entity.Edges.Instance.ResourceID != in.GetInstance().GetResourceId() {
				return nil, errors.Errorfc(codes.InvalidArgument, "Cannot change Instance for an existing OS Update Run")
			}

			updateBuilder := tx.OSUpdateRunResource.UpdateOneID(entity.ID)
			mut := updateBuilder.Mutation()

			// No need to update the edge, it's immutable!

			err = buildEntMutate(in, mut, OSUpdateRunEnumStateMap, fieldmask.GetPaths())
			if err != nil {
				return nil, err
			}

			_, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			res, err := getOSUpdateRunQuery(ctx, tx, id)
			if err != nil {
				return nil, err
			}

			protoRes := entOSUpdateRunResourceToProtoOSUpdateRunResource(res)
			return util.WrapResource(protoRes)
		})
	if err != nil {
		return nil, err
	}

	return res, err
}

func (is *InvStore) DeleteOSUpdateRun(ctx context.Context, id string) (*inv_v1.Resource, error) {
	// this is a "Hard Delete" as OSUpdateRun don't have state to reconcile
	zlog.Debug().Msgf("DeleteOSUpdateRun Hard Delete: %s", id)

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, deleteOSUpdateRun(id))

	return res, err
}

func deleteOSUpdateRun(resourceID string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		entity, err := tx.OSUpdateRunResource.Query().
			Where(our.ResourceID(resourceID)).
			Only(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = tx.OSUpdateRunResource.DeleteOneID(entity.ID).Exec(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		return util.WrapResource(entOSUpdateRunResourceToProtoOSUpdateRunResource(entity))
	}
}

func (is *InvStore) DeleteOSUpdateRuns(
	ctx context.Context, tenantID string, _ bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]
	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			collection, err := tx.OSUpdateRunResource.Query().Where(our.TenantID(tenantID)).All(ctx)
			if err != nil {
				return err
			}
			if _, err := tx.OSUpdateRunResource.Delete().Where(our.TenantID(tenantID)).Exec(ctx); err != nil {
				return err
			}
			for _, element := range collection {
				res, err := util.WrapResource(entOSUpdateRunResourceToProtoOSUpdateRunResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(HARD, res))
			}
			return nil
		})
	return deleted, txErr
}

func filterOSUpdateRuns(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) (
	[]*ent.OSUpdateRunResource, int, error,
) {
	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[our.OrderOption](filter.GetOrderBy(), our.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// perform query - And together all the predicates with eager loading
	query := client.OSUpdateRunResource.Query().
		WithAppliedPolicy().
		WithInstance().
		Where(pred).
		Order(orderOpts...).
		Offset(offset)

	// Limits number of query results if existent
	if limit != 0 {
		query = query.Limit(limit)
	}

	osList, err := query.All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	// Count total number of item without applying pagination limits, order, or loading edges.
	total, err := client.OSUpdateRunResource.Query().
		Where(pred).
		Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return osList, total, nil
}

func (is *InvStore) ListOSUpdateRuns(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.OSUpdateRunResource, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]*ent.OSUpdateRunResource, *int, error) {
			resources, total, err := filterOSUpdateRuns(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &resources, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[*ent.OSUpdateRunResource, *inv_v1.GetResourceResponse](*resources,
		func(res *ent.OSUpdateRunResource) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_OsUpdateRun{
						OsUpdateRun: entOSUpdateRunResourceToProtoOSUpdateRunResource(res),
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

func (is *InvStore) FilterOSUpdateRuns(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*cl.ResourceTenantIDCarrier, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.OSUpdateRunResource, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]*ent.OSUpdateRunResource, *int, error) {
			filtered, total, err := filterOSUpdateRuns(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}

	ids := collections.MapSlice[*ent.OSUpdateRunResource, *cl.ResourceTenantIDCarrier](
		*resources, func(c *ent.OSUpdateRunResource) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.TenantID, ResourceId: c.ResourceID}
		})

	return ids, *total, err
}
