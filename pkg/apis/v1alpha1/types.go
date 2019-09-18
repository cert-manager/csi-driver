package v1alpha1

const (
	CSIPodNameKey      = "csi.storage.k8s.io/pod.name"
	CSIPodNamespaceKey = "csi.storage.k8s.io/pod.namespace"
	CSIEphemeralKey    = "csi.storage.k8s.io/ephemeral"
)

const (
	IssuerNameKey  string = "csi.certmanager.k8s.io/issuer-name"
	IssuerKindKey  string = "csi.certmanager.k8s.io/issuer-kind"
	IssuerGroupKey string = "csi.certmanager.k8s.io/issuer-group"

	CommonNameKey string = "csi.certmanager.k8s.io/common-name"
	DNSNamesKey   string = "csi.certmanager.k8s.io/dns-names"
	IPSANsKey     string = "csi.certmanager.k8s.io/ip-sans"
	URISANsKey    string = "csi.certmanager.k8s.io/uri-sans"
	DurationKey   string = "csi.certmanager.k8s.io/duration"
	IsCAKey       string = "csi.certmanager.k8s.io/is-ca"

	CertFileKey  string = "csi.certmanager.k8s.io/certificate-file"
	KeyFileKey   string = "csi.certmanager.k8s.io/privatekey-file"
	NamespaceKey string = "csi.certmanager.k8s.io/namespace"

	RenewBeforeKey      string = "csi.certmanager.k8s.io/renew-before"
	DisableAutoRenewKey string = "csi.certmanager.k8s.io/disable-auto-renew"
	ReusePrivateKey     string = "csi.certmanager.k8s.io/reuse-private-key"
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
