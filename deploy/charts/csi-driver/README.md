# cert-manager csi-driver

<!-- see https://artifacthub.io/packages/helm/cert-manager/cert-manager-csi-driver for the rendered version -->

## Helm Values

<!-- AUTO-GENERATED -->

#### **metrics.enabled** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Enable the metrics server on csi-driver pods.  
If false, the metrics server will be disabled and the other metrics fields below will be ignored.
#### **metrics.port** ~ `number`
> Default value:
> ```yaml
> 9402
> ```

The TCP port on which the metrics server will listen.
#### **metrics.podmonitor.enabled** ~ `bool`
> Default value:
> ```yaml
> false
> ```

Create a PodMonitor to add csi-driver to Prometheus if you are using Prometheus Operator. See https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PodMonitor
#### **metrics.podmonitor.namespace** ~ `string`

The namespace that the pod monitor should live in, defaults to the cert-manager-csi-driver namespace.

#### **metrics.podmonitor.prometheusInstance** ~ `string`
> Default value:
> ```yaml
> default
> ```

Specifies the `prometheus` label on the created PodMonitor. This is used when different Prometheus instances have label selectors matching different PodMonitors.
#### **metrics.podmonitor.interval** ~ `string`
> Default value:
> ```yaml
> 60s
> ```

The interval to scrape metrics.
#### **metrics.podmonitor.scrapeTimeout** ~ `string`
> Default value:
> ```yaml
> 30s
> ```

The timeout before a metrics scrape fails.
#### **metrics.podmonitor.labels** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Additional labels to add to the PodMonitor.
#### **metrics.podmonitor.annotations** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Additional annotations to add to the PodMonitor.
#### **metrics.podmonitor.honorLabels** ~ `bool`
> Default value:
> ```yaml
> false
> ```

Keep labels from scraped data, overriding server-side labels.
#### **metrics.podmonitor.endpointAdditionalProperties** ~ `object`
> Default value:
> ```yaml
> {}
> ```

EndpointAdditionalProperties allows setting additional properties on the endpoint such as relabelings, metricRelabelings etc.  
  
For example:

```yaml
endpointAdditionalProperties:
 relabelings:
 - action: replace
   sourceLabels:
   - __meta_kubernetes_pod_node_name
   targetLabel: instance
```



#### **imageRegistry** ~ `string`
> Default value:
> ```yaml
> quay.io
> ```

The container registry used for csi-driver images by default. This can include path prefixes (e.g. "artifactory.example.com/docker").

#### **imageNamespace** ~ `string`
> Default value:
> ```yaml
> jetstack
> ```

The repository namespace used for csi-driver images by default.  
Examples:  
- jetstack  
- cert-manager

#### **image.registry** ~ `string`

Target image registry. This value is prepended to the target image repository, if set.  
For example:

```yaml
registry: quay.io
repository: jetstack/cert-manager-csi-driver
```

Deprecated: per-component registry prefix.  
  
If set, this value is *prepended* to the image repository that the chart would otherwise render. This applies both when `image.repository` is set and when the repository is computed from  
`imageRegistry` + `imageNamespace` + `image.name`.  
  
This can produce "double registry" style references such as  
`legacy.example.io/quay.io/jetstack/...`. Prefer using the global  
`imageRegistry`/`imageNamespace` values.

#### **image.repository** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Full repository override (takes precedence over `imageRegistry`, `imageNamespace`, and `image.name`).  
Example: quay.io/jetstack/cert-manager-csi-driver

#### **image.name** ~ `string`
> Default value:
> ```yaml
> cert-manager-csi-driver
> ```

The image name for the csi-driver.  
This is used (together with `imageRegistry` and `imageNamespace`) to construct the full image reference.

#### **image.tag** ~ `string`

Override the image tag to deploy by setting this variable. If no value is set, the chart's appVersion is used.

#### **image.digest** ~ `string`

Target image digest. Override any tag, if set.  
For example:

```yaml
digest: sha256:0e072dddd1f7f8fc8909a2ca6f65e76c5f0d2fcfb8be47935ae3457e8bbceb20
```

#### **image.pullPolicy** ~ `string`
> Default value:
> ```yaml
> IfNotPresent
> ```

Kubernetes imagePullPolicy on Deployment.
#### **imagePullSecrets** ~ `array`
> Default value:
> ```yaml
> []
> ```

Optional secrets used for pulling the csi-driver container image.  
  
For example:

```yaml
imagePullSecrets:
- name: secret-name
```
#### **commonLabels** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Labels to apply to all resources.
#### **nodeDriverRegistrarImage.registry** ~ `string`

Target image registry. This value is prepended to the target image repository, if set.  
For example:

```yaml
registry: registry.k8s.io
repository: sig-storage/csi-node-driver-registrar
```

Deprecated: per-component registry prefix.  
  
If set, this value is *prepended* to the image repository that the chart would otherwise render. This applies both when `image.repository` is set and when the repository is computed from  
`imageRegistry` + `imageNamespace` + `image.name`.  
  
This can produce "double registry" style references such as  
`legacy.example.io/quay.io/jetstack/...`. Prefer using the global  
`imageRegistry`/`imageNamespace` values.

