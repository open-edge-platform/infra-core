# Container Image Publishing

This document describes how container images are built and published for the Edge Infrastructure Manager Core components.

## Overview

All components in this repository are containerized using Docker and automatically published to AWS ECR (Elastic Container Registry) as part of the CI/CD pipeline.

## Published Images

The following container images are published:

| Component | Image Location | Workflow |
|-----------|---------------|----------|
| API | `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/apiv2` | [Post-Merge API](.github/workflows/post-merge-apiv2.yml) |
| Inventory | `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/inventory` | [Post-Merge Inventory](.github/workflows/post-merge-inventory.yml) |
| Inventory Exporter | `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/exporters-inventory` | [Post-Merge Exporters Inventory](.github/workflows/post-merge-exporters-inventory.yml) |
| Tenant Controller | `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/tenant-controller` | [Post-Merge Tenant Controller](.github/workflows/post-merge-tenant-controller.yml) |

## Image Versioning

Images are tagged with multiple tags:

1. **Version Tag**: Based on the version in the [VERSION](VERSION) file (e.g., `0.1.0-dev`)
2. **Branch Tag**: The name of the branch (e.g., `main`, `release-1.0`)
3. **Commit SHA**: The Git commit SHA for traceability

Example tags for an image:
- `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/inventory:0.1.0-dev`
- `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/inventory:main`
- `080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra/inventory:ebd292a`

## Build and Publish Process

### Trigger

Images are automatically built and published when:

- Code is merged to the `main` branch
- Code is merged to `release-*` branches
- Changes affect the component's directory (e.g., changes in `inventory/` trigger the Inventory workflow)

### Workflow Steps

Each post-merge workflow performs the following steps:

1. **Version Check**: Validates the version in the VERSION file
2. **Dependency Check**: Verifies dependency versions
3. **Build**: Compiles the Go binary and builds the Docker image
4. **Security Scans**: 
   - Trivy filesystem scan
   - Bandit security scan
   - Zizmor scan
5. **Push to ECR**: Uploads the image to AWS ECR
6. **Image Signing**: Signs the image using Sigstore
7. **Image Scanning**: Scans the published image for vulnerabilities
8. **Notifications**: Sends status updates to Microsoft Teams (on failure)

### Dockerfile Details

All components use multi-stage Dockerfiles with:

- **Build Stage**: `golang:1.25.5-bookworm` - Used for compiling Go code
- **Runtime Stage**: `gcr.io/distroless/static-debian12:nonroot` - Minimal, secure base image
- **Non-root User**: Runs as non-privileged user for security
- **OCI Labels**: Includes metadata (version, source, revision, created timestamp)

Example Dockerfile structure:

```dockerfile
# Build stage
FROM golang:1.25.5-bookworm AS build
WORKDIR /workspace
COPY . .
RUN make build

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /workspace/out/binary /usr/local/bin/binary
ENTRYPOINT ["/usr/local/bin/binary"]
```

## Checking Image Publishing Status

### Using GitHub Actions UI

1. Navigate to the [Actions tab](https://github.com/open-edge-platform/infra-core/actions)
2. Filter by the workflow name (e.g., "Post-Merge Inventory")
3. Check recent runs for success status
4. Click on a run to see detailed logs including:
   - Image build output
   - Push confirmation
   - Signing details
   - Security scan results

### Using the Status Check Script

Run the included script to check the status of all component images:

```bash
./check-image-status.sh
```

This script uses the GitHub CLI to fetch recent workflow runs and display their status.

### Manual Verification

You can also check manually using the GitHub CLI:

```bash
# List recent runs for a specific workflow
gh run list --repo open-edge-platform/infra-core --workflow post-merge-inventory.yml --limit 5

# View details of a specific run
gh run view <run-id> --repo open-edge-platform/infra-core

# View logs for a specific run
gh run view <run-id> --log --repo open-edge-platform/infra-core
```

## Deployment

For information about deploying these images in production environments, see:

- [infra-charts repository](https://github.com/open-edge-platform/infra-charts) - Helm charts for Kubernetes deployment
- [Edge Orchestrator User Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/index.html)

## Security

All published images are:

1. **Scanned for vulnerabilities** using Trivy
2. **Signed** using Sigstore for supply chain security
3. **Built from minimal base images** (distroless) to reduce attack surface
4. **Run as non-root** to limit privileges

## Troubleshooting

### Images Not Being Published

If images are not being published:

1. **Check Workflow Runs**: Look for failed workflows in the Actions tab
2. **Review Logs**: Examine the workflow logs for error messages
3. **Verify Secrets**: Ensure AWS ECR credentials are properly configured in repository secrets:
   - `NO_AUTH_ECR_PUSH_USERNAME`
   - `NO_AUTH_ECR_PUSH_PASSWD`
4. **Check Workflow Configuration**: Verify that `run_docker_push: true` is set in the workflow file

### Common Issues

- **Build Failures**: Usually due to compilation errors or test failures
- **Push Failures**: Often caused by authentication issues or network problems
- **Scan Failures**: May occur if critical vulnerabilities are found

## References

- [Post-Merge Workflows](.github/workflows/)
- [Reusable Workflow from orch-ci](https://github.com/open-edge-platform/orch-ci)
- [VERSION file](VERSION)
- [Component READMEs](README.md)
