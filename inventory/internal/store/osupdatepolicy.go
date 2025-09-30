// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

// osupdatepolicy.go  store information for OS Update Policy objects

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	oup "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdatepolicyresource"
	compute_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var osUpPolicyResourceCreationValidators = []resourceValidator[*compute_v1.OSUpdatePolicyResource]{
	protoValidator[*compute_v1.OSUpdatePolicyResource],
	validateOSUpdatePolicyProto,
	doNotAcceptResourceID[*compute_v1.OSUpdatePolicyResource],
}

func validateOSUpdatePolicyProto(in *compute_v1.OSUpdatePolicyResource) error {
	switch in.GetUpdatePolicy() {
	case compute_v1.UpdatePolicy_UPDATE_POLICY_UNSPECIFIED:
		return errors.Errorfc(codes.InvalidArgument, "OS Update Policy type cannot be unspecified")
	case compute_v1.UpdatePolicy_UPDATE_POLICY_TARGET:
		if !isValidTargetPolicy(in) {
			return errors.Errorfc(codes.InvalidArgument, "Fields for mutable and immutable OSes are mutually exclusive")
		}
	case compute_v1.UpdatePolicy_UPDATE_POLICY_LATEST:
		if !isValidLatestPolicy(in) {
			return errors.Errorfc(codes.InvalidArgument, "With Policy LATEST, no fields should be set")
		}
	default:
	}
	return nil
}

func isValidTargetPolicy(in *compute_v1.OSUpdatePolicyResource) bool {
	// Enforce mutually exclusive fields: either TargetOs OR the other fields, but not both or neither
	targetOsSet := in.GetTargetOs() != nil
	mutableOSFieldsSet := in.GetInstallPackages() != "" || in.GetUpdateSources() != nil || in.GetKernelCommand() != ""

	// Valid if exactly one group is set and at least one group is set
	return targetOsSet != mutableOSFieldsSet && (targetOsSet || mutableOSFieldsSet)
}

func isValidLatestPolicy(in *compute_v1.OSUpdatePolicyResource) bool {
	// All fields must be unset
	return in.GetTargetOs() == nil && in.GetInstallPackages() == "" &&
		in.GetUpdateSources() == nil && in.GetKernelCommand() == ""
}

// OSUpdatePolicyEnumStateMap maps proto enum fields to their Ent equivalents.
func OSUpdatePolicyEnumStateMap(fname string, eint int32) (ent.Value, error) {
	switch fname {
	case oup.FieldUpdatePolicy:
		return oup.UpdatePolicy(compute_v1.UpdatePolicy_name[eint]), nil
	default:
		zlog.InfraSec().InfraError("unknown Enum field %s", fname).Msg("")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Enum field %s", fname)
	}
}

func (is *InvStore) CreateOSUpdatePolicy(ctx context.Context, in *compute_v1.OSUpdatePolicyResource) (*inv_v1.Resource, error) {
	if err := validate(in, osUpPolicyResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, osUpdatePolicyResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("OS Update Policy Created: %s, %s", res.GetOsUpdatePolicy().GetResourceId(), res)
	return res, nil
}

func osUpdatePolicyResourceCreator(in *compute_v1.OSUpdatePolicyResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY)
		zlog.Debug().Msgf("CreateOSUpdatePolicy: %s", id)

		newEntity := tx.OSUpdatePolicyResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, OSUpdatePolicyEnumStateMap, nil); err != nil {
			return nil, err
		}

		// Look up the edges
		if err := setEdgeTargetOSIDForMut(ctx, tx.Client(), mut, in.GetTargetOs()); err != nil {
			return nil, err
		}

		if err := mut.SetField(oup.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getOSUpdatePolicyQuery(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(res))
	}
}

func (is *InvStore) GetOSUpdatePolicy(ctx context.Context, id string) (*inv_v1.Resource, error) {
	res, err := ExecuteInRoTxAndReturnSingle[ent.OSUpdatePolicyResource](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.OSUpdatePolicyResource, error) {
			return getOSUpdatePolicyQuery(ctx, tx, id)
		})
	if err != nil {
		return nil, err
	}

	apiResource := entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_OsUpdatePolicy{OsUpdatePolicy: apiResource}}, nil
}

func getOSUpdatePolicyQuery(ctx context.Context, tx *ent.Tx, resourceID string) (*ent.OSUpdatePolicyResource, error) {
	entity, err := tx.OSUpdatePolicyResource.Query().
		Where(oup.ResourceID(resourceID)).
		WithTargetOs().
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return entity, nil
}

