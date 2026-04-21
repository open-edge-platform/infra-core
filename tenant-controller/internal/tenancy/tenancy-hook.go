// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package tenancy replaces the Nexus SDK subscription with a Tenant Manager
// REST polling approach using the shared orch-library tenancy poller.
package tenancy

import (
	"context"
	"fmt"
	"os"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/controller"
	orchtenancy "github.com/open-edge-platform/orch-library/go/pkg/tenancy"
)

var log = logging.GetLogger("tc-tenancy")

const (
	// appName is the canonical controller ID registered in the Tenant Manager.
	// Must match the name in the registered-controller config.
	appName = "tenant-controller"

	defaultTenantManagerURL = "http://tenancy-manager.orch-iam:8080"
)

// tenantInitializationHandler is satisfied by controller.TenantInitializationController.
type tenantInitializationHandler interface {
	InitializeTenant(ctx context.Context, config controller.ProjectConfig) error
}

// tenantTerminationHandler is satisfied by controller.TerminationController.
type tenantTerminationHandler interface {
	TerminateTenant(ctx context.Context, tenantID string) error
}

// TenancyHook replaces the former Nexus-based DataModelController.
// It subscribes to project lifecycle events from the Tenant Manager REST API
// via the shared orch-library tenancy poller.
type TenancyHook struct {
	cancel context.CancelFunc
}

// NewTenancyHook creates a TenancyHook.
func NewTenancyHook() *TenancyHook {
	return &TenancyHook{}
}

// Subscribe starts the tenancy poller in a background goroutine.
func (h *TenancyHook) Subscribe(
	initializer tenantInitializationHandler,
	terminator tenantTerminationHandler,
) error {
	tenantManagerURL := os.Getenv("TENANT_MANAGER_URL")
	if tenantManagerURL == "" {
		tenantManagerURL = defaultTenantManagerURL
	}

	handler := &infraHandler{
		initializer: initializer,
		terminator:  terminator,
	}

	poller, err := orchtenancy.NewPoller(tenantManagerURL, appName, handler,
		func(cfg *orchtenancy.PollerConfig) {
			cfg.OnError = func(err error, msg string) {
				log.Warn().Err(err).Msgf("tenancy poller: %s", msg)
			}
		},
	)
	if err != nil {
		return fmt.Errorf("create tenancy poller: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	go func() {
		if err := poller.Run(ctx); err != nil && ctx.Err() == nil {
			log.Error().Err(err).Msg("tenancy poller stopped unexpectedly")
		}
	}()

	log.Info().Msgf("Tenancy hook subscribed: controller=%s url=%s", appName, tenantManagerURL)
	return nil
}

// Unsubscribe cancels the background poller.
func (h *TenancyHook) Unsubscribe() {
	if h.cancel != nil {
		h.cancel()
	}
}

// infraHandler implements orchtenancy.Handler for the infra tenant-controller.
// On project creation it initializes tenant resources in Inventory.
// On project deletion it terminates tenant resources in Inventory.
// Org events require no action — infra-TC only manages project-scoped resources.
type infraHandler struct {
	initializer tenantInitializationHandler
	terminator  tenantTerminationHandler
}

func (h *infraHandler) HandleEvent(ctx context.Context, event orchtenancy.Event) error {
	// Only handle project events; org events require no action.
	if event.ResourceType != orchtenancy.ResourceTypeProject {
		return nil
	}

	tenantID := event.ResourceID.String()

	switch event.EventType {
	case orchtenancy.EventTypeCreated:
		log.Info().Msgf("project created: initializing tenant(%s) name=%s", tenantID, event.ResourceName)
		return h.initializer.InitializeTenant(ctx, controller.ProjectConfig{TenantID: tenantID})
	case orchtenancy.EventTypeDeleted:
		log.Info().Msgf("project deleted: terminating tenant(%s) name=%s", tenantID, event.ResourceName)
		return h.terminator.TerminateTenant(ctx, tenantID)
	default:
		log.Warn().Msgf("unknown event type %s for project %s", event.EventType, tenantID)
		return nil
	}
}
