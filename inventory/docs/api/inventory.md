# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [localaccount/v1/localaccount.proto](#localaccount_v1_localaccount-proto)
    - [LocalAccountResource](#localaccount-v1-LocalAccountResource)
  
- [ou/v1/ou.proto](#ou_v1_ou-proto)
    - [OuResource](#ou-v1-OuResource)
  
- [provider/v1/provider.proto](#provider_v1_provider-proto)
    - [ProviderResource](#provider-v1-ProviderResource)
  
    - [ProviderKind](#provider-v1-ProviderKind)
    - [ProviderVendor](#provider-v1-ProviderVendor)
  
- [location/v1/location.proto](#location_v1_location-proto)
    - [RegionResource](#location-v1-RegionResource)
    - [SiteResource](#location-v1-SiteResource)
  
- [os/v1/os.proto](#os_v1_os-proto)
    - [OperatingSystemResource](#os-v1-OperatingSystemResource)
  
    - [OsProviderKind](#os-v1-OsProviderKind)
    - [OsType](#os-v1-OsType)
    - [SecurityFeature](#os-v1-SecurityFeature)
  
- [status/v1/status.proto](#status_v1_status-proto)
    - [StatusIndication](#status-v1-StatusIndication)
  
- [compute/v1/compute.proto](#compute_v1_compute-proto)
    - [CustomConfigResource](#compute-v1-CustomConfigResource)
    - [HostResource](#compute-v1-HostResource)
    - [HostgpuResource](#compute-v1-HostgpuResource)
    - [HostnicResource](#compute-v1-HostnicResource)
    - [HoststorageResource](#compute-v1-HoststorageResource)
    - [HostusbResource](#compute-v1-HostusbResource)
    - [InstanceResource](#compute-v1-InstanceResource)
    - [OSUpdatePolicyResource](#compute-v1-OSUpdatePolicyResource)
    - [OSUpdateRunResource](#compute-v1-OSUpdateRunResource)
    - [WorkloadMember](#compute-v1-WorkloadMember)
    - [WorkloadResource](#compute-v1-WorkloadResource)
  
    - [AmtState](#compute-v1-AmtState)
    - [BaremetalControllerKind](#compute-v1-BaremetalControllerKind)
    - [HostComponentState](#compute-v1-HostComponentState)
    - [HostState](#compute-v1-HostState)
    - [InstanceKind](#compute-v1-InstanceKind)
    - [InstanceState](#compute-v1-InstanceState)
    - [NetworkInterfaceLinkState](#compute-v1-NetworkInterfaceLinkState)
    - [PowerCommandPolicy](#compute-v1-PowerCommandPolicy)
    - [PowerState](#compute-v1-PowerState)
    - [UpdatePolicy](#compute-v1-UpdatePolicy)
    - [WorkloadKind](#compute-v1-WorkloadKind)
    - [WorkloadMemberKind](#compute-v1-WorkloadMemberKind)
    - [WorkloadState](#compute-v1-WorkloadState)
  
- [network/v1/network.proto](#network_v1_network-proto)
    - [EndpointResource](#network-v1-EndpointResource)
    - [IPAddressResource](#network-v1-IPAddressResource)
    - [NetlinkResource](#network-v1-NetlinkResource)
    - [NetworkSegment](#network-v1-NetworkSegment)
  
    - [IPAddressConfigMethod](#network-v1-IPAddressConfigMethod)
    - [IPAddressState](#network-v1-IPAddressState)
    - [IPAddressStatus](#network-v1-IPAddressStatus)
    - [NetlinkState](#network-v1-NetlinkState)
  
- [remoteaccess/v1/remoteaccess.proto](#remoteaccess_v1_remoteaccess-proto)
    - [RemoteAccessConfiguration](#remoteaccess-v1-RemoteAccessConfiguration)
  
    - [RemoteAccessState](#remoteaccess-v1-RemoteAccessState)
  
- [schedule/v1/schedule.proto](#schedule_v1_schedule-proto)
    - [RepeatedScheduleResource](#schedule-v1-RepeatedScheduleResource)
    - [SingleScheduleResource](#schedule-v1-SingleScheduleResource)
  
    - [ScheduleStatus](#schedule-v1-ScheduleStatus)
  
- [telemetry/v1/telemetry.proto](#telemetry_v1_telemetry-proto)
    - [TelemetryGroupResource](#telemetry-v1-TelemetryGroupResource)
    - [TelemetryProfile](#telemetry-v1-TelemetryProfile)
  
    - [CollectorKind](#telemetry-v1-CollectorKind)
    - [SeverityLevel](#telemetry-v1-SeverityLevel)
    - [TelemetryResourceKind](#telemetry-v1-TelemetryResourceKind)
  
- [tenant/v1/tenant.proto](#tenant_v1_tenant-proto)
    - [Tenant](#tenant-v1-Tenant)
  
    - [TenantState](#tenant-v1-TenantState)
  
- [inventory/v1/inventory.proto](#inventory_v1_inventory-proto)
    - [ChangeSubscribeEventsRequest](#inventory-v1-ChangeSubscribeEventsRequest)
    - [ChangeSubscribeEventsResponse](#inventory-v1-ChangeSubscribeEventsResponse)
    - [CreateResourceRequest](#inventory-v1-CreateResourceRequest)
    - [DeleteAllResourcesRequest](#inventory-v1-DeleteAllResourcesRequest)
    - [DeleteAllResourcesResponse](#inventory-v1-DeleteAllResourcesResponse)
    - [DeleteResourceRequest](#inventory-v1-DeleteResourceRequest)
    - [DeleteResourceResponse](#inventory-v1-DeleteResourceResponse)
    - [FindResourcesRequest](#inventory-v1-FindResourcesRequest)
    - [FindResourcesResponse](#inventory-v1-FindResourcesResponse)
    - [FindResourcesResponse.ResourceTenantIDCarrier](#inventory-v1-FindResourcesResponse-ResourceTenantIDCarrier)
    - [GetResourceRequest](#inventory-v1-GetResourceRequest)
    - [GetResourceResponse](#inventory-v1-GetResourceResponse)
    - [GetResourceResponse.ResourceMetadata](#inventory-v1-GetResourceResponse-ResourceMetadata)
    - [GetSitesPerRegionRequest](#inventory-v1-GetSitesPerRegionRequest)
    - [GetSitesPerRegionResponse](#inventory-v1-GetSitesPerRegionResponse)
    - [GetSitesPerRegionResponse.Node](#inventory-v1-GetSitesPerRegionResponse-Node)
    - [GetTreeHierarchyRequest](#inventory-v1-GetTreeHierarchyRequest)
    - [GetTreeHierarchyResponse](#inventory-v1-GetTreeHierarchyResponse)
    - [GetTreeHierarchyResponse.Node](#inventory-v1-GetTreeHierarchyResponse-Node)
    - [GetTreeHierarchyResponse.TreeNode](#inventory-v1-GetTreeHierarchyResponse-TreeNode)
    - [HeartbeatRequest](#inventory-v1-HeartbeatRequest)
    - [HeartbeatResponse](#inventory-v1-HeartbeatResponse)
    - [ListInheritedTelemetryProfilesRequest](#inventory-v1-ListInheritedTelemetryProfilesRequest)
    - [ListInheritedTelemetryProfilesRequest.InheritBy](#inventory-v1-ListInheritedTelemetryProfilesRequest-InheritBy)
    - [ListInheritedTelemetryProfilesResponse](#inventory-v1-ListInheritedTelemetryProfilesResponse)
    - [ListResourcesRequest](#inventory-v1-ListResourcesRequest)
    - [ListResourcesResponse](#inventory-v1-ListResourcesResponse)
    - [Resource](#inventory-v1-Resource)
    - [ResourceFilter](#inventory-v1-ResourceFilter)
    - [SubscribeEventsRequest](#inventory-v1-SubscribeEventsRequest)
    - [SubscribeEventsResponse](#inventory-v1-SubscribeEventsResponse)
    - [UpdateResourceRequest](#inventory-v1-UpdateResourceRequest)
  
    - [ClientKind](#inventory-v1-ClientKind)
    - [ResourceKind](#inventory-v1-ResourceKind)
    - [SubscribeEventsResponse.EventKind](#inventory-v1-SubscribeEventsResponse-EventKind)
  
    - [InventoryService](#inventory-v1-InventoryService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="localaccount_v1_localaccount-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## localaccount/v1/localaccount.proto



<a name="localaccount-v1-LocalAccountResource"></a>

### LocalAccountResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource identifier |
| username | [string](#string) |  | Username provided by admin |
| ssh_key | [string](#string) |  | SSH Public Key of EN |
| tenant_id | [string](#string) |  | Tenant Identifier. |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 

 

 

 



<a name="ou_v1_ou-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ou/v1/ou.proto



<a name="ou-v1-OuResource"></a>

### OuResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| name | [string](#string) |  | user-provided, human-readable name of OU |
| ou_kind | [string](#string) |  | kinds like &#34;Organization&#34;, &#34;BU&#34;... |
| parent_ou | [OuResource](#ou-v1-OuResource) |  | Optional parent OU. |
| children | [OuResource](#ou-v1-OuResource) | repeated | References to children OU. |
| metadata | [string](#string) |  | Record metadata with format as json string. Example: [{&#34;key&#34;:&#34;cluster-name&#34;,&#34;value&#34;:&#34;&#34;},{&#34;key&#34;:&#34;app-id&#34;,&#34;value&#34;:&#34;&#34;}] |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 

 

 

 



<a name="provider_v1_provider-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provider/v1/provider.proto



<a name="provider-v1-ProviderResource"></a>

### ProviderResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| provider_kind | [ProviderKind](#provider-v1-ProviderKind) |  | kind and vendor are used to diversify the provider |
| provider_vendor | [ProviderVendor](#provider-v1-ProviderVendor) |  |  |
| name | [string](#string) |  | Provider&#39;s name, unique in tenant context. |
| api_endpoint | [string](#string) |  | URI to contact the provider |
| api_credentials | [string](#string) | repeated | ID of credential in Vault |
| config | [string](#string) |  | Opaque provider configuration. |
| tenant_id | [string](#string) |  | Tenant Identifier. |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="provider-v1-ProviderKind"></a>

### ProviderKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| PROVIDER_KIND_UNSPECIFIED | 0 |  |
| PROVIDER_KIND_BAREMETAL | 1 |  |



<a name="provider-v1-ProviderVendor"></a>

### ProviderVendor


| Name | Number | Description |
| ---- | ------ | ----------- |
| PROVIDER_VENDOR_UNSPECIFIED | 0 |  |
| PROVIDER_VENDOR_LENOVO_LXCA | 1 |  |
| PROVIDER_VENDOR_LENOVO_LOCA | 2 |  |


 

 

 



<a name="location_v1_location-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## location/v1/location.proto



<a name="location-v1-RegionResource"></a>

### RegionResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| name | [string](#string) |  | user-provided, human-readable name of region |
| region_kind | [string](#string) |  | kinds like &#34;Country&#34;, &#34;State&#34; and &#34;City&#34; |
| parent_region | [RegionResource](#location-v1-RegionResource) |  | Optional parent region. |
| children | [RegionResource](#location-v1-RegionResource) | repeated | References to children regions. |
| metadata | [string](#string) |  | Record metadata with format as json string. Example: [{&#34;key&#34;:&#34;cluster-name&#34;,&#34;value&#34;:&#34;&#34;},{&#34;key&#34;:&#34;app-id&#34;,&#34;value&#34;:&#34;&#34;}] |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="location-v1-SiteResource"></a>

### SiteResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| name | [string](#string) |  | user-provided, human-readable name of site |
| region | [RegionResource](#location-v1-RegionResource) |  | Region this site is located in |
| ou | [ou.v1.OuResource](#ou-v1-OuResource) |  | OU this site is part of |
| address | [string](#string) |  |  |
| site_lat | [int32](#int32) |  | latitude |
| site_lng | [int32](#int32) |  | longitude |
| dns_servers | [string](#string) | repeated | list of DNS servers |
| docker_registries | [string](#string) | repeated |  |
| metrics_endpoint | [string](#string) |  |  |
| http_proxy | [string](#string) |  |  |
| https_proxy | [string](#string) |  |  |
| ftp_proxy | [string](#string) |  |  |
| no_proxy | [string](#string) |  |  |
| provider | [provider.v1.ProviderResource](#provider-v1-ProviderResource) |  | Provider this Site is managed by |
| metadata | [string](#string) |  | Record metadata with format as json string. Example: [{&#34;key&#34;:&#34;cluster-name&#34;,&#34;value&#34;:&#34;&#34;},{&#34;key&#34;:&#34;app-id&#34;,&#34;value&#34;:&#34;&#34;}] |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 

 

 

 



<a name="os_v1_os-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## os/v1/os.proto



<a name="os-v1-OperatingSystemResource"></a>

### OperatingSystemResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID of this OperatingSystemResource |
| name | [string](#string) |  | user-provided, human-readable name of OS |
| architecture | [string](#string) |  | CPU architecture supported |
| kernel_command | [string](#string) |  | Kernel Command Line Options. Deprecated in EMF-v3.1, use OSUpdatePolicy. |
| update_sources | [string](#string) | repeated | OS Update Sources. Should be in &#39;DEB822 Source Format&#39; for Debian style OSs. Deprecated in EMF-v3.1, use OSUpdatePolicy. |
| image_url | [string](#string) |  | OS image URL. URL of the original installation source. |
| image_id | [string](#string) |  | OS image ID. This must be a unique identifier of OS image that can be retrieved from running OS. Used by IMMUTABLE only. |
| sha256 | [string](#string) |  | SHA256 checksum of the OS resource in HEX. It&#39;s length is 32 bytes, but string representation of HEX is twice long (64 chars) |
| profile_name | [string](#string) |  | Name of an OS profile that the OS resource belongs to. Uniquely identifies family of OSResources. |
| profile_version | [string](#string) |  | Version of an OS profile that the OS resource belongs to. Along with profile_name uniquely identifies OS resource. |
| installed_packages | [string](#string) |  | Freeform text, OS-dependent. A list of package names, one per line (newline separated). Should not contain version info. Deprecated in EMF-v3.1, use OSUpdatePolicy. |
| installed_packages_url | [string](#string) |  | The URL of the OS manifest which contains install packages details. This will be used to fill the installed_packages field for the advance use case to allow manual creation of OSProfiles when supported from backend. |
| security_feature | [SecurityFeature](#os-v1-SecurityFeature) |  | Indicating if this OS is capable of supporting features like Secure Boot (SB) and Full Disk Encryption (FDE). |
| os_type | [OsType](#os-v1-OsType) |  | Indicating the type of OS (for example, mutable or immutable). |
| os_provider | [OsProviderKind](#os-v1-OsProviderKind) |  | Indicating the provider of OS (e.g., Infra or Lenovo). |
| platform_bundle | [string](#string) |  | An opaque JSON string storing a reference to custom installation script(s) that supplements the base OS with additional OS-level dependencies/configurations. If empty, the default OS installation will be used. |
| description | [string](#string) |  | user-provided, human-readable description of OS |
| metadata | [string](#string) |  | Opaque JSON field storing metadata associated to this OS resource. Expected to be a JSON object with string keys and values, or an empty string. |
| existing_cves_url | [string](#string) |  | URL of the file containing information about the existing CVEs on the Operating System. |
| existing_cves | [string](#string) |  | The CVEs that are currently present on the Operating System, encoded as a JSON list. |
| fixed_cves_url | [string](#string) |  | URL of the file containing information about the CVEs that have been fixed by this OS Resource version. |
| fixed_cves | [string](#string) |  | The CVEs that have been fixed by this OS Resource version, encoded as a JSON list. |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="os-v1-OsProviderKind"></a>

### OsProviderKind
OsProviderKind describes &#34;owner&#34; of the OS, that will drive OS provisioning.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OS_PROVIDER_KIND_UNSPECIFIED | 0 |  |
| OS_PROVIDER_KIND_INFRA | 1 |  |
| OS_PROVIDER_KIND_LENOVO | 2 |  |



<a name="os-v1-OsType"></a>

### OsType
OsType describes type of operating system.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OS_TYPE_UNSPECIFIED | 0 |  |
| OS_TYPE_MUTABLE | 1 |  |
| OS_TYPE_IMMUTABLE | 2 |  |



<a name="os-v1-SecurityFeature"></a>

### SecurityFeature
SecurityFeature describes the security capabilities of a resource.
Due to limitations of the Ent code generator, this enum cannot be a repeated
field in resource messages. Hence, we have to manually list composite options
like SB&#43;FDE.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SECURITY_FEATURE_UNSPECIFIED | 0 |  |
| SECURITY_FEATURE_NONE | 1 |  |
| SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION | 2 |  |


 

 

 



<a name="status_v1_status-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## status/v1/status.proto


 


<a name="status-v1-StatusIndication"></a>

### StatusIndication


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_INDICATION_UNSPECIFIED | 0 |  |
| STATUS_INDICATION_ERROR | 1 |  |
| STATUS_INDICATION_IN_PROGRESS | 2 |  |
| STATUS_INDICATION_IDLE | 3 |  |


 

 

 



<a name="compute_v1_compute-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## compute/v1/compute.proto



<a name="compute-v1-CustomConfigResource"></a>

### CustomConfigResource
---------------------------------------------------------------------------------------------------------------------
--------------------------------------------- Custom Config Resources -----------------------------------------------


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource identifier |
| name | [string](#string) |  | Unique name provided by admin |
| config | [string](#string) |  | Configuration file. |
| description | [string](#string) |  | Config Description |
| tenant_id | [string](#string) |  | Tenant Identifier. |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |
| instances | [InstanceResource](#compute-v1-InstanceResource) | repeated | back-reference to the Instances associated to this Custom Config, we don&#39;t handle setting edges via back-reference and we dont eager load this edge. |






<a name="compute-v1-HostResource"></a>

### HostResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| kind | [string](#string) |  | Kind of resource. Frequently tied to Provider |
| name | [string](#string) |  | user-provided, human-readable name of host |
| desired_state | [HostState](#compute-v1-HostState) |  |  |
| current_state | [HostState](#compute-v1-HostState) |  |  |
| site | [location.v1.SiteResource](#location-v1-SiteResource) |  | Site this VM is located at |
| provider | [provider.v1.ProviderResource](#provider-v1-ProviderResource) |  | Provider this host is onboarded through |
| note | [string](#string) |  | user-provided reason for change or a freeform field |
| hardware_kind | [string](#string) |  | FIXME: add validation rules on the below items

type such as &#34;XSPgen3&#34;, &#34;XDgen2&#34;, &#34;CI7gen12&#34; |
| serial_number | [string](#string) |  | SMBIOS device Serial Number |
| uuid | [string](#string) |  | SMBIOS device UUID. See pages 37-38 of https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.6.0.pdf |
| memory_bytes | [uint64](#uint64) |  | Quantity of memory (RAM) in the system in bytes. |
| cpu_model | [string](#string) |  | CPU model of the Host |
| cpu_sockets | [uint32](#uint32) |  | Number of physical CPU sockets |
| cpu_cores | [uint32](#uint32) |  | Number of CPU cores |
| cpu_capabilities | [string](#string) |  | String list of all CPU capabilities (possibly JSON) |
| cpu_architecture | [string](#string) |  | Architecture of the CPU model, e.g. x86_64 |
| cpu_threads | [uint32](#uint32) |  | Total Number of threads supported by the CPU |
| cpu_topology | [string](#string) |  | JSON field storing the CPU topology, refer to HDA/HRM docs for the JSON schema. |
| mgmt_ip | [string](#string) |  | IP address of management network |
| bmc_kind | [BaremetalControllerKind](#compute-v1-BaremetalControllerKind) |  | Kind of BMC |
| bmc_ip | [string](#string) |  | BMC IP address, such as &#34;192.0.0.1&#34; |
| bmc_username | [string](#string) |  | BMC user name, such as &#34;admin&#34; |
| bmc_password | [string](#string) |  | BMC password, such as &#34;admin&#34; |
| pxe_mac | [string](#string) |  | MAC address for PXE boot |
| hostname | [string](#string) |  | Hostname |
| product_name | [string](#string) |  | System Product Name |
| bios_version | [string](#string) |  | BIOS Version |
| bios_release_date | [string](#string) |  | BIOS Release Date |
| bios_vendor | [string](#string) |  | BIOS Vendor |
| metadata | [string](#string) |  | Record metadata with format as json string. Example: [{&#34;key&#34;:&#34;cluster-name&#34;,&#34;value&#34;:&#34;&#34;},{&#34;key&#34;:&#34;app-id&#34;,&#34;value&#34;:&#34;&#34;}] |
| desired_power_state | [PowerState](#compute-v1-PowerState) |  | Power management related fields

Desired power state of the host |
| current_power_state | [PowerState](#compute-v1-PowerState) |  | Current power state of the host |
| power_status | [string](#string) |  | A group of fields describing the Power status at runtime. The following 3 fields should always be updated in one shot. If power_status is empty during initialization, it is automatically set to a Unknown value.

textual message that describes the runtime status of Host power. Set by DM RM only. |
| power_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates dynamicity of the power_status. Set by DM RM only. |
| power_status_timestamp | [uint64](#uint64) |  | UTC timestamp when power_status was last changed. Set by DM RM only. |
| power_command_policy | [PowerCommandPolicy](#compute-v1-PowerCommandPolicy) |  | Power command policy of the host. By default, it is set to PowerCommandPolicy.POWER_COMMAND_POLICY_ORDERED. |
| power_on_time | [uint64](#uint64) |  | UTC timestamp when the host was powered on. Set by DM RM only. |
| host_status | [string](#string) |  | A group of fields describing the Host runtime status. host_status, host_status_indicator and host_status_timestamp should always be updated in one shot. If host_status is empty during initialization, it is automatically set to a default value.

textual message that describes the runtime status of Host. Set by RMs only. |
| host_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of host_status. Set by RMs only. |
| host_status_timestamp | [uint64](#uint64) |  | UTC timestamp when host_status was last changed. Set by RMs only. |
| onboarding_status | [string](#string) |  | A group of fields describing the Host onboarding status. onboarding_status, onboarding_status_indicator and onboarding_status_timestamp should always be updated in one shot. If onboarding_status is empty during initialization, it is automatically set to a default value.

textual message that describes the onboarding status of Host. Set by RMs only. |
| onboarding_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of onboarding_status. Set by RMs only. |
| onboarding_status_timestamp | [uint64](#uint64) |  | UTC timestamp when onboarding_status was last changed. Set by RMs only. |
| registration_status | [string](#string) |  | A group of fields describing the Host registration status. registration_status, registration_status_indicator and registration_status_timestamp should always be updated in one shot. If registration_status is empty during initialization, it is automatically set to a default value.

textual message that describes the onboarding status of Host. Set by RMs only. |
| registration_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of registration_status. Set by RMs only. |
| registration_status_timestamp | [uint64](#uint64) |  | UTC timestamp when registration_status was last changed. Set by RMs only. |
| host_storages | [HoststorageResource](#compute-v1-HoststorageResource) | repeated | Back-reference to attached host storage resources. This edge is read-only. |
| host_nics | [HostnicResource](#compute-v1-HostnicResource) | repeated | Back-reference to attached host NIC resources. This edge is read-only. |
| host_usbs | [HostusbResource](#compute-v1-HostusbResource) | repeated | Back-reference to attached host USB resources. This edge is read-only. |
| host_gpus | [HostgpuResource](#compute-v1-HostgpuResource) | repeated | Back-reference to attached host GPU resources. This edge is read-only. |
| instance | [InstanceResource](#compute-v1-InstanceResource) |  | back-reference to baremetal Instance associated to this host |
| amt_sku | [string](#string) |  | coming from device introspection |
| desired_amt_state | [AmtState](#compute-v1-AmtState) |  | Desired AMT/vPRO state of the host |
| current_amt_state | [AmtState](#compute-v1-AmtState) |  | Current AMT/vPRO state of the host |
| amt_status | [string](#string) |  | A group of fields describing the AMT status. amt_status, amt_status_indicator and amt_status_timestamp should always be updated in one shot. If amt_status is empty during initialization, it is automatically set to a Unknown value.

textual message that describes the AMT status of Host. Set by DM RM only. |
| amt_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates dynamicity of the amt_status. Set by DM and OM RM only. |
| amt_status_timestamp | [uint64](#uint64) |  | UTC timestamp when amt_status was last changed. Set by DM and OM RM only. |
| lvm_size | [uint32](#uint32) |  | LVM size in GB. |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-HostgpuResource"></a>

### HostgpuResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID |
| host | [HostResource](#compute-v1-HostResource) |  | Host this GPU device is installed in |
| pci_id | [string](#string) |  | The GPU device PCI identifier |
| product | [string](#string) |  | The GPU device model |
| vendor | [string](#string) |  | The GPU device vendor |
| description | [string](#string) |  | The human-readable GPU device description |
| device_name | [string](#string) |  | GPU name as reported by OS |
| features | [string](#string) |  | The features of this GPU device, comma separated |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-HostnicResource"></a>

### HostnicResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID |
| kind | [string](#string) |  | Kind of resource. Frequently tied to Provider |
| provider_status | [string](#string) |  | current status of the resource according to the provider |
| host | [HostResource](#compute-v1-HostResource) |  | The Host where this NIC is installed |
| device_name | [string](#string) |  | FIXME: Better validation on fields below

the device name (OS provided, like eth0, enp1s0, etc.) |
| pci_identifier | [string](#string) |  | PCI identifier string for this network interface |
| mac_addr | [string](#string) |  | MAC address |
| sriov_enabled | [bool](#bool) |  | has SRIOV |
| sriov_vfs_num | [uint32](#uint32) |  | The number of VFs currently provisioned on the interface, if SR-IOV is supported |
| sriov_vfs_total | [uint32](#uint32) |  | The maximum number of VFs the interface supports, if SR-IOV is supported. |
| peer_name | [string](#string) |  | the neighbor device (the other side of the link), collecting via LLDP |
| peer_description | [string](#string) |  | the neighbor device description |
| peer_mac | [string](#string) |  | the neighbor device MAC address |
| peer_mgmt_ip | [string](#string) |  | the neighbor device management IP address |
| peer_port | [string](#string) |  | the neighbor device port number |
| supported_link_mode | [string](#string) |  | the link mode supported by this interface, comma separated |
| advertising_link_mode | [string](#string) |  | the link mode advertising by this interface |
| current_speed_bps | [uint64](#uint64) |  | the current speed of this interface |
| current_duplex | [string](#string) |  | the current duplex of this interface |
| features | [string](#string) |  | the features of this interface, comma separated |
| mtu | [uint32](#uint32) |  | Maximum transmission unit of the interface |
| link_state | [NetworkInterfaceLinkState](#compute-v1-NetworkInterfaceLinkState) |  | link state of this interface |
| bmc_interface | [bool](#bool) |  | whether this is a bmc interface or not |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-HoststorageResource"></a>

### HoststorageResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID |
| kind | [string](#string) |  | Kind of resource. Frequently tied to Provider |
| provider_status | [string](#string) |  | current status of the resource according to the provider |
| host | [HostResource](#compute-v1-HostResource) |  | The Host where this storage device is installed |
| wwid | [string](#string) |  | FIXME: better validation of the below values

The storage device unique identifier. |
| serial | [string](#string) |  | The storage device unique serial number. |
| vendor | [string](#string) |  | The Storage device vendor |
| model | [string](#string) |  | The storage device model string |
| capacity_bytes | [uint64](#uint64) |  | The storage device Capacity (size) in bytes |
| device_name | [string](#string) |  | The storage device device name (OS provided, like sda, sdb, etc.) |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-HostusbResource"></a>

### HostusbResource
A USB resource


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID |
| kind | [string](#string) |  | Kind of resource. |
| host | [HostResource](#compute-v1-HostResource) |  | The Host where this USB device is installed |
| owner_id | [string](#string) |  | VM or container this usb device allocated to |
| idvendor | [string](#string) |  | FIXME: better validation of the below values

Hexadecimal number representing ID of the USB device vendor |
| idproduct | [string](#string) |  | Hexadecimal number representing ID of the USB device product |
| bus | [uint32](#uint32) |  | Bus number of device connected with |
| addr | [uint32](#uint32) |  | USB Device number assigned by OS. |
| class | [string](#string) |  | class defined by USB-IF |
| serial | [string](#string) |  | Serial number of device |
| device_name | [string](#string) |  | the OS-provided device name |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-InstanceResource"></a>

### InstanceResource
InstanceResource describes an instantiated OS install, running on either a
host or hypervisor.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create. |
| kind | [InstanceKind](#compute-v1-InstanceKind) |  | Kind of resource. Frequently tied to Provider. |
| name | [string](#string) |  | user-provided, human-readable name of Instance |
| desired_state | [InstanceState](#compute-v1-InstanceState) |  |  |
| current_state | [InstanceState](#compute-v1-InstanceState) |  |  |
| vm_memory_bytes | [uint64](#uint64) |  | Quantity of memory in the system, in bytes. Only applicable to VM instances. |
| vm_cpu_cores | [uint32](#uint32) |  | Number of CPU cores. Only applicable to VM instances. |
| vm_storage_bytes | [uint64](#uint64) |  | Storage quantity (primary), in bytes. Only applicable to VM instances. |
| host | [HostResource](#compute-v1-HostResource) |  | Host this Instance is placed on. Only applicable to baremetal instances. |
| desired_os | [os.v1.OperatingSystemResource](#os-v1-OperatingSystemResource) |  | Deprecated, use OSUpdatePolicy for driving day2, and os for day0 operations instead. OS resource that should be installed to this Instance. |
| current_os | [os.v1.OperatingSystemResource](#os-v1-OperatingSystemResource) |  | Deprecated, use os field instead. OS resource that is currently installed for this Instance. |
| os | [os.v1.OperatingSystemResource](#os-v1-OperatingSystemResource) |  | OS resource that is installed for this Instance. |
| security_feature | [os.v1.SecurityFeature](#os-v1-SecurityFeature) |  | Select to enable security features such as Secure Boot (SB) and Full Disk Encryption (FDE). |
| instance_status | [string](#string) |  | A group of fields describing the Instance runtime status. instance_status, instance_status_indicator and instance_status_timestamp should always be updated in one shot. If instance_status is empty during initialization, it is automatically set to a default value.

textual message that describes the current instance status. Set by RMs only. |
| instance_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of instance_status. Set by RMs only. |
| instance_status_timestamp | [uint64](#uint64) |  | UTC timestamp when instance_status was last changed. Set by RMs only. |
| provisioning_status | [string](#string) |  | A group of fields describing the Instance provisioning status. provisioning_status, provisioning_status_indicator and provisioning_status_timestamp should always be updated in one shot. If provisioning_status is empty during initialization, it is automatically set to a default value.

textual message that describes the provisioning status of Instance. Set by RMs only. |
| provisioning_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of provisioning_status. Set by RMs only. |
| provisioning_status_timestamp | [uint64](#uint64) |  | UTC timestamp when provisioning_status was last changed. Set by RMs only. |
| update_status | [string](#string) |  | A group of fields describing the Instance update status. update_status, update_status_indicator and update_status_timestamp should always be updated in one shot. update_status_detail should be populated when update status reports update finished successfully or failed. If update_status is empty during initialization, it is automatically set to a default value.

textual message that describes the update status of Instance. Set by RMs only. |
| update_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of update_status. Set by RMs only. |
| update_status_timestamp | [uint64](#uint64) |  | UTC timestamp when update_status was last changed. Set by RMs only. |
| update_status_detail | [string](#string) |  | Deprecated, will be removed in EMF v3.2.0, use OSUpdateRun instead. JSON field storing details of Instance update status. Set by RMs only. Beta, subject to change. |
| trusted_attestation_status | [string](#string) |  | A group of fields describing the Instance trusted_attestation status. trusted_attestation_status, trusted_attestation_status_indicator and trusted_attestation_status_timestamp should always be updated in one shot. If trusted_attestation_status is empty during initialization, it is automatically set to a default value.

textual message that describes the trusted_attestation status of Instance. Set by RMs only. |
| trusted_attestation_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of trusted_attestation_status. Set by RMs only. |
| trusted_attestation_status_timestamp | [uint64](#uint64) |  | UTC timestamp when trusted_attestation_status was last changed. Set by RMs only. |
| workload_members | [WorkloadMember](#compute-v1-WorkloadMember) | repeated | back-reference to the Workload Members associated to this Instance |
| provider | [provider.v1.ProviderResource](#provider-v1-ProviderResource) |  | Provider this Instance is provisioned through |
| localaccount | [localaccount.v1.LocalAccountResource](#localaccount-v1-LocalAccountResource) |  | Local Account associated with this Instance |
| existing_cves | [string](#string) |  | The CVEs that are currently present on the Instance, encoded as a JSON list. |
| os_update_policy | [OSUpdatePolicyResource](#compute-v1-OSUpdatePolicyResource) |  | OS Update Policy associated with this Instance. |
| runtime_packages | [string](#string) |  | The packages available on the Instance at runtime, represented as a JSON list. |
| os_update_available | [string](#string) |  | Details about OS Updates available for this Instance. If empty, there are no updates available. |
| custom_config | [CustomConfigResource](#compute-v1-CustomConfigResource) | repeated | List of Custom Config associated with this Instance |
| tenant_id | [string](#string) |  | Tenant Identifier |
| instance_status_detail | [string](#string) |  | textual message that gives detailed status of the instance&#39;s software components. |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-OSUpdatePolicyResource"></a>

### OSUpdatePolicyResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| name | [string](#string) |  | User-provided, human-readable name of OSUpdatePolicy |
| description | [string](#string) |  |  |
| install_packages | [string](#string) |  | Freeform text, OS-dependent. A list of package names, one per line (newline separated). Should not contain version info. Applies only to Mutable OSes. |
| update_sources | [string](#string) | repeated | OS Update Sources. Should be in &#39;DEB822 Source Format&#39; for Debian style OSs. Applies only to Mutable OSes. |
| kernel_command | [string](#string) |  | Kernel Command Line Options. Applies only to Mutable OSes. |
| target_os | [os.v1.OperatingSystemResource](#os-v1-OperatingSystemResource) |  | OS resource that should be installed to this Instance. Applies only to Immutable OSes for A/B upgrades. The field is immutable. |
| update_policy | [UpdatePolicy](#compute-v1-UpdatePolicy) |  | Update Policy for the OS update. This field is used to determine the update policy for the OS update. |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-OSUpdateRunResource"></a>

### OSUpdateRunResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| name | [string](#string) |  | User-provided, human-readable name of OSUpdateRun |
| description | [string](#string) |  |  |
| applied_policy | [OSUpdatePolicyResource](#compute-v1-OSUpdatePolicyResource) |  | OSUpdate Policy of this OSUpdate Run. The field is immutable. |
| instance | [InstanceResource](#compute-v1-InstanceResource) |  | The instance resource associated with this OS Update Run. This OS Update Run is executed for this instance and is immutable. |
| status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Status Indicator for the OS update run. This field is used to determine the status type for the OS update Run. |
| status | [string](#string) |  | Short message that describes what happened during the OS Update. |
| status_details | [string](#string) |  | Details about what happened during the OS Update. |
| status_timestamp | [uint64](#uint64) |  | UTC timestamp of OS Update status reported. |
| start_time | [uint64](#uint64) |  | UTC timestamp of OS Update started. This field is set by the RM when the OS Update Run is started. |
| end_time | [uint64](#uint64) |  | UTC timestamp of OS Update ended. This field is set by the RM when the OS Update Run is finished. |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-WorkloadMember"></a>

### WorkloadMember
Intermediate resource to represent a relation between a workload and a compute resource (i.e., instance).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| kind | [WorkloadMemberKind](#compute-v1-WorkloadMemberKind) |  | Type of member |
| workload | [WorkloadResource](#compute-v1-WorkloadResource) |  |  |
| instance | [InstanceResource](#compute-v1-InstanceResource) |  |  |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="compute-v1-WorkloadResource"></a>

### WorkloadResource
Represents a generic way to group compute resources (e.g., cluster, DHCP...).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| kind | [WorkloadKind](#compute-v1-WorkloadKind) |  | Type of workload |
| name | [string](#string) |  | user-provided, human-readable name of workload |
| external_id | [string](#string) |  | Edge towards a resource that sits outside infra realm (for example, ID of the Cluster managed by cluster orchestrator). We don&#39;t enforce any pattern, but the max length of the field is 40 chars. |
| desired_state | [WorkloadState](#compute-v1-WorkloadState) |  |  |
| current_state | [WorkloadState](#compute-v1-WorkloadState) |  |  |
| status | [string](#string) |  | Human-readable status of the workload |
| members | [WorkloadMember](#compute-v1-WorkloadMember) | repeated | Should not be used to set members |
| metadata | [string](#string) |  | Record metadata with format as json string. Example: [{&#34;key&#34;:&#34;cluster-name&#34;,&#34;value&#34;:&#34;&#34;},{&#34;key&#34;:&#34;app-id&#34;,&#34;value&#34;:&#34;&#34;}] |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="compute-v1-AmtState"></a>

### AmtState


| Name | Number | Description |
| ---- | ------ | ----------- |
| AMT_STATE_UNSPECIFIED | 0 |  |
| AMT_STATE_PROVISIONED | 1 |  |
| AMT_STATE_UNPROVISIONED | 2 |  |
| AMT_STATE_DISCONNECTED | 3 |  |



<a name="compute-v1-BaremetalControllerKind"></a>

### BaremetalControllerKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| BAREMETAL_CONTROLLER_KIND_UNSPECIFIED | 0 |  |
| BAREMETAL_CONTROLLER_KIND_NONE | 1 |  |
| BAREMETAL_CONTROLLER_KIND_IPMI | 2 |  |
| BAREMETAL_CONTROLLER_KIND_VPRO | 3 |  |
| BAREMETAL_CONTROLLER_KIND_PDU | 4 |  |



<a name="compute-v1-HostComponentState"></a>

### HostComponentState


| Name | Number | Description |
| ---- | ------ | ----------- |
| HOST_COMPONENT_STATE_UNSPECIFIED | 0 |  |
| HOST_COMPONENT_STATE_ERROR | 1 |  |
| HOST_COMPONENT_STATE_DELETED | 2 |  |
| HOST_COMPONENT_STATE_EXISTS | 3 |  |



<a name="compute-v1-HostState"></a>

### HostState
--------------------------------------------------- Host Resources --------------------------------------------------

| Name | Number | Description |
| ---- | ------ | ----------- |
| HOST_STATE_UNSPECIFIED | 0 |  |
| HOST_STATE_DELETED | 2 |  |
| HOST_STATE_ONBOARDED | 3 |  |
| HOST_STATE_UNTRUSTED | 4 |  |
| HOST_STATE_REGISTERED | 5 |  |



<a name="compute-v1-InstanceKind"></a>

### InstanceKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| INSTANCE_KIND_UNSPECIFIED | 0 |  |
| INSTANCE_KIND_VM | 1 |  |
| INSTANCE_KIND_METAL | 2 |  |



<a name="compute-v1-InstanceState"></a>

### InstanceState
--------------------------------------------- Instance Resources ----------------------------------------------------

| Name | Number | Description |
| ---- | ------ | ----------- |
| INSTANCE_STATE_UNSPECIFIED | 0 | unconfigured |
| INSTANCE_STATE_RUNNING | 1 | OS is Running |
| INSTANCE_STATE_DELETED | 2 | OS should be Deleted |
| INSTANCE_STATE_UNTRUSTED | 3 | OS should not be trusted anymore |



<a name="compute-v1-NetworkInterfaceLinkState"></a>

### NetworkInterfaceLinkState


| Name | Number | Description |
| ---- | ------ | ----------- |
| NETWORK_INTERFACE_LINK_STATE_UNSPECIFIED | 0 |  |
| NETWORK_INTERFACE_LINK_STATE_UP | 1 |  |
| NETWORK_INTERFACE_LINK_STATE_DOWN | 2 |  |



<a name="compute-v1-PowerCommandPolicy"></a>

### PowerCommandPolicy


| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_COMMAND_POLICY_UNSPECIFIED | 0 |  |
| POWER_COMMAND_POLICY_IMMEDIATE | 1 |  |
| POWER_COMMAND_POLICY_ORDERED | 2 |  |



<a name="compute-v1-PowerState"></a>

### PowerState


| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_STATE_UNSPECIFIED | 0 |  |
| POWER_STATE_ON | 2 |  |
| POWER_STATE_OFF | 3 |  |
| POWER_STATE_SLEEP | 4 |  |
| POWER_STATE_HIBERNATE | 5 |  |
| POWER_STATE_RESET | 6 |  |
| POWER_STATE_POWER_CYCLE | 7 |  |



<a name="compute-v1-UpdatePolicy"></a>

### UpdatePolicy
UPDATE_POLICY_LATEST:
- for mutable: unsupported
- for immutable: latest version of the OS Resource
UPDATE_POLICY_TARGET:
- for mutable: apply the install_packages, update_sources, kernel_command
- for immutable: install the version referenced by target_os

| Name | Number | Description |
| ---- | ------ | ----------- |
| UPDATE_POLICY_UNSPECIFIED | 0 | Should never be used |
| UPDATE_POLICY_LATEST | 1 | Upgrade to latest version |
| UPDATE_POLICY_TARGET | 2 | Upgrade to the target version |



<a name="compute-v1-WorkloadKind"></a>

### WorkloadKind
Represents the type of workload (e.g., cluster, DHCP, DNS...).

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKLOAD_KIND_UNSPECIFIED | 0 | Should never be used |
| WORKLOAD_KIND_CLUSTER | 1 |  |
| WORKLOAD_KIND_DHCP | 2 | currently unused, but useful to test 2-phase delete |



<a name="compute-v1-WorkloadMemberKind"></a>

### WorkloadMemberKind
Represents the type of the workload member.

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKLOAD_MEMBER_KIND_UNSPECIFIED | 0 | Should never be used |
| WORKLOAD_MEMBER_KIND_CLUSTER_NODE | 1 | Node of a cluster workload |



<a name="compute-v1-WorkloadState"></a>

### WorkloadState
--------------------------------------------- Workload Resources ----------------------------------------------------
Represents the Workload state, used for both current and desired state.

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKLOAD_STATE_UNSPECIFIED | 0 |  |
| WORKLOAD_STATE_ERROR | 1 |  |
| WORKLOAD_STATE_DELETING | 2 |  |
| WORKLOAD_STATE_DELETED | 3 |  |
| WORKLOAD_STATE_PROVISIONED | 4 |  |


 

 

 



<a name="network_v1_network-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## network/v1/network.proto



<a name="network-v1-EndpointResource"></a>

### EndpointResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| kind | [string](#string) |  | Kind of resource. Frequently tied to Provider |
| name | [string](#string) |  | user-provided, human-readable name of endpoint |
| host | [compute.v1.HostResource](#compute-v1-HostResource) |  | Host this Endpoint belongs to |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="network-v1-IPAddressResource"></a>

### IPAddressResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated by Inventory on Create |
| address | [string](#string) |  | An IP Address using CIDR notation (e.g., 192.168.1.12/24). Empty to allow the allocation in future |
| desired_state | [IPAddressState](#network-v1-IPAddressState) |  | Set to optional to allow the discovery |
| current_state | [IPAddressState](#network-v1-IPAddressState) |  |  |
| status | [IPAddressStatus](#network-v1-IPAddressStatus) |  |  |
| status_detail | [string](#string) |  | User-friendly status to provide details about the resource state |
| config_method | [IPAddressConfigMethod](#network-v1-IPAddressConfigMethod) |  | With user-assisted config we may need to use UNSPECIFIED for discovery |
| nic | [compute.v1.HostnicResource](#compute-v1-HostnicResource) |  | Nic this IPAddress is assigned to |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="network-v1-NetlinkResource"></a>

### NetlinkResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| kind | [string](#string) |  | Kind of resource. Frequently tied to Provider |
| name | [string](#string) |  | user-provided, human-readable name of netlink resource |
| desired_state | [NetlinkState](#network-v1-NetlinkState) |  |  |
| current_state | [NetlinkState](#network-v1-NetlinkState) |  |  |
| provider_status | [string](#string) |  |  |
| src | [EndpointResource](#network-v1-EndpointResource) |  |  |
| dst | [EndpointResource](#network-v1-EndpointResource) |  |  |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="network-v1-NetworkSegment"></a>

### NetworkSegment
NetworkSegment represents a logical Layer 1 (L1) of the network and a VLAN (i.e., broadcast domain)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| name | [string](#string) |  | user-provided, human-readable name of network segment |
| vlan_id | [int32](#int32) |  |  |
| site | [location.v1.SiteResource](#location-v1-SiteResource) |  | Site this NetworkSegment is located at, it can&#39;t be null |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="network-v1-IPAddressConfigMethod"></a>

### IPAddressConfigMethod


| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_ADDRESS_CONFIG_METHOD_UNSPECIFIED | 0 |  |
| IP_ADDRESS_CONFIG_METHOD_STATIC | 1 |  |
| IP_ADDRESS_CONFIG_METHOD_DYNAMIC | 2 |  |



<a name="network-v1-IPAddressState"></a>

### IPAddressState


| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_ADDRESS_STATE_UNSPECIFIED | 0 |  |
| IP_ADDRESS_STATE_ERROR | 1 |  |
| IP_ADDRESS_STATE_ASSIGNED | 2 |  |
| IP_ADDRESS_STATE_CONFIGURED | 3 |  |
| IP_ADDRESS_STATE_RELEASED | 4 |  |
| IP_ADDRESS_STATE_DELETED | 5 |  |



<a name="network-v1-IPAddressStatus"></a>

### IPAddressStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| IP_ADDRESS_STATUS_UNSPECIFIED | 0 |  |
| IP_ADDRESS_STATUS_ASSIGNMENT_ERROR | 1 |  |
| IP_ADDRESS_STATUS_ASSIGNED | 2 |  |
| IP_ADDRESS_STATUS_CONFIGURATION_ERROR | 3 |  |
| IP_ADDRESS_STATUS_CONFIGURED | 4 |  |
| IP_ADDRESS_STATUS_RELEASED | 5 |  |
| IP_ADDRESS_STATUS_ERROR | 6 |  |



<a name="network-v1-NetlinkState"></a>

### NetlinkState


| Name | Number | Description |
| ---- | ------ | ----------- |
| NETLINK_STATE_UNSPECIFIED | 0 |  |
| NETLINK_STATE_DELETED | 1 |  |
| NETLINK_STATE_ONLINE | 2 |  |
| NETLINK_STATE_OFFLINE | 3 |  |
| NETLINK_STATE_ERROR | 4 |  |


 

 

 



<a name="remoteaccess_v1_remoteaccess-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## remoteaccess/v1/remoteaccess.proto



<a name="remoteaccess-v1-RemoteAccessConfiguration"></a>

### RemoteAccessConfiguration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource identifier |
| instance | [compute.v1.InstanceResource](#compute-v1-InstanceResource) |  | Resource ID of related Instance resource |
| expiration_timestamp | [uint64](#uint64) |  | Remote access expiration timestamp |
| local_port | [uint32](#uint32) |  | Port terminating reverse SSH tunnel (on orchestrator side) Set by resource manager. |
| user | [string](#string) |  | Name of remote user configured on SSH server running on EN Set by resource manager. |
| current_state | [RemoteAccessState](#remoteaccess-v1-RemoteAccessState) |  | Expresses current state of remote access. Managed by resource manager on behalf of provider. |
| desired_state | [RemoteAccessState](#remoteaccess-v1-RemoteAccessState) |  | Expresses desired state of remote access. Set by an administrator. |
| configuration_status | [string](#string) |  | A group of fields describing the remote access configuration. Configuration status of the resource according to the provider. configuration_status, configuration_status_indicator and configuration_status_timestamp should always be updated in one shot.

textual message that describes the update status of Instance. Set by RMs only. |
| configuration_status_indicator | [status.v1.StatusIndication](#status-v1-StatusIndication) |  | Indicates interpretation of configuration_status. Set by RMs only. |
| configuration_status_timestamp | [uint64](#uint64) |  | UTC timestamp when status was last changed. Set by RMs only. |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="remoteaccess-v1-RemoteAccessState"></a>

### RemoteAccessState
Represents the Remote Access state, used for both current and desired state.

| Name | Number | Description |
| ---- | ------ | ----------- |
| REMOTE_ACCESS_STATE_UNSPECIFIED | 0 |  |
| REMOTE_ACCESS_STATE_DELETED | 1 |  |
| REMOTE_ACCESS_STATE_ERROR | 2 |  |
| REMOTE_ACCESS_STATE_ENABLED | 3 |  |


 

 

 



<a name="schedule_v1_schedule-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## schedule/v1/schedule.proto



<a name="schedule-v1-RepeatedScheduleResource"></a>

### RepeatedScheduleResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID of this RepeatedSchedule |
| schedule_status | [ScheduleStatus](#schedule-v1-ScheduleStatus) |  |  |
| name | [string](#string) |  | user-provided, human-readable name of repeated schedule |
| target_site | [location.v1.SiteResource](#location-v1-SiteResource) |  | Resource ID of Site this applies to |
| target_host | [compute.v1.HostResource](#compute-v1-HostResource) |  | Resource ID of Host this applies to |
| target_workload | [compute.v1.WorkloadResource](#compute-v1-WorkloadResource) |  | Resource ID of Workload this applies to |
| target_region | [location.v1.RegionResource](#location-v1-RegionResource) |  | Resource ID of Region this applies to |
| duration_seconds | [uint32](#uint32) |  | duration, in seconds |
| cron_minutes | [string](#string) |  | cron style minutes (0-59), it can be empty only when used in a Filter |
| cron_hours | [string](#string) |  | cron style hours (0-23), it can be empty only when used in a Filter |
| cron_day_month | [string](#string) |  | cron style day of month (1-31), it can be empty only when used in a Filter |
| cron_month | [string](#string) |  | cron style month (1-12), it can be empty only when used in a Filter |
| cron_day_week | [string](#string) |  | cron style day of week (0-6), it can be empty only when used in a Filter |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="schedule-v1-SingleScheduleResource"></a>

### SingleScheduleResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID of this SingleSchedule |
| schedule_status | [ScheduleStatus](#schedule-v1-ScheduleStatus) |  | status of one-time-schedule |
| name | [string](#string) |  | user-provided, human-readable name of one-time-schedule |
| target_site | [location.v1.SiteResource](#location-v1-SiteResource) |  | Resource ID of Site this applies to |
| target_host | [compute.v1.HostResource](#compute-v1-HostResource) |  | Resource ID of Host this applies to |
| target_workload | [compute.v1.WorkloadResource](#compute-v1-WorkloadResource) |  | Resource ID of Workload this applies to |
| target_region | [location.v1.RegionResource](#location-v1-RegionResource) |  | Resource ID of Region this applies to |
| start_seconds | [uint64](#uint64) |  | start of one-time schedule |
| end_seconds | [uint64](#uint64) |  | end of one-time schedule |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="schedule-v1-ScheduleStatus"></a>

### ScheduleStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| SCHEDULE_STATUS_UNSPECIFIED | 0 |  |
| SCHEDULE_STATUS_MAINTENANCE | 1 | generic maintenance |
| SCHEDULE_STATUS_SHIPPING | 2 | being shipped/in transit |
| SCHEDULE_STATUS_OS_UPDATE | 3 | for performing OS updates |
| SCHEDULE_STATUS_FIRMWARE_UPDATE | 4 | for peforming firmware updates |
| SCHEDULE_STATUS_CLUSTER_UPDATE | 5 | for peforming cluster updates |


 

 

 



<a name="telemetry_v1_telemetry-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## telemetry/v1/telemetry.proto



<a name="telemetry-v1-TelemetryGroupResource"></a>

### TelemetryGroupResource
TelemetryResource defines a concrete grouping of telemetry data (metrics, logs or traces).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID of this Telemetry |
| name | [string](#string) |  | user-provided, human-readable name |
| kind | [TelemetryResourceKind](#telemetry-v1-TelemetryResourceKind) |  |  |
| collector_kind | [CollectorKind](#telemetry-v1-CollectorKind) |  |  |
| groups | [string](#string) | repeated | list of metrics/logs/traces (depends on kind) groups to be gathered. It should always include entries of the same kind. |
| profiles | [TelemetryProfile](#telemetry-v1-TelemetryProfile) | repeated | back-reference to the TelemetryProfiles associated to this TelemetryGroup. Read only. |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |






<a name="telemetry-v1-TelemetryProfile"></a>

### TelemetryProfile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource ID, generated by inventory on Create |
| region | [location.v1.RegionResource](#location-v1-RegionResource) |  |  |
| site | [location.v1.SiteResource](#location-v1-SiteResource) |  |  |
| instance | [compute.v1.InstanceResource](#compute-v1-InstanceResource) |  |  |
| kind | [TelemetryResourceKind](#telemetry-v1-TelemetryResourceKind) |  |  |
| metrics_interval | [uint32](#uint32) |  | metrics interval in seconds, must be set for kind METRICS only |
| log_level | [SeverityLevel](#telemetry-v1-SeverityLevel) |  | log level, must be set for kind LOGS only |
| group | [TelemetryGroupResource](#telemetry-v1-TelemetryGroupResource) |  |  |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="telemetry-v1-CollectorKind"></a>

### CollectorKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| COLLECTOR_KIND_UNSPECIFIED | 0 |  |
| COLLECTOR_KIND_HOST | 1 | telemetry data collected from bare-metal host |
| COLLECTOR_KIND_CLUSTER | 2 | telemetry data collected from Kubernetes cluster |



<a name="telemetry-v1-SeverityLevel"></a>

### SeverityLevel


| Name | Number | Description |
| ---- | ------ | ----------- |
| SEVERITY_LEVEL_UNSPECIFIED | 0 |  |
| SEVERITY_LEVEL_CRITICAL | 1 |  |
| SEVERITY_LEVEL_ERROR | 2 |  |
| SEVERITY_LEVEL_WARN | 3 |  |
| SEVERITY_LEVEL_INFO | 4 |  |
| SEVERITY_LEVEL_DEBUG | 5 |  |



<a name="telemetry-v1-TelemetryResourceKind"></a>

### TelemetryResourceKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| TELEMETRY_RESOURCE_KIND_UNSPECIFIED | 0 |  |
| TELEMETRY_RESOURCE_KIND_METRICS | 1 |  |
| TELEMETRY_RESOURCE_KIND_LOGS | 2 |  |


 

 

 



<a name="tenant_v1_tenant-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## tenant/v1/tenant.proto



<a name="tenant-v1-Tenant"></a>

### Tenant



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource identifier |
| current_state | [TenantState](#tenant-v1-TenantState) |  | Expresses current state of tenant. |
| desired_state | [TenantState](#tenant-v1-TenantState) |  | Expresses desired state of tenant. |
| watcher_osmanager | [bool](#bool) |  | state of tenant initialization on osmanager side

-------------------------------------------------------------------------------- |
| tenant_id | [string](#string) |  | Tenant Identifier |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 


<a name="tenant-v1-TenantState"></a>

### TenantState
An Enum with the states defined by the Multi-tenant framework

| Name | Number | Description |
| ---- | ------ | ----------- |
| TENANT_STATE_UNSPECIFIED | 0 |  |
| TENANT_STATE_CREATED | 1 |  |
| TENANT_STATE_DELETED | 2 |  |


 

 

 



<a name="inventory_v1_inventory-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## inventory/v1/inventory.proto



<a name="inventory-v1-ChangeSubscribeEventsRequest"></a>

### ChangeSubscribeEventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  | The UUID of the client to change the subscriptions for. |
| subscribed_resource_kinds | [ResourceKind](#inventory-v1-ResourceKind) | repeated | The new resource kinds that the client subscribes to. Can be empty to not receive any events. Replaces the current subscriptions. |






<a name="inventory-v1-ChangeSubscribeEventsResponse"></a>

### ChangeSubscribeEventsResponse







<a name="inventory-v1-CreateResourceRequest"></a>

### CreateResourceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| resource | [Resource](#inventory-v1-Resource) |  |  |
| tenant_id | [string](#string) |  | Definition of tenant_id can be seen as redundant since tenant_id is also defined in the nested resource. Extracting tenant information from nested structs could be expensive. Tenant related requests handling strategy has been created based on convention assuming that tenant is available on top level of requests, this approach comes with clarity of implementation. Underlying implementation enforces that tenant_id is consistent with tenant_id provided in the nested resource. |






<a name="inventory-v1-DeleteAllResourcesRequest"></a>

### DeleteAllResourcesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| resource_kind | [ResourceKind](#inventory-v1-ResourceKind) |  |  |
| enforce | [bool](#bool) |  | Enforces deletion for resources supporting 2phase deletion. Transparent for all other resources. |
| tenant_id | [string](#string) |  |  |






<a name="inventory-v1-DeleteAllResourcesResponse"></a>

### DeleteAllResourcesResponse







<a name="inventory-v1-DeleteResourceRequest"></a>

### DeleteResourceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| resource_id | [string](#string) |  |  |
| tenant_id | [string](#string) |  |  |






<a name="inventory-v1-DeleteResourceResponse"></a>

### DeleteResourceResponse







<a name="inventory-v1-FindResourcesRequest"></a>

### FindResourcesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| filter | [ResourceFilter](#inventory-v1-ResourceFilter) |  |  |






<a name="inventory-v1-FindResourcesResponse"></a>

### FindResourcesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resources | [FindResourcesResponse.ResourceTenantIDCarrier](#inventory-v1-FindResourcesResponse-ResourceTenantIDCarrier) | repeated |  |
| has_next | [bool](#bool) |  | Deprecated. Use total_elements instead. |
| total_elements | [int32](#int32) |  | Total number of items the find request would return, if not limited by pagination. Callers can use this value to determine if there are more elements to be fetched, by comparing the supplied offset and returned items to the total: bool more = offset &#43; len(resource_id) &lt; total_elements |






<a name="inventory-v1-FindResourcesResponse-ResourceTenantIDCarrier"></a>

### FindResourcesResponse.ResourceTenantIDCarrier



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tenant_id | [string](#string) |  |  |
| resource_id | [string](#string) |  |  |






<a name="inventory-v1-GetResourceRequest"></a>

### GetResourceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| resource_id | [string](#string) |  |  |
| tenant_id | [string](#string) |  |  |






<a name="inventory-v1-GetResourceResponse"></a>

### GetResourceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#inventory-v1-Resource) |  |  |
| rendered_metadata | [GetResourceResponse.ResourceMetadata](#inventory-v1-GetResourceResponse-ResourceMetadata) |  |  |






<a name="inventory-v1-GetResourceResponse-ResourceMetadata"></a>

### GetResourceResponse.ResourceMetadata
Contains the rendered metadata with format as json string. Example: [{&#34;key&#34;:&#34;cluster-name&#34;,&#34;value&#34;:&#34;&#34;},{&#34;key&#34;:&#34;app-id&#34;,&#34;value&#34;:&#34;&#34;}]


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phy_metadata | [string](#string) |  |  |
| logi_metadata | [string](#string) |  |  |






<a name="inventory-v1-GetSitesPerRegionRequest"></a>

### GetSitesPerRegionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| filter | [string](#string) | repeated | List of resource ID to filter upon |
| tenant_id | [string](#string) |  | Definition of tenant_id can be seen as redundant since tenant_id is also defined in the nested resource. Extracting tenant information from nested structs could be expensive. Tenant related requests handling strategy has been created based on convention assuming that tenant is available on top level of requests, this approach comes with clarity of implementation. |






<a name="inventory-v1-GetSitesPerRegionResponse"></a>

### GetSitesPerRegionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regions | [GetSitesPerRegionResponse.Node](#inventory-v1-GetSitesPerRegionResponse-Node) | repeated | Ordered list of nodes |






<a name="inventory-v1-GetSitesPerRegionResponse-Node"></a>

### GetSitesPerRegionResponse.Node



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| child_sites | [int32](#int32) |  |  |






<a name="inventory-v1-GetTreeHierarchyRequest"></a>

### GetTreeHierarchyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| filter | [string](#string) | repeated | List of resource ID to filter upon

resource ID, generated by inventory on Create |
| descending | [bool](#bool) |  | Order the tree by descending depth (root to leaf), otherwise ordering is by ascending depth (leaf to root). |
| tenant_id | [string](#string) |  | Definition of tenant_id can be seen as redundant since it could be provided as part of nested filter. Extracting tenant information from nested structs could be expensive. Tenant related requests handling strategy has been created based on convention assuming that tenant is available on top level of requests, this approach comes with clarity of implementation. |






<a name="inventory-v1-GetTreeHierarchyResponse"></a>

### GetTreeHierarchyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tree | [GetTreeHierarchyResponse.TreeNode](#inventory-v1-GetTreeHierarchyResponse-TreeNode) | repeated | Ordered list of tree nodes by depth |






<a name="inventory-v1-GetTreeHierarchyResponse-Node"></a>

### GetTreeHierarchyResponse.Node



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  |  |
| resource_kind | [ResourceKind](#inventory-v1-ResourceKind) |  |  |






<a name="inventory-v1-GetTreeHierarchyResponse-TreeNode"></a>

### GetTreeHierarchyResponse.TreeNode



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| current_node | [GetTreeHierarchyResponse.Node](#inventory-v1-GetTreeHierarchyResponse-Node) |  |  |
| parent_nodes | [GetTreeHierarchyResponse.Node](#inventory-v1-GetTreeHierarchyResponse-Node) | repeated |  |
| name | [string](#string) |  | Name of the resource if available, otherwise unset |
| depth | [int32](#int32) |  | The depth in the tree of the current node |






<a name="inventory-v1-HeartbeatRequest"></a>

### HeartbeatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  | The UUID of the client. |






<a name="inventory-v1-HeartbeatResponse"></a>

### HeartbeatResponse







<a name="inventory-v1-ListInheritedTelemetryProfilesRequest"></a>

### ListInheritedTelemetryProfilesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| inherit_by | [ListInheritedTelemetryProfilesRequest.InheritBy](#inventory-v1-ListInheritedTelemetryProfilesRequest-InheritBy) |  | Specifies the base resource ID to inherit from (Instance, Site, or Region ID). |
| filter | [ResourceFilter](#inventory-v1-ResourceFilter) |  | Specify a filter on the inherited telemetry profiles. Allows also to specify pagination parameters (these must always be set) Note: we support ONLY the new `AIP-160`-style filter, so filter.fieldmask and filter.resource are not supported |
| tenant_id | [string](#string) |  | Definition of tenant_id can be seen as redundant since tenant_id is also defined in the nested resource. Extracting tenant information from nested structs could be expensive. Tenant related requests handling strategy has been created based on convention assuming that tenant is available on top level of requests, this approach comes with clarity of implementation. |






<a name="inventory-v1-ListInheritedTelemetryProfilesRequest-InheritBy"></a>

### ListInheritedTelemetryProfilesRequest.InheritBy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_id | [string](#string) |  |  |
| site_id | [string](#string) |  |  |
| region_id | [string](#string) |  |  |






<a name="inventory-v1-ListInheritedTelemetryProfilesResponse"></a>

### ListInheritedTelemetryProfilesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| telemetry_profiles | [telemetry.v1.TelemetryProfile](#telemetry-v1-TelemetryProfile) | repeated | The inherited Telemetry Profiles given the &#34;inherit_by&#34; param given in the request |
| total_elements | [int32](#int32) |  | Total number of Telemetry Profiles the request would return, if not limited by pagination. Callers can use this value to determine if there are more elements to be fetched, by comparing the supplied offset and returned items to the total: bool more = offset &#43; len(resource_id) &lt; total_elements |






<a name="inventory-v1-ListResourcesRequest"></a>

### ListResourcesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| filter | [ResourceFilter](#inventory-v1-ResourceFilter) |  |  |






<a name="inventory-v1-ListResourcesResponse"></a>

### ListResourcesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resources | [GetResourceResponse](#inventory-v1-GetResourceResponse) | repeated |  |
| has_next | [bool](#bool) |  | Deprecated. Use total_elements instead. |
| total_elements | [int32](#int32) |  | Total number of items the list request would return, if not limited by pagination. Callers can use this value to determine if there are more elements to be fetched, by comparing the supplied offset and returned items to the total: bool more = offset &#43; len(resources) &lt; total_elements |






<a name="inventory-v1-Resource"></a>

### Resource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | [location.v1.RegionResource](#location-v1-RegionResource) |  |  |
| site | [location.v1.SiteResource](#location-v1-SiteResource) |  |  |
| ou | [ou.v1.OuResource](#ou-v1-OuResource) |  |  |
| provider | [provider.v1.ProviderResource](#provider-v1-ProviderResource) |  |  |
| host | [compute.v1.HostResource](#compute-v1-HostResource) |  |  |
| hoststorage | [compute.v1.HoststorageResource](#compute-v1-HoststorageResource) |  |  |
| hostnic | [compute.v1.HostnicResource](#compute-v1-HostnicResource) |  |  |
| hostusb | [compute.v1.HostusbResource](#compute-v1-HostusbResource) |  |  |
| hostgpu | [compute.v1.HostgpuResource](#compute-v1-HostgpuResource) |  |  |
| instance | [compute.v1.InstanceResource](#compute-v1-InstanceResource) |  |  |
| ipaddress | [network.v1.IPAddressResource](#network-v1-IPAddressResource) |  |  |
| network_segment | [network.v1.NetworkSegment](#network-v1-NetworkSegment) |  |  |
| netlink | [network.v1.NetlinkResource](#network-v1-NetlinkResource) |  |  |
| endpoint | [network.v1.EndpointResource](#network-v1-EndpointResource) |  |  |
| os | [os.v1.OperatingSystemResource](#os-v1-OperatingSystemResource) |  |  |
| singleschedule | [schedule.v1.SingleScheduleResource](#schedule-v1-SingleScheduleResource) |  |  |
| repeatedschedule | [schedule.v1.RepeatedScheduleResource](#schedule-v1-RepeatedScheduleResource) |  |  |
| workload | [compute.v1.WorkloadResource](#compute-v1-WorkloadResource) |  |  |
| workload_member | [compute.v1.WorkloadMember](#compute-v1-WorkloadMember) |  |  |
| telemetry_group | [telemetry.v1.TelemetryGroupResource](#telemetry-v1-TelemetryGroupResource) |  |  |
| telemetry_profile | [telemetry.v1.TelemetryProfile](#telemetry-v1-TelemetryProfile) |  |  |
| tenant | [tenant.v1.Tenant](#tenant-v1-Tenant) |  |  |
| remote_access | [remoteaccess.v1.RemoteAccessConfiguration](#remoteaccess-v1-RemoteAccessConfiguration) |  |  |
| local_account | [localaccount.v1.LocalAccountResource](#localaccount-v1-LocalAccountResource) |  |  |
| os_update_policy | [compute.v1.OSUpdatePolicyResource](#compute-v1-OSUpdatePolicyResource) |  |  |
| custom_config | [compute.v1.CustomConfigResource](#compute-v1-CustomConfigResource) |  |  |
| os_update_run | [compute.v1.OSUpdateRunResource](#compute-v1-OSUpdateRunResource) |  |  |






<a name="inventory-v1-ResourceFilter"></a>

### ResourceFilter
Filter resources with the given filter. The filter requires a filter string and a resource (kind) to be specified.
Also, limit and offset parameter are used for pagination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#inventory-v1-Resource) |  | The resource kind to filter on, must always be specified. Generally the resource&#39;s fields are unset, except for metadata filters that include inherited metadata. |
| limit | [uint32](#uint32) |  |  |
| offset | [uint32](#uint32) |  |  |
| filter | [string](#string) |  | Optional filter to return only resources of interest. See https://google.aip.dev/160 for details. Note: for backwards compatability the fields `field_mask` and `resource` are used for filtering when `filter` is unset. This means an empty (=no) filter cannot be expressed at the moment. Clients wanting to use this filter mechanism must set `filter` and `resource` to select which resource type to return. Calls with an invalid filter will fail with `INVALID_ARGUMENT`. Limitations: - Timestamps are not supported beyond treating them as simple strings. - Filtering with only a naked literal (`filter: &#34;foo&#34;`) is not supported. Always provide a field. - Field names must be given as they appear in the protobuf message, but see the notes on casing. - The &#34;:&#34; (has) operator is not supported. Use the `has(&lt;edge name&gt;)` function extension instead. - Nested fields may be accessed up to 5 levels deep. I.e. `site.region.name = &#34;foo&#34;`. - If a string literal contains double quotes, the string itself must be single quoted. I.e. `metadata = &#39;{&#34;key&#34;: &#34;value&#34;}&#39;` Extensions: - All fields of the resource kind set in `resource` are hoisted into the global name space. I.e. can be accessed directly without prefixing: `resource_id = &#34;host-1234&#34;` instead of `host.resource_id = ...`. - Field names may be specified in both camelCase and snake_case. - To check for edge presence, use the `has(&lt;edge_name&gt;)` operator. E.g.: `has(site)` to filter by resources that are linked to a site. Can be used on nested edges: `has(site.region)`. - String equality comparisons are case insensitive. `name = &#34;foo&#34;` and `name = &#34;FOO&#34;` are equivalent. - String equality comparisons are fuzzy. `name = &#34;abc&#34;` will match `abc`, `abcd` and `123abc`. - String equality comparisons may contain one or multiple wildcards `*` which match any number of characters. |
| order_by | [string](#string) |  | Optional, comma-seperated list of fields that specify the sorting order of the requested resources. By default, resources are returned in alphanumerical and ascending order based on their resource ID. Fields can be given in either their proto `foo_bar` and JSON `fooBar` casing. See https://google.aip.dev/132 for details. Additional limitations: Ordering on nested fields, such as `foo.bar` is not supported. |






<a name="inventory-v1-SubscribeEventsRequest"></a>

### SubscribeEventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the API client |
| version | [string](#string) |  | version string of the Client |
| client_kind | [ClientKind](#inventory-v1-ClientKind) |  | the kind of API client |
| subscribed_resource_kinds | [ResourceKind](#inventory-v1-ResourceKind) | repeated | The resource kinds that this client provides or subscribes to. Can be empty to not receive any events. |






<a name="inventory-v1-SubscribeEventsResponse"></a>

### SubscribeEventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  | For the first event response sent to the client, a UUID will be generated and assigned to that client. Subsequent requests must use this UUID. |
| resource_id | [string](#string) |  | Deprecated, use resource instead. The resource ID that was changed. |
| resource | [Resource](#inventory-v1-Resource) |  | The changed resource. On delete events this contains the last known state. On create and update events this contains the new state. |
| event_kind | [SubscribeEventsResponse.EventKind](#inventory-v1-SubscribeEventsResponse-EventKind) |  |  |






<a name="inventory-v1-UpdateResourceRequest"></a>

### UpdateResourceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_uuid | [string](#string) |  |  |
| resource_id | [string](#string) |  |  |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  |  |
| resource | [Resource](#inventory-v1-Resource) |  |  |
| tenant_id | [string](#string) |  | Definition of tenant_id can be seen as redundant since tenant_id is also defined in the nested resource. Extracting tenant information from nested structs could be expensive. Tenant related requests handling strategy has been created based on convention assuming that tenant is available on top level of requests, this approach comes with clarity of implementation. Underlying implementation enforces that tenant_id is consistent with tenant_id provided in the nested resource. |





 


<a name="inventory-v1-ClientKind"></a>

### ClientKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| CLIENT_KIND_UNSPECIFIED | 0 | Unspecified |
| CLIENT_KIND_API | 1 | API server |
| CLIENT_KIND_RESOURCE_MANAGER | 2 | Resource manager |
| CLIENT_KIND_TENANT_CONTROLLER | 3 | Tenant Controller |



<a name="inventory-v1-ResourceKind"></a>

### ResourceKind


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESOURCE_KIND_UNSPECIFIED | 0 |  |
| RESOURCE_KIND_REGION | 8 |  |
| RESOURCE_KIND_SITE | 9 |  |
| RESOURCE_KIND_OU | 10 |  |
| RESOURCE_KIND_PROVIDER | 32 |  |
| RESOURCE_KIND_HOST | 48 |  |
| RESOURCE_KIND_HOSTSTORAGE | 49 |  |
| RESOURCE_KIND_HOSTNIC | 50 |  |
| RESOURCE_KIND_HOSTUSB | 51 |  |
| RESOURCE_KIND_HOSTGPU | 52 |  |
| RESOURCE_KIND_INSTANCE | 64 |  |
| RESOURCE_KIND_IPADDRESS | 95 |  |
| RESOURCE_KIND_NETWORKSEGMENT | 96 |  |
| RESOURCE_KIND_NETLINK | 97 |  |
| RESOURCE_KIND_ENDPOINT | 98 |  |
| RESOURCE_KIND_OS | 99 |  |
| RESOURCE_KIND_SINGLESCHEDULE | 100 |  |
| RESOURCE_KIND_REPEATEDSCHEDULE | 101 |  |
| RESOURCE_KIND_WORKLOAD | 110 |  |
| RESOURCE_KIND_WORKLOAD_MEMBER | 111 |  |
| RESOURCE_KIND_TELEMETRY_GROUP | 120 |  |
| RESOURCE_KIND_TELEMETRY_PROFILE | 121 |  |
| RESOURCE_KIND_TENANT | 130 |  |
| RESOURCE_KIND_RMT_ACCESS_CONF | 150 |  |
| RESOURCE_KIND_LOCALACCOUNT | 170 |  |
| RESOURCE_KIND_OSUPDATEPOLICY | 180 |  |
| RESOURCE_KIND_CUSTOMCONFIG | 190 |  |
| RESOURCE_KIND_OSUPDATERUN | 200 |  |



<a name="inventory-v1-SubscribeEventsResponse-EventKind"></a>

### SubscribeEventsResponse.EventKind
EventKind is a inventory operation event kind for event subscriptions.

| Name | Number | Description |
| ---- | ------ | ----------- |
| EVENT_KIND_UNSPECIFIED | 0 |  |
| EVENT_KIND_CREATED | 1 |  |
| EVENT_KIND_UPDATED | 2 |  |
| EVENT_KIND_DELETED | 3 |  |


 

 


<a name="inventory-v1-InventoryService"></a>

### InventoryService
Inventory Service (IS) provides an API for managing resources.
Selected RPCs operates on tenant context, each of them specifies obligatory tenant_id field.
Any RPC operations relying on request messages not specifying tenant_id are intended to operate cross-tenant.

Client (API, Resource Manager, etc) registration and event streaming

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SubscribeEvents | [SubscribeEventsRequest](#inventory-v1-SubscribeEventsRequest) | [SubscribeEventsResponse](#inventory-v1-SubscribeEventsResponse) stream | Registers a new client and subscribes to the requested events. All clients must open and maintain this stream before making any other requests. Closing this stream de-registers the client. |
| ChangeSubscribeEvents | [ChangeSubscribeEventsRequest](#inventory-v1-ChangeSubscribeEventsRequest) | [ChangeSubscribeEventsResponse](#inventory-v1-ChangeSubscribeEventsResponse) | Changes the resource kinds the given client will receive events for. See SubscribeEvents. |
| CreateResource | [CreateResourceRequest](#inventory-v1-CreateResourceRequest) | [Resource](#inventory-v1-Resource) | Create a new resource, returning it (or error). Returns UNKNOWN_CLIENT error if the UUID is not known. See SubscribeEvents. |
| FindResources | [FindResourcesRequest](#inventory-v1-FindResourcesRequest) | [FindResourcesResponse](#inventory-v1-FindResourcesResponse) | Find resource IDs given criteria. |
| GetResource | [GetResourceRequest](#inventory-v1-GetResourceRequest) | [GetResourceResponse](#inventory-v1-GetResourceResponse) | Get information about a single resource given resource ID. |
| UpdateResource | [UpdateResourceRequest](#inventory-v1-UpdateResourceRequest) | [Resource](#inventory-v1-Resource) | Update a resource with a given ID, returning the updated resource. If the update results in a hard-delete, the resource is returned in its last state before deletion. Returns UNKNOWN_CLIENT error if the UUID is not known. See SubscribeEvents. |
| DeleteResource | [DeleteResourceRequest](#inventory-v1-DeleteResourceRequest) | [DeleteResourceResponse](#inventory-v1-DeleteResourceResponse) | Delete a resource with a given ID. Returns UNKNOWN_CLIENT error if the UUID is not known. See SubscribeEvents. |
| ListResources | [ListResourcesRequest](#inventory-v1-ListResourcesRequest) | [ListResourcesResponse](#inventory-v1-ListResourcesResponse) | List resources given a criteria. |
| ListInheritedTelemetryProfiles | [ListInheritedTelemetryProfilesRequest](#inventory-v1-ListInheritedTelemetryProfilesRequest) | [ListInheritedTelemetryProfilesResponse](#inventory-v1-ListInheritedTelemetryProfilesResponse) | Custom RPC for Telemetry: Lists the inherited telemetry given a site, instance or region ID. |
| GetTreeHierarchy | [GetTreeHierarchyRequest](#inventory-v1-GetTreeHierarchyRequest) | [GetTreeHierarchyResponse](#inventory-v1-GetTreeHierarchyResponse) | Returns the upstream tree hierarchy given the resource ID in the request. The response contains a list of adjacent nodes, from which the tree can be reconstructed. |
| GetSitesPerRegion | [GetSitesPerRegionRequest](#inventory-v1-GetSitesPerRegionRequest) | [GetSitesPerRegionResponse](#inventory-v1-GetSitesPerRegionResponse) | Returns a list of the number of sites per region ID given the list of region IDs in the request. The response contains a list of objects with a region ID associated to the total amount of sites under it. The sites under a region account for all the sites under its child regions recursively, respecting the max-depth of parent relationships among regions. |
| DeleteAllResources | [DeleteAllResourcesRequest](#inventory-v1-DeleteAllResourcesRequest) | [DeleteAllResourcesResponse](#inventory-v1-DeleteAllResourcesResponse) | Deletes all resources of given kind for tenant. |
| Heartbeat | [HeartbeatRequest](#inventory-v1-HeartbeatRequest) | [HeartbeatResponse](#inventory-v1-HeartbeatResponse) | Custom RPC to establish clients heartbeat and subscription verification. |

 



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

