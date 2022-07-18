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
	"fmt"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/cert-manager/csi-lib/metadata"
	"software.sslmate.com/src/go-pkcs12"
)

// Create combines the inputs to a single PKCS12 keystore file.
// Private key must be PKCS1 or PKCS8 encoded. Certificates must be PEM
// encoded.
func Create(meta metadata.Metadata, pk crypto.PrivateKey, chainPEM []byte) ([]byte, error) {
	chain, err := pki.DecodeX509CertificateChainBytes(chainPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate chain: %w", err)
	}

	if len(chain) == 0 {
		return nil, fmt.Errorf("no certificates decoded in certificate chain: %w", err)
	}

	pfx, err := pkcs12.Encode(rand.Reader, pk, chain[0], chain[1:], meta.VolumeContext[csiapi.KeyStorePKCS12PasswordKey])
	if err != nil {
		return nil, fmt.Errorf("failed to encode the PKCS12 certificate chain file: %v", err)
	}

	return pfx, nil
}
