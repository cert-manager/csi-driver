<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="256" width="256" alt="cert-manager project logo" />
</p>
<p align="center">
  <a href="https://godoc.org/github.com/cert-manager/csi-driver"><img src="https://godoc.org/github.com/cert-manager/csi-driver?status.svg" alt="csi-driver godoc"></a>
  <a href="https://goreportcard.com/report/github.com/cert-manager/csi-driver"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/cert-manager/csi-driver" /></a>
  <a href="https://artifacthub.io/packages/search?repo=cert-manager"><img alt="Artifact Hub" src="https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/cert-manager" /></a>
</p>

# csi-driver

csi-driver is a Container Storage Interface (CSI) driver plugin for Kubernetes
to work along [cert-manager](https://cert-manager.io/). The goal for this plugin
is to facilitate requesting and mounting certificate key pairs to pods
seamlessly. This is useful for facilitating mTLS, or otherwise securing
connections of pods with guaranteed present certificates whilst having all of
the features that cert-manager provides.

## Why a CSI Driver?

- Ensure private keys never leave the node and are never sent over the network.
  All private keys are stored locally on the node.
- Unique key and certificate per application replica with a grantee to be
  present on application run time.
- Reduce resource management overhead by defining certificate request spec
  in-line of the Kubernetes Pod template.
- Automatic renewal of certificates based on expiry of each individual
  certificate.
- Keys and certificates are destroyed during application termination.
- Scope for extending plugin behaviour with visibility on each replica's
  certificate request and termination.

## Documentation

Please follow the documentation at
[cert-manager.io](https://cert-manager.io/docs/projects/csi-driver/) for
installing and using csi-driver.

## Release Process

There is a semi-automated release process for csi-driver.
When you create a Git tag with a tagname that has a `v` prefix and push it to GitHub
it will trigger the [release workflow].

This will:

1. Create and push a Docker image to `quay.io/jetstack/cert-manager-csi-driver:${{ github.ref_name }}`
2. Create a Helm chart
3. Create a *draft* GitHub release with the Helm chart file attached and containing a reference to the Docker image.

To perform a release:

1. Create and push a Git tag

   ```sh
   export VERSION=v0.5.0-alpha.0
   git tag --annotate --message="Release ${VERSION}" "${VERSION}"
   git push origin "${VERSION}"
   ```

2. Wait for the [release workflow] to succeed and if successful visit the draft release page to download the attached Helm chart attachment.

3. Create a PR in the [jetstack/jetstack-charts repository on GitHub](https://github.com/jetstack/jetstack-charts), containing the Helm chart file that is attached to the draft GitHub release. This is only currently possible for maintainers inside Venafi, but will be changed in the future.

4. Wait for the PR to be merged and verify that the Helm chart is available from https://charts.jetstack.io.

5. Visit the [releases page], edit the draft release, click "Generate release notes", and publish the release.

[release workflow]: https://github.com/cert-manager/csi-driver/actions/workflows/release.yaml
[releases page]: https://github.com/cert-manager/csi-driver/releases
