# Edge Infrastructure Manager API

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Get Started](#get-started)
- [Contribute](#contribute)
- [Using API](#using-api)
- [Integration Tests](#integration-tests)

## Overview

This sub-repository contains the implementation of the Northbound APIs for Edge Infrastructure Manager, the OpenAPI
definitions and the auto-generated golang code which can be used in other components as well as external projects to
do API calls using native golang code.

Additionally, the repo contains the Edge Infrastructure Manager Integration Tests which are used as sanity tests to
evaluate a new release.

## Features

- REST APIs with Role based access control (RBAC);
- Stateless service with the capability to be horizontally scaled;
- Adapts user oriented abstractions to/from Protobuf resources which are consumed by
[Inventory](../inventory/README.md) and by other Edge Infrastructure Manager components;
- Built with the support for Multitenancy;
- Flexible deployments that span from a standalone binary to container-based orchestrations.

## Get Started

Instructions on how to install and set up the API on your development machine.

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
- [gocover-cobertura](github.com/boumenot/gocover-cobertura) - check [$GOCOBERTURAVERSION_REQ](../version.mk)
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - check [$OAPI_CODEGEN_VERSION_REQ](../version.mk)
- [swagger-cli](https://github.com/APIDevTools/swagger-cli)
- [openapi-spec-validator](https://github.com/p1c2u/openapi-spec-validator)

You can install Go dependencies by running `make go-dependency`.

### Build the Binary

Build the project as follows:

```bash
# Build go binary
make build
```

The binary is installed in the [$OUT_DIR](../common.mk) folder.

### Usage

> NOTE: This guide shows how to deploy API for local development or testing. For production deployments use the
[Edge Infrastructure Manager charts][inframanager-charts].

```bash
make run

# Or
make go-run
```

See the [documentation][user-guide-url] if you want to learn more about using Edge Orchestrator.

## Contribute

To learn how to contribute to the project, see the [contributor's guide][contributors-guide-url]. The project will
accept contributions through Pull-Requests (PRs). PRs must be built successfully by the CI pipeline, pass linters
verifications and the unit tests.

There are several convenience make targets to support developer activities, you can use `help` to see a list of makefile
targets. The following is a list of makefile targets that support developer activities:

- `generate` to generate the API definitions and Golang bindings
- `lint` to run a list of linting targets
- `test` to run the API unit test
- `go-tidy` to update the Go dependencies and regenerate the `go.sum` file
- `build` to build the project and generate executable files
- `docker-build` to build the API Docker container

See the [docs](docs) for advanced development topics:

- [OAPI definitions](docs/oapi.md)

To learn more about internals and software architecture, see
[Edge Infrastructure Manager developer documentation][inframanager-dev-guide-url].

## Using API

To use API code as library in other applications check the folder [pkg](pkg), it contains the list of the
exported packages.

API auto-generated clients are available under [pkg/api/v0](pkg/api/v0/). Other bindings can be generated using
the [OAPI definition](api/openapi/edge-infrastructure-manager-openapi-all.yaml), see the documentation for your
language of preference.

## Integration tests

To run the integration tests for API, please refer to this [README](test/README.md). We also provide useful make
targets to run the integration tests (int-test*).

[user-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/index.html
[inframanager-dev-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/infra_manager/index.html
[contributors-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html
[inframanager-charts]: https://github.com/open-edge-platform/infra-charts
