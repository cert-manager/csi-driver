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

cert_manager_version := v1.12.3

images_amd64 += quay.io/jetstack/cert-manager-controller:$(cert_manager_version)@sha256:6b9b696c2e56aaef5bf7e0b659ee91a773d0bb8f72b0eb4914a9db7e87578d47
images_amd64 += quay.io/jetstack/cert-manager-cainjector:$(cert_manager_version)@sha256:31ffa7640020640345a34f3fe6964560665e7ca89d818a6c455e63f5c4f5eb14
images_amd64 += quay.io/jetstack/cert-manager-webhook:$(cert_manager_version)@sha256:292facf28fd4f0db074fed12437669eef9c0ab8c1b9812d2c91e42b4a7448a36
images_amd64 += quay.io/jetstack/cert-manager-ctl:$(cert_manager_version)@sha256:5c985c4ebd8da6592cbe0249936f7513c0527488d754198699b3be9389b8b587

images_arm64 += quay.io/jetstack/cert-manager-controller:$(cert_manager_version)@sha256:f2adb86c11c305dcb78607cdf86fa232e657d196f82d0592799aebbfea22dec8
images_arm64 += quay.io/jetstack/cert-manager-cainjector:$(cert_manager_version)@sha256:118b985b0f0051ee9c428a3736c47bea92c3d8e7cb7c6eda881f7ecd4430cbed
images_arm64 += quay.io/jetstack/cert-manager-webhook:$(cert_manager_version)@sha256:0195441dc0f7f81e7514e6497bf68171bc54ef8481efc5fa0efe51892bd28c36
images_arm64 += quay.io/jetstack/cert-manager-ctl:$(cert_manager_version)@sha256:f376994ae17c519b12dd59c406a0abf8c6265c5f0c57431510eee15eaa40e4eb
