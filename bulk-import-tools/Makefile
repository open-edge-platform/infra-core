# Bulk Import Tools Makefile

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help
.PHONY: test lint license help generate

# Optionally include tool version checks, not used in Docker builds
TOOL_VERSION_CHECK ?= 1

##### Variables #####

# Required Go Dependencies
GOLINT        := true
GOJUNITREPORT := true
MOCKGEN       := true
PROTOCGENDOC  := true

# Project variables
PROJECT_NAME	:= bulk-import-tools
PROJECT_NICKNAME := bit
BULK_IMPORT_BINARY_NAME	:= orch-host-bulk-import
PREFLIGHT_BINARY_NAME	:= orch-host-preflight

# Versioning variables
VERSION          := $(shell cat VERSION)
GIT_COMMIT       ?= $(shell git rev-parse --short HEAD)
GIT_TAG_PREFIX   := $(PROJECT_NAME)/v

# Test variables
# Set TEST_TARGET to '<testname1>' or '<testname1\|testname2>' to run specific tests or use regex '<testname.*>' - example:
#  TEST_TARGET='TestAuth.*' make test
TEST_TARGET ?= .
# default TEST_PKG, overwrite to select particular test pakages
TEST_PKG ?= $(shell go list ./internal/... | tr '\n' ,)
TEST_COVER := atomic ./internal/...
TEST_USE_DB := false

# Include shared makefile
include ../common.mk

#### Security Config & Targets ####
# Security config for Go Builds - see:
#   https://readthedocs.intel.com/SecureCodingStandards/latest/compiler/golang/
# -trimpath: Remove all file system paths from the resulting executable.
# -gcflags="all=-m": Print optimizations applied by the compiler for review and verification against security requirements.
# -gcflags="all=-spectre=all" Enable all available Spectre mitigations
# -ldflags="all=-s -w" remove the symbol and debug info
# -ldflags="all=-X ..." Embed binary build stamping information
# TODO: currently not in line with other projects, but should be updated to match.
ifeq ($(GOARCH),arm64)
	# Note that arm64 (Apple, similar) does not support any spectre mititations.
  GOEXTRAFLAGS := -trimpath -gcflags="all=-spectre= -N -l" -asmflags="all=-spectre=" -ldflags="all=-s -w -extldflags=-static \
	-X github.com/open-edge-platform/infra-core/bulk-import-tools/info.version=$(VERSION) \
	-X github.com/open-edge-platform/infra-core/bulk-import-tools/info.commit=$(GIT_COMMIT)"
else
  GOEXTRAFLAGS := -trimpath -gcflags="all=-spectre=all -N -l" -asmflags="all=-spectre=all" -ldflags="all=-s -w -extldflags=-static \
	-X github.com/open-edge-platform/infra-core/bulk-import-tools/info.version=$(VERSION) \
	-X github.com/open-edge-platform/infra-core/bulk-import-tools/info.commit=$(GIT_COMMIT)"
endif

#### Development targets ####

build: go-build ## Build local binaries

lint: $(OUT_DIR) license go-lint mdlint ## Run all lint tools

test: $(OUT_DIR) go-test go-integration ## Run all unit tests & integration tests

#### Sub-targets ####

go-build: $(OUT_DIR) $(OUT_DIR)/$(BULK_IMPORT_BINARY_NAME) $(OUT_DIR)/$(PREFLIGHT_BINARY_NAME) ## Build resource manager binary

$(OUT_DIR)/$(BULK_IMPORT_BINARY_NAME): $(SRC) $(DEPS) ## Build if sources or dependencies changed
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BULK_IMPORT_BINARY_NAME) cmd/$(BULK_IMPORT_BINARY_NAME)/orch_host_bulk_import.go

$(OUT_DIR)/$(PREFLIGHT_BINARY_NAME): $(SRC) $(DEPS) ## Build if sources or dependencies changed
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(PREFLIGHT_BINARY_NAME) cmd/$(PREFLIGHT_BINARY_NAME)/orch_host_preflight.go

go-integration: go-build ## Run integration tests
	mkdir -p $(OUT_DIR)/coverage/integration
	@grep -q "127.0.0.1 api.test.svc keycloak.test.svc" /etc/hosts || echo "127.0.0.1 api.test.svc keycloak.test.svc" | sudo tee -a /etc/hosts
	PREFLIGHT_BIN_PATH=$(shell pwd)/$(OUT_DIR)/$(PREFLIGHT_BINARY_NAME) \
	$(GOCMD) test -v ./cmd/$(PREFLIGHT_BINARY_NAME)/orch_host_preflight_test.go 
	BI_BIN_PATH=$(shell pwd)/$(OUT_DIR)/$(BULK_IMPORT_BINARY_NAME) \
	$(GOCMD) test -v ./cmd/$(BULK_IMPORT_BINARY_NAME)/orch_host_bulk_import_test.go

go-run: go-build ## Run the resource manager
	$(GOCMD) run $(GOEXTRAFLAGS) ./cmd/$(BULK_IMPORT_BINARY_NAME)/orch_host_bulk_import.go
	$(GOCMD) run $(GOEXTRAFLAGS) ./cmd/$(PREFLIGHT_BINARY_NAME)/orch_host_preflight.go

#### Arifact publishing targets ####

binary-release: ## Publish binaries to Production Release Service
	@echo "Publishing orch-host-bulk-import to Production Release Service."
	cd $(OUT_DIR) && \
	oras push $(OCI_REGISTRY)/$(OCI_REPOSITORY)/files/orch-host-bulk-import:$(TAG) --artifact-type application/vnd.intel.orch.file ./orch-host-bulk-import

	@echo "Publishing orch-host-preflight to Production Release Service."
	cd $(OUT_DIR) && \
	oras push $(OCI_REGISTRY)/$(OCI_REPOSITORY)/files/orch-host-preflight:$(TAG) --artifact-type application/vnd.intel.orch.file ./orch-host-preflight

artifact-publish: build binary-release ## Publish binaries

#### Unsupported targets ###

docker-build docker-push buf-update buf-lint db-start db-stop db-shell: ## Unsupported target
	echo '"make $@" is unsupported'
