# Copyright 2019 The Jetstack cert-manager contributors.
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

help:  ## display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: help build docker_build test depend verify all clean generate

all: test build image ## runs test, build and image build

clean: ## clean all bin data
	rm -rf ./bin

build: ## build cert-manager-csi
	GO111MODULE=on CGO_ENABLED=0 go build -v -o ./bin/cert-manager-csi ./cmd/.

verify: test boilerplate ## verify codebase

test: ## offline test cert-manager-csi
	go test -v ./pkg/...

boilerplate: ## verify boilerplate headers
	./hack/verify_boilerplate.py

image: build ## build cert-manager-csi docker image
	docker build -t gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.1 .

publish: image ## build cert-manager-csi docker image and publish image
	docker push gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.1

e2e: ## run end to end tests
	./test/run.sh
