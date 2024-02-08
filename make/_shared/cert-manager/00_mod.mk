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

cert_manager_version := v1.14.2

images_amd64 += quay.io/jetstack/cert-manager-controller:$(cert_manager_version)@sha256:8c5b6bfe84dfb0127aaab3b64ffcb4a2741aa5a7aa4736cc6edeb481a5ac47cf
images_amd64 += quay.io/jetstack/cert-manager-cainjector:$(cert_manager_version)@sha256:67ccc881f34b2d1dd3fa3d422d37bbb6934d268f08aefadb33e55cb0e515e270
images_amd64 += quay.io/jetstack/cert-manager-webhook:$(cert_manager_version)@sha256:dfc8027dd294d29bda073b3ceef06be4b4a0cada853b1eb170b4686e10fcce47
images_amd64 += quay.io/jetstack/cert-manager-ctl:$(cert_manager_version)@sha256:4e603416401ba94e773ec710236a417481c8f039336895a592e3727900a3dff0

images_arm64 += quay.io/jetstack/cert-manager-controller:$(cert_manager_version)@sha256:01a0388d5b1f15c8ae6cfd0672558697eef2992b956690a6cf335ccebeb646c2
images_arm64 += quay.io/jetstack/cert-manager-cainjector:$(cert_manager_version)@sha256:ffbb4641da6562c965db8e9cc28b653a89fc4bbbf22639d014bac6fc837db5e7
images_arm64 += quay.io/jetstack/cert-manager-webhook:$(cert_manager_version)@sha256:fdf548c247afbbf25ca8f74c5fde98f58976169504cf75c79d03ced552b93626
images_arm64 += quay.io/jetstack/cert-manager-ctl:$(cert_manager_version)@sha256:d0ab0106c62621b85614afae26c54e89d4bc5ac222b3402f7deb8f16fa0aa852
