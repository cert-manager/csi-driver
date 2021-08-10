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
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/jetstack/cert-manager/pkg/util/pki"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager/test/e2e/framework/log"
)

func (h *Helper) CertificateKeyInPodPath(namespace, podName, containerName, mountPath string,
	attr map[string]string) ([]byte, []byte, error) {
	certPath, ok := attr[csiapi.CertFileKey]
	if !ok {
		certPath = "tls.crt"
	}
	certPath = filepath.Join(mountPath, certPath)

	keyPath, ok := attr[csiapi.KeyFileKey]
	if !ok {
		keyPath = "tls.key"
	}
	keyPath = filepath.Join(mountPath, keyPath)

	certData, err := h.readFilePathFromContainer(namespace, podName, containerName, certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read cert data from pod: %s", err)
	}

	keyData, err := h.readFilePathFromContainer(namespace, podName, containerName, keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key data from pod: %s", err)
	}

	return certData, keyData, nil
}

func (h *Helper) CertificateKeyMatch(cr *cmapi.CertificateRequest, certData, keyData []byte) error {
	if !bytes.Equal(certData, cr.Status.Certificate) {
		return fmt.Errorf("certificate does not match that in the CertificateRequest %q, exp=%s got=%s",
			cr.Name, cr.Status.Certificate, certData)
	}

	cert, err := pki.DecodeX509CertificateBytes(certData)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %s", err)
	}

	key, err := pki.DecodePrivateKeyBytes(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse key data: %s", err)
	}

	ok, err := pki.PublicKeyMatchesCertificate(key.Public(), cert)
	if err != nil {
		return fmt.Errorf("failed to check key matches certificate: %s", err)
	}

	if !ok {
		return fmt.Errorf("private key does not match certificate %s\n%s",
			certData, keyData)
	}

	return nil
}

func (h *Helper) readFilePathFromContainer(namespace, podName, containerName, path string) ([]byte, error) {
	coreclient, err := corev1client.NewForConfig(h.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build core client form rest config: %s", err)
	}

	log.Logf("helper: reading from file %s:%s:%s:%s",
		namespace, podName, containerName, path)

	// TODO (@joshvanl): use tar compression
	req := coreclient.RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"cat", path},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(h.RestConfig, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	execOut, execErr := new(bytes.Buffer), new(bytes.Buffer)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: execOut,
		Stderr: execErr,
		Tty:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create exec stream (%s): %s", execErr.String(), err)
	}

	return execOut.Bytes(), nil
}

func (h *Helper) WaitForPodReady(namespace, name string, timeout time.Duration) error {
	log.Logf("Waiting for Pod to become ready %s/%s", namespace, name)

	err := wait.PollImmediate(time.Second/2, timeout, func() (bool, error) {
		pod, err := h.KubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		var ready bool
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady &&
				cond.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}

		if !ready {
			log.Logf("helper: pod not ready %s/%s: %v",
				pod.Namespace, pod.Name, pod.Status.Conditions)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		h.Kubectl(namespace).DescribeResource("pod", name)
		return err
	}

	return nil
}

func (h *Helper) WaitForPodDeletion(namespace, name string, timeout time.Duration) error {
	log.Logf("Waiting for Pod to be deleted %s/%s", namespace, name)
	err := wait.PollImmediate(time.Second/2, timeout, func() (bool, error) {
		pod, err := h.KubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) {
			return true, nil
		}

		if err != nil {
			return false, err
		}

		log.Logf("helper: pod not deleted %s/%s: %v",
			pod.Namespace, pod.Name, pod.Status.Conditions)

		return false, nil
	})
	if err != nil {
		h.Kubectl(namespace).DescribeResource("pod", name)
		return err
	}

	return nil
}
