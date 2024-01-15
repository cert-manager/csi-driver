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

################
# Check Inputs #
################

ifndef helm_chart_source_dir
$(error helm_chart_source_dir is not set)
endif

################
# Add targets #
################

.PHONY: generate-crds
## Generate CRD manifests.
## @category [shared] Generate/ Verify
generate-crds: | $(NEEDS_CONTROLLER-GEN) $(NEEDS_YQ)
	$(eval crds_gen_temp := $(bin_dir)/scratch/crds)
	$(eval directories := $(shell ls -d */ | grep -v '_bin' | grep -v 'make'))

	rm -rf $(crds_gen_temp)
	mkdir -p $(crds_gen_temp)

	$(CONTROLLER-GEN) crd \
		$(directories:%=paths=./%...) \
		output:crd:artifacts:config=$(crds_gen_temp)

	echo "Updating CRDs with helm templating, writing to $(helm_chart_source_dir)/templates"

	@for i in $$(ls $(crds_gen_temp)); do \
		$(YQ) $(crds_gen_temp)/$$i > $(helm_chart_source_dir)/templates/crd-$$i; \
	done

shared_generate_targets += generate-crds
