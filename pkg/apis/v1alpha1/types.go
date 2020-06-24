/*
Copyright 2019 The Jetstack cert-manager contributors.

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
	MetaDataFileName = "metadata.json"
)

const (
	CSIPodNameKey      = "csi.storage.k8s.io/pod.name"
	CSIPodNamespaceKey = "csi.storage.k8s.io/pod.namespace"
	CSIPodUIDKey       = "csi.storage.k8s.io/pod.uid"
	CSIEphemeralKey    = "csi.storage.k8s.io/ephemeral"
)

const (
	IssuerNameKey  string = "csi.cert-manager.io/issuer-name"
	IssuerKindKey  string = "csi.cert-manager.io/issuer-kind"
	IssuerGroupKey string = "csi.cert-manager.io/issuer-group"

	CommonNameKey string = "csi.cert-manager.io/common-name"
	DNSNamesKey   string = "csi.cert-manager.io/dns-names"
	IPSANsKey     string = "csi.cert-manager.io/ip-sans"
	URISANsKey    string = "csi.cert-manager.io/uri-sans"
	DurationKey   string = "csi.cert-manager.io/duration"
	IsCAKey       string = "csi.cert-manager.io/is-ca"
	KeyUsagesKey  string = "csi.cert-manager.io/key-usages"

	CAFileKey   string = "csi.cert-manager.io/ca-file"
	CertFileKey string = "csi.cert-manager.io/certificate-file"
	KeyFileKey  string = "csi.cert-manager.io/privatekey-file"

	RenewBeforeKey      string = "csi.cert-manager.io/renew-before"
	DisableAutoRenewKey string = "csi.cert-manager.io/disable-auto-renew"
	ReusePrivateKey     string = "csi.cert-manager.io/reuse-private-key"
)

type MetaData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`

	// real file path in the host file system
	Path string `json:"path"`
	// target path to mount to
	TargetPath string `json:"targetPath"`

	Attributes map[string]string `json:"attributes"`
}
