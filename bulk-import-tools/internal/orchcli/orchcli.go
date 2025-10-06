// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package orchcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	u "github.com/google/uuid"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/authn"
	e "github.com/open-edge-platform/infra-core/bulk-import-tools/internal/errors"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/types"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/validator"
)

const kVSize = 2

type OrchCli struct {
	SvcURL         *url.URL
	Project        string
	Jwt            string
	OSProfileCache map[string]api.OperatingSystemResource
	SiteCache      map[string]api.Site
	LACache        map[string]api.LocalAccount
	HostCache      map[string]api.Host
}

type MetadataItem = struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewOrchCli(ctx context.Context, svcURL, project string) (*OrchCli, error) {
	// Parse the input URL string into a url.URL object.
	uParsed, err := url.Parse(svcURL)
	if err != nil {
		return &OrchCli{}, e.NewCustomError(e.ErrURL)
	}
	keycloakURL := *uParsed
	//nolint:mnd // split required into 2 parts to get service name and sub-domain
	keycloakURL.Host = "keycloak" + strings.TrimPrefix(keycloakURL.Host, strings.SplitN(keycloakURL.Host, ".", 2)[0])
	fmt.Println("url is:")
	fmt.Println(keycloakURL)
	// get credentials & authenticate
	jwt, err := authn.Authenticate(ctx, &keycloakURL)
	if err != nil {
		return &OrchCli{}, e.NewCustomError(e.ErrAuthNFailed)
	}
	return &OrchCli{
		SvcURL:         uParsed,
		Project:        project,
		Jwt:            jwt,
		OSProfileCache: make(map[string]api.OperatingSystemResource),
		SiteCache:      make(map[string]api.Site),
		LACache:        make(map[string]api.LocalAccount),
		HostCache:      make(map[string]api.Host),
	}, nil
}

// errors are customized here to record in error log.
func (oC *OrchCli) RegisterHost(ctx context.Context, host, sNo, uuid string, autoOnboard bool) (string, error) {
	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/hosts/register", oC.Project))

	fmt.Printf("DEBUG RegisterHost: URL: %s\n", uParsed.String())
	fmt.Printf("DEBUG RegisterHost: Input - host: %s, sNo: %s, uuid: %s, autoOnboard: %v\n", host, sNo, uuid, autoOnboard)

	// Prepare the form data
	payload := &api.HostRegisterInfo{
		Name:        &host,
		AutoOnboard: &autoOnboard,
	}

	if sNo != "" {
		payload.SerialNumber = &sNo
		fmt.Printf("DEBUG RegisterHost: Added SerialNumber: %s\n", sNo)
	}
	if uuid != "" {
		uObj := u.MustParse(uuid)
		payload.Uuid = &uObj
		fmt.Printf("DEBUG RegisterHost: Added UUID: %s\n", uuid)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("DEBUG RegisterHost: JSON Marshal failed: %v\n", err)
		return "", e.NewCustomError(e.ErrInternal)
	}

	fmt.Printf("DEBUG RegisterHost: Request payload: %s\n", string(jsonData))

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodPost, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("DEBUG RegisterHost: HTTP request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	
	fmt.Printf("DEBUG RegisterHost: Response Status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG RegisterHost: Response Body: %s\n", bodyString)

	if resp.StatusCode == http.StatusPreconditionFailed {
		fmt.Printf("DEBUG RegisterHost: Host already registered (412)\n")
		return "", e.NewCustomError(e.ErrAlreadyRegistered)
	}

	// Accept both 200 and 201 as success
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		fmt.Printf("DEBUG RegisterHost: Registration failed with status %d\n", resp.StatusCode)
		return "", e.NewCustomError(e.ErrRegisterFailed)
	}

	// Use generic JSON parsing to avoid struct mismatch issues
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Printf("DEBUG RegisterHost: JSON unmarshal failed: %v\n", err)
		return "", e.NewCustomError(e.ErrInternal)
	}

	// Extract resourceId directly from the map
	resourceId, ok := response["resourceId"].(string)
	if !ok {
		fmt.Printf("DEBUG RegisterHost: Could not extract resourceId from response\n")
		return "", e.NewCustomError(e.ErrInternal)
	}

	fmt.Printf("DEBUG RegisterHost: Successfully registered host with ID: %s\n", resourceId)

	// Note: We're not caching the full host object due to struct mismatch
	// If you need caching, you'll need to fix the api.Host struct definition
	
	return resourceId, nil
}

