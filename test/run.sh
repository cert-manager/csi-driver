#!/usr/bin/env bash

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

set -o nounset
set -o errexit
set -o pipefail

# Sets up the end-to-end test environment by:
# - creating a kind cluster
# - deploying cert-manager
# - deploying cert-manager-csi-driver
# The end-to-end test suite will then be run against this environment.
# The cluster will be deleted after tests have run.

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
cd "$REPO_ROOT"

ARTIFACTS="${ARTIFACTS:-$REPO_ROOT/_artifacts}"

BIN_DIR="$REPO_ROOT/bin"
mkdir -p "$BIN_DIR"
# install_multiplatform will install a binary for either Linux of macOS
# $1 = path to install to
# $2 = filename to save as
# $3 = linux-specific URL
# $4 = mac-specific URL
install_multiplatform() {
  case "$(uname -s)" in

   Darwin)
     curl -Lo "$1/$2" "$4"
     ;;

   Linux)
     curl -Lo "$1/$2" "$3"
     ;;

   *)
     echo 'Unsupported OS!'
     exit 1
     ;;
  esac

  chmod +x "$1/$2"
}

if ! command -v kind; then
  echo "'kind' command not found - installing..."
  install_multiplatform "${BIN_DIR}" kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-linux-amd64" "https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-darwin-amd64"
fi

if ! command -v kubectl; then
  echo "'kubectl' command not found - installing..."
  install_multiplatform "${BIN_DIR}" kubectl "https://dl.k8s.io/release/v1.16.15/bin/linux/amd64/kubectl" "https://dl.k8s.io/release/v1.16.15/bin/darwin/amd64/kubectl"
fi

if ! command -v go; then
  echo "'go' command not found - please install from https://golang.org"
  exit 1
fi

if ! command -v docker; then
  echo "'docker' command not found - please install from https://docker.com"
  exit 1
fi

export PATH="$BIN_DIR:$PATH"

CLUSTER_NAME="cert-manager-csi-driver-cluster"
exit_command() {
  kind export logs "${ARTIFACTS}" --name="$CLUSTER_NAME"
  if [ -z "${SKIP_CLEANUP:-}" ]; then
    kind delete cluster --name="$CLUSTER_NAME"
  else
    echo "Skipping cleanup due to SKIP_CLEANUP flag set - run 'kind delete cluster --name=$CLUSTER_NAME' to cleanup"
  fi
}
trap exit_command EXIT

echo "Pre-creating 'kind' docker network to avoid networking issues in CI"
# When running in our CI environment the Docker network's subnet choice will cause issues with routing
# This works this around till we have a way to properly patch this.
docker network create --driver=bridge --subnet=192.168.0.0/16 --gateway 192.168.0.1 kind || true
# Sleep for 2s to avoid any races between docker's network subcommand and 'kind create'
sleep 2

echo "Creating kind cluster named '$CLUSTER_NAME'"
# Kind image at 1.16.15, compatible with kind v0.11.1
kind create cluster --image=kindest/node@sha256:83067ed51bf2a3395b24687094e283a7c7c865ccc12a8b1d7aa673ba0c5e8861 --name="$CLUSTER_NAME"
export KUBECONFIG="$(kind get kubeconfig-path --name="$CLUSTER_NAME")"

CERT_MANAGER_MANIFEST_URL="https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml"
echo "Installing cert-manager in test cluster using manifest URL '$CERT_MANAGER_MANIFEST_URL'"
kubectl create -f "$CERT_MANAGER_MANIFEST_URL"

CERT_MANAGER_CSI_DOCKER_IMAGE="quay.io/jetstack/cert-manager-csi-driver"
CERT_MANAGER_CSI_DOCKER_TAG="canary"
echo "Building cert-manager-csi-driver container"
docker build -t "$CERT_MANAGER_CSI_DOCKER_IMAGE:$CERT_MANAGER_CSI_DOCKER_TAG" .

echo "Loading '$CERT_MANAGER_CSI_DOCKER_IMAGE:$CERT_MANAGER_CSI_DOCKER_TAG' image into kind cluster"
kind load docker-image --name="$CLUSTER_NAME" "$CERT_MANAGER_CSI_DOCKER_IMAGE:$CERT_MANAGER_CSI_DOCKER_TAG"

echo "Deploying cert-manager-csi-driver into test cluster"
./bin/helm upgrade --install -n cert-manager cert-manager-csi-driver ./deploy/charts/csi-driver --set image.repository=$CERT_MANAGER_CSI_DOCKER_IMAGE --set image.tag=$CERT_MANAGER_CSI_DOCKER_TAG

echo "Waiting 30s to allow Deployment & DaemonSet controllers to create pods"
sleep 30

kubectl get pods -A
echo "Waiting for all pods to be ready..."
kubectl wait --for=condition=Ready pod --all --all-namespaces --timeout=5m

echo "Executing end-to-end test suite"

# Export variables used by test suite
export REPO_ROOT
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
export CLUSTER_NAME
export KUBECTL=$(command -v kubectl)
go test -v -timeout 30m "./test/e2e/suite"
