package helper

import (
	"bytes"
	"fmt"
	"path/filepath"
	"time"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func (h *Helper) CertificateKeyExistInPodPath(namespace, podName, containerName, mountPath string,
	cr *cmapi.CertificateRequest, attr map[string]string) error {

	certPath, ok := attr[csiapi.CertFileKey]
	if !ok {
		certPath = "crt.pem"
	}
	certPath = filepath.Join(mountPath, certPath)

	keyPath, ok := attr[csiapi.KeyFileKey]
	if !ok {
		keyPath = "key.pem"
	}
	keyPath = filepath.Join(mountPath, keyPath)

	restClient, err := rest.RESTClientFor(h.RestConfig)
	if err != nil {
		return fmt.Errorf("failed to build rest client form rest config: %s", err)
	}

	// TODO (@joshvanl): use tar compression
	req := restClient.Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command: []string{
				"cat", certPath,
			},
			Stdin:  false,
			Stdout: true,
			Stderr: true,
			TTY:    false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(h.RestConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create SPDY executor: %s", err)
	}
	execOut := &bytes.Buffer{}
	execErr := &bytes.Buffer{}

	sopt := remotecommand.StreamOptions{
		Stdout: execOut,
		Stderr: execErr,
		Tty:    false,
	}

	err = exec.Stream(sopt)
	if err != nil {
		fmt.Errorf("failed to execute stream command: %s", err)
	}

	return nil
}

func (h *Helper) WaitForPodReady(namespace, name string) error {
	err := wait.PollImmediate(time.Second*5, time.Minute, func() (bool, error) {
		pod, err := h.KubeClient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if pod.Status.Phase != corev1.PodRunning {
			log.Infof("helper: pod not ready %s:%s %v",
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
