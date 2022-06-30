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

package pkcs12

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"software.sslmate.com/src/go-pkcs12"
)

var (
	notBefore = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	notAfter  = time.Date(1970, time.January, 4, 0, 0, 0, 0, time.UTC)
)

func generateKeyAndCert(t *testing.T) (*rsa.PrivateKey, []byte, []byte) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), 128),
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	intermediateTemplate := x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), 128),
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	caTemplate := x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), 128),
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	db, err := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	if err != nil {
		t.Fatalf("x509.CreateCertificate: %v", err)
	}

	//cert, err := x509.ParseCertificate(db)
	//if err != nil {
	//	t.Fatalf("x509.ParseCertificate: %v", err)
	//}

	idb, err := x509.CreateCertificate(rand.Reader, &intermediateTemplate, &intermediateTemplate, &pk.PublicKey, pk)
	if err != nil {
		t.Fatalf("x509.CreateCertificate(intermediate): %v", err)
	}

	rdb, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &pk.PublicKey, pk)
	if err != nil {
		t.Fatalf("x509.CreateCertificate(ca): %v", err)
	}

	// TODO: can we use the concatenated PEM chain?
	_ = append(idb, rdb...)

	//var blocks []*pem.Block
	//for {
	//	b, rest := pem.Decode(cdb)
	//
	//	blocks = append(blocks, b)
	//
	//	if len(rest) == 0 {
	//		break
	//	}
	//}

	//chain, err := x509.ParseCertificates(cb)
	//if err != nil {
	//	t.Fatalf("x509.ParseCertificates: %v", err)
	//}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: db})
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rdb})

	//return pk, db, cb
	return pk, certPEM, caPEM
}

func TestCreate(t *testing.T) {
	key, leaf, chain := generateKeyAndCert(t)
	tests := map[string]struct {
		key    *rsa.PrivateKey
		leaf   []byte
		chain  []byte
		expErr bool
	}{
		"happy path": {
			key:    key,
			leaf:   leaf,
			chain:  chain,
			expErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p12, err := Create(test.key, test.leaf, test.chain)

			if test.expErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			pk, cert, ca, err := pkcs12.DecodeChain(p12, pkcs12.DefaultPassword)
			assert.NoError(t, err)
			assert.NotNil(t, pk)
			assert.NotNil(t, cert)
			assert.NotNil(t, ca)

			//assert.Equal(t, test.key, pk.(*rsa.PrivateKey))
			//assert.Equal(t, test.cert, cert)
			//assert.Equal(t, []*x509.Certificate{test.testBundle.ca}, ca)
		})
	}
}
