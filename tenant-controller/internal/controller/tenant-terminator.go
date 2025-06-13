// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_util "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/invclient"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/util"
)

// Please be aware that order of definitions below is the key.
// Inventory resources need to be deleted in this particular order, mixing it can cause unpredictable problems.
var (
	tenantTerminationSteps = []*terminationStep{
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_TENANT, terminationFunc: tenantSoftDeletion, watchEvents: true},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD, terminationFunc: workloadsDeletion, watchEvents: true},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE, terminationFunc: softDeletionNoFallback, watchEvents: true},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_HOST, terminationFunc: softDeletionNoFallback, watchEvents: true},
		// {resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_NETLINK, terminationFunc: hardDeletion}, // not used.
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_OS, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_SITE, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_REGION, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER, terminationFunc: hardDeletion},
		// {resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF, terminationType: RESOURCE_SOFT_DELETION}, // not used
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_OU, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG, terminationFunc: hardDeletion},
		{resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_TENANT, terminationFunc: tenantHardDeletion},
	}

	eventDispatcherTimeout = time.Second / 2
)

func NewTerminationController(ic *invclient.TCInventoryClient) *TerminationController {
	return &TerminationController{
		terminators: &syncMap[string, *TenantTerminator]{
			data: map[string]*TenantTerminator{},
		},
		ic: ic,
	}
}

type TerminationController struct {
	terminators *syncMap[string, *TenantTerminator]
	ic          *invclient.TCInventoryClient
}

func (t *TerminationController) HandleEvent(we *client.WatchEvents) {
	t.terminators.ForEach(func(v *TenantTerminator) {
		select {
		case v.invEvents <- we:
		case <-time.After(eventDispatcherTimeout):
		}
	})
}

func (t *TerminationController) TerminateTenant(ctx context.Context, tenantID string) error {
	log.Debug().Msgf("TerminateTenant(%s)", tenantID)
	events := make(chan *client.WatchEvents)
	terminator := NewTenantTerminator(t.ic, events, tenantID, tenantTerminationSteps)
	defer func() {
		log.Debug().Msgf("Unregistering TenantTerminator(%s)", tenantID)
		t.terminators.Remove(tenantID)
		close(events)
	}()
	if !t.terminators.PutIfAbsent(tenantID, terminator) {
		return errors.Errorfc(codes.AlreadyExists, "Tenant(%s) termination is already running", tenantID)
	}
	return terminator.Run(ctx)
}

func NewTenantTerminator(
	ic InventoryClient, events chan *client.WatchEvents, tenantID string, steps []*terminationStep,
) *TenantTerminator {
	c := steps[0]
	for idx, next := range steps {
		if idx == 0 {
			continue
		}

		c.next = next
		c = next
	}
	return &TenantTerminator{
		ic:        ic,
		tenantID:  tenantID,
		chain:     steps[0],
		invEvents: events,
	}
}

type TenantTerminator struct {
	tenantID  string
	ic        InventoryClient
	chain     *terminationStep
	invEvents chan *client.WatchEvents
}

func (t *TenantTerminator) Run(ctx context.Context) error {
	log.Debug().Msgf("Run(): %s", t.String())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for underProcessing := t.chain; underProcessing != nil; underProcessing = underProcessing.next {
		rk := underProcessing.resourceKind
		log.Debug().Msgf("Terminator(tenantID=%s): Executing deletion of %s", t.tenantID, rk)
		if err := underProcessing.terminationFunc(ctx, t, rk); err != nil {
			log.Err(err).Msgf("Unexpected error when terminating %s for %s", rk, t.tenantID)
			return err
		}

		underProcessing.status.executed = true
		log.Debug().Msgf("Terminator(tenantID=%s): %s terminaiton done", t.tenantID, rk)
		log.Debug().Msgf("TenantTerminationStatus: %s", t)
	}
	log.Debug().Msgf("Terminator(tenantID=%s): All resources deleted", t.tenantID)
	return nil
}

func (t *TenantTerminator) executeHardDeletion(ctx context.Context, rk inv_v1.ResourceKind) error {
	if err := t.ic.DeleteAllResources(ctx, t.tenantID, rk, true); err != nil {
		log.Err(err).Msgf("Unexpected error when inv.DeleteResources[kind:%s, tid:%s, enforce: true]", rk, t.tenantID)
		return err
	}
	return nil
}

func (t *TenantTerminator) executeSoftDeletion(ctx context.Context, rk inv_v1.ResourceKind) error {
	allDeleted, err := t.checkIfAllDeleted(ctx, rk)
	if err != nil {
		return err
	}
	if allDeleted {
		log.Debug().Msgf("Terminator(tenantID=%s): zero instances of %s detected, nothing to do", t.tenantID, rk)
		return nil
	}

	condition := func(c context.Context) (bool, error) { return t.checkIfAllDeleted(c, rk) }
	promise := util.Run[any](ctx, waitUntil(t.tenantID, rk, t.invEvents, condition))

	if err := t.ic.DeleteAllResources(ctx, t.tenantID, rk, false); err != nil {
		log.Err(err).Msgf("Unexpected error on inv.DeleteResources[kind:%s, tid:%s, enforce: false]", rk, t.tenantID)
		return err
	}

	if _, err = promise.Await(); err != nil {
		return err
	}
	return nil
}

