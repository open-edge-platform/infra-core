// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	hostdevice "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostdeviceresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var hostdeviceResourceCreationValidators = []resourceValidator[*computev1.HostdeviceResource]{
	protoValidator[*computev1.HostdeviceResource],
	doNotAcceptResourceID[*computev1.HostdeviceResource],
}

func (is *InvStore) CreateHostdevice(ctx context.Context, in *computev1.HostdeviceResource) (*inv_v1.Resource, error) {
	if err := validate(in, hostdeviceResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, hostdeviceResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("HostDevice Created: %s, %s", res.GetHostDevice().GetResourceId(), res)
	return res, nil
}

func hostdeviceResourceCreator(in *computev1.HostdeviceResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE)
		zlog.Debug().Msgf("CreateHostdevice: %s", id)

		newEntity := tx.HostdeviceResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, EmptyEnumStateMap, nil); err != nil {
			return nil, err
		}
		// Look up the optional host ID for this device.
		if err := setEdgeHostIDForMut(ctx, tx.Client(), mut, in.GetHost()); err != nil {
			return nil, err
		}

		// Set the resource_id field last.
		if err := mut.SetField(hostdevice.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getHostdeviceQuery(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entHostDeviceResourceToProtoHostDeviceResource(res))
	}
}

func (is *InvStore) GetHostdevice(ctx context.Context, id string) (*inv_v1.Resource, error) {
	res, err := ExecuteInRoTxAndReturnSingle[ent.HostdeviceResource](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.HostdeviceResource, error) {
			return getHostdeviceQuery(ctx, tx, id)
		})
	if err != nil {
		return nil, err
	}

	apiResource := entHostDeviceResourceToProtoHostDeviceResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_HostDevice{HostDevice: apiResource}}, nil
}

func getHostdeviceQuery(ctx context.Context, tx *ent.Tx, resourceID string) (*ent.HostdeviceResource, error) {
	entity, err := tx.HostdeviceResource.Query().
		Where(hostdevice.ResourceID(resourceID)).
		WithHost().
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return entity, nil
}

func (is *InvStore) UpdateHostdevice(
	ctx context.Context, id string, in *computev1.HostdeviceResource, fieldmask *fieldmaskpb.FieldMask,
) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("UpdateHostdevice (%s): %v, fm: %v", id, in, fieldmask)

	return ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx,
		func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
			entity, err := tx.HostdeviceResource.Query().
				Select(hostdevice.FieldID).
				Where(hostdevice.ResourceID(id)).
				Only(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			updateBuilder := tx.HostdeviceResource.UpdateOneID(entity.ID)
			mut := updateBuilder.Mutation()

			// Look up the (new) referenced edges for this device.
			err = setRelationsForHostdeviceMutIfNeeded(ctx, tx.Client(), mut, in, fieldmask)
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

			res, err := getHostdeviceQuery(ctx, tx, id)
			if err != nil {
				return nil, err
			}
			toBeReturned, err := util.WrapResource(entHostDeviceResourceToProtoHostDeviceResource(res))

			return toBeReturned, errors.Wrap(err)
		},
	)
}

func (is *InvStore) DeleteHostdevice(ctx context.Context, id string) (*inv_v1.Resource, error) {
	// this is a "Hard Delete" as Hostdevices don't have state
	zlog.Debug().Msgf("DeleteHostdevice Hard Delete: %s", id)

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, deleteHostDevice(id))

	return res, err
}

func deleteHostDevice(id string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		entity, err := tx.HostdeviceResource.Query().
			Where(hostdevice.ResourceID(id)).
			Only(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = tx.HostdeviceResource.DeleteOneID(entity.ID).Exec(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		return util.WrapResource(entHostDeviceResourceToProtoHostDeviceResource(entity))
	}
}

func (is *InvStore) DeleteHostDevices(
	ctx context.Context, tenantID string, _ bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]
	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			all, err := tx.HostdeviceResource.Query().Where(hostdevice.TenantID(tenantID)).All(ctx)
			if err != nil {
				return err
			}
			if _, err := tx.HostdeviceResource.Delete().Where(hostdevice.TenantID(tenantID)).Exec(ctx); err != nil {
				return err
			}
			for _, element := range all {
				res, err := util.WrapResource(entHostDeviceResourceToProtoHostDeviceResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(HARD, res))
			}
			return nil
		})
	return deleted, txErr
}

func filterHostdevices(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) (
	[]*ent.HostdeviceResource,
	int,
	error,
) {
	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[hostdevice.OrderOption](filter.GetOrderBy(), hostdevice.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// perform query - And together all the predicates
	query := client.HostdeviceResource.Query().
		WithHost().
		Where(pred).
		Order(orderOpts...).
		Offset(offset)

	// Limits number of query results if existent
	if limit != 0 {
		query = query.Limit(limit)
	}

	hostdeviceList, err := query.All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	// Count total number of item without applying pagination limits, order, or loading edges.
	total, err := client.HostdeviceResource.Query().
		Where(pred).
		Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return hostdeviceList, total, nil
}

func (is *InvStore) ListHostdevice(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.HostdeviceResource, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]*ent.HostdeviceResource, *int, error) {
			filtered, total, err := filterHostdevices(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[*ent.HostdeviceResource, *inv_v1.GetResourceResponse](*resources,
		func(res *ent.HostdeviceResource) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_HostDevice{
						HostDevice: entHostDeviceResourceToProtoHostDeviceResource(res),
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

func (is *InvStore) FilterHostdevice(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*cl.ResourceTenantIDCarrier, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.HostdeviceResource, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]*ent.HostdeviceResource, *int, error) {
			filtered, total, err := filterHostdevices(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}

	ids := collections.MapSlice[*ent.HostdeviceResource, *cl.ResourceTenantIDCarrier](
		*resources, func(c *ent.HostdeviceResource) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.TenantID, ResourceId: c.ResourceID}
		})

	return ids, *total, err
}

func setRelationsForHostdeviceMutIfNeeded(
	ctx context.Context,
	client *ent.Client,
	mut *ent.HostdeviceResourceMutation,
	in *computev1.HostdeviceResource,
	fieldmask *fieldmaskpb.FieldMask,
) error {
	mut.ResetHost()
	if slices.Contains(fieldmask.GetPaths(), hostdevice.EdgeHost) {
		if err := setEdgeHostIDForMut(ctx, client, mut, in.GetHost()); err != nil {
			return err
		}
	}
	return nil
}
