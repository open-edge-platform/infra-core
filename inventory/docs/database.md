<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Database use in Inventory Service

The inventory service relies on a database for storing its information. While
using an [ORM](entgo.io) abstracts away most details about the database for
developers, operators still need to be aware of some details. At this time, inventory
relies on the platform team to provide a Postgres-compatible database in
production and test deployments. This module can be found [here](missing aurora config file from pod-configs)

For local testing and development, we use a docker container that mimics Aurora
as close as possible. All make targets take care of setup and teardown on their
own, so there is no action required. Should you need to spin up this container
manually, use the provided `db-start`, `db-stop` and `db-shell` make targets. Using `gen_er_diagram` make target you can generate the ER-diagram of the database. 

## Providing Database Configuration to `inventory`

Upon startup, `inventory` will read the following environment variables to determine
its database configuration:

- `PGHOST`
- `PGPORT`
- `PGDATABASE`
- `PGUSER`
- `PGPASSWORD`
- `PGSSLMODE`

These values are (assumed to be) automatically provided in any k8s setup, but
need to be manually set before running `inventory` locally. This is due to
limitations on the the degree of which make and modify the caller's bash
environment. The correct values for testing can be found in the [Makefile](../Makefile).