func (oC *OrchCli) CreateInstance(ctx context.Context, hostID string, r *types.HostRecord) (string, error) {
	fmt.Printf("DEBUG CreateInstance: Starting - hostID: %s, Serial: %s, UUID: %s\n", hostID, r.Serial, r.UUID)
	
	if exists, err := oC.InstanceExists(ctx, r.Serial, r.UUID); exists {
		fmt.Printf("DEBUG CreateInstance: Instance already exists\n")
		return "", e.NewCustomError(e.ErrAlreadyRegistered)
	} else if err != nil {
		fmt.Printf("DEBUG CreateInstance: Error checking if instance exists: %v\n", err)
		return "", err
	}

	fmt.Printf("DEBUG CreateInstance: Validating OS Profile: %s\n", r.OSProfile)
	if err := validateOSProfile(r.OSProfile); err != nil {
		fmt.Printf("DEBUG CreateInstance: OS Profile validation failed: %v\n", err)
		return "", err
	}

	fmt.Printf("DEBUG CreateInstance: Preparing instance payload\n")
	payload, err := oC.prepareInstancePayload(hostID, r)
	if err != nil {
		fmt.Printf("DEBUG CreateInstance: Failed to prepare payload: %v\n", err)
		return "", err
	}

	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/instances", oC.Project))

	fmt.Printf("DEBUG CreateInstance: URL: %s\n", uParsed.String())
	fmt.Printf("DEBUG CreateInstance: Payload: %s\n", string(payload))

	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodPost, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("DEBUG CreateInstance: HTTP request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	
	fmt.Printf("DEBUG CreateInstance: Response Status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG CreateInstance: Response Body: %s\n", bodyString)

	// Accept both 200 and 201 as success (like we did for RegisterHost)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		fmt.Printf("DEBUG CreateInstance: Instance creation failed with status %d\n", resp.StatusCode)
		fmt.Printf("DEBUG CreateInstance: Expected status 200 or 201, got %d\n", resp.StatusCode)
		return "", e.NewCustomError(e.ErrInstanceFailed)
	}

	// Use generic JSON parsing to avoid struct mismatch issues
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Printf("DEBUG CreateInstance: JSON unmarshal failed: %v\n", err)
		fmt.Printf("DEBUG CreateInstance: Response body was: %s\n", bodyString)
		return "", e.NewCustomError(e.ErrInternal)
	}

	// Extract resourceId directly from the map
	resourceId, ok := response["resourceId"].(string)
	if !ok {
		fmt.Printf("DEBUG CreateInstance: Could not extract resourceId from response\n")
		fmt.Printf("DEBUG CreateInstance: Available keys in response: %v\n", getMapKeys(response))
		return "", e.NewCustomError(e.ErrInternal)
	}

	fmt.Printf("DEBUG CreateInstance: Successfully created instance with ID: %s\n", resourceId)
	return resourceId, nil
}

// Helper function to debug available keys in response
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func validateOSProfile(osProfile string) error {
	osRe := regexp.MustCompile(validator.OSPIDPATTERN)
	if !osRe.MatchString(osProfile) {
		return e.NewCustomError(e.ErrInvalidOSProfile)
	}
	return nil
}

func (oC *OrchCli) prepareInstancePayload(hostID string, r *types.HostRecord) ([]byte, error) {
	payload := &api.Instance{
		HostID:          &hostID,
		OsID:            &r.OSProfile,
		SecurityFeature: new(api.SecurityFeature),
		Kind:            new(api.InstanceKind),
	}

	if r.RemoteUser != "" {
		payload.LocalAccountID = &r.RemoteUser
	}
	*payload.Kind = api.INSTANCEKINDUNSPECIFIED

	osResource, ok := oC.OSProfileCache[r.OSProfile]
	if !ok {
		return nil, e.NewCustomError(e.ErrInternal)
	}

	*payload.SecurityFeature = *osResource.SecurityFeature
	if r.Secure != types.SecureTrue {
		*payload.SecurityFeature = api.SECURITYFEATURENONE
	}

	return json.Marshal(payload)
}

