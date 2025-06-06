// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

// store.go - store for core inventory client/registration/streaming
// currently just keeps in-memory maps
// future work will persist data externally

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	entgosql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // blank import to make sure is included in go.mod
	"google.golang.org/grpc/codes"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/intercept"
	regions "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/status"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
)

var zlog = logging.GetLogger("InfraInvStore")

const (
	createdAtFieldName = "created_at"
	updatedAtFieldName = "updated_at"
	ISO8601Format      = "2006-01-02T15:04:05.999Z"
)

type InvStore struct {
	entClient ent.Client
}

// tenantFilterApplyingInterceptor - provides interceptor automatically applying tenant filter.
func tenantFilterApplyingInterceptor(ctx context.Context, query intercept.Query) error {
	tenantID, ok := tenant.GetTenantIDFromContext(ctx)
	if ok && tenantID != "" {
		query.WhereP(func(selector *entgosql.Selector) {
			selector.Where(entgosql.EQ("tenant_id", tenantID))
		})
	}
	return nil
}

func newEntClient(dbURLWriter, dbURLReader string) ent.Client {
	client := *ConnectEntDB(dbURLWriter, dbURLReader)
	client.Intercept(intercept.TraverseFunc(tenantFilterApplyingInterceptor))
	client.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Store time with millisecond precision, in ISO8601 format.
			now := time.Now().UTC().Format(ISO8601Format)
			switch m.Op() {
			case ent.OpCreate:
				err := applyStringFields(m, true, map[string]string{
					createdAtFieldName: now,
					updatedAtFieldName: now,
				})
				if err != nil {
					return nil, err
				}

				switch m.(type) {
				case *ent.HostResourceMutation:
					err = applyStringFields(m, false, map[string]string{
						hostresource.FieldHostStatus:         status.DefaultHostStatus,
						hostresource.FieldOnboardingStatus:   status.DefaultOnboardingStatus,
						hostresource.FieldRegistrationStatus: status.DefaultRegistrationStatus,
						hostresource.FieldPowerStatus:        status.DefaultPowerStatus,
						hostresource.FieldAmtStatus:          status.DefaultAmtStatus,
						hostresource.FieldAmtSku:             "Unknown",
					})
					if err != nil {
						return nil, err
					}
				case *ent.InstanceResourceMutation:
					err = applyStringFields(m, false, map[string]string{
						instanceresource.FieldInstanceStatus:           status.DefaultInstanceStatus,
						instanceresource.FieldProvisioningStatus:       status.DefaultProvisioningStatus,
						instanceresource.FieldUpdateStatus:             status.DefaultUpdateStatus,
						instanceresource.FieldTrustedAttestationStatus: status.DefaultTrustedAttestationStatus,
					})
					if err != nil {
						return nil, err
					}
				default:
					// Do nothing.
				}
			case ent.OpUpdate, ent.OpUpdateOne:
				err := m.SetField(updatedAtFieldName, now)
				if err != nil {
					return nil, err
				}
			default:
				// Do nothing.
			}
			return next.Mutate(ctx, m)
		})
	})
	return client
}

func applyStringFields(m ent.Mutation, allowOverride bool, fields map[string]string) error {
	for field, value := range fields {
		if _, isSet := m.Field(field); isSet && !allowOverride {
			continue
		}

		if err := m.SetField(field, value); err != nil {
			zlog.InfraSec().InfraErr(err).Msgf("failed to set field %s", field)
			return err
		}
	}
	return nil
}

func NewStore(dbURLWriter, dbURLReader string) *InvStore {
	is := new(InvStore)

	// connect to DB
	is.entClient = newEntClient(dbURLWriter, dbURLReader)

	resourceTranspilerRegistry = newRegistry()
	return is
}

func setEdgeRegionIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, regionres *locationv1.RegionResource,
) error {
	if regionres == nil {
		return nil
	}

	reg, qerr := client.RegionResource.Query().
		Where(regions.ResourceID(regionres.ResourceId)).
		Only(ctx)
	if qerr != nil {
		return errors.Wrap(qerr)
	}
	regID := reg.ID

	switch mut := mut.(type) {
	case *ent.TelemetryProfileMutation:
		mut.SetRegionID(regID)
	case *ent.RepeatedScheduleResourceMutation:
		mut.SetTargetRegionID(regID)
	case *ent.SingleScheduleResourceMutation:
		mut.SetTargetRegionID(regID)
	case *ent.SiteResourceMutation:
		mut.SetRegionID(regID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeSiteIDForMut(ctx context.Context, client *ent.Client, mut ent.Mutation, siteres *locationv1.SiteResource) error {
	if siteres == nil {
		return nil
	}
	siteID, qerr := getSiteIDFromResourceID(ctx, client, siteres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.HostResourceMutation:
		mut.SetSiteID(siteID)
	case *ent.NetworkSegmentMutation:
		mut.SetSiteID(siteID)
	case *ent.RepeatedScheduleResourceMutation:
		mut.SetTargetSiteID(siteID)
	case *ent.SingleScheduleResourceMutation:
		mut.SetTargetSiteID(siteID)
	case *ent.TelemetryProfileMutation:
		mut.SetSiteID(siteID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeProviderIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, provres *providerv1.ProviderResource,
) error {
	if provres == nil {
		return nil
	}
	providerID, qerr := getProviderIDFromResourceID(ctx, client, provres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.HostResourceMutation:
		mut.SetProviderID(providerID)
	case *ent.InstanceResourceMutation:
		mut.SetProviderID(providerID)
	case *ent.SiteResourceMutation:
		mut.SetProviderID(providerID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeLocalAccountIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, localaccountres *localaccountv1.LocalAccountResource,
) error {
	if localaccountres == nil {
		return nil
	}
	localaccountID, qerr := getLocalAccountIDFromResourceID(ctx, client, localaccountres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.InstanceResourceMutation:
		mut.SetLocalaccountID(localaccountID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeHostIDForMut(ctx context.Context, client *ent.Client, mut ent.Mutation, hostres *computev1.HostResource) error {
	if hostres == nil {
		return nil
	}
	hostID, qerr := getHostIDFromResourceID(ctx, client, hostres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.HostnicResourceMutation:
		mut.SetHostID(hostID)
	case *ent.HoststorageResourceMutation:
		mut.SetHostID(hostID)
	case *ent.HostusbResourceMutation:
		mut.SetHostID(hostID)
	case *ent.HostgpuResourceMutation:
		mut.SetHostID(hostID)
	case *ent.RepeatedScheduleResourceMutation:
		mut.SetTargetHostID(hostID)
	case *ent.SingleScheduleResourceMutation:
		mut.SetTargetHostID(hostID)
	case *ent.InstanceResourceMutation:
		mut.SetHostID(hostID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeDesiredOSIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, osres *osv1.OperatingSystemResource,
) error {
	if osres == nil {
		return nil
	}
	osID, qerr := getOSIDFromResourceID(ctx, client, osres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.InstanceResourceMutation:
		mut.SetDesiredOsID(osID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msgf("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeCurrentOSIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, osres *osv1.OperatingSystemResource,
) error {
	if osres == nil {
		return nil
	}
	osID, qerr := getOSIDFromResourceID(ctx, client, osres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.InstanceResourceMutation:
		mut.SetCurrentOsID(osID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msgf("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeOSIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, osres *osv1.OperatingSystemResource,
) error {
	if osres == nil {
		return nil
	}
	osID, qerr := getOSIDFromResourceID(ctx, client, osres)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.InstanceResourceMutation:
		mut.SetOsID(osID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msgf("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeWorkloadIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, workloadRes *computev1.WorkloadResource,
) error {
	if workloadRes == nil {
		return nil
	}
	workloadID, qerr := getWorkloadIDFromResourceID(ctx, client, workloadRes)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.WorkloadMemberMutation:
		mut.SetWorkloadID(workloadID)
	case *ent.RepeatedScheduleResourceMutation:
		mut.SetTargetWorkloadID(workloadID)
	case *ent.SingleScheduleResourceMutation:
		mut.SetTargetWorkloadID(workloadID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

type InstanceCarrier interface {
	SetInstanceID(id int)
}

func setEdgeInstanceIDForMut(
	ctx context.Context, client *ent.Client, ic InstanceCarrier, instanceRes *computev1.InstanceResource,
) error {
	if instanceRes == nil {
		return nil
	}
	instanceID, err := getInstanceIDFromResourceID(ctx, client, instanceRes)
	if err != nil {
		return err
	}
	ic.SetInstanceID(instanceID)
	return nil
}

func setParentRegionForRegionMut(
	ctx context.Context, client *ent.Client, mut *ent.RegionResourceMutation, regionRes *locationv1.RegionResource,
) error {
	if regionRes == nil {
		return nil
	}

	regionID, err := getRegionIDFromResourceID(ctx, client, regionRes)
	if err != nil {
		return err
	}
	mut.SetParentRegionID(regionID)
	return nil
}

func setEdgeTelemetryGroupIDForMut(
	ctx context.Context, client *ent.Client, mut ent.Mutation, group *telemetry_v1.TelemetryGroupResource,
) error {
	if group == nil {
		return nil
	}
	groupID, qerr := getTelemetryGroupIDFromResourceID(ctx, client, group)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.TelemetryProfileMutation:
		mut.SetGroupID(groupID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func setEdgeNicIDForMut(ctx context.Context, client *ent.Client, mut ent.Mutation, nicRes *computev1.HostnicResource) error {
	if nicRes == nil {
		return nil
	}
	nicID, qerr := getNicIDFromResourceID(ctx, client, nicRes)
	if qerr != nil {
		return qerr
	}
	switch mut := mut.(type) {
	case *ent.IPAddressResourceMutation:
		mut.SetNicID(nicID)
	default:
		zlog.InfraSec().InfraError("unknown mutation kind: %T", mut).Msg("")
		return errors.Errorfc(codes.InvalidArgument, "unknown mutation kind: %T", mut)
	}
	return nil
}

func isEntLeaf(err error) bool {
	if ent.IsNotFound(err) || ent.IsNotLoaded(err) {
		return true // We found a leaf.
	}
	return false
}

func isEntRoot(err error) bool {
	return isEntLeaf(err)
}

// Starts a new transaction with RepeatableRead isolation policy.
func (is *InvStore) startTransaction(ctx context.Context) (*ent.Tx, error) {
	txOpts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	tx, err := is.entClient.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return tx, nil
}

// Starts a new read-only transaction with RepeatableRead isolation policy.
func (is *InvStore) startReadTransaction(ctx context.Context) (*ent.Tx, error) {
	txOpts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}
	tx, err := is.entClient.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return tx, nil
}

// Commit the given transaction.
func (is *InvStore) commitTransaction(tx *ent.Tx) error {
	return errors.Wrap(tx.Commit())
}

// rollback calls to tx.Rollback and wraps the given error with the rollback error if occurred.
func (is *InvStore) rollbackTransaction(tx *ent.Tx, err error) error {
	if rollbackError := tx.Rollback(); rollbackError != nil {
		err = errors.Wrap(fmt.Errorf("%w: %w", err, rollbackError))
	}
	return err
}

// Close the ent client.
func (is *InvStore) CloseEntClient() error {
	return is.entClient.Close()
}

type DeletionKind string

const (
	SOFT DeletionKind = "soft"
	HARD DeletionKind = "hard"
)

func (is *InvStore) TestGetEntClient() *ent.Client {
	return &is.entClient
}
