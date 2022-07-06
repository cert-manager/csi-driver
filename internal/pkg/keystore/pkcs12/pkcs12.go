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
)

// Create combines the inputs to a single PKCS12 keystore file.
// key must be PKCS1 or PKCS8 encoded. certificates must be PEM encoded.
func Create(key crypto.PrivateKey, chainPEM []byte, rootPEM []byte) ([]byte, error) {
	if key == nil {
		return nil, errors.New("key must not be nil")
	}

	if len(chainPEM) == 0 {
		return nil, errors.New("chain must not be empty")
	}

	if len(rootPEM) == 0 {
		return nil, errors.New("root must not be empty")
	}

	rc, err := pki.DecodeX509CertificateBytes(rootPEM)
	if err != nil {
		return nil, fmt.Errorf("pki.DecodeX509CertificateChainBytes(rootPEM): %v", err)
	}

	cc, err := pki.DecodeX509CertificateChainBytes(chainPEM)
	if err != nil {
		return nil, fmt.Errorf("pki.DecodeX509CertificateChainBytes(chainPEM): %v", err)
	}

	// we need to grab the leaf cert from chain
	// TODO: is it the first cert or the last?
	// leaf is the last cert - right?
	leaf := cc[len(cc)-1]
	cc = cc[:len(cc)-1]

	// add the root cert to the back of the chain
	cc = append(cc, rc)

	pfx, err := pkcs12.Encode(rand.Reader, key, leaf, cc, pkcs12.DefaultPassword)
	if err != nil {
		return nil, fmt.Errorf("pkcs12.Encode: %v", err)
	}

	return pfx, nil
}
