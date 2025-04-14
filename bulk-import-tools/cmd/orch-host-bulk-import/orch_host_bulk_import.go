// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/info"
	e "github.com/open-edge-platform/infra-core/bulk-import-tools/internal/errors"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/files"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/orchcli"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/types"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/validator"
)

const (
	idxAfterFlags = 2
	numArgs       = 2
	importNumArgs = 3
)

func main() {
	// Check for subcommands
	if len(os.Args) < numArgs {
		displayHelp()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "import":
		handleImportCommand()
	case "help":
		displayHelp()
	case "version":
		fmt.Printf("Version %s\n\n", info.Version)
	default:
		fmt.Printf("error: Unknown command '%s'\n\n", os.Args[1])
		displayHelp()
		os.Exit(1)
	}
}

func handleImportCommand() {
	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	importCmd.Usage = displayHelp
	onboardFlag := importCmd.Bool("onboard", false, "")
	projectNameIn := importCmd.String("project", "", "")
	osProfileIn := importCmd.String("os-profile", "", "")
	siteIn := importCmd.String("site", "", "")
	secureFlag := importCmd.Bool("secure", false, "")
	remoteUserIn := importCmd.String("remote-user", "", "")
	metadataIn := importCmd.String("metadata", "", "")
	err := importCmd.Parse(os.Args[idxAfterFlags:])

	// Check for the correct number of arguments after flags
	if err != nil || importCmd.NArg() < numArgs-1 {
		fmt.Println("error: Filename & url required as arguments")
		displayHelp()
		os.Exit(1)
	}

	filePath := importCmd.Arg(0)
	serverURL := importCmd.Arg(1)

	// Check if project name is not provided, use the environment variable EDGEORCH_PROJECT
	projectName := *projectNameIn
	if projectName == "" {
		projectName = os.Getenv("EDGEORCH_PROJECT")
	}

	if projectName == "" {
		fmt.Println("error: Project name required as argument or set env variable EDGEORCH_PROJECT")
		displayHelp()
		os.Exit(1)
	}

	// Check if osprofile is not provided, use the environment variable EDGEORCH_OSPROFILE
	osProfile := *osProfileIn
	if osProfile == "" {
		osProfile = os.Getenv("EDGEORCH_OSPROFILE")
	}

	// Check if site is not provided, use the environment variable EDGEORCH_SITE
	site := *siteIn
	if site == "" {
		site = os.Getenv("EDGEORCH_SITE")
	}

	// Check if secure flag is not provided, use the environment variable EDGEORCH_SECURE
	secure := *secureFlag
	if !secure {
		secureEnv := os.Getenv("EDGEORCH_SECURE")
		if secureEnv == "true" {
			secure = true
		}
	}

	// Check if remote user is not provided, use the environment variable EDGEORCH_REMOTEUSER
	remoteUser := *remoteUserIn
	if remoteUser == "" {
		remoteUser = os.Getenv("EDGEORCH_REMOTEUSER")
	}

	// Check if metadata is not provided, use the environment variable EDGEORCH_METADATA
	metadata := *metadataIn
	if metadata == "" {
		metadata = os.Getenv("EDGEORCH_METADATA")
	}

	globalAttr := &types.HostRecord{
		OSProfile:  osProfile,
		Site:       site,
		Secure:     secure,
		RemoteUser: remoteUser,
		Metadata:   metadata,
	}

	fmt.Printf("Importing hosts from file: %s to server: %s\n", filePath, serverURL)

	// Implement the import functionality here
	if err := doImport(*onboardFlag, filePath, serverURL, projectName, globalAttr); err != nil {
		fmt.Printf("error: %v\n\n", err.Error())
		os.Exit(1)
	}
	fmt.Print("CSV import successful\n\n")
}

// displayHelp prints the help information for the utility.
func displayHelp() {
	fmt.Print("\n\nImport host data from input file into the Edge Orchestrator.\n\n")
	fmt.Print("Usage: orch-host-bulk-import COMMAND\n\n")
	fmt.Print("\nCOMMANDS:\n")
	fmt.Println("\timport [OPTIONS] <file> <url> Import data from given CSV file to orchestrator URL")
	fmt.Println("\t        file       Required source CSV file to read data from")
	fmt.Println("\t        url        Required Edge Orchestrator URL")
	fmt.Println("\tversion            Display version information")
	fmt.Print("\thelp               Show this help message\n")
	fmt.Println("OPTIONS:")
	fmt.Println("\t--onboard          If set, hosts will be automatically onboarded when connected")
	fmt.Println("\t--project <name>   Optional project name in Edge Orchestrator.",
		"Alternatively, set env variable EDGEORCH_PROJECT")
	fmt.Print("\t--os-profile <id>  Optional operating system profile name/id to configure for hosts.",
		"Alternatively, set env variable EDGEORCH_OSPROFILE\n\n")
}

