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
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"software.sslmate.com/src/go-pkcs12"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

// Handle will handle PKCS12 keystore options in the given Volume attributes.
// If enabled, A PKCS12 keystore file will be encoded and written to the given
// file store.
func Handle(attributes map[string]string, files map[string][]byte, pk crypto.PrivateKey, chainPEM []byte) error {
	// If PKCS12 support is not enabled, return early.
	if attributes[csiapi.KeyStorePKCS12EnableKey] != "true" {
		return nil
	}

	pfx, err := create(attributes[csiapi.KeyStorePKCS12PasswordKey], pk, chainPEM)
	if err != nil {
		return fmt.Errorf("failed to create pkcs12 file: %w", err)
	}

	// Write PKCS12 file to the file store.
	files[attributes[csiapi.KeyStorePKCS12FileKey]] = pfx

	return nil
}

// create combines the inputs to a single PKCS12 keystore file. Private key
// must be PKCS1 or PKCS8 encoded. Certificates must be PEM encoded.
func create(password string, pk crypto.PrivateKey, chainPEM []byte) ([]byte, error) {
	chain, err := pki.DecodeX509CertificateChainBytes(chainPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate chain: %w", err)
	}

	if len(chain) == 0 {
		return nil, errors.New("no certificates decoded in certificate chain")
	}

	pfx, err := pkcs12.Encode(rand.Reader, pk, chain[0], chain[1:], password)
	if err != nil {
		return nil, fmt.Errorf("failed to encode the PKCS12 certificate chain file: %v", err)
	}

	return pfx, nil
}
