// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"testing"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/proxy"
)

func TestBuildAllowedClientList(t *testing.T) {

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
			scenarioName: "test-scenario-full",
			allowlist: map[string][]string{
				"test-scenario-full": {"HostService", "LocationService", "ProviderService"},
			},
			wantLen:      3,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService", "ProviderService"},
		},
		{
			name:         "unknown scenario",
			scenarioName: "test-scenario-unknown",
			allowlist: map[string][]string{
				"test-scenario-full": {"HostService"},
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:         "scenario with empty service names",
			scenarioName: "test-scenario-empty",
			allowlist: map[string][]string{
				"test-scenario-empty": {"HostService", "  ", "", "LocationService"},
			},
			wantLen:      2,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService"},
		},
		{
			name:         "scenario with no services",
			scenarioName: "test-scenario-minimal",
			allowlist: map[string][]string{
				"test-scenario-minimal": {},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:         "scenario with whitespace in service names",
			scenarioName: "test-scenario-whitespace",
			allowlist: map[string][]string{
				"test-scenario-whitespace": {"  HostService  ", "LocationService "},
			},
			wantLen:      2,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := proxy.BuildAllowedClientList(tt.scenarioName, tt.allowlist)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildAllowedClientList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Expected error case
			}

			if len(got) != tt.wantLen {
				t.Errorf("BuildAllowedClientList() got map length %d, want %d", len(got), tt.wantLen)
			}

			for _, service := range tt.wantServices {
				if _, exists := got[service]; !exists {
					t.Errorf("BuildAllowedClientList() missing service %q in result", service)
				}
			}
		})
	}
}

func TestBuildAllowedClientList_UnregisteredService(t *testing.T) {
	allowlist := map[string][]string{
		"test-scenario-unregistered": {"HostService", "UnknownService"},
	}

	list, err := proxy.BuildAllowedClientList("test-scenario-unregistered", allowlist)

	if err != nil {
		t.Errorf("BuildAllowedClientList() unexpected error: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("BuildAllowedClientList() got map length %d, want 1", len(list))
	}

	// Only known services should be in the allowed map
	if _, exists := list["HostService"]; !exists {
		t.Error("BuildAllowedClientList() missing HostService in result")
	}
	if _, exists := list["UnknownService"]; exists {
		t.Error("BuildAllowedClientList() should not include UnknownService in result")
	}
}
