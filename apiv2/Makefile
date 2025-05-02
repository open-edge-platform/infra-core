# API Makefile

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help
.PHONY: test lint license help generate

# Optionally include tool version checks, not used in Docker builds
TOOL_VERSION_CHECK ?= 1

##### Variables #####

# Required Go Dependencies
GOLINT          := true
GOJUNITREPORT   := true
OAPI_CODEGEN    := true
GOCOBERTURA     := true
PROTOCGENDOC    := true
BUF             := true
PROTOCGENGO     := true
PROTOCGENGOGRPC := true

# Defining the shell, users and groups
CURRENT_UID := $(shell id -u)
CURRENT_GID := $(shell id -g)

# Project variables
PROJECT_NAME       := api
BINARY_NAME_API    := api
BINARY_NAME_PROXY  := proxy
SUBPROJECT_DIR     := apiv2

# Code versions, tags, and so on
VERSION       := $(shell cat VERSION)
# Hardcode 2 here, to avoid changing API version right before. We should get the major from the VERSION file via  $(shell cut -c 1 VERSION)
VERSION_MAJOR    := 2
VERSION_SUFFIX   := v2
DOCKER_IMG_NAME  := $(PROJECT_NAME)$(VERSION_SUFFIX)
DOCKER_VERSION   ?= $(shell git branch --show-current)
GIT_COMMIT       ?= $(shell git rev-parse HEAD)

# Test variables
# Set TEST_TARGET to '<testname1>' or '<testname1\|testname2>' to run specific tests or use other regex '<testname.*>' - example:
#  TEST_TARGET='TestAuth.*' make test
TEST_TARGET      ?= .
# default TEST_PKG, overwrite to select particular test pakages
TEST_PKG         ?= $(shell go list ./... | grep -v "mocks" | grep -v "pkg/api" | grep -v "internal/pbapi" | grep -v "/test" | tr '\n' ,)
TEST_COVER       := atomic ./cmd/... ./pkg/... ./internal/...
TEST_USE_DB      := true

# yamllint variables
YAML_IGNORE := vendor, api/openapi/openapi.yaml, .github/workflows, ./$(VENV_NAME)

# Lint variables
LINT_DIRS := cmd/... pkg/... internal/...

DIR_TO_CLEAN := $$PWD/internal/pbapi/* pkg/api/*

# Include shared makefile
include ../common.mk

# Go command invocation
API_URL					   ?= "http://localhost:8080"
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

generate: buf-gen generate-api ## generate protos, types, server and client

lint: $(OUT_DIR) license yamllint buf-lint go-lint hadolint mdlint ## Run all lint, add `generate` target

test: $(OUT_DIR) go-test ## Run all tests

generate-api: $(TYPES) $(SERVER) $(CLIENT) ## Generate Go code from Openapi Specs

build: go-build ## Build project

run: build ## Run API binary
	$(OUT_DIR)/$(BINARY_NAME)

oapi-banner: ## generate a single/complete OpenAPI file
	sed -i $(COPYRIGHT_BANNER) $$PWD/api/openapi/openapi.yaml

oapi-docs:  ## generates a single/complete documentation of the OpenAPI spec in a html static file
	docker run --rm -v "$$PWD/api/openapi:/local" \
		redocly/cli build-docs  /local/openapi.yaml  --output=/local/openapi-static-doc.html

$(TYPES): api/openapi/openapi.yaml ## generate types API from OpenAPI specs
	mkdir -p $(API_GEN_FOLDER)
	oapi-codegen --config=api/openapi/types.cfg.yaml $<
	mv edge-infrastructure-manager-openapi-types.gen.go $@

$(SERVER): api/openapi/openapi.yaml ## generate server API from OpenAPI specs
	mkdir -p $(API_GEN_FOLDER)
	oapi-codegen --config=api/openapi/server.cfg.yaml $<
	mv edge-infrastructure-manager-openapi-server.gen.go $@

$(CLIENT): api/openapi/openapi.yaml ## generate client API from OpenAPI specs
	mkdir -p $(API_GEN_FOLDER)
	oapi-codegen --config=api/openapi/client.cfg.yaml $<
	mv edge-infrastructure-manager-openapi-client.gen.go $@

go-build: $(OUT_DIR) ## go build
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BINARY_NAME_API) cmd/api/main.go
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BINARY_NAME_PROXY) cmd/proxy/main.go

# NOTE this will recompile
go-run: generate go-run-api go-run-proxy ## go run

go-run-api: ## go run api
	$(GOCMD) run $(GOEXTRAFLAGS) cmd/api/main.go

go-run-proxy: ## go run proxy
	$(GOCMD) run $(GOEXTRAFLAGS) cmd/proxy/main.go

docker-build: common-docker-build ## build Docker image

docker-push: docker-build common-docker-push # build and push development docker container. Only run by automated builds

db-start: common-db-start ## Start the local postgres database. See: db-stop

db-stop: common-db-stop ## Stop the local postgres database. See: db-start

db-shell: common-db-shell ## Run the postgres shell connected to a local database. See: db-start

validate-openapi: $(VENV_NAME) ## Check OpenAPI spec validity
	. ./$</bin/activate ; set -u ;\
  openapi-spec-validator --errors=all api/openapi/openapi.yaml

##### Integration tests from test/client #####

int-test: ## Runs integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN)

int-test-host: ## Runs host integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestHost

int-test-location: ## Runs phy location integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestLocation
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestRegion
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestSite

int-test-os: ## Runs os integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestOS

int-test-sr: ## Runs repeated schedule integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestSchedR

int-test-ss: ## Runs repeated schedule integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestSchedS

int-test-workload: ## Runs workload integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestCluster
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestWorkload

int-test-instance: ## Runs instance integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestInstance

int-test-provider: ## Runs provider integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestProvider

int-test-telemetry: ## Runs telemetry integration tests from test/client
	@echo "!!!!! Follow instruction in test/client/README.md !!!!!"
	@if [ -z $(JWT_TOKEN) ]; then echo "Error: JWT_TOKEN is not defined" >&2; false; fi
	@if [ -z $(PROJECT_ID) ]; then echo "Error: PROJECT_ID is not defined" >&2; false; fi
	$(GOCMD_TESTCLIENTJWT_TOKEN) -run TestTelemetry

buf-update: common-buf-update ## Update buf modules

buf-gen-api: $(VENV_NAME) ## Compile protoc api files into code
	set +u; . ./$</bin/activate; set -u ;\
  buf --version ;\
  HUMAN=true buf generate

buf-gen: buf-gen-api oapi-banner ## compile protoc files

buf-lint: ## Lint and format protobuf files
	buf --version
	# TODO: currently failing buf format and buf lint - need to be fixed
