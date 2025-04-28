# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [resources/common/v1/common.proto](#resources_common_v1_common-proto)
    - [MetadataItem](#resources-common-v1-MetadataItem)
    - [Timestamps](#resources-common-v1-Timestamps)
  
- [resources/provider/v1/provider.proto](#resources_provider_v1_provider-proto)
    - [ProviderResource](#resources-provider-v1-ProviderResource)
  
    - [ProviderKind](#resources-provider-v1-ProviderKind)
    - [ProviderVendor](#resources-provider-v1-ProviderVendor)
  
- [resources/location/v1/location.proto](#resources_location_v1_location-proto)
    - [RegionResource](#resources-location-v1-RegionResource)
    - [SiteResource](#resources-location-v1-SiteResource)
  
- [resources/os/v1/os.proto](#resources_os_v1_os-proto)
    - [OperatingSystemResource](#resources-os-v1-OperatingSystemResource)
  
    - [OsProviderKind](#resources-os-v1-OsProviderKind)
    - [OsType](#resources-os-v1-OsType)
    - [SecurityFeature](#resources-os-v1-SecurityFeature)
  
- [resources/localaccount/v1/localaccount.proto](#resources_localaccount_v1_localaccount-proto)
    - [LocalAccountResource](#resources-localaccount-v1-LocalAccountResource)
  
- [resources/status/v1/status.proto](#resources_status_v1_status-proto)
    - [StatusIndication](#resources-status-v1-StatusIndication)
  
- [resources/compute/v1/compute.proto](#resources_compute_v1_compute-proto)
    - [HostResource](#resources-compute-v1-HostResource)
    - [HostgpuResource](#resources-compute-v1-HostgpuResource)
    - [HostnicResource](#resources-compute-v1-HostnicResource)
    - [HoststorageResource](#resources-compute-v1-HoststorageResource)
    - [HostusbResource](#resources-compute-v1-HostusbResource)
    - [InstanceResource](#resources-compute-v1-InstanceResource)
    - [WorkloadMember](#resources-compute-v1-WorkloadMember)
    - [WorkloadResource](#resources-compute-v1-WorkloadResource)
  
    - [BaremetalControllerKind](#resources-compute-v1-BaremetalControllerKind)
    - [HostComponentState](#resources-compute-v1-HostComponentState)
    - [HostState](#resources-compute-v1-HostState)
    - [InstanceKind](#resources-compute-v1-InstanceKind)
    - [InstanceState](#resources-compute-v1-InstanceState)
    - [NetworkInterfaceLinkState](#resources-compute-v1-NetworkInterfaceLinkState)
    - [WorkloadKind](#resources-compute-v1-WorkloadKind)
    - [WorkloadMemberKind](#resources-compute-v1-WorkloadMemberKind)
    - [WorkloadState](#resources-compute-v1-WorkloadState)
  
- [resources/schedule/v1/schedule.proto](#resources_schedule_v1_schedule-proto)
    - [RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource)
    - [SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource)
  
    - [ScheduleStatus](#resources-schedule-v1-ScheduleStatus)
  
- [resources/telemetry/v1/telemetry.proto](#resources_telemetry_v1_telemetry-proto)
    - [TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource)
    - [TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource)
    - [TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource)
    - [TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource)
  
    - [CollectorKind](#resources-telemetry-v1-CollectorKind)
    - [SeverityLevel](#resources-telemetry-v1-SeverityLevel)
    - [TelemetryResourceKind](#resources-telemetry-v1-TelemetryResourceKind)
  
- [services/v1/services.proto](#services_v1_services-proto)
    - [CreateHostRequest](#services-v1-CreateHostRequest)
    - [CreateHostResponse](#services-v1-CreateHostResponse)
    - [CreateInstanceRequest](#services-v1-CreateInstanceRequest)
    - [CreateInstanceResponse](#services-v1-CreateInstanceResponse)
    - [CreateLocalAccountRequest](#services-v1-CreateLocalAccountRequest)
    - [CreateLocalAccountResponse](#services-v1-CreateLocalAccountResponse)
    - [CreateOperatingSystemRequest](#services-v1-CreateOperatingSystemRequest)
    - [CreateOperatingSystemResponse](#services-v1-CreateOperatingSystemResponse)
    - [CreateProviderRequest](#services-v1-CreateProviderRequest)
    - [CreateProviderResponse](#services-v1-CreateProviderResponse)
    - [CreateRegionRequest](#services-v1-CreateRegionRequest)
    - [CreateRegionResponse](#services-v1-CreateRegionResponse)
    - [CreateRepeatedScheduleRequest](#services-v1-CreateRepeatedScheduleRequest)
    - [CreateRepeatedScheduleResponse](#services-v1-CreateRepeatedScheduleResponse)
    - [CreateSingleScheduleRequest](#services-v1-CreateSingleScheduleRequest)
    - [CreateSingleScheduleResponse](#services-v1-CreateSingleScheduleResponse)
    - [CreateSiteRequest](#services-v1-CreateSiteRequest)
    - [CreateSiteResponse](#services-v1-CreateSiteResponse)
    - [CreateTelemetryLogsGroupRequest](#services-v1-CreateTelemetryLogsGroupRequest)
    - [CreateTelemetryLogsGroupResponse](#services-v1-CreateTelemetryLogsGroupResponse)
    - [CreateTelemetryLogsProfileRequest](#services-v1-CreateTelemetryLogsProfileRequest)
    - [CreateTelemetryLogsProfileResponse](#services-v1-CreateTelemetryLogsProfileResponse)
    - [CreateTelemetryMetricsGroupRequest](#services-v1-CreateTelemetryMetricsGroupRequest)
    - [CreateTelemetryMetricsGroupResponse](#services-v1-CreateTelemetryMetricsGroupResponse)
    - [CreateTelemetryMetricsProfileRequest](#services-v1-CreateTelemetryMetricsProfileRequest)
    - [CreateTelemetryMetricsProfileResponse](#services-v1-CreateTelemetryMetricsProfileResponse)
    - [CreateWorkloadMemberRequest](#services-v1-CreateWorkloadMemberRequest)
    - [CreateWorkloadMemberResponse](#services-v1-CreateWorkloadMemberResponse)
    - [CreateWorkloadRequest](#services-v1-CreateWorkloadRequest)
    - [CreateWorkloadResponse](#services-v1-CreateWorkloadResponse)
    - [DeleteHostRequest](#services-v1-DeleteHostRequest)
    - [DeleteHostResponse](#services-v1-DeleteHostResponse)
    - [DeleteInstanceRequest](#services-v1-DeleteInstanceRequest)
    - [DeleteInstanceResponse](#services-v1-DeleteInstanceResponse)
    - [DeleteLocalAccountRequest](#services-v1-DeleteLocalAccountRequest)
    - [DeleteLocalAccountResponse](#services-v1-DeleteLocalAccountResponse)
    - [DeleteOperatingSystemRequest](#services-v1-DeleteOperatingSystemRequest)
    - [DeleteOperatingSystemResponse](#services-v1-DeleteOperatingSystemResponse)
    - [DeleteProviderRequest](#services-v1-DeleteProviderRequest)
    - [DeleteProviderResponse](#services-v1-DeleteProviderResponse)
    - [DeleteRegionRequest](#services-v1-DeleteRegionRequest)
    - [DeleteRegionResponse](#services-v1-DeleteRegionResponse)
    - [DeleteRepeatedScheduleRequest](#services-v1-DeleteRepeatedScheduleRequest)
    - [DeleteRepeatedScheduleResponse](#services-v1-DeleteRepeatedScheduleResponse)
    - [DeleteSingleScheduleRequest](#services-v1-DeleteSingleScheduleRequest)
    - [DeleteSingleScheduleResponse](#services-v1-DeleteSingleScheduleResponse)
    - [DeleteSiteRequest](#services-v1-DeleteSiteRequest)
    - [DeleteSiteResponse](#services-v1-DeleteSiteResponse)
    - [DeleteTelemetryLogsGroupRequest](#services-v1-DeleteTelemetryLogsGroupRequest)
    - [DeleteTelemetryLogsGroupResponse](#services-v1-DeleteTelemetryLogsGroupResponse)
    - [DeleteTelemetryLogsProfileRequest](#services-v1-DeleteTelemetryLogsProfileRequest)
    - [DeleteTelemetryLogsProfileResponse](#services-v1-DeleteTelemetryLogsProfileResponse)
    - [DeleteTelemetryMetricsGroupRequest](#services-v1-DeleteTelemetryMetricsGroupRequest)
    - [DeleteTelemetryMetricsGroupResponse](#services-v1-DeleteTelemetryMetricsGroupResponse)
    - [DeleteTelemetryMetricsProfileRequest](#services-v1-DeleteTelemetryMetricsProfileRequest)
    - [DeleteTelemetryMetricsProfileResponse](#services-v1-DeleteTelemetryMetricsProfileResponse)
    - [DeleteWorkloadMemberRequest](#services-v1-DeleteWorkloadMemberRequest)
    - [DeleteWorkloadMemberResponse](#services-v1-DeleteWorkloadMemberResponse)
    - [DeleteWorkloadRequest](#services-v1-DeleteWorkloadRequest)
    - [DeleteWorkloadResponse](#services-v1-DeleteWorkloadResponse)
    - [GetHostRequest](#services-v1-GetHostRequest)
    - [GetHostResponse](#services-v1-GetHostResponse)
    - [GetHostSummaryRequest](#services-v1-GetHostSummaryRequest)
    - [GetHostSummaryResponse](#services-v1-GetHostSummaryResponse)
    - [GetInstanceRequest](#services-v1-GetInstanceRequest)
    - [GetInstanceResponse](#services-v1-GetInstanceResponse)
    - [GetLocalAccountRequest](#services-v1-GetLocalAccountRequest)
    - [GetLocalAccountResponse](#services-v1-GetLocalAccountResponse)
    - [GetOperatingSystemRequest](#services-v1-GetOperatingSystemRequest)
    - [GetOperatingSystemResponse](#services-v1-GetOperatingSystemResponse)
    - [GetProviderRequest](#services-v1-GetProviderRequest)
    - [GetProviderResponse](#services-v1-GetProviderResponse)
    - [GetRegionRequest](#services-v1-GetRegionRequest)
    - [GetRegionResponse](#services-v1-GetRegionResponse)
    - [GetRepeatedScheduleRequest](#services-v1-GetRepeatedScheduleRequest)
    - [GetRepeatedScheduleResponse](#services-v1-GetRepeatedScheduleResponse)
    - [GetSingleScheduleRequest](#services-v1-GetSingleScheduleRequest)
    - [GetSingleScheduleResponse](#services-v1-GetSingleScheduleResponse)
    - [GetSiteRequest](#services-v1-GetSiteRequest)
    - [GetSiteResponse](#services-v1-GetSiteResponse)
    - [GetTelemetryLogsGroupRequest](#services-v1-GetTelemetryLogsGroupRequest)
    - [GetTelemetryLogsGroupResponse](#services-v1-GetTelemetryLogsGroupResponse)
    - [GetTelemetryLogsProfileRequest](#services-v1-GetTelemetryLogsProfileRequest)
    - [GetTelemetryLogsProfileResponse](#services-v1-GetTelemetryLogsProfileResponse)
    - [GetTelemetryMetricsGroupRequest](#services-v1-GetTelemetryMetricsGroupRequest)
    - [GetTelemetryMetricsGroupResponse](#services-v1-GetTelemetryMetricsGroupResponse)
    - [GetTelemetryMetricsProfileRequest](#services-v1-GetTelemetryMetricsProfileRequest)
    - [GetTelemetryMetricsProfileResponse](#services-v1-GetTelemetryMetricsProfileResponse)
    - [GetWorkloadMemberRequest](#services-v1-GetWorkloadMemberRequest)
    - [GetWorkloadMemberResponse](#services-v1-GetWorkloadMemberResponse)
    - [GetWorkloadRequest](#services-v1-GetWorkloadRequest)
    - [GetWorkloadResponse](#services-v1-GetWorkloadResponse)
    - [HostRegister](#services-v1-HostRegister)
    - [InvalidateHostRequest](#services-v1-InvalidateHostRequest)
    - [InvalidateHostResponse](#services-v1-InvalidateHostResponse)
    - [InvalidateInstanceRequest](#services-v1-InvalidateInstanceRequest)
    - [InvalidateInstanceResponse](#services-v1-InvalidateInstanceResponse)
    - [ListHostsRequest](#services-v1-ListHostsRequest)
    - [ListHostsResponse](#services-v1-ListHostsResponse)
    - [ListInstancesRequest](#services-v1-ListInstancesRequest)
    - [ListInstancesResponse](#services-v1-ListInstancesResponse)
    - [ListLocalAccountsRequest](#services-v1-ListLocalAccountsRequest)
    - [ListLocalAccountsResponse](#services-v1-ListLocalAccountsResponse)
    - [ListLocationsRequest](#services-v1-ListLocationsRequest)
    - [ListLocationsResponse](#services-v1-ListLocationsResponse)
    - [ListLocationsResponse.LocationNode](#services-v1-ListLocationsResponse-LocationNode)
    - [ListOperatingSystemsRequest](#services-v1-ListOperatingSystemsRequest)
    - [ListOperatingSystemsResponse](#services-v1-ListOperatingSystemsResponse)
    - [ListProvidersRequest](#services-v1-ListProvidersRequest)
    - [ListProvidersResponse](#services-v1-ListProvidersResponse)
    - [ListRegionsRequest](#services-v1-ListRegionsRequest)
    - [ListRegionsResponse](#services-v1-ListRegionsResponse)
    - [ListRepeatedSchedulesRequest](#services-v1-ListRepeatedSchedulesRequest)
    - [ListRepeatedSchedulesResponse](#services-v1-ListRepeatedSchedulesResponse)
    - [ListSchedulesRequest](#services-v1-ListSchedulesRequest)
    - [ListSchedulesResponse](#services-v1-ListSchedulesResponse)
    - [ListSingleSchedulesRequest](#services-v1-ListSingleSchedulesRequest)
    - [ListSingleSchedulesResponse](#services-v1-ListSingleSchedulesResponse)
    - [ListSitesRequest](#services-v1-ListSitesRequest)
    - [ListSitesResponse](#services-v1-ListSitesResponse)
    - [ListTelemetryLogsGroupsRequest](#services-v1-ListTelemetryLogsGroupsRequest)
    - [ListTelemetryLogsGroupsResponse](#services-v1-ListTelemetryLogsGroupsResponse)
    - [ListTelemetryLogsProfilesRequest](#services-v1-ListTelemetryLogsProfilesRequest)
    - [ListTelemetryLogsProfilesResponse](#services-v1-ListTelemetryLogsProfilesResponse)
    - [ListTelemetryMetricsGroupsRequest](#services-v1-ListTelemetryMetricsGroupsRequest)
    - [ListTelemetryMetricsGroupsResponse](#services-v1-ListTelemetryMetricsGroupsResponse)
    - [ListTelemetryMetricsProfilesRequest](#services-v1-ListTelemetryMetricsProfilesRequest)
    - [ListTelemetryMetricsProfilesResponse](#services-v1-ListTelemetryMetricsProfilesResponse)
    - [ListWorkloadMembersRequest](#services-v1-ListWorkloadMembersRequest)
    - [ListWorkloadMembersResponse](#services-v1-ListWorkloadMembersResponse)
    - [ListWorkloadsRequest](#services-v1-ListWorkloadsRequest)
    - [ListWorkloadsResponse](#services-v1-ListWorkloadsResponse)
    - [OnboardHostRequest](#services-v1-OnboardHostRequest)
    - [OnboardHostResponse](#services-v1-OnboardHostResponse)
    - [PatchHostRequest](#services-v1-PatchHostRequest)
    - [PatchInstanceRequest](#services-v1-PatchInstanceRequest)
    - [PatchOperatingSystemRequest](#services-v1-PatchOperatingSystemRequest)
    - [PatchProviderRequest](#services-v1-PatchProviderRequest)
    - [PatchRegionRequest](#services-v1-PatchRegionRequest)
    - [PatchRepeatedScheduleRequest](#services-v1-PatchRepeatedScheduleRequest)
    - [PatchSingleScheduleRequest](#services-v1-PatchSingleScheduleRequest)
    - [PatchSiteRequest](#services-v1-PatchSiteRequest)
    - [PatchTelemetryLogsProfileRequest](#services-v1-PatchTelemetryLogsProfileRequest)
    - [PatchTelemetryMetricsProfileRequest](#services-v1-PatchTelemetryMetricsProfileRequest)
    - [PatchWorkloadMemberRequest](#services-v1-PatchWorkloadMemberRequest)
    - [PatchWorkloadRequest](#services-v1-PatchWorkloadRequest)
    - [RegisterHostRequest](#services-v1-RegisterHostRequest)
    - [UpdateHostRequest](#services-v1-UpdateHostRequest)
    - [UpdateInstanceRequest](#services-v1-UpdateInstanceRequest)
    - [UpdateOperatingSystemRequest](#services-v1-UpdateOperatingSystemRequest)
    - [UpdateProviderRequest](#services-v1-UpdateProviderRequest)
    - [UpdateRegionRequest](#services-v1-UpdateRegionRequest)
    - [UpdateRepeatedScheduleRequest](#services-v1-UpdateRepeatedScheduleRequest)
    - [UpdateSingleScheduleRequest](#services-v1-UpdateSingleScheduleRequest)
    - [UpdateSiteRequest](#services-v1-UpdateSiteRequest)
    - [UpdateTelemetryLogsProfileRequest](#services-v1-UpdateTelemetryLogsProfileRequest)
    - [UpdateTelemetryMetricsProfileRequest](#services-v1-UpdateTelemetryMetricsProfileRequest)
    - [UpdateWorkloadMemberRequest](#services-v1-UpdateWorkloadMemberRequest)
    - [UpdateWorkloadRequest](#services-v1-UpdateWorkloadRequest)
  
    - [ListLocationsResponse.ResourceKind](#services-v1-ListLocationsResponse-ResourceKind)
  
    - [HostService](#services-v1-HostService)
    - [InstanceService](#services-v1-InstanceService)
    - [LocalAccountService](#services-v1-LocalAccountService)
    - [LocationService](#services-v1-LocationService)
    - [OperatingSystemService](#services-v1-OperatingSystemService)
    - [ProviderService](#services-v1-ProviderService)
    - [RegionService](#services-v1-RegionService)
    - [ScheduleService](#services-v1-ScheduleService)
    - [SiteService](#services-v1-SiteService)
    - [TelemetryLogsGroupService](#services-v1-TelemetryLogsGroupService)
    - [TelemetryLogsProfileService](#services-v1-TelemetryLogsProfileService)
    - [TelemetryMetricsGroupService](#services-v1-TelemetryMetricsGroupService)
    - [TelemetryMetricsProfileService](#services-v1-TelemetryMetricsProfileService)
    - [WorkloadMemberService](#services-v1-WorkloadMemberService)
    - [WorkloadService](#services-v1-WorkloadService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="resources_common_v1_common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/common/v1/common.proto



<a name="resources-common-v1-MetadataItem"></a>

### MetadataItem
A metadata item, represented by a key:value pair.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  | The metadata key. |
| value | [string](#string) |  | The metadata value. |






<a name="resources-common-v1-Timestamps"></a>

### Timestamps



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time when the resource was created. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time when the resource was last updated. |





 

 

 

 



<a name="resources_provider_v1_provider-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/provider/v1/provider.proto



<a name="resources-provider-v1-ProviderResource"></a>

### ProviderResource
A provider resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated by the inventory on Create. |
| provider_kind | [ProviderKind](#resources-provider-v1-ProviderKind) |  | The provider kind. |
| provider_vendor | [ProviderVendor](#resources-provider-v1-ProviderVendor) |  | The provider vendor. |
| name | [string](#string) |  | The provider resource&#39;s name. |
| api_endpoint | [string](#string) |  | The provider resource&#39;s API endpoint. |
| api_credentials | [string](#string) | repeated | The provider resource&#39;s list of credentials. |
| config | [string](#string) |  | Opaque provider configuration. |
| provider_id | [string](#string) |  | The provider resource&#39;s unique identifier. Alias of resourceId. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 


<a name="resources-provider-v1-ProviderKind"></a>

### ProviderKind
Kind of provider.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PROVIDER_KIND_UNSPECIFIED | 0 |  |
| PROVIDER_KIND_BAREMETAL | 1 |  |



<a name="resources-provider-v1-ProviderVendor"></a>

### ProviderVendor
Vendor of the provider.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PROVIDER_VENDOR_UNSPECIFIED | 0 |  |
| PROVIDER_VENDOR_LENOVO_LXCA | 1 |  |
| PROVIDER_VENDOR_LENOVO_LOCA | 2 |  |


 

 

 



<a name="resources_location_v1_location-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/location/v1/location.proto



<a name="resources-location-v1-RegionResource"></a>

### RegionResource
A region resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by the inventory on Create. |
| name | [string](#string) |  | The user-provided, human-readable name of region |
| parent_region | [RegionResource](#resources-location-v1-RegionResource) |  | The parent Region associated to the Region, when existent. |
| region_id | [string](#string) |  | The Region unique identifier. Alias of resourceId. |
| metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The metadata associated to the Region, represented by a list of key:value pairs. |
| inherited_metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The rendered metadata from the Region parent(s) that can be inherited by the Region, represented by a list of key:value pairs. This field can not be used in filter. |
| total_sites | [int32](#int32) |  | The total number of sites in the region. |
| parent_id | [string](#string) |  | The parent Region unique identifier that the region is associated to, when existent. This field can not be used in filter. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-location-v1-SiteResource"></a>

### SiteResource
A site resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by the inventory on Create. |
| name | [string](#string) |  | The site&#39;s human-readable name. |
| region | [RegionResource](#resources-location-v1-RegionResource) |  | Region this site is located in |
| site_lat | [int32](#int32) |  | The geolocation latitude of the site. Points are represented as latitude-longitude pairs in the E7 representation (degrees are multiplied by 10**7 and rounded to the nearest integer). siteLat must be in the range of &#43;/- 90 degrees. |
| site_lng | [int32](#int32) |  | The geolocation longitude of the site. Points are represented as latitude-longitude pairs in the E7 representation (degrees are multiplied by 10**7 and rounded to the nearest integer). siteLng must be in the range of &#43;/- 180 degrees (inclusive). |
| provider | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) |  | Provider this Site is managed by |
| site_id | [string](#string) |  | The site unique identifier. Alias of resourceId. |
| metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The metadata associated to the Region, represented by a list of key:value pairs. |
| inherited_metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The rendered metadata from the Region parent(s) that can be inherited by the Region, represented by a list of key:value pairs. This field can not be used in filter. |
| region_id | [string](#string) |  | The region&#39;s unique identifier that the site is associated to. This field cannot be used in filter. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 

 

 

 



<a name="resources_os_v1_os-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/os/v1/os.proto



<a name="resources-os-v1-OperatingSystemResource"></a>

### OperatingSystemResource
An OS resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated by inventory on Create. |
| name | [string](#string) |  | The OS resource&#39;s name. |
| architecture | [string](#string) |  | The OS resource&#39;s CPU architecture. |
| kernel_command | [string](#string) |  | The OS resource&#39;s kernel Command Line Options. |
| update_sources | [string](#string) | repeated | The list of OS resource update sources. Should be in &#39;DEB822 Source Format&#39; for Debian style OSs |
| image_url | [string](#string) |  | The URL repository of the OS image. |
| image_id | [string](#string) |  | A unique identifier of the OS image that can be retrieved from the running OS. |
| sha256 | [string](#string) |  | SHA256 checksum of the OS resource in hexadecimal representation. |
| profile_name | [string](#string) |  | Name of an OS profile that the OS resource belongs to. Uniquely identifies a family of OS resources. |
| profile_version | [string](#string) |  | Version of OS profile that the OS resource belongs to. |
| installed_packages | [string](#string) |  | Freeform text, OS-dependent. A list of package names, one per line (newline separated). Must not contain version information. |
| security_feature | [SecurityFeature](#resources-os-v1-SecurityFeature) |  | Indicating if this OS is capable of supporting features like Secure Boot (SB) and Full Disk Encryption (FDE). Immutable after creation. |
| os_type | [OsType](#resources-os-v1-OsType) |  | Indicating the type of OS (for example, mutable or immutable). |
| os_provider | [OsProviderKind](#resources-os-v1-OsProviderKind) |  | Indicating the provider of OS (e.g., Infra or Lenovo). |
| platform_bundle | [string](#string) |  | Opaque JSON field storing references to custom installation script(s) that supplements the base OS with additional OS-level dependencies/configurations. If empty, the default OS installation will be used. |
| os_resource_id | [string](#string) |  | The OS resource&#39;s unique identifier. Alias of resourceId. |
| repo_url | [string](#string) |  | OS image URL. URL of the original installation source. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 


<a name="resources-os-v1-OsProviderKind"></a>

### OsProviderKind
OsProviderKind describes &#34;owner&#34; of the OS, that will drive OS provisioning.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OS_PROVIDER_KIND_UNSPECIFIED | 0 |  |
| OS_PROVIDER_KIND_INFRA | 1 |  |
| OS_PROVIDER_KIND_LENOVO | 2 |  |



<a name="resources-os-v1-OsType"></a>

### OsType
OsType describes type of operating system.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OS_TYPE_UNSPECIFIED | 0 |  |
| OS_TYPE_MUTABLE | 1 |  |
| OS_TYPE_IMMUTABLE | 2 |  |



<a name="resources-os-v1-SecurityFeature"></a>

### SecurityFeature
SecurityFeature describes the security capabilities of a resource.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SECURITY_FEATURE_UNSPECIFIED | 0 |  |
| SECURITY_FEATURE_NONE | 1 |  |
| SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION | 2 |  |


 

 

 



<a name="resources_localaccount_v1_localaccount-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/localaccount/v1/localaccount.proto



<a name="resources-localaccount-v1-LocalAccountResource"></a>

### LocalAccountResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource identifier |
| username | [string](#string) |  | Username provided by admin |
| ssh_key | [string](#string) |  | SSH Public Key of EN |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 

 

 

 



<a name="resources_status_v1_status-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/status/v1/status.proto


 


<a name="resources-status-v1-StatusIndication"></a>

### StatusIndication
The status indicator.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_INDICATION_UNSPECIFIED | 0 |  |
| STATUS_INDICATION_ERROR | 1 |  |
| STATUS_INDICATION_IN_PROGRESS | 2 |  |
| STATUS_INDICATION_IDLE | 3 |  |


 

 

 



<a name="resources_compute_v1_compute-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/compute/v1/compute.proto



<a name="resources-compute-v1-HostResource"></a>

### HostResource
A Host resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated on Create. |
| name | [string](#string) |  | The host name. |
| desired_state | [HostState](#resources-compute-v1-HostState) |  | The desired state of the Host. |
| current_state | [HostState](#resources-compute-v1-HostState) |  | The current state of the Host. |
| site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | The site resource associated with the host. |
| note | [string](#string) |  | The note associated with the host. |
| serial_number | [string](#string) |  | SMBIOS device serial number. |
| uuid | [string](#string) |  | The host UUID identifier; UUID is unique and immutable. |
| memory_bytes | [string](#string) |  | Quantity of memory (RAM) in the system in bytes. |
| cpu_model | [string](#string) |  | CPU model of the Host. |
| cpu_sockets | [uint32](#uint32) |  | Number of physical CPU sockets. |
| cpu_cores | [uint32](#uint32) |  | Number of CPU cores. |
| cpu_capabilities | [string](#string) |  | String list of all CPU capabilities (possibly JSON). |
| cpu_architecture | [string](#string) |  | Architecture of the CPU model, e.g. x86_64. |
| cpu_threads | [uint32](#uint32) |  | Total Number of threads supported by the CPU. |
| cpu_topology | [string](#string) |  | JSON field storing the CPU topology, refer to HDA/HRM docs for the JSON schema. |
| bmc_kind | [BaremetalControllerKind](#resources-compute-v1-BaremetalControllerKind) |  | Kind of BMC. |
| bmc_ip | [string](#string) |  | BMC IP address, such as &#34;192.0.0.1&#34;. |
| hostname | [string](#string) |  | Hostname. |
| product_name | [string](#string) |  | System Product Name. |
| bios_version | [string](#string) |  | BIOS Version. |
| bios_release_date | [string](#string) |  | BIOS Release Date. |
| bios_vendor | [string](#string) |  | BIOS Vendor. |
| host_status | [string](#string) |  | textual message that describes the runtime status of Host. Set by RMs only. |
| host_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of host_status. Set by RMs only. |
| host_status_timestamp | [string](#string) |  | UTC timestamp when host_status was last changed. Set by RMs only. |
| onboarding_status | [string](#string) |  | textual message that describes the onboarding status of Host. Set by RMs only. |
| onboarding_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of onboarding_status. Set by RMs only. |
| onboarding_status_timestamp | [string](#string) |  | UTC timestamp when onboarding_status was last changed. Set by RMs only. |
| registration_status | [string](#string) |  | textual message that describes the onboarding status of Host. Set by RMs only. |
| registration_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of registration_status. Set by RMs only. |
| registration_status_timestamp | [string](#string) |  | UTC timestamp when registration_status was last changed. Set by RMs only. |
| host_storages | [HoststorageResource](#resources-compute-v1-HoststorageResource) | repeated | Back-reference to attached host storage resources. |
| host_nics | [HostnicResource](#resources-compute-v1-HostnicResource) | repeated | Back-reference to attached host NIC resources. |
| host_usbs | [HostusbResource](#resources-compute-v1-HostusbResource) | repeated | Back-reference to attached host USB resources. |
| host_gpus | [HostgpuResource](#resources-compute-v1-HostgpuResource) | repeated | Back-reference to attached host GPU resources. |
| instance | [InstanceResource](#resources-compute-v1-InstanceResource) |  | The instance associated with the host. |
| host_id | [string](#string) |  | Resource ID, generated on Create. |
| site_id | [string](#string) |  | The site where the host is located. |
| metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The metadata associated with the host, represented by a list of key:value pairs. |
| inherited_metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The metadata inherited by the host, represented by a list of key:value pairs, rendered by location and logical structures. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-HostgpuResource"></a>

### HostgpuResource
The set of available host GPU cards.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| pci_id | [string](#string) |  | The GPU device PCI identifier. |
| product | [string](#string) |  | The GPU device model. |
| vendor | [string](#string) |  | The GPU device vendor. |
| description | [string](#string) |  | The human-readable GPU device description. |
| device_name | [string](#string) |  | GPU name as reported by OS. |
| features | [string](#string) |  | The features of this GPU device, comma separated. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-HostnicResource"></a>

### HostnicResource
The set of available host interfaces.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| device_name | [string](#string) |  | The device name (OS provided, like eth0, enp1s0, etc.). |
| pci_identifier | [string](#string) |  | PCI identifier string for this network interface. |
| mac_addr | [string](#string) |  | The interface MAC address. |
| sriov_enabled | [bool](#bool) |  | If the interface has SRIOV enabled. |
| sriov_vfs_num | [uint32](#uint32) |  | The number of VFs currently provisioned on the interface, if SR-IOV is supported. |
| sriov_vfs_total | [uint32](#uint32) |  | The maximum number of VFs the interface supports, if SR-IOV is supported. |
| features | [string](#string) |  | The features of this interface, comma separated. |
| mtu | [uint32](#uint32) |  | Maximum transmission unit of the interface. |
| link_state | [NetworkInterfaceLinkState](#resources-compute-v1-NetworkInterfaceLinkState) |  | Link state of this interface. |
| bmc_interface | [bool](#bool) |  | Whether this is a bmc interface or not. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-HoststorageResource"></a>

### HoststorageResource
The set of available host storage capabilities.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| wwid | [string](#string) |  | The storage device unique identifier. |
| serial | [string](#string) |  | The storage device unique serial number. |
| vendor | [string](#string) |  | The Storage device vendor. |
| model | [string](#string) |  | The storage device model. |
| capacity_bytes | [string](#string) |  | The storage device Capacity (size) in bytes. |
| device_name | [string](#string) |  | The storage device device name (OS provided, like sda, sdb, etc.) |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-HostusbResource"></a>

### HostusbResource
The set of host USB resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| idvendor | [string](#string) |  | Hexadecimal number representing ID of the USB device vendor. |
| idproduct | [string](#string) |  | Hexadecimal number representing ID of the USB device product. |
| bus | [uint32](#uint32) |  | Bus number of device connected with. |
| addr | [uint32](#uint32) |  | USB Device number assigned by OS. |
| class | [string](#string) |  | class defined by USB-IF. |
| serial | [string](#string) |  | Serial number of device. |
| device_name | [string](#string) |  | the OS-provided device name. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-InstanceResource"></a>

### InstanceResource
InstanceResource describes an instantiated OS install, running on either a
host or hypervisor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated on Create. |
| kind | [InstanceKind](#resources-compute-v1-InstanceKind) |  | Kind of resource. Frequently tied to Provider. |
| name | [string](#string) |  | The instance&#39;s human-readable name. |
| desired_state | [InstanceState](#resources-compute-v1-InstanceState) |  | The Instance desired state. |
| current_state | [InstanceState](#resources-compute-v1-InstanceState) |  | The Instance current state. |
| host | [HostResource](#resources-compute-v1-HostResource) |  | Host this Instance is placed on. Only applicable to baremetal instances. |
| desired_os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | OS resource that should be installed to this Instance. |
| current_os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | OS resource that is currently installed for this Instance. |
| security_feature | [resources.os.v1.SecurityFeature](#resources-os-v1-SecurityFeature) |  | Select to enable security features such as Secure Boot (SB) and Full Disk Encryption (FDE). |
| instance_status | [string](#string) |  | textual message that describes the current instance status. Set by RMs only. |
| instance_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of instance_status. Set by RMs only. |
| instance_status_timestamp | [string](#string) |  | UTC timestamp when instance_status was last changed. Set by RMs only. |
| provisioning_status | [string](#string) |  | textual message that describes the provisioning status of Instance. Set by RMs only. |
| provisioning_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of provisioning_status. Set by RMs only. |
| provisioning_status_timestamp | [string](#string) |  | UTC timestamp when provisioning_status was last changed. Set by RMs only. |
| update_status | [string](#string) |  | textual message that describes the update status of Instance. Set by RMs only. |
| update_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of update_status. Set by RMs only. |
| update_status_timestamp | [string](#string) |  | UTC timestamp when update_status was last changed. Set by RMs only. |
| update_status_detail | [string](#string) |  | JSON field storing details of Instance update status. Set by RMs only. Beta, subject to change. |
| trusted_attestation_status | [string](#string) |  | textual message that describes the trusted_attestation status of Instance. Set by RMs only. |
| trusted_attestation_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of trusted_attestation_status. Set by RMs only. |
| trusted_attestation_status_timestamp | [string](#string) |  | UTC timestamp when trusted_attestation_status was last changed. Set by RMs only. |
| workload_members | [WorkloadMember](#resources-compute-v1-WorkloadMember) | repeated | The workload members associated with the instance.

back-reference to the Workload Members associated to this Instance |
| localaccount | [resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) |  | Local Account associated with this Instance |
| instance_id | [string](#string) |  | The instance&#39;s unique identifier. Alias of resourceID. |
| host_id | [string](#string) |  | The host&#39;s unique identifier associated with the instance. |
| os_id | [string](#string) |  | The unique identifier of OS resource that must be installed on the instance. |
| local_account_id | [string](#string) |  | The unique identifier of local account will be associated with the instance. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-WorkloadMember"></a>

### WorkloadMember
Intermediate resource to represent a relation between a workload and a compute resource (i.e., instance).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated by the inventory on Create. |
| kind | [WorkloadMemberKind](#resources-compute-v1-WorkloadMemberKind) |  | The kind of the workload member. |
| workload | [WorkloadResource](#resources-compute-v1-WorkloadResource) |  | The workload resource associated with the workload member. |
| instance | [InstanceResource](#resources-compute-v1-InstanceResource) |  | The instance resource associated with the workload member. |
| workload_member_id | [string](#string) |  | The workload unique identifier. Alias of resourceId. |
| member | [InstanceResource](#resources-compute-v1-InstanceResource) |  | The reference of the Instance member of the workload. |
| workload_id | [string](#string) |  | The workload unique identifier. |
| instance_id | [string](#string) |  | The unique identifier of the instance. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-compute-v1-WorkloadResource"></a>

### WorkloadResource
A generic way to group compute resources to obtain a workload.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by the inventory on Create. |
| kind | [WorkloadKind](#resources-compute-v1-WorkloadKind) |  | Type of workload. |
| name | [string](#string) |  | Human-readable name for the workload. |
| external_id | [string](#string) |  | The ID of the external resource, used to link to resources outside the realm of Edge Infrastructure Manager. |
| status | [string](#string) |  | Human-readable status of the workload. |
| members | [WorkloadMember](#resources-compute-v1-WorkloadMember) | repeated | The members of the workload. |
| workload_id | [string](#string) |  | The workload unique identifier. Alias of resourceId. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 


<a name="resources-compute-v1-BaremetalControllerKind"></a>

### BaremetalControllerKind
The type of BMC.

| Name | Number | Description |
| ---- | ------ | ----------- |
| BAREMETAL_CONTROLLER_KIND_UNSPECIFIED | 0 |  |
| BAREMETAL_CONTROLLER_KIND_NONE | 1 |  |
| BAREMETAL_CONTROLLER_KIND_IPMI | 2 |  |
| BAREMETAL_CONTROLLER_KIND_VPRO | 3 |  |
| BAREMETAL_CONTROLLER_KIND_PDU | 4 |  |



<a name="resources-compute-v1-HostComponentState"></a>

### HostComponentState
The state of the Host component.

| Name | Number | Description |
| ---- | ------ | ----------- |
| HOST_COMPONENT_STATE_UNSPECIFIED | 0 |  |
| HOST_COMPONENT_STATE_ERROR | 1 |  |
| HOST_COMPONENT_STATE_DELETED | 2 |  |
| HOST_COMPONENT_STATE_EXISTS | 3 |  |



<a name="resources-compute-v1-HostState"></a>

### HostState
States of the host.

| Name | Number | Description |
| ---- | ------ | ----------- |
| HOST_STATE_UNSPECIFIED | 0 |  |
| HOST_STATE_DELETING | 1 |  |
| HOST_STATE_DELETED | 2 |  |
| HOST_STATE_ONBOARDED | 3 |  |
| HOST_STATE_UNTRUSTED | 4 |  |
| HOST_STATE_REGISTERED | 5 |  |



<a name="resources-compute-v1-InstanceKind"></a>

### InstanceKind
The Instance kind.

| Name | Number | Description |
| ---- | ------ | ----------- |
| INSTANCE_KIND_UNSPECIFIED | 0 |  |
| INSTANCE_KIND_METAL | 2 | INSTANCE_KIND_VM = 1; |



<a name="resources-compute-v1-InstanceState"></a>

### InstanceState
The Instance States.

| Name | Number | Description |
| ---- | ------ | ----------- |
| INSTANCE_STATE_UNSPECIFIED | 0 | unconfigured |
| INSTANCE_STATE_RUNNING | 1 | OS is Running |
| INSTANCE_STATE_DELETED | 2 | OS should be Deleted |
| INSTANCE_STATE_UNTRUSTED | 3 | OS should not be trusted anymore |



<a name="resources-compute-v1-NetworkInterfaceLinkState"></a>

### NetworkInterfaceLinkState
The state of the network interface.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NETWORK_INTERFACE_LINK_STATE_UNSPECIFIED | 0 |  |
| NETWORK_INTERFACE_LINK_STATE_UP | 1 |  |
| NETWORK_INTERFACE_LINK_STATE_DOWN | 2 |  |



<a name="resources-compute-v1-WorkloadKind"></a>

### WorkloadKind
Represents the type of workload.

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKLOAD_KIND_UNSPECIFIED | 0 | Should never be used. |
| WORKLOAD_KIND_CLUSTER | 1 | Cluster workload. |
| WORKLOAD_KIND_DHCP | 2 | currently unused, but useful to test 2-phase delete. |



<a name="resources-compute-v1-WorkloadMemberKind"></a>

### WorkloadMemberKind
Represents the type of the workload member.

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKLOAD_MEMBER_KIND_UNSPECIFIED | 0 | Should never be used. |
| WORKLOAD_MEMBER_KIND_CLUSTER_NODE | 1 | Node of a cluster workload. |



<a name="resources-compute-v1-WorkloadState"></a>

### WorkloadState
Represents the Workload state, used for both current and desired state.

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKLOAD_STATE_UNSPECIFIED | 0 |  |
| WORKLOAD_STATE_ERROR | 1 |  |
| WORKLOAD_STATE_DELETING | 2 |  |
| WORKLOAD_STATE_DELETED | 3 |  |
| WORKLOAD_STATE_PROVISIONED | 4 |  |


 

 

 



<a name="resources_schedule_v1_schedule-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/schedule/v1/schedule.proto



<a name="resources-schedule-v1-RepeatedScheduleResource"></a>

### RepeatedScheduleResource
A repeated-schedule resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated by the inventory on Create. |
| schedule_status | [ScheduleStatus](#resources-schedule-v1-ScheduleStatus) |  | The schedule status. |
| name | [string](#string) |  | The schedule&#39;s name. |
| target_site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | Resource ID of Site this applies to. |
| target_host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | Resource ID of Host this applies to. |
| target_region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | Resource ID of Region this applies to. |
| duration_seconds | [uint32](#uint32) |  | The duration in seconds of the repeated schedule, per schedule. |
| cron_minutes | [string](#string) |  | cron style minutes (0-59), it can be empty only when used in a Filter. |
| cron_hours | [string](#string) |  | cron style hours (0-23), it can be empty only when used in a Filter |
| cron_day_month | [string](#string) |  | cron style day of month (1-31), it can be empty only when used in a Filter |
| cron_month | [string](#string) |  | cron style month (1-12), it can be empty only when used in a Filter |
| cron_day_week | [string](#string) |  | cron style day of week (0-6), it can be empty only when used in a Filter |
| repeated_schedule_id | [string](#string) |  | The repeated schedule&#39;s unique identifier. Alias of resourceId. |
| target_host_id | [string](#string) |  | The target region ID of the schedule. Only one target can be provided per schedule. This field cannot be used as filter. |
| target_site_id | [string](#string) |  | The target site ID of the schedule. Only one target can be provided per schedule. This field cannot be used as filter. |
| target_region_id | [string](#string) |  | The target region ID of the schedule. Only one target can be provided per schedule. This field cannot be used as filter. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-schedule-v1-SingleScheduleResource"></a>

### SingleScheduleResource
A single schedule resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated by the inventory on Create. |
| schedule_status | [ScheduleStatus](#resources-schedule-v1-ScheduleStatus) |  | The schedule status.

status of one-time-schedule |
| name | [string](#string) |  | The schedule&#39;s name. |
| target_site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | Resource ID of Site this applies to. |
| target_host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | Resource ID of Host this applies to. |
| target_region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | Resource ID of Region this applies to. |
| start_seconds | [uint32](#uint32) |  | The start time in seconds, of the single schedule. |
| end_seconds | [uint32](#uint32) |  | The end time in seconds, of the single schedule. The value of endSeconds must be equal to or bigger than the value of startSeconds. |
| single_schedule_id | [string](#string) |  | The single schedule resource&#39;s unique identifier. Alias of resourceId. |
| target_host_id | [string](#string) |  | The target host ID of the schedule. Only one target can be provided per schedule. This field cannot be used as filter. |
| target_site_id | [string](#string) |  | The target site ID of the schedule. Only one target can be provided per schedule. This field cannot be used as filter. |
| target_region_id | [string](#string) |  | The target region ID of the schedule. Only one target can be provided per schedule. This field cannot be used as filter. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 


<a name="resources-schedule-v1-ScheduleStatus"></a>

### ScheduleStatus
The representation of a schedule&#39;s status.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SCHEDULE_STATUS_UNSPECIFIED | 0 |  |
| SCHEDULE_STATUS_MAINTENANCE | 1 | Generic maintenance.

SCHEDULE_STATUS_SHIPPING = 2; // being shipped/in transit |
| SCHEDULE_STATUS_OS_UPDATE | 3 | for performing OS updates.

SCHEDULE_STATUS_FIRMWARE_UPDATE = 4; // for peforming firmware updates SCHEDULE_STATUS_CLUSTER_UPDATE = 5; // for peforming cluster updates |


 

 

 



<a name="resources_telemetry_v1_telemetry-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/telemetry/v1/telemetry.proto



<a name="resources-telemetry-v1-TelemetryLogsGroupResource"></a>

### TelemetryLogsGroupResource
TelemetryLogsGroupResource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Unique ID of the telemetry group. |
| telemetry_logs_group_id | [string](#string) |  | Unique ID of the telemetry group. Alias of resource_id. |
| name | [string](#string) |  | Human-readable name for the log group. |
| collector_kind | [CollectorKind](#resources-telemetry-v1-CollectorKind) |  | The collector kind. |
| groups | [string](#string) | repeated | A list of log groups to collect. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-telemetry-v1-TelemetryLogsProfileResource"></a>

### TelemetryLogsProfileResource
A telemetry log profile for a hierarchy object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | The ID of the telemetry profile. |
| profile_id | [string](#string) |  | The ID of the telemetry profile. |
| target_instance | [string](#string) |  | The ID of the instance that the telemetry profile is assigned to. Can only be one of targetInstance, targetSite, or targetRegion. |
| target_site | [string](#string) |  | The ID of the site where the telemetry profile is assigned to. Can only be one of targetInstance, targetSite, or targetRegion. |
| target_region | [string](#string) |  | The ID of the region where the telemetry profile is assigned to. Can only be one of targetInstance, targetSite, or targetRegion. |
| log_level | [SeverityLevel](#resources-telemetry-v1-SeverityLevel) |  | The log level og the telemetry profile. |
| logs_group_id | [string](#string) |  | The unique identifier of the telemetry log group. |
| logs_group | [TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) |  | The log group associated with the telemetry profile. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-telemetry-v1-TelemetryMetricsGroupResource"></a>

### TelemetryMetricsGroupResource
TelemetryMetricsGroupResource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Unique ID of the telemetry group. |
| telemetry_metrics_group_id | [string](#string) |  | Unique ID of the telemetry group. Alias of resource_id. |
| name | [string](#string) |  | Human-readable name for the log group. |
| collector_kind | [CollectorKind](#resources-telemetry-v1-CollectorKind) |  | The collector kind. |
| groups | [string](#string) | repeated | A list of log groups to collect. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |






<a name="resources-telemetry-v1-TelemetryMetricsProfileResource"></a>

### TelemetryMetricsProfileResource
A telemetry metric profile for a hierarchy object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | The ID of the telemetry profile. |
| profile_id | [string](#string) |  | The ID of the telemetry profile. |
| target_instance | [string](#string) |  | The ID of the instance that the telemetry profile is assigned to. Can only be one of targetInstance, targetSite, or targetRegion. |
| target_site | [string](#string) |  | The ID of the site where the telemetry profile is assigned to. Can only be one of targetInstance, targetSite, or targetRegion. |
| target_region | [string](#string) |  | The ID of the region where the telemetry profile is assigned to. Can only be one of targetInstance, targetSite, or targetRegion. |
| metrics_interval | [uint32](#uint32) |  | Metric interval (in seconds) for the telemetry profile. This field must only be defined if the type equals to TELEMETRY_CONFIG_KIND_METRICS. |
| metrics_group_id | [string](#string) |  | The unique identifier of the telemetry metric group. |
| metrics_group | [TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) |  | The metric group associated with the telemetry profile. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 


<a name="resources-telemetry-v1-CollectorKind"></a>

### CollectorKind
The collector kind.

| Name | Number | Description |
| ---- | ------ | ----------- |
| COLLECTOR_KIND_UNSPECIFIED | 0 |  |
| COLLECTOR_KIND_HOST | 1 | telemetry data collected from bare-metal host. |
| COLLECTOR_KIND_CLUSTER | 2 | telemetry data collected from Kubernetes cluster. |



<a name="resources-telemetry-v1-SeverityLevel"></a>

### SeverityLevel
Log level used for the telemetry config.
This field must only be defined if kind equals to TELEMETRY_CONFIG_KIND_LOGS.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SEVERITY_LEVEL_UNSPECIFIED | 0 |  |
| SEVERITY_LEVEL_CRITICAL | 1 |  |
| SEVERITY_LEVEL_ERROR | 2 |  |
| SEVERITY_LEVEL_WARN | 3 |  |
| SEVERITY_LEVEL_INFO | 4 |  |
| SEVERITY_LEVEL_DEBUG | 5 |  |



<a name="resources-telemetry-v1-TelemetryResourceKind"></a>

### TelemetryResourceKind
Kind of telemetry collector.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TELEMETRY_RESOURCE_KIND_UNSPECIFIED | 0 |  |
| TELEMETRY_RESOURCE_KIND_METRICS | 1 |  |
| TELEMETRY_RESOURCE_KIND_LOGS | 2 |  |


 

 

 



<a name="services_v1_services-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## services/v1/services.proto



<a name="services-v1-CreateHostRequest"></a>

### CreateHostRequest
Request message for the CreateHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | The host to create. |






<a name="services-v1-CreateHostResponse"></a>

### CreateHostResponse
Response message for the CreateHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | The created host. |






<a name="services-v1-CreateInstanceRequest"></a>

### CreateInstanceRequest
Request message for the CreateInstance method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) |  | The instance to create. |






<a name="services-v1-CreateInstanceResponse"></a>

### CreateInstanceResponse
Response message for the CreateInstance method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) |  | The created instance. |






<a name="services-v1-CreateLocalAccountRequest"></a>

### CreateLocalAccountRequest
Request message for the CreateLocalAccount method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| local_account | [resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) |  | The localaccount to create. |






<a name="services-v1-CreateLocalAccountResponse"></a>

### CreateLocalAccountResponse
Response message for the CreateLocalAccount method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| local_account | [resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) |  | The created localaccount. |






<a name="services-v1-CreateOperatingSystemRequest"></a>

### CreateOperatingSystemRequest
Request message for the CreateOperatingSystem method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | The os to create. |






<a name="services-v1-CreateOperatingSystemResponse"></a>

### CreateOperatingSystemResponse
Response message for the CreateOperatingSystem method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | The created os. |






<a name="services-v1-CreateProviderRequest"></a>

### CreateProviderRequest
Request message for the CreateProvider method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) |  | The provider to create. |






<a name="services-v1-CreateProviderResponse"></a>

### CreateProviderResponse
Response message for the CreateProvider method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) |  | The created provider. |






<a name="services-v1-CreateRegionRequest"></a>

### CreateRegionRequest
Request message for the CreateRegion method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | The region to create. |






<a name="services-v1-CreateRegionResponse"></a>

### CreateRegionResponse
Response message for the CreateRegion method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | The created region. |






<a name="services-v1-CreateRepeatedScheduleRequest"></a>

### CreateRepeatedScheduleRequest
Request message for the CreateRepeatedSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repeated_schedule | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) |  | The repeated_schedule to create. |






<a name="services-v1-CreateRepeatedScheduleResponse"></a>

### CreateRepeatedScheduleResponse
Response message for the CreateRepeatedSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repeated_schedule | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) |  | The created repeated_schedule. |






<a name="services-v1-CreateSingleScheduleRequest"></a>

### CreateSingleScheduleRequest
Request message for the CreateSingleSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_schedule | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) |  | The single_schedule to create. |






<a name="services-v1-CreateSingleScheduleResponse"></a>

### CreateSingleScheduleResponse
Response message for the CreateSingleSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_schedule | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) |  | The created single_schedule. |






<a name="services-v1-CreateSiteRequest"></a>

### CreateSiteRequest
Request message for the CreateSite method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | The site to create. |






<a name="services-v1-CreateSiteResponse"></a>

### CreateSiteResponse
Response message for the CreateSite method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | The created site. |






<a name="services-v1-CreateTelemetryLogsGroupRequest"></a>

### CreateTelemetryLogsGroupRequest
Request message for the CreateTelemetryLogsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_group | [resources.telemetry.v1.TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) |  | The telemetry_logs_group to create. |






<a name="services-v1-CreateTelemetryLogsGroupResponse"></a>

### CreateTelemetryLogsGroupResponse
Response message for the CreateTelemetryLogsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_group | [resources.telemetry.v1.TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) |  | The created telemetry_logs_group. |






<a name="services-v1-CreateTelemetryLogsProfileRequest"></a>

### CreateTelemetryLogsProfileRequest
Request message for the CreateTelemetryLogsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_profile | [resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) |  | The telemetry_logs_profile to create. |






<a name="services-v1-CreateTelemetryLogsProfileResponse"></a>

### CreateTelemetryLogsProfileResponse
Response message for the CreateTelemetryLogsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_profile | [resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) |  | The created telemetry_logs_profile. |






<a name="services-v1-CreateTelemetryMetricsGroupRequest"></a>

### CreateTelemetryMetricsGroupRequest
Request message for the CreateTelemetryMetricsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_group | [resources.telemetry.v1.TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) |  | The telemetry_metrics_group to create. |






<a name="services-v1-CreateTelemetryMetricsGroupResponse"></a>

### CreateTelemetryMetricsGroupResponse
Response message for the CreateTelemetryMetricsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_group | [resources.telemetry.v1.TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) |  | The created telemetry_metrics_group. |






<a name="services-v1-CreateTelemetryMetricsProfileRequest"></a>

### CreateTelemetryMetricsProfileRequest
Request message for the CreateTelemetryMetricsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_profile | [resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) |  | The telemetry_metrics_profile to create. |






<a name="services-v1-CreateTelemetryMetricsProfileResponse"></a>

### CreateTelemetryMetricsProfileResponse
Response message for the CreateTelemetryMetricsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_profile | [resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) |  | The created telemetry_metrics_profile. |






<a name="services-v1-CreateWorkloadMemberRequest"></a>

### CreateWorkloadMemberRequest
Request message for the CreateWorkloadMember method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload_member | [resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) |  | The workload_member to create. |






<a name="services-v1-CreateWorkloadMemberResponse"></a>

### CreateWorkloadMemberResponse
Response message for the CreateWorkloadMember method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload_member | [resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) |  | The created workload_member. |






<a name="services-v1-CreateWorkloadRequest"></a>

### CreateWorkloadRequest
Request message for the CreateWorkload method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload | [resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) |  | The workload to create. |






<a name="services-v1-CreateWorkloadResponse"></a>

### CreateWorkloadResponse
Response message for the CreateWorkload method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload | [resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) |  | The created workload. |






<a name="services-v1-DeleteHostRequest"></a>

### DeleteHostRequest
Request message for DeleteHost.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the host host to be deleted. |






<a name="services-v1-DeleteHostResponse"></a>

### DeleteHostResponse
Reponse message for DeleteHost.






<a name="services-v1-DeleteInstanceRequest"></a>

### DeleteInstanceRequest
Request message for DeleteInstance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the instance instance to be deleted. |






<a name="services-v1-DeleteInstanceResponse"></a>

### DeleteInstanceResponse
Response message for DeleteInstance.






<a name="services-v1-DeleteLocalAccountRequest"></a>

### DeleteLocalAccountRequest
Request message for DeleteLocalAccount.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the localaccount to be deleted. |






<a name="services-v1-DeleteLocalAccountResponse"></a>

### DeleteLocalAccountResponse
Response message for DeleteLocalAccount.






<a name="services-v1-DeleteOperatingSystemRequest"></a>

### DeleteOperatingSystemRequest
Request message for DeleteOperatingSystem.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the os os to be deleted. |






<a name="services-v1-DeleteOperatingSystemResponse"></a>

### DeleteOperatingSystemResponse
Response message for DeleteOperatingSystem.






<a name="services-v1-DeleteProviderRequest"></a>

### DeleteProviderRequest
Request message for DeleteProvider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the provider provider to be deleted. |






<a name="services-v1-DeleteProviderResponse"></a>

### DeleteProviderResponse
Response message for DeleteProvider.






<a name="services-v1-DeleteRegionRequest"></a>

### DeleteRegionRequest
Request message for DeleteRegion.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the region region to be deleted. |






<a name="services-v1-DeleteRegionResponse"></a>

### DeleteRegionResponse
Response message for DeleteRegion.






<a name="services-v1-DeleteRepeatedScheduleRequest"></a>

### DeleteRepeatedScheduleRequest
Request message for DeleteRepeatedSchedule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the repeated_schedule repeated_schedule to be deleted. |






<a name="services-v1-DeleteRepeatedScheduleResponse"></a>

### DeleteRepeatedScheduleResponse
Response message for DeleteRepeatedSchedule.






<a name="services-v1-DeleteSingleScheduleRequest"></a>

### DeleteSingleScheduleRequest
Request message for DeleteSingleSchedule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the single_schedule single_schedule to be deleted. |






<a name="services-v1-DeleteSingleScheduleResponse"></a>

### DeleteSingleScheduleResponse
Response message for DeleteSingleSchedule.






<a name="services-v1-DeleteSiteRequest"></a>

### DeleteSiteRequest
Request message for DeleteSite.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the site site to be deleted. |






<a name="services-v1-DeleteSiteResponse"></a>

### DeleteSiteResponse
Response message for DeleteSite.






<a name="services-v1-DeleteTelemetryLogsGroupRequest"></a>

### DeleteTelemetryLogsGroupRequest
Request message for DeleteTelemetryLogsGroup.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the telemetry_logs_group telemetry_logs_group to be deleted. |






<a name="services-v1-DeleteTelemetryLogsGroupResponse"></a>

### DeleteTelemetryLogsGroupResponse
Response message for DeleteTelemetryLogsGroup.






<a name="services-v1-DeleteTelemetryLogsProfileRequest"></a>

### DeleteTelemetryLogsProfileRequest
Request message for DeleteTelemetryLogsProfile.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the telemetry_logs_profile telemetry_logs_profile to be deleted. |






<a name="services-v1-DeleteTelemetryLogsProfileResponse"></a>

### DeleteTelemetryLogsProfileResponse
Response message for DeleteTelemetryLogsProfile.






<a name="services-v1-DeleteTelemetryMetricsGroupRequest"></a>

### DeleteTelemetryMetricsGroupRequest
Request message for DeleteTelemetryMetricsGroup.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the telemetry_metrics_group telemetry_metrics_group to be deleted. |






<a name="services-v1-DeleteTelemetryMetricsGroupResponse"></a>

### DeleteTelemetryMetricsGroupResponse
Response message for DeleteTelemetryMetricsGroup.






<a name="services-v1-DeleteTelemetryMetricsProfileRequest"></a>

### DeleteTelemetryMetricsProfileRequest
Request message for DeleteTelemetryMetricsProfile.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the telemetry_metrics_profile telemetry_metrics_profile to be deleted. |






<a name="services-v1-DeleteTelemetryMetricsProfileResponse"></a>

### DeleteTelemetryMetricsProfileResponse
Response message for DeleteTelemetryMetricsProfile.






<a name="services-v1-DeleteWorkloadMemberRequest"></a>

### DeleteWorkloadMemberRequest
Request message for DeleteWorkloadMember.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the workload_member workload_member to be deleted. |






<a name="services-v1-DeleteWorkloadMemberResponse"></a>

### DeleteWorkloadMemberResponse
Response message for DeleteWorkloadMember.






<a name="services-v1-DeleteWorkloadRequest"></a>

### DeleteWorkloadRequest
Request message for DeleteWorkload.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the workload workload to be deleted. |






<a name="services-v1-DeleteWorkloadResponse"></a>

### DeleteWorkloadResponse
Response message for DeleteWorkload.






<a name="services-v1-GetHostRequest"></a>

### GetHostRequest
Request message for the GetHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested host. |






<a name="services-v1-GetHostResponse"></a>

### GetHostResponse
Response message for the GetHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | The requested host. |






<a name="services-v1-GetHostSummaryRequest"></a>

### GetHostSummaryRequest
Request the summary of Hosts resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |






<a name="services-v1-GetHostSummaryResponse"></a>

### GetHostSummaryResponse
Summary of the hosts status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| total | [uint32](#uint32) |  | The total number of hosts. |
| error | [uint32](#uint32) |  | The total number of hosts presenting an Error. |
| running | [uint32](#uint32) |  | The total number of hosts in Running state. |
| unallocated | [uint32](#uint32) |  | The total number of hosts without a site. |






<a name="services-v1-GetInstanceRequest"></a>

### GetInstanceRequest
Request message for the GetInstance method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested instance. |






<a name="services-v1-GetInstanceResponse"></a>

### GetInstanceResponse
Response message for the GetInstance method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) |  | The requested instance. |






<a name="services-v1-GetLocalAccountRequest"></a>

### GetLocalAccountRequest
Request message for the GetLocalAccount method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested localaccount. |






<a name="services-v1-GetLocalAccountResponse"></a>

### GetLocalAccountResponse
Response message for the GetLocalAccount method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| local_account | [resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) |  | The requested localaccount. |






<a name="services-v1-GetOperatingSystemRequest"></a>

### GetOperatingSystemRequest
Request message for the GetOperatingSystem method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested os. |






<a name="services-v1-GetOperatingSystemResponse"></a>

### GetOperatingSystemResponse
Response message for the GetOperatingSystem method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | The requested os. |






<a name="services-v1-GetProviderRequest"></a>

### GetProviderRequest
Request message for the GetProvider method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested provider. |






<a name="services-v1-GetProviderResponse"></a>

### GetProviderResponse
Response message for the GetProvider method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) |  | The requested provider. |






<a name="services-v1-GetRegionRequest"></a>

### GetRegionRequest
Request message for the GetRegion method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested region. |






<a name="services-v1-GetRegionResponse"></a>

### GetRegionResponse
Response message for the GetRegion method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | The requested region. |






<a name="services-v1-GetRepeatedScheduleRequest"></a>

### GetRepeatedScheduleRequest
Request message for the GetRepeatedSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested repeated_schedule. |






<a name="services-v1-GetRepeatedScheduleResponse"></a>

### GetRepeatedScheduleResponse
Response message for the GetRepeatedSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repeated_schedule | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) |  | The requested repeated_schedule. |






<a name="services-v1-GetSingleScheduleRequest"></a>

### GetSingleScheduleRequest
Request message for the GetSingleSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested single_schedule. |






<a name="services-v1-GetSingleScheduleResponse"></a>

### GetSingleScheduleResponse
Response message for the GetSingleSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_schedule | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) |  | The requested single_schedule. |






<a name="services-v1-GetSiteRequest"></a>

### GetSiteRequest
Request message for the GetSite method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested site. |






<a name="services-v1-GetSiteResponse"></a>

### GetSiteResponse
Response message for the GetSite method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | The requested site. |






<a name="services-v1-GetTelemetryLogsGroupRequest"></a>

### GetTelemetryLogsGroupRequest
Request message for the GetTelemetryLogsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested telemetry_logs_group. |






<a name="services-v1-GetTelemetryLogsGroupResponse"></a>

### GetTelemetryLogsGroupResponse
Response message for the GetTelemetryLogsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_group | [resources.telemetry.v1.TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) |  | The requested telemetry_logs_group. |






<a name="services-v1-GetTelemetryLogsProfileRequest"></a>

### GetTelemetryLogsProfileRequest
Request message for the GetTelemetryLogsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested telemetry_logs_profile. |






<a name="services-v1-GetTelemetryLogsProfileResponse"></a>

### GetTelemetryLogsProfileResponse
Response message for the GetTelemetryLogsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_profile | [resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) |  | The requested telemetry_logs_profile. |






<a name="services-v1-GetTelemetryMetricsGroupRequest"></a>

### GetTelemetryMetricsGroupRequest
Request message for the GetTelemetryMetricsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested telemetry_metrics_group. |






<a name="services-v1-GetTelemetryMetricsGroupResponse"></a>

### GetTelemetryMetricsGroupResponse
Response message for the GetTelemetryMetricsGroup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_group | [resources.telemetry.v1.TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) |  | The requested telemetry_metrics_group. |






<a name="services-v1-GetTelemetryMetricsProfileRequest"></a>

### GetTelemetryMetricsProfileRequest
Request message for the GetTelemetryMetricsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested telemetry_metrics_profile. |






<a name="services-v1-GetTelemetryMetricsProfileResponse"></a>

### GetTelemetryMetricsProfileResponse
Response message for the GetTelemetryMetricsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_profile | [resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) |  | The requested telemetry_metrics_profile. |






<a name="services-v1-GetWorkloadMemberRequest"></a>

### GetWorkloadMemberRequest
Request message for the GetWorkloadMember method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested workload_member. |






<a name="services-v1-GetWorkloadMemberResponse"></a>

### GetWorkloadMemberResponse
Response message for the GetWorkloadMember method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload_member | [resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) |  | The requested workload_member. |






<a name="services-v1-GetWorkloadRequest"></a>

### GetWorkloadRequest
Request message for the GetWorkload method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the requested workload. |






<a name="services-v1-GetWorkloadResponse"></a>

### GetWorkloadResponse
Response message for the GetWorkload method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload | [resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) |  | The requested workload. |






<a name="services-v1-HostRegister"></a>

### HostRegister
Message to register a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The host name. |
| serial_number | [string](#string) |  | The host serial number. |
| uuid | [string](#string) |  | The host UUID. |
| auto_onboard | [bool](#bool) |  | Flag ot signal to automatically onboard the host. |






<a name="services-v1-InvalidateHostRequest"></a>

### InvalidateHostRequest
Request to invalidate/untrust a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Host resource ID |
| note | [string](#string) |  | user-provided reason for change or a freeform field |






<a name="services-v1-InvalidateHostResponse"></a>

### InvalidateHostResponse
Response message for InvalidateHost.






<a name="services-v1-InvalidateInstanceRequest"></a>

### InvalidateInstanceRequest
Request message for Invalidate Instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Instance resource ID |






<a name="services-v1-InvalidateInstanceResponse"></a>

### InvalidateInstanceResponse
Response message for Invalidate Instance.






<a name="services-v1-ListHostsRequest"></a>

### ListHostsRequest
Request message for the ListHosts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListHostsResponse"></a>

### ListHostsResponse
Response message for the ListHosts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | repeated | Sorted and filtered list of hosts. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListInstancesRequest"></a>

### ListInstancesRequest
Request message for the ListInstances method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListInstancesResponse"></a>

### ListInstancesResponse
Response message for the ListInstances method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instances | [resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) | repeated | Sorted and filtered list of instances. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListLocalAccountsRequest"></a>

### ListLocalAccountsRequest
Request message for the ListLocalAccounts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListLocalAccountsResponse"></a>

### ListLocalAccountsResponse
Response message for the ListLocalAccounts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| local_accounts | [resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) | repeated | Sorted and filtered list of localaccounts. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListLocationsRequest"></a>

### ListLocationsRequest
Request message for the ListLocations method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Filter locations by name |
| show_sites | [bool](#bool) |  | Return site locations |
| show_regions | [bool](#bool) |  | Return region locations |






<a name="services-v1-ListLocationsResponse"></a>

### ListLocationsResponse
Response message for the ListLocations method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nodes | [ListLocationsResponse.LocationNode](#services-v1-ListLocationsResponse-LocationNode) | repeated | Sorted and filtered list of regions. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| output_elements | [int32](#int32) |  | Amount of items in the returned list. |






<a name="services-v1-ListLocationsResponse-LocationNode"></a>

### ListLocationsResponse.LocationNode
A node in the location tree.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | The associated node resource ID, generated by inventory on Create. |
| parent_id | [string](#string) |  | The associated resource ID, of the parent resource of this Location node. In the case of a region, it could be empty or a regionId. In the case of a site, it could be empty or a regionId. |
| name | [string](#string) |  | The node human readable name. |
| type | [ListLocationsResponse.ResourceKind](#services-v1-ListLocationsResponse-ResourceKind) |  | The node type |






<a name="services-v1-ListOperatingSystemsRequest"></a>

### ListOperatingSystemsRequest
Request message for the ListOperatingSystems method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListOperatingSystemsResponse"></a>

### ListOperatingSystemsResponse
Response message for the ListOperatingSystems method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operating_systems | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) | repeated | Sorted and filtered list of oss. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListProvidersRequest"></a>

### ListProvidersRequest
Request message for the ListProviders method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListProvidersResponse"></a>

### ListProvidersResponse
Response message for the ListProviders method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| providers | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) | repeated | Sorted and filtered list of providers. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListRegionsRequest"></a>

### ListRegionsRequest
Request message for the ListRegions method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| show_total_sites | [bool](#bool) |  | Flag to signal if the total amount of site in a region should be returned. |






<a name="services-v1-ListRegionsResponse"></a>

### ListRegionsResponse
Response message for the ListRegions method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regions | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) | repeated | Sorted and filtered list of regions. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListRepeatedSchedulesRequest"></a>

### ListRepeatedSchedulesRequest
Request message for the ListRepeatedSchedules method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| host_id | [string](#string) |  | The host ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified host ID applied to them, i.e., target including the inherited ones (parent site if not null). If null, returns all the schedules without a host ID as target. |
| site_id | [string](#string) |  | The site ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified site ID applied to them, i.e., target including the inherited ones. If null, returns all the schedules without a site ID as target |
| region_id | [string](#string) |  | The region ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified region ID applied to them, i.e., target including the inherited ones (parent region if not null). If null, returns all the schedules without a region ID as target. |
| unix_epoch | [string](#string) |  | Filter based on the timestamp, expected to be UNIX epoch UTC timestamp in seconds. |






<a name="services-v1-ListRepeatedSchedulesResponse"></a>

### ListRepeatedSchedulesResponse
Response message for the ListRepeatedSchedules method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repeated_schedules | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) | repeated | Sorted and filtered list of repeated_schedules. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListSchedulesRequest"></a>

### ListSchedulesRequest
Request message for the ListSchedules method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| host_id | [string](#string) |  | The host ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified host ID applied to them, i.e., target including the inherited ones (parent site if not null). If null, returns all the schedules without a host ID as target. |
| site_id | [string](#string) |  | The site ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified site ID applied to them, i.e., target including the inherited ones. If null, returns all the schedules without a site ID as target |
| region_id | [string](#string) |  | The region ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified region ID applied to them, i.e., target including the inherited ones (parent region if not null). If null, returns all the schedules without a region ID as target. |
| unix_epoch | [string](#string) |  | Filter based on the timestamp, expected to be UNIX epoch UTC timestamp in seconds. |






<a name="services-v1-ListSchedulesResponse"></a>

### ListSchedulesResponse
Response message for the ListSchedulesResponse method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_schedules | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) | repeated | Sorted and filtered list of single_schedules. |
| repeated_schedules | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) | repeated | Sorted and filtered list of repeated_schedules. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListSingleSchedulesRequest"></a>

### ListSingleSchedulesRequest
Request message for the ListSingleSchedules method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| host_id | [string](#string) |  | The host ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified host ID applied to them, i.e., target including the inherited ones (parent site if not null). If null, returns all the schedules without a host ID as target. |
| site_id | [string](#string) |  | The site ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified site ID applied to them, i.e., target including the inherited ones. If null, returns all the schedules without a site ID as target |
| region_id | [string](#string) |  | The region ID target of the schedules. If not specified, returns all schedules (given the other query params). If specified, returns the schedules that have the specified region ID applied to them, i.e., target including the inherited ones (parent region if not null). If null, returns all the schedules without a region ID as target. |
| unix_epoch | [string](#string) |  | Filter based on the timestamp, expected to be UNIX epoch UTC timestamp in seconds. |






<a name="services-v1-ListSingleSchedulesResponse"></a>

### ListSingleSchedulesResponse
Response message for the ListSingleSchedules method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| single_schedules | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) | repeated | Sorted and filtered list of single_schedules. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListSitesRequest"></a>

### ListSitesRequest
Request message for the ListSites method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListSitesResponse"></a>

### ListSitesResponse
Response message for the ListSites method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sites | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) | repeated | Sorted and filtered list of sites. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListTelemetryLogsGroupsRequest"></a>

### ListTelemetryLogsGroupsRequest
Request message for the ListTelemetryLogsGroups method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |






<a name="services-v1-ListTelemetryLogsGroupsResponse"></a>

### ListTelemetryLogsGroupsResponse
Response message for the ListTelemetryLogsGroups method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_groups | [resources.telemetry.v1.TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) | repeated | Sorted and filtered list of telemetry_logs_groups. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListTelemetryLogsProfilesRequest"></a>

### ListTelemetryLogsProfilesRequest
Request message for the ListTelemetryLogsProfiles method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| instance_id | [string](#string) |  | Returns only the telemetry profiles that are assigned with the given instance identifier. |
| site_id | [string](#string) |  | Returns only the telemetry profiles that are assigned with the given siteID. |
| region_id | [string](#string) |  | Returns only the telemetry profiles that are assigned with the given regionID. |
| show_inherited | [bool](#bool) |  | Indicates if listed telemetry profiles should be extended with telemetry profiles rendered from hierarchy. This flag is only used along with one of siteId, regionId or instanceId. If siteId, regionId or instanceId are not set, this flag is ignored. |






<a name="services-v1-ListTelemetryLogsProfilesResponse"></a>

### ListTelemetryLogsProfilesResponse
Response message for the ListTelemetryLogsProfiles method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_logs_profiles | [resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) | repeated | Sorted and filtered list of telemetry_logs_profiles. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListTelemetryMetricsGroupsRequest"></a>

### ListTelemetryMetricsGroupsRequest
Request message for the ListTelemetryMetricsGroups method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |






<a name="services-v1-ListTelemetryMetricsGroupsResponse"></a>

### ListTelemetryMetricsGroupsResponse
Response message for the ListTelemetryMetricsGroups method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_groups | [resources.telemetry.v1.TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) | repeated | Sorted and filtered list of telemetry_metrics_groups. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListTelemetryMetricsProfilesRequest"></a>

### ListTelemetryMetricsProfilesRequest
Request message for the ListTelemetryMetricsProfiles method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| instance_id | [string](#string) |  | Returns only the telemetry profiles that are assigned with the given instance identifier. |
| site_id | [string](#string) |  | Returns only the telemetry profiles that are assigned with the given siteID. |
| region_id | [string](#string) |  | Returns only the telemetry profiles that are assigned with the given regionID. |
| show_inherited | [bool](#bool) |  | Indicates if listed telemetry profiles should be extended with telemetry profiles rendered from hierarchy. This flag is only used along with one of siteId, regionId or instanceId. If siteId, regionId or instanceId are not set, this flag is ignored. |






<a name="services-v1-ListTelemetryMetricsProfilesResponse"></a>

### ListTelemetryMetricsProfilesResponse
Response message for the ListTelemetryMetricsProfiles method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_metrics_profiles | [resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) | repeated | Sorted and filtered list of telemetry_metrics_profiles. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListWorkloadMembersRequest"></a>

### ListWorkloadMembersRequest
Request message for the ListWorkloadMembers method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListWorkloadMembersResponse"></a>

### ListWorkloadMembersResponse
Response message for the ListWorkloadMembers method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workload_members | [resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) | repeated | Sorted and filtered list of workload_members. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-ListWorkloadsRequest"></a>

### ListWorkloadsRequest
Request message for the ListWorkloads method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [int32](#int32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [int32](#int32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListWorkloadsResponse"></a>

### ListWorkloadsResponse
Response message for the ListWorkloads method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloads | [resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) | repeated | Sorted and filtered list of workloads. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-OnboardHostRequest"></a>

### OnboardHostRequest
Request to onboard a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Host resource ID |






<a name="services-v1-OnboardHostResponse"></a>

### OnboardHostResponse
Response of a Host Register request.






<a name="services-v1-PatchHostRequest"></a>

### PatchHostRequest
Request message for the PatchHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | Updated values for the host. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of host. |






<a name="services-v1-PatchInstanceRequest"></a>

### PatchInstanceRequest
Request message for the PatchInstance method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| instance | [resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) |  | Updated values for the instance. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of instance. |






<a name="services-v1-PatchOperatingSystemRequest"></a>

### PatchOperatingSystemRequest
Request message for the PatchOperatingSystem method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | Updated values for the os. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of os. |






<a name="services-v1-PatchProviderRequest"></a>

### PatchProviderRequest
Request message for the PatchProvider method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| provider | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) |  | Updated values for the provider. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of provider. |






<a name="services-v1-PatchRegionRequest"></a>

### PatchRegionRequest
Request message for the PatchRegion method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | Updated values for the region. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of region. |






<a name="services-v1-PatchRepeatedScheduleRequest"></a>

### PatchRepeatedScheduleRequest
Request message for the PatchRepeatedSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| repeated_schedule | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) |  | Updated values for the repeated_schedule. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of repeated_schedule. |






<a name="services-v1-PatchSingleScheduleRequest"></a>

### PatchSingleScheduleRequest
Request message for the PatchSingleSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| single_schedule | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) |  | Updated values for the single_schedule. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of single_schedule. |






<a name="services-v1-PatchSiteRequest"></a>

### PatchSiteRequest
Request message for the PatchSite method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | Updated values for the site. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of site. |






<a name="services-v1-PatchTelemetryLogsProfileRequest"></a>

### PatchTelemetryLogsProfileRequest
Request message for the PatchTelemetryLogs method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| telemetry_logs_profile | [resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) |  | Updated values for the telemetry_logs_profile. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of telemetry_logs_profile. |






<a name="services-v1-PatchTelemetryMetricsProfileRequest"></a>

### PatchTelemetryMetricsProfileRequest
Request message for the PatchTelemetryMetricsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| telemetry_metrics_profile | [resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) |  | Updated values for the telemetry_metrics_profile. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of telemetry_metrics_profile. |






<a name="services-v1-PatchWorkloadMemberRequest"></a>

### PatchWorkloadMemberRequest
Request message for the PatchWorkloadMember method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| workload_member | [resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) |  | Updated values for the workload_member. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of workload_member. |






<a name="services-v1-PatchWorkloadRequest"></a>

### PatchWorkloadRequest
Request message for the PatchWorkload method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| workload | [resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) |  | Updated values for the workload. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of workload. |






<a name="services-v1-RegisterHostRequest"></a>

### RegisterHostRequest
Request to register a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| host | [HostRegister](#services-v1-HostRegister) |  |  |






<a name="services-v1-UpdateHostRequest"></a>

### UpdateHostRequest
Request message for the UpdateHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the host host to be updated. |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | Updated values for the host. |






<a name="services-v1-UpdateInstanceRequest"></a>

### UpdateInstanceRequest
Request message for the UpdateInstance method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | ID of the resource to be updated. |
| instance | [resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) |  | Updated values for the instance. |






<a name="services-v1-UpdateOperatingSystemRequest"></a>

### UpdateOperatingSystemRequest
Request message for the UpdateOperatingSystem method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the os os to be updated. |
| os | [resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) |  | Updated values for the os. |






<a name="services-v1-UpdateProviderRequest"></a>

### UpdateProviderRequest
Request message for the UpdateProvider method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the provider provider to be updated. |
| provider | [resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) |  | Updated values for the provider. |






<a name="services-v1-UpdateRegionRequest"></a>

### UpdateRegionRequest
Request message for the UpdateRegion method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the region region to be updated. |
| region | [resources.location.v1.RegionResource](#resources-location-v1-RegionResource) |  | Updated values for the region. |






<a name="services-v1-UpdateRepeatedScheduleRequest"></a>

### UpdateRepeatedScheduleRequest
Request message for the UpdateRepeatedSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the repeated_schedule repeated_schedule to be updated. |
| repeated_schedule | [resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) |  | Updated values for the repeated_schedule. |






<a name="services-v1-UpdateSingleScheduleRequest"></a>

### UpdateSingleScheduleRequest
Request message for the UpdateSingleSchedule method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the single_schedule single_schedule to be updated. |
| single_schedule | [resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) |  | Updated values for the single_schedule. |






<a name="services-v1-UpdateSiteRequest"></a>

### UpdateSiteRequest
Request message for the UpdateSite method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the site site to be updated. |
| site | [resources.location.v1.SiteResource](#resources-location-v1-SiteResource) |  | Updated values for the site. |






<a name="services-v1-UpdateTelemetryLogsProfileRequest"></a>

### UpdateTelemetryLogsProfileRequest
Request message for the UpdateTelemetryLogsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the telemetry_logs_profile telemetry_logs_profile to be updated. |
| telemetry_logs_profile | [resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) |  | Updated values for the telemetry_logs_profile. |






<a name="services-v1-UpdateTelemetryMetricsProfileRequest"></a>

### UpdateTelemetryMetricsProfileRequest
Request message for the UpdateTelemetryMetricsProfile method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the telemetry_metrics_profile telemetry_metrics_profile to be updated. |
| telemetry_metrics_profile | [resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) |  | Updated values for the telemetry_metrics_profile. |






<a name="services-v1-UpdateWorkloadMemberRequest"></a>

### UpdateWorkloadMemberRequest
Request message for the UpdateWorkloadMember method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the workload_member workload_member to be updated. |
| workload_member | [resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) |  | Updated values for the workload_member. |






<a name="services-v1-UpdateWorkloadRequest"></a>

### UpdateWorkloadRequest
Request message for the UpdateWorkload method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Name of the workload workload to be updated. |
| workload | [resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) |  | Updated values for the workload. |





 


<a name="services-v1-ListLocationsResponse-ResourceKind"></a>

### ListLocationsResponse.ResourceKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESOURCE_KIND_UNSPECIFIED | 0 |  |
| RESOURCE_KIND_REGION | 1 |  |
| RESOURCE_KIND_SITE | 2 |  |


 

 


<a name="services-v1-HostService"></a>

### HostService
Host.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetHostsSummary | [GetHostSummaryRequest](#services-v1-GetHostSummaryRequest) | [GetHostSummaryResponse](#services-v1-GetHostSummaryResponse) | Get a specific host. |
| CreateHost | [CreateHostRequest](#services-v1-CreateHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Create a host. |
| ListHosts | [ListHostsRequest](#services-v1-ListHostsRequest) | [ListHostsResponse](#services-v1-ListHostsResponse) | Get a list of hosts. |
| GetHost | [GetHostRequest](#services-v1-GetHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Get a specific host. |
| UpdateHost | [UpdateHostRequest](#services-v1-UpdateHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Update a host. |
| PatchHost | [PatchHostRequest](#services-v1-PatchHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Patch a host. |
| DeleteHost | [DeleteHostRequest](#services-v1-DeleteHostRequest) | [DeleteHostResponse](#services-v1-DeleteHostResponse) | Delete a host. |
| InvalidateHost | [InvalidateHostRequest](#services-v1-InvalidateHostRequest) | [InvalidateHostResponse](#services-v1-InvalidateHostResponse) | Invalidate a host. |
| RegisterHost | [RegisterHostRequest](#services-v1-RegisterHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Register a host. |
| RegisterUpdateHost | [RegisterHostRequest](#services-v1-RegisterHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Update a host register. |
| OnboardHost | [OnboardHostRequest](#services-v1-OnboardHostRequest) | [OnboardHostResponse](#services-v1-OnboardHostResponse) | Onboard a host. |


<a name="services-v1-InstanceService"></a>

### InstanceService
Instance.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateInstance | [CreateInstanceRequest](#services-v1-CreateInstanceRequest) | [.resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) | Create a instance. |
| ListInstances | [ListInstancesRequest](#services-v1-ListInstancesRequest) | [ListInstancesResponse](#services-v1-ListInstancesResponse) | Get a list of instances. |
| GetInstance | [GetInstanceRequest](#services-v1-GetInstanceRequest) | [.resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) | Get a specific instance. |
| UpdateInstance | [UpdateInstanceRequest](#services-v1-UpdateInstanceRequest) | [.resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) | Update a instance. |
| PatchInstance | [PatchInstanceRequest](#services-v1-PatchInstanceRequest) | [.resources.compute.v1.InstanceResource](#resources-compute-v1-InstanceResource) | Patch a instance. |
| DeleteInstance | [DeleteInstanceRequest](#services-v1-DeleteInstanceRequest) | [DeleteInstanceResponse](#services-v1-DeleteInstanceResponse) | Delete a instance. |
| InvalidateInstance | [InvalidateInstanceRequest](#services-v1-InvalidateInstanceRequest) | [InvalidateInstanceResponse](#services-v1-InvalidateInstanceResponse) | Invalidate a instance. |


<a name="services-v1-LocalAccountService"></a>

### LocalAccountService
LocalAccount.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateLocalAccount | [CreateLocalAccountRequest](#services-v1-CreateLocalAccountRequest) | [.resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) | Create a localAccount. |
| ListLocalAccounts | [ListLocalAccountsRequest](#services-v1-ListLocalAccountsRequest) | [ListLocalAccountsResponse](#services-v1-ListLocalAccountsResponse) | Get a list of providers. |
| GetLocalAccount | [GetLocalAccountRequest](#services-v1-GetLocalAccountRequest) | [.resources.localaccount.v1.LocalAccountResource](#resources-localaccount-v1-LocalAccountResource) | Get a specific provider. |
| DeleteLocalAccount | [DeleteLocalAccountRequest](#services-v1-DeleteLocalAccountRequest) | [DeleteLocalAccountResponse](#services-v1-DeleteLocalAccountResponse) | Delete a provider. |


<a name="services-v1-LocationService"></a>

### LocationService
Location.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListLocations | [ListLocationsRequest](#services-v1-ListLocationsRequest) | [ListLocationsResponse](#services-v1-ListLocationsResponse) | Get a list of locations. |


<a name="services-v1-OperatingSystemService"></a>

### OperatingSystemService
OperatingSystem.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateOperatingSystem | [CreateOperatingSystemRequest](#services-v1-CreateOperatingSystemRequest) | [.resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) | Create an OS |
| ListOperatingSystems | [ListOperatingSystemsRequest](#services-v1-ListOperatingSystemsRequest) | [ListOperatingSystemsResponse](#services-v1-ListOperatingSystemsResponse) | Get a list of OSs. |
| GetOperatingSystem | [GetOperatingSystemRequest](#services-v1-GetOperatingSystemRequest) | [.resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) | Get a specific OS. |
| UpdateOperatingSystem | [UpdateOperatingSystemRequest](#services-v1-UpdateOperatingSystemRequest) | [.resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) | Update an OS. |
| PatchOperatingSystem | [PatchOperatingSystemRequest](#services-v1-PatchOperatingSystemRequest) | [.resources.os.v1.OperatingSystemResource](#resources-os-v1-OperatingSystemResource) | Patch an OS. |
| DeleteOperatingSystem | [DeleteOperatingSystemRequest](#services-v1-DeleteOperatingSystemRequest) | [DeleteOperatingSystemResponse](#services-v1-DeleteOperatingSystemResponse) | Delete an OS. |


<a name="services-v1-ProviderService"></a>

### ProviderService
Provider.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateProvider | [CreateProviderRequest](#services-v1-CreateProviderRequest) | [.resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) | Create a provider. |
| ListProviders | [ListProvidersRequest](#services-v1-ListProvidersRequest) | [ListProvidersResponse](#services-v1-ListProvidersResponse) | Get a list of providers. |
| GetProvider | [GetProviderRequest](#services-v1-GetProviderRequest) | [.resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) | Get a specific provider. |
| UpdateProvider | [UpdateProviderRequest](#services-v1-UpdateProviderRequest) | [.resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) | Update a provider. |
| PatchProvider | [PatchProviderRequest](#services-v1-PatchProviderRequest) | [.resources.provider.v1.ProviderResource](#resources-provider-v1-ProviderResource) | Patch a provider. |
| DeleteProvider | [DeleteProviderRequest](#services-v1-DeleteProviderRequest) | [DeleteProviderResponse](#services-v1-DeleteProviderResponse) | Delete a provider. |


<a name="services-v1-RegionService"></a>

### RegionService
Region.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateRegion | [CreateRegionRequest](#services-v1-CreateRegionRequest) | [.resources.location.v1.RegionResource](#resources-location-v1-RegionResource) | Create a region. |
| ListRegions | [ListRegionsRequest](#services-v1-ListRegionsRequest) | [ListRegionsResponse](#services-v1-ListRegionsResponse) | Get a list of regions. |
| GetRegion | [GetRegionRequest](#services-v1-GetRegionRequest) | [.resources.location.v1.RegionResource](#resources-location-v1-RegionResource) | Get a specific region. |
| UpdateRegion | [UpdateRegionRequest](#services-v1-UpdateRegionRequest) | [.resources.location.v1.RegionResource](#resources-location-v1-RegionResource) | Update a region. |
| PatchRegion | [PatchRegionRequest](#services-v1-PatchRegionRequest) | [.resources.location.v1.RegionResource](#resources-location-v1-RegionResource) | Patch a region. |
| DeleteRegion | [DeleteRegionRequest](#services-v1-DeleteRegionRequest) | [DeleteRegionResponse](#services-v1-DeleteRegionResponse) | Delete a region. |


<a name="services-v1-ScheduleService"></a>

### ScheduleService
Schedules.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListSchedules | [ListSchedulesRequest](#services-v1-ListSchedulesRequest) | [ListSchedulesResponse](#services-v1-ListSchedulesResponse) | Get a list of schedules (single/repeated). |
| CreateSingleSchedule | [CreateSingleScheduleRequest](#services-v1-CreateSingleScheduleRequest) | [.resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) | Create a single_schedule. |
| ListSingleSchedules | [ListSingleSchedulesRequest](#services-v1-ListSingleSchedulesRequest) | [ListSingleSchedulesResponse](#services-v1-ListSingleSchedulesResponse) | Get a list of singleSchedules. |
| GetSingleSchedule | [GetSingleScheduleRequest](#services-v1-GetSingleScheduleRequest) | [.resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) | Get a specific single_schedule. |
| UpdateSingleSchedule | [UpdateSingleScheduleRequest](#services-v1-UpdateSingleScheduleRequest) | [.resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) | Update a single_schedule. |
| PatchSingleSchedule | [PatchSingleScheduleRequest](#services-v1-PatchSingleScheduleRequest) | [.resources.schedule.v1.SingleScheduleResource](#resources-schedule-v1-SingleScheduleResource) | Patch a single_schedule. |
| DeleteSingleSchedule | [DeleteSingleScheduleRequest](#services-v1-DeleteSingleScheduleRequest) | [DeleteSingleScheduleResponse](#services-v1-DeleteSingleScheduleResponse) | Delete a single_schedule. |
| CreateRepeatedSchedule | [CreateRepeatedScheduleRequest](#services-v1-CreateRepeatedScheduleRequest) | [.resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) | Create a repeated_schedule. |
| ListRepeatedSchedules | [ListRepeatedSchedulesRequest](#services-v1-ListRepeatedSchedulesRequest) | [ListRepeatedSchedulesResponse](#services-v1-ListRepeatedSchedulesResponse) | Get a list of repeatedSchedules. |
| GetRepeatedSchedule | [GetRepeatedScheduleRequest](#services-v1-GetRepeatedScheduleRequest) | [.resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) | Get a specific repeated_schedule. |
| UpdateRepeatedSchedule | [UpdateRepeatedScheduleRequest](#services-v1-UpdateRepeatedScheduleRequest) | [.resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) | Update a repeated_schedule. |
| PatchRepeatedSchedule | [PatchRepeatedScheduleRequest](#services-v1-PatchRepeatedScheduleRequest) | [.resources.schedule.v1.RepeatedScheduleResource](#resources-schedule-v1-RepeatedScheduleResource) | Patch a repeated_schedule. |
| DeleteRepeatedSchedule | [DeleteRepeatedScheduleRequest](#services-v1-DeleteRepeatedScheduleRequest) | [DeleteRepeatedScheduleResponse](#services-v1-DeleteRepeatedScheduleResponse) | Delete a repeated_schedule. |


<a name="services-v1-SiteService"></a>

### SiteService
Site.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateSite | [CreateSiteRequest](#services-v1-CreateSiteRequest) | [.resources.location.v1.SiteResource](#resources-location-v1-SiteResource) | Create a site. |
| ListSites | [ListSitesRequest](#services-v1-ListSitesRequest) | [ListSitesResponse](#services-v1-ListSitesResponse) | Get a list of sites. |
| GetSite | [GetSiteRequest](#services-v1-GetSiteRequest) | [.resources.location.v1.SiteResource](#resources-location-v1-SiteResource) | Get a specific site. |
| UpdateSite | [UpdateSiteRequest](#services-v1-UpdateSiteRequest) | [.resources.location.v1.SiteResource](#resources-location-v1-SiteResource) | Update a site. |
| PatchSite | [PatchSiteRequest](#services-v1-PatchSiteRequest) | [.resources.location.v1.SiteResource](#resources-location-v1-SiteResource) | Patch a site. |
| DeleteSite | [DeleteSiteRequest](#services-v1-DeleteSiteRequest) | [DeleteSiteResponse](#services-v1-DeleteSiteResponse) | Delete a site. |


<a name="services-v1-TelemetryLogsGroupService"></a>

### TelemetryLogsGroupService
TelemetryLogsGroup.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateTelemetryLogsGroup | [CreateTelemetryLogsGroupRequest](#services-v1-CreateTelemetryLogsGroupRequest) | [.resources.telemetry.v1.TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) | Create a telemetry_logs_group. |
| ListTelemetryLogsGroups | [ListTelemetryLogsGroupsRequest](#services-v1-ListTelemetryLogsGroupsRequest) | [ListTelemetryLogsGroupsResponse](#services-v1-ListTelemetryLogsGroupsResponse) | Get a list of telemetry_logs_groups. |
| GetTelemetryLogsGroup | [GetTelemetryLogsGroupRequest](#services-v1-GetTelemetryLogsGroupRequest) | [.resources.telemetry.v1.TelemetryLogsGroupResource](#resources-telemetry-v1-TelemetryLogsGroupResource) | Get a specific telemetry_logs_group. |
| DeleteTelemetryLogsGroup | [DeleteTelemetryLogsGroupRequest](#services-v1-DeleteTelemetryLogsGroupRequest) | [DeleteTelemetryLogsGroupResponse](#services-v1-DeleteTelemetryLogsGroupResponse) | Delete a telemetry_logs_group. |


<a name="services-v1-TelemetryLogsProfileService"></a>

### TelemetryLogsProfileService
TelemetryLogsProfile.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateTelemetryLogsProfile | [CreateTelemetryLogsProfileRequest](#services-v1-CreateTelemetryLogsProfileRequest) | [.resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) | Create a telemetry_logs_profile. |
| ListTelemetryLogsProfiles | [ListTelemetryLogsProfilesRequest](#services-v1-ListTelemetryLogsProfilesRequest) | [ListTelemetryLogsProfilesResponse](#services-v1-ListTelemetryLogsProfilesResponse) | Get a list of telemetryLogsProfiles. |
| GetTelemetryLogsProfile | [GetTelemetryLogsProfileRequest](#services-v1-GetTelemetryLogsProfileRequest) | [.resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) | Get a specific telemetry_logs_profile. |
| UpdateTelemetryLogsProfile | [UpdateTelemetryLogsProfileRequest](#services-v1-UpdateTelemetryLogsProfileRequest) | [.resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) | Update a telemetry_logs_profile. |
| PatchTelemetryLogsProfile | [PatchTelemetryLogsProfileRequest](#services-v1-PatchTelemetryLogsProfileRequest) | [.resources.telemetry.v1.TelemetryLogsProfileResource](#resources-telemetry-v1-TelemetryLogsProfileResource) | Patch a telemetry_logs_profile. |
| DeleteTelemetryLogsProfile | [DeleteTelemetryLogsProfileRequest](#services-v1-DeleteTelemetryLogsProfileRequest) | [DeleteTelemetryLogsProfileResponse](#services-v1-DeleteTelemetryLogsProfileResponse) | Delete a telemetry_logs_profile. |


<a name="services-v1-TelemetryMetricsGroupService"></a>

### TelemetryMetricsGroupService
TelemetryMetricsGroup.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateTelemetryMetricsGroup | [CreateTelemetryMetricsGroupRequest](#services-v1-CreateTelemetryMetricsGroupRequest) | [.resources.telemetry.v1.TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) | Create a telemetry_metrics_group. |
| ListTelemetryMetricsGroups | [ListTelemetryMetricsGroupsRequest](#services-v1-ListTelemetryMetricsGroupsRequest) | [ListTelemetryMetricsGroupsResponse](#services-v1-ListTelemetryMetricsGroupsResponse) | Get a list of telemetryMetricsGroups. |
| GetTelemetryMetricsGroup | [GetTelemetryMetricsGroupRequest](#services-v1-GetTelemetryMetricsGroupRequest) | [.resources.telemetry.v1.TelemetryMetricsGroupResource](#resources-telemetry-v1-TelemetryMetricsGroupResource) | Get a specific telemetry_metrics_group. |
| DeleteTelemetryMetricsGroup | [DeleteTelemetryMetricsGroupRequest](#services-v1-DeleteTelemetryMetricsGroupRequest) | [DeleteTelemetryMetricsGroupResponse](#services-v1-DeleteTelemetryMetricsGroupResponse) | Delete a telemetry_metrics_group. |


<a name="services-v1-TelemetryMetricsProfileService"></a>

### TelemetryMetricsProfileService
TelemetryMetricsProfile.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateTelemetryMetricsProfile | [CreateTelemetryMetricsProfileRequest](#services-v1-CreateTelemetryMetricsProfileRequest) | [.resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) | Create a telemetry_metrics_profile. |
| ListTelemetryMetricsProfiles | [ListTelemetryMetricsProfilesRequest](#services-v1-ListTelemetryMetricsProfilesRequest) | [ListTelemetryMetricsProfilesResponse](#services-v1-ListTelemetryMetricsProfilesResponse) | Get a list of telemetryMetricsProfiles. |
| GetTelemetryMetricsProfile | [GetTelemetryMetricsProfileRequest](#services-v1-GetTelemetryMetricsProfileRequest) | [.resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) | Get a specific telemetry_metrics_profile. |
| UpdateTelemetryMetricsProfile | [UpdateTelemetryMetricsProfileRequest](#services-v1-UpdateTelemetryMetricsProfileRequest) | [.resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) | Update a telemetry_metrics_profile. |
| PatchTelemetryMetricsProfile | [PatchTelemetryMetricsProfileRequest](#services-v1-PatchTelemetryMetricsProfileRequest) | [.resources.telemetry.v1.TelemetryMetricsProfileResource](#resources-telemetry-v1-TelemetryMetricsProfileResource) | Patch a telemetry_metrics_profile. |
| DeleteTelemetryMetricsProfile | [DeleteTelemetryMetricsProfileRequest](#services-v1-DeleteTelemetryMetricsProfileRequest) | [DeleteTelemetryMetricsProfileResponse](#services-v1-DeleteTelemetryMetricsProfileResponse) | Delete a telemetry_metrics_profile. |


<a name="services-v1-WorkloadMemberService"></a>

### WorkloadMemberService
WorkloadMember.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateWorkloadMember | [CreateWorkloadMemberRequest](#services-v1-CreateWorkloadMemberRequest) | [.resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) | Create a workload_member. |
| ListWorkloadMembers | [ListWorkloadMembersRequest](#services-v1-ListWorkloadMembersRequest) | [ListWorkloadMembersResponse](#services-v1-ListWorkloadMembersResponse) | Get a list of workload_members. |
| GetWorkloadMember | [GetWorkloadMemberRequest](#services-v1-GetWorkloadMemberRequest) | [.resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) | Get a specific workload_member. |
| UpdateWorkloadMember | [UpdateWorkloadMemberRequest](#services-v1-UpdateWorkloadMemberRequest) | [.resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) | Update a workload_member. |
| PatchWorkloadMember | [PatchWorkloadMemberRequest](#services-v1-PatchWorkloadMemberRequest) | [.resources.compute.v1.WorkloadMember](#resources-compute-v1-WorkloadMember) | Patch a workload_member. |
| DeleteWorkloadMember | [DeleteWorkloadMemberRequest](#services-v1-DeleteWorkloadMemberRequest) | [DeleteWorkloadMemberResponse](#services-v1-DeleteWorkloadMemberResponse) | Delete a workload_member. |


<a name="services-v1-WorkloadService"></a>

### WorkloadService
Workload.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateWorkload | [CreateWorkloadRequest](#services-v1-CreateWorkloadRequest) | [.resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) | Create a workload. |
| ListWorkloads | [ListWorkloadsRequest](#services-v1-ListWorkloadsRequest) | [ListWorkloadsResponse](#services-v1-ListWorkloadsResponse) | Get a list of workloads. |
| GetWorkload | [GetWorkloadRequest](#services-v1-GetWorkloadRequest) | [.resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) | Get a specific workload. |
| UpdateWorkload | [UpdateWorkloadRequest](#services-v1-UpdateWorkloadRequest) | [.resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) | Update a workload. |
| PatchWorkload | [PatchWorkloadRequest](#services-v1-PatchWorkloadRequest) | [.resources.compute.v1.WorkloadResource](#resources-compute-v1-WorkloadResource) | Patch a workload. |
| DeleteWorkload | [DeleteWorkloadRequest](#services-v1-DeleteWorkloadRequest) | [DeleteWorkloadResponse](#services-v1-DeleteWorkloadResponse) | Delete a workload. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

