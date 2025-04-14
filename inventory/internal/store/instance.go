// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

// instance.go  store information for Instance objects

import (
	"context"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/utils/booleans"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var instanceResourceCreationValidators = []resourceValidator[*computev1.InstanceResource]{
	protoValidator[*computev1.InstanceResource],
	validateInstanceProto,
	doNotAcceptResourceID[*computev1.InstanceResource],
}

// validateInstanceProto checks that all constraints on an Instance proto message are fulfilled.
func validateInstanceProto(in *computev1.InstanceResource) error {
	// Check that only fields applicable to the instance kind are set.
	if in.GetKind() == computev1.InstanceKind_INSTANCE_KIND_METAL {
		if in.GetVmMemoryBytes() > 0 || in.GetVmCpuCores() > 0 || in.GetVmStorageBytes() > 0 {
			return errors.Errorfc(codes.InvalidArgument, "VM fields cannot be set on a baremetal instance")
		}
	}
	if in.GetKind() == computev1.InstanceKind_INSTANCE_KIND_VM {
		if in.GetHost() != nil {
			return errors.Errorfc(codes.InvalidArgument, "host cannot be set on a VM instance")
		}
	}

	return nil
}

// InstanceEnumStateMap maps proto enum fields to their Ent equivalents.
func InstanceEnumStateMap(fname string, eint int32) (ent.Value, error) {
	switch fname {
	case instanceresource.FieldDesiredState:
		return instanceresource.DesiredState(computev1.InstanceState_name[eint]), nil

	case instanceresource.FieldCurrentState:
		return instanceresource.CurrentState(computev1.InstanceState_name[eint]), nil

	case instanceresource.FieldKind:
		return instanceresource.Kind(computev1.InstanceKind_name[eint]), nil

	case instanceresource.FieldInstanceStatusIndicator:
		return instanceresource.InstanceStatusIndicator(statusv1.StatusIndication_name[eint]), nil

	case instanceresource.FieldProvisioningStatusIndicator:
		return instanceresource.ProvisioningStatusIndicator(statusv1.StatusIndication_name[eint]), nil

	case instanceresource.FieldUpdateStatusIndicator:
		return instanceresource.UpdateStatusIndicator(statusv1.StatusIndication_name[eint]), nil

	case instanceresource.FieldTrustedAttestationStatusIndicator:
		return instanceresource.TrustedAttestationStatusIndicator(statusv1.StatusIndication_name[eint]), nil

	case instanceresource.FieldSecurityFeature:
		return instanceresource.SecurityFeature(osv1.SecurityFeature_name[eint]), nil

	default:
		zlog.InfraSec().InfraError("unknown Enum field %s", fname).Msg("")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Enum field %s", fname)
	}
}

func (is *InvStore) CreateInstance(ctx context.Context, in *computev1.InstanceResource) (*inv_v1.Resource, error) {
	if err := validate(in, instanceResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, instanceResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Instance Created: %s, %s", res.GetInstance().GetResourceId(), res)
	return res, nil
}

func instanceResourceCreator(in *computev1.InstanceResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE)
		zlog.Debug().Msgf("CreateInstance: %s", id)

		newEntity := tx.InstanceResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, InstanceEnumStateMap, nil); err != nil {
			return nil, err
		}

		// Look up the optional host ID for this Instance.
		if err := setEdgeHostIDForMut(ctx, tx.Client(), mut, in.GetHost()); err != nil {
			return nil, err
		}
		// Look up the optional Desired OS ID for this Instance.
		if err := setEdgeDesiredOSIDForMut(ctx, tx.Client(), mut, in.GetDesiredOs()); err != nil {
			return nil, err
		}
		// Look up the optional Desired OS ID for this Instance.
		if err := setEdgeCurrentOSIDForMut(ctx, tx.Client(), mut, in.GetCurrentOs()); err != nil {
			return nil, err
		}
		// Look up the optional provider ID for this host.
		if err := setEdgeProviderIDForMut(ctx, tx.Client(), mut, in.GetProvider()); err != nil {
			return nil, err
		}

		// Look up the optional LocalAccount ID for this Instance.
		if err := setEdgeLocalAccountIDForMut(ctx, tx.Client(), mut, in.GetLocalaccount()); err != nil {
			return nil, err
		}

		if err := mut.SetField(instanceresource.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getInstanceQuery(ctx, tx, id, false)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entInstanceResourceToProtoInstanceResource(res))
	}
}

