package v1alpha1

import (
	"time"
)

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

	CertFileKey string = "csi.cert-manager.io/certificate-file"
	KeyFileKey  string = "csi.cert-manager.io/privatekey-file"

	RenewBeforeKey      string = "csi.cert-manager.io/renew-before"
	DisableAutoRenewKey string = "csi.cert-manager.io/disable-auto-renew"
	ReusePrivateKey     string = "csi.cert-manager.io/reuse-private-key"
)

type DriverID struct {
	NodeID     string `json:"nodeID"`
	DriverName string `json:"driverName"`
}

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

type WebhookClient interface {
	Register(*DriverID) error
	Create(*MetaData)
	Renew(*MetaData)
	Destroy(*MetaData)
}

type WebhookClientPost struct {
	Timestamp time.Time `json:"timestamp"`
	DriverID  *DriverID `json:"driverID"`
	*MetaData
}
