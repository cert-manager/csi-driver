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

oci_platforms ?= linux/amd64,linux/arm/v7,linux/arm64,linux/ppc64le

# Use distroless as minimal base image to package the manager binary
# To get latest SHA run crane digest quay.io/jetstack/base-static:latest
base_image_static := quay.io/jetstack/base-static@sha256:ba3cff0a4cacc5ae564e04c1f645000e8c9234c0f4b09534be1dee7874a42141

# Use custom apko-built image as minimal base image to package the manager binary
# To get latest SHA run crane digest quay.io/jetstack/base-static-csi:latest
base_image_csi-static := quay.io/jetstack/base-static-csi@sha256:54bacd13cccc385ef66730dbc7eb13bdb6a9ff8853e7f551d025ccb0e8c6bf83

ifndef bin_dir
$(error bin_dir is not set)
endif

ifndef build_names
$(error build_names is not set)
endif

fatal_if_undefined = $(if $(findstring undefined,$(origin $1)),$(error $1 is not set))

define check_variables
$(call fatal_if_undefined,go_$1_ldflags)
$(call fatal_if_undefined,go_$1_main_dir)
$(call fatal_if_undefined,go_$1_mod_dir)
$(call fatal_if_undefined,oci_$1_base_image_flavor)
$(call fatal_if_undefined,oci_$1_image_name)
$(call fatal_if_undefined,oci_$1_image_name_development)

ifeq ($(oci_$1_base_image_flavor),static)
    oci_$1_base_image := $(base_image_static)
else ifeq ($(oci_$1_base_image_flavor),csi-static)
    oci_$1_base_image := $(base_image_csi-static)
else ifeq ($(oci_$1_base_image_flavor),custom)
    $$(call fatal_if_undefined,oci_$1_base_image)
else
    $$(error oci_$1_base_image_flavor has unknown value "$(oci_$1_base_image_flavor)")
endif

ifneq ($(go_$1_main_dir:.%=.),.)
$$(error go_$1_main_dir "$(go_$1_main_dir)" should be a directory path that DOES start with ".")
endif
ifeq ($(go_$1_main_dir:%/=/),/)
$$(error go_$1_main_dir "$(go_$1_main_dir)" should be a directory path that DOES NOT end with "/")
endif
ifeq ($(go_$1_main_dir:%.go=.go),.go)
$$(error go_$1_main_dir "$(go_$1_main_dir)" should be a directory path that DOES NOT end with ".go")
endif
ifneq ($(go_$1_mod_dir:.%=.),.)
$$(error go_$1_mod_dir "$(go_$1_mod_dir)" should be a directory path that DOES start with ".")
endif
ifeq ($(go_$1_mod_dir:%/=/),/)
$$(error go_$1_mod_dir "$(go_$1_mod_dir)" should be a directory path that DOES NOT end with "/")
endif
ifeq ($(go_$1_mod_dir:%.go=.go),.go)
$$(error go_$1_mod_dir "$(go_$1_mod_dir)" should be a directory path that DOES NOT end with ".go")
endif

endef

$(foreach build_name,$(build_names),$(eval $(call check_variables,$(build_name))))

##########################################

RELEASE_DRYRUN ?= false

CGO_ENABLED ?= 0
GOEXPERIMENT ?=  # empty by default

COSIGN_FLAGS ?= # empty by default
OCI_SIGN_ON_PUSH ?= true

oci_build_targets := $(build_names:%=oci-build-%)
oci_push_targets := $(build_names:%=oci-push-%)
oci_push_no_sign_targets := $(build_names:%=oci-push-no-sign-%)
oci_sign_targets := $(build_names:%=oci-sign-%)
oci_maybe_push_targets := $(build_names:%=oci-maybe-push-%)
oci_load_targets := $(build_names:%=oci-load-%)
docker_tarball_targets := $(build_names:%=docker-tarball-%)

$(foreach build_name,$(build_names),$(eval oci_layout_path_$(build_name) := $(bin_dir)/scratch/image/oci-layout-$(build_name).$(oci_$(build_name)_image_tag)))
$(foreach build_name,$(build_names),$(eval docker_tarball_path_$(build_name) := $(CURDIR)/$(oci_layout_path_$(build_name)).docker.tar))

image_tool_dir := $(dir $(lastword $(MAKEFILE_LIST)))/image_tool/

