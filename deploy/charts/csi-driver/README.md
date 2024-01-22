# cert-manager-csi-driver

![Version: v0.6.1](https://img.shields.io/badge/Version-v0.6.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.6.1](https://img.shields.io/badge/AppVersion-v0.6.1-informational?style=flat-square)

cert-manager-csi-driver enables issuing secretless X.509 certificates for pods using cert-manager

**Homepage:** <https://github.com/cert-manager/csi-driver>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| cert-manager-maintainers | <cert-manager-maintainers@googlegroups.com> | <https://cert-manager.io> |

## Source Code

* <https://github.com/cert-manager/csi-driver>

## Values
<!-- AUTO-GENERATED -->


<table>
<tr>
<th>Property</th>
<th>Description</th>
<th>Type</th>
<th>Default</th>
</tr>
<tr>

<td>image.repository</td>
<td>

Target image repository.

</td>
<td>string</td>
<td>

```yaml
quay.io/jetstack/cert-manager-csi-driver
```

</td>
</tr>
<tr>

<td>image.tag</td>
<td>

Target image version tag.

</td>
<td>string</td>
<td>

```yaml
v0.0.0
```

</td>
</tr>
<tr>

<td>image.pullPolicy</td>
<td>

Kubernetes imagePullPolicy on csi-driver.

</td>
<td>string</td>
<td>

```yaml
IfNotPresent
```

</td>
</tr>
<tr>

<td>imagePullSecrets</td>
<td>

Optional secrets used for pulling the csi-driver container image  
  
For example:

```yaml
imagePullSecrets:
- name: secret-name
```

</td>
<td>array</td>
<td>

```yaml
[]
```

</td>
</tr>
<tr>

<td>commonLabels</td>
<td>

Labels to apply to all resources

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>nodeDriverRegistrarImage.repository</td>
<td>

Target image repository.

</td>
<td>string</td>
<td>

```yaml
registry.k8s.io/sig-storage/csi-node-driver-registrar
```

</td>
</tr>
<tr>

<td>nodeDriverRegistrarImage.tag</td>
<td>

Target image version tag.

</td>
<td>string</td>
<td>

```yaml
v2.10.0
```

</td>
</tr>
<tr>

<td>nodeDriverRegistrarImage.pullPolicy</td>
<td>

Kubernetes imagePullPolicy on node-driver.

</td>
<td>string</td>
<td>

```yaml
IfNotPresent
```

</td>
</tr>
<tr>

<td>livenessProbeImage.repository</td>
<td>

Target image repository.

</td>
<td>string</td>
<td>

```yaml
registry.k8s.io/sig-storage/livenessprobe
```

</td>
</tr>
<tr>

<td>livenessProbeImage.tag</td>
<td>

Target image version tag.

</td>
<td>string</td>
<td>

```yaml
v2.12.0
```

</td>
</tr>
<tr>

<td>livenessProbeImage.pullPolicy</td>
<td>

Kubernetes imagePullPolicy on liveness probe.

</td>
<td>string</td>
<td>

```yaml
IfNotPresent
```

</td>
</tr>
<tr>

<td>app.logLevel</td>
<td>

Verbosity of cert-manager-csi-driver logging.

</td>
<td>number</td>
<td>

```yaml
1
```

</td>
</tr>
<tr>

<td>app.driver.name</td>
<td>

Name of the driver which will be registered with Kubernetes.

</td>
<td>string</td>
<td>

```yaml
csi.cert-manager.io
```

</td>
</tr>
<tr>

<td>app.driver.useTokenRequest</td>
<td>

If enabled, will use CSI token request for creating. CertificateRequests. CertificateRequests will be created via mounting pod's service accounts.

</td>
<td>bool</td>
<td>

```yaml
false
```

</td>
</tr>
<tr>

<td>app.driver.csiDataDir</td>
<td>

Configures the hostPath directory that the driver will write and mount volumes from.

</td>
<td>string</td>
<td>

```yaml
/tmp/cert-manager-csi-driver
```

</td>
</tr>
<tr>

<td>app.livenessProbe.port</td>
<td>

The port that will expose the livness of the csi-driver

</td>
<td>number</td>
<td>

```yaml
9809
```

</td>
</tr>
<tr>

<td>app.kubeletRootDir</td>
<td>

Overrides path to root kubelet directory in case of a non-standard k8s install.

</td>
<td>string</td>
<td>

```yaml
/var/lib/kubelet
```

</td>
</tr>
<tr>

<td>daemonSetAnnotations</td>
<td>

Optional additional annotations to add to the csi-driver DaemonSet

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>podAnnotations</td>
<td>

Optional additional annotations to add to the csi-driver Pods

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>podLabels</td>
<td>

Optional additional labels to add to the csi-driver Pods

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>resources</td>
<td>

Kubernetes pod resources requests/limits for cert-manager-csi-driver  
  
For example:

```yaml
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>nodeSelector</td>
<td>

Kubernetes node selector: node labels for pod assignment. For example, to allow scheduling of DaemonSet on linux nodes only:

```yaml
nodeSelector:
  kubernetes.io/os: linux
```

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>affinity</td>
<td>

Kubernetes affinity: constraints for pod assignment  
  
For example:

```yaml
affinity:
  nodeAffinity:
   requiredDuringSchedulingIgnoredDuringExecution:
     nodeSelectorTerms:
     - matchExpressions:
       - key: foo.bar.com/role
         operator: In
         values:
         - master
```

</td>
<td>object</td>
<td>

```yaml
{}
```

</td>
</tr>
<tr>

<td>tolerations</td>
<td>

Kubernetes pod tolerations for cert-manager-csi-driver  
  
For example:

```yaml
tolerations:
- operator: "Exists"
```

</td>
<td>array</td>
<td>

```yaml
[]
```

</td>
</tr>
<tr>

<td>priorityClassName</td>
<td>

Optional priority class to be used for the csi-driver pods.

</td>
<td>string</td>
<td>

```yaml
""
```

</td>
</tr>
</table>

<!-- /AUTO-GENERATED -->