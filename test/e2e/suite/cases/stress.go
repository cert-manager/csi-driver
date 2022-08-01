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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cert-manager/csi-driver/test/e2e/framework"
	"github.com/cert-manager/csi-driver/test/e2e/util"
)

var _ = framework.CasesDescribe("Normal CSI behaviour", func() {
	f := framework.NewDefaultFramework("stress-test")

	It("should create a pod with a single volume key pair mounted with all attributes set", func() {
		testVolume, testPod := basePod(f, map[string]string{
			"csi.cert-manager.io/issuer-name":  f.Issuer.Name,
			"csi.cert-manager.io/issuer-kind":  f.Issuer.Kind,
			"csi.cert-manager.io/issuer-group": f.Issuer.Group,
			"csi.cert-manager.io/dns-names":    "a.example.com,b.example.com",
			"csi.cert-manager.io/uri-sans":     "spiffe://my-service.sandbox.cluster.local,http://foo.bar",
			"csi.cert-manager.io/ip-sans":      "192.168.0.1,123.4.5.6",
			"csi.cert-manager.io/duration":     "123h",
			"csi.cert-manager.io/is-ca":        "true",
			"csi.cert-manager.io/common-name":  "foo-bar",
		})

		By("Creating a Pod")
		testPod, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), testPod, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Pod to become ready")
		err = f.Helper().WaitForPodReady(f.Namespace.Name, testPod.Name, time.Minute)
		Expect(err).NotTo(HaveOccurred())

		testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(context.TODO(), testPod.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the corresponding CertificateRequests should exist with the correct spec")
		crs, err := f.Helper().WaitForCertificateRequestsReady(testPod, time.Second)
		Expect(err).NotTo(HaveOccurred())
		Expect(crs).To(HaveLen(1))

		err = util.CertificateRequestMatchesSpec(crs[0], testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the certificate key pair exists in the pod and matches that in the CertificateRequest")
		certData, keyData, err := f.Helper().CertificateKeyInPodPath(f.Namespace.Name, testPod.Name, "test-container-1", "/tls",
			testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())

		err = f.Helper().CertificateKeyMatch(crs[0], certData, keyData)
		Expect(err).NotTo(HaveOccurred())

		deletePod(f, testPod)
	})
})

func deletePod(f *framework.Framework, pod *corev1.Pod) {
	By("Deleting Pod " + pod.Name)
	err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = f.Helper().WaitForPodDeletion(pod.Namespace, pod.Name, time.Second*90)
	Expect(err).NotTo(HaveOccurred())

	By("Pod Deleted " + pod.Name)
}

func testPod(f *framework.Framework, pod *corev1.Pod) {
	By(fmt.Sprintf("Ensuring corresponding CertificateRequests exists with the correct spec: %s/%s", pod.Namespace, pod.Name))

	attributesMap := make(map[string]*map[string]string)

	// Not all defined volumes will be mounted. This means that the
	// NodePublishVolume will not be called and therefore no
	// CertificateRequest will be created. This is by design.
	for _, vol := range pod.Spec.Volumes {
		// Ignore non csi volumes
		if vol.VolumeSource.CSI == nil {
			continue
		}

		attributesMap[vol.Name] = &vol.CSI.VolumeAttributes
	}

	crs, err := f.CertManagerClientSet.CertmanagerV1().CertificateRequests(f.Namespace.Name).List(context.TODO(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	for _, container := range pod.Spec.Containers {
		By(fmt.Sprintf("Ensure the certificate key pairs exists in the pod's container and matches that in the CertificateRequest: %s/%s:%s", pod.Namespace, pod.Name, container.Name))
		for _, vol := range container.VolumeMounts {
			// Ignore non csi volumes
			if _, ok := attributesMap[vol.Name]; !ok {
				continue
			}

			crs, err := f.Helper().FindCertificateRequestsReady(crs.Items, pod)
			Expect(err).NotTo(HaveOccurred())

			var matchedCR *cmapi.CertificateRequest
			for _, cr := range crs {
				if err = util.CertificateRequestMatchesSpec(cr, *attributesMap[vol.Name]); err == nil {
					matchedCR = cr
					break
				}
			}
			Expect(matchedCR).ShouldNot(BeNil(), "expected one CertificateRequest to match the volume spec")

			certData, keyData, err := f.Helper().CertificateKeyInPodPath(f.Namespace.Name, pod.Name, container.Name, vol.MountPath,
				*attributesMap[vol.Name])

			err = f.Helper().CertificateKeyMatch(matchedCR, certData, keyData)
			Expect(err).NotTo(HaveOccurred())
		}
	}
}
