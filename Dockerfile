# Copyright 2021 The cert-manager Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build the cert-manager-csi-driver binary
FROM docker.io/library/golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source files
COPY Makefile Makefile
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/

# Build
RUN make build

FROM alpine:3.15
LABEL description="cert-manager CSI Driver"

WORKDIR /

COPY --from=builder /workspace/bin/cert-manager-csi-driver /usr/bin/cert-manager-csi-driver

# Add util-linux to get a new version of losetup.
RUN apk add util-linux

ENTRYPOINT ["/usr/bin/cert-manager-csi-driver"]
