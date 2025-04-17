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
	secureFlag := importCmd.String("secure", string(types.SecureUnspecified), "")
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
	secure := getGlobalSecureAttr(secureFlag)

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
		Secure:     types.StringToRecordSecure(secure),
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

func getGlobalSecureAttr(secureFlag *string) string {
	secure := *secureFlag
	if secure == "" {
		secureEnv := os.Getenv("EDGEORCH_SECURE")
		if secureEnv == "" {
			secure = string(types.SecureUnspecified)
		} else {
			secure = secureEnv
		}
	}
	return secure
}

// displayHelp prints the help information for the utility.
func displayHelp() {
	fmt.Print("\n\nImport host data from input file into the Edge Orchestrator.\n\n")
	fmt.Print("Usage: orch-host-bulk-import COMMAND\n\n")
	fmt.Print("COMMANDS:\n\n")
	fmt.Println("import [OPTIONS] <file> <url>  Import data from given CSV file to orchestrator URL")
	fmt.Println("        file                   Required source CSV file to read data from")
	fmt.Println("        url                    Required Edge Orchestrator URL")
	fmt.Println("version                        Display version information")
	fmt.Print("help                           Show this help message\n\n")
	fmt.Print("OPTIONS:\n\n")
	fmt.Println("--onboard                      Optional onboard flag.",
		"If set, hosts will be automatically onboarded when connected")
	fmt.Println("--project <name>               Required project name in Edge Orchestrator.",
		"Alternatively, set env variable EDGEORCH_PROJECT")
	fmt.Println("--os-profile <name/id>         Optional operating system profile name/id to configure for hosts.",
		"Alternatively, set env variable EDGEORCH_OSPROFILE")
	fmt.Println("--site <name/id>               Optional site name/id to configure for hosts.",
		"Alternatively, set env variable EDGEORCH_SITE")
	fmt.Println("--secure <value>               Optional security feature to configure for hosts.",
		"Alternatively, set env variable EDGEORCH_SECURE. Valid values: true, false")
	fmt.Println("--remote-user <name/id>        Optional remote user name/id to configure for hosts.",
		"Alternatively, set env variable EDGEORCH_REMOTEUSER")
	fmt.Print("--metadata <data>              Optional metadata to configure for hosts. ",
		"Alternatively, set env variable EDGEORCH_METADATA. Metadata format: key=value&key=value\n\n")
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
	isSecure := resolveSecure(record.Secure, globalAttr.Secure)
	osProfileID, err := resolveOSProfile(ctx, oClient, record.OSProfile, globalAttr.OSProfile, record, erringRecords)
	if err != nil {
		return nil, err
	}

	if valErr := validateSecurityFeature(oClient, osProfileID, isSecure, record, erringRecords); valErr != nil {
		return nil, valErr
	}

	siteID, err := resolveSite(ctx, oClient, record.Site, globalAttr.Site, record, erringRecords)
	if err != nil {
		return nil, err
	}

	laID, err := resolveRemoteUser(ctx, oClient, record.RemoteUser, globalAttr.RemoteUser, record, erringRecords)
	if err != nil {
		return nil, err
	}

	metadataToUse := resolveMetadata(record.Metadata, globalAttr.Metadata)

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

func resolveSecure(recordSecure, globalSecure types.RecordSecure) types.RecordSecure {
	if globalSecure != recordSecure && globalSecure != types.SecureUnspecified {
		return globalSecure
	}
	return recordSecure
}

func resolveOSProfile(ctx context.Context, oClient *orchcli.OrchCli, recordOSProfile, globalOSProfile string,
	record types.HostRecord, erringRecords *[]types.HostRecord,
) (string, error) {
	osProfileID := recordOSProfile
	if globalOSProfile != "" {
		osProfileID = globalOSProfile
	}

	osProfileID, err := oClient.GetOsProfileID(ctx, osProfileID)
	if err != nil {
		record.Error = err.Error()
		*erringRecords = append(*erringRecords, record)
		return "", err
	}
	return osProfileID, nil
}

func validateSecurityFeature(oClient *orchcli.OrchCli, osProfileID string, isSecure types.RecordSecure,
	record types.HostRecord, erringRecords *[]types.HostRecord,
) error {
	osProfile, ok := oClient.OSProfileCache[osProfileID]
	if !ok || (*osProfile.SecurityFeature != api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION && isSecure == types.SecureTrue) {
		record.Error = e.NewCustomError(e.ErrOSSecurityMismatch).Error()
		*erringRecords = append(*erringRecords, record)
		return e.NewCustomError(e.ErrOSSecurityMismatch)
	}
	return nil
}

func resolveSite(ctx context.Context, oClient *orchcli.OrchCli, recordSite, globalSite string,
	record types.HostRecord, erringRecords *[]types.HostRecord,
) (string, error) {
	siteToQuery := recordSite
	if globalSite != "" {
		siteToQuery = globalSite
	}

	siteID, err := oClient.GetSiteID(ctx, siteToQuery)
	if err != nil {
		record.Error = err.Error()
		*erringRecords = append(*erringRecords, record)
		return "", err
	}
	return siteID, nil
}

func resolveRemoteUser(ctx context.Context, oClient *orchcli.OrchCli, recordRemoteUser, globalRemoteUser string,
	record types.HostRecord, erringRecords *[]types.HostRecord,
) (string, error) {
	remoteUserToQuery := recordRemoteUser
	if globalRemoteUser != "" {
		remoteUserToQuery = globalRemoteUser
	}

	laID, err := oClient.GetLocalAccountID(ctx, remoteUserToQuery)
	if err != nil {
		record.Error = err.Error()
		*erringRecords = append(*erringRecords, record)
		return "", err
	}
	return laID, nil
}

func resolveMetadata(recordMetadata, globalMetadata string) string {
	if globalMetadata != "" {
		return globalMetadata
	}
	return recordMetadata
}
