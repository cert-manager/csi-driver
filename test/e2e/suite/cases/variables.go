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
	"net/url"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/cert-manager/cert-manager/pkg/util/pki"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cert-manager/csi-driver/test/e2e/framework"
)

var _ = framework.CasesDescribe("Should correctly substitute out SANs with variables", func() {
	setupPod := func(f *framework.Framework, annotations map[string]string) (*corev1.Pod, *cmapi.CertificateRequest) {
		By("Creating a ServiceAccount")
		testServiceAccount, err := f.KubeClientSet.CoreV1().ServiceAccounts(f.Namespace.Name).Create(context.TODO(), &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: f.BaseName + "-", Namespace: f.Namespace.Name},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		_, testPod := basePod(f, annotations)
		testPod.Spec.ServiceAccountName = testServiceAccount.Name

		By("Creating a Pod")
		testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), testPod, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Pod to become ready")
		err = f.Helper().WaitForPodReady(f.Namespace.Name, testPod.Name, time.Minute)
		Expect(err).NotTo(HaveOccurred())

		testPod, err = f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).Get(context.TODO(), testPod.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Ensure the corresponding CertificateRequest should exist with the correct spec")
		crs, err := f.Helper().WaitForCertificateRequestsReady(testPod, time.Second)
		Expect(err).NotTo(HaveOccurred())
		Expect(crs).To(HaveLen(1))
		return testPod, crs[0]
	}

	mustParseURI := func(uri string) *url.URL {
		puri, err := url.Parse(uri)
		Expect(err).NotTo(HaveOccurred())
		return puri
	}

	f := framework.NewDefaultFramework("san-variables")
	It("should create a pod with variables on SAN values", func() {
		pod, cr := setupPod(f, map[string]string{
			"csi.cert-manager.io/issuer-name":  f.Issuer.Name,
			"csi.cert-manager.io/issuer-kind":  f.Issuer.Kind,
			"csi.cert-manager.io/issuer-group": f.Issuer.Group,
			"csi.cert-manager.io/common-name":  "$POD_NAME.${POD_NAMESPACE}",
			"csi.cert-manager.io/dns-names":    "$POD_NAME-my-dns-$POD_NAMESPACE-${POD_UID},${POD_NAME},${POD_NAME}.${POD_NAMESPACE},$POD_NAME.${POD_NAMESPACE}.svc,${POD_UID},${SERVICE_ACCOUNT_NAME}",
			"csi.cert-manager.io/uri-sans":     "spiffe://foo.bar/${POD_NAMESPACE}/$POD_NAME/$POD_UID,file://foo-bar,${POD_UID}",
		})

		request, err := pki.DecodeX509CertificateRequestBytes(cr.Spec.Request)
		Expect(err).NotTo(HaveOccurred())

		Expect(request.Subject.CommonName).To(Equal(fmt.Sprintf("%s.%s", pod.Name, pod.Namespace)))
		Expect(request.DNSNames).To(ConsistOf([]string{
			fmt.Sprintf("%s-my-dns-%s-%s", pod.Name, pod.Namespace, pod.UID),
			pod.Name,
			fmt.Sprintf("%s.%s", pod.Name, pod.Namespace),
			fmt.Sprintf("%s.%s.svc", pod.Name, pod.Namespace),
			string(pod.UID),
			pod.Spec.ServiceAccountName,
		}))
		Expect(request.URIs).To(ConsistOf([]*url.URL{
			mustParseURI(fmt.Sprintf("spiffe://foo.bar/%s/%s/%s", pod.Namespace, pod.Name, pod.UID)),
			mustParseURI("file://foo-bar"),
			mustParseURI(string(pod.UID)),
		}))
	})
})
