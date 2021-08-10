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

func newTestBundle(t *testing.T) testBundle {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	pkBytes := x509.MarshalPKCS1PrivateKey(pk)
	pkPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkBytes})

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

	return testBundle{[]byte("CA Certificate"), certPEM, pk, pkPEM}
}

func Test_calculateNextIssuanceTime(t *testing.T) {
	testBundle := newTestBundle(t)

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
		"if disable auto renew present and `true`, return year 9999": {
			attrs: map[string]string{
				"csi.cert-manager.io/disable-auto-renew": "true",
			},
			expTime: time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC),
			expErr:  false,
		},
		"if disable auto renew present and `false`, return 2/3rds": {
			attrs: map[string]string{
				"csi.cert-manager.io/disable-auto-renew": "false",
			},
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
		"if renew before present but also disable auto-renew, return year 9999": {
			attrs: map[string]string{
				"csi.cert-manager.io/disable-auto-renew": "true",
				"csi.cert-manager.io/renew-before":       "48h",
			},
			expTime: time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC),
			expErr:  false,
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
	testBundle := newTestBundle(t)

	tests := map[string]struct {
		meta metadata.Metadata

		expFiles map[string][]byte
		expErr   bool
	}{
		"if no additional attributes given, expect files to be written with NextIssuanceTime 2/3rds": {
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name": "ca-issuer",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  testBundle.caPEM,
				"tls.crt": testBundle.certPEM,
				"tls.key": testBundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer"}}`,
				),
			},
			expErr: false,
		},
		"if renew before present, use that renew before": {
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":  "ca-issuer",
					"csi.cert-manager.io/renew-before": "48h",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  testBundle.caPEM,
				"tls.crt": testBundle.certPEM,
				"tls.key": testBundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-02T00:00:00Z","volumeContext":{"csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/renew-before":"48h"}}`,
				),
			},
			expErr: false,
		},
		"if renew before present in metadata but given a bad string, return error": {
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
		"if disable renew before, set NextIssuanceTime to year 9999": {
			meta: metadata.Metadata{
				VolumeID:   "vol-id",
				TargetPath: "/target-path",
				VolumeContext: map[string]string{
					"csi.cert-manager.io/issuer-name":        "ca-issuer",
					"csi.cert-manager.io/disable-auto-renew": "true",
				},
			},
			expFiles: map[string][]byte{
				"ca.crt":  testBundle.caPEM,
				"tls.crt": testBundle.certPEM,
				"tls.key": testBundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"9999-01-01T00:00:00Z","volumeContext":{"csi.cert-manager.io/disable-auto-renew":"true","csi.cert-manager.io/issuer-name":"ca-issuer"}}`,
				),
			},
			expErr: false,
		},
		"if custom file paths, write to those file paths": {
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
				"foo/bar":         testBundle.caPEM,
				"my-crt":          testBundle.certPEM,
				"hello/world/key": testBundle.pkPEM,
				"metadata.json": []byte(
					`{"volumeID":"vol-id","targetPath":"/target-path","nextIssuanceTime":"1970-01-03T00:00:00Z","volumeContext":{"csi.cert-manager.io/ca-file":"foo/bar","csi.cert-manager.io/certificate-file":"my-crt","csi.cert-manager.io/issuer-name":"ca-issuer","csi.cert-manager.io/privatekey-file":"hello/world/key"}}`,
				),
			},
			expErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			store := storage.NewMemoryFS()
			w := &Writer{store}

			_, err := w.Store.RegisterMetadata(test.meta)
			assert.NoError(t, err)

			err = w.WriteKeypair(test.meta, testBundle.pk, testBundle.certPEM, testBundle.caPEM)
			assert.Equal(t, test.expErr, err != nil, "%v", err)

			files, err := store.ReadFiles("vol-id")
			assert.NoError(t, err)
			assert.Equal(t, test.expFiles, files)
		})
	}
}
