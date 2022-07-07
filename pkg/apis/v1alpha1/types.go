/*
Copyright 2021 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

const (
	IssuerNameKey  = "csi.cert-manager.io/issuer-name"
	IssuerKindKey  = "csi.cert-manager.io/issuer-kind"
	IssuerGroupKey = "csi.cert-manager.io/issuer-group"

	CommonNameKey  = "csi.cert-manager.io/common-name"
	DNSNamesKey    = "csi.cert-manager.io/dns-names"
	IPSANsKey      = "csi.cert-manager.io/ip-sans"
	URISANsKey     = "csi.cert-manager.io/uri-sans"
	DurationKey    = "csi.cert-manager.io/duration"
	IsCAKey        = "csi.cert-manager.io/is-ca"
	KeyUsagesKey   = "csi.cert-manager.io/key-usages"
	KeyEncodingKey = "csi.cert-manager.io/key-encoding"

	CAFileKey   = "csi.cert-manager.io/ca-file"
	CertFileKey = "csi.cert-manager.io/certificate-file"
	KeyFileKey  = "csi.cert-manager.io/privatekey-file"
	FSGroupKey  = "csi.cert-manager.io/fs-group"

	RenewBeforeKey  = "csi.cert-manager.io/renew-before"
	ReusePrivateKey = "csi.cert-manager.io/reuse-private-key"

	KeystoreTypeKey = "csi.cert-manager.io/keystore-type"
	KeystoreFileKey = "csi.cert-manager.io/keystore-file"
)

const (
	// Well-known attribute keys that are present in the volume context, passed
	// from the Kubelet during PublishVolume calls.
	K8sVolumeContextKeyPodName      = "csi.storage.k8s.io/pod.name"
	K8sVolumeContextKeyPodNamespace = "csi.storage.k8s.io/pod.namespace"
	K8sVolumeContextKeyPodUID       = "csi.storage.k8s.io/pod.uid"
)
