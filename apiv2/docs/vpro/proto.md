# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [resources/common/v1/common.proto](#resources_common_v1_common-proto)
    - [MetadataItem](#resources-common-v1-MetadataItem)
    - [Timestamps](#resources-common-v1-Timestamps)
  
- [resources/status/v1/status.proto](#resources_status_v1_status-proto)
    - [StatusIndication](#resources-status-v1-StatusIndication)
  
- [resources/compute/v1/host_vpro.proto](#resources_compute_v1_host_vpro-proto)
    - [HostResource](#resources-compute-v1-HostResource)
  
    - [AmtSku](#resources-compute-v1-AmtSku)
    - [AmtState](#resources-compute-v1-AmtState)
    - [BaremetalControllerKind](#resources-compute-v1-BaremetalControllerKind)
    - [HostState](#resources-compute-v1-HostState)
    - [PowerCommandPolicy](#resources-compute-v1-PowerCommandPolicy)
    - [PowerState](#resources-compute-v1-PowerState)
  
- [services/onboarding/v1/compute.proto](#services_onboarding_v1_compute-proto)
    - [CreateHostRequest](#services-v1-CreateHostRequest)
    - [CreateHostResponse](#services-v1-CreateHostResponse)
    - [DeleteHostRequest](#services-v1-DeleteHostRequest)
    - [DeleteHostResponse](#services-v1-DeleteHostResponse)
    - [GetHostRequest](#services-v1-GetHostRequest)
    - [GetHostResponse](#services-v1-GetHostResponse)
    - [GetHostSummaryRequest](#services-v1-GetHostSummaryRequest)
    - [GetHostSummaryResponse](#services-v1-GetHostSummaryResponse)
    - [HostRegister](#services-v1-HostRegister)
    - [InvalidateHostRequest](#services-v1-InvalidateHostRequest)
    - [InvalidateHostResponse](#services-v1-InvalidateHostResponse)
    - [ListHostsRequest](#services-v1-ListHostsRequest)
    - [ListHostsResponse](#services-v1-ListHostsResponse)
    - [OnboardHostRequest](#services-v1-OnboardHostRequest)
    - [OnboardHostResponse](#services-v1-OnboardHostResponse)
    - [PatchHostRequest](#services-v1-PatchHostRequest)
    - [RegisterHostRequest](#services-v1-RegisterHostRequest)
    - [UpdateHostRequest](#services-v1-UpdateHostRequest)
  
    - [HostService](#services-v1-HostService)
  
- [services/v1/services.proto](#services_v1_services-proto)
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


 

 

 



<a name="resources_compute_v1_host_vpro-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## resources/compute/v1/host_vpro.proto



<a name="resources-compute-v1-HostResource"></a>

### HostResource
A Host resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | Resource ID, generated on Create. |
| name | [string](#string) |  | The host name. |
| desired_state | [HostState](#resources-compute-v1-HostState) |  | The desired state of the Host. |
| current_state | [HostState](#resources-compute-v1-HostState) |  | The current state of the Host. |
| note | [string](#string) |  | The note associated with the host. |
| serial_number | [string](#string) |  | SMBIOS device serial number. |
| uuid | [string](#string) |  | The host UUID identifier; UUID is unique and immutable. |
| bmc_kind | [BaremetalControllerKind](#resources-compute-v1-BaremetalControllerKind) |  | Kind of BMC. |
| bmc_ip | [string](#string) |  | BMC IP address, such as &#34;192.0.0.1&#34;. |
| hostname | [string](#string) |  | Hostname. |
| product_name | [string](#string) |  | System Product Name. |
| bios_version | [string](#string) |  | BIOS Version. |
| bios_release_date | [string](#string) |  | BIOS Release Date. |
| bios_vendor | [string](#string) |  | BIOS Vendor. |
| desired_power_state | [PowerState](#resources-compute-v1-PowerState) |  | Desired power state of the host |
| current_power_state | [PowerState](#resources-compute-v1-PowerState) |  | Current power state of the host |
| power_status | [string](#string) |  | textual message that describes the runtime status of Host power. Set by DM RM only. |
| power_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates dynamicity of the power_status. Set by DM RM only. |
| power_status_timestamp | [uint32](#uint32) |  | UTC timestamp when power_status was last changed. Set by DM RM only. |
| power_command_policy | [PowerCommandPolicy](#resources-compute-v1-PowerCommandPolicy) |  | Power command policy of the host. By default, it is set to PowerCommandPolicy.POWER_COMMAND_POLICY_ORDERED. |
| power_on_time | [uint32](#uint32) |  | UTC timestamp when the host was powered on. Set by DM RM only. |
| host_status | [string](#string) |  | textual message that describes the runtime status of Host. Set by RMs only. |
| host_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of host_status. Set by RMs only. |
| host_status_timestamp | [uint32](#uint32) |  | UTC timestamp when host_status was last changed. Set by RMs only. |
| onboarding_status | [string](#string) |  | textual message that describes the onboarding status of Host. Set by RMs only. |
| onboarding_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of onboarding_status. Set by RMs only. |
| onboarding_status_timestamp | [uint32](#uint32) |  | UTC timestamp when onboarding_status was last changed. Set by RMs only. |
| registration_status | [string](#string) |  | textual message that describes the onboarding status of Host. Set by RMs only. |
| registration_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates interpretation of registration_status. Set by RMs only. |
| registration_status_timestamp | [uint32](#uint32) |  | UTC timestamp when registration_status was last changed. Set by RMs only. |
| amt_sku | [AmtSku](#resources-compute-v1-AmtSku) |  | coming from device introspection |
| desired_amt_state | [AmtState](#resources-compute-v1-AmtState) |  | Desired AMT/vPRO state of the host |
| current_amt_state | [AmtState](#resources-compute-v1-AmtState) |  | Current AMT/vPRO state of the host |
| amt_status | [string](#string) |  | coming from device introspection. Set only by the DM RM. |
| amt_status_indicator | [resources.status.v1.StatusIndication](#resources-status-v1-StatusIndication) |  | Indicates dynamicity of the amt_status. Set by DM and OM RM only. |
| amt_status_timestamp | [uint32](#uint32) |  | UTC timestamp when amt_status was last changed. Set by DM and OM RM only. |
| user_lvm_size | [uint32](#uint32) |  | LVM size in GB. |
| metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The metadata associated with the host, represented by a list of key:value pairs. |
| inherited_metadata | [resources.common.v1.MetadataItem](#resources-common-v1-MetadataItem) | repeated | The metadata inherited by the host, represented by a list of key:value pairs, rendered by location and logical structures. |
| timestamps | [resources.common.v1.Timestamps](#resources-common-v1-Timestamps) |  | Timestamps associated to the resource. |





 


<a name="resources-compute-v1-AmtSku"></a>

### AmtSku


| Name | Number | Description |
| ---- | ------ | ----------- |
| AMT_SKU_UNSPECIFIED | 0 |  |
| AMT_SKU_AMT | 1 |  |
| AMT_SKU_ISM | 2 |  |



<a name="resources-compute-v1-AmtState"></a>

### AmtState
The state of the AMT (Active Management Technology) component.

| Name | Number | Description |
| ---- | ------ | ----------- |
| AMT_STATE_UNSPECIFIED | 0 |  |
| AMT_STATE_PROVISIONED | 1 |  |
| AMT_STATE_UNPROVISIONED | 2 |  |
| AMT_STATE_DISCONNECTED | 3 |  |



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



<a name="resources-compute-v1-HostState"></a>

### HostState
States of the host.

| Name | Number | Description |
| ---- | ------ | ----------- |
| HOST_STATE_UNSPECIFIED | 0 |  |
| HOST_STATE_DELETED | 2 |  |
| HOST_STATE_ONBOARDED | 3 |  |
| HOST_STATE_UNTRUSTED | 4 |  |
| HOST_STATE_REGISTERED | 5 |  |



<a name="resources-compute-v1-PowerCommandPolicy"></a>

### PowerCommandPolicy
The policy for handling power commands.

| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_COMMAND_POLICY_UNSPECIFIED | 0 |  |
| POWER_COMMAND_POLICY_IMMEDIATE | 1 |  |
| POWER_COMMAND_POLICY_ORDERED | 2 |  |



<a name="resources-compute-v1-PowerState"></a>

### PowerState
The host power state.

| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_STATE_UNSPECIFIED | 0 |  |
| POWER_STATE_ON | 2 |  |
| POWER_STATE_OFF | 3 |  |
| POWER_STATE_SLEEP | 4 |  |
| POWER_STATE_HIBERNATE | 5 |  |
| POWER_STATE_RESET | 6 |  |
| POWER_STATE_POWER_CYCLE | 7 |  |
| POWER_STATE_RESET_REPEAT | 8 | For consecutive reset operations |


 

 

 



<a name="services_onboarding_v1_compute-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## services/onboarding/v1/compute.proto



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






<a name="services-v1-DeleteHostRequest"></a>

### DeleteHostRequest
Request message for DeleteHost.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  | Name of the host host to be deleted. |






<a name="services-v1-DeleteHostResponse"></a>

### DeleteHostResponse
Reponse message for DeleteHost.






<a name="services-v1-GetHostRequest"></a>

### GetHostRequest
Request message for the GetHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  | Name of the requested host. |






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






<a name="services-v1-HostRegister"></a>

### HostRegister
Message to register a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The host name. |
| serial_number | [string](#string) |  | The host serial number. |
| uuid | [string](#string) |  | The host UUID. |
| auto_onboard | [bool](#bool) |  | Flag to signal to automatically onboard the host. |
| enable_vpro | [bool](#bool) |  | Flag to signal to enable vPRO on the host. |
| user_lvm_size | [uint32](#uint32) |  | LVM size in GB |






<a name="services-v1-InvalidateHostRequest"></a>

### InvalidateHostRequest
Request to invalidate/untrust a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  | Host resource ID |
| note | [string](#string) |  | user-provided reason for change or a freeform field |






<a name="services-v1-InvalidateHostResponse"></a>

### InvalidateHostResponse
Response message for InvalidateHost.






<a name="services-v1-ListHostsRequest"></a>

### ListHostsRequest
Request message for the ListHosts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Optional comma separated list of fields to specify a sorting order. See https://google.aip.dev/132 for details. |
| filter | [string](#string) |  | Optional filter to return only item of interest. See https://google.aip.dev/160 for details. |
| page_size | [uint32](#uint32) |  | Defines the amount of items to be contained in a single page. Default of 20. |
| offset | [uint32](#uint32) |  | Index of the first item to return. This allows skipping items. |






<a name="services-v1-ListHostsResponse"></a>

### ListHostsResponse
Response message for the ListHosts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | repeated | Sorted and filtered list of hosts. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |
| has_next | [bool](#bool) |  | Inform if there are more elements |






<a name="services-v1-OnboardHostRequest"></a>

### OnboardHostRequest
Request to onboard a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  | Host resource ID |






<a name="services-v1-OnboardHostResponse"></a>

### OnboardHostResponse
Response of a Host Register request.






<a name="services-v1-PatchHostRequest"></a>

### PatchHostRequest
Request message for the PatchHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  | ID of the resource to be updated. |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | Updated values for the host. |
| field_mask | [google.protobuf.FieldMask](#google-protobuf-FieldMask) |  | Field mask to be applied on the patch of host. |






<a name="services-v1-RegisterHostRequest"></a>

### RegisterHostRequest
Request to register a Host.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  |  |
| host | [HostRegister](#services-v1-HostRegister) |  |  |






<a name="services-v1-UpdateHostRequest"></a>

### UpdateHostRequest
Request message for the UpdateHost method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceId | [string](#string) |  | Name of the host host to be updated. |
| host | [resources.compute.v1.HostResource](#resources-compute-v1-HostResource) |  | Updated values for the host. |





 

 

 


<a name="services-v1-HostService"></a>

### HostService
Host.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetHostsSummary | [GetHostSummaryRequest](#services-v1-GetHostSummaryRequest) | [GetHostSummaryResponse](#services-v1-GetHostSummaryResponse) | Get a summary of the hosts status. |
| CreateHost | [CreateHostRequest](#services-v1-CreateHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Create a host. |
| ListHosts | [ListHostsRequest](#services-v1-ListHostsRequest) | [ListHostsResponse](#services-v1-ListHostsResponse) | Get a list of hosts. |
| GetHost | [GetHostRequest](#services-v1-GetHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Get a specific host. |
| UpdateHost | [UpdateHostRequest](#services-v1-UpdateHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Update a host. |
| PatchHost | [PatchHostRequest](#services-v1-PatchHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Patch a host. |
| DeleteHost | [DeleteHostRequest](#services-v1-DeleteHostRequest) | [DeleteHostResponse](#services-v1-DeleteHostResponse) | Delete a host. |
| InvalidateHost | [InvalidateHostRequest](#services-v1-InvalidateHostRequest) | [InvalidateHostResponse](#services-v1-InvalidateHostResponse) | Invalidate a host. |
| RegisterHost | [RegisterHostRequest](#services-v1-RegisterHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Register a host. |
| PatchRegisterHost | [RegisterHostRequest](#services-v1-RegisterHostRequest) | [.resources.compute.v1.HostResource](#resources-compute-v1-HostResource) | Update a host registration. |
| OnboardHost | [OnboardHostRequest](#services-v1-OnboardHostRequest) | [OnboardHostResponse](#services-v1-OnboardHostResponse) | Onboard a host. |

 



<a name="services_v1_services-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## services/v1/services.proto


 

 

 

 



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

