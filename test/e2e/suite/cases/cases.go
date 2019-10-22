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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	csi "github.com/jetstack/cert-manager-csi/pkg/apis"
	"github.com/jetstack/cert-manager-csi/pkg/util"
	"github.com/jetstack/cert-manager-csi/test/e2e/framework"
)

const (
	testImage = "busybox"
)

var _ = framework.CasesDescribe("Normal CSI behaviour", func() {
	f := framework.NewDefaultFramework("ca-issuer")

	It("should create a pod with a single volume key pair mounted with all attributes set", func() {
		testVolume := corev1.Volume{
			Name: "tls",
			VolumeSource: corev1.VolumeSource{
				CSI: &corev1.CSIVolumeSource{
					Driver: csi.GroupName,
					VolumeAttributes: map[string]string{
						"csi.cert-manager.io/issuer-name":  f.Issuer.Name,
						"csi.cert-manager.io/issuer-kind":  f.Issuer.Kind,
						"csi.cert-manager.io/issuer-group": f.Issuer.Group,
						"csi.cert-manager.io/dns-names":    "a.example.com,b.example.com",
						"csi.cert-manager.io/uri-sans":     "spiffe://my-service.sandbox.cluster.local,http://foo.bar",
						"csi.cert-manager.io/ip-sans":      "192.168.0.1,123.4.5.6",
						"csi.cert-manager.io/duration":     "123h",
						"csi.cert-manager.io/is-ca":        "true",
						"csi.cert-manager.io/common-name":  "foo-bar",
					},
				},
			},
		}

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
		err = f.Helper().MetaDataCertificateKeyExistInHostPath(cr, testPod, testVolume.CSI.VolumeAttributes, testVolume.Name, "/tmp/cert-manager-csi")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create 2 pods with random containers, volumes, and attributes set", func() {
		// TODO (@joshvanl): fix pointer nonsense

		pods := make([]corev1.Pod, 2)
		for i := range pods {
			pods[i] = *f.RandomPod()
			By(fmt.Sprintf("%v\n", pods[i].Spec))
		}

		for i := range pods {
			By(fmt.Sprintf("Creating a Pod %d", i))
			_, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(&pods[i])
			Expect(err).NotTo(HaveOccurred())
		}

		podsList, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		pods = podsList.Items

		for i := range pods {
			By(fmt.Sprintf("Waiting for Pod to become ready %d: %s", i, pods[i].Name))
			err := f.Helper().WaitForPodReady(f.Namespace.Name, pods[i].Name)
			Expect(err).NotTo(HaveOccurred())

			pod, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(pods[i].Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			pods[i] = *pod
		}

		for i, pod := range pods {
			By(fmt.Sprintf("Ensuring corresponding CertificateRequests exists with the correct spec %d: %s", i, pods))

			crMap := make(map[string]*cmapi.CertificateRequest)

			for _, vol := range pod.Spec.Volumes {
				crName := util.BuildVolumeName(pod.Name, util.BuildVolumeID(string(pod.GetUID()), vol.Name))

				cr, err := f.Helper().WaitForCertificateRequestReady(f.Namespace.Name, crName, time.Second)
				Expect(err).NotTo(HaveOccurred())

				err = f.Helper().CertificateRequestMatchesSpec(cr, vol.CSI.VolumeAttributes)
				Expect(err).NotTo(HaveOccurred())

				crMap[vol.Name] = cr
			}

			for _, container := range pod.Spec.Containers {
				By(fmt.Sprintf("Ensure the certificate key pairs exists in the pod's container and matches that in the CertificateRequest %d: %s:%s", i, pod.Name, container.Name))
				for _, vol := range pod.Spec.Volumes {
					err := f.Helper().CertificateKeyExistInPodPath(f.Namespace.Name, pod.Name, container.Name, vol.Name, crMap[vol.Name], vol.CSI.VolumeAttributes)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			for _, vol := range pod.Spec.Volumes {
				By(fmt.Sprintf("Ensure the certificate key pairs and metadata files exists in the host's data directory and matches that in the CertificateRequests %d: %s", i, pod.Name))
				err := f.Helper().MetaDataCertificateKeyExistInHostPath(crMap[vol.Name], &pod, vol.CSI.VolumeAttributes, vol.Name, "/tmp/cert-manager-csi")
				Expect(err).NotTo(HaveOccurred())
			}
		}
	})
})