.PHONY: $(oci_build_targets)
## Build the OCI image.
## @category [shared] Build
$(oci_build_targets): oci-build-%: | $(NEEDS_KO) $(NEEDS_GO) $(NEEDS_YQ) $(bin_dir)/scratch/image
	rm -rf $(CURDIR)/$(oci_layout_path_$*)

	@if [ ! -f "$(go_$*_mod_dir)/go.mod" ]; then \
		echo "ERROR: Specified directory "$(go_$*_mod_dir)" does not contain a go.mod file."; \
		exit 1; \
	fi

	@if [ ! -f "$(go_$*_mod_dir)/$(go_$*_main_dir)/main.go" ]; then \
		echo "ERROR: Specified directory "$(go_$*_mod_dir)$(go_$*_main_dir)" does not contain a main.go file."; \
		exit 1; \
	fi

	echo '{}' | \
		$(YQ) '.defaultBaseImage = "$(oci_$*_base_image)"' | \
		$(YQ) '.builds[0].id = "$*"' | \
		$(YQ) '.builds[0].dir = "$(go_$*_mod_dir)"' | \
		$(YQ) '.builds[0].main = "$(go_$*_main_dir)"' | \
		$(YQ) '.builds[0].env[0] = "CGO_ENABLED=$(CGO_ENABLED)"' | \
		$(YQ) '.builds[0].env[1] = "GOEXPERIMENT=$(GOEXPERIMENT)"' | \
		$(YQ) '.builds[0].ldflags[0] = "-s"' | \
		$(YQ) '.builds[0].ldflags[1] = "-w"' | \
		$(YQ) '.builds[0].ldflags[2] = "{{.Env.LDFLAGS}}"' \
		> $(CURDIR)/$(oci_layout_path_$*).ko_config.yaml

	GOWORK=off \
	KO_DOCKER_REPO=$(oci_$*_image_name_development) \
	KOCACHE=$(CURDIR)/$(bin_dir)/scratch/image/ko_cache \
	KO_CONFIG_PATH=$(CURDIR)/$(oci_layout_path_$*).ko_config.yaml \
	SOURCE_DATE_EPOCH=$(GITEPOCH) \
	KO_GO_PATH=$(GO) \
	LDFLAGS="$(go_$*_ldflags)" \
	$(KO) build $(go_$*_mod_dir)/$(go_$*_main_dir) \
		--platform=$(oci_platforms) \
		--oci-layout-path=$(oci_layout_path_$*) \
		--sbom-dir=$(CURDIR)/$(oci_layout_path_$*).sbom \
		--sbom=spdx \
		--push=false \
		--bare

	cd $(image_tool_dir) && GOWORK=off $(GO) run . list-digests \
		$(CURDIR)/$(oci_layout_path_$*) \
		> $(CURDIR)/$(oci_layout_path_$*).digests

# Function for ensuring the .digests file exists. In the use case where pushing
# and signing happen independently, we need to ensure this file exists for 
# signing
define oci_digest_ensure 
ifeq ($(call oci_digest,$1),)
$$(error "$(oci_layout_path_$1).digests" does not exist, has this image been built?)
endif
endef

# Functions for pushing and signing. We have a few targets that push/sign, this
# use of functions means we can define the commands once.
oci_digest = $(shell head -1 $(CURDIR)/$(oci_layout_path_$1).digests)
oci_push_command = $(foreach oci_image_name,$(oci_$1_image_name),$(CRANE) push "$(oci_layout_path_$1)" "$(oci_image_name):$(oci_$1_image_tag)";)
oci_sign_command = $(foreach oci_image_name,$(oci_$1_image_name),$(COSIGN) sign --yes=true $(COSIGN_FLAGS) "$(oci_image_name)@$(call oci_digest,$1)";)

.PHONY: $(oci_push_targets)
## Build and push OCI image.
## If the tag already exists, this target will overwrite it.
## If an identical image was already built before, we will add a new tag to it, but we will not sign it again.
## Expected pushed images:
## - :v1.2.3, @sha256:0000001
## - :v1.2.3.sig, :sha256-0000001.sig
## @category [shared] Build
$(oci_push_targets): oci-push-%: oci-build-% | $(NEEDS_CRANE) $(NEEDS_COSIGN) $(NEEDS_YQ) $(bin_dir)/scratch/image
ifneq ($(RELEASE_DRYRUN),true)
	if $(CRANE) image digest $(oci_$*_image_name)@$(call oci_digest,$*) >/dev/null 2>&1; then \
		echo "Digest already exists, will retag without resigning."; \
		$(call oci_push_command,$*) \
	else \
		echo "Digest does not yet exist, pushing image and signing."; \
		$(call oci_push_command,$*) \
		$(call oci_sign_command,$*) \
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

.PHONY: $(oci_push_no_sign_targets)
## Build and push OCI image.
## If the tag already exists, this target will overwrite it.
## If an identical image was already built before, we will add a new tag to it.
## This target will not sign the image
## Expected pushed images:
## - :v1.2.3, @sha256:0000001
## @category [shared] Build
$(oci_push_no_sign_targets): oci-push-no-sign-%: oci-build-% | $(NEEDS_CRANE) $(bin_dir)/scratch/image
	$(call oci_push_command,$*)

.PHONY: $(oci_sign_targets)
## Run 'make oci-sign-...' to force a sign of the image.
## @category [shared] Build
$(oci_sign_targets): oci-sign-%: | $(NEEDS_COSIGN)
	$(eval $(call oci_digest_ensure,$*)) 
	$(call oci_sign_command,$*)

.PHONY: $(oci_load_targets)
## Build OCI image for the local architecture and load
## it into the $(kind_cluster_name) kind cluster.
## @category [shared] Build
$(oci_load_targets): oci-load-%: docker-tarball-% | kind-cluster $(NEEDS_KIND)
	$(KIND) load image-archive --name $(kind_cluster_name) $(docker_tarball_path_$*)

## Build Docker tarball image for the local architecture
## @category [shared] Build
.PHONY: $(docker_tarball_targets)
$(docker_tarball_targets): oci_platforms := ""
$(docker_tarball_targets): docker-tarball-%: oci-build-%
	cd $(image_tool_dir) && GOWORK=off $(GO) run . convert-to-docker-tar $(CURDIR)/$(oci_layout_path_$*) $(docker_tarball_path_$*) $(oci_$*_image_name_development):$(oci_$*_image_tag)