func obtainRequestPath(oC *OrchCli, input, pattern, pathByID, pathByName, filter string) (url.URL, *regexp.Regexp) {
	uParsed := *oC.SvcURL
	// match os to id pattern
	re := regexp.MustCompile(pattern)
	if re.MatchString(input) {
		// if successful, query db to check if site exists by id
		uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf(pathByID, oC.Project, input))
		fmt.Printf("Constructed URL: %s\n", uParsed.String())
		fmt.Printf("RawQuery: %s\n", uParsed.RawQuery)
	} else {
		// else query db to check if site exists by name
		uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf(pathByName, oC.Project))
		// REPLACE THIS SECTION:
		// query := uParsed.Query()
		// query.Set("filter", fmt.Sprintf("%s=%q", filter, input))
		// uParsed.RawQuery = query.Encode()
		
		// WITH THIS:
		uParsed.RawQuery = fmt.Sprintf("filter=%s=\"%s\"", filter, input)
		// Debug logging
		fmt.Printf("Constructed URL: %s\n", uParsed.String())
		fmt.Printf("RawQuery: %s\n", uParsed.RawQuery)
	}
	return uParsed, re
}

func (oC *OrchCli) InstanceExists(ctx context.Context, sn, uuid string) (bool, error) {
	pathByName := "/v1/projects/%s/compute/instances"
	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf(pathByName, oC.Project))
	query := uParsed.Query()
	switch {
	case sn != "" && uuid != "":
		query.Set("filter", fmt.Sprintf("%s=%q AND %s=%q", "host.serialNumber", sn, "host.uuid", uuid))
	case sn != "":
		query.Set("filter", fmt.Sprintf("%s=%q", "host.serialNumber", sn))
	case uuid != "":
		query.Set("filter", fmt.Sprintf("%s=%q", "host.uuid", uuid))
	default:
		return false, nil
	}
	uParsed.RawQuery = query.Encode()

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodGet, http.NoBody)
	if err != nil {
		return false, e.NewCustomError(e.ErrInternal)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, e.NewCustomError(e.ErrInternal)
	}

	var instances api.InstanceList

	if err := json.NewDecoder(resp.Body).Decode(&instances); err != nil {
		return false, e.NewCustomError(e.ErrInternal)
	}

	if *instances.TotalElements > 0 {
		return true, nil
	}

	return false, nil
}

// GetHostID is invoked when a pre-registered host has caused StatusPreconditionFailed error
// in /register call. HostID is queried in such cases to attempt to complete the import
// ( i.e. - instance creation & site,metadata allocation).
// Note that the post register steps would be executed only for a strict match of the Serial
// Number and UUID. A partial match might indicate intentional registration of another host.
func (oC *OrchCli) GetHostID(ctx context.Context, sn, uuid string) (string, error) {
	if sn == "" && uuid == "" {
		return "", e.NewCustomError(e.ErrInternal)
	}

	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/hosts", oC.Project))
	query := uParsed.Query()
	query.Set("filter", fmt.Sprintf("%s=%q AND %s=%q", "serialNumber", sn, "uuid", uuid))
	uParsed.RawQuery = query.Encode()

	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodGet, http.NoBody)
	if err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", e.NewCustomError(e.ErrInternal)
	}

	var hosts api.HostsList
	err = json.NewDecoder(resp.Body).Decode(&hosts)
	if err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	if *hosts.TotalElements != 1 {
		return "", e.NewCustomError(e.ErrHostDetailMismatch)
	}
	// Get host at index 0 as this is the only host available
	host := (*hosts.Hosts)[0]

	// If an instance for the host already exists, return an error
	if host.Instance != nil {
		return "", e.NewCustomError(e.ErrAlreadyRegistered)
	}
	oC.HostCache[*host.ResourceId] = host
	return *host.ResourceId, nil
}

