package defaults

import (
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func SetDefaults(attr map[v1alpha1.Attribute]string) map[v1alpha1.Attribute]string {
	setDefaultIfEmpty(attr, v1alpha1.IssuerKindKey, "Issuer")
	setDefaultIfEmpty(attr, v1alpha1.IssuerGroupKey, "certmanager.k8s.io")

	setDefaultIfEmpty(attr, v1alpha1.IsCAKey, "false")
	setDefaultIfEmpty(attr, v1alpha1.DurationKey, cmapi.DefaultCertificateDuration.String())

	setDefaultIfEmpty(attr, v1alpha1.CertFileKey, "crt.pem")
	setDefaultIfEmpty(attr, v1alpha1.KeyFileKey, "key.pem")

	// use given pod namespace if one not set
	setDefaultIfEmpty(attr, v1alpha1.NamespaceKey, attr[v1alpha1.CSIPodNamespaceKey])

	return attr
}

func setDefaultIfEmpty(attr map[v1alpha1.Attribute]string, k v1alpha1.Attribute, s string) {
	if len(attr[k]) == 0 {
		attr[k] = s
	}
}
