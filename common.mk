# common.mk - common targets for Infra Core repository

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

# Makefile Style Guide:
# - Help will be generated from ## comments at end of any target line
# - Use smooth parens $() for variables over curly brackets ${} for consistency
# - Continuation lines (after an \ on previous line) should start with spaces
#   not tabs - this will cause editor highlighting to point out editing mistakes
# - When creating targets that run a lint or similar testing tool, print the
#   tool version first so that issues with versions in CI or other remote
#   environments can be caught

# Optionally include tool version checks, not used in Docker builds
ifeq ($(TOOL_VERSION_CHECK), 1)
	include ../version.mk
endif

#### Variables ####

# Shell config variable
SHELL	:= bash -eu -o pipefail

# GO variables
GOARCH	:= $(shell go env GOARCH)
GOCMD   := go

# Path variables
OUT_DIR	    := out
SRC         := $(shell find . -type f -name '*.go' ! -name '*_test.go')
DEPS        := go.mod go.sum
BASE_BRANCH := main

# Docker variables
DOCKER_ENV              := DOCKER_BUILDKIT=1
OCI_REGISTRY            ?= 080137407410.dkr.ecr.us-west-2.amazonaws.com
OCI_REPOSITORY          ?= edge-orch
DOCKER_SECTION          := infra
DOCKER_REGISTRY         ?= $(OCI_REGISTRY)
DOCKER_REPOSITORY       ?= $(OCI_REPOSITORY)
DOCKER_TAG              := $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_SECTION)/$(DOCKER_IMG_NAME):$(VERSION)
DOCKER_TAG_BRANCH       := $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_SECTION)/$(DOCKER_IMG_NAME):$(DOCKER_VERSION)
# Decides if we shall push image tagged with the branch name or not.
DOCKER_TAG_BRANCH_PUSH	?= true
LABEL_REPO_URL          ?= $(shell git remote get-url $(shell git remote | head -n 1))
LABEL_VERSION           ?= $(VERSION)
LABEL_REVISION          ?= $(GIT_COMMIT)
LABEL_BUILD_DATE        ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")

DB_CONTAINER_NAME       := $(PROJECT_NAME)-db

YAML_FILES := $(shell find . -type f \( -name '*.yaml' -o -name '*.yml' \) -print )
YAML_IGNORE ?= vendor, .github/workflows, $(VENV_NAME)
YAML_LINE_LENGTH ?= 99

# Docker networking flags for the database container.
# The problem is as follows: On a local MacOS machine we want to expose the port
# of the DB to the native host to enable smooth tooling and unit tests. During
# CI we're already inside a container, hence have to attach the DB container to
# the same network stack as the job. Because the port (-p) syntax cannot be used
# at the same time as the --network container:x flag, we need this variable.
ifeq ($(shell echo $${CI_CONTAINER:-false}), true)
  DOCKER_NETWORKING_FLAGS = --network container:$$HOSTNAME
else
  DOCKER_NETWORKING_FLAGS = -p 5432:5432
endif

# Security config for Go Builds - see:
#   https://readthedocs.intel.com/SecureCodingStandards/latest/compiler/golang/
# -trimpath: Remove all file system paths from the resulting executable.
# -gcflags="all=-m": Print optimizations applied by the compiler for review and verification against security requirements.
# -gcflags="all=-spectre=all" Enable all available Spectre mitigations
# -ldflags="all=-s -w" remove the symbol and debug info
# -ldflags="all=-X ..." Embed binary build stamping information
ifeq ($(GOARCH),arm64)
	# Note that arm64 (Apple, similar) does not support any spectre mititations.
  COMMON_GOEXTRAFLAGS := -trimpath -gcflags="all=-spectre= -N -l" -asmflags="all=-spectre=" -ldflags="all=-s -w -X 'main.RepoURL=$(LABEL_REPO_URL)' -X 'main.Version=$(LABEL_VERSION)' -X 'main.Revision=$(LABEL_REVISION)' -X 'main.BuildDate=$(LABEL_BUILD_DATE)'"
else
  COMMON_GOEXTRAFLAGS := -trimpath -gcflags="all=-spectre=all -N -l" -asmflags="all=-spectre=all" -ldflags="all=-s -w -X 'main.RepoURL=$(LABEL_REPO_URL)' -X 'main.Version=$(LABEL_VERSION)' -X 'main.Revision=$(LABEL_REVISION)' -X 'main.BuildDate=$(LABEL_BUILD_DATE)'"
