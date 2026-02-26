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

$(kind_cluster_config): make/config/kind/cluster.yaml | $(bin_dir)/scratch
	cat $< | \
	sed -e 's|{{KIND_IMAGES}}|$(CURDIR)/$(images_tar_dir)|g' \
	> $@

include make/test-e2e.mk
include make/test-unit.mk

.PHONY: release
## Publish all release artifacts (image + helm chart)
## @category [shared] Release
release: | $(NEEDS_CRANE)
	$(MAKE) oci-push-manager
	$(CRANE) "$(livenessprobe_image_name_source):$(livenessprobe_image_tag)" "$(livenessprobe_image_name):$(livenessprobe_image_tag)"
	$(CRANE) "$(nodedriverregistrar_image_name_source):$(nodedriverregistrar_image_tag)" "$(nodedriverregistrar_image_name):$(nodedriverregistrar_image_tag)"
	$(MAKE) helm-chart-oci-push

	@echo "RELEASE_OCI_MANAGER_IMAGE=$(oci_manager_image_name)" >> "$(GITHUB_OUTPUT)"
	@echo "RELEASE_OCI_MANAGER_TAG=$(oci_manager_image_tag)" >> "$(GITHUB_OUTPUT)"
	@echo "RELEASE_HELM_CHART_IMAGE=$(helm_chart_image_name)" >> "$(GITHUB_OUTPUT)"
	@echo "RELEASE_HELM_CHART_VERSION=$(helm_chart_version)" >> "$(GITHUB_OUTPUT)"

	@echo "Release complete!"
