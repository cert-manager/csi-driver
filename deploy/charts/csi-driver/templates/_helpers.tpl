{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cert-manager-csi-driver.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cert-manager-csi-driver.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels

Labels are merged as a map, not concatenated as YAML text, so a key in
.Values.commonLabels cannot produce a duplicate key. On conflict
.Values.commonLabels wins.

Note that the PodMonitor selects Pods on app.kubernetes.io/name and
app.kubernetes.io/instance, both of which are set here. The pod template
therefore re-applies those two at a higher precedence, so that overriding
them cannot leave the PodMonitor matching nothing.

IMPORTANT: This function is standardized across all charts in the cert-manager GH organization.
Any changes to this function should also be made in cert-manager, trust-manager, approver-policy, ...
See https://github.com/cert-manager/cert-manager/issues/9348 for a list of linked PRs.
*/}}
{{- define "cert-manager-csi-driver.labels" -}}
{{- $labels := dict
  "app.kubernetes.io/name" (include "cert-manager-csi-driver.name" .)
  "helm.sh/chart" (include "cert-manager-csi-driver.chart" .)
  "app.kubernetes.io/instance" .Release.Name
  "app.kubernetes.io/managed-by" .Release.Service
-}}
{{- if .Chart.AppVersion }}
{{- $labels = set $labels "app.kubernetes.io/version" .Chart.AppVersion }}
{{- end }}
{{- toYaml (mergeOverwrite $labels (.Values.commonLabels | default dict)) }}
{{- end -}}

{{/*
Util function for generating the image URL based on the provided options.
IMPORTANT: This function is standardized across all charts in the cert-manager GH organization.
Any changes to this function should also be made in cert-manager, trust-manager, approver-policy, ...
See https://github.com/cert-manager/cert-manager/issues/6329 for a list of linked PRs.
*/}}
{{- define "cert-manager-csi-driver.image" -}}
{{- /*
Calling convention:

- (tuple <imageValues> <imageRegistry> <imageNamespace> <defaultReference>)

We intentionally pass imageRegistry/imageNamespace as explicit arguments rather than reading
from `.Values` inside this helper, because `helm-tool lint` does not reliably track `.Values.*`
usage through tuple/variable indirection.
*/ -}}

{{- if ne (len .) 4 -}}
    {{- fail (printf "ERROR: template \"cert-manager-csi-driver.image\" expects (tuple <imageValues> <imageRegistry> <imageNamespace> <defaultReference>), got %d arguments" (len .)) -}}
{{- end -}}

{{- $image := index . 0 -}}
{{- $imageRegistry := index . 1 | default "" -}}
{{- $imageNamespace := index . 2 | default "" -}}
{{- $defaultReference := index . 3 -}}

{{- $repository := "" -}}
{{- if $image.repository -}}
    {{- $repository = $image.repository -}}

    {{- /*
        Backwards compatibility: if image.registry is set, additionally prefix the repository with this registry.
    */ -}}
    {{- if $image.registry -}}
        {{- $repository = printf "%s/%s" $image.registry $repository -}}
    {{- end -}}
{{- else -}}
    {{- $name := required "ERROR: image.name must be set when image.repository is empty" $image.name -}}
    {{- $repository = $name -}}

    {{- if $imageNamespace -}}
        {{- $repository = printf "%s/%s" $imageNamespace $repository -}}
    {{- end -}}

    {{- if $imageRegistry -}}
        {{- $repository = printf "%s/%s" $imageRegistry $repository -}}
    {{- end -}}

    {{- /*
        Backwards compatibility: if image.registry is set, additionally prefix the repository with this registry.
    */ -}}
    {{- if $image.registry -}}
        {{- $repository = printf "%s/%s" $image.registry $repository -}}
    {{- end -}}
{{- end -}}

{{- $repository -}}
{{- if and $image.tag $image.digest -}}
    {{- printf ":%s@%s" $image.tag $image.digest -}}
{{- else if $image.tag -}}
    {{- printf ":%s" $image.tag -}}
{{- else if $image.digest -}}
    {{- printf "@%s" $image.digest -}}
{{- else -}}
    {{- printf "%s" $defaultReference -}}
{{- end -}}
{{- end -}}
