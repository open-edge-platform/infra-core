# Edge Infrastructure Manager Core

## Overview

The repository includes the core micro-services of the Edge Infrastructure Manager of the Edge Manageability Framework.
In particular, the repository comprises the following components and services:

- [**API**](api/): provides a northbound REST API that can be accessed by users and other Edge Manageability Framework
services.
- [**Inventory**](inventory/): is the state store and the only component that persists state in Edge Infrastructure Manager.
- [**Inventory Exporter**](exporters-inventory/): exports, using a [Prometheus\* toolkit](https://prometheus.io/)-compatible
interface, some Inventory metrics that cannot be collected directly from the edge node software.
- [**Bulk Import Tools**](bulk-import-tools/): are tools that automate the registration of multiple edge nodes in
Edge Infrastructure Manager.
- [**Tenant Controller**](tenant-controller/): implements a controller for tenant creation and deletion.

Read more about Edge Orchestrator in the [User Guide][user-guide-url].

Navigate through the folders to get started, develop, and contribute to Edge Infrastructure
Manager.

Last Updated Date: January 10, 2025

[user-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html
