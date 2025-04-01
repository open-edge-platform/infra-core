# Exporters-inventory Makefile

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help
.PHONY: test lint license help generate


# Optionally include tool version checks, not used in Docker builds
TOOL_VERSION_CHECK ?= 1

##### Variables #####

# Required Go Dependencies
GOLINT           := true
GOJUNITREPORT    := true
GOCOBERTURA      := true

# Project variables
PROJECT_NAME     := exporter
BINARY_NAME      := $(PROJECT_NAME)

# Code versions, tags, and so on
VERSION          := $(shell cat VERSION)
VERSION_MAJOR    := $(shell cut -c 1 VERSION)
DOCKER_IMG_NAME  := $(PROJECT_NAME)
DOCKER_VERSION   ?= $(shell git branch --show-current | sed 's/\//_/g')
GIT_COMMIT       ?= $(shell git rev-parse HEAD)

# Test variables
# Set TEST_TARGET to '<testname1>' or '<testname1\|testname2>' to run specific tests or use regex '<testname.*>' - example:
#  TEST_TARGET='TestAuth.*' make test
TEST_TARGET     ?= .
# Default TEST_PKG, overwrite to select particular test pakages
TEST_PKG        ?= $(shell go list ./internal/... | grep -v "cmd" | tr '\n' ,)
TEST_COVER      := atomic ./internal/...
TEST_USE_DB     := true

# Lint variables
LINT_DIRS       := cmd/... internal/...

# Include shared makefile
include ../common.mk

# Security config for Go builds
GOEXTRAFLAGS := $(COMMON_GOEXTRAFLAGS)

lint: $(OUT_DIR) license yamllint go-lint hadolint mdlint ## Run all lint

test: $(OUT_DIR) go-test ## Run all tests

build: go-build ## Build project

run: build ## Run exporter binary
	$(OUT_DIR)/$(BINARY_NAME)

go-build: $(OUT_DIR) $(OUT_DIR)/$(BINARY_NAME) ## Build resource manager binary

$(OUT_DIR)/$(BINARY_NAME): $(SRC) $(DEPS) ## Build if sources or dependencies changed
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BINARY_NAME) cmd/$(BINARY_NAME).go

go-run: $(OUT_DIR) ## Run go run
	$(GOCMD) run $(GOEXTRAFLAGS) cmd/$(BINARY_NAME).go

db-start: common-db-start ## Start the local postgres database. See: db-stop

db-stop: common-db-stop ## Stop the local postgres database. See: db-start

db-shell: common-db-shell ## Run the postgres shell connected to a local database. See: db-start

docker-build: common-docker-build ## Build Docker image

docker-push: common-docker-push ## Push Docker image

#### Unsupported targets ###

buf-update buf-lint: ## Unsupported target
	echo '"make $@" is unsupported'
