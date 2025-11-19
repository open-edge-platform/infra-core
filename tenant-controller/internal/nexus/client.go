// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package nexus

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
	baseprojectactivewatcherinfrahostcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	projectwatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectwatcher.edge-orchestrator.intel.com/v1"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
)

var log = logging.GetLogger("tc-nexus")

const (
	// cacheSyncDelay is the time to wait for nexus client cache synchronization.
	cacheSyncDelay = 100 * time.Millisecond
)

// SetupClient creates a new Nexus client using k8s ServiceAccount, this works only for k8s deployment.
// TODO: extend to use local kubeconfig if not running in a pod.
func SetupClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	nxc, err := nexus_client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return NewClient(nxc), err
}

func NewClient(nxc *nexus_client.Clientset) *Client {
	return &Client{nxc}
}

type Client struct {
	*nexus_client.Clientset
}

func (c *Client) GetRuntimeProjectByUID(
	ctx context.Context, tenantID string,
) (*nexus_client.RuntimeprojectRuntimeProject, error) {
	time.Sleep(cacheSyncDelay) // wait for cache sync
	runtime, err := c.TenancyMultiTenancy().GetRuntime(ctx)
	if err != nil {
		log.Err(err).Msgf("Cannot get Runtime")
		return nil, err
	}

	orgIterator := runtime.GetAllOrgsIter(ctx)
	for org, err := orgIterator.Next(ctx); org != nil; org, err = orgIterator.Next(ctx) {
		if err != nil {
			log.Err(err).Msgf("Cannot get Org from runtime")
			return nil, errors.Errorfc(codes.NotFound, "rPROJECT[UID=%s] not found: %s", tenantID, err)
		}
		folderIterator := org.GetAllFoldersIter(ctx)
		for folder, err := folderIterator.Next(ctx); folder != nil; folder, err = folderIterator.Next(ctx) {
			if err != nil {
				log.Err(err).Msgf("Cannot get folder from org")
				return nil, errors.Errorfc(codes.NotFound, "rPROJECT[UID=%s] not found: %s", tenantID, err)
			}
			projectIterator := folder.GetAllProjectsIter(ctx)
			for project, err := projectIterator.Next(ctx); project != nil; project, err = projectIterator.Next(ctx) {
				if err != nil {
					log.Err(err).Msgf("Cannot get project from org")
					return nil, errors.Errorfc(codes.NotFound, "rPROJECT[UID=%s] not found: %s", tenantID, err)
				}
				if string(project.UID) == tenantID {
					return project, nil
				}
			}
		}
	}
	log.Err(err).Msgf("No org found in runtime")
	return nil, errors.Errorfc(codes.NotFound, "rPROJECT[UID=%s] not found", tenantID)
}

func (c *Client) TryToSetActiveWatcherStatusIdle(runtimeProjectID string) error {
	runtimeProject, err := c.GetRuntimeProjectByUID(context.Background(), runtimeProjectID)
	if err != nil {
		return errors.Errorfc(codes.NotFound, "Cannot get rPROJECT[%s]: %v", runtimeProjectID, err)
	}

	activeWatcher, err := runtimeProject.GetActiveWatchers(context.TODO(), configuration.AppName)
	if err != nil {
		return errors.Errorfc(codes.NotFound, "Cannot get rACTIVEWATCHER[%s] for rPROJECT[%s]: %v",
			configuration.AppName, runtimeProjectID, err)
	}

	if activeWatcher.Spec.StatusIndicator == baseprojectactivewatcherinfrahostcomv1.StatusIndicationIdle {
		log.Trace().Msgf("Tenant[%s] already initialized, nothing to do", runtimeProjectID)
		return nil
	}

	activeWatcher.Spec.StatusIndicator = baseprojectactivewatcherinfrahostcomv1.StatusIndicationIdle
	activeWatcher.Spec.Message = "Tenant on INFRA side successfully initialized"
	activeWatcher.Spec.TimeStamp = uint64(time.Now().Unix()) //nolint:gosec // disable G115
	if err := activeWatcher.Update(context.TODO()); err != nil {
		return errors.Errorf("Cannot update rACTIVEWATCHER[%s] for rPROJECT[%s]: %v",
			configuration.AppName, runtimeProjectID, err)
	}
	return nil
}

// RegisterActiveWatcher for given RuntimeProject(rp).
// Created watcher has IDLE status.
func (c *Client) RegisterActiveWatcher(
	rp *nexus_client.RuntimeprojectRuntimeProject,
) (*nexus_client.ProjectactivewatcherProjectActiveWatcher, error) {
	watcher, err := rp.AddActiveWatchers(context.Background(),
		&baseprojectactivewatcherinfrahostcomv1.ProjectActiveWatcher{
			ObjectMeta: CreateObjectMeta(configuration.AppName),
			Spec: baseprojectactivewatcherinfrahostcomv1.ProjectActiveWatcherSpec{
				StatusIndicator: baseprojectactivewatcherinfrahostcomv1.StatusIndicationInProgress,
				Message:         "Initializing Tenant on INFRA side",
				TimeStamp:       uint64(time.Now().Unix()), //nolint:gosec // disable G115
			},
		})

	if nexus_client.IsAlreadyExists(err) {
		log.Debug().Msgf("Watcher[%s] already exists for rPROJECT[%s/%s]",
			watcher.DisplayName(), rp.GetUID(), rp.DisplayName())
		return rp.GetActiveWatchers(context.Background(), configuration.AppName)
	} else if err != nil {
		log.InfraSec().Err(err).Msgf(
			"Cannot register active watcher for rPROJECT[%s/%s]",
			rp.GetUID(),
			rp.DisplayName(),
		)
		return nil, err
	}

	return watcher, nil
}

func (c *Client) ReportError(aw *nexus_client.ProjectactivewatcherProjectActiveWatcher, msg, tenantID string) error {
	aw.Spec.StatusIndicator = baseprojectactivewatcherinfrahostcomv1.StatusIndicationError
	aw.Spec.Message = msg
	aw.Spec.TimeStamp = uint64(time.Now().Unix()) //nolint:gosec // disable G115

	if err := aw.Update(context.TODO()); err != nil {
		return errors.Errorf("Cannot update rACTIVEWATCHER[%s] for rPROJECT[%s]: %v", aw.DisplayName(), tenantID, err)
	}
	return nil
}

func (c *Client) SetupWatcherConfig() error {
	tenancy := c.TenancyMultiTenancy()
	if tenancy == nil {
		return errors.Errorfc(codes.FailedPrecondition, "unexpected nexus client configuration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *configuration.DataModelTimeout)
	defer cancel()
	projWatcher, err := tenancy.Config().AddProjectWatchers(
		ctx,
		&projectwatcherv1.ProjectWatcher{
			ObjectMeta: CreateObjectMeta(configuration.AppName),
		},
	)
	if nexus_client.IsAlreadyExists(err) {
		log.Debug().Msgf("Project watcher already exist: PROJECTWATCHER[%s]", configuration.AppName)
	} else if err != nil {
		log.InfraSec().Err(err).Msgf("Failed to create project watcher PROJECTWATCHER[%s]", configuration.AppName)
		return err
	}
	log.Info().Msgf("Created project watcher: PROJECTWATCHER[%s]", projWatcher.DisplayName())
	return nil
}

// CreateObjectMeta - this has to be configurable since available test framework requires special ObjectMeta creation.
var CreateObjectMeta = func(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
	}
}
