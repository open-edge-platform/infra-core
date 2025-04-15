// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var bIBinPath = os.Getenv("BI_BIN_PATH")

func TestBinaryImportCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectOutput string
		expectErr    bool
	}{
		{
			name:         "import without project name",
			args:         []string{"import", "--onboard", "input.csv", "https://xyz.com"},
			expectOutput: "Project name required as argument or set env variable EDGEORCH_PROJECT",
			expectErr:    true,
		},
		{ // Should not complain about missing project name, Setting env variable in loop
			name:         "env",
			args:         []string{"import", "--onboard", "input.csv", "https://xyz.com"},
			expectOutput: "Importing hosts from file: input.csv to server: https://xyz.com\nOnboarding is enabled\n",
			expectErr:    true,
		},
		{
			name:         "import with invalid url",
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "https://xyz.com"},
			expectOutput: "Importing hosts from file: input.csv to server: https://xyz.com\nOnboarding is enabled\n",
			expectErr:    true,
		},
		{
			name: "import with all flags",
			args: []string{"import", "--onboard", "--project", "test", "--os-profile", "test", "--site",
				"test", "--secure", "test", "--remote-user", "test", "--metadata", "test", "input.csv", "https://xyz.com"},
			expectOutput: "Importing hosts from file: input.csv to server: https://xyz.com\nOnboarding is enabled\n",
			expectErr:    true,
		},
		{
			name: "import with all flags with equals sign",
			args: []string{"import", "--onboard", "--project=test", "--os-profile=test", "--site=test", "--secure=true",
				"--remote-user=test", "--metadata=test", "input.csv", "https://xyz.com"},
			expectOutput: "Importing hosts from file: input.csv to server: https://xyz.com\nOnboarding is enabled\n",
			expectErr:    true,
		},
		{
			name: "import with invalid flag",
			args: []string{"import", "--onboard", "--project", "test", "--osprofile", "test", "--site",
				"test", "--secure", "test", "--remote-user", "test", "--metadata", "test", "input.csv", "https://xyz.com"},
			expectOutput: "flag provided but not defined: -osprofile\n",
			expectErr:    true,
		},
		{
			name:         "help",
			args:         []string{"help"},
			expectOutput: "Usage",
			expectErr:    false,
		},
		{
			name:         "version",
			args:         []string{"version"},
			expectOutput: "Version",
			expectErr:    false,
		},
		{
			name:         "import with missing arguments",
			args:         []string{"import"},
			expectOutput: "error: Filename & url required as arguments",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "env" {
				os.Setenv("EDGEORCH_PROJECT", "test")
				defer os.Unsetenv("EDGEORCH_PROJECT")
			}
			// #nosec G204
			cmd := exec.Command(bIBinPath, tt.args...)
			outputBytes, err := cmd.CombinedOutput()
			output := string(outputBytes)

			if (err != nil) != tt.expectErr {
				t.Errorf("executeBulkImport() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !strings.Contains(output, tt.expectOutput) {
				t.Errorf("executeBulkImport() output = %v, should contain %v", output, tt.expectOutput)
			}
		})
	}
}
