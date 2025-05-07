# Edge Infrastructure Manager Inventory Exporter

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Get Started](#get-started)
- [Contribute](#contribute)
- [Integration tests](#integration-tests)

## Overview

This sub-repository contains the Observability exporter for the Inventory service. It exports, using a
[Prometheus\* toolkit](https://prometheus.io/)-compatible interface, some Inventory metrics that cannot be collected
directly from the edge node software.

## Features

- Prometheus* compatible northbound APIs
- Exportes active maintanance window indication
- Exportes host status information such as provisioning, onboarding and update statuses
- Exportes edge device specific information such as hostname, serial number, UUID
- Built with the support for Multitenancy
- Flexible deployments that span from a standalone binary to container-based orchestrations

## Get Started

Instructions on how to install and set up the Inventory Exporter on your development machine.

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

> NOTE: This guide shows how to deploy Inventory Exporter for local development or testing. For production deployments
use the [Edge Infrastructure Manager charts][inframanager-charts].

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

- `lint` to run a list of linting targets
- `test` to run the Exporter unit test
- `go-tidy` to update the Go dependencies and regenerate the `go.sum` file
- `build` to build the project and generate executable files
- `docker-build` to build the Exporter Docker container

See the [docs](docs) for advanced development topics:

- [Internals](docs/internals.md)

To learn more about internals and software architecture, see
[Edge Infrastructure Manager developer documentation][inframanager-dev-guide-url].

## Integration tests

To run the integration tests for Inventory Exporter, please refer to this [README](test/README.md).

[user-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html
[inframanager-dev-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/infra_manager/index.html
[contributors-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html
[inframanager-charts]: https://github.com/open-edge-platform/infra-charts
