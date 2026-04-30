# SPDX-FileCopyrightText: (C) 2026 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26.1-bookworm@sha256:ab3d6955bbc813a0f3fdf220c1d817dd89c0b3f283777db8ece4a32fe7858edd AS build

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

ARG MAKE_TARGET=go-build

# Arguments used to stamp the binary
ARG REPO_URL
ARG VERSION
ARG REVISION
ARG BUILD_DATE

COPY common.mk version.mk /go/src/github.com/open-edge-platform/infra-core/
COPY Makefile requirements.txt VERSION go.mod go.sum /go/src/github.com/open-edge-platform/infra-core/tenant-controller/
COPY cmd/ /go/src/github.com/open-edge-platform/infra-core/tenant-controller/cmd/
COPY internal/ /go/src/github.com/open-edge-platform/infra-core/tenant-controller/internal/
COPY vendor/ /go/src/github.com/open-edge-platform/infra-core/tenant-controller/vendor/

WORKDIR /go/src/github.com/open-edge-platform/infra-core/tenant-controller

RUN CGO_ENABLED=0 TOOL_VERSION_CHECK=0 \
  LABEL_REPO_URL=${REPO_URL} \
  LABEL_VERSION=${VERSION} \
  LABEL_REVISION=${REVISION} \
  LABEL_BUILD_DATE=${BUILD_DATE} \
  make ${MAKE_TARGET}

# From: https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

# Run as non-privileged user
USER nobody

ARG REPO_URL
ARG VERSION
ARG REVISION
ARG BUILD_DATE

LABEL org.opencontainers.image.version=${VERSION} \
      org.opencontainers.image.source=${REPO_URL} \
      org.opencontainers.image.revision=${REVISION} \
      org.opencontainers.image.created=${BUILD_DATE}

COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/tenant-controller/out/tenant-controller /usr/local/bin/tenant-controller

ENTRYPOINT ["tenant-controller"]
