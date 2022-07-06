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

func generateKeyAndCert(t *testing.T) (*rsa.PrivateKey, *x509.Certificate, []byte, []*x509.Certificate, []byte, *x509.Certificate, []byte, []*x509.Certificate) {
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

	leaf, err := x509.ParseCertificate(db)
	if err != nil {
		t.Fatalf("x509.ParseCertificate: %v", err)
	}

	idb, err := x509.CreateCertificate(rand.Reader, &intermediateTemplate, &intermediateTemplate, &pk.PublicKey, pk)
	if err != nil {
		t.Fatalf("x509.CreateCertificate(intermediate): %v", err)
	}

	ic, err := x509.ParseCertificate(idb)
	if err != nil {
		t.Errorf("x509.ParseCertificate(idb): %v", err)
	}

	rootDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &pk.PublicKey, pk)
	if err != nil {
		t.Fatalf("x509.CreateCertificate(ca): %v", err)
	}

	root, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatalf("x509.ParseCertificate: %v", err)
	}

	caChain := []*x509.Certificate{ic, root}
	chainDER := append(db, idb...)

	chain, err := x509.ParseCertificates(chainDER)
	if err != nil {
		t.Fatalf("x509.ParseCertificates(chainDER): %v", err)
	}

	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: db})
	rootPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})
	intermediatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: idb})

	chainPEM := append(leafPEM, intermediatePEM...)

	return pk, leaf, leafPEM, chain, chainPEM, root, rootPEM, caChain
}

func TestCreate(t *testing.T) {
	key, leaf, leafPEM, chain, chainPEM, root, rootPEM, caChain := generateKeyAndCert(t)
	tests := map[string]struct {
		key      *rsa.PrivateKey
		leaf     *x509.Certificate
		leafPEM  []byte
		chain    []*x509.Certificate
		chainPEM []byte
		root     *x509.Certificate
		rootPEM  []byte
		caChain  []*x509.Certificate
		expErr   bool
	}{
		"happy path": {
			key:      key,
			leaf:     leaf,
			leafPEM:  leafPEM,
			chain:    chain,
			chainPEM: chainPEM,
			root:     root,
			rootPEM:  rootPEM,
			caChain:  caChain,
			expErr:   false,
		},
		"without intermediate succeeds": {
			key:      key,
			leaf:     leaf,
			chainPEM: leafPEM,
			rootPEM:  rootPEM,
			caChain:  []*x509.Certificate{root},
		},
		"nil key": {
			key:    nil,
			expErr: true,
		},
		"empty chain": {
			key:     key,
			rootPEM: rootPEM,
			expErr:  true,
		},
		"empty root": {
			key:      key,
			chainPEM: chainPEM,
			expErr:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p12, err := Create(test.key, test.chainPEM, test.rootPEM)

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

			assert.Equal(t, test.key, pk.(*rsa.PrivateKey))
			assert.Equal(t, test.leaf, cert)
			assert.Equal(t, test.caChain, ca)
		})
	}
}
