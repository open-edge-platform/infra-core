#!/bin/bash
# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

# Script to check the status of container image publishing for infra-core components

REPO="open-edge-platform/infra-core"
REGISTRY="080137407410.dkr.ecr.us-west-2.amazonaws.com/edge-orch/infra"

echo "========================================="
echo "Container Image Publishing Status Check"
echo "========================================="
echo ""

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed."
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Array of components
components=("apiv2" "inventory" "exporters-inventory" "tenant-controller")

echo "Checking recent workflow runs for each component..."
echo ""

for component in "${components[@]}"; do
    echo "Component: $component"
    echo "Image: $REGISTRY/$component"
    
    # Get the most recent post-merge workflow run for this component
    workflow_name="Post-Merge ${component^}"
    
    # Use gh CLI to get recent runs
    recent_run=$(gh run list --repo "$REPO" --workflow "post-merge-$component.yml" --limit 1 --json conclusion,createdAt,headSha,htmlUrl 2>/dev/null || echo "[]")
    
    if [ "$recent_run" != "[]" ]; then
        conclusion=$(echo "$recent_run" | jq -r '.[0].conclusion')
        created_at=$(echo "$recent_run" | jq -r '.[0].createdAt')
        sha=$(echo "$recent_run" | jq -r '.[0].headSha' | cut -c1-7)
        url=$(echo "$recent_run" | jq -r '.[0].htmlUrl')
        
        if [ "$conclusion" = "success" ]; then
            echo "  ✓ Status: Published successfully"
        else
            echo "  ✗ Status: Last build $conclusion"
        fi
        echo "  Date: $created_at"
        echo "  Commit: $sha"
        echo "  Details: $url"
    else
        echo "  ⚠ No recent workflow runs found"
    fi
    echo ""
done

echo "========================================="
echo "Note: Images are automatically published to AWS ECR when code is merged to main or release-* branches."
echo "For more details, see the post-merge workflow files in .github/workflows/"
echo "========================================="