endif

# Postgres DB configuration and credentials for testing. This mimics the Aurora
# production environment.
export PGUSER=admin
export PGHOST=localhost
export PGDATABASE=postgres
export PGPORT=5432
export PGPASSWORD=pass
export PGSSLMODE=disable

# Artifact publishing variables
BRANCH_NAME := $(shell git rev-parse --abbrev-ref HEAD | sed 's/\//_/g')
ifeq ($(findstring -dev,$(VERSION)), -dev)
  TAG := $(VERSION)-$(GIT_COMMIT),latest-$(BRANCH_NAME)-dev,v$(VERSION)-$(GIT_COMMIT)
else
  TAG := $(VERSION),latest-$(BRANCH_NAME),v$(VERSION)
endif

#### Path Target ####

$(OUT_DIR): ## Create out directory
	mkdir -p $(OUT_DIR)

#### Docker Targets ####

common-docker-build: ## Build Docker image
	$(GOCMD) mod vendor
	cp ../common.mk ../version.mk .
	docker build . -f Dockerfile \
		-t $(DOCKER_IMG_NAME):$(VERSION) \
		--build-arg http_proxy="$(http_proxy)" --build-arg HTTP_PROXY="$(HTTP_PROXY)" \
		--build-arg https_proxy="$(https_proxy)" --build-arg HTTPS_PROXY="$(HTTPS_PROXY)" \
		--build-arg no_proxy="$(no_proxy)" --build-arg NO_PROXY="$(NO_PROXY)" \
		--build-arg REPO_URL="$(LABEL_REPO_URL)" \
		--build-arg VERSION="$(LABEL_VERSION)" \
		--build-arg REVISION="$(LABEL_REVISION)" \
		--build-arg BUILD_DATE="$(LABEL_BUILD_DATE)"
	@rm -rf vendor common.mk version.mk

common-docker-push: ## Tag and push Docker image
	docker tag $(DOCKER_IMG_NAME):$(VERSION) $(DOCKER_TAG_BRANCH)
	docker tag $(DOCKER_IMG_NAME):$(VERSION) $(DOCKER_TAG)
	docker push $(DOCKER_TAG)
ifeq ($(DOCKER_TAG_BRANCH_PUSH), true)
	docker push $(DOCKER_TAG_BRANCH)
endif

docker-list: ## Print name of docker container image
	@echo "  $(DOCKER_IMG_NAME):"
	@echo "    name: '$(DOCKER_TAG)'"
	@echo "    version: '$(VERSION)'"
	@echo "    gitTagPrefix: '$(GIT_TAG_PREFIX)'"
	@echo "    buildTarget: '$(PROJECT_NICKNAME)-docker-build'"

#### Python venv Target ####

VENV_NAME	:= venv_$(PROJECT_NAME)

$(VENV_NAME): requirements.txt ## Create Python venv
	python3 -m venv $@ ;\
  set +u; . ./$@/bin/activate; set -u ;\
  python -m pip install --upgrade pip ;\
  python -m pip install -r requirements.txt

#### Maintenance Targets ####

go-tidy: ## Run go mod tidy
	$(GOCMD) mod tidy

go-lint-fix: ## Apply automated lint/formatting fixes to go files
	golangci-lint run --fix --config .golangci.yml

#### Test Targets ####

# https://github.com/koalaman/shellcheck
SH_FILES := $(shell find . -type f \( -name '*.sh' \) -print )
shellcheck: ## lint shell scripts with shellcheck
	shellcheck --version
	shellcheck -x -S style $(SH_FILES)

# https://pypi.org/project/reuse/
license: $(VENV_NAME) ## Check licensing with the reuse tool
	set +u; . ./$</bin/activate; set -u ;\
  reuse --version ;\
  reuse --root . lint

hadolint: ## Check Dockerfile with Hadolint
	hadolint Dockerfile

checksec: go-build ## Check various security properties that are available for executable,like RELRO, STACK CANARY, NX,PIE etc
	$(GOCMD) version -m $(OUT_DIR)/$(BINARY_NAME)
	checksec --output=json --file=$(OUT_DIR)/$(BINARY_NAME)
	checksec --fortify-file=$(OUT_DIR)/$(BINARY_NAME)