func (is *InvStore) GetInstance(ctx context.Context, id string) (*inv_v1.Resource, error) {
	res, err := ExecuteInRoTxAndReturnSingle[ent.InstanceResource](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*ent.InstanceResource, error) {
			return getInstanceQuery(ctx, tx, id, true)
		})
	if err != nil {
		return nil, err
	}

	apiResource := entInstanceResourceToProtoInstanceResource(res)
	if err = validator.ValidateMessage(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}
	if err = validateInstanceProto(apiResource); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, errors.Wrap(err)
	}

	return &inv_v1.Resource{Resource: &inv_v1.Resource_Instance{Instance: apiResource}}, nil
}

func getInstanceQuery(ctx context.Context, tx *ent.Tx, resourceID string, nestedLoad bool) (*ent.InstanceResource, error) {
	query := tx.InstanceResource.Query().
		Where(instanceresource.ResourceID(resourceID)).
		WithDesiredOs().
		WithCurrentOs().
		WithProvider().
		WithLocalaccount()
	if nestedLoad {
		query.
			WithHost(func(q *ent.HostResourceQuery) {
				q.WithSite()     // Populate the site of each host
				q.WithProvider() // Populate the provider of each host
			}).
			WithWorkloadMembers(func(q *ent.WorkloadMemberQuery) {
				q.WithWorkload() // Populate the workload of each member
			})
	} else {
		query.WithHost().WithWorkloadMembers()
	}
	entity, err := query.Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return entity, nil
}

