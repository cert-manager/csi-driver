help:  ## display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: help build docker_build test depend verify all clean generate

clean: ## clean all bin data
	rm -rf ./bin

build: ## build cert-manager-csi
	GO111MODULE=on CGO_ENABLED=0 go build -v -o ./bin/cert-manager-csi ./cmd/.

test: ## offline test cert-manager-csi
	go test -v ./pkg/...

image: build ## build cert-manager-csi docker image
	docker build -t gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.1 .

publish: image ## build cert-manager-csi docker image and publish image
	docker push gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.1

e2e: ## run end to end tests
	CERT_MANAGER_CSI_ROOT_PATH="$$(pwd)" go test -v ./test/e2e/suite/.

dev_cluster_create: ## create dev cluster for development testing
	CERT_MANAGER_CSI_ROOT_PATH="$$(pwd)" go run -v ./test/e2e/environment/dev create

dev_cluster_destroy: ## destroy dev cluster
	CERT_MANAGER_CSI_ROOT_PATH="$$(pwd)" go run -v ./test/e2e/environment/dev destroy