func (oC *OrchCli) GetOsProfileID(ctx context.Context, os string) (string, error) {
	if os == "" {
		return "", e.NewCustomError(e.ErrInvalidOSProfile)
	}
	if osResource, ok := oC.OSProfileCache[os]; ok {
		return *osResource.ResourceId, nil
	}

	pathByID := "/v1/projects/%s/compute/os/%s"
	pathByName := "/v1/projects/%s/compute/os"
	uParsed, oSPRe := obtainRequestPath(oC, os, validator.OSPIDPATTERN, pathByID, pathByName, "profileName")

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodGet, http.NoBody)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", e.NewCustomError(e.ErrInvalidOSProfile)
	}

	if oSPRe.MatchString(os) {
		var osResource api.OperatingSystemResource
		if err := json.NewDecoder(resp.Body).Decode(&osResource); err != nil {
			return "", e.NewCustomError(e.ErrInternal)
		}
		oC.OSProfileCache[os] = osResource
		oC.OSProfileCache[*osResource.ProfileName] = osResource
		return *osResource.ResourceId, nil
	}

	var osResources api.OperatingSystemResourceList

	if err := json.NewDecoder(resp.Body).Decode(&osResources); err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	for _, osResource := range *osResources.OperatingSystemResources {
		if *osResource.ProfileName == os {
			oC.OSProfileCache[os] = osResource
			oC.OSProfileCache[*osResource.ResourceId] = osResource
			return *osResource.ResourceId, nil
		}
	}

	return "", e.NewCustomError(e.ErrInvalidOSProfile)
}

func (oC *OrchCli) GetSiteID(ctx context.Context, site string) (string, error) {
	if site == "" {
		return "", nil
	}
	if siteResource, ok := oC.SiteCache[site]; ok {
		return *siteResource.ResourceId, nil
	}
	
	pathByID := "/v1/projects/%s/regions/regionID/sites/%s"
	pathByName := "/v1/projects/%s/regions/regionID/sites"
	uParsed, siteRe := obtainRequestPath(oC, site, validator.SITEIDPATTERN, pathByID, pathByName, "name")
	
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodGet, http.NoBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", e.NewCustomError(e.ErrInvalidSite)
	}

	if siteRe.MatchString(site) {
		// Handle single site response (by ID)
		var siteResource api.Site
		if err := json.NewDecoder(resp.Body).Decode(&siteResource); err != nil {
			return "", e.NewCustomError(e.ErrInternal)
		}
		oC.SiteCache[site] = siteResource
		oC.SiteCache[*siteResource.Name] = siteResource
		return *siteResource.ResourceId, nil
	}

	// Handle sites list response (by name) - parse manually to avoid struct issues
	bodyBytes, _ := io.ReadAll(resp.Body)
	
	// Use a generic map to parse the JSON
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}
	
	sites, ok := response["sites"].([]interface{})
	if !ok {
		return "", e.NewCustomError(e.ErrInvalidSite)
	}
	
	for _, siteInterface := range sites {
		siteMap, ok := siteInterface.(map[string]interface{})
		if !ok {
			continue
		}
		
		siteName, ok := siteMap["name"].(string)
		if !ok {
			continue
		}
		
		if siteName == site {
			resourceId, ok := siteMap["resourceId"].(string)
			if !ok {
				return "", e.NewCustomError(e.ErrInternal)
			}
			return resourceId, nil
		}
	}
	
	return "", e.NewCustomError(e.ErrInvalidSite)
}

