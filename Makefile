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

.DEFAULT_GOAL := help

BINDIR ?= $(CURDIR)/bin
ARCH   ?= $(shell go env GOARCH)
HELM_VERSION ?= 3.4.1
IMAGE_PLATFORMS ?= linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OS := linux
endif
ifeq ($(UNAME_S),Darwin)
	OS := darwin
endif

.PHONY: help
help:  ## display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: all
all: verify build image chart-readme  ## runs test, build and image build

.PHONY: clean
clean: ## clean all bin data
	rm -rf ./bin

.PHONY: build
build:  | $(BINDIR) ## build cert-manager-csi-driver
	GO111MODULE=on CGO_ENABLED=0 go build -v -o $(BINDIR)/cert-manager-csi-driver ./cmd/.

.PHONY: verify
verify: test boilerplate ## verify codebase

.PHONY: test
test: ## offline test cert-manager-csi-driver
	go test -v ./pkg/...

.PHONY: boilerplate
boilerplate: ## verify boilerplate headers
	./hack/verify-boilerplate.sh

# image will only build and store the image locally, targeted in OCI format.
# To actually push an image to the public repo, replace the `--output` flag and
# arguments to `--push`.
.PHONY: image
image: ## build cert-manager-csi-driver docker image targeting all supported platforms
	docker buildx build --platform=$(IMAGE_PLATFORMS) -t quay.io/jetstack/cert-manager-csi-driver:v0.3.0 --output type=oci,dest=./bin/cert-manager-csi-driver-oci .

.PHONY: e2e
e2e: depend ## run end to end tests
	./test/run.sh

CHART_YAML := $(shell find deploy/charts/csi-driver -name "*.yaml")

.PHONY: chart-readme
chart-readme: deploy/charts/csi-driver/README.md  ## update helm chart README file

deploy/charts/csi-driver/README.md: $(BINDIR)/helm-docs $(CHART_YAML)
	$(BINDIR)/helm-docs

.PHONY: depend
depend: $(BINDIR)/helm $(BINDIR)/helm-docs

$(BINDIR)/helm: | $(BINDIR)
	curl -o $(BINDIR)/helm.tar.gz -LO "https://get.helm.sh/helm-v$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz"
	tar -C $(BINDIR) -xzf $(BINDIR)/helm.tar.gz
	cp $(BINDIR)/$(OS)-$(ARCH)/helm $(BINDIR)/helm
	rm -r $(BINDIR)/$(OS)-$(ARCH) $(BINDIR)/helm.tar.gz

HELM_DOCS_VERSION=1.10.0

$(BINDIR)/helm-docs: | $(BINDIR)
	GOBIN=$(BINDIR) go install github.com/norwoodj/helm-docs/cmd/helm-docs@v1.10.0

$(BINDIR):
	mkdir -p $@
