// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	customconfigs "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/customconfigresource"
	instances "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
	"google.golang.org/grpc/codes"
)

var customConfigResourceCreationValidators = []resourceValidator[*computev1.CustomConfigResource]{
	protoValidator[*computev1.CustomConfigResource],
	doNotAcceptResourceID[*computev1.CustomConfigResource],
}

func (is *InvStore) CreateCustomConfig(ctx context.Context, in *computev1.CustomConfigResource) (*inv_v1.Resource, error) {
	if err := validate(in, customConfigResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, customConfigResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("CustomConfig Created: %s, %s", res.GetCustomConfig().GetResourceId(), res)
	return res, nil
}

func customConfigResourceCreator(in *computev1.CustomConfigResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG)
		zlog.Debug().Msgf("CustomConfig: %s", id)

		newEntity := tx.CustomConfigResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, EmptyEnumStateMap, nil); err != nil {
			return nil, err
		}

		if err := mut.SetField(customconfigs.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getCustomConfigQuery(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entCustomConfigResourceToProtoCustomConfigResource(res))
	}
}

func getCustomConfigQuery(ctx context.Context, tx *ent.Tx, resourceID string) (*ent.CustomConfigResource, error) {
	entity, err := tx.CustomConfigResource.Query().
		Where(customconfigs.ResourceID(resourceID)).
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return entity, nil
}

func (is *InvStore) GetCustomConfig(ctx context.Context, id string) (*inv_v1.Resource, error) {
	res, err := ExecuteInRoTxAndReturnSingle[ent.CustomConfigResource](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.CustomConfigResource, error) {
			return getCustomConfigQuery(ctx, tx, id)
		})
	if err != nil {
		return nil, err
	}

	apiResource := entCustomConfigResourceToProtoCustomConfigResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_CustomConfig{CustomConfig: apiResource}}, nil
}

func (is *InvStore) DeleteCustomConfig(ctx context.Context, id string) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("DeleteCustomConfig Delete: %s", id)

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, deleteCustomConfig(id))

	return res, err
}

func deleteCustomConfig(id string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		entity, qerr := tx.CustomConfigResource.Query().
			Where(customconfigs.ResourceID(id)).
			Only(ctx)
		if qerr != nil {
			return nil, errors.Wrap(qerr)
		}

		// Error is already wrapped
		if err := verifyCustomConfigStrongRelations(ctx, tx, id); err != nil {
			return nil, err
		}

		if err := tx.CustomConfigResource.DeleteOneID(entity.ID).Exec(ctx); err != nil {
			return nil, errors.Wrap(err)
		}

		return util.WrapResource(entCustomConfigResourceToProtoCustomConfigResource(entity))
	}
}

func verifyCustomConfigStrongRelations(ctx context.Context, tx *ent.Tx, id string) error {
	_, err := tx.InstanceResource.Query().
		Where(instances.HasCustomConfigWith(customconfigs.ResourceID(id))).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return errors.Wrap(err)
	}

	if err == nil {
		zlog.InfraSec().InfraError("the CustomConfig has a relation with instance and cannot be deleted").Msg("")
		return errors.Errorfc(codes.FailedPrecondition,
			"the CustomConfig has a relation with instance and cannot be deleted")
	}

	return nil
}

func (is *InvStore) DeleteCustomConfigs(
	ctx context.Context, tenantID string, _ bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]
	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			collection, err := tx.CustomConfigResource.Query().Where(customconfigs.TenantID(tenantID)).All(ctx)
			if err != nil {
				return err
			}
			if _, err := tx.CustomConfigResource.Delete().Where(customconfigs.TenantID(tenantID)).Exec(ctx); err != nil {
				return err
			}
			for _, element := range collection {
				res, err := util.WrapResource(entCustomConfigResourceToProtoCustomConfigResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(HARD, res))
			}
			return nil
		})
	return deleted, txErr
}

func (is *InvStore) ListCustomConfigs(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.CustomConfigResource, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]*ent.CustomConfigResource, *int, error) {
			filtered, total, err := filterCustomConfigs(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[*ent.CustomConfigResource, *inv_v1.GetResourceResponse](*resources,
		func(res *ent.CustomConfigResource) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_CustomConfig{
						CustomConfig: entCustomConfigResourceToProtoCustomConfigResource(res),
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

func filterCustomConfigs(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) (
	[]*ent.CustomConfigResource, int, error,
) {
	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[customconfigs.OrderOption](filter.GetOrderBy(), customconfigs.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// perform query - And together all the predicates
	query := client.CustomConfigResource.Query().
		Where(pred).
		Order(orderOpts...).
		Offset(offset)

	// Limits number of query results if existent
	if limit != 0 {
		query = query.Limit(limit)
	}

	provs, err := query.All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	// Count total number of item without applying pagination limits, order, or loading edges.
	total, err := client.CustomConfigResource.Query().
		Where(pred).
		Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return provs, total, nil
}

func (is *InvStore) FilterCustomConfigs(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*cl.ResourceTenantIDCarrier, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.CustomConfigResource, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]*ent.CustomConfigResource, *int, error) {
			filtered, total, err := filterCustomConfigs(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}

	ids := collections.MapSlice[*ent.CustomConfigResource, *cl.ResourceTenantIDCarrier](
		*resources, func(c *ent.CustomConfigResource) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.TenantID, ResourceId: c.ResourceID}
		})

	return ids, *total, err
}

func getCustomConfigIDFromResourceID(
	ctx context.Context, client *ent.Client, customConfigRes *computev1.CustomConfigResource,
) (int, error) {
	customConfig, qerr := client.CustomConfigResource.Query().
		Where(customconfigs.ResourceID(customConfigRes.ResourceId)).
		Only(ctx)
	if qerr != nil {
		return 0, errors.Wrap(qerr)
	}
	return customConfig.ID, nil
}
