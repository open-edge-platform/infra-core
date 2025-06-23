// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package collections

import (
	"sort"
	"strconv"
	"strings"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/function"
)

// MapSlice converts each element of `a` by applying function `f`.
func MapSlice[T any, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

// FirstError executes function `f` for elements of `a` and returns first error reported by `f`.
func FirstError[T any](a []T, f func(T) error) error {
	for _, e := range a {
		r := f(e)
		if r != nil {
			return r
		}
	}

	return nil
}

func Filter[T any](c []T, predicate function.Predicate[T]) []T {
	var res []T
	for _, e := range c {
		if predicate(e) {
			res = append(res, e)
		}
	}
	return res
}

func ForEach[T any](c []T, f func(T)) {
	for _, e := range c {
		f(e)
	}
}

// ConcatMapValuesSorted takes a map of string keys and values, sorts the keys alphabetically,
// concatenates the corresponding non-empty values using the specified delimiter, and returns
// the resulting string wrapped in double quotes. If the map is empty or all values are empty,
// it returns an empty string.
func ConcatMapValuesSorted(m map[string]string, delimiter string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Sort keys alphabetically
	sort.Strings(keys)
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		if m[k] != "" {
			values = append(values, m[k])
		}
	}

	if len(values) == 0 {
		return ""
	}

	return strconv.Quote(strings.Join(values, delimiter))
}
