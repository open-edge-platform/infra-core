# Edge Infrastructure Manager Core

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/open-edge-platform/infra-core/badge)](https://scorecard.dev/viewer/?uri=github.com/open-edge-platform/infra-core)

## Overview

The repository includes the core micro-services of the Edge Infrastructure Manager of the Edge Manageability Framework.

## Get Started

The repository comprises the following components and services:

- [**API**](apiv2/): provides a northbound REST API that can be accessed by users and other Edge Manageability Framework
services.
- [**Inventory**](inventory/): is the state store and the only component that persists state in Edge Infrastructure Manager.
- [**Inventory Exporter**](exporters-inventory/): exports, using a [Prometheus\* toolkit](https://prometheus.io/)-compatible
interface, some Inventory metrics that cannot be collected directly from the edge node software.
- [**Tenant Controller**](tenant-controller/): implements a controller for tenant creation and deletion.

Read more about Edge Orchestrator in the [User Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/index.html).

## Develop

To develop one of the several core components, please follow its guide in README.md located in its respective folder.

## Contribute

To learn how to contribute to the project, see the [Contributor's
Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html).

## Community and Support

To learn more about the project, its community, and governance, visit
the [Edge Orchestrator Community](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/index.html).

For support, start with [Troubleshooting](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/troubleshooting/index.html)

## License

Each component of the Edge Infrastructure external is licensed under [Apache 2.0][apache-license].

Last Updated Date: April 7, 2025

[apache-license]: https://www.apache.org/licenses/LICENSE-2.0
