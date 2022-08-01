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
	"crypto/rand"
	"crypto/rsa"
	"errors"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/cert-manager/csi-lib/metadata"
	"github.com/cert-manager/csi-lib/storage"

	"github.com/cert-manager/csi-driver/pkg/apis/defaults"
	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	"github.com/cert-manager/csi-driver/pkg/apis/validation"
)

// Generator wraps the storage backend to allow for re-using private keys when
// re-issuing a certificate.
// It generates 2048-bit RSA private keys.
type Generator struct {
	Store *storage.Filesystem
}

// KeyForMetadata generates a 2048-bit RSA private key, or returns an existing
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
		return new2048BitRSAKey()
	}

	bytes, err := k.Store.ReadFile(meta.VolumeID, attrs[csiapi.KeyFileKey])
	if errors.Is(err, storage.ErrNotFound) {
		// Generate a new key if one is not found on disk
		return new2048BitRSAKey()
	}
	if err != nil {
		return nil, err
	}

	pk, err := pki.DecodePrivateKeyBytes(bytes)
	if err != nil {
		// Generate a new key if the existing one cannot be decoded
		return new2048BitRSAKey()
	}

	return pk, nil
}

func new2048BitRSAKey() (crypto.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}
