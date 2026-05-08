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
	"bytes"
	"crypto"
	"crypto/x509"
	"testing"

	"github.com/cert-manager/csi-lib/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	"github.com/cert-manager/csi-driver/test/unit"
)

func Test_Handle(t *testing.T) {
	root := unit.MustCreateBundle(t, nil, "root")

	tests := map[string]struct {
		meta       metadata.Metadata
		attributes map[string]string
		pk         crypto.PrivateKey
		chainPEM   []byte
		expFiles   []string
		expErr     bool
	}{
		"if no PKCS12 attributes provided, expect no files written": {
			meta:       metadata.Metadata{},
			attributes: map[string]string{},
			pk:         root.PK,
			chainPEM:   root.PEM,
			expFiles:   []string{},
			expErr:     false,
		},
		"if PKCS12 enabled with password in attribute, expect file written": {
			meta: metadata.Metadata{},
			attributes: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-password": "my-password",
				"csi.cert-manager.io/pkcs12-filename": "crt.p12",
			},
			pk:       root.PK,
			chainPEM: root.PEM,
			expFiles: []string{"crt.p12"},
			expErr:   false,
		},
		"if PKCS12 enabled with password in secret, expect file written": {
			meta: metadata.Metadata{
				Secrets: map[string]string{
					csiapi.KeyStorePKCS12PasswordSecretKey: "secret-password",
				},
			},
			attributes: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-filename": "crt.p12",
			},
			pk:       root.PK,
			chainPEM: root.PEM,
			expFiles: []string{"crt.p12"},
			expErr:   false,
		},
		"if PKCS12 enabled with password in both, secret takes precedence": {
			meta: metadata.Metadata{
				Secrets: map[string]string{
					csiapi.KeyStorePKCS12PasswordSecretKey: "secret-password",
				},
			},
			attributes: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-password": "attr-password",
				"csi.cert-manager.io/pkcs12-filename": "crt.p12",
			},
			pk:       root.PK,
			chainPEM: root.PEM,
			expFiles: []string{"crt.p12"},
			expErr:   false,
		},
		"if PKCS12 enabled with no password in attribute or secret, expect error": {
			meta: metadata.Metadata{},
			attributes: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-filename": "crt.p12",
			},
			pk:       root.PK,
			chainPEM: root.PEM,
			expFiles: []string{},
			expErr:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			files := make(map[string][]byte)
			err := Handle(test.meta, test.attributes, files, test.pk, test.chainPEM)
			assert.Equal(t, test.expErr, err != nil, "unexpected error: %v", err)

			var gotFiles []string
			for k := range files {
				gotFiles = append(gotFiles, k)
			}
			assert.ElementsMatch(t, test.expFiles, gotFiles)
		})
	}
}

func Test_create(t *testing.T) {
	root := unit.MustCreateBundle(t, nil, "root")
	int1 := unit.MustCreateBundle(t, root, "int1")
	int2 := unit.MustCreateBundle(t, int1, "int2")

	tests := map[string]struct {
		pk       crypto.PrivateKey
		chainPEM []byte
		expPK    crypto.PrivateKey
		expCert  *x509.Certificate
		expCAs   []*x509.Certificate
		expErr   bool
	}{
		"if chain is empty, then expect error": {
			pk:       int2.PK,
			chainPEM: []byte{},
			expPK:    nil,
			expCert:  nil,
			expCAs:   nil,
			expErr:   true,
		},
		"if chain contains single certificate, expect it is encoded": {
			pk:       int2.PK,
			chainPEM: int2.PEM,
			expPK:    int2.PK,
			expCert:  int2.Cert,
			expCAs:   nil,
			expErr:   false,
		},
		"if chain contains multiple certificates, expect it is encoded and splits cas from leaf": {
			pk:       int2.PK,
			chainPEM: bytes.Join([][]byte{int2.PEM, int1.PEM, root.PEM}, []byte("\n")),
			expPK:    int2.PK,
			expCert:  int2.Cert,
			expCAs:   []*x509.Certificate{int1.Cert, root.Cert},
			expErr:   false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := create("test-password", test.pk, test.chainPEM)
			require.Equal(t, test.expErr, err != nil, "%v", err)

			if !test.expErr {
				pk, cert, cas, err := pkcs12.DecodeChain(resp, "test-password")
				require.NoError(t, err)

				assert.Equal(t, test.expPK, pk)
				assert.Equal(t, test.expCert, cert)
				assert.Equal(t, test.expCAs, cas)
			}
		})
	}
}
