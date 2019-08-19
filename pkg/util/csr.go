package util

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
)

// EncodeCSR calls x509.CreateCertificateRequest to sign the given CSR.
// It returns a PEM encoded signed CSR.
func EncodeCSR(csr *x509.CertificateRequest, key crypto.Signer) ([]byte, error) {
	derBytes, err := x509.CreateCertificateRequest(rand.Reader, csr, key)
	if err != nil {
		return nil, fmt.Errorf("error creating x509 certificate: %s", err.Error())
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: derBytes,
	})

	return csrPEM, nil
}

func CertificateRequestReady(cr *cmapi.CertificateRequest) bool {
	readyType := cmapi.CertificateRequestConditionReady
	readyStatus := cmapi.ConditionTrue

	existingConditions := cr.Status.Conditions
	for _, cond := range existingConditions {
		if readyType == cond.Type && readyStatus == cond.Status {
			return true
		}
	}

	return false
}

func CertificateRequestFailed(cr *cmapi.CertificateRequest) (string, bool) {
	readyType := cmapi.CertificateRequestConditionReady

	for _, con := range cr.Status.Conditions {
		if readyType == con.Type && con.Reason == "Failed" {
			return con.Message, true
		}
	}

	return "", false
}
