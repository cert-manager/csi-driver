# cert-manager csi-driver

<!-- see https://artifacthub.io/packages/helm/cert-manager/cert-manager-csi-driver for the rendered version -->

## Helm Values

<!-- AUTO-GENERATED -->

#### **image.repository** ~ `string`
> Default value:
> ```yaml
> quay.io/jetstack/cert-manager-csi-driver
> ```

Target image repository.
#### **image.registry** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

Target image registry. This value is prepended to the target image repository, if set.
#### **image.tag** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

Target image version tag. Defaults to the chart's appVersion.
#### **image.digest** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

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
#### **nodeDriverRegistrarImage.repository** ~ `string`
> Default value:
> ```yaml
> registry.k8s.io/sig-storage/csi-node-driver-registrar
> ```

Target image repository.
#### **nodeDriverRegistrarImage.registry** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

Target image registry. This value is prepended to the target image repository, if set.
#### **nodeDriverRegistrarImage.tag** ~ `string`
> Default value:
> ```yaml
> v2.10.0
> ```

Target image version tag. Defaults to the chart's appVersion.
#### **nodeDriverRegistrarImage.digest** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

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
#### **livenessProbeImage.repository** ~ `string`
> Default value:
> ```yaml
> registry.k8s.io/sig-storage/livenessprobe
> ```

Target image repository.
#### **livenessProbeImage.registry** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

Target image registry. This value is prepended to the target image repository, if set.
#### **livenessProbeImage.tag** ~ `string`
> Default value:
> ```yaml
> v2.12.0
> ```

Target image version tag. Defaults to the chart's appVersion.
#### **livenessProbeImage.digest** ~ `unknown`
> Default value:
> ```yaml
> null
> ```

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
> {}
> ```

Kubernetes node selector: node labels for pod assignment. For example, use this to allow scheduling of DaemonSet on linux nodes only:

```yaml
nodeSelector:
  kubernetes.io/os: linux
```
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

<!-- /AUTO-GENERATED -->