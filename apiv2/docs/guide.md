# Infrastructure Manager API Guidelines

## Table of Contents

- [Design](#overview)
- [Compile](#compile)
- [Implement](#implement)
- [Breaking Changes](#breaking-changes)

## Overview

For "historichal" reasons the Infrastructure Manager REST API was designed as a standalone  
component to be used to translate HTTP requests into internal gRPC calls to the Inventory component.
The reasons to decouple an API model to an Inventory model encompassed:
(i) customize the processing of data in REST API; (ii) sharding inventory
(per region/site); (iii) load balancing.

While proven useful decoupling the REST API from Inventory database model,
the following problems emerged:

- An extra layer of resource modelling, to be manually designed/written and maintained:
prone to semantycally diverge from Inventory/DB model;
- A standalone component to be managed by Infrastructure Manager team,
included unit test coverage, integration tests, SDLe analysis:
possibly enabling bugs in Infrastructure Manager as its cumbersome methods
to add/remove endpoints/paths/resource.

Examples of issues: remodeling REST API fields due to filtering feature;
bug in API reconnection (not yet solved);
different deprecation of fields in Inventory and REST/API resources.

As Infrastructure Manager grows in complexity, new resources and APIs emerge,
demanding proper modeling. Designing/writing/maintaining these in both places,
API and Inventory, becomes technically unfeasible.
This document has a core motivation that adding a new Infrastructure Manager
feature and maintaining it should be a simple task, not cumbersome due to REST/API development.

This documents instructs how to auto-generate a OpenAPI spec and its source code
from Inventory protobuf definitions, and an approach to migrate from the current
REST API component to another one that uses the auto-generated code.

In summary the proposal consists in: defining a protobuf file for REST API;
compile this file to generate the OpenAPI spec using buf plugins;
associate the protobuf implementation with the OpenAPI spec via the package grpc-gateway;
develop the stubs for the compiled protobuf in Inventory using the existing code
(e.g., methods in internal/store package). And finally the breaking changes of the
auto-generate API is presented regarding the previous one.

## Design

The goal is to write a `services.proto` file that contains the service rpc definitions
for each resource in the current REST API.
The example below showcases the service rpc definitions for the `RegionResource`.

It is important to note, in this file there are the annotations that generate the OpenAPI
specification for the REST API. Those are done using the `option` keyword in each rpc definition.
With those, it is possible to specify REST API endpoints and their parameters.

```proto
// Region.
service RegionService {
  // Create a region.
  rpc CreateRegion(CreateRegionRequest) returns (resources.location.v1.RegionResource) {
    option (google.api.http) = {
      post: "/edge-infra.orchestrator.apis/v2/regions"
      body: "region"
    };
  }
  // Get a list of regions.
  rpc ListRegions(ListRegionsRequest) returns (ListRegionsResponse) {
    option (google.api.http) = {get: "/edge-infra.orchestrator.apis/v2/regions"};
  }
  // Get a specific region.
  rpc GetRegion(GetRegionRequest) returns (resources.location.v1.RegionResource) {
    option (google.api.http) = {get: "/edge-infra.orchestrator.apis/v2/regions/{resource_id}"};
  }
  // Update a region.
  rpc UpdateRegion(UpdateRegionRequest) returns (resources.location.v1.RegionResource) {
    option (google.api.http) = {
      put: "/edge-infra.orchestrator.apis/v2/regions/{resource_id}"
      body: "region"
    };
  }
  // Delete a region.
  rpc DeleteRegion(DeleteRegionRequest) returns (DeleteRegionResponse) {
    option (google.api.http) = {delete: "/edge-infra.orchestrator.apis/v2/regions/{resource_id}"};
  }
}

// Request message for the CreateRegion method.
message CreateRegionRequest {
  // The region to create.
  resources.location.v1.RegionResource region = 1 [(google.api.field_behavior) = REQUIRED];
}

// Response message for the CreateRegion method.
message CreateRegionResponse {
  // The created region.
  resources.location.v1.RegionResource region = 1 [(google.api.field_behavior) = REQUIRED];
}
...
```

## Compile

The `services.proto` is compiled into the golang grpc code and into a OpenAPI spec,
given its google.api.http annotations. These compilations are done by buf plugins.
The definition below in the file `buf.gen.yaml` can be added to autogenerate the OpenAPI
specification from the annotations in the protobuf file.
The other existing definitions for go and grpc plugins in `buf.gen.yaml` file of inventory
already generate the gRPC client/server stubs for the `services.proto` file.

```yaml
  # openapi v3 - https://github.com/kollalabs/protoc-gen-openapi
  # If running locally install to match the path on
  #  devops-docker: oie_ci_testing
  # GOBIN=/tmp go install "github.com/kollalabs/protoc-gen-openapi@latest
  # cp /tmp/protoc-gen-openapi /usr/local/bin/protoc-gen-openapi-kollalabs
  - name: kollalabs
    path: protoc-gen-openapi-kollalabs
    out: api/openapi
    strategy: all
    opt:
      - title=Edge Infrastructure Manager
      - version=0.2.0-dev
      - validate=true
      - description=Edge Infrastructure Manager API - License Apache 2.0
```

Thus the buf compilation generates the code base to implement a gRPC
client and server of the `services.proto`.
With the code auto-generated, using the golang project `grpc-gateway`
it is possible implement a REST API server by linking the auto-generated OpenAPI
spec with a gRPC client of the generated code.

Notice, the linking of OpenAPI spec and gRPC client/server done by grpc-gateway
takes place with the specification of a URI path, which can contain the version of this REST API.

## Implement

Using the existing code (e.g., methods in internal/store package) of inventory package,
the gRPC stub server generated from `services.proto` can be implemented.
All CRUD methods already exist for all the resources.
Thus, it is possible to link the stub with the store methods.

If a new service is added, make sure to modify the `internal/proxy/server.go`
file adding it to the list `servicesClients`.
And in the `internal/server/server.go` make sure to register the new service at the
`func (is *InventorygRPCServer) Start` method.

If there are fields or new resources, make sure to add them to the `internal/server`
similar to the existing one. And add a method to convert `toInv*` and `fromInv*`
the message from/to inventory gRPC message models.

Finally add the unit tests related to the changes, as shown in `internal/server/host_test.go`.
Mocks are used, and auto-generated by the mockery package, regarding the inventory client interface.

And in the folder `test/client` there are the integration tests of the API,
make sure to validate your changes against them before pushing pull requests.

## Changes

Compared to the previous API, the following breaking changes are introduced in apiv2.

- Replies 200/OK for all operations or Error (convert grpc to http, as before):
before 201 defined for Created and 204 for Deleted, other errors converted from
gRPC to HTTP automatically by grpc-gateway.
- All return error is set by default: it contains the following fields `code, message, details`.
Those are directly translated from the gRPC error codes.
- All paths/resources moved to use resourceId by default
- Removal of OU resources and paths
- Enum UNSPECIFIED = 0 are not present in all enums.
- Removal of /compute: all /compute paths are changed to the /host (/compute/hosts -> /hosts).
- OS related changes: the path was moved from as `/OSResources -> /operating_systems`;
in the GET (aka LIST) method the reference of resources is changed as
`OperatingSystemResources -> OperatingSystems`.
- The Region resource is changed with the field: totalSites `<int32>`
- Removes unused site fields: `dnsServers, proxy, dockerRegistries, metricsEndpoint, ou`
- Moves schedule start/end seconds from uint64 to int64: this was done given that the
default JSON unmarshall in golang converts uint64 to string.
Another option is to move uint64 fields to string.
A long discussion can be found here `https://github.com/grpc-ecosystem/grpc-gateway/issues/438`.
- Moves all remaining uint64 to uint32: e.g., timestamp in status.
