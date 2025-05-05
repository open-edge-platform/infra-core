// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package orchcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	u "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	e "github.com/open-edge-platform/infra-core/bulk-import-tools/internal/errors"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/orchcli"
	"github.com/open-edge-platform/infra-core/bulk-import-tools/internal/types"
)

var (
	sn                  = "test-serial"
	uuid                = "4c4c4544-0000-5555-8888-cac04f515233"
	autoOnboard         = false
	jwt                 = "test-token"
	resourceID          = "testResource"
	project             = "testProject"
	hostID              = "host-12345678"
	oSResourceID        = "os-12345678"
	localAccountID      = "localaccount-12345678"
	siteID              = "site-12345678"
	siteName            = "test-site"
	kind                = api.INSTANCEKINDUNSPECIFIED
	securityFeatureNone = api.SECURITYFEATURENONE
)

const (
	tokStr      = "Bearer test-token"
	contentType = "application/json"
)

// Mock server to simulate edge infrastructure manager service.
func setupMockServer(status int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != fmt.Sprintf("/v1/projects/%s/compute/hosts/register", project) {
			http.Error(w, "Invalid request method", http.StatusNotFound)
			return
		}

		// Check if the request content type is application/json
		if r.Header.Get("Content-Type") != contentType {
			http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
			return
		}
		// Check if bearer token available
		if r.Header.Get("Authorization") != tokStr {
			http.Error(w, "Invalid content type", http.StatusUnauthorized)
			return
		}
		// Read the body of the request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Unmarshal the JSON data into the struct
		var data api.HostRegisterInfo
		if err1 := json.Unmarshal(body, &data); err1 != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		assert.Equal(nil, sn, *data.SerialNumber)
		assert.Equal(nil, uuid, data.Uuid.String())
		assert.Equal(nil, autoOnboard, *data.AutoOnboard)

		rID := resourceID
		resp := api.Host{
			ResourceId: &rID,
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		err = json.NewEncoder(w).Encode(resp)
		assert.NoError(nil, err)
	})

	return httptest.NewServer(handler)
}

func newOrchCli(t *testing.T, u, p, jwt string) *orchcli.OrchCli {
	t.Helper()
	uParsed, err := url.Parse(u)
	require.NoError(t, err)
	oc := &orchcli.OrchCli{
		SvcURL:         uParsed,
		Project:        p,
		Jwt:            jwt,
		OSProfileCache: make(map[string]api.OperatingSystemResource),
		SiteCache:      make(map[string]api.Site),
		LACache:        make(map[string]api.LocalAccount),
		HostCache:      make(map[string]api.Host),
	}
	return oc
}

func TestRegisterHost(t *testing.T) {
	// Set up a mock server
	mockServer := setupMockServer(http.StatusCreated)
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)

	resp, err := oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)
	assert.NoError(t, err)

	assert.Equal(t, resourceID, resp)
}

func TestRegisterHostFail(t *testing.T) {
	// Set up a mock server
	mockServer := setupMockServer(http.StatusPreconditionFailed)
	defer mockServer.Close()

	oc := newOrchCli(t, "", project, jwt)

	resp, err := oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)
	assert.Error(t, err)
	assert.Empty(t, resp)

	oc = newOrchCli(t, mockServer.URL, project, "")

	// Auth Error
	resp, err = oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)
	assert.Error(t, err)
	assert.Empty(t, resp)

	oc = newOrchCli(t, mockServer.URL, project, jwt)
	resp, err = oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)
	assert.Error(t, err)
	assert.Empty(t, resp)
}

func TestRegisterHostAlreadyExist(t *testing.T) {
	// Set up a mock server
	mockServer := setupMockServer(http.StatusPreconditionFailed)
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)

	resp, err := oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)

	assert.Error(t, err)
	assert.Empty(t, resp)
}

func TestRegisterHostProjectName(t *testing.T) {
	// Set up a mock server
	mockServer := setupMockServer(http.StatusCreated)
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, "", jwt)

	resp, err := oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)
	assert.Error(t, err)
	assert.Empty(t, resp)

	oc = newOrchCli(t, mockServer.URL, project, jwt)
	// project name required
	resp, err = oc.RegisterHost(context.Background(), "", sn, uuid, autoOnboard)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

func TestDecodeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		expected *api.Metadata
		err      error
	}{
		{
			name:     "Empty metadata",
			metadata: "",
			expected: &api.Metadata{},
			err:      nil,
		},
		{
			name:     "Valid metadata",
			metadata: "key1=value1&key2=value2",
			expected: &api.Metadata{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			err: nil,
		},
		{
			name:     "Invalid metadata format",
			metadata: "key1=value1&key2",
			expected: &api.Metadata{},
			err:      e.NewCustomError(e.ErrInvalidMetadata),
		},
		{
			name:     "Single valid metadata",
			metadata: "key1=value1",
			expected: &api.Metadata{
				{Key: "key1", Value: "value1"},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := orchcli.DecodeMetadata(tt.metadata)
			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func setupMockServerForCreateInstance(t *testing.T, status int, expectedPayload *api.Instance) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != fmt.Sprintf("/v1/projects/%s/compute/instances", project) {
			http.Error(w, "Invalid request path", http.StatusNotFound)
			return
		}

		// Check if the request content type is application/json
		if r.Header.Get("Content-Type") != contentType {
			http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
			return
		}

		// Check if bearer token is available
		if r.Header.Get("Authorization") != tokStr {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Read the body of the request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Unmarshal the JSON data into the struct
		var data api.Instance
		if err = json.Unmarshal(body, &data); err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		assert.EqualValues(t, expectedPayload, &data)

		if status == http.StatusCreated {
			resp := api.Instance{
				ResourceId: &resourceID,
			}
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(status)
			err = json.NewEncoder(w).Encode(resp)
			assert.NoError(nil, err)
		} else {
			w.WriteHeader(status)
		}
	})

	return httptest.NewServer(handler)
}

func TestCreateInstance(t *testing.T) {
	mockServer := setupMockServerForCreateInstance(t, http.StatusCreated, &api.Instance{
		HostID:          &hostID,
		OsID:            &oSResourceID,
		SecurityFeature: new(api.SecurityFeature),
		Kind:            &kind,
	})
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)
	oc.OSProfileCache[oSResourceID] = api.OperatingSystemResource{
		SecurityFeature: new(api.SecurityFeature),
	}

	hr := &types.HostRecord{
		OSProfile:  oSResourceID,
		RemoteUser: "",
		Secure:     types.SecureTrue,
	}

	resp, err := oc.CreateInstance(context.Background(), hostID, hr)
	assert.NoError(t, err)
	assert.Equal(t, resourceID, resp)
}

func TestCreateInstanceInvalidOSProfile(t *testing.T) {
	mockServer := setupMockServerForCreateInstance(t, http.StatusBadRequest, nil)
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)

	hr := &types.HostRecord{
		OSProfile:  "invalid-os-profile",
		RemoteUser: "",
		Secure:     types.SecureTrue,
	}

	resp, err := oc.CreateInstance(context.Background(), hostID, hr)
	assert.Error(t, err)
	assert.Empty(t, resp)
}

func TestCreateInstanceInternalError(t *testing.T) {
	mockServer := setupMockServerForCreateInstance(t, http.StatusInternalServerError, nil)
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)

	hr := &types.HostRecord{
		OSProfile:  oSResourceID,
		RemoteUser: "",
		Secure:     types.SecureTrue,
	}

	resp, err := oc.CreateInstance(context.Background(), hostID, hr)
	assert.Error(t, err)
	assert.Empty(t, resp)

	oc.OSProfileCache[oSResourceID] = api.OperatingSystemResource{
		SecurityFeature: new(api.SecurityFeature),
	}

	resp, err = oc.CreateInstance(context.Background(), hostID, hr)
	assert.Error(t, err)
	assert.Empty(t, resp)
}

