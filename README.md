<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d7a3a3976785bd5717d9d06b115878feaf257597/logo/logo.png" height="241" width="250" alt="cert-manager project logo" />
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

## Requirements and Installation

This CSI driver plugin makes use of the 'CSI inline volume' feature - Alpha as
of `v1.15` and beta in `v1.16`. Kubernetes versions `v1.16` and higher require
no extra configuration however `v1.15` requires the following feature gate set:
```
--feature-gates=CSIInlineVolume=true
```

You must have a working installation of cert-manager present on the cluster.
Instructions on how to install cert-manager can be found
[on cert-manager.io](https://cert-manager.io/docs/installation/).

To install the csi-driver, use helm install:

```terminal
 $ helm repo add jetstack https://charts.jetstack.io --force-update
 $ helm upgrade -i -n cert-manager cert-manager-csi-driver jetstack/cert-manager-csi-driver --wait
```

Or apply the static manifests to your cluster:

```terminal
 $ helm template ./deploy/charts/csi-driver | kubectl apply -n cert-manager -f -
```


You can verify the installation has completed correctly by checking the presence
of the CSIDriver resource as well as a CSINode resource present for each node,
referencing `csi.cert-manager.io`.

```
$ kubectl get csidrivers
NAME                     CREATED AT
csi.cert-manager.io   2019-09-06T16:55:19Z

$ kubectl get csinodes -o yaml
apiVersion: v1
items:
- apiVersion: storage.k8s.io/v1beta1
  kind: CSINode
  metadata:
    name: kind-control-plane
    ownerReferences:
    - apiVersion: v1
      kind: Node
      name: kind-control-plane
...
  spec:
    drivers:
    - name: csi.cert-manager.io
      nodeID: kind-control-plane
      topologyKeys: null
...
```

The CSI driver is now installed and is ready to be used for pods in the cluster.

## Requesting and Mounting Certificates

To request certificates from cert-manager, simply define a volume mount where
the key and certificate will be written to, along with a volume with attributes
that define the cert-manager request. The following is a dummy app that mounts a
key certificate pair to `/tls` and has been signed by the `ca-issuer` with a DNS
name valid for `my-service.sandbox.svc.cluster.local`.

```
apiVersion: v1
kind: Pod
metadata:
  name: my-csi-app
  namespace: sandbox
  labels:
    app: my-csi-app
spec:
  containers:
    - name: my-frontend
      image: busybox
      volumeMounts:
      - mountPath: "/tls"
        name: tls
      command: [ "sleep", "1000000" ]
  volumes:
    - name: tls
      csi:
        driver: csi.cert-manager.io
        volumeAttributes:
              csi.cert-manager.io/issuer-name: ca-issuer
              csi.cert-manager.io/dns-names: {{.POD_NAME}}.{{.POD_NAMESPACE}}.svc.cluster.local
```

Once created, the CSI driver will generate a private key locally, request a
certificate from cert-manager based on the given attributes, then store both
locally to be mounted to the pod. The pod will remain in a pending state until
this process has been completed.

For more information on how to set up issuers for your cluster, refer to the
cert-manager documentation
[here](https://cert-manager.io/docs/configuration/). **Note** it is not
possible to use `SelfSigned` Issuers with the CSI Driver. In order for
cert-manager to self sign a certificate, it needs access to the secret
containing the private key that signed the certificate request to sign the end
certificate. This secret is not used and so not available in the CSI driver use
case.

## Supported Volume Attributes

The csi-driver driver aims to have complete feature parity with all possible
values available through the cert-manager API however currently supports the
following values;

| Attribute                               | Description                                                                                                                | Default                              | Example                          |
|-----------------------------------------|----------------------------------------------------------------------------------------------------------------------------|--------------------------------------|----------------------------------|
| `csi.cert-manager.io/issuer-name`       | The Issuer name to sign the certificate request.                                                                           |                                      | `ca-issuer`                      |
| `csi.cert-manager.io/issuer-kind`       | The Issuer kind to sign the certificate request.                                                                           | `Issuer`                             | `ClusterIssuer`                  |
| `csi.cert-manager.io/issuer-group`      | The group name the Issuer belongs to.                                                                                      | `cert-manager.io`                    | `out.of.tree.foo`                |
| `csi.cert-manager.io/common-name`       | Certificate common name (supports variables).                                                                              |                                      | `my-cert.foo`                    |
| `csi.cert-manager.io/dns-names`         | DNS names the certificate will be requested for. At least a DNS Name, IP or URI name must be present (supports variables). |                                      | `a.b.foo.com,c.d.foo.com`        |
| `csi.cert-manager.io/ip-sans`           | IP addresses the certificate will be requested for.                                                                        |                                      | `192.0.0.1,192.0.0.2`            |
| `csi.cert-manager.io/uri-sans`          | URI names the certificate will be requested for (supports variables).                                                      |                                      | `spiffe://foo.bar.cluster.local` |
| `csi.cert-manager.io/duration`          | Requested duration the signed certificate will be valid for.                                                               | `720h`                               | `1880h`                          |
| `csi.cert-manager.io/is-ca`             | Mark the certificate as a certificate authority.                                                                           | `false`                              | `true`                           |
| `csi.cert-manager.io/key-usages`        | Set the key usages on the certificate request.                                                                             | `digital signature,key encipherment` | `server auth,client auth`        |
| `csi.cert-manager.io/key-encoding`      | Set the key encoding format (PKCS1 or PKCS8).                                                                              | `PKCS1`                              | `PKCS8`                          |
| `csi.cert-manager.io/certificate-file`  | File name to store the certificate file at.                                                                                | `tls.crt`                            | `bar/foo.crt`                    |
| `csi.cert-manager.io/ca-file`           | File name to store the ca certificate file at.                                                                             | `ca.crt`                             | `bar/foo.ca`                     |
| `csi.cert-manager.io/privatekey-file`   | File name to store the key file at.                                                                                        | `tls.key`                            | `bar/foo.key`                    |
| `csi.cert-manager.io/fs-group`          | Set the FS Group of written files. Should be paired with and match the value of the consuming container `runAsGroup`.      |                                      | `2000`                           |
| `csi.cert-manager.io/renew-before`      | The time to renew the certificate before expiry. Defaults to a third of the requested duration.                            | `$CERT_DURATION/3`                   | `72h`                            |
| `csi.cert-manager.io/reuse-private-key` | Re-use the same private when when renewing certificates.                                                                   | `false`                              | `true`                           |

### Variables

The following attributes support variables that are evaluated when a request is
made for the mounting Pod. These variables are useful for constructing requests
with SANs that contain values from the mounting Pod.

```
`csi.cert-manager.io/common-name`
`csi.cert-manager.io/dns-names`
`csi.cert-manager.io/uri-sans`
```

Variables follow the [go os.Expand](https://pkg.go.dev/os#Expand) structure,
which is generally what you would expect on a UNIX shell. The CSI driver has
access to the following variables:

```
${POD_NAME}
${POD_NAMESPACE}
${POD_UID}
```

#### Example Usage

```yaml
volumeAttributes:
  csi.cert-manager.io/issuer-name: ca-issuer
  csi.cert-manager.io/dns-names: "${POD_NAME}.${POD_NAMESPACE}.svc.cluster.local"
  csi.cert-manager.io/uri-sans: "spiffe://cluster.local/ns/${POD_NAMESPACE}/pod/${POD_NAME}/${POD_UID}"
  csi.cert-manager.io/common-name: "${POD_NAME}.${POD_NAMESPACE}"
```

## Design Documents
 - [Certificate Renewal](./docs/design/20190914.certificaterenewal.md)
