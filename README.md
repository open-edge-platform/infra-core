# Edge Infrastructure Manager Core

## Overview

The repository includes the core micro-services of the Edge Infrastructure Manager of the Edge Manageability Framework.

## Get Started

The repository comprises the following components and services:

- [**API**](api/): provides a northbound REST API that can be accessed by users and other Edge Manageability Framework
services.
- [**Inventory**](inventory/): is the state store and the only component that persists state in Edge Infrastructure Manager.
- [**Inventory Exporter**](exporters-inventory/): exports, using a [Prometheus\* toolkit](https://prometheus.io/)-compatible
interface, some Inventory metrics that cannot be collected directly from the edge node software.
- [**Bulk Import Tools**](bulk-import-tools/): are tools that automate the registration of multiple edge nodes in
Edge Infrastructure Manager.
- [**Tenant Controller**](tenant-controller/): implements a controller for tenant creation and deletion.

Read more about Edge Orchestrator in the [User Guide][user-guide-url] to get started
using Edge Infrastructure Manager.

## Develop

To develop one of the several core components, please follow its guide in README.md located in its respective folder..

## Contribute

To learn how to contribute to the project, see the \[Contributor's
Guide\](<https://website-name.com>).

## Community and Support

To learn more about the project, its community, and governance, visit
the \[Edge Orchestrator Community\](<https://website-name.com>).

For support, start with \[Troubleshooting\](<https://website-name.com>) or
\[contact us\](<https://website-name.com>).

## License

Each component of the Edge Infrastructure core is licensed under
[Apache 2.0][apache-license].

Last Updated Date: April 7, 2025

[user-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html
[apache-license]: https://www.apache.org/licenses/LICENSE-2.0
