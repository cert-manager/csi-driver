package util

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
)

type KeyBundle struct {
	PrivateKey         crypto.Signer
	SignatureAlgorithm x509.SignatureAlgorithm
	PublicKeyAlgorithm x509.PublicKeyAlgorithm
}

func NewRSAKey(keyPath string) (*KeyBundle, error) {
	sk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(sk),
		},
	)

	err = WriteFile(keyPath, keyPEM, 0600)
	if err != nil {
		return nil, err
	}

	return &KeyBundle{
		PrivateKey:         sk,
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.RSA,
	}, nil
}

func WriteFile(path string, b []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		return err
	}

	return ioutil.WriteFile(path, b, perm)
}
