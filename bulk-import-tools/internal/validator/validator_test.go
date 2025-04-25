// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/files"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/types"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/validator"
)

//nolint:funlen // Test function, so len does not matter
func TestSanitizeEntries(t *testing.T) {
	// Test Cases
	tests := []struct {
		name      string
		lines     []types.HostRecord
		expectErr bool
		expectStr []types.HostRecord
	}{
		{
			name: "Empty OSProfile field",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "Non-empty OSProfile field",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os1"},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: "os2"},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os1"},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: "os2"},
			},
		},
		{
			name: "Successfully validates content1",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "Successfully validates content2",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "Empty line",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "", UUID: "", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "", UUID: "", OSProfile: "", Error: "One of Serial number or UUID required;"},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "Empty line spaces",
			lines: []types.HostRecord{
				{Serial: " ", UUID: "", OSProfile: ""},
				{Serial: "", UUID: "", OSProfile: " "},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "", UUID: "", OSProfile: "", Error: "One of Serial number or UUID required;"},
				{Serial: "", UUID: "", OSProfile: "", Error: "One of Serial number or UUID required;"},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "Error column failure",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os-ubuntu", Error: "err"},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: true,
			expectStr: nil,
		},
		{
			name: "Serial number unavailable",
			lines: []types.HostRecord{
				{Serial: "", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{Serial: "", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "UUID unavailable",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: ""},
			},
		},
		{
			name: "SN UUID empty",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "", UUID: "", OSProfile: ""},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "", UUID: "", OSProfile: "", Error: "One of Serial number or UUID required;"},
			},
		},
		{
			name: "Invalid SN1",
			lines: []types.HostRecord{
				{Serial: "ABCD-123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{
					Serial: "ABCD-123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "",
					Error: "Invalid Serial number;",
				},
			},
		},
		{
			name: "Invalid SN2",
			lines: []types.HostRecord{
				{Serial: "ABCD", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "", Error: "Invalid Serial number;"},
			},
		},
		{
			name: "Invalid SN3",
			lines: []types.HostRecord{
				{Serial: "ABCD123ABCD123ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123ABCD123ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", Error: "Invalid Serial number;"},
			},
		},
		{
			name: "Invalid SN4",
			lines: []types.HostRecord{
				{Serial: "ABCD\t123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD\t123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", Error: "Invalid Serial number;"},
			},
		},
		{
			name: "Duplicate SN1",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "2c2c2c2c-0000-1111-2222-333333333333"},
				{Serial: "ABCD123", UUID: "1c1c1c1c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "2c2c2c2c-0000-1111-2222-333333333333", Error: "Duplicate Serial number : Row 2;"},
				{Serial: "ABCD123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", Error: "Duplicate Serial number : Row 1;"},
			},
		},
		{
			name: "Duplicate SN2",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "2c2c2c2c-0000-1111-2222-333333333333"},
				{Serial: "ABC*D123", UUID: "1c1c1c1c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "2c2c2c2c-0000-1111-2222-333333333333", Error: "Duplicate Serial number : Row 2;"},
				{Serial: "ABC*D123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", Error: "Invalid Serial number;"},
			},
		},
		{
			name: "Duplicate SN3 case",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "qwerty123", UUID: "2c2c2c2c-0000-1111-2222-333333333333"},
				{Serial: "AbcD123", UUID: "1c1c1c1c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "qwerty123", UUID: "2c2c2c2c-0000-1111-2222-333333333333", Error: "Duplicate Serial number : Row 2;"},
				{Serial: "AbcD123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", Error: "Duplicate Serial number : Row 1;"},
			},
		},
		{
			name: "Invalid UUID1",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c1-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c1-0000-1111-2222-333333333333", Error: "Invalid UUID;"},
			},
		},
		{
			name: "Invalid UUID2",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4-0000-1111-2222-333333333333", Error: "Invalid UUID;"},
			},
		},
		{
			name: "Invalid UUID3",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-11112222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-11112222-333333333333", Error: "Invalid UUID;"},
			},
		},
		{
			name: "Invalid UUID4",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "^4c4c4c4c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "^4c4c4c4c-0000-1111-2222-333333333333", Error: "Invalid UUID;"},
			},
		},
		{
			name: "Invalid UUID5",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111\t-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111\t-2222-333333333333", Error: "Invalid UUID;"},
			},
		},
		{
			name: "Duplicate UUID1",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "ABCD124", UUID: "4c4c4c4c-0000-1111-2222-444444444444"},
				{Serial: "ABCD125", UUID: "4c4c4c4c-0000-1111-22x22-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "ABCD124", UUID: "4c4c4c4c-0000-1111-2222-444444444444"},
				{Serial: "ABCD125", UUID: "4c4c4c4c-0000-1111-22x22-333333333333", Error: "Invalid UUID;"},
			},
		},
		{
			name: "Duplicate UUID2",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "ABCD124", UUID: "4c4c4c4c-0000-1111-2222-444444444444"},
				{Serial: "ABCD125", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "ABCD124", UUID: "4c4c4c4c-0000-1111-2222-444444444444"},
				{Serial: "ABCD125", UUID: "4c4c4c4c-0000-1111-2222-333333333333", Error: "Duplicate UUID : Row 1;"},
			},
		},
		{
			name: "Duplicate UUID3 case",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4a4a4a4a-0000-1111-2222-333333333333"},
				{Serial: "ABCD124", UUID: "4a4a4a4a-0000-1111-2222-444444444444"},
				{Serial: "ABCD125", UUID: "4A4A4A4A-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4a4a4a4a-0000-1111-2222-333333333333"},
				{Serial: "ABCD124", UUID: "4a4a4a4a-0000-1111-2222-444444444444"},
				{Serial: "ABCD125", UUID: "4A4A4A4A-0000-1111-2222-333333333333", Error: "Duplicate UUID : Row 1;"},
			},
		},
		{
			name: "Duplicate SN & UUID",
			lines: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "AbcD123", UUID: "1c1c1c1c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
				{Serial: "QWERTY123", UUID: "3c3c3c3c-0000-1111-2222-333333333333"},
				{
					Serial: "QWERTY123", UUID: "4c4c4c4c-0000-1111-2222-333333333333",
					Error: "Duplicate Serial number : Row 2;Duplicate UUID : Row 1;",
				},
				{
					Serial: "AbcD123", UUID: "1c1c1c1c-0000-1111-2222-333333333333",
					Error: "Duplicate Serial number : Row 1;",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := validator.SanitizeEntries(tt.lines)

			if tt.expectErr {
				assert.Error(t, err, "SanitizeEntries() should return an error")
			} else {
				assert.NoError(t, err, "SanitizeEntries() should not return an error")
			}
			assert.Equal(t, tt.expectStr, out, "File content should match expected output")
		})
	}
}

