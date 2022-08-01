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

package unit

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/stretchr/testify/assert"
)

// CertBundle contains a signed certificate and corresponding private key, to
// be used for testing.
type CertBundle struct {
	Cert *x509.Certificate
	PEM  []byte
	PK   crypto.PrivateKey
}

// MustCreateBundle creating a CertBundle for testing. If issuer bundle is
// empty, certificate will be self signed.
func MustCreateBundle(t *testing.T, issuer *CertBundle, name string) *CertBundle {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	assert.NoError(t, err)

	template := &x509.Certificate{
		Version:               3,
		BasicConstraintsValid: true,
		SerialNumber:          serialNumber,
		PublicKeyAlgorithm:    x509.ECDSA,
		PublicKey:             pk.Public(),
		IsCA:                  true,
		Subject: pkix.Name{
			CommonName: name,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Minute),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	var (
		issuerKey  crypto.PrivateKey
		issuerCert *x509.Certificate
	)

	if issuer == nil {
		// No issuer implies the cert should be self signed.
		issuerKey = pk
		issuerCert = template
	} else {
		issuerKey = issuer.PK
		issuerCert = issuer.Cert
	}

	certPEM, cert, err := pki.SignCertificate(template, issuerCert, pk.Public(), issuerKey)
	assert.NoError(t, err)

	return &CertBundle{PEM: certPEM, Cert: cert, PK: pk}
}