#### **nodeDriverRegistrarImage.repository** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Full repository override (takes precedence over `imageRegistry`, `imageNamespace`, and `image.name`).  
Example: quay.io/jetstack/cert-manager-csi-driver

#### **nodeDriverRegistrarImage.name** ~ `string`
> Default value:
> ```yaml
> csi-node-driver-registrar
> ```

The image name for the node-driver-registrar.  
This is used to construct the full image reference if `repository` is empty.

#### **nodeDriverRegistrarImage.tag** ~ `string`

Override the image tag to deploy by setting this variable. If no value is set, the chart's appVersion is used.

#### **nodeDriverRegistrarImage.digest** ~ `string`

Target image digest. Override any tag, if set.  
For example:

```yaml
digest: sha256:0e072dddd1f7f8fc8909a2ca6f65e76c5f0d2fcfb8be47935ae3457e8bbceb20
```

#### **nodeDriverRegistrarImage.pullPolicy** ~ `string`
> Default value:
> ```yaml
> IfNotPresent
> ```

Kubernetes imagePullPolicy on Deployment.
#### **livenessProbeImage.registry** ~ `string`

Target image registry. This value is prepended to the target image repository, if set.  
For example:

```yaml
registry: registry.k8s.io
repository: sig-storage/livenessprobe
```

Deprecated: per-component registry prefix.  
  
If set, this value is *prepended* to the image repository that the chart would otherwise render. This applies both when `image.repository` is set and when the repository is computed from  
`imageRegistry` + `imageNamespace` + `image.name`.  
  
This can produce "double registry" style references such as  
`legacy.example.io/quay.io/jetstack/...`. Prefer using the global  
`imageRegistry`/`imageNamespace` values.

#### **livenessProbeImage.repository** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Full repository override (takes precedence over `imageRegistry`, `imageNamespace`, and `image.name`).  
Example: quay.io/jetstack/cert-manager-csi-driver

#### **livenessProbeImage.name** ~ `string`
> Default value:
> ```yaml
> livenessprobe
> ```

The image name for the liveness probe.  
This is used (together with `imageRegistry` and `imageNamespace`) to construct the full image reference.

#### **livenessProbeImage.tag** ~ `string`

Override the image tag to deploy by setting this variable. If no value is set, the chart's appVersion is used.

#### **livenessProbeImage.digest** ~ `string`

Target image digest. Override any tag, if set.  
For example:

```yaml
digest: sha256:0e072dddd1f7f8fc8909a2ca6f65e76c5f0d2fcfb8be47935ae3457e8bbceb20
```

#### **livenessProbeImage.pullPolicy** ~ `string`
> Default value:
> ```yaml
> IfNotPresent
> ```

Kubernetes imagePullPolicy on Deployment.
#### **app.logLevel** ~ `number`
> Default value:
> ```yaml
> 1
> ```

Verbosity of cert-manager-csi-driver logging.
#### **app.driver.name** ~ `string`
> Default value:
> ```yaml
> csi.cert-manager.io
> ```

Name of the driver to be registered with Kubernetes.
#### **app.driver.useTokenRequest** ~ `bool`
> Default value:
> ```yaml
> false
> ```

If enabled, this uses a CSI token request for creating. CertificateRequests. CertificateRequests are created by mounting the pod's service accounts.
#### **app.driver.csiDataDir** ~ `string`
> Default value:
> ```yaml
> /tmp/cert-manager-csi-driver
> ```

Configures the hostPath directory that the driver writes and mounts volumes from.
#### **app.livenessProbe.port** ~ `number`
> Default value:
> ```yaml
> 9809
> ```

The port that will expose the liveness of the csi-driver.
#### **app.kubeletRootDir** ~ `string`
> Default value:
> ```yaml
> /var/lib/kubelet
> ```

Overrides the path to root kubelet directory in case of a non-standard Kubernetes install.
#### **daemonSetAnnotations** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Optional additional annotations to add to the csi-driver DaemonSet.
#### **podAnnotations** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Optional additional annotations to add to the csi-driver pods.
#### **podLabels** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Optional additional labels to add to the csi-driver pods.
#### **resources** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Kubernetes pod resources requests/limits for cert-manager-csi-driver.  
  
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
#### **nodeSelector** ~ `object`
> Default value:
> ```yaml
> kubernetes.io/os: linux
> ```

Kubernetes node selector: node labels for pod assignment.

#### **affinity** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Kubernetes affinity: constraints for pod assignment.  
  
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
#### **tolerations** ~ `array`
> Default value:
> ```yaml
> []
> ```

Kubernetes pod tolerations for cert-manager-csi-driver.  
  
For example:

```yaml
tolerations:
- operator: "Exists"
```
#### **priorityClassName** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Optional priority class to be used for the csi-driver pods.
#### **openshift.securityContextConstraint.enabled** ~ `boolean,string,null`
> Default value:
> ```yaml
> detect
> ```

Include RBAC to allow the DaemonSet to "use" the specified  
SecurityContextConstraints.  
  
This value can either be a boolean true or false, or the string "detect". If set to "detect" then the securityContextConstraint is automatically enabled for openshift installs.

#### **openshift.securityContextConstraint.name** ~ `string`
> Default value:
> ```yaml
> privileged
> ```

Name of the SecurityContextConstraints to create RBAC for.

<!-- /AUTO-GENERATED -->