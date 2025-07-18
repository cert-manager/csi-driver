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

package keygen

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"strings"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/cert-manager/csi-driver/pkg/apis/defaults"
	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	"github.com/cert-manager/csi-driver/pkg/apis/validation"
)

// Generator wraps the storage backend to allow for re-using private keys when
// re-issuing a certificate.
type Generator struct {
	Store *storage.Filesystem
}

// KeyForMetadata generates a new private key, or returns an existing
// one if the reuse private key attribute is present.
func (k *Generator) KeyForMetadata(meta metadata.Metadata) (crypto.PrivateKey, error) {
	attrs, err := defaults.SetDefaultAttributes(meta.VolumeContext)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateAttributes(attrs); err != nil {
		return nil, err.ToAggregate()
	}

	// By default, generate a new private key each time.
	if attrs[csiapi.ReusePrivateKey] != "true" {
		return newKey(attrs)
	}

	bytes, err := k.Store.ReadFile(meta.VolumeID, attrs[csiapi.KeyFileKey])
	if errors.Is(err, storage.ErrNotFound) {
		// Generate a new key if one is not found on disk
		return newKey(attrs)
	}
	if err != nil {
		return nil, err
	}

	pk, err := pki.DecodePrivateKeyBytes(bytes)
	if err != nil {
		// Generate a new key if the existing one cannot be decoded
		return newKey(attrs)
	}

	return pk, nil
}

func newKey(attrs map[string]string) (crypto.PrivateKey, error) {
	switch algo := attrs[csiapi.KeyAlgorithmKey]; algo {
	case string(cmapi.RSAKeyAlgorithm):
		size, err := strconv.Atoi(attrs[csiapi.KeySizeKey])
		if err != nil {
			return nil, err
		}
		return rsa.GenerateKey(rand.Reader, size)
	case string(cmapi.ECDSAKeyAlgorithm):
		size, err := strconv.Atoi(attrs[csiapi.KeySizeKey])
		if err != nil {
			return nil, err
		}
		var curve elliptic.Curve
		switch size {
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		}
		return ecdsa.GenerateKey(curve, rand.Reader)
	case string(cmapi.Ed25519KeyAlgorithm):
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		return privateKey, err
	default:
		validValues := []cmapi.PrivateKeyAlgorithm{cmapi.RSAKeyAlgorithm, cmapi.ECDSAKeyAlgorithm, cmapi.Ed25519KeyAlgorithm}
		quotedValues := make([]string, len(validValues))
		for i, v := range validValues {
			quotedValues[i] = strconv.Quote(fmt.Sprint(v))
		}
		return nil, &field.Error{
			Type:     field.ErrorTypeNotSupported,
			Field:    csiapi.KeyAlgorithmKey,
			BadValue: algo,
			Detail:   "supported values: " + strings.Join(quotedValues, ", "),
		}
	}
}
