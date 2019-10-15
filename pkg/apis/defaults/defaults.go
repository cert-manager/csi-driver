package defaults

import (
	"time"

	"github.com/jetstack/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	"github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func SetDefaultAttributes(attr map[string]string) (map[string]string, error) {
	setDefaultIfEmpty(attr, v1alpha1.IssuerKindKey, cmapi.IssuerKind)
	setDefaultIfEmpty(attr, v1alpha1.IssuerGroupKey, certmanager.GroupName)

	setDefaultIfEmpty(attr, v1alpha1.IsCAKey, "false")
	setDefaultIfEmpty(attr, v1alpha1.DurationKey, cmapi.DefaultCertificateDuration.String())

	setDefaultIfEmpty(attr, v1alpha1.CertFileKey, "crt.pem")
	setDefaultIfEmpty(attr, v1alpha1.KeyFileKey, "key.pem")

	// TODO (@joshvanl): add a smarter defaulting mechanism
	dur, err := time.ParseDuration(attr[string(v1alpha1.DurationKey)])
	if err != nil {
		return nil, err
	}
	dur = dur / 3
	setDefaultIfEmpty(attr, v1alpha1.RenewBeforeKey, dur.String())

	// use given pod namespace if one not set
	setDefaultIfEmpty(attr, v1alpha1.NamespaceKey, attr[v1alpha1.CSIPodNamespaceKey])

	return attr, nil
}

func setDefaultIfEmpty(attr map[string]string, k, s string) {
	if len(attr[string(k)]) == 0 {
		attr[string(k)] = s
	}
}
