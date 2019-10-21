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

package cases

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/cert-manager-csi/pkg/util"
	"github.com/jetstack/cert-manager-csi/test/e2e/framework"
	testutil "github.com/jetstack/cert-manager-csi/test/e2e/suite/util"
)

const (
	testImage = "busybox"
)

var _ = framework.CasesDescribe("Normal CSI behaviour", func() {
	f := framework.NewDefaultFramework("ca-issuer")

	It("should create a pod with a single volume key pair mounted", func() {
		testVolume := testutil.ConstructCSIVolume("tls", map[string]string{
			"csi.cert-manager.io/issuer-name":  "ca-issuer",
			"csi.cert-manager.io/issuer-kind":  "Issuer",
			"csi.cert-manager.io/issuer-group": "cert-manager.io",
			"csi.cert-manager.io/dns-names":    "a.example.com,b.example.com",
			"csi.cert-manager.io/uri-sans":     "spiffe://my-service.sandbox.cluster.local,http://foo.bar",
			"csi.cert-manager.io/ip-sans":      "192.168.0.1,123.4.5.6",
			"csi.cert-manager.io/duration":     "123h",
			"csi.cert-manager.io/is-ca":        "true",
			"csi.cert-manager.io/common-name":  "foo-bar",
		})

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
					},
				},
				Volumes: []corev1.Volume{
					testVolume,
				},
			},
		}

		By("Creating a Pod")
		testPod, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(testPod)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Pod to become ready")
		err = f.Helper().WaitForPodReady(f.Namespace.Name, testPod.Name)
		Expect(err).NotTo(HaveOccurred())

		testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(testPod.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the corresponding CertificateRequest should exist with the correct spec")
		crName := util.BuildVolumeName(testPod.Name,
			util.BuildVolumeID(string(testPod.GetUID()), "tls"))
		cr, err := f.Helper().WaitForCertificateRequestReady(f.Namespace.Name, crName, time.Second)
		Expect(err).NotTo(HaveOccurred())

		err = f.Helper().CertificateRequestMatchesSpec(cr, testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the certificate key pair exists in the pod and matches that in the CertificateRequest")
		err = f.Helper().CertificateKeyExistInPodPath(f.Namespace.Name, testPod.Name, "test-container-1", "/tls",
			cr, testVolume.CSI.VolumeAttributes)
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the certificate key pair and metadata file exists in the local data directory and matches that in the CertificateRequest")
		err = f.Helper().MetaDataCertificateKeyExistInHostPath(cr, testPod, testVolume.CSI.VolumeAttributes, "tls", "/tmp/cert-manager-csi")
		Expect(err).NotTo(HaveOccurred())
	})
})