YAML_FILES := $(shell find . -type f \( -name '*.yaml' -o -name '*.yml' \) -print )
YAML_IGNORE ?= vendor, .github/workflows, $(VENV_NAME)
yamllint: $(VENV_NAME) ## Lint YAML files
	. ./$</bin/activate; set -u ;\
  yamllint --version ;\
  yamllint -d '{extends: default, rules: {line-length: {max: $(YAML_LINE_LENGTH)}}, ignore: [$(YAML_IGNORE)]}' -s $(YAML_FILES)

mdlint: ## Link MD files
	markdownlint --version ;\
	markdownlint "**/*.md" -c ../.markdownlint.yml

go-lint: $(OUT_DIR) ## Run go lint
	golangci-lint --version
	golangci-lint run $(LINT_DIRS) --config .golangci.yml

go-test: $(OUT_DIR) $(GO_TEST_DEPS) ## Run go test and calculate code coverage
ifeq ($(TEST_USE_DB), true)
	$(MAKE) db-stop
	$(MAKE) db-start
endif
	$(GOCMD) test -count=1 -race -v -p 1 \
	-coverpkg=$(TEST_PKG) -run $(TEST_TARGET) \
	-coverprofile=$(OUT_DIR)/coverage.out \
	-covermode $(TEST_COVER) $(if $(TEST_ARGS),-args $(TEST_ARGS)) \
	| tee >(go-junit-report -set-exit-code > $(OUT_DIR)/report.xml)
	gocover-cobertura $(if $(TEST_IGNORE_FILES),-ignore-files $(TEST_IGNORE_FILES)) < $(OUT_DIR)/coverage.out > $(OUT_DIR)/coverage.xml
	$(GOCMD) tool cover -html=$(OUT_DIR)/coverage.out -o $(OUT_DIR)/coverage.html
	$(GOCMD) tool cover -func=$(OUT_DIR)/coverage.out -o $(OUT_DIR)/function_coverage.log
ifeq ($(TEST_USE_DB), true)
	$(MAKE) db-stop
endif

#### Postgress DB Targets ####

common-db-start: ## Start the local postgres database. See: db-stop
	if [ -z "`docker ps -aq -f name=^$(DB_CONTAINER_NAME)`" ]; then \
		echo POSTGRES_PASSWORD=$$PGPASSWORD -e POSTGRES_DB=$$PGDATABASE -e POSTGRES_USER=$$PGUSER -d postgres:$(POSTGRES_VERSION); \
		docker run --name $(DB_CONTAINER_NAME) --rm $(DOCKER_NETWORKING_FLAGS) -e POSTGRES_PASSWORD=$$PGPASSWORD -e POSTGRES_DB=$$PGDATABASE -e POSTGRES_USER=$$PGUSER -d postgres:$(POSTGRES_VERSION); \
	fi

common-db-stop: ## Stop the local postgres database. See: db-start
	@if [ -n "`docker ps -aq -f name=^$(DB_CONTAINER_NAME)`" ]; then \
		docker container kill $(DB_CONTAINER_NAME); \
	fi

common-db-shell: ## Run the postgres shell connected to a local database. See: db-start
	docker run -it --network=host -e PGPASSWORD=$(PGPASSWORD) --name inv-shell --rm postgres:$(POSTGRES_VERSION) psql -h $$PGHOST -U $$PGUSER -d $$PGDATABASE

#### Buf protobuf code generation tooling ###

common-buf-update: $(VENV_NAME) ## Update buf modules
	set +u; . ./$</bin/activate; set -u ;\
  buf --version ;\
  pushd api; buf dep update; popd ;\
  buf build

common-buf-lint: $(VENV_NAME) ## Lint and format protobuf files
	buf --version
	buf format -d --exit-code
	buf lint

common-buf-breaking:  $(VENV_NAME) ## Find breaking changes in protobuf files
	buf breaking --against 'https://github.com/open-edge-platform/infra-core.git#branch=${BASE_BRANCH},subdir=${SUBPROJECT_DIR}'

#### Clean Targets ###

clean: ## Delete build and vendor directories
	rm -rf $(OUT_DIR) vendor $(DIR_TO_CLEAN)

clean-venv: ## Delete Python venv
	rm -rf "$(VENV_NAME)"

clean-all: clean clean-venv ## Delete all built artifacts and downloaded tools

#### Help Target ####

help: ## Print help for each target
	@echo $(PROJECT_NAME) make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'
