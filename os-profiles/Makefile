# Exporters-inventory Makefile

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help
.PHONY: test lint license help generate

# Optionally include tool version checks, not used in Docker builds
TOOL_VERSION_CHECK ?= 0

##### Variables #####

# Project variables
PROJECT_NAME     := os-profiles

# Code versions, tags, and so on
GIT_COMMIT       ?= $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD)

VERSION          := $(shell cat VERSION)
TAGS             ?= "$(VERSION),$(BRANCH_NAME)"

ALL_PROFILES    := $(shell find *.yaml)

# Include shared makefile
include ../common.mk

# Lint variables
YAML_LINE_LENGTH := 150

OCI_REPOSITORY := edge-orch/en/files/os-profile

lint: license yamllint mdlint ## Run all lint

test: validate-profile-name ## Run all tests

validate-profile-name: ## Validate that the profile name matches the file name
	for file in $(ALL_PROFILES); do \
		profile_name=$$(yq e '.spec.profileName' $$file); \
		file_name=$$(basename $$file .yaml); \
		if [ "$$profile_name" != "$$file_name" ]; then \
			echo "Profile name $$profile_name does not match file name $$file_name in $$file"; \
			exit 1; \
		fi; \
	done

artifact-publish:
	for file in $(ALL_PROFILES); do \
  		echo $$file; \
  		profile_name=$$(basename $$file .yaml); \
		yq e -i '.metadata.release = "'$(VERSION)'"' $$file; \
		aws ecr create-repository --region us-west-2 --repository-name $(OCI_REPOSITORY)/$$profile_name || true; \
        oras push $(OCI_REGISTRY)/$(OCI_REPOSITORY)/$$profile_name:$(TAGS) --artifact-type application/vnd.intel.orch.en $$file; \
	done


#### Unsupported targets ###
# TODO: fill more targets
buf-update buf-lint build: ## Unsupported target
	echo '"make $@" is unsupported'
