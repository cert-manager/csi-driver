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

images_amd64 ?=
images_arm64 ?=

cert_manager_version := v1.13.3

images_amd64 += quay.io/jetstack/cert-manager-controller:$(cert_manager_version)@sha256:3490283b7a8fb4e13d9864963f94ab710fc2d669d3d53608f01608c887d4f741
images_amd64 += quay.io/jetstack/cert-manager-cainjector:$(cert_manager_version)@sha256:c3a5ce22b8521e1a0a792152540cbbcb3ef8d509c9f3583cffb6b4e9a5b7bd60
images_amd64 += quay.io/jetstack/cert-manager-webhook:$(cert_manager_version)@sha256:afe9a27be1e6b3847d6483eb9a83b20fb8576ba5c314f381a90b185af022a105
images_amd64 += quay.io/jetstack/cert-manager-ctl:$(cert_manager_version)@sha256:38f549eba224094c6810c088e5f8e257690dc882956234b8db3cad15c6253822

images_arm64 += quay.io/jetstack/cert-manager-controller:$(cert_manager_version)@sha256:2ec91011538846690da6c236e51ef0221b2e3dbd45de41cce6dfa16d531a4dc4
images_arm64 += quay.io/jetstack/cert-manager-cainjector:$(cert_manager_version)@sha256:1bdddcf53317991f01be03cffc126b10df4136556e0440836bee07a192dcc3f5
images_arm64 += quay.io/jetstack/cert-manager-webhook:$(cert_manager_version)@sha256:04c79086f1e3440bac8b584304fe5444d6184c5345c7e7115147ee39d4591d2e
images_arm64 += quay.io/jetstack/cert-manager-ctl:$(cert_manager_version)@sha256:f0755d949d0facd64550d2f1e2f974ce5592a199b84772fd9ab4a97a2a19a609
