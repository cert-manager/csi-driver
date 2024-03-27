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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/cert-manager/csi-driver/test/e2e/framework"
	"github.com/cert-manager/csi-driver/test/e2e/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = framework.CasesDescribe("Should pick-up correct FSGroup on Pods", func() {
	f := framework.NewDefaultFramework("fs-group")

	It("should create a pod with a Group of 2000 and be able to read files with FS Group of 2000", func() {
		testVolume, testPod := basePod(f, map[string]string{
			"csi.cert-manager.io/issuer-name": f.Issuer.Name,
			"csi.cert-manager.io/fs-group":    "2000",
		})

		testPod.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			RunAsGroup: pointer.Int64(2000),
		}

		By("Creating Pod")
		testPod, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), testPod, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Pod to become ready")
		err = f.Helper().WaitForPodReady(f.Namespace.Name, testPod.Name, time.Minute)
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the corresponding CertificateRequest should exist with the correct spec")
		crs, err := f.Helper().WaitForCertificateRequestsReady(testPod, time.Second)
		Expect(err).NotTo(HaveOccurred())

		err = util.CertificateRequestMatchesSpec(crs[0], testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())
		Expect(crs).To(HaveLen(1))

		By("Ensure the certificate key pair exists in the pod and can be read by the pod")
		certData, keyData, err := f.Helper().CertificateKeyInPodPath(f.Namespace.Name, testPod.Name, "test-container-1", "/tls",
			testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())

		By("Ensure certificate key pair matches spec")
		err = f.Helper().CertificateKeyMatch(crs[0], certData, keyData)
		Expect(err).NotTo(HaveOccurred())
	})
})
