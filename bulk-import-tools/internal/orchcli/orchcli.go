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
	}, nil
}

// errors are customized here to record in error log.
func (oC *OrchCli) RegisterHost(ctx context.Context, host, sNo, uuid string, autoOnboard bool) (string, error) {
	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/hosts/register", oC.Project))

	// Prepare the form data
	payload := &api.HostRegisterInfo{
		Name:        &host,
		AutoOnboard: &autoOnboard,
	}

	if sNo != "" {
		payload.SerialNumber = &sNo
	}
	if uuid != "" {
		uObj := u.MustParse(uuid)
		payload.Uuid = &uObj
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodPost, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPreconditionFailed {
		return "", e.NewCustomError(e.ErrAlreadyRegistered)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", e.NewCustomError(e.ErrRegisterFailed)
	}

	var hostInfo api.Host

	if err := json.NewDecoder(resp.Body).Decode(&hostInfo); err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	return *hostInfo.ResourceId, nil
}

func (oC *OrchCli) CreateInstance(ctx context.Context, hostID string, r *types.HostRecord) (string, error) {
	if exists, err := oC.InstanceExists(ctx, r.Serial, r.UUID); exists {
		return "", e.NewCustomError(e.ErrAlreadyRegistered)
	} else if err != nil {
		return "", err
	}

	if err := validateOSProfile(r.OSProfile); err != nil {
		return "", err
	}

	payload, err := oC.prepareInstancePayload(hostID, r)
	if err != nil {
		return "", err
	}

	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/instances", oC.Project))

	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodPost, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", e.NewCustomError(e.ErrInstanceFailed)
	}

	var instanceInfo api.Instance
	if err := json.NewDecoder(resp.Body).Decode(&instanceInfo); err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	return *instanceInfo.ResourceId, nil
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
	if !r.Secure {
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
	} else {
		// else query db to check if site exists by name
		uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf(pathByName, oC.Project))
		query := uParsed.Query()
		query.Set("filter", fmt.Sprintf("%s=%q", filter, input))
		uParsed.RawQuery = query.Encode()
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

	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodGet, http.NoBody)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", e.NewCustomError(e.ErrInvalidSite)
	}

	if siteRe.MatchString(site) {
		var siteResource api.Site
		if err := json.NewDecoder(resp.Body).Decode(&siteResource); err != nil {
			return "", e.NewCustomError(e.ErrInternal)
		}
		oC.SiteCache[site] = siteResource
		oC.SiteCache[*siteResource.Name] = siteResource
		return *siteResource.ResourceId, nil
	}

	var sites api.SitesList

	if err := json.NewDecoder(resp.Body).Decode(&sites); err != nil {
		return "", e.NewCustomError(e.ErrInternal)
	}

	for _, siteItem := range *sites.Sites {
		if *siteItem.Name == site {
			oC.SiteCache[site] = siteItem
			oC.SiteCache[*siteItem.ResourceId] = siteItem
			return *siteItem.ResourceId, nil
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
	if siteID == "" && metadata == "" {
		return nil
	}

	uParsed := *oC.SvcURL
	uParsed.Path = path.Join(uParsed.Path, fmt.Sprintf("/v1/projects/%s/compute/hosts/%s", oC.Project, hostID))

	metadataToSend, err := DecodeMetadata(metadata)
	if err != nil {
		return err
	}
	// Prepare the form data
	payload := &api.Host{}
	if siteID != "" {
		payload.SiteId = &siteID
	}
	if metadata != "" {
		payload.Metadata = metadataToSend
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return e.NewCustomError(e.ErrInternal)
	}
	// Create the HTTP client and make request
	resp, err := oC.doRequest(ctx, uParsed.String(), http.MethodPatch, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return e.NewCustomError(e.ErrHostSiteMetadataFailed)
	}

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
