# Allowed Services Generator

This tool generates the scenario allowlist code from YAML manifest files
and validates service consistency.

## Features

1. **Code Generation**: Generates `internal/scenario/allowlist_gen.go` from YAML manifests
in `manifests/` directory
2. **Service Validation**: Validates that all services in manifests have matching service
hadlers in the generated go code.

## Usage

```bash
# Default usage (part of `make gen-allowed-services` and `make generate`)
go run ./tools/allowedservicesgen

# Custom paths
go run ./tools/allowedservicesgen \
  -manifests manifests \
  -out internal/scenario/allowlist_gen.go
```

## Validation Rules

### Manifest Service Existence

**Error if:** A service in a manifest YAML doesn't have a matching service handler.

**Example:**
```
validation failed: service in manifest does not have a matching service
handler: [InvalidServiceName]
```

**Note:** It is perfectly valid for services to have their handlers in the code but
not be used in any scenario manifest.

## Manifest Format

Manifests are YAML files in the `manifests/` directory. The filename (without extension)
becomes the scenario name.

**Example:** `manifests/vpro.yaml`
```yaml
services:
  - HostService
```

This creates a scenario named `"vpro"` with 1 service.

## Integration

This tool is automatically run as part of:
- `make generate` - Full code generation
- `make gen-allowed-services` - Run this tool only

## Exit Codes

- `0` - Success
- `1` - Validation failed or error occurred
