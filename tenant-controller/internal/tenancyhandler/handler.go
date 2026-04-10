// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenancyhandler

import (
	"context"
	"fmt"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/controller"
	"github.com/open-edge-platform/orch-library/go/pkg/tenancy"
)

var log = logging.GetLogger("tenancy-handler")

// Handler implements tenancy.Handler for the infra tenant-controller.
// It only handles project events (created / deleted).
type Handler struct {
	initializer tenantInitializer
	terminator  tenantTerminator
}

type tenantInitializer interface {
	InitializeTenant(ctx context.Context, config controller.ProjectConfig) error
}

type tenantTerminator interface {
	TerminateTenant(ctx context.Context, tenantID string) error
}

// NewHandler creates a Handler wired to the existing initialization and
// termination controllers.
func NewHandler(initializer tenantInitializer, terminator tenantTerminator) *Handler {
	return &Handler{
		initializer: initializer,
		terminator:  terminator,
	}
}

// HandleEvent is called by the Poller for each tenancy event. The Poller
// manages status updates (in_progress / completed / error) automatically,
// so this method only needs to call the business logic and return an error
// on failure.
func (h *Handler) HandleEvent(ctx context.Context, event tenancy.Event) error {
	// This controller only handles project events.
	if event.ResourceType != "project" {
		log.Debug().Msgf("Ignoring non-project event: %s/%s", event.ResourceType, event.EventType)
		return nil
	}

	tenantID := event.ResourceID.String()

	switch event.EventType {
	case "created":
		log.Info().Msgf("Handling project created event for tenant(%s) project(%s)",
			tenantID, event.ResourceName)
		if err := h.initializer.InitializeTenant(ctx, controller.ProjectConfig{
			TenantID: tenantID,
		}); err != nil {
			log.Err(err).Msgf("Failed to initialize tenant(%s)", tenantID)
			return fmt.Errorf("initialize tenant(%s): %w", tenantID, err)
		}
		return nil

	case "deleted":
		log.Info().Msgf("Handling project deleted event for tenant(%s) project(%s)",
			tenantID, event.ResourceName)
		if err := h.terminator.TerminateTenant(ctx, tenantID); err != nil {
			log.Err(err).Msgf("Failed to terminate tenant(%s)", tenantID)
			return fmt.Errorf("terminate tenant(%s): %w", tenantID, err)
		}
		return nil

	default:
		log.Debug().Msgf("Ignoring unknown event type %q for tenant(%s)", event.EventType, tenantID)
		return nil
	}
}
