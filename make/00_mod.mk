# Copyright 2023 The cert-manager Authors.
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

repo_name := github.com/cert-manager/csi-driver

kind_cluster_name := csi-driver
kind_cluster_config := $(bin_dir)/scratch/kind_cluster.yaml

build_names := manager

go_manager_main_dir := ./cmd
go_manager_mod_dir := .
go_manager_ldflags := -X $(repo_name)/internal/version.AppVersion=$(VERSION) -X $(repo_name)/internal/version.GitCommit=$(GITCOMMIT)
oci_manager_base_image_flavor := csi-static
oci_manager_image_name := quay.io/jetstack/cert-manager-csi-driver
oci_manager_image_tag := $(VERSION)
oci_manager_image_name_development := cert-manager.local/cert-manager-csi-driver

deploy_name := csi-driver
deploy_namespace := cert-manager

api_docs_outfile := docs/api/api.md
api_docs_package := $(repo_name)/pkg/apis/trust/v1alpha1
api_docs_branch := main

helm_chart_source_dir := deploy/charts/csi-driver
helm_chart_image_name := quay.io/jetstack/charts/cert-manager-csi-driver
helm_chart_version := $(VERSION)
helm_labels_template_name := cert-manager-csi-driver.labels

golangci_lint_config := .golangci.yaml

livenessprobe_image_name_source := registry.k8s.io/sig-storage/livenessprobe
livenessprobe_image_name := quay.io/jetstack/livenessprobe
livenessprobe_image_tag := v2.18.0

nodedriverregistrar_image_name_source := registry.k8s.io/sig-storage/csi-node-driver-registrar
nodedriverregistrar_image_name := quay.io/jetstack/csi-node-driver-registrar
nodedriverregistrar_image_tag := v2.16.0

define helm_values_mutation_function
$(YQ) \
	'( .livenessProbeImage._defaultReference = ":$(livenessprobe_image_tag)" ) | \
	( .nodeDriverRegistrarImage._defaultReference = ":$(nodedriverregistrar_image_tag)" )' \
	$1 --inplace
endef