//nolint:cyclop // high cyclomatic complexity due to hard-delete.
func (is *InvStore) UpdateInstance(
	ctx context.Context, id string, in *computev1.InstanceResource, fieldmask *fieldmaskpb.FieldMask,
) (*inv_v1.Resource, bool, error) {
	if err := validate(in, validateInstanceProto); err != nil {
		return nil, false, errors.Wrap(err)
	}

	zlog.Debug().Msgf("UpdateInstance (%s): %v, fm: %v", id, in, fieldmask)

	res, hardDelete, err := ExecuteInTxAndReturnDouble[inv_v1.Resource, bool](is)(ctx,
		func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, *bool, error) {
			entity, err := tx.InstanceResource.Query().
				Where(instanceresource.ResourceID(id)).
				Only(ctx)
			if err != nil {
				return nil, booleans.Pointer(false), errors.Wrap(err)
			}

			// hard delete - if both Desired and Current state are Deleted, remove
			if isInstanceHardDelete(fieldmask, entity, in) {
				zlog.Debug().Msgf("UpdateInstance Hard Delete: %s", id)

				// should be nil on success
				err = tx.InstanceResource.DeleteOneID(entity.ID).Exec(ctx)
				if err != nil {
					return nil, booleans.Pointer(false), errors.Wrap(err)
				}

				var wrapped *inv_v1.Resource
				// Set current state to be consistent on the returned value on events and upon update.
				entity.CurrentState = instanceresource.CurrentStateINSTANCE_STATE_DELETED
				wrapped, err = util.WrapResource(entInstanceResourceToProtoInstanceResource(entity))
				if err != nil {
					return nil, booleans.Pointer(false), err
				}
				return wrapped, booleans.Pointer(true), nil
			}

			if isNotValidInstanceTransition(fieldmask, entity, in) {
				zlog.InfraSec().InfraError("%s from %s to %s is not allowed",
					id, entity.CurrentState, in.DesiredState).Msgf("UpdateInstance")
				return nil, booleans.Pointer(false),
					errors.Errorfc(codes.InvalidArgument, "UpdateInstance %s from %s to %s is not allowed",
						id, entity.CurrentState, in.DesiredState)
			}

			if isValidLocalAccountTransition(fieldmask, entity, in) {
				zlog.InfraSec().InfraError("%s from %s to %s is not allowed",
					id, entity.CurrentState, in.GetLocalaccount()).Msgf("UpdateInstance")
				return nil, booleans.Pointer(false),
					errors.Errorfc(codes.InvalidArgument, "UpdateInstance %s from %s to %s LocalAccount is not allowed %s, out %t, cmp %t",
						id, entity.CurrentState, in.DesiredState, in.GetLocalaccount(),
						isValidLocalAccountTransition(fieldmask, entity, in),
						(slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeLocalaccount) &&
							(entity.CurrentState == instanceresource.CurrentStateINSTANCE_STATE_UNSPECIFIED ||
								len(entity.CurrentState) == 0) && in.GetLocalaccount() != nil),
					)
			}

			// Because the instance-to-host edge is O2O and Ent has a limitation that does not allow
			// updating an already set O2O edge, we have to clear it before setting it in the mutation.
			if in.GetHost().GetResourceId() != "" {
				_, err = tx.InstanceResource.UpdateOneID(entity.ID).
					ClearHost().
					Save(ctx)
				if err != nil {
					return nil, booleans.Pointer(false), errors.Wrap(err)
				}
			}

			updateBuilder := tx.InstanceResource.UpdateOneID(entity.ID)
			mut := updateBuilder.Mutation()

			// Look up the (new) referenced edges for this Instance.
			err = setRelationsForInstanceMutIfNeeded(ctx, tx.Client(), mut, in, fieldmask)
			if err != nil {
				return nil, booleans.Pointer(false), err
			}

			err = buildEntMutate(in, mut, InstanceEnumStateMap, fieldmask.GetPaths())
			if err != nil {
				return nil, booleans.Pointer(false), err
			}

			_, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, booleans.Pointer(false), errors.Wrap(err)
			}

			// Get updated resource including eager loaded edges
			res, err := getInstanceQuery(ctx, tx, id, false)
			if err != nil {
				return nil, booleans.Pointer(false), err
			}
			toBeReturned, err := util.WrapResource(entInstanceResourceToProtoInstanceResource(res))
			return toBeReturned, booleans.Pointer(false), err
		},
	)
	if err != nil {
		return nil, false, err
	}

	return res, *hardDelete, err
}

func (is *InvStore) DeleteInstance(ctx context.Context, id string) (*inv_v1.Resource, bool, error) {
	// this is a "Soft Delete" - it only sets the Desired State to Deleted
	// Hard delete happens in Update, when both Desired and Current state are
	// both Deleted.
	zlog.Debug().Msgf("DeleteInstance Soft Delete: %s", id)

	res, isSoftDelete, err := ExecuteInTxAndReturnDouble[inv_v1.Resource, bool](is)(
		ctx,
		deleteInstance(id))
	if err != nil {
		return nil, false, err
	}

	return res, *isSoftDelete, err
}

func deleteInstance(id string) func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, *bool, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, *bool, error) {
		entity, err := tx.InstanceResource.Query().
			Where(instanceresource.ResourceID(id)).
			Only(ctx)
		if err != nil {
			return nil, booleans.Pointer(false), errors.Wrap(err)
		}

		_, err = tx.InstanceResource.UpdateOneID(entity.ID).
			SetDesiredState(instanceresource.DesiredStateINSTANCE_STATE_DELETED).
			Save(ctx)
		if err != nil {
			return nil, booleans.Pointer(false), errors.Wrap(err)
		}

		// Get updated resource including eager loaded edges
		res, err := getInstanceQuery(ctx, tx, id, false)
		if err != nil {
			return nil, booleans.Pointer(false), err
		}
		toBeReturned, err := util.WrapResource(entInstanceResourceToProtoInstanceResource(res))
		if err != nil {
			return nil, booleans.Pointer(false), err
		}

		return toBeReturned, booleans.Pointer(true), nil
	}
}

