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

package defaults

import (
	"strings"

	"github.com/cert-manager/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

// SetDefaultAttributes will set default values on the given attribute map.
// It will not modify the attributes in-place, and instead will return a copy.
func SetDefaultAttributes(attrOriginal map[string]string) (map[string]string, error) {
	attr := make(map[string]string)
	for k, v := range attrOriginal {
		attr[k] = v
	}

	setDefaultIfEmpty(attr, csiapi.IssuerKindKey, cmapi.IssuerKind)
	setDefaultIfEmpty(attr, csiapi.IssuerGroupKey, certmanager.GroupName)

	setDefaultIfEmpty(attr, csiapi.IsCAKey, "false")
	setDefaultIfEmpty(attr, csiapi.DurationKey, cmapi.DefaultCertificateDuration.String())

	setDefaultIfEmpty(attr, csiapi.CAFileKey, "ca.crt")
	setDefaultIfEmpty(attr, csiapi.CertFileKey, "tls.crt")
	setDefaultIfEmpty(attr, csiapi.KeyFileKey, "tls.key")

	setDefaultIfEmpty(attr, csiapi.KeyEncodingKey, "PKCS1")

	setDefaultIfEmpty(attr, csiapi.KeyUsagesKey, strings.Join([]string{string(cmapi.UsageDigitalSignature), string(cmapi.UsageKeyEncipherment)}, ","))

	setDefaultKeyStorePKCS12(attr)

	return attr, nil
}

func setDefaultIfEmpty(attr map[string]string, k, v string) {
	if len(attr[string(k)]) == 0 {
		attr[string(k)] = v
	}
}

// setDefaultKeystorePKCS12 sets the default values for the PKCS12 relevant
// attributes. If the csiapi.KeyStorePKCS12EnableKey key is not defined, omit
// setting defaults on the other PKCS12 keys, since they should not be present
// in the attributes at all. If the other attributes are present, then a
// validation error will be picked up by validation.
func setDefaultKeyStorePKCS12(attr map[string]string) {
	if _, ok := attr[csiapi.KeyStorePKCS12EnableKey]; ok {
		setDefaultIfEmpty(attr, csiapi.KeyStorePKCS12FileKey, "keystore.p12")
	}
}
