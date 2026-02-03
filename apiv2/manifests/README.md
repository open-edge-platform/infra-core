# Deployment Scenario Manifests

This directory contains YAML manifests that define which API services are enabled for each EIM deployment scenario.

## Overview

Each manifest file defines an  **EIM scenario** - a specific configuration of enabled API services for different
deployment contexts. The filename (without extension) becomes the scenario name used at runtime.

## Usage

### Running with a Scenario

```bash
# Build and run with specific scenario
./out/api -eimScenario=vpro
./out/proxy -eimScenario=vpro
```

### Code Generation

Manifests are automatically processed during build:

```bash
make generate              # Generates go code from all protobuf definitions and manifests
make gen-allowed-services  # Regenerates just the scenario allowed list of API services
```

This generates `internal/scenario/allowlist_gen.go` containing the scenario-to-service mappings.

## Manifest Format

```yaml
---
name: scenario-name
description: Human-readable description of this scenario
services:
  - ServiceName1
  - ServiceName2
  - ServiceName3
```

### Fields

- **`name`** (required): Scenario identifier (must match filename)
- **`description`** (optional): Information about the scenario's purpose
- **`services`** (required): List of API service names to enable - must match
the service name in the protobuf definition file (`api/proto/services/v1/services.proto`)

## Adding a New Scenario

**Note:** All services defined in `services.proto` must have their handler registration functions
mapped in `internal/proxy/server.go` and `internal/server/server.go`.

1. Create manifest file

   ```bash
   # Filename becomes scenario name
   touch manifests/custom-scenario.yaml
   ```

2. Define services

   ```yaml
   name: custom-scenario
   description: Custom deployment configuration
   services:
     - HostService
     - RegionService
     - SiteService
   ```

3. Regenerate code

   ```bash
   make generate
   ```

4. Validation happens automatically:
   - Service names must match service names in `api/proto/services/v1/services.proto`
   - Build fails if invalid services are referenced
   - No code changes needed to add scenarios

## Validation Rules

The code generator (`tools/allowservicesgen`) validates:

- All service names must exist in the `services.proto` file.
- Services in manifests that don't exist in code return build error.
- Manifest files are valid YAML.
- Each manifest has at least one service.
- Services can exist in `services.proto` but not be used in any scenario.

## Related Files

- [Allowed Services Generator](../tools/allowedservicesgen/README.md) - Code generation tool
- [Services Proto](../api/proto/services/v1/services.proto) - API service definitions
- [Server Implementation](../internal/server/server.go) - Service registration