func (is *InvStore) UpdateOSUpdatePolicy(
	ctx context.Context,
	id string,
	in *compute_v1.OSUpdatePolicyResource,
	fieldmask *fieldmaskpb.FieldMask,
) (*inv_v1.Resource, error) {
	zlog.Debug().Msgf("UpdateosUpdatePolicy (%s): %v, fm: %v", id, in, fieldmask)
	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx,
		func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
			entity, err := tx.OSUpdatePolicyResource.Query().
				// we need to also retrieve immutable fields to do the check
				Select(oup.FieldID).
				Where(oup.ResourceID(id)).
				WithTargetOs().
				Only(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			// Target OS is immutable.
			if in.GetTargetOs() != nil && entity.Edges.TargetOs.ResourceID != in.GetTargetOs().GetResourceId() {
				return nil, errors.Errorfc(codes.InvalidArgument, "Cannot change target OS for an existing OS Update Policy")
			}

			updateBuilder := tx.OSUpdatePolicyResource.UpdateOneID(entity.ID)
			mut := updateBuilder.Mutation()

			// No need to update the edge, it's immutable!

			err = buildEntMutate(in, mut, OSUpdatePolicyEnumStateMap, fieldmask.GetPaths())
			if err != nil {
				return nil, err
			}

			_, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			res, err := getOSUpdatePolicyQuery(ctx, tx, id)
			if err != nil {
				return nil, err
			}

			protoRes := entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(res)
			if err := validateOSUpdatePolicyProto(protoRes); err != nil {
				return nil, err
			}
			return util.WrapResource(protoRes)
		})
	if err != nil {
		return nil, err
	}

	return res, err
}

func (is *InvStore) DeleteOSUpdatePolicy(ctx context.Context, id string) (*inv_v1.Resource, error) {
	// this is a "Hard Delete" as os don't have state to reconcile
	zlog.Debug().Msgf("DeleteOSUpdatePolicy Hard Delete: %s", id)

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, deleteOSUpdatePolicy(id))

	return res, err
}

func deleteOSUpdatePolicy(resourceID string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		count, err := tx.InstanceResource.Query().
			Where(instanceresource.HasOsUpdatePolicyWith(oup.ResourceID(resourceID))).
			Count(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if count > 0 {
			return nil, errors.Errorfc(codes.FailedPrecondition,
				"Cannot delete OS Update Policy %s, it is still referenced by %d instances",
				resourceID, count)
		}
		entity, err := tx.OSUpdatePolicyResource.Query().
			Where(oup.ResourceID(resourceID)).
			Only(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		err = tx.OSUpdatePolicyResource.DeleteOneID(entity.ID).Exec(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		return util.WrapResource(entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(entity))
	}
}

func (is *InvStore) DeleteOSUpdatePolicies(
	ctx context.Context, tenantID string, _ bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]
	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			collection, err := tx.OSUpdatePolicyResource.Query().Where(oup.TenantID(tenantID)).All(ctx)
			if err != nil {
				return err
			}
			if _, err := tx.OSUpdatePolicyResource.Delete().Where(oup.TenantID(tenantID)).Exec(ctx); err != nil {
				return err
			}
			for _, element := range collection {
				res, err := util.WrapResource(entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(HARD, res))
			}
			return nil
		})
	return deleted, txErr
}

func filterOSUpdatePolicies(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) (
	[]*ent.OSUpdatePolicyResource, int, error,
) {
	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[oup.OrderOption](filter.GetOrderBy(), oup.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// perform query - And together all the predicates
	query := client.OSUpdatePolicyResource.Query().
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
	total, err := client.OSUpdatePolicyResource.Query().
		Where(pred).
		Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return osList, total, nil
}

func (is *InvStore) ListOSUpdatePolicies(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.OSUpdatePolicyResource, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]*ent.OSUpdatePolicyResource, *int, error) {
			resources, total, err := filterOSUpdatePolicies(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &resources, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[*ent.OSUpdatePolicyResource, *inv_v1.GetResourceResponse](*resources,
		func(res *ent.OSUpdatePolicyResource) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_OsUpdatePolicy{
						OsUpdatePolicy: entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(res),
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

func (is *InvStore) FilterOSUpdatePolicies(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*cl.ResourceTenantIDCarrier, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.OSUpdatePolicyResource, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]*ent.OSUpdatePolicyResource, *int, error) {
			filtered, total, err := filterOSUpdatePolicies(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}

	ids := collections.MapSlice[*ent.OSUpdatePolicyResource, *cl.ResourceTenantIDCarrier](
		*resources, func(c *ent.OSUpdatePolicyResource) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.TenantID, ResourceId: c.ResourceID}
		})

	return ids, *total, err
}

func getOSPolicyUpdateIDFromResourceID(
	ctx context.Context, client *ent.Client, osPolicyUpdate *compute_v1.OSUpdatePolicyResource,
) (int, error) {
	oup, qerr := client.OSUpdatePolicyResource.Query().
		Where(oup.ResourceID(osPolicyUpdate.ResourceId)).
		Only(ctx)
	if qerr != nil {
		return 0, errors.Wrap(qerr)
	}
	return oup.ID, nil
}
