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

package filestore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"
	"github.com/stretchr/testify/assert"
)

var (
	notBefore = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	notAfter  = time.Date(1970, time.January, 4, 0, 0, 0, 0, time.UTC)
)

type testBundle struct {
	caPEM   []byte
	certPEM []byte
	pk      *rsa.PrivateKey
	pkPEM   []byte
}

type keyEncoder func(key *rsa.PrivateKey) (*pem.Block, error)

var (
	pkcs1Encoder keyEncoder = func(key *rsa.PrivateKey) (*pem.Block, error) {
		pkBytes := x509.MarshalPKCS1PrivateKey(key)
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkBytes}, nil
	}
	pkcs8Encoder keyEncoder = func(key *rsa.PrivateKey) (*pem.Block, error) {
		pkBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "PRIVATE KEY", Bytes: pkBytes}, nil
	}
)

func newTestBundle(t *testing.T, encoder keyEncoder) testBundle {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	pemBlock, err := encoder(pk)
	if err != nil {
		t.Fatal(err)
	}
	pkPEM := pem.EncodeToMemory(pemBlock)

	template := x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), 128),
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	rootTemplate := &x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), 128),
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	rootDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &pk.PublicKey, pk)
	if err != nil {
		t.Fatal(err)
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})

	return testBundle{caPEM, certPEM, pk, pkPEM}
}

func Test_calculateNextIssuanceTime(t *testing.T) {
	testBundle := newTestBundle(t, pkcs1Encoder)

	tests := map[string]struct {
		attrs   map[string]string
		expTime time.Time
		expErr  bool
	}{
		"if no attributes given, return 2/3rd certificate lifetime": {
			attrs:   map[string]string{},
			expTime: notBefore.AddDate(0, 0, 2),
			expErr:  false,
		},
		"if renew before present, return renew before time": {
			attrs: map[string]string{
				"csi.cert-manager.io/renew-before": "48h",
			},
			expTime: notBefore.AddDate(0, 0, 1),
			expErr:  false,
		},
		"if renew before present but is before NotBefore, return 2/3rds": {
			attrs: map[string]string{
				"csi.cert-manager.io/renew-before": "100h",
			},
			expTime: notBefore.AddDate(0, 0, 2),
			expErr:  false,
		},
		"if renew before present but given a bad string, return error": {
			attrs: map[string]string{
				"csi.cert-manager.io/renew-before": "bad-duration",
			},
			expTime: time.Time{},
			expErr:  true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			renewTime, err := calculateNextIssuanceTime(test.attrs, testBundle.certPEM)
			assert.Equal(t, test.expErr, err != nil)
			assert.Equal(t, test.expTime, renewTime)
		})
	}
}

