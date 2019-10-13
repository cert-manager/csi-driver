# Copyright 2019 The Jetstack cert-manager contributors.
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

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file", "http_archive")
load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("@bazel_gazelle//:deps.bzl", "go_repository")

def install():
    install_misc()
    install_kubectl()
    install_kind()

def install_misc():
    http_file(
        name = "jq_linux",
        executable = 1,
        sha256 = "c6b3a7d7d3e7b70c6f51b706a3b90bd01833846c54d32ca32f0027f00226ff6d",
        urls = ["https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64"],
    )

    http_file(
        name = "jq_osx",
        executable = 1,
        sha256 = "386e92c982a56fe4851468d7a931dfca29560cee306a0e66c6a1bd4065d3dac5",
        urls = ["https://github.com/stedolan/jq/releases/download/jq-1.5/jq-osx-amd64"],
    )

# Define rules for different kubectl versions
def install_kubectl():
    http_file(
        name = "kubectl_1_12_darwin",
        executable = 1,
        sha256 = "ccddf5b78cd24d5782f4fbe436eee974ca3d901a2d850c24693efa8824737979",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.12.3/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_12_linux",
        executable = 1,
        sha256 = "a93cd2ffd146bbffb6ea651b71b57fe377ba1f158c7c0eb16c14aa93394cd576",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.12.3/bin/linux/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_13_darwin",
        executable = 1,
        sha256 = "e656a8ac9272d04febf2ed29b2e8866bfdb73f55e098026384268851d7aeba74",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.13.2/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_13_linux",
        executable = 1,
        sha256 = "2c7ab398559c7f4f91102c4a65184e0a5a3a137060c3179e9361d9c20b467181",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.13.2/bin/linux/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_14_darwin",
        executable = 1,
        sha256 = "b4f6d583014f3dc9f3912d68b5aaa20a25394ecc43b42b2df3d37ef7c4a6f819",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.14.3/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_14_linux",
        executable = 1,
        sha256 = "ebc8c2fadede148c2db1b974f0f7f93f39f19c8278619893fd530e20e9bec98f",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.14.3/bin/linux/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_15_darwin",
        executable = 1,
        sha256 = "63f1ace419edffa1f5ebb64a6c63597afd48f8d94a61d4fb44e820139adbbe54",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_15_linux",
        executable = 1,
        sha256 = "ecec7fe4ffa03018ff00f14e228442af5c2284e57771e4916b977c20ba4e5b39",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_16_darwin",
        executable = 1,
        sha256 = "ab04b4e950fb7a8fa24da1d646af6d2fd7c1c7f09254af3783c920d258a94b1a",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.16.0-alpha.1/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_16_linux",
        executable = 1,
        sha256 = "05942f4d57305dedeb76102a8d7ba0476914a1cd373e51d503923e6c96c4dc45",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.16.0-alpha.1/bin/linux/amd64/kubectl"],
    )

## Fetch kind images used during e2e tests
def install_kind():
    # install kind binary
    http_file(
        name = "kind_darwin",
        executable = 1,
        sha256 = "023f1886207132dcfc62139a86f09488a79210732b00c9ec6431d6f6b7e9d2d3",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.4.0/kind-darwin-amd64"],
    )

    http_file(
        name = "kind_linux",
        executable = 1,
        sha256 = "a97f7d6d97bc0e261ea85433ca564269f117baf0fae051f16b296d2d7541f8dd",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.4.0/kind-linux-amd64"],
    )

    container_pull(
        name = "kind-1.12",
        registry = "index.docker.io",
        repository = "kindest/node",
        tag = "v1.12.9",
        digest = "sha256:bcb79eb3cd6550c1ba9584ce57c832dcd6e442913678d2785307a7ad9addc029",
    )

    container_pull(
        name = "kind-1.13",
        registry = "index.docker.io",
        repository = "kindest/node",
        tag = "v1.13.7",
        digest = "sha256:f3f1cfc2318d1eb88d91253a9c5fa45f6e9121b6b1e65aea6c7ef59f1549aaaf",
    )

    container_pull(
        name = "kind-1.14",
        registry = "index.docker.io",
        repository = "kindest/node",
        tag = "v1.14.3",
        digest = "sha256:583166c121482848cd6509fbac525dd62d503c52a84ff45c338ee7e8b5cfe114",
    )

    container_pull(
        name = "kind-1.15",
        registry = "index.docker.io",
        repository = "kindest/node",
        tag = "v1.15.0",
        digest = "sha256:b4d092fd2b507843dd096fe6c85d06a27a0cbd740a0b32a880fe61aba24bb478",
    )

    container_pull(
        name = "kind-1.16",
        registry = "eu.gcr.io",
        repository = "jetstack-build-infra-images/kind-node",
        tag = "1.16.0-alpha.1",
        digest = "sha256:b9775b688fda2e6434cda1b9016baf876f381a8325961f59b9ae238166259885",
    )