func doImport(autoOnboard bool, filePath, serverURL, projectName string, globalAttr *types.HostRecord) error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	erringRecords := []types.HostRecord{}

	// Check if hosts expected to be onboarded or registered
	if autoOnboard {
		fmt.Println("Onboarding is enabled")
	}
	// validate input file
	validated, err := validator.CheckCSV(filePath)
	if err != nil {
		return err
	}

	oClient, err := orchcli.NewOrchCli(ctx, serverURL, projectName)
	if err != nil {
		return err
	}

	// registerHost
	// iterate over all entries available
	for _, record := range validated {
		doRegister(ctx, oClient, autoOnboard, globalAttr, record, &erringRecords)
	}
	// write import error to import_error_<rfc3339_timestamp>_<filename>
	// if there is any error record after header
	if len(erringRecords) > 0 {
		newFilename := fmt.Sprintf("%s_%s_%s", "import_error",
			time.Now().Format(time.RFC3339), filepath.Base(filePath))
		fmt.Printf("Generating error file: %s\n", newFilename)
		if err := files.WriteHostRecords(newFilename, erringRecords); err != nil {
			return e.NewCustomError(e.ErrFileRW)
		}
		return e.NewCustomError(e.ErrImportFailed)
	}
	return nil
}

func doRegister(ctx context.Context, oClient *orchcli.OrchCli, autoOnboard bool,
	globalAttr *types.HostRecord, rIn types.HostRecord, erringRecords *[]types.HostRecord,
) {
	// get the required fields from the record
	sNo := rIn.Serial
	uuid := rIn.UUID

	rOut, err := sanitizeProvisioningFields(ctx, oClient, rIn, erringRecords, globalAttr)
	if err != nil {
		return
	}

	// Register host
	hostID, err := oClient.RegisterHost(ctx, "", sNo, uuid, autoOnboard)
	if err != nil && !e.Is(e.ErrAlreadyRegistered, err) {
		// add to reject list if failed
		rIn.Error = err.Error()
		*erringRecords = append(*erringRecords, rIn)
		return
	}
	// Create instance if osProfileID is available else append to error list
	// Need not notify user of instance ID. Unnecessary detail for user.
	_, err = oClient.CreateInstance(ctx, hostID, rOut)
	if err != nil {
		rIn.Error = err.Error()
		*erringRecords = append(*erringRecords, rIn)
		return
	}

	if err := oClient.AllocateHostToSiteAndAddMetadata(ctx, hostID, rOut.Site, rOut.Metadata); err != nil {
		rIn.Error = err.Error()
		*erringRecords = append(*erringRecords, rIn)
		return
	}
	// Print host_id from response if successful
	fmt.Printf("✔ Host Serial number : %s  UUID : %s registered. Name : %s\n", sNo, uuid, hostID)
}

func sanitizeProvisioningFields(ctx context.Context, oClient *orchcli.OrchCli, record types.HostRecord,
	erringRecords *[]types.HostRecord, globalAttr *types.HostRecord,
) (*types.HostRecord, error) {
	var siteID, laID, osProfileID string
	// TODO - Can be a different case as absence of secure is similar to false
	isSecure := record.Secure
	if globalAttr.Secure != isSecure {
		isSecure = globalAttr.Secure
	}

	// If globalAttr.OSProfile is non-empty, use it; otherwise, use the record's OSProfile
	osProfileID = record.OSProfile
	if globalAttr.OSProfile != "" {
		osProfileID = globalAttr.OSProfile
	}

	var err error

	if osProfileID, err = oClient.GetOsProfileID(ctx, osProfileID); err != nil {
		record.Error = err.Error()
		*erringRecords = append(*erringRecords, record)
		return nil, err
	}

	// osProfile must be in cache as if the flow is here.
	// Check for security feature mismatch.
	osProfile, ok := oClient.OSProfileCache[osProfileID]
	if !ok || (*osProfile.SecurityFeature != api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION && isSecure) {
		record.Error = e.NewCustomError(e.ErrOSSecurityMismatch).Error()
		*erringRecords = append(*erringRecords, record)
		return nil, e.NewCustomError(e.ErrOSSecurityMismatch)
	}

	// If globalAttr.Site is non-empty, use it; otherwise, use the record's Site
	siteToQuery := record.Site
	if globalAttr.Site != "" {
		siteToQuery = globalAttr.Site
	}
	if siteID, err = oClient.GetSiteID(ctx, siteToQuery); err != nil {
		record.Error = err.Error()
		*erringRecords = append(*erringRecords, record)
		return nil, err
	}

	// If globalAttr.RemoteUser is non-empty, use it; otherwise, use the record's RemoteUser
	remoteUserToQuery := record.RemoteUser
	if globalAttr.RemoteUser != "" {
		remoteUserToQuery = globalAttr.RemoteUser
	}
	if laID, err = oClient.GetLocalAccountID(ctx, remoteUserToQuery); err != nil {
		record.Error = err.Error()
		*erringRecords = append(*erringRecords, record)
		return nil, err
	}

	// If globalAttr.Metadata is non-empty, use it; otherwise, use the record's Metadata
	metadataToUse := record.Metadata
	if globalAttr.Metadata != "" {
		metadataToUse = globalAttr.Metadata
	}

	return &types.HostRecord{
		OSProfile:  osProfileID,
		RemoteUser: laID,
		Site:       siteID,
		Secure:     isSecure,
		UUID:       record.UUID,
		Serial:     record.Serial,
		Metadata:   metadataToUse,
	}, nil
}
