# Inventory Makefile

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help
.PHONY: test lint license help generate

# Optionally include tool version checks, not used in Docker builds
TOOL_VERSION_CHECK ?= 1

##### Variables #####

# Required Go Dependencies
GOLINT             := true
GOJUNITREPORT      := true
GOCOBERTURA        := true
MOCKGEN            := true
ATLAS              := true
GCC                := true
PROTOCGENDOC       := true
BUF                := true
OPA                := true
PROTOCGENENT       := true
PROTOCGENGO        := true
PROTOCGENGOGRPC    := true

# Project variables
PROJECT_NAME       := inventory
PROJECT_NICKNAME   := inv
BINARY_NAME        := $(PROJECT_NAME)

# Code versions, tags, and so on
VERSION            := $(shell cat VERSION)
VERSION_MAJOR      := $(shell cut -c 1 VERSION)
DOCKER_IMG_NAME    := $(PROJECT_NAME)
DOCKER_VERSION     ?= $(shell git branch --show-current | sed 's/\//_/g')
GIT_COMMIT         ?= $(shell git rev-parse HEAD)
GIT_TAG_PREFIX     := $(PROJECT_NAME)/v

# Test variables
# Set TEST_TARGET to '<testname1>' or '<testname1\|testname2>' to run specific tests or use other  regex '<testname.*>' - example:
#  TEST_TARGET='TestAuth.*' make test
TEST_TARGET       ?= .
# Default TEST_PKG, overwrite to select particular test pakages
TEST_PKG          ?= $(shell go list ./... | grep -v "internal/ent" | grep -v "pkg/api" | grep -v "pkg/fuzz" | grep -v "pkg/testing" | grep -v "pkg/cli" | grep -v "cmd/" | grep -v "pkg/infrainv" | grep -v "pkg/errors" | grep -v "pkg/status" | grep -v "pkg/mocks" | tr '\n' ,)
TEST_ARGS         := -policyBundle=$(BUNDLE) -globalLogLevel=error
TEST_IGNORE_FILES := pkg/errors/errors.pb.go
TEST_COVER        := atomic ./...
TEST_USE_DB       := true
GO_TEST_DEPS      := policy-build certificates

