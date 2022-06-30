package pkcs12

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"fmt"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"software.sslmate.com/src/go-pkcs12"
)

// Create combines the inputs to a single pfx/p12 file
func Create(key crypto.PrivateKey, leaf []byte, chain []byte) ([]byte, error) {
	cert, err := pki.DecodeX509CertificateBytes(leaf)
	if err != nil {
		return nil, fmt.Errorf("pki.DecodeX509CertificateChainBytes(leaf): %v", err)
	}

	var cas []*x509.Certificate
	if len(chain) > 0 {
		cas, err = pki.DecodeX509CertificateChainBytes(chain)
		if err != nil {
			return nil, fmt.Errorf("pki.DecodeX509CertificateChainBytes(chain): %v", err)
		}
	}

	pfx, err := pkcs12.Encode(rand.Reader, key, cert, cas, pkcs12.DefaultPassword)
	if err != nil {
		return nil, fmt.Errorf("pkcs12.Encode: %v", err)
	}

	return pfx, nil
}
