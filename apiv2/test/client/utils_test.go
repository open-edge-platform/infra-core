// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
)

// ListAllInstances retrieves all InstanceResource objects by iterating over paginated responses.
func ListAllInstances(
	ctx context.Context,
	client *api.ClientWithResponses,
	projectName string,
	pageSize int,
) ([]api.InstanceResource, error) {
	var allInstances []api.InstanceResource
	offset := 0

	for {
		// Call the API to get a paginated list of instances
		response, err := client.InstanceServiceListInstancesWithResponse(
			ctx,
			projectName,
			&api.InstanceServiceListInstancesParams{
				PageSize: &pageSize,
				Offset:   &offset,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %w", err)
		}

		// Check if the response is valid
		if response.JSON200 == nil {
			return nil, fmt.Errorf("unexpected response: %v", response.HTTPResponse.Status)
		}

		// Append the instances from the current page to the result
		allInstances = append(allInstances, response.JSON200.Instances...)

		// Check if there are more pages
		if !response.JSON200.HasNext {
			break
		}

		// Increment the offset for the next page
		offset += pageSize
	}

	return allInstances, nil
}

// ListAllHosts retrieves all HostResource objects by iterating over paginated responses.
func ListAllHosts(
	ctx context.Context,
	client *api.ClientWithResponses,
	projectName string,
	pageSize int,
) ([]api.HostResource, error) {
	var allHosts []api.HostResource
	offset := 0

	for {
		// Call the API to get a paginated list of hosts
		response, err := client.HostServiceListHostsWithResponse(ctx, projectName, &api.HostServiceListHostsParams{
			PageSize: &pageSize,
			Offset:   &offset,
		},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to list hosts: %w", err)
		}

		// Check if the response is valid
		if response.JSON200 == nil {
			return nil, fmt.Errorf("unexpected response: %v", response.HTTPResponse.Status)
		}

		// Append the hosts from the current page to the result
		allHosts = append(allHosts, response.JSON200.Hosts...)

		// Check if there are more pages
		if !response.JSON200.HasNext {
			break
		}

		// Increment the offset for the next page
		offset += pageSize
	}

	return allHosts, nil
}

// ListAllRegions retrieves all RegionResource objects by iterating over paginated responses.
func ListAllRegions(
	ctx context.Context, t *testing.T, client *api.ClientWithResponses, pageSize int,
) ([]api.RegionResource, error) {
	t.Helper()
	var allRegions []api.RegionResource
	offset := 0

	projectName := getProjectID(t)

	for {
		// Call the API to get a paginated list of regions
		response, err := client.RegionServiceListRegionsWithResponse(ctx, projectName, &api.RegionServiceListRegionsParams{
			PageSize: &pageSize,
			Offset:   &offset,
		},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to list regions: %w", err)
		}

		// Check if the response is valid
		if response.JSON200 == nil {
			return nil, fmt.Errorf("unexpected response: %v", response.HTTPResponse.Status)
		}

		// Append the regions from the current page to the result
		allRegions = append(allRegions, response.JSON200.Regions...)

		// Check if there are more pages
		if !response.JSON200.HasNext {
			break
		}

		// Increment the offset for the next page
		offset += pageSize
	}

	return allRegions, nil
}

// ListAllSites retrieves all SiteResource objects by iterating over paginated responses.
func ListAllSites(
	ctx context.Context, client *api.ClientWithResponses, projectName string, pageSize int,
) ([]api.SiteResource, error) {
	var allSites []api.SiteResource
	offset := 0

	for {
		// Call the API to get a paginated list of sites
		response, err := client.SiteServiceListSites2WithResponse(ctx, &api.SiteServiceListSites2Params{
			ProjectName: projectName,
			PageSize:    &pageSize,
			Offset:      &offset,
		},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to list sites: %w", err)
		}

		// Check if the response is valid
		if response.JSON200 == nil {
			return nil, fmt.Errorf("unexpected response: %v", response.HTTPResponse.Status)
		}

		// Append the sites from the current page to the result
		allSites = append(allSites, response.JSON200.Sites...)

		// Check if there are more pages
		if !response.JSON200.HasNext {
			break
		}

		// Increment the offset for the next page
		offset += pageSize
	}

	return allSites, nil
}

// ListAllWorkloads retrieves all WorkloadResource objects by iterating over paginated responses.
func ListAllWorkloads(
	ctx context.Context, client *api.ClientWithResponses, projectName string, pageSize int,
) ([]api.WorkloadResource, error) {
	var allWorkloads []api.WorkloadResource
	offset := 0

	for {
		// Call the API to get a paginated list of workloads
		response, err := client.WorkloadServiceListWorkloadsWithResponse(
			ctx, projectName,
			&api.WorkloadServiceListWorkloadsParams{
				PageSize: &pageSize,
				Offset:   &offset,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to list workloads: %w", err)
		}

		// Check if the response is valid
		if response.JSON200 == nil {
			return nil, fmt.Errorf("unexpected response: %v", response.HTTPResponse.Status)
		}

		// Append the workloads from the current page to the result
		allWorkloads = append(allWorkloads, response.JSON200.Workloads...)

		// Check if there are more pages
		if !response.JSON200.HasNext {
			break
		}

		// Increment the offset for the next page
		offset += pageSize
	}

	return allWorkloads, nil
}

// DeleteAllOSUpdatePolicies deletes all OS update policies in the system.
func DeleteAllOSUpdatePolicies(ctx context.Context, t *testing.T, apiClient *api.ClientWithResponses) {
	t.Helper()

	projectName := getProjectID(t)

	// List all OS update policies
	listResp, err := apiClient.OSUpdatePolicyListOSUpdatePolicyWithResponse(
		ctx, projectName, &api.OSUpdatePolicyListOSUpdatePolicyParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode())

	if listResp.JSON200 != nil && len(listResp.JSON200.OsUpdatePolicies) > 0 {
		t.Logf("Deleting %d OS update policies", len(listResp.JSON200.OsUpdatePolicies))
		for _, policy := range listResp.JSON200.OsUpdatePolicies {
			_, err := apiClient.OSUpdatePolicyDeleteOSUpdatePolicyWithResponse(
				ctx,
				projectName,
				*policy.ResourceId,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			if err != nil {
				t.Logf("Failed to delete policy %s: %v", *policy.ResourceId, err)
			}
		}
	}
}

// Example test case for ListAllRegions.
func TestDeleteAllRegions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	pageSize := 10
	regions, err := ListAllRegions(ctx, t, apiClient, pageSize)
	if err != nil {
		t.Fatalf("failed to list all regions: %v", err)
	}

	t.Logf("Retrieved %d regions", len(regions))
	for _, region := range regions {
		t.Logf("Region ID: %s, Name: %s", *region.ResourceId, *region.Name)
		projectName := getProjectID(t)
		resDelRegion, err := apiClient.RegionServiceDeleteRegionWithResponse(
			ctx,
			projectName,
			*region.ResourceId,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		if resDelRegion.StatusCode() != http.StatusOK {
			t.Logf("Skipping delete for region %s: status %d", *region.ResourceId, resDelRegion.StatusCode())
		}
	}
}

// Example test case for ListAllSites.
func TestDeleteAllSites(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	pageSize := 10
	sites, err := ListAllSites(ctx, apiClient, projectName, pageSize)
	if err != nil {
		t.Fatalf("failed to list all sites: %v", err)
	}

	t.Logf("Retrieved %d sites", len(sites))
	for _, site := range sites {
		t.Logf("Site ID: %s, Name: %s", *site.ResourceId, *site.Name)
		resDelSite, err := apiClient.SiteServiceDeleteSite2WithResponse(
			ctx,
			*site.ResourceId,
			&api.SiteServiceDeleteSite2Params{ProjectName: projectName},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		if resDelSite.StatusCode() != http.StatusOK {
			t.Logf("Skipping delete for site %s: status %d", *site.ResourceId, resDelSite.StatusCode())
		}
	}
}

// Example test case for ListAllWorkloads.
func TestDeleteAllWorkloads(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	pageSize := 10
	workloads, err := ListAllWorkloads(ctx, apiClient, projectName, pageSize)
	if err != nil {
		t.Fatalf("failed to list all workloads: %v", err)
	}

	t.Logf("Retrieved %d workloads", len(workloads))
	for _, workload := range workloads {
		t.Logf("Workload ID: %s, Name: %s", *workload.WorkloadId, *workload.Name)
		var workloadID string
		if workload.WorkloadId != nil {
			workloadID = *workload.WorkloadId
		} else if workload.ResourceId != nil {
			workloadID = *workload.ResourceId
		}
		if workloadID == "" {
			continue
		}
		resp, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
			ctx,
			projectName,
			workloadID,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		if resp.StatusCode() != http.StatusOK {
			t.Logf("WARNING: Failed to delete workload %s - Status Code: %d", workloadID, resp.StatusCode())
		}
	}
}

// Example test case for ListAllInstances.
func TestDeleteAllInstances(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	pageSize := 10
	instances, err := ListAllInstances(ctx, apiClient, projectName, pageSize)
	if err != nil {
		t.Fatalf("failed to list all instances: %v", err)
	}

	t.Logf("Retrieved %d instances", len(instances))
	for _, instance := range instances {
		t.Logf("Instance ID: %s, Name: %s", *instance.ResourceId, *instance.Name)
		if instance.Name != nil {
			if *instance.Name == utils.Inst1Name || *instance.Name == utils.Inst2Name {
				DeleteInstance(ctx, t, apiClient, *instance.ResourceId)
				continue
			}
		}
		if instance.Host != nil && strings.HasPrefix(instance.Host.Name, "Test Host ") {
			DeleteInstance(ctx, t, apiClient, *instance.ResourceId)
		}
	}
}

// Example test case for ListAllHosts.
func TestDeleteAllHosts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	pageSize := 10
	hosts, err := ListAllHosts(ctx, apiClient, projectName, pageSize)
	if err != nil {
		t.Fatalf("failed to list all hosts: %v", err)
	}

	t.Logf("Retrieved %d hosts", len(hosts))
	for _, host := range hosts {
		t.Logf("Host ID: %s, Name: %s", *host.ResourceId, host.Name)
		if host.Name != "" && strings.HasPrefix(host.Name, "Test Host ") {
			SoftDeleteHost(ctx, t, apiClient, &host)
		}
	}
}

func ListAllLocalAccounts(
	ctx context.Context, t *testing.T, apiClient *api.ClientWithResponses, projectName string,
) []api.LocalAccountResource {
	t.Helper()

	var allAccounts []api.LocalAccountResource
	var offset int
	pageSize := 100 // Adjust page size as needed

	for {
		resList, err := apiClient.LocalAccountServiceListLocalAccountsWithResponse(
			ctx,
			projectName,
			&api.LocalAccountServiceListLocalAccountsParams{
				Offset:   &offset,
				PageSize: &pageSize,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resList.StatusCode())

		allAccounts = append(allAccounts, resList.JSON200.LocalAccounts...)

		if !resList.JSON200.HasNext {
			break
		}
		offset += pageSize
	}

	return allAccounts
}

func TestDeleteAllLocalAccounts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	accounts := ListAllLocalAccounts(ctx, t, apiClient, projectName)

	for _, account := range accounts {
		_, err := apiClient.LocalAccountServiceDeleteLocalAccountWithResponse(
			ctx,
			projectName,
			*account.ResourceId,
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
	}
}

func TestDeleteAllOSUpdatePolicies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	// List all OS update policies
	listResp, err := apiClient.OSUpdatePolicyListOSUpdatePolicyWithResponse(
		ctx, projectName, &api.OSUpdatePolicyListOSUpdatePolicyParams{}, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode())

	if listResp.JSON200 != nil && len(listResp.JSON200.OsUpdatePolicies) > 0 {
		t.Logf("Retrieved %d OS update policies", len(listResp.JSON200.OsUpdatePolicies))
		for _, policy := range listResp.JSON200.OsUpdatePolicies {
			t.Logf("Deleting OS Update Policy ID: %s, Name: %s", *policy.ResourceId, policy.Name)
			_, err := apiClient.OSUpdatePolicyDeleteOSUpdatePolicyWithResponse(
				ctx,
				projectName,
				*policy.ResourceId,
				AddJWTtoTheHeader, AddProjectIDtoTheHeader,
			)
			if err != nil {
				t.Logf("Failed to delete policy %s: %v", *policy.ResourceId, err)
			}
		}
	} else {
		t.Logf("Retrieved 0 OS update policies")
	}
}

func TestDeleteTestOSResources(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	projectName := getProjectID(t)

	pageSize := 50
	offset := 0
	for {
		resList, err := apiClient.OperatingSystemServiceListOperatingSystemsWithResponse(
			ctx,
			projectName,
			&api.OperatingSystemServiceListOperatingSystemsParams{
				Offset:   &offset,
				PageSize: &pageSize,
			},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resList.StatusCode())

		for _, osRes := range resList.JSON200.OperatingSystemResources {
			name := ""
			if osRes.Name != nil {
				name = *osRes.Name
			}
			profileName := ""
			if osRes.ProfileName != nil {
				profileName = *osRes.ProfileName
			}
			metadata := ""
			if osRes.Metadata != nil {
				metadata = *osRes.Metadata
			}

			if strings.HasPrefix(name, "OSName") ||
				strings.HasPrefix(name, "Ubuntu 22.04 LTS generic EXT (24.08.0-n20240816) ") ||
				strings.Contains(name, "TestName#") ||
				strings.Contains(metadata, "\"createdby\":\"int-test\"") ||
				strings.HasPrefix(profileName, "ubuntu-22.04-lts-generic-ext:1.0.2-") ||
				(name == "" && profileName == "") {
				t.Logf("Deleting OS resource %s (name=%q profile=%q)", *osRes.ResourceId, name, profileName)
				DeleteOS(ctx, t, apiClient, *osRes.ResourceId)
			}
		}

		if !resList.JSON200.HasNext {
			break
		}
		offset += pageSize
	}
}
