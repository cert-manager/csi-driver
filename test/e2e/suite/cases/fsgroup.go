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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	csi "github.com/cert-manager/csi-driver/pkg/apis"
	"github.com/cert-manager/csi-driver/test/e2e/framework"
	"github.com/cert-manager/csi-driver/test/e2e/util"
)

var _ = framework.CasesDescribe("Should pick-up correct FSGroup on Pods", func() {
	f := framework.NewDefaultFramework("fs-group")

	It("should create a pod with a Group of 2000 and be able to read files with FS Group of 2000", func() {
		testVolume := corev1.Volume{
			Name: "tls",
			VolumeSource: corev1.VolumeSource{
				CSI: &corev1.CSIVolumeSource{
					Driver:   csi.GroupName,
					ReadOnly: boolPtr(true),
					VolumeAttributes: map[string]string{
						"csi.cert-manager.io/issuer-name": f.Issuer.Name,
						"csi.cert-manager.io/fs-group":    "2000",
					},
				},
			},
		}

		var group int64 = 2000
		testPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: f.BaseName + "-",
				Namespace:    f.Namespace.Name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:    "test-container-1",
						Image:   "busybox",
						Command: []string{"sleep", "10000"},
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tls",
								Name:      "tls",
							},
						},
						SecurityContext: &corev1.SecurityContext{
							RunAsGroup: &group,
						},
					},
				},
				Volumes: []corev1.Volume{
					testVolume,
				},
			},
		}

		By("Creating a Pod")
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