func TestCheckCSV(t *testing.T) {
	// Setup temporary directory for test files
	tmpDir := t.TempDir()

	// Test Cases
	tests := []struct {
		name         string
		content      []types.HostRecord
		expectErr    bool
		expectStr    []types.HostRecord
		expectErrStr string
	}{
		{
			name: "Valid CSV",
			content: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: ""},
				{Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: "os2"},
			},
			expectErr: false,
			expectStr: []types.HostRecord{
				{
					Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "",
					RawRecord: "ABCD123,4c4c4c4c-0000-1111-2222-333333333333,,,,,,",
				},
				{
					Serial: "QWERTY123", UUID: "1c1c1c1c-0000-1111-2222-333333333333", OSProfile: "os2",
					RawRecord: "QWERTY123,1c1c1c1c-0000-1111-2222-333333333333,os2,,,,,",
				},
			},
		},
		{
			name: "Invalid Serial Number",
			content: []types.HostRecord{
				{Serial: "ABCD-123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os1"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{
					Serial: "ABCD-123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os1",
					Error: "Invalid Serial number;", RawRecord: "ABCD-123,4c4c4c4c-0000-1111-2222-333333333333,os1,,,,,",
				},
			},
			expectErrStr: "Pre-flight check failed",
		},
		{
			name: "Duplicate UUID",
			content: []types.HostRecord{
				{Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os1"},
				{Serial: "QWERTY123", UUID: "4c4c4c4c-0000-1111-2222-333333333333"},
			},
			expectErr: true,
			expectStr: []types.HostRecord{
				{
					Serial: "ABCD123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", OSProfile: "os1",
					RawRecord: "ABCD123,4c4c4c4c-0000-1111-2222-333333333333,os1,,,,,",
				},
				{
					Serial: "QWERTY123", UUID: "4c4c4c4c-0000-1111-2222-333333333333", Error: "Duplicate UUID : Row 1;",
					RawRecord: "QWERTY123,4c4c4c4c-0000-1111-2222-333333333333,,,,,,",
				},
			},
			expectErrStr: "Pre-flight check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary CSV file
			tmpFile := filepath.Join(tmpDir, fmt.Sprintf("%s.csv", tt.name))
			err := files.WriteHostRecords(tmpFile, tt.content)
			assert.NoError(t, err, "Failed to write temporary CSV file")

			// Run CheckCSV
			out, err := validator.CheckCSV(tmpFile)

			if tt.expectErr {
				assert.Error(t, err, "CheckCSV() should return an error")
				assert.Contains(t, err.Error(), tt.expectErrStr, "Error message should contain expected string")
			} else {
				assert.NoError(t, err, "CheckCSV() should not return an error")
			}
			assert.Equal(t, tt.expectStr, out, "File content should match expected output")

			// Check if error file is generated
			if tt.expectErr {
				errorFiles, err := filepath.Glob("preflight_error*")
				assert.NoError(t, err, "Failed to list error files")
				assert.NotEmpty(t, errorFiles, "Error file should be generated")

				// Delete error files
				for _, file := range errorFiles {
					err := os.Remove(file)
					assert.NoError(t, err, "Failed to delete error file")
				}
			}
		})
	}
}
