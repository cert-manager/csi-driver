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
	"bytes"
	"context"
	"fmt"
	"maps"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cert-manager/csi-driver/test/e2e/framework"
	"github.com/cert-manager/csi-driver/test/e2e/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = framework.CasesDescribe("Normal certificate renew behaviour", func() {
	f := framework.NewDefaultFramework("renew-test")

	It("should renew certificates with the same private key if set", func() {
		pod, attr := newRenewingTestPod(f, map[string]string{"csi.cert-manager.io/reuse-private-key": "true"})
		defer deletePod(f, pod)

		By("Wait for certificate to be renewed twice but keep the same private key throughout")
		cert, key, err := f.Helper().CertificateKeyInPodPath(context.TODO(), f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name, "/tls", attr)
		Expect(err).NotTo(HaveOccurred())

		for i := range 2 {
			By(fmt.Sprintf("Wait for certificate to be renewed %d", i+1))
			Eventually(func() bool {
				By("Testing pod for new certificate file")
				newCert, newKey, err := f.Helper().CertificateKeyInPodPath(context.TODO(), f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name, "/tls", attr)
				Expect(err).NotTo(HaveOccurred())

				if !bytes.Equal(cert, newCert) {
					cert = newCert
				} else {
					return false
				}

				if bytes.Equal(key, newKey) {
					return true
				}

				return false
			}, "20s", "0.5s").Should(BeTrue(), "expected renewed certificate to use same private key")
		}
	})

	It("should renew certificates with a new private key with no attribute set", func() {
		pod, attr := newRenewingTestPod(f, map[string]string{})
		defer deletePod(f, pod)

		By("Wait for certificate to be renewed and have a new private key")
		cert, key, err := f.Helper().CertificateKeyInPodPath(context.TODO(), f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name, "/tls", attr)
		Expect(err).NotTo(HaveOccurred())

		for i := range 2 {
			By(fmt.Sprintf("Wait for certificate to be renewed %d", i+1))
			Eventually(func() bool {
				By("Testing pod for new certificate file")
				newCert, newKey, err := f.Helper().CertificateKeyInPodPath(context.TODO(), f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name, "/tls", attr)
				Expect(err).NotTo(HaveOccurred())

				if !bytes.Equal(cert, newCert) {
					Expect(key).ShouldNot(Equal(newKey))
					cert = newCert
					key = newKey
					return true
				}

				return false
			}, "20s", "0.5s").Should(BeTrue(), "expected renewed certificate to use different private key")
		}
	})

	It("should renew certificates with a new private key with attribute set to false", func() {
		pod, attr := newRenewingTestPod(f, map[string]string{"csi.cert-manager.io/reuse-private-key": "false"})
		defer deletePod(f, pod)

		By("Wait for certificate to be renewed and have a new private key")
		cert, key, err := f.Helper().CertificateKeyInPodPath(context.TODO(), f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name, "/tls", attr)
		Expect(err).NotTo(HaveOccurred())

		for i := range 2 {
			By(fmt.Sprintf("Wait for certificate to be renewed %d", i+1))
			Eventually(func() bool {
				By("Testing pod for new certificate file")
				newCert, newKey, err := f.Helper().CertificateKeyInPodPath(context.TODO(), f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name, "/tls", attr)
				Expect(err).NotTo(HaveOccurred())

				if !bytes.Equal(cert, newCert) {
					Expect(key).ShouldNot(Equal(newKey))
					cert = newCert
					key = newKey
					return true
				}

				return false
			}, "20s", "0.5s").Should(BeTrue(), "expected renewed certificate to use different private key")
		}
	})
})

func newRenewingTestPod(f *framework.Framework, extraAttributes map[string]string) (*corev1.Pod, map[string]string) {
	attributes := map[string]string{
		"csi.cert-manager.io/issuer-name":  f.Issuer.Name,
		"csi.cert-manager.io/issuer-kind":  f.Issuer.Kind,
		"csi.cert-manager.io/issuer-group": f.Issuer.Group,
		"csi.cert-manager.io/dns-names":    "a.example.com",
		"csi.cert-manager.io/duration":     "10s",
	}

	maps.Copy(attributes, extraAttributes)

	testVolume, testPod := basePod(f, attributes)

	By("Creating a Pod")
	testPod, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), testPod, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	By("Waiting for Pod to become ready")
	err = f.Helper().WaitForPodReady(context.TODO(), f.Namespace.Name, testPod.Name, time.Minute)
	Expect(err).NotTo(HaveOccurred())

	testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(context.TODO(), testPod.Name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	By("Ensure the corresponding CertificateRequest should exist with the correct spec")
	crs, err := f.Helper().WaitForCertificateRequestsReady(context.TODO(), testPod, time.Second)
	Expect(err).NotTo(HaveOccurred())
	Expect(crs).To(HaveLen(1))

	err = util.CertificateRequestMatchesSpec(crs[0], testVolume.CSI.VolumeAttributes)
	Expect(err).NotTo(HaveOccurred())

	return testPod, attributes
}
