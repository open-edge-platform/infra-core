// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"testing"

	api "github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/stretchr/testify/require"
)

// ListAllInstances retrieves all InstanceResource objects by iterating over paginated responses.
func ListAllInstances(ctx context.Context, client *api.ClientWithResponses, pageSize int32) ([]api.InstanceResource, error) {
	var allInstances []api.InstanceResource
	var offset int32 = 0

	for {

		// Call the API to get a paginated list of instances
		response, err := client.InstanceServiceListInstancesWithResponse(ctx, &api.InstanceServiceListInstancesParams{
			PageSize: &pageSize,
			Offset:   &offset,
		},
			AddJWTtoTheHeader, AddProjectIDtoTheHeader)
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
		offset += int32(pageSize)
	}

	return allInstances, nil
}

// ListAllHosts retrieves all HostResource objects by iterating over paginated responses.
func ListAllHosts(ctx context.Context, client *api.ClientWithResponses, pageSize int32) ([]api.HostResource, error) {
	var allHosts []api.HostResource
	var offset int32 = 0

	for {
		// Call the API to get a paginated list of hosts
		response, err := client.HostServiceListHostsWithResponse(ctx, &api.HostServiceListHostsParams{
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
func ListAllRegions(ctx context.Context, client *api.ClientWithResponses, pageSize int32) ([]api.RegionResource, error) {
	var allRegions []api.RegionResource
	var offset int32 = 0

	for {
		// Call the API to get a paginated list of regions
		response, err := client.RegionServiceListRegionsWithResponse(ctx, &api.RegionServiceListRegionsParams{
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
func ListAllSites(ctx context.Context, client *api.ClientWithResponses, pageSize int32) ([]api.SiteResource, error) {
	var allSites []api.SiteResource
	var offset int32 = 0

	for {
		// Call the API to get a paginated list of sites
		response, err := client.SiteServiceListSitesWithResponse(ctx, &api.SiteServiceListSitesParams{
			PageSize: &pageSize,
			Offset:   &offset,
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
func ListAllWorkloads(ctx context.Context, client *api.ClientWithResponses, pageSize int32) ([]api.WorkloadResource, error) {
	var allWorkloads []api.WorkloadResource
	var offset int32 = 0

	for {
		// Call the API to get a paginated list of workloads
		response, err := client.WorkloadServiceListWorkloadsWithResponse(ctx, &api.WorkloadServiceListWorkloadsParams{
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

// Example test case for ListAllRegions
func TestDeleteAllRegions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	var pageSize int32 = 10
	regions, err := ListAllRegions(ctx, apiClient, pageSize)
	if err != nil {
		t.Fatalf("failed to list all regions: %v", err)
	}

	t.Logf("Retrieved %d regions", len(regions))
	for _, region := range regions {
		t.Logf("Region ID: %s, Name: %s", *region.ResourceId, *region.Name)
		DeleteRegion(t, ctx, apiClient, *region.ResourceId)
	}
}

// Example test case for ListAllSites
func TestDeleteAllSites(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	var pageSize int32 = 10
	sites, err := ListAllSites(ctx, apiClient, pageSize)
	if err != nil {
		t.Fatalf("failed to list all sites: %v", err)
	}

	t.Logf("Retrieved %d sites", len(sites))
	for _, site := range sites {
		t.Logf("Site ID: %s, Name: %s", *site.ResourceId, *site.Name)
		DeleteSite(t, ctx, apiClient, *site.ResourceId)
	}
}

// Example test case for ListAllWorkloads
func TestDeleteAllWorkloads(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	var pageSize int32 = 10
	workloads, err := ListAllWorkloads(ctx, apiClient, pageSize)
	if err != nil {
		t.Fatalf("failed to list all workloads: %v", err)
	}

	t.Logf("Retrieved %d workloads", len(workloads))
	for _, workload := range workloads {
		t.Logf("Workload ID: %s, Name: %s", *workload.WorkloadId, *workload.Name)
		DeleteWorkload(t, ctx, apiClient, *workload.ResourceId)
	}
}

// Example test case for ListAllInstances
func TestDeleteAllInstances(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	var pageSize int32 = 10
	instances, err := ListAllInstances(ctx, apiClient, pageSize)
	if err != nil {
		t.Fatalf("failed to list all instances: %v", err)
	}

	t.Logf("Retrieved %d instances", len(instances))
	for _, instance := range instances {
		t.Logf("Instance ID: %s, Name: %s", *instance.ResourceId, *instance.Name)
		DeleteInstance(t, ctx, apiClient, *instance.ResourceId)
	}
}

// Example test case for ListAllHosts
func TestDeleteAllHosts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := GetAPIClient()
	require.NoError(t, err)

	var pageSize int32 = 10
	hosts, err := ListAllHosts(ctx, apiClient, pageSize)
	if err != nil {
		t.Fatalf("failed to list all hosts: %v", err)
	}

	t.Logf("Retrieved %d hosts", len(hosts))
	for _, host := range hosts {
		t.Logf("Host ID: %s, Name: %s", *host.ResourceId, host.Name)
		SoftDeleteHost(t, ctx, apiClient, &host)
	}
}
