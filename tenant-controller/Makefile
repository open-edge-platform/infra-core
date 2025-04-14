# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

# Tenant Controller Makefile

.DEFAULT_GOAL := help
.PHONY: test lint license help generate

# Optionally include tool version checks, not used in Docker builds
TOOL_VERSION_CHECK ?= 1

##### Variables #####

# Required Go Dependencies
GOLINT          := true
GOJUNITREPORT   := true
MOCKGEN         := true
OPA             := true
PROTOCGENENT    := true

# Project variables
PROJECT_NAME     := tenant-controller
PROJECT_NICKNAME := tc
BINARY_NAME      := $(PROJECT_NAME)

# Versioning variables
VERSION         := $(shell cat VERSION)
VERSION_MAJOR   := $(shell cut -c 1 VERSION)
DOCKER_IMG_NAME := $(PROJECT_NAME)
DOCKER_VERSION  ?= $(shell git branch --show-current | sed 's/\//_/g')
GIT_COMMIT      ?= $(shell git rev-parse HEAD)
GIT_TAG_PREFIX  := $(PROJECT_NAME)/v

# Test variables
# Set TEST_TARGET to '<testname1>' or '<testname1\|testname2>' to run specific tests or use regex '<testname.*>' - example:
#  TEST_TARGET='TestAuth.*' make test
TEST_TARGET     ?= .
# Default TEST_PKG, overwrite to select particular test pakages
TEST_PKG        ?= $(shell go list ./... | grep -v "pkg/api" | grep -v "cmd" | tr '\n' ,)
TEST_ARGS       := -globalLogLevel=error
TEST_COVER      := atomic ./...
TEST_USE_DB     := true

# Include shared makefile
include ../common.mk

# Security config for Go builds
GOEXTRAFLAGS    := $(COMMON_GOEXTRAFLAGS)

#### Buf protobuf code generation tooling ###

APIPKG_DIR      ?= pkg/api

buf-generate: $(VENV_NAME) ## Compile protobuf files in api into code
	set +u; . ./$</bin/activate; set -u ;\
  buf --version ;\
  buf generate

buf-lint-fix: $(VENV_NAME) ## Lint and when possible fix protobuf files
	buf --version
	buf format -d -w
	buf lint

buf-update: common-buf-update ## Update buf modules

buf-lint: common-buf-lint ## Lint and format protobuf files

#### Development Targets ####

build: go-build ## Build local binaries

lint: $(OUT_DIR) license yamllint hadolint go-lint mdlint ## Run all lint tools

test: $(OUT_DIR) go-test ## Run all unit tests

db-start: common-db-start ## Start the local postgres database. See: db-stop

db-stop: common-db-stop ## Stop the local postgres database. See: db-start

db-shell: common-db-shell ## Run the postgres shell connected to a local database. See: db-start

#### Sub-targets ####

go-build: $(OUT_DIR) $(OUT_DIR)/$(BINARY_NAME) ## Build resource manager binary

$(OUT_DIR)/$(BINARY_NAME): $(SRC) $(DEPS) ## Build if sources or dependencies changed
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BINARY_NAME) cmd/$(BINARY_NAME).go

go-run: go-build ## Run the resource manager
	$(GOCMD) run $(GOEXTRAFLAGS) ./cmd/$(BINARY_NAME).go --initResourcesDefinitionPath ./configuration/default/resources.json

docker-build: common-docker-build ## Build Docker image

docker-push: common-docker-push ## Tag and push Docker image
