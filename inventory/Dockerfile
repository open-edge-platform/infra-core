# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.25.5-bookworm as build
ARG OPA_VERSION
ARG OPA_SHA
ARG ATLAS_VERSION
ARG ATLAS_SHA

# Set proxy environment variables using build arguments
ARG http_proxy
ARG https_proxy
ARG no_proxy

SHELL ["/bin/bash", "-eo", "pipefail", "-c"]

# Install OPA to build policy bundle
RUN curl -L -o opa "https://openpolicyagent.org/downloads/v$OPA_VERSION/opa_linux_amd64_static" \
    && echo "$OPA_SHA  opa" | sha256sum -c - \
    && chmod +x opa \
    && mv opa /usr/local/bin

ENV GO111MODULE=on
ARG MAKE_TARGET=go-build
# Arguments used to stamp the binary
ARG REPO_URL
ARG VERSION
ARG REVISION
ARG BUILD_DATE

COPY common.mk version.mk /go/src/github.com/open-edge-platform/infra-core/
COPY Makefile go.mod go.sum /go/src/github.com/open-edge-platform/infra-core/inventory/
COPY internal/ /go/src/github.com/open-edge-platform/infra-core/inventory/internal/
COPY cmd/ /go/src/github.com/open-edge-platform/infra-core/inventory/cmd/
COPY pkg/ /go/src/github.com/open-edge-platform/infra-core/inventory/pkg/
COPY vendor/ /go/src/github.com/open-edge-platform/infra-core/inventory/vendor/
COPY api/ /go/src/github.com/open-edge-platform/infra-core/inventory/api/
COPY rego/ /go/src/github.com/open-edge-platform/infra-core/inventory/rego/

WORKDIR /go/src/github.com/open-edge-platform/infra-core/inventory

RUN CGO_ENABLED=0 LABEL_REPO_URL=${REPO_URL} DLABEL_VERSION=${VERSION} LABEL_REVISION=${REVISION} LABEL_BUILD_DATE=${BUILD_DATE} make ${MAKE_TARGET}

# download atlas tool
RUN curl -Lo /tmp/atlas "https://release.ariga.io/atlas/atlas-community-linux-amd64-v$ATLAS_VERSION" \
    && echo "$ATLAS_SHA  /tmp/atlas" | sha256sum -c - \
    && install -o root -g root -m 0755 /tmp/atlas /usr/local/bin/atlas

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

COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/inventory/out/inventory /usr/local/bin/inventory
# Copy policy bundle
COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/inventory/out/policy_bundle.tar.gz /rego/policy_bundle.tar.gz
# Copy database migrations
COPY --from=build --chown=nobody:nobody /go/src/github.com/open-edge-platform/infra-core/inventory/internal/ent/migrate/migrations /usr/share/migrations
# Copy atlas tool
COPY --from=build --chown=nobody:nobody /usr/local/bin/atlas /usr/local/bin/atlas


ENTRYPOINT ["inventory"]
