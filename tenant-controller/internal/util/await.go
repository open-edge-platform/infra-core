// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util

import "context"

type Task[T any] func(context.Context) (T, error)

type response[T any] struct {
	value T
	err   error
}

type Promise[T any] struct {
	responses chan response[T]
	ctx       context.Context
	cancel    context.CancelFunc
}

func (t *Promise[T]) Await() (T, error) {
	for {
		select {
		case <-t.ctx.Done():
			return *new(T), t.ctx.Err()
		case rsp := <-t.responses:
			return rsp.value, rsp.err
		}
	}
}

func (t *Promise[T]) Cancel() {
	t.cancel()
}

func Run[T any](ctx context.Context, f Task[T]) *Promise[T] {
	ctx, cancel := context.WithCancel(ctx)
	responses := make(chan response[T], 16) //nolint:mnd // default size of buffer
	go func() {
		v, e := f(ctx)
		responses <- response[T]{
			value: v,
			err:   e,
		}
	}()
	return &Promise[T]{
		responses: responses,
		ctx:       ctx,
		cancel:    cancel,
	}
}