func TestCreateInstanceWithLocalAccount(t *testing.T) {
	mockServer := setupMockServerForCreateInstance(t, http.StatusCreated, &api.Instance{
		HostID:          &hostID,
		OsID:            &oSResourceID,
		LocalAccountID:  &localAccountID,
		SecurityFeature: new(api.SecurityFeature),
		Kind:            &kind,
	})
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)
	oc.OSProfileCache[oSResourceID] = api.OperatingSystemResource{
		SecurityFeature: new(api.SecurityFeature),
	}

	hr := &types.HostRecord{
		OSProfile:  oSResourceID,
		RemoteUser: localAccountID,
		Secure:     types.SecureTrue,
	}

	resp, err := oc.CreateInstance(context.Background(), hostID, hr)
	assert.NoError(t, err)
	assert.Equal(t, resourceID, resp)
}

func TestCreateInstanceSecurityFeatureNone(t *testing.T) {
	mockServer := setupMockServerForCreateInstance(t, http.StatusCreated, &api.Instance{
		HostID:          &hostID,
		OsID:            &oSResourceID,
		SecurityFeature: &securityFeatureNone,
		Kind:            &kind,
	})
	defer mockServer.Close()

	oc := newOrchCli(t, mockServer.URL, project, jwt)
	oc.OSProfileCache[oSResourceID] = api.OperatingSystemResource{
		SecurityFeature: new(api.SecurityFeature),
	}
	hr := &types.HostRecord{
		OSProfile:  oSResourceID,
		RemoteUser: "",
		Secure:     types.SecureFalse,
	}

	resp, err := oc.CreateInstance(context.Background(), hostID, hr)
	assert.NoError(t, err)
	assert.Equal(t, resourceID, resp)
}

func setupMockServerForGetResources(t *testing.T, status int, responseBody interface{}) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is GET
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the request content type is application/json
		if r.Header.Get("Accept") != contentType {
			http.Error(w, "Invalid accept header", http.StatusUnsupportedMediaType)
			return
		}

		// Check if bearer token is available
		if r.Header.Get("Authorization") != tokStr {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)

		if responseBody != nil {
			err := json.NewEncoder(w).Encode(responseBody)
			assert.NoError(t, err)
		}
	})

	return httptest.NewServer(handler)
}

func TestGetOsProfileID(t *testing.T) {
	osProfileID := "os-12345678"
	osProfileName := "test-os-profile"
	osResource := api.OperatingSystemResource{
		ResourceId:      &osProfileID,
		ProfileName:     &osProfileName,
		SecurityFeature: new(api.SecurityFeature),
	}
	osResourceList := api.OperatingSystemResourceList{
		TotalElements: new(int),
		OperatingSystemResources: &[]api.OperatingSystemResource{
			osResource,
		},
	}
	*osResourceList.TotalElements = 1

	t.Run("Valid OS Profile ID", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, osResource)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetOsProfileID(context.Background(), osProfileID)
		assert.NoError(t, err)
		assert.Equal(t, osProfileID, resp)
	})

	t.Run("Valid OS Profile Name", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, osResourceList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetOsProfileID(context.Background(), osProfileName)
		assert.NoError(t, err)
		assert.Equal(t, osProfileID, resp)
	})

	t.Run("Invalid OS Profile", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusNotFound, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetOsProfileID(context.Background(), "invalid-os-profile")
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Empty OS Profile", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)

		resp, err := oc.GetOsProfileID(context.Background(), "")
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("OS Profile Cached", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)
		oc.OSProfileCache[osProfileID] = osResource

		resp, err := oc.GetOsProfileID(context.Background(), osProfileID)
		assert.NoError(t, err)
		assert.Equal(t, osProfileID, resp)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusInternalServerError, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetOsProfileID(context.Background(), osProfileID)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})
}

func TestGetSiteID(t *testing.T) {
	siteResource := api.Site{
		ResourceId: &siteID,
		Name:       &siteName,
	}
	siteList := api.SitesList{
		TotalElements: new(int),
		Sites: &[]api.Site{
			siteResource,
		},
	}
	*siteList.TotalElements = 1

	t.Run("Valid Site ID", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, siteResource)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetSiteID(context.Background(), siteID)
		assert.NoError(t, err)
		assert.Equal(t, siteID, resp)
	})

	t.Run("Valid Site Name", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, siteList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetSiteID(context.Background(), siteName)
		assert.NoError(t, err)
		assert.Equal(t, siteID, resp)
	})

	t.Run("Invalid Site", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusNotFound, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetSiteID(context.Background(), "invalid-site")
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Empty Site", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)

		resp, err := oc.GetSiteID(context.Background(), "")
		assert.NoError(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Site Cached", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)
		oc.SiteCache[siteID] = siteResource

		resp, err := oc.GetSiteID(context.Background(), siteID)
		assert.NoError(t, err)
		assert.Equal(t, siteID, resp)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusInternalServerError, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetSiteID(context.Background(), siteID)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})
}

