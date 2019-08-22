help:  ## display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: help build docker_build test depend verify all clean generate

build: ## build cert-manager-csi
	GO111MODULE=on CGO_ENABLED=0 go build -v

docker: build ## build cert-manager-csi docker image
	docker build -t gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.0 .
