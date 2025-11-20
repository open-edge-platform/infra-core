// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package datamodel

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/controller"
	recv2 "github.com/open-edge-platform/orch-library/go/pkg/controller/v2"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
)

const (
	parallelism = 1

	exponentialBackoffMinDelay = 3 * time.Second
	exponentialBackoffMaxDelay = time.Minute
)

type ProjectID string

func (id ProjectID) String() string {
	return string(id)
}

func NewProjectID(projectID types.UID) ProjectID {
	return ProjectID(projectID)
}

type ProjectReconciliationController interface {
	ControlLoop(chan bool)
}

type tenantTerminationHandler interface {
	TerminateTenant(ctx context.Context, tenantID string) error
}

type tenantInitializationHandler interface {
	InitializeTenant(ctx context.Context, config controller.ProjectConfig) error
}

type nexusClient interface {
	GetRuntimeProjectByUID(
		ctx context.Context, tenantID string,
	) (*nexus_client.RuntimeprojectRuntimeProject, error)
	RegisterActiveWatcher(
		rp *nexus_client.RuntimeprojectRuntimeProject,
	) (*nexus_client.ProjectactivewatcherProjectActiveWatcher, error)
	ReportError(aw *nexus_client.ProjectactivewatcherProjectActiveWatcher, msg, tenantID string) error
}

func newProjectReconciliationController(
	nxc nexusClient, trace bool,
	invTenantTerminationHandler tenantTerminationHandler,
	invTenantInitializationHandler tenantInitializationHandler,
	initializations, terminations chan *nexus_client.RuntimeprojectRuntimeProject,
) ProjectReconciliationController {
	initializer := recv2.NewController[ProjectID](
		reconcileProjectCreation(nxc, invTenantInitializationHandler, trace),
		recv2.WithParallelism(parallelism), recv2.WithTimeout(*configuration.TenantInitializationTimeout),
	)

	terminator := recv2.NewController[ProjectID](
		reconcileProjectTermination(nxc, invTenantTerminationHandler, trace),
		recv2.WithParallelism(parallelism), recv2.WithTimeout(*configuration.TenantTerminationTimeout),
	)

	return &reconciler{
		initializer:             initializer,
		terminator:              terminator,
		termChan:                make(chan struct{}),
		nxc:                     nxc,
		projectsToBeInitialized: initializations,
		projectsToBeTerminated:  terminations,
	}
}

func reconcileProjectCreation(
	nxc nexusClient, invTenantInitializationHandler tenantInitializationHandler, trace bool,
) recv2.Reconciler[ProjectID] {
	return func(ctx context.Context, request recv2.Request[ProjectID]) recv2.Directive[ProjectID] {
		if trace {
			ctx = tracing.StartTrace(ctx, "infra-tc", "ProjectCreationReconciler")
			defer tracing.StopTrace(ctx)
		}
		log.Info().Msgf("Reconciling rPROJECT[%s]", request.ID)

		rp, err := nxc.GetRuntimeProjectByUID(ctx, request.ID.String())
		if err != nil {
			log.Err(err).Msgf("rPROJECT[%s] does not exists, cannot reconcile project creation", request.ID)
			return request.Fail(err)
		}

		_, err = nxc.RegisterActiveWatcher(rp)
		if err != nil {
			log.Err(err).Msgf("Cannot register actiwe watcher for rPROJECT[%s/%s]", rp.GetUID(), rp.DisplayName())
			return request.Retry(err).With(recv2.ExponentialBackoff(exponentialBackoffMinDelay, exponentialBackoffMaxDelay))
		}

		projectConfig := controller.ProjectConfig{
			TenantID: string(rp.GetUID()),
		}

		ctx, cancel := context.WithTimeout(context.Background(), *configuration.InvTimeout)
		defer cancel()

		// initialize tenant on INV side
		if err := invTenantInitializationHandler.InitializeTenant(ctx, projectConfig); err != nil {
			log.InfraSec().Err(err).Msgf("Failed to initialize tenant[%s], requeing", rp.GetUID())
			return request.Retry(err).With(recv2.ExponentialBackoff(exponentialBackoffMinDelay, exponentialBackoffMaxDelay))
		}

		// all good!
		return request.Ack()
	}
}

func reconcileProjectTermination(nxc nexusClient, terminator tenantTerminationHandler, trace bool) recv2.Reconciler[ProjectID] {
	return func(ctx context.Context, request recv2.Request[ProjectID]) recv2.Directive[ProjectID] {
		if trace {
			ctx = tracing.StartTrace(ctx, "infra-tc", "ProjectTerminationReconciler")
			defer tracing.StopTrace(ctx)
		}
		rp, err := nxc.GetRuntimeProjectByUID(ctx, request.ID.String())
		if err != nil {
			log.Err(err).Msgf("rPROJECT[%s] does not exists, cannot reconcile project termination", request.ID)
			return request.Retry(err).With(recv2.ExponentialBackoff(exponentialBackoffMinDelay, exponentialBackoffMaxDelay))
		}
		if err = terminator.TerminateTenant(ctx, string(rp.GetUID())); err != nil {
			log.InfraSec().Err(err).Msgf("error occurred during rPROJECT[%s/%s] termination, requeing request %s",
				rp.GetUID(), rp.DisplayName(), request.ID)
			return request.Retry(err).With(recv2.ExponentialBackoff(exponentialBackoffMinDelay, exponentialBackoffMaxDelay))
		}
		if err := rp.DeleteActiveWatchers(ctx, configuration.AppName); err != nil {
			return request.Fail(err)
		}
		return request.Ack()
	}
}

type reconciler struct {
	initializer             *recv2.Controller[ProjectID]
	terminator              *recv2.Controller[ProjectID]
	termChan                chan struct{}
	nxc                     nexusClient
	projectsToBeInitialized chan *nexus_client.RuntimeprojectRuntimeProject
	projectsToBeTerminated  chan *nexus_client.RuntimeprojectRuntimeProject
}

func (s *reconciler) ControlLoop(termChan chan bool) {
	log.Info().Msgf("ProjectReconciliationController.ControlLoop is running")
	for {
		select {
		case rp, ok := <-s.projectsToBeTerminated:
			if !ok {
				// Note this will cover the sigterm scenario as well
				log.InfraSec().Fatal().Msg("data-model project creation notifications stream closed")
			}
			log.Debug().Msgf("Project to be terminated[%s/%s]", rp.GetUID(), rp.DisplayName())
			if err := s.terminator.Reconcile(NewProjectID(rp.GetUID())); err != nil {
				log.Err(err).Msg("")
			}
		case rp, ok := <-s.projectsToBeInitialized:
			if !ok {
				// Note this will cover the sigterm scenario as well
				log.InfraSec().Fatal().Msg("data-model project update/deletions notifications stream closed")
			}
			log.Debug().Msgf("Project to be initialized[%s/%s]", rp.GetUID(), rp.DisplayName())
			if err := s.initializer.Reconcile(NewProjectID(rp.GetUID())); err != nil {
				log.Err(err).Msgf("Cannot initialize project[%s/%s]", rp.GetUID(), rp.DisplayName())
				continue
			}
		case <-termChan:
			return
		}
	}
}