func Test_WriteKeypair(t *testing.T) {
	pkcs1Bundle := newTestBundle(t, pkcs1Encoder)
	pkcs8Bundle := newTestBundle(t, pkcs8Encoder)

	tests := map[string]struct {
		meta metadata.Metadata

		testBundle testBundle
		expFiles   map[string][]byte
		expErr     bool
		isPKCS12   bool
	}{
		"if no additional attributes given, expect files to be written with NextIssuanceTime 2/3rds": {
			testBundle: pkcs1Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name": "ca-issuer",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  pkcs1Bundle.caPEM,
				"tls.crt": pkcs1Bundle.certPEM,
				"tls.key": pkcs1Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer"}}`,
				),
			},
			expErr: false,
		},
		"if renew before present, use that renew before": {
			testBundle: pkcs1Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":  "ca-issuer",
					"csi.cert-manager.io/renew-before": "48h",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  pkcs1Bundle.caPEM,
				"tls.crt": pkcs1Bundle.certPEM,
				"tls.key": pkcs1Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-02T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/renew-before":"48h"}}`,
				),
			},
			expErr: false,
		},
		"if renew before present in metadata but given a bad string, return error": {
			testBundle: pkcs1Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":  "ca-issuer",
					"csi.cert-manager.io/renew-before": "bad-duration",
				},
			},
			expFiles: map[string][]byte{
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/renew-before":"bad-duration"}}`,
				),
			},
			expErr: true,
		},
		"if custom file paths, write to those file paths": {
			testBundle: pkcs1Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":      "ca-issuer",
					"csi.cert-manager.io/ca-file":          "foo/bar",
					"csi.cert-manager.io/certificate-file": "my-crt",
					"csi.cert-manager.io/privatekey-file":  "hello/world/key",
				},
			},
			expFiles: map[string][]byte{
				"foo/bar":         pkcs1Bundle.caPEM,
				"my-crt":          pkcs1Bundle.certPEM,
				"hello/world/key": pkcs1Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/ca-file":"foo/bar","csi.cert-manager.io/certificate-file":"my-crt","csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/privatekey-file":"hello/world/key"}}`,
				),
			},
			expErr: false,
		},

		"if encoder is PKCS8, use the correct encoder": {
			testBundle: pkcs8Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":  "ca-issuer",
					"csi.cert-manager.io/key-encoding": "PKCS8",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  pkcs8Bundle.caPEM,
				"tls.crt": pkcs8Bundle.certPEM,
				"tls.key": pkcs8Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/key-encoding":"PKCS8"}}`,
				),
			},
			expErr: false,
		},

		"if encoder is unknown, return an error": {
			testBundle: pkcs8Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":  "ca-issuer",
					"csi.cert-manager.io/key-encoding": "UNKNOWN_ENCODER",
				},
			},
			expFiles: map[string][]byte{
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/key-encoding":"UNKNOWN_ENCODER"}}`,
				),
			},
			expErr: true,
		},

		"if encoder is empty, use default encoder PKCS1": {
			testBundle: pkcs1Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name": "ca-issuer",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  pkcs1Bundle.caPEM,
				"tls.crt": pkcs1Bundle.certPEM,
				"tls.key": pkcs1Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer"}}`,
				),
			},
			expErr: false,
		},

		"if encoder is empty string, use default encoder PKCS1": {
			testBundle: pkcs1Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":  "ca-issuer",
					"csi.cert-manager.io/key-encoding": "",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  pkcs1Bundle.caPEM,
				"tls.crt": pkcs1Bundle.certPEM,
				"tls.key": pkcs1Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/key-encoding":""}}`,
				),
			},
			expErr: false,
		},
		"if keystore-file is set, correct metadata should be written": {
			testBundle: pkcs8Bundle,
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":   "ca-issuer",
					"csi.cert-manager.io/key-encoding":  "PKCS8",
					"csi.cert-manager.io/keystore-file": "my-file.pfx",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  pkcs8Bundle.caPEM,
				"tls.crt": pkcs8Bundle.certPEM,
				"tls.key": pkcs8Bundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/key-encoding":"PKCS8","csi.cert-manager.io/keystore-file":"my-file.pfx","csi.cert-manager.io/keystore-type":"PKCS12"}}`,
				),
			},
			expErr:   false,
			isPKCS12: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			store := storage.NewMemoryFS()
			w := &Writer{store}

			_, err := w.Store.RegisterMetadata(test.meta)
			assert.NoError(t, err)

			testBundle := test.testBundle
			err = w.WriteKeypair(test.meta, testBundle.pk, testBundle.certPEM, testBundle.caPEM)
			assert.Equal(t, test.expErr, err != nil, "%v", err)

			files, err := store.ReadFiles("vol-id")
			assert.NoError(t, err)

			// there is some entropy in the PKCS12 process, so we cannot predict the contents of the file
			if test.isPKCS12 {
				key := test.meta.VolumeContext["csi.cert-manager.io/keystore-file"]
				assert.NotEmpty(t, files[key])

				// we'll delete the key to let the assertion for expFiles proceed
				delete(files, key)
			}
			assert.Equal(t, test.expFiles, files)
		})
	}
}