DIR_TO_CLEAN      := docs/api/* pkg/api/* pkg/errors/*.pb.go internal/ent/schema/* python/infra_inventory/* errors/ inventory/ localaccount/ provider/ schedule/ tenant/ types/ vendor/ os/ ou/ compute/ cert/certificates

# Include shared makefile
include ../common.mk

# Security config for Go builds
GOEXTRAFLAGS      := $(COMMON_GOEXTRAFLAGS)

# Migrations related variables
export MIGR_PATH  := internal/ent/migrate/migrations
MIGR_DIR          := "$(shell pwd)/$(MIGR_PATH)"
# Policy related
REGODIR           := $(shell pwd)/rego/
BUNDLE            := "$(shell pwd)/$(OUT_DIR)/policy_bundle.tar.gz"

generate: buf-gen ent-gen mock-gen cli-gen ## Generate buf and ent

lint: $(OUT_DIR) generate migration-test migration-lint license go-lint yamllint buf-lint hadolint mdlint ## Run all linters, generate targets and validate migrations

test: $(OUT_DIR) go-test ## Run all tests

build: go-build

policy-build: $(OUT_DIR) ## Rebuild the opa bundle
	opa build $(REGODIR) -o $(BUNDLE)

certificates: ## Generate certificates for unit tests
	cd cert && mkdir -p certificates && cd certificates && ../gen.sh

go-build: $(OUT_DIR) policy-build $(OUT_DIR)/$(BINARY_NAME) ## Build resource manager binary

$(OUT_DIR)/$(BINARY_NAME): $(SRC) $(DEPS) ## Build if sources or dependencies changed
	$(GOCMD) build $(GOEXTRAFLAGS) -o $(OUT_DIR)/$(BINARY_NAME) cmd/server/$(BINARY_NAME).go

go-run: build ## Run go run
	$(GOCMD) run $(GOEXTRAFLAGS) ./cmd/server/inventory.go --policyBundlePath=$(BUNDLE) -migrationsDir $(MIGR_DIR)

run: go-build ## Run go build
	$(OUT_DIR)/$(BINARY_NAME) -policyBundle=$(BUNDLE) -migrationsDir $(MIGR_DIR)

buf-update: common-buf-update ## Update buf modules

buf-gen-infrainv-schema-extender: $(VENV_NAME) ## Compile protoc api files into code
	set +u; . ./$</bin/activate; set -u ;\
  buf --version ;\
  HUMAN=true buf generate --template '{"version":"v1","plugins":[{"plugin":"buf.build/protocolbuffers/go:v1.28.1","out":"pkg", "opt":["paths=source_relative"]}]}' --path api/infrainv

buf-gen-api: $(VENV_NAME) ## Compile protoc api files into code
	set +u; . ./$</bin/activate; set -u ;\
  buf --version ;\
  HUMAN=true buf generate --exclude-path api/infrainv --exclude-path api/ent --exclude-path api/errors;\
  HUMAN=true buf generate --template cmd/protoc-gen-ent-schema-extensions/buf.gen.yaml \
  	--exclude-path api/infrainv --exclude-path api/ent --exclude-path api/errors

buf-gen-errors: $(VENV_NAME) ## Compile protoc errors files into code
	set +u; . ./$</bin/activate; set -u ;\
  buf --version ;\
  buf generate --template buf.gen.errors.yaml --exclude-path api/ent --exclude-path api/inventory --exclude-path api/compute \
	--exclude-path api/location --exclude-path api/network --exclude-path api/os --exclude-path api/ou \
	--exclude-path api/provider --exclude-path api/schedule --exclude-path api/tenant --exclude-path api/telemetry \
	--exclude-path api/status --exclude-path api/remoteaccess --exclude-path api/localaccount --exclude-path api/infrainv

buf-gen: buf-gen-infrainv-schema-extender buf-gen-api buf-gen-errors ## Compile protoc files

buf-lint: common-buf-lint ## Lint and format protobuf files

ent-gen: ## Generate ent schema
	$(GOCMD) generate ./internal/ent

mock-gen: ## Generate mocks
	mockgen --build_flags=--mod=mod -package=mocks -destination=pkg/mocks/inventory_mock.go \
		github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1 \
			InventoryServiceClient,InventoryServiceServer
	$(GOCMD) generate -run "mockgen" ./...

cli-gen: ## Generate CLI library code
	$(GOCMD) generate ./pkg/cli/gen.go

migration-generate: writeSqlSchema ## Generate DB migration "make migration-generate MIGRATION=<migration-name>"
	@if test -z $(MIGRATION); then echo "Please specify migration name" && exit 1; fi
	$(MAKE) db-stop
	$(MAKE) db-start
	sleep 5;
	atlas migrate diff $(MIGRATION) \
		--env local \
		--to "file://sql/inventory.sql"
	@$(MAKE) db-stop
	$(GOCMD) run internal/ent/migrate/gen_code_migrations.go

migration-lint: ## Validate the DB migration
	-$(MAKE) db-stop
	$(MAKE) db-start
	sleep 5;
	atlas migrate lint \
		--env local \
		--latest=1
	@$(MAKE) db-stop

migration-hash: ## Updates the atlas.sum checksums with the local migrations.
	atlas migrate hash \
		--env local

migration-test: ## Validates that no migrations are needed
	$(MAKE) migration-generate MIGRATION="testing"
	diff -u <(echo -n) <(git diff .)

docker-build: policy-build ## Build Docker image
	$(GOCMD) mod vendor
	cp ../common.mk ../version.mk .
	docker build . -f Dockerfile \
		-t $(DOCKER_IMG_NAME):$(VERSION) \
		--build-arg http_proxy="$(http_proxy)" --build-arg https_proxy="$(https_proxy)" --build-arg no_proxy="$(no_proxy)" \
		--build-arg HTTP_PROXY="$(HTTP_PROXY)" --build-arg HTTPS_PROXY="$(HTTPS_PROXY)" --build-arg NO_PROXY="$(NO_PROXY)" \
		--build-arg REPO_URL="$(LABEL_REPO_URL)" --build-arg VERSION="$(LABEL_VERSION)"  \
		--build-arg REVISION="$(LABEL_REVISION)" --build-arg BUILD_DATE="$(LABEL_BUILD_DATE)" \
		--build-arg OPA_VERSION=$(OPAVERSION_REQ) --build-arg OPA_SHA=$(OPA_SHA) --build-arg ATLAS_VERSION=$(ATLASVERSION_REQ) --build-arg ATLAS_SHA=$(ATLAS_SHA)
	@rm -rf vendor common.mk version.mk

docker-push: common-docker-push ## Push Docker image

docker-run: ## Run the inventory Docker image, e.g. for local testing. It requies a PG database running and is exposed to other components.
	docker run --rm \
		-p 50051:50051 \
		--env HUMAN=1 \
		--env PGHOST=$(shell docker inspect --format '(( .NetworkSettings.IPAddress ))' $(DB_CONTAINER_NAME)) \
		--env PGUSER --env PGDATABASE --env PGPORT --env PGPASSWORD --env PGSSLMODE \
		$(DOCKER_IMG_NAME):$(DOCKER_VERSION)

db-start: common-db-start ## Start the local postgres database. See: db-stop

db-stop: common-db-stop ## Stop the local postgres database. See: db-start

db-shell: common-db-shell ## Run the postgres shell connected to a local database. See: db-start

jaeger-start: ## Start a local Jaeger telemetry collector instance.
	if [ -z "`docker ps -aq -f name=^jaeger`" ]; then \
		docker run --rm -d --name jaeger \
			-e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
			-p 6831:6831/udp \
			-p 6832:6832/udp \
			-p 5778:5778 \
			-p 16686:16686 \
			-p 4317:4317 \
			-p 4318:4318 \
			-p 14250:14250 \
			-p 14268:14268 \
			-p 14269:14269 \
			-p 9411:9411 \
			jaegertracing/all-in-one:1.56; \
	fi

jaeger-stop: ## Stop the local Jaeger instance. See: jaeger-start
	@if [ -n "`docker ps -aq -f name=^jaeger`" ]; then \
		docker container kill jaeger; \
	fi

writeSqlSchema: ## Write sql schema
	$(MAKE) db-stop
	$(MAKE) db-start
	sleep 5
	$(GOCMD) run cmd/sqlSchemaWriter/main.go
	@$(MAKE) db-stop
	$(MAKE) gen-er-diagram

FUZZ_SECONDS	?= 60
go-fuzz: policy-build certificates ## go fuzz test
	@# Help: Runs the fuzz test stage (for 60s by default)
	$(MAKE) db-stop
	$(MAKE) db-start
	$(GOCMD) test -fuzz FuzzCreateRegion -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateSite -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateOu -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateProvider -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateHostResource -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateHoststorage -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateHostnic -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateHostusb -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateHostgpu -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateInstance -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateOs -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateSingleschedule -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateRepeatedschedule -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateWorkloadResource -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateWorkloadMember -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateTelemetryGroup -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateTelemetryProfile -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateRemoteAccess -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateIpaddress -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateNetworkSegment -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateNetlink -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateEndpoint -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(GOCMD) test -fuzz FuzzCreateRemoteAccess -fuzztime=${FUZZ_SECONDS}s ./pkg/fuzz -args -policyBundle=${BUNDLE} -globalLogLevel=error
	$(MAKE) db-stop

gen-er-diagram: ## Generate ER diagram
	npm install -g @dbml/cli
	sql2dbml sql/inventory.sql -o output.dbml
	npm install -g @softwaretechnik/dbml-renderer
	dbml-renderer --input output.dbml --output inventory-er-diagram.svg

