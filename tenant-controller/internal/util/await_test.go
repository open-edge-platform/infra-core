// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/tenant-controller/internal/util"
)

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	canceled := new(atomic.Bool)
	promise := util.Run(ctx, func(c context.Context) (any, error) {
		<-c.Done()
		canceled.Store(true)
		return new(interface{}), nil
	})

	promise.Cancel()
	require.Eventually(t, func() bool { return canceled.Load() },
		time.Second, 100*time.Millisecond, "running task shall be canceled")
}

func TestCancelWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	promise := util.Run(ctx, func(c context.Context) (any, error) {
		<-c.Done()
		return new(interface{}), nil
	})

	go time.AfterFunc(time.Second, func() {
		cancel()
	})
	_, err := promise.Await()
	require.ErrorIs(t, err, context.Canceled)
}
