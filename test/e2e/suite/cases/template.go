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

	csi "github.com/jetstack/cert-manager-csi/pkg/apis"
	"github.com/jetstack/cert-manager-csi/pkg/util"
	"github.com/jetstack/cert-manager-csi/test/e2e/framework"
)

var _ = framework.CasesDescribe("Should template all templatable attributes correctly", func() {
	f := framework.NewDefaultFramework("attr-templating")

	It("should create a pod with a fully templated certificate request", func() {
		testVolume := corev1.Volume{
			Name: "tls",
			VolumeSource: corev1.VolumeSource{
				CSI: &corev1.CSIVolumeSource{
					Driver: csi.GroupName,
					VolumeAttributes: map[string]string{
						"csi.cert-manager.io/issuer-name":  f.Issuer.Name,
						"csi.cert-manager.io/issuer-kind":  f.Issuer.Kind,
						"csi.cert-manager.io/issuer-group": f.Issuer.Group,
						"csi.cert-manager.io/dns-names":    "a.example.com,b.example.com,{{ .PodName }}.my-service.{{ .PodNamespace }}",
						"csi.cert-manager.io/uri-sans":     "spiffe://my-service.sandbox.cluster.local,http://{{ .PodName }}.my-service.{{ .PodNamespace }}",
						"csi.cert-manager.io/ip-sans":      "192.168.0.1,123.4.5.6",
						"csi.cert-manager.io/duration":     "123h",
						"csi.cert-manager.io/is-ca":        "true",
						"csi.cert-manager.io/common-name":  "{{ .PodName }}.my-service.{{ .PodNamespace }}",
					},
				},
			},
		}

		renderedAttributes := map[string]string{
			"csi.cert-manager.io/issuer-name":  f.Issuer.Name,
			"csi.cert-manager.io/issuer-kind":  f.Issuer.Kind,
			"csi.cert-manager.io/issuer-group": f.Issuer.Group,
			"csi.cert-manager.io/dns-names":    "a.example.com,b.example.com,test-container-1.my-service." + f.Namespace.Name,
			"csi.cert-manager.io/uri-sans":     "spiffe://my-service.sandbox.cluster.local,http://test-container-1.my-service." + f.Namespace.Name,
			"csi.cert-manager.io/ip-sans":      "192.168.0.1,123.4.5.6",
			"csi.cert-manager.io/duration":     "123h",
			"csi.cert-manager.io/is-ca":        "true",
			"csi.cert-manager.io/common-name":  "test-container-1.my-service." + f.Namespace.Name,
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
		err = f.Helper().WaitForPodReady(f.Namespace.Name, testPod.Name, time.Second*10)
		Expect(err).NotTo(HaveOccurred())

		testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(testPod.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the corresponding CertificateRequest should exist with the correct spec")
		crName := util.BuildVolumeID(string(testPod.GetUID()), "tls")
		cr, err := f.Helper().WaitForCertificateRequestReady(f.Namespace.Name, crName, time.Second)
		Expect(err).NotTo(HaveOccurred())

		err = util.CertificateRequestMatchesSpec(cr, renderedAttributes)
		Expect(err).NotTo(HaveOccurred())
	})
})
