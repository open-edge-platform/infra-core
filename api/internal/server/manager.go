// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/open-edge-platform/infra-core/api/internal/common"
	"github.com/open-edge-platform/infra-core/api/internal/types"
	api "github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

const (
	waitTimeout  = 10 * time.Second
	apiTraceName = "APIEchoServer"
)

var log = logging.GetLogger("server")

type Manager struct {
	echoServer *echo.Echo
	cfg        *common.GlobalConfig
	jobCh      chan types.Job
	Ready      chan bool
	Term       chan bool
	wg         *sync.WaitGroup
}

func NewManager(
	cfg *common.GlobalConfig,
	jobCh chan types.Job,
	ready, termChan chan bool,
	wg *sync.WaitGroup,
) (*Manager, error) {
	return &Manager{
		echoServer: echo.New(),
		cfg:        cfg,
		jobCh:      jobCh,
		Ready:      ready,
		Term:       termChan,
		wg:         wg,
	}, nil
}

func (m *Manager) Start() error {
	log.Info().Msg("Starting REST Manager")
	e := echo.New()

	m.setOptions(e)

	if m.cfg.RestServer.EnableMetrics {
		log.Info().Msgf("Metrics exporter is enabled")
		m.enableMetrics(e)
	}

	handlers, err := NewHandlers(m.jobCh, m.cfg.RestServer.Timeout)
	if err != nil {
		return err
	}

	log.Info().Str("baseUrl", m.cfg.RestServer.BaseURL).Msgf("Registering handlers")
	api.RegisterHandlersWithBaseURL(e, handlers, m.cfg.RestServer.BaseURL)
	log.Info().Str("address", m.cfg.RestServer.Address).Msgf("Starting REST server")

	m.echoServer = e
	m.wg.Add(1)
	go func() {
		err := e.Start(m.cfg.RestServer.Address)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("fatal error")
		}
	}()

	go func() {
		err := m.waitTerm()
		if err != nil {
			log.Fatal().Err(err).Msg("fatal error")
		}
	}()

	m.Ready <- true
	log.Info().Msg("Started REST Manager")
	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	return m.echoServer.Shutdown(ctx)
}

func (m *Manager) waitTerm() error {
	<-m.Term
	log.Info().Msg("Stopping REST Manager")
	defer m.wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	if err := m.Stop(ctx); err != nil {
		return err
	}
	log.Info().Msg("Stopped REST Manager")
	return nil
}
