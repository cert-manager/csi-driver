# cert-manager-csi

![Version: v0.1.0](https://img.shields.io/badge/Version-v0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.1.0](https://img.shields.io/badge/AppVersion-v0.1.0-informational?style=flat-square)

A Helm chart for cert-manager-csi

**Homepage:** <https://github.com/cert-manager/csi>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| joshvanl | joshua.vanleeuwen@jetstack.io | https://cert-manager.io |

## Source Code

* <https://github.com/cert-manager/csi>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| app.driver | object | `{"name":"csi.cert-manager.io","useRequestToken":false}` | Options for CSI driver |
| app.driver.name | string | `"csi.cert-manager.io"` | Name of the driver which will be registered with Kubernetes. |
| app.driver.useRequestToken | bool | `false` | Use the empty audience request token for creating CertificateRequests. Created CertificateRequests will have UserInfo fields of the mounting Pod's ServiceAccount. |
| app.logLevel | int | `1` | Verbosity of cert-manager-csi logging. |
| image.pullPolicy | string | `"IfNotPresent"` | Kubernetes imagePullPolicy on DaemonSet. |
| image.repository | string | `"quay.io/jetstack/cert-manager-csi"` | Target image repository. |
| image.tag | string | `"v0.1.0"` | Target image version tag. |
| resources | object | `{}` |  |

