# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24.4-bookworm as build

ENV GO111MODULE=on
ARG MAKE_TARGET=go-build
# Arguments used to stamp the binary
ARG REPO_URL
ARG VERSION
ARG REVISION
ARG BUILD_DATE

COPY common.mk version.mk /go/src/github.com/open-edge-platform/infra-core/
COPY Makefile go.mod go.sum /go/src/github.com/open-edge-platform/infra-core/exporters-inventory/
COPY cmd/ /go/src/github.com/open-edge-platform/infra-core/exporters-inventory/cmd/
COPY internal/ /go/src/github.com/open-edge-platform/infra-core/exporters-inventory/internal/
COPY vendor/ /go/src/github.com/open-edge-platform/infra-core/exporters-inventory/vendor/

WORKDIR /go/src/github.com/open-edge-platform/infra-core/exporters-inventory
RUN CGO_ENABLED=0 LABEL_REPO_URL=${REPO_URL} LABEL_VERSION=${VERSION} LABEL_REVISION=${REVISION} LABEL_BUILD_DATE=${BUILD_DATE} make ${MAKE_TARGET}

FROM gcr.io/distroless/static-debian12:nonroot
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

COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/exporters-inventory/out/exporter /usr/local/bin/exporter

ENTRYPOINT ["exporter"]
