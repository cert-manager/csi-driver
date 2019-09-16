package defaults

import (
	"github.com/jetstack/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func SetDefaults(attr v1alpha1.Attributes) v1alpha1.Attributes {
	setDefaultIfEmpty(attr, v1alpha1.IssuerKindKey, cmapi.IssuerKind)
	setDefaultIfEmpty(attr, v1alpha1.IssuerGroupKey, certmanager.GroupName)

	setDefaultIfEmpty(attr, v1alpha1.IsCAKey, "false")
	setDefaultIfEmpty(attr, v1alpha1.DurationKey, cmapi.DefaultCertificateDuration.String())

	setDefaultIfEmpty(attr, v1alpha1.CertFileKey, "crt.pem")
	setDefaultIfEmpty(attr, v1alpha1.KeyFileKey, "key.pem")

	// use given pod namespace if one not set
	setDefaultIfEmpty(attr, v1alpha1.NamespaceKey, attr[v1alpha1.CSIPodNamespaceKey])

	return attr
}

func setDefaultIfEmpty(attr v1alpha1.Attributes, k v1alpha1.Attribute, s string) {
	if len(attr[k]) == 0 {
		attr[k] = s
	}
}