func (is *InvStore) DeleteInstances(
	ctx context.Context, tenantID string, enforce bool,
) ([]*util.Tuple[DeletionKind, *inv_v1.Resource], error) {
	deletionStrategies := map[bool]func(ctx context.Context, tx *ent.Tx, tenantID string) (DeletionKind, int, error){
		true: func(ctx context.Context, tx *ent.Tx, tenantID string) (DeletionKind, int, error) {
			i, e := tx.InstanceResource.Delete().Where(instanceresource.TenantID(tenantID)).Exec(ctx)
			return HARD, i, e
		},
		false: func(ctx context.Context, tx *ent.Tx, tenantID string) (DeletionKind, int, error) {
			i, e := tx.InstanceResource.Update().
				Where(instanceresource.TenantID(tenantID)).
				SetDesiredState(instanceresource.DesiredStateINSTANCE_STATE_DELETED).
				Save(ctx)
			return SOFT, i, e
		},
	}

	var deleted []*util.Tuple[DeletionKind, *inv_v1.Resource]
	txErr := ExecuteInTx(is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) error {
			collection, err := tx.InstanceResource.Query().Where(instanceresource.TenantID(tenantID)).All(ctx)
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
				collections.ForEach(collection, func(i *ent.InstanceResource) {
					i.DesiredState = instanceresource.DesiredStateINSTANCE_STATE_DELETED
				})
			}
			for _, element := range collection {
				res, err := util.WrapResource(entInstanceResourceToProtoInstanceResource(element))
				if err != nil {
					return err
				}
				deleted = append(deleted, util.NewTuple(dk, res))
			}

			return nil
		})
	return deleted, txErr
}

func (is *InvStore) ListInstances(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.InstanceResource, int](is)(
		ctx,
		func(ctx context.Context, tx *ent.Tx) (*[]*ent.InstanceResource, *int, error) {
			filtered, total, err := filterInstances(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, err
		},
	)
	if err != nil {
		return nil, 0, err
	}

	resps := collections.MapSlice[*ent.InstanceResource, *inv_v1.GetResourceResponse](*resources,
		func(res *ent.InstanceResource) *inv_v1.GetResourceResponse {
			return &inv_v1.GetResourceResponse{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Instance{
						Instance: entInstanceResourceToProtoInstanceResource(res),
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

func filterInstances(ctx context.Context, client *ent.Client, filter *inv_v1.ResourceFilter) (
	[]*ent.InstanceResource, int, error,
) {
	pred, err := getPredicate(inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE, filter.GetFilter())
	if err != nil {
		return nil, 0, err
	}

	orderOpts, err := GetOrderByOptions[instanceresource.OrderOption](filter.GetOrderBy(), instanceresource.ValidColumn)
	if err != nil {
		return nil, 0, err
	}

	offset, limit, err := getOffsetAndLimit(filter)
	if err != nil {
		return nil, 0, err
	}

	// perform query - And together all the predicates with eager loading
	query := client.Debug().InstanceResource.Query().
		WithHost(func(q *ent.HostResourceQuery) {
			q.WithSite()     // Populate the site of each host
			q.WithProvider() // Populate the provider of each host
		}).
		WithDesiredOs().
		WithCurrentOs().
		WithWorkloadMembers(func(q *ent.WorkloadMemberQuery) {
			q.WithWorkload() // Populate the workload of each member
		}).
		WithProvider().
		WithLocalaccount().
		Where(pred).
		Order(orderOpts...).
		Offset(offset)

	// Limits number of query results if existent
	if limit != 0 {
		query = query.Limit(limit)
	}

	instanceresourceList, err := query.All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	// Count total number of item without applying pagination limits, order, or loading edges.
	total, err := client.InstanceResource.Query().
		Where(pred).
		Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return instanceresourceList, total, nil
}

func (is *InvStore) FilterInstances(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*cl.ResourceTenantIDCarrier, int, error,
) {
	resources, total, err := ExecuteInRoTxAndReturnDouble[[]*ent.InstanceResource, int](is)(
		ctx, func(ctx context.Context, tx *ent.Tx) (*[]*ent.InstanceResource, *int, error) {
			filtered, total, err := filterInstances(ctx, tx.Client(), filter)
			if err != nil {
				return nil, nil, err
			}
			return &filtered, &total, nil
		})
	if err != nil {
		return nil, 0, err
	}

	ids := collections.MapSlice[*ent.InstanceResource, *cl.ResourceTenantIDCarrier](
		*resources, func(c *ent.InstanceResource) *cl.ResourceTenantIDCarrier {
			return &cl.ResourceTenantIDCarrier{TenantId: c.TenantID, ResourceId: c.ResourceID}
		})

	return ids, *total, err
}

//nolint:cyclop // high cyclomatic complexity due to large conditions.
func setRelationsForInstanceMutIfNeeded(
	ctx context.Context,
	client *ent.Client,
	mut *ent.InstanceResourceMutation,
	in *computev1.InstanceResource,
	fieldmask *fieldmaskpb.FieldMask,
) error {
	mut.ResetHost()
	if slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeHost) {
		if err := setEdgeHostIDForMut(ctx, client, mut, in.GetHost()); err != nil {
			return err
		}
	}
	mut.ResetDesiredOs()
	if slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeDesiredOs) {
		if err := setEdgeDesiredOSIDForMut(ctx, client, mut, in.GetDesiredOs()); err != nil {
			return err
		}
	}
	mut.ResetCurrentOs()
	if slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeCurrentOs) {
		if err := setEdgeCurrentOSIDForMut(ctx, client, mut, in.GetCurrentOs()); err != nil {
			return err
		}
	}
	mut.ResetProvider()
	if slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeProvider) {
		if err := setEdgeProviderIDForMut(ctx, client, mut, in.GetProvider()); err != nil {
			return err
		}
	}
	mut.ResetLocalaccount()
	if slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeLocalaccount) {
		if err := setEdgeLocalAccountIDForMut(ctx, client, mut, in.GetLocalaccount()); err != nil {
			return err
		}
	}
	return nil
}

