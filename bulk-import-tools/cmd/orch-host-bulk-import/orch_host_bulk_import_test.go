// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	e "github.com/open-edge-platform/infra-core/bulk-import-tools/internal/errors"
)

var (
	bIBinPath = os.Getenv("BI_BIN_PATH")
	inputFile = "input.csv"
)

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
			args: []string{
				"import", "--onboard", "--project", "test", "--os-profile", "test", "--site",
				"test", "--secure", "test", "--remote-user", "test", "--metadata", "test", "input.csv", "https://xyz.com",
			},
			expectOutput: "Importing hosts from file: input.csv to server: https://xyz.com\nOnboarding is enabled\n",
			expectErr:    true,
		},
		{
			name: "import with all flags with equals sign",
			args: []string{
				"import", "--onboard", "--project=test", "--os-profile=test", "--site=test", "--secure=true",
				"--remote-user=test", "--metadata=test", "input.csv", "https://xyz.com",
			},
			expectOutput: "Importing hosts from file: input.csv to server: https://xyz.com\nOnboarding is enabled\n",
			expectErr:    true,
		},
		{
			name: "import with invalid flag",
			args: []string{
				"import", "--onboard", "--project", "test", "--osprofile", "test", "--site",
				"test", "--secure", "test", "--remote-user", "test", "--metadata", "test", "input.csv", "https://xyz.com",
			},
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

//nolint:cyclop,funlen // Test function is long due to multiple test cases
func TestEndToEndWithMockServer(t *testing.T) {
	mockServer := startMockServer(t, false, false)
	defer mockServer.Close()

	destFile := inputFile
	err := os.WriteFile(destFile, []byte(`Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill
JQWRQR3,4c4c4533-0046-5310-8052-cac04f515233,ubuntu-22.04-lts-generic,Folsom,true,testaccount,cluster=test&group=hr
JQWRQR4,4c4c4533-0046-5310-8052-cac04f515234,ubuntu-20.04-lts-generic-ext,Folsom,false,testaccount,cluster=dev`), 0o600)
	if err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}

	defer os.Remove(destFile)
	tests := []struct {
		name         string
		env          map[string]string
		args         []string
		expectOutput string
		csvError     string
		expectErr    bool
	}{
		{
			name:         "import successfully",
			env:          map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test"},
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "CSV import successful",
			csvError:     "",
			expectErr:    false,
		},
		{
			name:         "import successfully project from env",
			env:          map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_PROJECT": "test"},
			args:         []string{"import", "--onboard", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "CSV import successful",
			csvError:     "",
			expectErr:    false,
		},
		{
			// os-profile in env var overrides all csv row entries
			name: "os profile not found env",
			env: map[string]string{
				"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test",
				"EDGEORCH_OSPROFILE": "microvisor",
			},
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidOSProfile).Error(),
			expectErr:    true,
		},
		{
			// os-profile in options overrides all csv row entries
			name: "os profile not found option",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test"},
			args: []string{
				"import", "--onboard", "--project", "test", "--os-profile", "microvisor",
				"input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidOSProfile).Error(),
			expectErr:    true,
		},
		{
			// if both supplied, option has precedence. env var is ignored which is good
			name: "os profile not found option takes precedence",
			env: map[string]string{
				"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test",
				"EDGEORCH_OSPROFILE": "ubuntu-22.04-lts-generic",
			},
			args: []string{
				"import", "--onboard", "--project", "test", "--os-profile", "microvisor",
				"input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidOSProfile).Error(),
			expectErr:    true,
		},
		{
			name:         "security feature mismatch env",
			env:          map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_SECURE": "true"},
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrOSSecurityMismatch).Error(),
			expectErr:    true,
		},
		{
			name: "security feature mismatch option",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test"},
			args: []string{
				"import", "--onboard", "--project", "test", "--secure", "true", "input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrOSSecurityMismatch).Error(),
			expectErr:    true,
		},
		{
			name: "security feature mismatch option takes precedence",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_SECURE": "true"},
			args: []string{
				"import", "--onboard", "--project", "test", "--secure", "true", "input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrOSSecurityMismatch).Error(),
			expectErr:    true,
		},
		{
			name:         "site not found env",
			env:          map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_SITE": "Honolulu"},
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidSite).Error(),
			expectErr:    true,
		},
		{
			name: "site not found option",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test"},
			args: []string{
				"import", "--onboard", "--project", "test", "--site", "Honolulu", "input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidSite).Error(),
			expectErr:    true,
		},
		{
			name: "site not found option takes precedence",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_SITE": "Folsom"},
			args: []string{
				"import", "--onboard", "--project", "test", "--site", "Honolulu", "input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidSite).Error(),
			expectErr:    true,
		},
		{
			name: "localaccount not found env",
			env: map[string]string{
				"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_REMOTEUSER": "anonymous",
			},
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidLocalAccount).Error(),
			expectErr:    true,
		},
		{
			name: "localaccount not found option",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test"},
			args: []string{
				"import", "--onboard", "--project", "test", "--remote-user", "anonymous", "input.csv",
				"http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidLocalAccount).Error(),
			expectErr:    true,
		},
		{
			name: "localaccount not found option takes precedence",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_REMOTEUSER": "testaccount"},
			args: []string{
				"import", "--onboard", "--project", "test", "--remote-user", "anonymous",
				"input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidLocalAccount).Error(),
			expectErr:    true,
		},
		{
			name: "metadata env",
			env: map[string]string{
				"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_METADATA": "cluster=test&grouphr",
			},
			args:         []string{"import", "--onboard", "--project", "test", "input.csv", "http://" + mockServer.Addr},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidMetadata).Error(),
			expectErr:    true,
		},
		{
			name: "metadata option",
			env:  map[string]string{"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test"},
			args: []string{
				"import", "--onboard", "--project", "test", "--metadata", "cluster=test&grouphr",
				"input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidMetadata).Error(),
			expectErr:    true,
		},
		{
			name: "metadata option takes precedence",
			env: map[string]string{
				"EDGEORCH_USER": "test", "EDGEORCH_PASSWORD": "test", "EDGEORCH_METADATA": "cluster=test",
			},
			args: []string{
				"import", "--onboard", "--project", "test", "--metadata", "cluster=test&grouphr",
				"input.csv", "http://" + mockServer.Addr,
			},
			expectOutput: "error: Failed to import all hosts",
			csvError:     e.NewCustomError(e.ErrInvalidMetadata).Error(),
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.env {
				t.Setenv(key, value)
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
			if tt.csvError != "" {
				// Extract the filename from the output
				var errorFileName string
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "Generating error file:") {
						parts := strings.Split(line, ": ")
						if len(parts) == 2 {
							errorFileName = strings.TrimSpace(parts[1])
							break
						}
					}
				}

				if errorFileName == "" {
					t.Fatalf("Error file not created")
				}

				// Read the content of the error file
				defer os.Remove(errorFileName)
				content, err := os.ReadFile(errorFileName)
				if err != nil {
					t.Fatalf("Failed to read error file %s: %v", errorFileName, err)
				}

				// Check if the content contains the expected error
				if !strings.Contains(string(content), tt.csvError) {
					t.Errorf("Error file %s should contain %v, but it does not", errorFileName, tt.csvError)
				}
			}
		})
	}
}

func TestReentrancyAlreadyExists(t *testing.T) {
	mockServer := startMockServer(t, true, true)
	defer mockServer.Close()

	destFile := inputFile
	err := os.WriteFile(destFile, []byte(`Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill
JQWRQR3,4c4c4533-0046-5310-8052-cac04f515233,ubuntu-22.04-lts-generic,Folsom,true,testaccount,cluster=test&group=hr`), 0o600)
	if err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}
	defer os.Remove(destFile)

	t.Setenv("EDGEORCH_USER", "test")
	t.Setenv("EDGEORCH_PASSWORD", "test")

	// #nosec G204
	cmd := exec.Command(bIBinPath, "import", "--onboard", "--project", "test", "input.csv", "http://"+mockServer.Addr)
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	assert.Error(t, err)
	if !strings.Contains(output, "error: Failed to import all hosts") {
		t.Errorf("executeBulkImport() output = %v, should contain %v", output, "error: Failed to import all hosts")
	}

	// Extract the filename from the output
	var errorFileName string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Generating error file:") {
			parts := strings.Split(line, ": ")
			if len(parts) == 2 {
				errorFileName = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	if errorFileName == "" {
		t.Fatalf("Error file not created")
	}

	// Read the content of the error file
	defer os.Remove(errorFileName)
	content, err := os.ReadFile(errorFileName)
	if err != nil {
		t.Fatalf("Failed to read error file %s: %v", errorFileName, err)
	}

	// Check if the content contains the expected error
	if !strings.Contains(string(content), e.NewCustomError(e.ErrAlreadyRegistered).Error()) {
		t.Errorf("Error file %s should contain %v, but it does not", errorFileName,
			e.NewCustomError(e.ErrAlreadyRegistered).Error())
	}
}

func TestReentrancyPass(t *testing.T) {
	mockServer := startMockServer(t, true, false)
	defer mockServer.Close()

	destFile := inputFile
	err := os.WriteFile(destFile, []byte(`Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill
JQWRQR3,4c4c4533-0046-5310-8052-cac04f515233,ubuntu-22.04-lts-generic,Folsom,true,testaccount,cluster=test&group=hr`), 0o600)
	if err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}
	defer os.Remove(destFile)

	t.Setenv("EDGEORCH_USER", "test")
	t.Setenv("EDGEORCH_PASSWORD", "test")

	// #nosec G204
	cmd := exec.Command(bIBinPath, "import", "--onboard", "--project", "test", "input.csv", "http://"+mockServer.Addr)
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	assert.NoError(t, err)
	if !strings.Contains(output, "CSV import successful") {
		t.Errorf("executeBulkImport() output = %v, should contain %v", output, "CSV import successful")
	}
}

func startMockServer(t *testing.T, alreadyRegistered, instanceExists bool) *http.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/projects/test/compute/hosts/register", func(w http.ResponseWriter, _ *http.Request) {
		if alreadyRegistered {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"resourceId": "host-12345678",
			"status": "success",
			"message": "Host registered successfully"
		}`))
	})
	mux.HandleFunc("/v1/projects/test/compute/hosts", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"totalElements": 1,
			"hosts": [
				{
					"serialNumber": "JQWRQR3",
					"uuid": "4c4c4533-0046-5310-8052-cac04f515233",
					"resourceId": "host-12345678"
				}
			]
		}`))
	})
	mux.HandleFunc("/v1/projects/test/compute/instances", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			if instanceExists {
				w.Write([]byte(`{ "totalElements": 1 }`))
			} else {
				w.Write([]byte(`{ "totalElements": 0 }`))
			}
		} else {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{ "resourceId": "inst-12345678" }`))
		}
	})
	mux.HandleFunc("/v1/projects/test/compute/os", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
        "operatingSystemResources": [
            {
                "resourceId": "os-12345678",
                "profileName": "ubuntu-22.04-lts-generic",
                "securityFeature": "SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION"
            },
            {
                "resourceId": "os-87654321",
                "profileName": "ubuntu-20.04-lts-generic-ext",
                "securityFeature": "SECURITY_FEATURE_NONE"
            }
        ]
    }`))
	})
	mux.HandleFunc("/v1/projects/test/regions/regionID/sites", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"sites": [
				{
					"resourceId": "site-12345678",
					"name": "Folsom"
				},
				{
					"resourceId": "site-87654321",
					"name": "SanJose"
				}
			]
		}`))
	})
	mux.HandleFunc("/v1/projects/test/localAccounts", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"localAccounts": [
				{
					"resourceId": "account123",
					"username": "testaccount"
				},
				{
					"resourceId": "account124",
					"username": "admin"
				}
			]
		}`))
	})
	mux.HandleFunc("/v1/projects/test/compute/hosts/host-12345678", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"allocated"}`))
	})
	mux.HandleFunc("/realms/master/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"test-token"}`))
	})

	return createStartServerWith(t, mux)
}

func createStartServerWith(t *testing.T, mux *http.ServeMux) *http.Server {
	t.Helper()
	server := &http.Server{
		Addr:              "api.test.svc:9090",
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		assert.Equal(t, err, http.ErrServerClosed)
	}()

	return server
}
