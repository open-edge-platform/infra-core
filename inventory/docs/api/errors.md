# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [customconfig/v1/customconfig.proto](#customconfig_v1_customconfig-proto)
    - [CustomConfigResource](#customconfig-v1-CustomConfigResource)
  
- [errors/errors.proto](#errors_errors-proto)
    - [ErrorInfo](#errors-ErrorInfo)
  
    - [Reason](#errors-Reason)
  
- [Scalar Value Types](#scalar-value-types)



<a name="customconfig_v1_customconfig-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## customconfig/v1/customconfig.proto



<a name="customconfig-v1-CustomConfigResource"></a>

### CustomConfigResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource_id | [string](#string) |  | resource identifier |
| name | [string](#string) |  | Unique name provided by admin |
| config | [string](#string) |  | Configuration file. |
| tenant_id | [string](#string) |  | Tenant Identifier. |
| created_at | [string](#string) |  | Creation timestamp |
| updated_at | [string](#string) |  | Update timestamp |





 

 

 

 



<a name="errors_errors-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## errors/errors.proto



<a name="errors-ErrorInfo"></a>

### ErrorInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reason | [Reason](#errors-Reason) |  | The reason of the error. This is a constant value that identifies the proximate cause of the error. Error reasons are unique within a particular domain of errors. |
| stacktrace | [string](#string) |  | The full error stack. Including the linenumber from which the error originated. This might require a wrap before generating this error info if the error is coming from a package outside our code. This information is only for internal debugging and not meant to be shared outside. |
| details | [google.protobuf.Any](#google-protobuf-Any) | repeated | A list of messages that carry additional error details to be standardized within this file. |





 


<a name="errors-Reason"></a>

### Reason
These are our error codes, meant to be processed by
machines or programs. Not really useful for humans.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OK | 0 | First value must be 0 and specified |
| UNKNOWN_CLIENT | 40 | UNKNOWN_CLIENT means client is unknown to the server and a new registration must be re-issued |
| OPERATION_IN_PROGRESS | 41 | OPERATION_IN_PROGRESS means that some action cannot be performed because there is other operation on a given resource in progress. |


 

 

 



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

