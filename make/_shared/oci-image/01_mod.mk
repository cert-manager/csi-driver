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

oci_platforms := linux/amd64,linux/arm/v7,linux/arm64,linux/ppc64le

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# To get latest SHA run crane digest gcr.io/distroless/static-debian12:nonroot
base_image_static := gcr.io/distroless/static-debian12@sha256:39ae7f0201fee13b777a3e4a5a9326a8889269172c8b4f4289d9f19c831f45f4

# Use custom apko-built image as minimal base image to package the manager binary
# To get latest SHA run crane digest quay.io/jetstack/base-static-csi:latest
base_image_csi-static := quay.io/jetstack/base-static-csi@sha256:f8463a8a6d2265a15a982cf3d6cb08b685ab7220828885fbbb528135baa0a951

ifndef bin_dir
$(error bin_dir is not set)
endif

ifndef build_names
$(error build_names is not set)
endif

fatal_if_undefined = $(if $(findstring undefined,$(origin $1)),$(error $1 is not set))

define check_variables
$(call fatal_if_undefined,go_$1_ldflags)
$(call fatal_if_undefined,go_$1_source_path)
$(call fatal_if_undefined,oci_$1_base_image_flavor)
$(call fatal_if_undefined,oci_$1_image_name)
$(call fatal_if_undefined,oci_$1_image_name_development)

ifeq ($(oci_$1_base_image_flavor),static)
    oci_$1_base_image := $(base_image_static)
else ifeq ($(oci_$1_base_image_flavor),csi-static)
    oci_$1_base_image := $(base_image_csi-static)
else
    $$(error oci_$1_base_image_flavor has unknown value "$(oci_$1_base_image_flavor)")
endif

endef

$(foreach build_name,$(build_names),$(eval $(call check_variables,$(build_name))))

##########################################

RELEASE_DRYRUN ?= false

CGO_ENABLED ?= 0
GOEXPERIMENT ?=  # empty by default

oci_build_targets := $(build_names:%=oci-build-%)
oci_push_targets := $(build_names:%=oci-push-%)
oci_maybe_push_targets := $(build_names:%=oci-maybe-push-%)
oci_load_targets := $(build_names:%=oci-load-%)

image_tool_dir := $(dir $(lastword $(MAKEFILE_LIST)))/image_tool/

.PHONY: $(oci_build_targets)
## Build the OCI image.
## @category [shared] Build
$(oci_build_targets): oci-build-%: | $(NEEDS_KO) $(NEEDS_GO) $(NEEDS_YQ) $(bin_dir)/scratch/image
	$(eval oci_layout_path := $(bin_dir)/scratch/image/oci-layout-$*.$(oci_$*_image_tag))
	rm -rf $(CURDIR)/$(oci_layout_path)

	echo '{}' | \
		$(YQ) '.defaultBaseImage = "$(oci_$*_base_image)"' | \
		$(YQ) '.builds[0].id = "$*"' | \
		$(YQ) '.builds[0].main = "$(go_$*_source_path)"' | \
		$(YQ) '.builds[0].env[0] = "CGO_ENABLED={{.Env.CGO_ENABLED}}"' | \
		$(YQ) '.builds[0].env[1] = "GOEXPERIMENT={{.Env.GOEXPERIMENT}}"' | \
		$(YQ) '.builds[0].ldflags[0] = "-s"' | \
		$(YQ) '.builds[0].ldflags[1] = "-w"' | \
		$(YQ) '.builds[0].ldflags[2] = "{{.Env.LDFLAGS}}"' \
		> $(CURDIR)/$(oci_layout_path).ko_config.yaml

	KO_DOCKER_REPO=$(oci_$*_image_name_development) \
	KOCACHE=$(bin_dir)/scratch/image/ko_cache \
	KO_CONFIG_PATH=$(CURDIR)/$(oci_layout_path).ko_config.yaml \
	SOURCE_DATE_EPOCH=$(GITEPOCH) \
	KO_GO_PATH=$(GO) \
	LDFLAGS="$(go_$*_ldflags)" \
	CGO_ENABLED=$(CGO_ENABLED) \
	GOEXPERIMENT=$(GOEXPERIMENT) \
	$(KO) build $(go_$*_source_path) \
		--platform=$(oci_platforms) \
		--oci-layout-path=$(oci_layout_path) \
		--sbom-dir=$(CURDIR)/$(oci_layout_path).sbom \
		--sbom=spdx \
		--push=false \
		--bare

	cd $(image_tool_dir) && $(GO) run . list-digests \
		$(CURDIR)/$(oci_layout_path) \
		> $(CURDIR)/$(oci_layout_path).digests

.PHONY: $(oci_push_targets)
## Build and push OCI image.
## If the tag already exists, this target will overwrite it.
## If an identical image was already built before, we will add a new tag to it, but we will not sign it again.
## Expected pushed images:
## - :v1.2.3, @sha256:0000001
## - :v1.2.3.sig, :sha256-0000001.sig
## @category [shared] Build
$(oci_push_targets): oci-push-%: oci-build-% | $(NEEDS_CRANE) $(NEEDS_COSIGN) $(NEEDS_YQ) $(bin_dir)/scratch/image
	$(eval oci_layout_path := $(bin_dir)/scratch/image/oci-layout-$*.$(oci_$*_image_tag))
	$(eval image_ref := $(shell head -1 $(CURDIR)/$(oci_layout_path).digests))

ifneq ($(RELEASE_DRYRUN),true)
	if $(CRANE) image digest $(oci_$*_image_name)@$(image_ref) >/dev/null 2>&1; then \
		echo "Digest already exists, will retag without resigning."; \
		$(CRANE) push "$(oci_layout_path)" "$(oci_$*_image_name):$(oci_$*_image_tag)"; \
	else \
		echo "Digest does not yet exist, pushing image and signing."; \
		$(CRANE) push "$(oci_layout_path)" "$(oci_$*_image_name):$(oci_$*_image_tag)"; \
		$(COSIGN) sign --yes=true "$(oci_$*_image_name)@$(image_ref)"; \
	fi
endif

.PHONY: $(oci_maybe_push_targets)
## Run 'make oci-push-...' if tag does not already exist in registry.
## @category [shared] Build
$(oci_maybe_push_targets): oci-maybe-push-%: | $(NEEDS_CRANE)
	if $(CRANE) manifest digest $(oci_$*_image_name):$(oci_$*_image_tag) > /dev/null 2>&1; then \
		echo "Image $(oci_$*_image_name):$(oci_$*_image_tag) already exists in registry"; \
	else \
		echo "Image $(oci_$*_image_name):$(oci_$*_image_tag) does not exist in registry"; \
		$(MAKE) oci-push-$*; \
	fi

.PHONY: $(oci_load_targets)
## Build OCI image for the local architecture and load
## it into the $(kind_cluster_name) kind cluster.
## @category [shared] Build
$(oci_load_targets): oci_platforms := ""
$(oci_load_targets): oci-load-%: oci-build-% | kind-cluster $(NEEDS_KIND)
	$(eval oci_layout_path := $(bin_dir)/scratch/image/oci-layout-$*.$(oci_$*_image_tag))

	cd $(image_tool_dir) && $(GO) run . convert-to-docker-tar $(CURDIR)/$(oci_layout_path) $(CURDIR)/$(oci_layout_path).docker.tar $(oci_$*_image_name_development):$(oci_$*_image_tag)
	$(KIND) load image-archive --name $(kind_cluster_name) $(oci_layout_path).docker.tar
