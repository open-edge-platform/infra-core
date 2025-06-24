// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package collections

import (
	"testing"
)

func TestConcatMapValuesSorted(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "NilMap",
			input:    nil,
			expected: "",
		},
		{
			name:     "EmptyMap",
			input:    map[string]string{},
			expected: "",
		},
		{
			name: "SingleKeyValue",
			input: map[string]string{
				"a": "foo",
			},
			expected: "foo",
		},
		{
			name: "SingleKeyEmptyValue",
			input: map[string]string{
				"a": "",
			},
			expected: "",
		},
		{
			name: "MultipleKeysSorted",
			input: map[string]string{
				"b": "bar",
				"a": "foo",
				"c": "baz",
			},
			expected: "foo\x1fbar\x1fbaz",
		},
		{
			name: "KeysWithEmptyValue",
			input: map[string]string{
				"a": "",
				"b": "bar",
			},
			expected: "bar",
		},
		{
			name: "KeysWithInterspersedEmptyValue",
			input: map[string]string{
				"a": "foo",
				"b": "",
				"c": "baz",
			},
			expected: "foo\x1fbaz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConcatMapValuesSorted(tt.input, "\x1f")
			if got != tt.expected {
				t.Errorf("ConcatMapValuesSorted() = %q, want %q", got, tt.expected)
			}
		})
	}
}