func getInstanceIDFromResourceID(ctx context.Context, client *ent.Client, instanceRes *computev1.InstanceResource) (int, error) {
	site, qerr := client.InstanceResource.Query().
		Where(instanceresource.ResourceID(instanceRes.ResourceId)).
		Only(ctx)
	if qerr != nil {
		return 0, errors.Wrap(qerr)
	}
	return site.ID, nil
}

func isInstanceHardDelete(
	fieldmask *fieldmaskpb.FieldMask, instanceq *ent.InstanceResource, in *computev1.InstanceResource,
) bool {
	return slices.Contains(fieldmask.GetPaths(), hostresource.FieldCurrentState) &&
		instanceq.DesiredState == instanceresource.DesiredStateINSTANCE_STATE_DELETED &&
		in.CurrentState == computev1.InstanceState_INSTANCE_STATE_DELETED
}

func isNotValidInstanceTransition(
	fieldmask *fieldmaskpb.FieldMask,
	instanceq *ent.InstanceResource,
	in *computev1.InstanceResource,
) bool {

	// transition from Untrusted to any other state than DELETED is not allowed
	return slices.Contains(fieldmask.GetPaths(), instanceresource.FieldDesiredState) &&
		instanceq.CurrentState == instanceresource.CurrentStateINSTANCE_STATE_UNTRUSTED &&
		in.DesiredState != computev1.InstanceState_INSTANCE_STATE_DELETED
}

func isValidLocalAccountTransition(
	fieldmask *fieldmaskpb.FieldMask,
	instanceq *ent.InstanceResource,
	in *computev1.InstanceResource,
) bool {

	// return failure in case of dont allow - abc state ,
	if slices.Contains(fieldmask.GetPaths(), instanceresource.EdgeLocalaccount) && in.GetLocalaccount() != nil {

		if instanceq.CurrentState != instanceresource.CurrentStateINSTANCE_STATE_UNSPECIFIED {
			// We are in non UNSPECIFIED state, exception is only for "" state - so allow update of LC
			if len(instanceq.CurrentState) == 0 {
				return false // Allow update
			} else {
				return true // Dont allow other state might be running etc
			}
		} else {
			return true
		}
	}
	return false
}
