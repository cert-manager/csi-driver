package v1alpha1

type Attribute string

type Attributes map[Attribute]string

const (
	CSIPodNameKey      = "csi.storage.k8s.io/pod.name"
	CSIPodNamespaceKey = "csi.storage.k8s.io/pod.namespace"
	CSIEphemeralKey    = "csi.storage.k8s.io/ephemeral"
)

const (
	IssuerNameKey  Attribute = "csi.certmanager.k8s.io/issuer-name"
	IssuerKindKey  Attribute = "csi.certmanager.k8s.io/issuer-kind"
	IssuerGroupKey Attribute = "csi.certmanager.k8s.io/issuer-group"

	CommonNameKey Attribute = "csi.certmanager.k8s.io/common-name"
	DNSNamesKey   Attribute = "csi.certmanager.k8s.io/dns-names"
	IPSANsKey     Attribute = "csi.certmanager.k8s.io/ip-sans"
	URISANsKey    Attribute = "csi.certmanager.k8s.io/uri-sans"
	DurationKey   Attribute = "csi.certmanager.k8s.io/duration"
	IsCAKey       Attribute = "csi.certmanager.k8s.io/is-ca"

	CertFileKey  Attribute = "csi.certmanager.k8s.io/certificate-file"
	KeyFileKey   Attribute = "csi.certmanager.k8s.io/privatekey-file"
	NamespaceKey Attribute = "csi.certmanager.k8s.io/namespace"

	RenewBeforeKey      Attribute = "csi.certmanager.k8s.io/renew-before"
	DisableAutoRenewKey Attribute = "csi.certmanager.k8s.io/disable-auto-renew"
	ReusePrivateKey     Attribute = "csi.certmanager.k8s.io/reuse-private-key"
)