func TestGetLocalAccountID(t *testing.T) {
	localAccountID := "localaccount-12345678"
	localAccountName := "test-local-account"
	localAccountResource := api.LocalAccount{
		ResourceId: &localAccountID,
		Username:   localAccountName,
	}
	localAccountList := api.LocalAccountList{
		TotalElements: new(int),
		LocalAccounts: &[]api.LocalAccount{
			localAccountResource,
		},
	}
	*localAccountList.TotalElements = 1

	t.Run("Valid Local Account ID", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, localAccountResource)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetLocalAccountID(context.Background(), localAccountID)
		assert.NoError(t, err)
		assert.Equal(t, localAccountID, resp)
	})

	t.Run("Valid Local Account Name", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, localAccountList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetLocalAccountID(context.Background(), localAccountName)
		assert.NoError(t, err)
		assert.Equal(t, localAccountID, resp)
	})

	t.Run("Invalid Local Account", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusNotFound, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetLocalAccountID(context.Background(), "invalid-local-account")
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Empty Local Account", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)

		resp, err := oc.GetLocalAccountID(context.Background(), "")
		assert.NoError(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Local Account Cached", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)
		oc.LACache[localAccountID] = localAccountResource

		resp, err := oc.GetLocalAccountID(context.Background(), localAccountID)
		assert.NoError(t, err)
		assert.Equal(t, localAccountID, resp)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusInternalServerError, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetLocalAccountID(context.Background(), localAccountID)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})
}

func setupMockServerForAllocateHost(t *testing.T, status int, expectedPayload *api.Host) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is PATCH
		if r.Method != http.MethodPatch {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the request path matches
		expectedPath := fmt.Sprintf("/v1/projects/%s/compute/hosts/%s", project, hostID)
		if r.URL.Path != expectedPath {
			http.Error(w, "Invalid request path", http.StatusNotFound)
			return
		}

		// Check if the request content type is application/json
		if r.Header.Get("Content-Type") != contentType {
			http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
			return
		}

		// Check if bearer token is available
		if r.Header.Get("Authorization") != tokStr {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Read the body of the request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Unmarshal the JSON data into the struct
		var data api.Host
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		assert.EqualValues(t, expectedPayload, &data)

		w.WriteHeader(status)
	})

	return httptest.NewServer(handler)
}

func TestAllocateHostToSiteAndAddMetadata(t *testing.T) {
	t.Run("Allocate Host with Site ID and Metadata", func(t *testing.T) {
		metadata := "key1=value1&key2=value2"
		decodedMetadata, err := orchcli.DecodeMetadata(metadata)
		assert.NoError(t, err)
		expectedPayload := &api.Host{
			SiteId:   &siteID,
			Metadata: decodedMetadata,
		}

		mockServer := setupMockServerForAllocateHost(t, http.StatusOK, expectedPayload)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		err = oc.AllocateHostToSiteAndAddMetadata(context.Background(), hostID, siteID, metadata)
		assert.NoError(t, err)
	})

	t.Run("Allocate Host with only Site ID", func(t *testing.T) {
		expectedPayload := &api.Host{
			SiteId: &siteID,
		}

		mockServer := setupMockServerForAllocateHost(t, http.StatusOK, expectedPayload)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		err := oc.AllocateHostToSiteAndAddMetadata(context.Background(), hostID, siteID, "")
		assert.NoError(t, err)
	})

	t.Run("Allocate Host with only Metadata", func(t *testing.T) {
		metadata := "key1=value1"
		decodedMetadata, err := orchcli.DecodeMetadata(metadata)
		assert.NoError(t, err)
		expectedPayload := &api.Host{
			Metadata: decodedMetadata,
		}

		mockServer := setupMockServerForAllocateHost(t, http.StatusOK, expectedPayload)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		err = oc.AllocateHostToSiteAndAddMetadata(context.Background(), hostID, "", metadata)
		assert.NoError(t, err)
	})

	t.Run("No Site ID and Metadata", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)

		err := oc.AllocateHostToSiteAndAddMetadata(context.Background(), hostID, "", "")
		assert.NoError(t, err)
	})

	t.Run("Invalid Metadata Format", func(t *testing.T) {
		metadata := "key1=value1&key2"

		oc := newOrchCli(t, "", project, jwt)

		err := oc.AllocateHostToSiteAndAddMetadata(context.Background(), hostID, siteID, metadata)
		assert.Error(t, err)
		assert.Equal(t, e.NewCustomError(e.ErrInvalidMetadata), err)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockServer := setupMockServerForAllocateHost(t, http.StatusInternalServerError, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		err := oc.AllocateHostToSiteAndAddMetadata(context.Background(), hostID, siteID, "")
		assert.Error(t, err)
		assert.Equal(t, e.NewCustomError(e.ErrHostSiteMetadataFailed), err)
	})
}

