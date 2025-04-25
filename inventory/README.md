# Edge Infrastructure Manager Inventory Service

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Get Started](#get-started)
- [Contribute](#contribute)
- [Using Inventory Pkgs](#using-inventory-pkgs)

## Overview

This sub-repository contains the Inventory implementation for Edge Infrastructure Manager, gRPC API definitions and
implementation for managing the state of the infrastructure, as well as shared libraries such as Inventory client or
helper packages used across different Edge Infrastructure Manager components.

Inventory micro-service is the state store and the only component that has persistent state in
Edge Infrastructure Manager.

## Features

- Declarative APIs for Lifecycle Managment (LCM) of the persistent state
- API subscription and notification mechanism (Publish/Subscribe)
- Built with the support for Multitenancy
- Hierarchical data-model at its core and Metadata-based categorization
- Top-down policy-based management and grouping of abstractions into projects, regions and sites
- Role-based access control (RBAC)
- Attribute-based access control (ABAC)
- Flexible deployments that span from a standalone binary to container-based orchestrations
- Scalable up to 10k of edge devices

## Get Started

Instructions on how to install and set up the Inventory on your development machine.

### Dependencies

Firstly, please verify that all dependencies have been installed.

```bash
# Return errors if any dependency is missing
make dependency-check
```

This code requires the following tools to be installed on your development machine:

- [Go\* programming language](https://go.dev) - check [$GOVERSION_REQ](../version.mk)
- [golangci-lint](https://github.com/golangci/golangci-lint) - check [$GOLINTVERSION_REQ](../version.mk)
- [go-junit-report](https://github.com/jstemmer/go-junit-report) - check [$GOJUNITREPORTVERSION_REQ](../version.mk)
- [Go mockgen](https://github.com/golang/mock) - check [$MOCKGENVERSION_REQ](../version.mk)
- Python\* programming language version 3.10 or later
- [buf](https://github.com/bufbuild/buf) - check [$BUFVERSION_REQ](../version.mk)
- [protoc-gen-doc](https://github.com/pseudomuto/protoc-gen-doc) - check [$PROTOCGENDOCVERSION_REQ](../version.mk)
- [python-betterproto](https://github.com/danielgtaylor/python-betterproto)
- [protoc-gen-ent](https://github.com/ent/contrib/tree/master/entproto/cmd/protoc-gen-ent) - check
[$PROTOCGENENTVERSION_REQ](../version.mk)
- [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf) - check [$PROTOCGENGOVERSION_REQ](../version.mk)
- [protoc-gen-go-grpc](https://pkg.go.dev/google.golang.org/grpc) - check [$PROTOCGENGOGRPCVERSION_REQ](../version.mk)
- [Open Policy Agent\* (OPA) policy engine](https://www.openpolicyagent.org) - check [$OPAVERSION_REQ](../version.mk)
- [Community Edition of Atlas tool](https://atlasgo.io/community-edition) - export the `ATLAS_VERSION` variable accordingly
to [$ATLASVERSION_REQ](../version.mk)
- [gocover-cobertura](github.com/boumenot/gocover-cobertura) - check [$GOCOBERTURAVERSION_REQ](../version.mk)
- [dbml/cli](https://github.com/holistics/dbml/blob/master/packages/dbml-cli) - check [$DBMLCLI_REQ](../version.mk)
- [dbml-renderer](https://www.npmjs.com/package/@softwaretechnik/dbml-renderer) - check [$DBMLRENDERER_REQ](../version.mk)
- GNU Compiler Collection (GCC)

You can install Go dependencies by running `make go-dependency`.

### Build the Binary

Build the project as follows:

```bash
# Build go binary
make build
```

The binary is installed in the [$OUT_DIR](../common.mk) folder.

### Usage

> NOTE: This guide shows how to deploy Inventory for local development or testing. For production deployments use the
[Edge Infrastructure Manager charts][inframanager-charts].

There are two options to run the Inventory service for local development or testing. In both cases the first step is
to deploy a local database (Postgres) using `db-start`.

#### Run Inventory as standalone binary

```bash
make run

# Or
make go-run
```

#### Run Inventory as Docker container

```bash
make docker-run
```

Note that when running Inventory as a standalone project, the default configuration defined in
[common.mk](../common.mk) will be used (see `PG*` variables).

#### Generate ER diagram

```bash
make gen-er-diagram
```

Generates ER diagram of the inventory data model.

See the [documentation][user-guide-url] if you want to learn more about using Edge Orchestrator.

## Contribute

To learn how to contribute to the project, see the [contributor's guide][contributors-guide-url]. The project will
accept contributions through Pull-Requests (PRs). PRs must be built successfully by the CI pipeline, pass linters
verifications and the unit tests.

There are several convenience make targets to support developer activities, you can use `help` to see a list of makefile
targets. The following is a list of makefile targets that support developer activities:

- `generate` to generate the database schema, Go code, and the Python binding from the protobuf definition of the APIs
- `lint` to run a list of linting targets
- `test` to run the Inventory unit test
- `go-tidy` to update the Go dependencies and regenerate the `go.sum` file
- `build` to build the project and generate executable files
- `docker-build` to build the Inventory Docker container

Automatically-generated documentation of the Inventory API can be found [here](docs/api) in the following:

- [Inventory API docs](docs/api/inventory.md)
- [Errors API docs](docs/api/errors.md)

See the [docs](docs) for advanced development topics:

- [Database usage](docs/database.md)
- [Telemetry](docs/telemetry_workflow.md)
- [Versioned Schema Migrations](docs/versioned_migrations.md)

To learn more about internals and software architecture, see
[Edge Infrastructure Manager developer documentation][inframanager-dev-guide-url].

## Using Inventory pkgs

To use Inventory code as library in other applications check the folder [pkg](pkg), it contains the list of the
exported packages. See the documentation for the exported packages, as follows:

- [Auditing library](pkg/auditing/auditing.md)
- [Inventory gRPC client](pkg/client/client.md)
- [Errors library](pkg/errors/errors.md)
- [Logging library](pkg/logging/logging.md)
- [OAM gRPC server](pkg/oam/oam.md)
- [Perf tooling](pkg/perf/perf.md)
- [Testing library](pkg/testing/testing.md)
- [Tracing pkg](pkg/tracing/tracing.md)

Note that the documentation also contains runnable examples that demonstrate how to use those libraries in your code.

[user-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html
[inframanager-dev-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/infra_manager/index.html
[contributors-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html
[inframanager-charts]: https://github.com/open-edge-platform/infra-charts
