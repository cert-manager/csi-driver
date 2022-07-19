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

package cases

import (
	"context"
	"time"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"software.sslmate.com/src/go-pkcs12"

	"github.com/cert-manager/csi-driver/test/e2e/framework"
)

var _ = framework.CasesDescribe("Should write keystore pkcs12 file correctly", func() {
	f := framework.NewDefaultFramework("pkcs12")

	It("should create a pod with the pkcs12 file written", func() {
		testVolume, testPod := basePod(f, map[string]string{
			"csi.cert-manager.io/issuer-name":              f.Issuer.Name,
			"csi.cert-manager.io/issuer-kind":              f.Issuer.Kind,
			"csi.cert-manager.io/issuer-group":             f.Issuer.Group,
			"csi.cert-manager.io/common-name":              "foo-bar",
			"csi.cert-manager.io/pkcs12-enable":   "true",
			"csi.cert-manager.io/pkcs12-password": "a-random-password",
			"csi.cert-manager.io/pkcs12-file":     "foo.p12",
		})

		By("Creating a Pod")
		testPod, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), testPod, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Pod to become ready")
		err = f.Helper().WaitForPodReady(f.Namespace.Name, testPod.Name, time.Minute)
		Expect(err).NotTo(HaveOccurred())

		testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(context.TODO(), testPod.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		_, err = f.Helper().WaitForCertificateRequestsReady(testPod, time.Second)
		Expect(err).NotTo(HaveOccurred())

		By("Extracting certificate and private key")
		certPEM, pkPEM, err := f.Helper().CertificateKeyInPodPath(f.Namespace.Name, testPod.Name, "test-container-1", "/tls",
			testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())

		By("Extracting PKCS12 file from Pod VolumeMount")
		pkcs12File, err := f.Helper().ReadFilePathFromContainer(f.Namespace.Name, testPod.Name, "test-container-1", "/tls/foo.p12")
		Expect(err).NotTo(HaveOccurred())

		pkcs12pk, pkcs12cert, _, err := pkcs12.DecodeChain(pkcs12File, "a-random-password")
		Expect(err).NotTo(HaveOccurred())

		cert, err := pki.DecodeX509CertificateBytes(certPEM)
		Expect(err).NotTo(HaveOccurred())
		pk, err := pki.DecodePrivateKeyBytes(pkPEM)
		Expect(err).NotTo(HaveOccurred())

		Expect(pkcs12pk).To(Equal(pk))
		Expect(pkcs12cert).To(Equal(cert))
	})
})
