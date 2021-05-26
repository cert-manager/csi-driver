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

package helper

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/jetstack/cert-manager-csi/pkg/util"
	apiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/jetstack/cert-manager/test/e2e/framework/log"
)

// WaitForCertificateRequestReady waits for the CertificateRequest resource to
// enter a Ready state.
func (h *Helper) WaitForCertificateRequestReady(namespace, name string, timeout time.Duration) (*cmapi.CertificateRequest, error) {
	var cr *cmapi.CertificateRequest

	err := wait.PollImmediate(time.Second/4, timeout,
		func() (bool, error) {
			var err error
			log.Logf("Waiting for CertificateRequest %s/%s to be ready", namespace, name)
			cr, err = h.CMClient.CertmanagerV1().CertificateRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if k8sErrors.IsNotFound(err) {
				return false, nil
			}

			if err != nil {
				return false, fmt.Errorf("error getting CertificateRequest %s: %v", name, err)
			}

			isReady := apiutil.CertificateRequestHasCondition(cr, cmapi.CertificateRequestCondition{
				Type:   cmapi.CertificateRequestConditionReady,
				Status: cmmeta.ConditionTrue,
			})
			if !isReady {
				log.Logf("Expected CertificateRequest %s/%s to have Ready condition 'true' but it has: %v",
					namespace, name, cr.Status.Conditions)
				return false, nil
			}
			return true, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (h *Helper) FindCertificateRequestReady(crs []cmapi.CertificateRequest, pod *corev1.Pod, volM *corev1.VolumeMount) (*cmapi.CertificateRequest, error) {
	crName := util.BuildVolumeID(string(pod.GetUID()), volM.Name)
	cr, err := h.findCertificateRequest(crs, crName)
	if err != nil {
		return nil, err
	}

	if !util.CertificateRequestReady(cr) {
		return nil, fmt.Errorf("CertificateRequest not ready: %+v", cr)
	}

	return cr, nil
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

func (h *Helper) findCertificateRequest(crs []cmapi.CertificateRequest, name string) (*cmapi.CertificateRequest, error) {
	for _, cr := range crs {
		if cr.Name == name {
			return &cr, nil
		}
	}

	return nil, fmt.Errorf("failed to find CertificateRequest %q", name)
}
