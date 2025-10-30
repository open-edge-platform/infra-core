# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24.8-bookworm AS build

ENV GO111MODULE=on
ARG MAKE_TARGET=go-build
# Arguments used to stamp the binary
ARG REPO_URL
ARG VERSION
ARG REVISION
ARG BUILD_DATE

COPY common.mk version.mk /go/src/github.com/open-edge-platform/infra-core/
COPY Makefile go.mod go.sum /go/src/github.com/open-edge-platform/infra-core/api/
COPY cmd/ /go/src/github.com/open-edge-platform/infra-core/api/cmd/
COPY internal/ /go/src/github.com/open-edge-platform/infra-core/api/internal/
COPY pkg/ /go/src/github.com/open-edge-platform/infra-core/api/pkg/
COPY vendor/ /go/src/github.com/open-edge-platform/infra-core/api/vendor/
COPY api/ /go/src/github.com/open-edge-platform/infra-core/api/api/
COPY rego/ /go/src/github.com/open-edge-platform/infra-core/api/rego/

WORKDIR /go/src/github.com/open-edge-platform/infra-core/api
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

COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/api/out/api /usr/local/bin/api
COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/api/out/proxy /usr/local/bin/proxy
# Copy policy bundle
COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/api/rego/authz.rego /rego/authz.rego
      
# No entrypoint - container is used for both API SERVER and PROXY
# ENTRYPOINT [""]
