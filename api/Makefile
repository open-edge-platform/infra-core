# API Makefile

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
OAPI_CODEGEN     := true
GOCOBERTURA      := true

# Required system Dependencies
SWAGGERCLI       := true

# Project variables
PROJECT_NAME     := api
PROJECT_NICKNAME := api
BINARY_NAME      := $(PROJECT_NAME)

# Code versions, tags, and so on
VERSION          := $(shell cat VERSION)
# Hardcode 0 here, to avoid changing API version right before. We should get the major from the VERSION file via  $(shell cut -c 1 VERSION)
VERSION_MAJOR    := 0
DOCKER_IMG_NAME  := $(PROJECT_NAME)
DOCKER_VERSION   ?= $(shell git branch --show-current | sed 's/\//_/g')
GIT_COMMIT       ?= $(shell git rev-parse HEAD)
GIT_TAG_PREFIX   := $(PROJECT_NAME)/v

# Test variables
# Set TEST_TARGET to '<testname1>' or '<testname1\|testname2>' to run specific tests or use other regex '<testname.*>' - example:
#  TEST_TARGET='TestAuth.*' make test
TEST_TARGET      ?= .
# default TEST_PKG, overwrite to select particular test pakages
TEST_PKG         ?= $(shell go list ./... | grep -v "pkg/api" | grep -v "/test" | tr '\n' ,)
TEST_COVER       := atomic ./cmd/... ./pkg/... ./internal/...
TEST_USE_DB      := true

# yamllint variables
YAML_IGNORE := vendor, api/openapi/edge-infrastructure-manager-openapi-all.yaml, .github/workflows, ./$(VENV_NAME)

# Lint variables
LINT_DIRS := cmd/... pkg/... internal/...

DIR_TO_CLEAN := $$PWD/api/openapi/edge-infrastructure-manager-openapi-all.yaml pkg/api/*

# Include shared makefile
include ../common.mk

# Go command invocation
API_URL					   ?= "http://localhost:8080/edge-infra.orchestrator.apis/v1"
CA_PATH					   ?= ""
GOCMD_TESTCLIENTJWT_TOKEN  := PROJECT_ID=$(PROJECT_ID) JWT_TOKEN=$(JWT_TOKEN) $(GOCMD) test -timeout=20m -count=1 -failfast -v ./test/client/ -apiurl=$(API_URL) -caPath=$(CA_PATH)

API_GEN_FOLDER := pkg/api/v$(VERSION_MAJOR)

# Target variables
TYPES  := "$(API_GEN_FOLDER)/edge-infrastructure-manager-openapi-types.gen.go"
SERVER := "$(API_GEN_FOLDER)/edge-infrastructure-manager-openapi-server.gen.go"
CLIENT := "$(API_GEN_FOLDER)/edge-infrastructure-manager-openapi-client.gen.go"

# Misc variables
# REUSE-IgnoreStart
COPYRIGHT_BANNER := '1i---\n\# SPDX-FileCopyrightText: (C) 2025 Intel Corporation\n\# SPDX-License-Identifier: Apache-2.0\n\# Generated file do not edit !!!'
# REUSE-IgnoreEnd

# Security config for Go builds
GOEXTRAFLAGS := $(COMMON_GOEXTRAFLAGS)

generate: oapi-bundle $(TYPES) $(SERVER) $(CLIENT) ## generate types, server and client

lint: $(OUT_DIR) generate license yamllint go-lint hadolint mdlint ## Run all lint

test: $(OUT_DIR) validate-openapi go-test ## Run all tests

generate-api: oapi-bundle $(TYPES) $(SERVER) $(CLIENT) ## Generate Go code from Openapi Specs

build: go-build ## Build project

run: build ## Run API binary
	$(OUT_DIR)/$(BINARY_NAME)

oapi-bundle: ## Generate a single/complete OpenAPI file
	swagger-cli bundle $$PWD/api/openapi/edge-infrastructure-manager-openapi.yaml \
    		--outfile $$PWD/api/openapi/edge-infrastructure-manager-openapi-all.yaml --type yaml
	sed -i $(COPYRIGHT_BANNER) $$PWD/api/openapi/edge-infrastructure-manager-openapi-all.yaml

oapi-docs:  ## Generates a single/complete documentation of the OpenAPI spec in a html static file
	docker run --rm -v "$$PWD/api/openapi:/local" \
		redocly/cli build-docs  /local/edge-infrastructure-manager-openapi-all.yaml  --output=/local/edge-infrastructure-manager-openapi-static-doc.html

$(TYPES): api/openapi/edge-infrastructure-manager-openapi-all.yaml ## Generate types API from OpenAPI specs
	mkdir -p $(API_GEN_FOLDER)
	oapi-codegen --config=api/openapi/types.cfg.yaml $<
	mv edge-infrastructure-manager-openapi-types.gen.go $@

$(SERVER): api/openapi/edge-infrastructure-manager-openapi-all.yaml ## Generate server API from OpenAPI specs
	mkdir -p $(API_GEN_FOLDER)
	oapi-codegen --config=api/openapi/server.cfg.yaml $<
	mv edge-infrastructure-manager-openapi-server.gen.go $@

$(CLIENT): api/openapi/edge-infrastructure-manager-openapi-all.yaml ## Generate client API from OpenAPI specs
	mkdir -p $(API_GEN_FOLDER)
	oapi-codegen --config=api/openapi/client.cfg.yaml $<
	mv edge-infrastructure-manager-openapi-client.gen.go $@

go-build: $(OUT_DIR) $(OUT_DIR)/$(BINARY_NAME) ## Build resource manager binary

$(OUT_DIR)/$(BINARY_NAME): $(SRC) $(DEPS) ## Build if sources or dependencies changed
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BINARY_NAME) cmd/$(BINARY_NAME).go

# NOTE this will recompile
go-run: $(TYPES) $(SERVER) $(CLIENT) ## Run go run
	$(GOCMD) run $(GOEXTRAFLAGS) cmd/api.go

docker-build: common-docker-build ## Build Docker image

docker-push: common-docker-push  ## Push Docker image

db-start: common-db-start ## Start the local postgres database. See: db-stop

db-stop: common-db-stop ## Stop the local postgres database. See: db-start

db-shell: common-db-shell ## Run the postgres shell connected to a local database. See: db-start

validate-openapi: $(VENV_NAME) ## Check OpenAPI spec validity
	. ./$</bin/activate ; set -u ;\
  openapi-spec-validator --errors=all api/openapi/edge-infrastructure-manager-openapi-all.yaml

##### Integration tests from test/client #####

int-test: ## Run integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN)

int-test-compute: ## Run compute integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestCompute

int-test-host: ## Run host integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestHost

int-test-location: ## Run phy location integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestLocation
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestRegion
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestSite

int-test-os: ## Run os integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestOS

int-test-ou: ## Run ou integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestOU

int-test-sr: ## Run repeated schedule integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestSchedR

int-test-ss: ## Run repeated schedule integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestSchedS

int-test-workload: ## Run workload integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestCluster
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestWorkload

int-test-instance: ## Run instance integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestInstance

int-test-provider: ## Run provider integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestProvider

int-test-telemetry: ## Run telemetry integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestTelemetry

int-test-with-trace: ## Run integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	RUN_TRACE_TEST="" $(GOCMD_TESTCLIENTJWT_TOKEN)

int-test-non-auth: ## Run integration tests for ECM and Observability test clients from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestECM
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestObservabilityClient

#### Unsupported targets ###

buf-update buf-lint: ## Unsupported target
	echo '"make $@" is unsupported'
