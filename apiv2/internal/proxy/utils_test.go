// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"testing"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/proxy"
	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/scenario"
)

func TestBuildAllowedClientList(t *testing.T) {
	// Save original state
	originalAllowlist := scenario.Allowlist

	tests := []struct {
		name         string
		scenarioName string
		allowlist    map[string][]string
		wantLen      int
		wantErr      bool
		wantServices []string
	}{
		{
			name:         "valid scenario with services",
			scenarioName: "full",
			allowlist: map[string][]string{
				"full": {"HostService", "LocationService", "ProviderService"},
			},
			wantLen:      3,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService", "ProviderService"},
		},
		{
			name:         "unknown scenario",
			scenarioName: "unknown",
			allowlist: map[string][]string{
				"full": {"HostService"},
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:         "scenario with empty service names",
			scenarioName: "vpro",
			allowlist: map[string][]string{
				"vpro": {"HostService", "  ", "", "LocationService"},
			},
			wantLen:      2,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService"},
		},
		{
			name:         "scenario with no services",
			scenarioName: "minimal",
			allowlist: map[string][]string{
				"minimal": {},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:         "scenario with whitespace in service names",
			scenarioName: "test",
			allowlist: map[string][]string{
				"test": {"  HostService  ", "LocationService "},
			},
			wantLen:      2,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test scenario.Allowlist
			scenario.Allowlist = tt.allowlist

			got, err := proxy.BuildAllowedClientList(tt.scenarioName)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildAllowedClientList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Expected error case
			}

			if len(got) != tt.wantLen {
				t.Errorf("buildAllowedClientList() got map length %d, want %d", len(got), tt.wantLen)
			}

			for _, service := range tt.wantServices {
				if _, exists := got[service]; !exists {
					t.Errorf("buildAllowedClientList() missing service %q in result", service)
				}
			}
		})
	}

	// Restore original state
	scenario.Allowlist = originalAllowlist
}

func TestBuildAllowedClientList_UnregisteredService(t *testing.T) {
	originalAllowlist := scenario.Allowlist

	// Create test scenario
	scenario.Allowlist = map[string][]string{
		"test": {"HostService", "UnknownService"},
	}

	list, err := proxy.BuildAllowedClientList("test")

	if err != nil {
		t.Errorf("buildAllowedClientList() unexpected error: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("buildAllowedClientList() got map length %d, want 1", len(list))
	}

	// Only known services should be in the allowed map
	if _, exists := list["HostService"]; !exists {
		t.Error("buildAllowedClientList() missing HostService in result")
	}
	if _, exists := list["UnknownService"]; exists {
		t.Error("buildAllowedClientList() should not include UnknownService in result")
	}

	// Restore original state
	scenario.Allowlist = originalAllowlist
}
