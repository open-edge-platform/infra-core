// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package files_test

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/files"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/types"
)

func TestCreateFile(t *testing.T) {
	// Set NonRoot user to avoid permission overrides with root user
	currentUser := setNonRootUser(t)
	defer resetUser(t, currentUser)
	tempDir := t.TempDir()

	// Test cases
	tests := []struct {
		name      string
		filePath  string
		setup     func() // Optional setup function to run before the test
		expectErr bool
	}{
		{
			name:     "File already exists",
			filePath: filepath.Join(tempDir, "existing.csv"),
			setup: func() {
				// Create a file that already exists
				_, err := os.Create(filepath.Join(tempDir, "existing.csv"))
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			},
			expectErr: false,
		},
		{
			name:      "Create new file",
			filePath:  filepath.Join(tempDir, "subdir", "testfile.csv"),
			expectErr: false,
		},
		{
			name:      "Invalid file path",
			filePath:  "",
			expectErr: true,
		},
		{
			name:     "Permission denied",
			filePath: filepath.Join(tempDir, "subdir2", "testfile.csv"),
			setup: func() {
				// Create a directory with no write permission
				dirPath := filepath.Join(tempDir, "subdir2")
				err := os.Mkdir(dirPath, 0o444) // Read-only permissions
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := files.CreateFile(tt.filePath)

			if tt.expectErr {
				assert.Error(t, err, "CreateFile() should return an error")
			} else {
				assert.NoError(t, err, "CreateFile() should not return an error")
			}

			// If no error is expected, check if the file exists
			if !tt.expectErr {
				_, err := os.Stat(tt.filePath)
				assert.NoError(t, err, "CreateFile() should create the file")
			}
		})
	}
}

func TestReadHostRecords(t *testing.T) {
	// Set NonRoot user to avoid permission overrides with root user
	currentUser := setNonRootUser(t)
	defer resetUser(t, currentUser)

	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "testfile.csv")

	// Test cases
	tests := []struct {
		name      string
		setup     func() // Function to set up the test environment
		expectRet []types.HostRecord
		expectErr bool
	}{
		{
			name:      "File does not exist",
			expectRet: nil,
			expectErr: true,
		},
		{
			name: "Successfully read host records",
			setup: func() {
				// Create and write to the test file
				// [{"key":"cluster-name","value":"cl1"},{"key":"app-id","value":""}]
				content := []byte("Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error\n" +
					"1234,uuid-1234,profile1,site1,true,user1,cluster-name=test&app-id=testApp\n" +
					"5678,uuid-5678,profile2,site2,,user2,meta2")
				err := os.WriteFile(testFilePath, content, 0o600)
				assert.NoError(t, err)
			},
			expectRet: []types.HostRecord{
				{
					Serial:     "1234",
					UUID:       "uuid-1234",
					OSProfile:  "profile1",
					Site:       "site1",
					Secure:     types.SecureTrue,
					RemoteUser: "user1",
					Metadata:   "cluster-name=test&app-id=testApp",
					Error:      "",
					RawRecord:  "1234,uuid-1234,profile1,site1,true,user1,cluster-name=test&app-id=testApp,",
				},
				{
					Serial:     "5678",
					UUID:       "uuid-5678",
					OSProfile:  "profile2",
					Site:       "site2",
					Secure:     types.SecureUnspecified,
					RemoteUser: "user2",
					Metadata:   "meta2",
					Error:      "",
					RawRecord:  "5678,uuid-5678,profile2,site2,,user2,meta2,",
				},
			},
			expectErr: false,
		},
		{
			name: "Error during file reading",
			setup: func() {
				// Set the file permissions to cause a read error
				err := os.Chmod(testFilePath, 0o222) // Write-only permissions
				assert.NoError(t, err)
			},
			expectRet: nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			readRecords, err := files.ReadHostRecords(testFilePath)

			if tt.expectErr {
				assert.Error(t, err, "ReadHostRecords() should return an error")
			} else {
				assert.NoError(t, err, "ReadHostRecords() should not return an error")
			}

			// Check the content
			assert.Equal(t, tt.expectRet, readRecords, "ReadHostRecords() should return the expected content")
		})
	}
}

func TestWriteHostRecords(t *testing.T) {
	// Set NonRoot user to avoid permission overrides with root user
	currentUser := setNonRootUser(t)
	defer resetUser(t, currentUser)

	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "testfile.csv")

	// Test cases
	tests := []struct {
		name      string
		records   []types.HostRecord
		expectErr bool
		expectStr string
		setup     func() // Optional setup function to run before the test
	}{
		{
			name: "Error during file writing",
			records: []types.HostRecord{
				{
					Serial:     "1234",
					UUID:       "uuid-1234",
					OSProfile:  "profile1",
					Site:       "site1",
					Secure:     types.SecureTrue,
					RemoteUser: "user1",
					Metadata:   "meta1",
					Error:      "error1",
				},
			},
			expectErr: true,
			setup: func() {
				// Set the file permissions to cause a write error
				err := os.Chmod(tempDir, 0o555) // Read and execute permissions, no write permission
				assert.NoError(t, err)
			},
		},
		{
			name: "Successfully write host records",
			records: []types.HostRecord{
				{
					Serial:     "1234",
					UUID:       "uuid-1234",
					OSProfile:  "profile1",
					Site:       "site1",
					Secure:     types.SecureTrue,
					RemoteUser: "user1",
					Metadata:   "meta1",
					Error:      "error1",
				},
				{
					Serial:     "5678",
					UUID:       "uuid-5678",
					OSProfile:  "profile2",
					Site:       "site2",
					Secure:     types.SecureFalse,
					RemoteUser: "user2",
					Metadata:   "meta2",
					Error:      "error2",
				},
			},
			expectErr: false,
			expectStr: "Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill\n" +
				"1234,uuid-1234,profile1,site1,true,user1,meta1,error1\n" +
				"5678,uuid-5678,profile2,site2,false,user2,meta2,error2\n",
			setup: func() {
				err := os.Chmod(tempDir, 0o700) // Full permissions
				assert.NoError(t, err)
			},
		},
		{
			name:      "Error during file creation",
			records:   []types.HostRecord{},
			expectErr: true,
			setup: func() {
				// Provide an invalid file path to simulate a file creation error
				testFilePath = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := files.WriteHostRecords(testFilePath, tt.records)

			if tt.expectErr {
				assert.Error(t, err, "WriteHostRecords() should return an error")
			} else {
				assert.NoError(t, err, "WriteHostRecords() should not return an error")

				// Verify the file content if no error is expected
				content, readErr := os.ReadFile(testFilePath)
				assert.NoError(t, readErr, "Reading the written file should not produce an error")
				assert.Equal(t, tt.expectStr, string(content), "File content should match expected output")
			}
		})
	}
}

func setNonRootUser(t *testing.T) int {
	t.Helper()
	currenteUID := syscall.Geteuid()

	if syscall.Geteuid() < 1000 {
		err := syscall.Seteuid(1000)
		assert.Nil(t, err, fmt.Sprintf("Could not set non root user %v", err))
	}
	return currenteUID
}

func resetUser(t *testing.T, originalUID int) {
	t.Helper()
	if syscall.Geteuid() != originalUID {
		err := syscall.Seteuid(originalUID)
		assert.Nil(t, err, fmt.Sprintf("Could not reset user configuration %v", err))
	}
}