func (oC *OrchCli) GetLocalAccountID(ctx context.Context, lAName string) (string, error) {
	if lAName == "" {
		return "", nil
	}
	if lAResource, ok := oC.LACache[lAName]; ok {
		return *lAResource.ResourceId, nil
	}

	pathByID := "/v1/projects/%s/localAccounts/%s"
	pathByName := "/v1/projects/%s/localAccounts"
	uParsed, lAIDRe := obtainRequestPath(oC, lAName, validator.LAIDPATTERN, pathByID, pathByName, "username")

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodGet, http.NoBody)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", e.NewCustomError(e.ErrInvalidLocalAccount)
	}

	if lAIDRe.MatchString(lAName) {
		var localAcc api.LocalAccount
		if err := json.NewDecoder(resp.Body).Decode(&localAcc); err != nil {
			return "", e.NewCustomError(e.ErrInternal)
		}
		oC.LACache[lAName] = localAcc
		oC.LACache[localAcc.Username] = localAcc
		return *localAcc.ResourceId, nil
	}

	var lAs api.LocalAccountList

	if err := json.NewDecoder(resp.Body).Decode(&lAs); err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	for _, la := range *lAs.LocalAccounts {
		if la.Username == lAName {
			oC.LACache[lAName] = la
			oC.LACache[*la.ResourceId] = la
			return *la.ResourceId, nil
		}
	}

	return "", e.NewCustomError(e.ErrInvalidLocalAccount)
}

func (oC *OrchCli) AllocateHostToSiteAndAddMetadata(ctx context.Context, hostID, siteID, metadata string) error {
	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Starting - hostID: %s, siteID: %s, metadata: %s\n", hostID, siteID, metadata)
	
	if siteID == "" && metadata == "" {
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Nothing to do - both siteID and metadata are empty\n")
		return nil
	}

	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/hosts/%s", oC.Project, hostID))

	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: URL: %s\n", uParsed.String())

	metadataToSend, err := DecodeMetadata(metadata)
	if err != nil {
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Failed to decode metadata: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Decoded metadata: %+v\n", metadataToSend)

	// Prepare the payload using a generic map to avoid struct issues
	payload := make(map[string]interface{})
	
	if host, ok := oC.HostCache[hostID]; ok {
		if host.Name != "" {  // Changed: removed nil check and dereference
			payload["name"] = host.Name  // Changed: removed dereference
			fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Added name from cache: %s\n", host.Name)  // Changed: removed dereference
		}
	} else {
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Host not found in cache, will send without name\n")
		payload["name"] =""
	}
	
	if siteID != "" {
		payload["siteId"] = siteID
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Added siteId: %s\n", siteID)
	}
	
	if metadata != "" && metadataToSend != nil {
		payload["metadata"] = metadataToSend
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Added metadata\n")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: JSON marshal failed: %v\n", err)
		return e.NewCustomError(e.ErrInternal)
	}

	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Request payload: %s\n", string(jsonData))

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodPatch, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: HTTP request failed: %v\n", err)
		return err
	}

	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	
	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Response Status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Response Body: %s\n", bodyString)

	// Accept both 200 and 204 as success for PATCH operations
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Operation failed with status %d\n", resp.StatusCode)
		fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Expected status 200 or 204, got %d\n", resp.StatusCode)
		
		// Try to parse error response
		if bodyString != "" {
			var errorResponse map[string]interface{}
			if json.Unmarshal(bodyBytes, &errorResponse) == nil {
				fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Error response: %+v\n", errorResponse)
			}
		}
		
		return e.NewCustomError(e.ErrHostSiteMetadataFailed)
	}

	fmt.Printf("DEBUG AllocateHostToSiteAndAddMetadata: Successfully updated host\n")
	return nil
}

func (oC *OrchCli) doRequest(ctx context.Context, targetURL, method string, payload io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, targetURL, payload)
	if err != nil {
		return nil, e.NewCustomError(e.ErrHTTPReq)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oC.Jwt))

	resp, err := client.Do(req)
	if err != nil {
		return nil, e.NewCustomError(e.ErrHTTPReq)
	}
	return resp, nil
}

func DecodeMetadata(metadata string) (*api.Metadata, error) {
	metadataList := make(api.Metadata, 0)
	if metadata == "" {
		return &metadataList, nil
	}
	metadataPairs := strings.Split(metadata, "&")
	for _, pair := range metadataPairs {
		kv := strings.Split(pair, "=")
		if len(kv) != kVSize {
			return &metadataList, e.NewCustomError(e.ErrInvalidMetadata)
		}
		mItem := MetadataItem{
			Key:   kv[0],
			Value: kv[1],
		}
		metadataList = append(metadataList, mItem)
	}
	return &metadataList, nil
}
