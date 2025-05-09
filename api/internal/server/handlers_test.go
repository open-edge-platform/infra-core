// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/api/internal/common"
	"github.com/open-edge-platform/infra-core/api/internal/dispatcher"
	"github.com/open-edge-platform/infra-core/api/internal/server"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
)

func TestHandlers(t *testing.T) {
	cfg := common.DefaultConfig()
	cfg.Inventory.Address = localhostAddress
	cfg.RestServer.Timeout = 1 * time.Second
	assert.NotEqual(t, cfg, nil)
	dispChan := make(chan bool, 1)
	termChan := make(chan bool, 1)
	wg := sync.WaitGroup{}
	disp := dispatcher.NewDispatcher(cfg, dispChan, termChan, &wg)
	assert.NotEqual(t, disp, nil)
	restMgrChan := make(chan bool, 1)
	man, err := server.NewManager(cfg, disp.JobQueue, restMgrChan, termChan, &wg)
	assert.NotEqual(t, man, nil)
	assert.NoError(t, err)

	h, err := server.NewHandlers(disp.JobQueue, cfg.RestServer.Timeout)
	require.NoError(t, err)
	require.NotNil(t, h)

	siteName := "site1"
	body := api.Site{Name: &siteName}
	req, err := api.NewPostSitesRequest("http://localhost:8080/edge-infra.orchestrator.apis/v1", body)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	respbody := new(api.Host)
	err = ctx.Bind(respbody)
	assert.NoError(t, err)

	err = h.PostComputeHosts(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusRequestTimeout, rec.Code)
}

func TestHandlersBody(t *testing.T) {
	cfg := common.DefaultConfig()
	cfg.Inventory.Address = localhostAddress
	cfg.RestServer.Timeout = 1 * time.Second
	assert.NotEqual(t, cfg, nil)
	dispChan := make(chan bool, 1)
	termChan := make(chan bool, 1)
	wg := sync.WaitGroup{}
	disp := dispatcher.NewDispatcher(cfg, dispChan, termChan, &wg)
	assert.NotEqual(t, disp, nil)
	restMgrChan := make(chan bool, 1)
	man, err := server.NewManager(cfg, disp.JobQueue, restMgrChan, termChan, &wg)
	assert.NotEqual(t, man, nil)
	assert.NoError(t, err)

	h, err := server.NewHandlers(disp.JobQueue, cfg.RestServer.Timeout)
	require.NoError(t, err)
	require.NotNil(t, h)

	siteName := "site2"
	body := api.Site{Name: &siteName}
	req, err := api.NewPostSitesRequest("http://localhost:8080/edge-infra.orchestrator.apis/v1", body)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	err = h.PostSites(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusRequestTimeout, rec.Code)
}