func TestInstanceExists(t *testing.T) {
	instanceList := api.InstanceList{
		TotalElements: new(int),
	}
	*instanceList.TotalElements = 1

	t.Run("Instance Exists with Serial Number and UUID", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, instanceList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		exists, err := oc.InstanceExists(context.Background(), sn, uuid)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Instance Exists with Serial Number Only", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, instanceList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		exists, err := oc.InstanceExists(context.Background(), sn, "")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Instance Exists with UUID Only", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, instanceList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		exists, err := oc.InstanceExists(context.Background(), "", uuid)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Instance Does Not Exist", func(t *testing.T) {
		instanceList.TotalElements = new(int) // Set TotalElements to 0
		mockServer := setupMockServerForGetResources(t, http.StatusOK, instanceList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		exists, err := oc.InstanceExists(context.Background(), sn, uuid)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Empty Serial Number and UUID", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)

		exists, err := oc.InstanceExists(context.Background(), "", "")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusInternalServerError, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		exists, err := oc.InstanceExists(context.Background(), sn, uuid)
		assert.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid Response Format", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, "invalid-response")
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		exists, err := oc.InstanceExists(context.Background(), sn, uuid)
		assert.Error(t, err)
		assert.False(t, exists)
	})
}

func TestGetHostID(t *testing.T) {
	hostResource := api.Host{
		ResourceId:   &hostID,
		SerialNumber: &sn,
		Uuid:         func() *u.UUID { id := u.MustParse(uuid); return &id }(),
	}
	hostList := api.HostsList{
		TotalElements: new(int),
		Hosts: &[]api.Host{
			hostResource,
		},
	}
	*hostList.TotalElements = 1

	t.Run("Valid Host ID", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, hostList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetHostID(context.Background(), sn, uuid)
		assert.NoError(t, err)
		assert.Equal(t, hostID, resp)
	})

	t.Run("Valid Host Serial Number and UUID", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, hostList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetHostID(context.Background(), sn, uuid)
		assert.NoError(t, err)
		assert.Equal(t, hostID, resp)
	})

	t.Run("Host Not Found", func(t *testing.T) {
		hostList.TotalElements = new(int) // Set TotalElements to 0
		mockServer := setupMockServerForGetResources(t, http.StatusOK, hostList)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetHostID(context.Background(), sn, uuid)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Empty Serial Number and UUID", func(t *testing.T) {
		oc := newOrchCli(t, "", project, jwt)

		resp, err := oc.GetHostID(context.Background(), "", "")
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusInternalServerError, nil)
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetHostID(context.Background(), sn, uuid)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("Invalid Response Format", func(t *testing.T) {
		mockServer := setupMockServerForGetResources(t, http.StatusOK, "invalid-response")
		defer mockServer.Close()

		oc := newOrchCli(t, mockServer.URL, project, jwt)

		resp, err := oc.GetHostID(context.Background(), sn, uuid)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})
}
