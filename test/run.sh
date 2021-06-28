#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

# Sets up the end-to-end test environment by:
# - creating a kind cluster
# - deploying cert-manager
# - deploying cert-manager-csi
# The end-to-end test suite will then be run against this environment.
# The cluster will be deleted after tests have run.

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
cd "$REPO_ROOT"

if ! command -v kind; then
  echo "'kind' command not found - please install from https://kind.sigs.k8s.io"
  exit 1
fi

if ! command -v kubectl; then
  echo "'kubectl' command not found - please install from https://k8s.io"
fi

if ! command -v go; then
  echo "'go' command not found - please install from https://golang.org"
fi

if ! command -v docker; then
  echo "'docker' command not found - please install from https://docker.com"
fi

CLUSTER_NAME="cert-manager-csi-cluster"
if [ -z "${SKIP_CLEANUP:-}" ]; then
  trap "kind delete cluster --name=$CLUSTER_NAME" EXIT
else
  echo "Skipping cleanup due to SKIP_CLEANUP flag set - run 'kind delete cluster --name=$CLUSTER_NAME' to cleanup"
fi
echo "Creating kind cluster named '$CLUSTER_NAME'"
kind create cluster --image=kindest/node:v1.16.15 --name="$CLUSTER_NAME"

CERT_MANAGER_MANIFEST_URL="https://github.com/jetstack/cert-manager/releases/download/v0.12.0/cert-manager.yaml"
echo "Installing cert-manager in test cluster using manifest URL '$CERT_MANAGER_MANIFEST_URL'"
kubectl create -f "$CERT_MANAGER_MANIFEST_URL"

echo "Building cert-manager-csi binary"
mkdir -p bin/
GOARCH=amd64 GOOS=linux go build -o ./bin/cert-manager-csi ./cmd

CERT_MANAGER_CSI_DOCKER_IMAGE="gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.1"
echo "Building cert-manager-csi container"
docker build -t "$CERT_MANAGER_CSI_DOCKER_IMAGE" .

echo "Loading '$CERT_MANAGER_CSI_DOCKER_IMAGE' image into kind cluster"
kind load docker-image "$CERT_MANAGER_CSI_DOCKER_IMAGE"

echo "Deploying cert-manager-csi into test cluster"
kubectl create -f "./deploy/cert-manager-csi-driver.yaml"

echo "Waiting 5s to allow Deployment controller to create pods"
sleep 5

echo "Waiting for all pods to be ready..."
kubectl wait --for=condition=Ready pod --all --all-namespaces --timeout=5m

echo "Executing end-to-end test suite"

# Export variables used by test suite
export REPO_ROOT
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
export CLUSTER_NAME
export KUBECTL=$(command -v kubectl)
go test -v "./test/e2e/suite"
