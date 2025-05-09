// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	e "github.com/open-edge-platform/infra-core/bulk-import-tools/internal/errors"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/types"
)

const HEADER = "Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill"

func CreateFile(filePath string) error {
	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, no need to create it
		return nil
	} else if !os.IsNotExist(err) {
		// An error other than "not exists" occurred, return the error
		return err
	}

	// File does not exist, create the parent directory
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		return e.NewCustomError(e.ErrFileCreate)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return e.NewCustomError(e.ErrFileCreate)
	}
	defer file.Close()

	if err = writeHeaders(file); err != nil {
		return e.NewCustomError(e.ErrFileCreate)
	}

	return nil
}

func writeHeaders(file *os.File) error {
	writer := csv.NewWriter(file)
	defer writer.Flush() // Ensure any buffered data is written to the file

	// Write the CSV header (column titles)
	if err := writer.Write(strings.Split(HEADER, string(writer.Comma))); err != nil {
		return e.NewCustomError(e.ErrFileCreate)
	}

	return nil
}

//nolint:mnd // indices of fields are fixed in csv
func ReadHostRecords(filePath string) ([]types.HostRecord, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, e.NewCustomError(e.ErrFileRW)
	}
	defer file.Close() // Ensure the file is closed when the function returns

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header line
	if _, err := reader.Read(); err != nil {
		return nil, e.NewCustomError(e.ErrFileRW)
	}

	var records []types.HostRecord

	// Read each record from the CSV file
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			// Will still continue if there is a mismatch in the number of fields
			// as empty fields are allowed & they can be trailing as well
			if !strings.Contains(err.Error(), csv.ErrFieldCount.Error()) {
				return nil, e.NewCustomError(e.ErrFileRW)
			}
		}

		// Ensure the record has at least 8 fields
		for len(record) < 8 {
			record = append(record, "")
		}

		// Create a HostRecord from the CSV record
		hostRecord := types.HostRecord{
			Serial:     getField(record, 0),
			UUID:       getField(record, 1),
			OSProfile:  getField(record, 2),
			Site:       getField(record, 3),
			Secure:     types.StringToRecordSecure(getField(record, 4)),
			RemoteUser: getField(record, 5),
			Metadata:   getField(record, 6),
			Error:      getField(record, 7),
			RawRecord:  strings.Join(record, ","),
		}

		// Append the HostRecord to the slice
		records = append(records, hostRecord)
	}

	return records, nil
}

// getField safely retrieves a field from the record.
func getField(record []string, index int) string {
	if index < len(record) {
		return strings.TrimSpace(record[index])
	}
	return ""
}

func WriteHostRecords(filePath string, records []types.HostRecord) error {
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return e.NewCustomError(e.ErrFileCreate)
	}
	defer file.Close() // Ensure the file is closed when the function returns

	// Create a CSV writer
	writer := csv.NewWriter(file)

	defer writer.Flush() // Ensure any buffered data is written to the file

	// Write the header line
	if err := writer.Write(strings.Split(HEADER, string(writer.Comma))); err != nil {
		return e.NewCustomError(e.ErrFileRW)
	}

	// Write each HostRecord to the CSV file
	for _, record := range records {
		// Create a slice of fields from the HostRecord
		fields := []string{
			record.Serial,
			record.UUID,
			record.OSProfile,
			record.Site,
			string(record.Secure),
			record.RemoteUser,
			record.Metadata,
			record.Error,
		}

		// Write the fields to the CSV file
		if err := writer.Write(fields); err != nil {
			return e.NewCustomError(e.ErrFileRW)
		}
	}

	// Check for any errors that occurred during the write
	if err := writer.Error(); err != nil {
		return e.NewCustomError(e.ErrFileRW)
	}

	return nil
}
