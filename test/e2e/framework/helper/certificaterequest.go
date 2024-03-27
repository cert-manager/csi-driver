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

package helper

import (
	"context"
	"fmt"
	"time"

	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/cert-manager/csi-driver/test/e2e/framework/log"
	"github.com/cert-manager/csi-driver/test/e2e/util"
)

// WaitForCertificateRequestReady waits for the CertificateRequest resources to
// enter a Ready state.
func (h *Helper) WaitForCertificateRequestsReady(pod *corev1.Pod, timeout time.Duration) ([]*cmapi.CertificateRequest, error) {
	var crs []*cmapi.CertificateRequest

	err := wait.PollImmediate(time.Second/4, timeout,
		func() (bool, error) {
			crList, err := h.CMClient.CertmanagerV1().CertificateRequests(pod.Namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return false, err
			}

			crs, err = h.findCertificateRequests(crList.Items, pod.UID)
			if err != nil {
				log.Logf("Cannot find CertificateRequests for pod, waiting...")
				return false, nil
			}

			for _, cr := range crs {
				isReady := apiutil.CertificateRequestHasCondition(cr, cmapi.CertificateRequestCondition{
					Type:   cmapi.CertificateRequestConditionReady,
					Status: cmmeta.ConditionTrue,
				})
				if !isReady {
					log.Logf("Expected CertificateRequest for Pod %s/%s to have Ready condition 'true' but it has: %v",
						pod.Namespace, pod.Name, cr.Status.Conditions)
					return false, nil
				}
			}

			return true, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return crs, nil
}

func (h *Helper) FindCertificateRequestsReady(crs []cmapi.CertificateRequest, pod *corev1.Pod) ([]*cmapi.CertificateRequest, error) {
	podCRs, err := h.findCertificateRequests(crs, pod.GetUID())
	if err != nil {
		return nil, err
	}

	for _, cr := range podCRs {
		if !util.CertificateRequestReady(cr) {
			return nil, fmt.Errorf("CertificateRequest not ready: %+v", cr)
		}
	}

	return podCRs, nil
}

func (h *Helper) WaitForCertificateRequestDeletion(namespace, name string, timeout time.Duration) error {
	log.Logf("Waiting for CertificateRequest to be deleted %s/%s", namespace, name)
	err := wait.PollImmediate(time.Second/2, timeout, func() (bool, error) {
		cr, err := h.CMClient.CertmanagerV1().CertificateRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) {
			return true, nil
		}

		if err != nil {
			return false, err
		}

		log.Logf("helper: CertificateRequest not deleted %s/%s: %v",
			cr.Namespace, cr.Name, cr.Status.Conditions)

		return false, nil
	})
	if err != nil {
		h.Kubectl(namespace).DescribeResource("certificaterequest", name)
		return err
	}

	return nil
}

func (h *Helper) findCertificateRequests(crs []cmapi.CertificateRequest, podUID types.UID) ([]*cmapi.CertificateRequest, error) {
	var podCRs []*cmapi.CertificateRequest

	for i, cr := range crs {
		if len(cr.OwnerReferences) == 0 {
			continue
		}
		if cr.OwnerReferences[0].UID == podUID {
			podCRs = append(podCRs, &crs[i])
		}
	}

	if len(podCRs) == 0 {
		return nil, fmt.Errorf("failed to find CertificateRequest owned by pod with UID %q", podUID)
	}

	return podCRs, nil
}
