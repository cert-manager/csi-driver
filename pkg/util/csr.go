/*
Copyright 2019 The Jetstack cert-manager contributors.

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

package util

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
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
	readyStatus := cmmeta.ConditionTrue

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

func RenewTimeFromNotAfter(notBefore time.Time, notAfter time.Time, renewBeforeString string) (time.Duration, error) {
	renewBefore, err := time.ParseDuration(renewBeforeString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse renew before: %s", err)
	}

	validity := notAfter.Sub(notBefore)
	if renewBefore > validity {
		return 0, fmt.Errorf("renewal duration is longer than certificate validity: %s %s",
			renewBefore, validity)
	}

	dur := notAfter.Add(-renewBefore).Sub(time.Now())

	return dur, nil
}

func KeyUsagesFromAttributes(attr map[string]string) []cmapi.KeyUsage {
	usageCSV := attr[csiapi.KeyUsagesKey]

	if len(usageCSV) == 0 {
		return nil
	}

	var keyUsages []cmapi.KeyUsage
	for _, usage := range strings.Split(usageCSV, ",") {
		keyUsages = append(keyUsages, cmapi.KeyUsage(usage))
	}

	return keyUsages

}