func (t *TenantTerminator) checkIfAllDeleted(ctx context.Context, rk inv_v1.ResourceKind) (bool, error) {
	log.Debug().Msgf("checkIfAllDeleted(tenantID=%s, rk=%s)", t.tenantID, rk)
	filter, err := createFindFirstResourceForTenant(rk, t.tenantID)
	if err != nil {
		// shall never happen
		return false, errors.Errorfc(
			codes.FailedPrecondition, "Cannot create query filter(resourceKind=%s, tenantID=%s): %s", rk, t.tenantID, err)
	}

	all, err := t.ic.FindAll(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(all) == 0, nil
}

func (t *TenantTerminator) String() string {
	var current *terminationStep
	var steps []string
	for {
		if current == nil {
			current = t.chain
		}
		steps = append(steps, current.String())
		if current.next == nil {
			break
		}
		current = current.next
	}

	return fmt.Sprintf("TenantTeminator(tenantID=%s): %s", t.tenantID, strings.Join(steps, " >>> "))
}

type terminationStrategy func(ctx context.Context, t *TenantTerminator, rk inv_v1.ResourceKind) error

// workloadsDeletion performs workloads/members deletion.
// Workloads and WorkloadMembers are managed by CO so deletion shall be executed by CO.
// TC just waits for CO to delete them, if CO cannot perform deletion within defined time(workloadSoftDeletionTimeout),
// TC will execute hard deletion for members and workloads.
func workloadsDeletion(ctx context.Context, t *TenantTerminator, rk inv_v1.ResourceKind) error {
	waitUntilCtx, cancel := context.WithTimeout(ctx, *configuration.WorkloadSoftDeletionTimeout)
	defer cancel()

	condition := func(c context.Context) (bool, error) { return t.checkIfAllDeleted(c, rk) }
	promise := util.Run[any](waitUntilCtx, waitUntil(t.tenantID, rk, t.invEvents, condition))
	if _, err := promise.Await(); err != nil {
		log.Warn().Msgf(`Timeout: Workloads expected to be deleted by CO are still 
there, executing hard deletion for members/workloads`)

		return util.Retry(configuration.DefaultBackoff, func() error {
			log.Debug().Msgf("Hard deleting workloadMembers for tenantID=%s", t.tenantID)
			membersDeletionErr := t.executeHardDeletion(ctx, inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER)
			if membersDeletionErr != nil {
				return err
			}
			log.Debug().Msgf("Hard deleting workloads for tenantID=%s", t.tenantID)
			return t.executeHardDeletion(ctx, inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD)
		}, "")
	}
	return nil
}

func tenantHardDeletion(ctx context.Context, t *TenantTerminator, rk inv_v1.ResourceKind) error {
	return util.Retry(
		configuration.DefaultBackoff,
		func() error {
			tid, rid, err := t.ic.GetTenantResource(ctx, t.tenantID)
			if errors.IsNotFound(err) {
				log.Debug().Msgf("Tenant(%s) does not exist, nothing to do", t.tenantID)
				return nil
			}
			if err != nil {
				return err
			}

			if err := t.ic.HardDeleteTenantResource(ctx, tid, rid); err != nil {
				log.Err(err).Msgf("Cannot hard delete Tenant(tid=%s, rid=%s)", tid, rid)
				return err
			}
			return nil
		},
		"Error during hard deletion of %s(%s)", rk, t.tenantID,
	)
}

func softDeletionNoFallback(ctx context.Context, t *TenantTerminator, rk inv_v1.ResourceKind) error {
	return util.RetryAndHandleError(
		configuration.DefaultBackoff,
		func() error {
			ctx, cancel := context.WithTimeout(ctx, *configuration.ResourceSoftDeletionTimeout)
			defer cancel()
			return t.executeSoftDeletion(ctx, rk)
		},
		func(actualError error) error {
			if *configuration.EnableHardDeletionFallback {
				return t.executeHardDeletion(ctx, rk)
			}
			return actualError
		},
		"Error during soft deletion of %s for tenant %s", rk, t.tenantID,
	)
}

func hardDeletion(ctx context.Context, t *TenantTerminator, rk inv_v1.ResourceKind) error {
	return util.Retry(
		configuration.DefaultBackoff,
		func() error { return t.executeHardDeletion(ctx, rk) },
		"Error during hard deletion of %s for tenant %s", rk, t.tenantID,
	)
}

func tenantSoftDeletion(ctx context.Context, t *TenantTerminator, rk inv_v1.ResourceKind) error {
	tid, rid, err := t.ic.GetTenantResource(ctx, t.tenantID)
	if errors.IsNotFound(err) {
		log.Debug().Msgf("Tenant(%s) does not exist nothing to do", t.tenantID)
		return nil
	}
	if err != nil {
		return err
	}

	condition := func(c context.Context) (bool, error) {
		tenantResource, err := t.ic.GetTenantResourceInstance(c, t.tenantID)
		if err != nil {
			return false, err
		}
		return !tenantResource.WatcherOsmanager, nil
	}
	promise := util.Run[any](ctx, waitUntil(t.tenantID, rk, t.invEvents, condition))

	if _, err := t.ic.Delete(ctx, tid, rid); err != nil {
		log.Err(err).Msgf("Unexpected error on inv.DeleteResource(tid=%s, rid=%s)", tid, rid)
		return err
	}

	log.Debug().Msgf("Waiting until no watchers for Tenant(%s)", t.tenantID)
	if _, err = promise.Await(); err != nil {
		return err
	}
	return nil
}

type terminationStep struct {
	resourceKind    inv_v1.ResourceKind
	next            *terminationStep
	status          terminationStepStatus
	terminationFunc terminationStrategy
	watchEvents     bool
}

func (t *terminationStep) String() string {
	return fmt.Sprintf("TerminationStep[rk=%s, status=(%v)]", t.resourceKind, t.status)
}

type terminationStepStatus struct {
	executed bool
}

func (t terminationStepStatus) String() string {
	if t.executed {
		return "EXECUTED"
	}
	return "PENDING"
}

func createFindFirstResourceForTenant(rk inv_v1.ResourceKind, tenantID string) (*inv_v1.ResourceFilter, error) {
	resource, err := inv_util.GetResourceFromKind(rk)
	if err != nil {
		log.Err(err).
			Msgf("Unrecognized resource kind: %s, termination of that kind cannot be continued", rk)
		return nil, err
	}

	return &inv_v1.ResourceFilter{
		Resource: resource, Filter: filters.NewBuilderWith(filters.ValEq("tenant_id", tenantID)).Build(), Limit: 1,
	}, nil
}

type untilCondition func(ctx context.Context) (bool, error)

func waitUntil(
	tid string, rk inv_v1.ResourceKind, invEvents <-chan *client.WatchEvents, condition untilCondition,
) util.Task[any] {
	return func(ctx context.Context) (any, error) {
		if ok, _ := condition(ctx); ok { //nolint:errcheck // do not care about error, if condition evaluate to false, continue
			return nil, nil
		}
		return waitUntilLoop(ctx, invEvents, tid, rk, condition)
	}
}

func waitUntilLoop(
	ctx context.Context, invEvents <-chan *client.WatchEvents, tid string, rk inv_v1.ResourceKind, condition untilCondition,
) (any, error) {
	log.Debug().Msgf("Waiting until no %s for Tenant(%s)", rk, tid)
	ticker := time.NewTicker(*configuration.TenantTerminatorWaitUntilTickerInterval)
	defer ticker.Stop()

	for {
		select {
		case ie, ok := <-invEvents:
			if !ok {
				log.Debug().Msgf("waitUntil(tid=%s, rk=%s) - events channel has been closed", tid, rk)
				return ok, nil
			}
			eventMatchingTenantID := isEventRelatedWithTenant(ie.Event, tid)
			eventMatchingResourceKind := hasEventResourceKind(ie.Event, rk)
			if !eventMatchingTenantID || !eventMatchingResourceKind {
				log.Trace().Msgf("WaitUntilCondition(tid=%s, rk=%s) - ignoring event(%v)", tid, rk, ie.Event)
				continue
			}
		case <-ticker.C:
			log.Debug().Msgf("WaitUntilCondition(tid=%s, rk=%s) - resumed by ticker", tid, rk)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		ok, err := condition(ctx)
		if err != nil {
			log.Err(err).Msgf("Error while checking condition, context[tenantID: %s, rk=%s]", tid, rk)
			return nil, err
		}
		if ok {
			return ok, nil
		}
	}
}

func isEventRelatedWithTenant(event *inv_v1.SubscribeEventsResponse, tenantID string) bool {
	tid, _, err := inv_util.GetResourceKeyFromResource(event.GetResource())
	if err != nil {
		// shall never happen
		log.Err(err).Msgf("cannot exctract event from event(%v)", event)
		return false
	}
	return tid == tenantID
}

func hasEventResourceKind(event *inv_v1.SubscribeEventsResponse, rk inv_v1.ResourceKind) bool {
	erk := inv_util.GetResourceKindFromResource(event.GetResource())
	return erk == rk
}

type syncMap[K comparable, V any] struct {
	sync.RWMutex
	data map[K]V
}

func (s *syncMap[K, V]) PutIfAbsent(k K, v V) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.data[k]
	if ok {
		return false
	}
	s.data[k] = v
	return true
}

func (s *syncMap[K, V]) Remove(k K) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, k)
}

func (s *syncMap[K, V]) ForEach(f func(v V)) {
	s.RLock()
	defer s.RUnlock()
	for _, v := range s.data {
		f(v)
	}
}
