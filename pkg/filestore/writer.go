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
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"

	"github.com/cert-manager/csi-driver/pkg/apis/defaults"
	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	"github.com/cert-manager/csi-driver/pkg/apis/validation"
)

// Writer wraps the storage backend to allow access for writing data.
type Writer struct {
	Store storage.Interface
}

// WriteKeypair writes the given certificate, CA, and private key data to their
// respective file locations, according to the volume attributes. Also writes
// or updates the metadata file, including a calculated NextIssuanceTime.
func (w *Writer) WriteKeypair(meta metadata.Metadata, key crypto.PrivateKey, chain []byte, ca []byte) error {
	attrs, err := defaults.SetDefaultAttributes(meta.VolumeContext)
	if err != nil {
		return err
	}
	if err := validation.ValidateAttributes(attrs); err != nil {
		return err.ToAggregate()
	}

	var pemBlock *pem.Block

	switch keyEncodingFormat := attrs[csiapi.KeyEncodingKey]; keyEncodingFormat {
	case string(cmapi.PKCS1):
		pemBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key.(*rsa.PrivateKey)),
		}
	case string(cmapi.PKCS8):
		bytes, err := x509.MarshalPKCS8PrivateKey(key.(*rsa.PrivateKey))
		if err != nil {
			return fmt.Errorf("marshalling pkcs8 private key: %w", err)
		}

		pemBlock = &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: bytes,
		}
	default:
		return fmt.Errorf("invalid key encoding format: %s", keyEncodingFormat)
	}

	keyPEM := pem.EncodeToMemory(pemBlock)

	// Calculate the next issuance time and check errors before writing files.
	// This prevents cases where we write files but also have errors in the
	// nextIssuanceTime, putting the volume into a bad state.
	nextIssuanceTime, err := calculateNextIssuanceTime(attrs, chain)
	if err != nil {
		return fmt.Errorf("calculating next issuance time: %w", err)
	}

	if err := w.Store.WriteFiles(meta, map[string][]byte{
		attrs[csiapi.KeyFileKey]:  keyPEM,
		attrs[csiapi.CertFileKey]: chain,
		attrs[csiapi.CAFileKey]:   ca,
	}); err != nil {
		return fmt.Errorf("writing data: %w", err)
	}

	meta.NextIssuanceTime = &nextIssuanceTime
	if err := w.Store.WriteMetadata(meta.VolumeID, meta); err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}

	return nil
}

// calculateNextIssuanceTime will return the time at when the certificate
// should be renewed by the driver. By default, this will return the time at
// when the issued certificate is 2/3rds through its lifetime (NotAfter -
// NotBefore).
//
// The volume attribute `csi.cert-manager.io/renew-before` can be used to
// overwrite the default behaviour with a custom renew time. If this duration
// results in a renew time before the NotBefore of the signed certificate
// itself, it will fall back to returning 2/3rds the certificate lifetime.
func calculateNextIssuanceTime(attrs map[string]string, chain []byte) (time.Time, error) {
	block, _ := pem.Decode(chain)
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing issued certificate: %w", err)
	}

	actualDuration := crt.NotAfter.Sub(crt.NotBefore)

	// if not explicitly set, renew once a certificate is 2/3rds of the way
	// through its lifetime.
	renewBeforeNotAfter := actualDuration / 3

	if v, ok := attrs[csiapi.RenewBeforeKey]; ok {
		renewBeforeDuration, err := time.ParseDuration(v)
		if err != nil {
			return time.Time{}, fmt.Errorf("parsing requested renew-before duration %q: %w", csiapi.RenewBeforeKey, err)
		}

		// If the requested renewBefore would cause the certificate to be
		// immediately re-issued, ignore the requested renew before and renew
		// 2/3rds of the way through its lifetime.
		if crt.NotBefore.Add(renewBeforeDuration).Before(crt.NotAfter) {
			renewBeforeNotAfter = renewBeforeDuration
		}
	}

	return crt.NotAfter.Add(-renewBeforeNotAfter), nil
}
