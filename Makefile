# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

SUBPROJECTS := api apiv2 bulk-import-tools tenant-controller exporters-inventory inventory

.DEFAULT_GOAL := help
.PHONY: all build clean clean-all help lint test

SHELL	:= bash -eu -o pipefail

# Repo root directory, where base makefiles are located
REPO_ROOT := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

#### Python venv Target ####
VENV_DIR := venv_core

$(VENV_DIR): requirements.txt ## Create Python venv
	python3 -m venv $@ ;\
  set +u; . ./$@/bin/activate; set -u ;\
  python -m pip install --upgrade pip ;\
  python -m pip install -r requirements.txt

#### common targets ####
all: lint build test ## run lint, build, test for all subprojects

dependency-check: $(VENV_DIR)

lint: $(VENV_DIR) mdlint license ## lint common and all subprojects
	for dir in $(SUBPROJECTS); do $(MAKE) -C $$dir lint; done

MD_FILES := $(shell find . -type f \( -name '*.md' \) -print )
mdlint: ## lint all markdown README.md files
	markdownlint --version
	markdownlint *.md

license: $(VENV_DIR) ## Check licensing with the reuse tool
	set +u; . ./$</bin/activate; set -u ;\
  reuse --version ;\
  reuse --root . lint

build: ## build in all subprojects
	for dir in $(SUBPROJECTS); do $(MAKE) -C $$dir build; done

DOCKER_PROJECTS := api exporters-inventory inventory tenant-controller
docker-build: ## build all docker containers
	for dir in $(DOCKER_PROJECTS); do $(MAKE) -C $$dir $@; done

docker-push: ## push all docker containers
	for dir in $(DOCKER_PROJECTS); do $(MAKE) -C $$dir $@; done

docker-list: ## list all docker containers
	@echo "images:"
	@for dir in $(DOCKER_PROJECTS); do $(MAKE) -C $$dir $@; done

test: ## test in all subprojects
	for dir in $(SUBPROJECTS); do $(MAKE) -C $$dir test; done

clean: ## clean in all subprojects
	for dir in $(SUBPROJECTS); do $(MAKE) -C $$dir clean; done

clean-all: ## clean-all in all subprojects
	for dir in $(SUBPROJECTS); do $(MAKE) -C $$dir clean-all; done
	rm -rf $(VENV_DIR)

api-%: ## Run api subproject's tasks, e.g. api-test
	$(MAKE) -C api $*

bit-%: ## Run bulk-import-tools subproject's tasks, e.g. bit-test
	$(MAKE) -C bulk-import-tools $*

einv-%: ## Run exporters-inventory subproject's tasks, e.g. einv-test
	$(MAKE) -C exporters-inventory $*

inv-%: ## Run inventory subproject's tasks, e.g. inv-test
	$(MAKE) -C inventory $*

tc-%: ## Run tenant-controller subproject's tasks, e.g. tc-test
	$(MAKE) -C tenant-controller $*

#### Help Target ####
help: ## print help for each target
	@echo infra-core make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]%_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'
