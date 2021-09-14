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

BINDIR ?= $(CURDIR)/bin
ARCH   ?= $(shell go env GOARCH)
HELM_VERSION ?= 3.4.1

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OS := linux
endif
ifeq ($(UNAME_S),Darwin)
	OS := darwin
endif

help:  ## display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: help build docker_build test depend verify all clean generate

all: verify build image ## runs test, build and image build

clean: ## clean all bin data
	rm -rf ./bin

build: ## build cert-manager-csi-driver
	mkdir -p $(BINDIR)
	GO111MODULE=on CGO_ENABLED=0 go build -v -o ./bin/cert-manager-csi-driver ./cmd/.

verify: test boilerplate ## verify codebase

test: ## offline test cert-manager-csi-driver
	go test -v ./pkg/...

boilerplate: ## verify boilerplate headers
	./hack/verify-boilerplate.sh

image: build ## build cert-manager-csi-driver docker image
	docker build -t quay.io/jetstack/cert-manager-csi-driver:v0.1.0 .

e2e: depend ## run end to end tests
	./test/run.sh

depend: $(BINDIR)/helm

$(BINDIR)/helm:
	mkdir -p $(BINDIR)
	curl -o $(BINDIR)/helm.tar.gz -LO "https://get.helm.sh/helm-v$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz"
	tar -C $(BINDIR) -xzf $(BINDIR)/helm.tar.gz
	cp $(BINDIR)/$(OS)-$(ARCH)/helm $(BINDIR)/helm
	rm -r $(BINDIR)/$(OS)-$(ARCH) $(BINDIR)/helm.tar.gz
