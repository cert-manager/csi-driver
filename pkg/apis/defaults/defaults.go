package defaults

import (
	"time"

	"github.com/jetstack/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func SetDefaultAttributes(attr map[string]string) (map[string]string, error) {
	setDefaultIfEmpty(attr, csiapi.IssuerKindKey, cmapi.IssuerKind)
	setDefaultIfEmpty(attr, csiapi.IssuerGroupKey, certmanager.GroupName)

	setDefaultIfEmpty(attr, csiapi.IsCAKey, "false")
	setDefaultIfEmpty(attr, csiapi.DurationKey, cmapi.DefaultCertificateDuration.String())

	setDefaultIfEmpty(attr, csiapi.CAFileKey, "ca.pem")
	setDefaultIfEmpty(attr, csiapi.CertFileKey, "crt.pem")
	setDefaultIfEmpty(attr, csiapi.KeyFileKey, "key.pem")

	// TODO (@joshvanl): add a smarter defaulting mechanism
	dur, err := time.ParseDuration(attr[string(csiapi.DurationKey)])
	if err != nil {
		return nil, err
	}
	dur = dur / 3
	setDefaultIfEmpty(attr, csiapi.RenewBeforeKey, dur.String())

	return attr, nil
}

func setDefaultIfEmpty(attr map[string]string, k, v string) {
	if len(attr[string(k)]) == 0 {
		attr[string(k)] = v
	}
}
