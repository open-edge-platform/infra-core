<!---
   SPDX-FileCopyrightText: (C) 2025 Intel Corporation
   SPDX-License-Identifier: Apache-2.0
-->

# Downloading Released Tools

As part of the Continuous Integration pipeline, both the tools `orch-host-preflight` and `orch-host-bulk-import` are
pushed to release registries. Both the artifacts are available in OCI (Open Container Registry) compliant registries
and it's recommended to use `oras` client to interact with them. So, ensure you have `oras` available if you want to
download these tools. Follow instructions in [public documentation](https://oras.land/docs/installation) to install
`oras` if not already done.

Download the tools as follows:

The tools are made available in the public AWS Elastic Container Registry. They can be pulled without any credentials using commands like below:

   ```bash
      oras pull registry-rs.edgeorchestration.intel.com/edge-orch/files/orch-host-preflight:3.0
      oras pull registry-rs.edgeorchestration.intel.com/edge-orch/files/orch-host-bulk-import:3.0
   ```
