package defaults

import (
	"time"

	"github.com/jetstack/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func SetDefaultAttributes(attr v1alpha1.Attributes) (v1alpha1.Attributes, error) {
	setDefaultIfEmpty(attr, v1alpha1.IssuerKindKey, cmapi.IssuerKind)
	setDefaultIfEmpty(attr, v1alpha1.IssuerGroupKey, certmanager.GroupName)

	setDefaultIfEmpty(attr, v1alpha1.IsCAKey, "false")
	setDefaultIfEmpty(attr, v1alpha1.DurationKey, cmapi.DefaultCertificateDuration.String())

	setDefaultIfEmpty(attr, v1alpha1.CertFileKey, "crt.pem")
	setDefaultIfEmpty(attr, v1alpha1.KeyFileKey, "key.pem")

	// TODO (@joshvanl): add a smarter defaulting mechanism
	dur, err := time.ParseDuration(attr[v1alpha1.DurationKey])
	if err != nil {
		return nil, err
	}
	dur = dur / 3
	setDefaultIfEmpty(attr, v1alpha1.RenewBeforeKey, dur.String())

	// use given pod namespace if one not set
	setDefaultIfEmpty(attr, v1alpha1.NamespaceKey, attr[v1alpha1.CSIPodNamespaceKey])

	return attr, nil
}

func setDefaultIfEmpty(attr v1alpha1.Attributes, k v1alpha1.Attribute, s string) {
	if len(attr[k]) == 0 {
		attr[k] = s
	}
}
