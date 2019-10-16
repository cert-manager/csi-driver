package kind

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	configv1alpha3 "sigs.k8s.io/kind/pkg/apis/config/v1alpha3"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
)

const (
	kubectlURL       = "https://storage.googleapis.com/kubernetes-release/release/v1.16.1/bin/%s/amd64/kubectl"
	kubectlSHALinux  = "69cfb3eeaa0b77cc4923428855acdfc9ca9786544eeaff9c21913be830869d29"
	kubectlSHADarwin = "9b45260bb16f251cf2bb4b4c5f90bc847ab752c9c936b784dc2bae892e10205a"
)

type Kind struct {
	rootPath string

	ctx    *cluster.Context
	client *kubernetes.Clientset
}

func New(rootPath, nodeImage string, masterNodes, workerNodes int) (*Kind, error) {
	log.Infof("kind: using k8s node image %q", nodeImage)

	k := &Kind{
		rootPath: rootPath,
		ctx:      cluster.NewContext("cert-manager-csi-e2e"),
	}

	conf := new(configv1alpha3.Cluster)
	configv1alpha3.SetDefaults_Cluster(conf)
	conf.Nodes = nil

	for i := 0; i < masterNodes; i++ {
		conf.Nodes = append(conf.Nodes,
			configv1alpha3.Node{
				Image: nodeImage,
				Role:  configv1alpha3.ControlPlaneRole,
			})
	}
	for i := 0; i < workerNodes; i++ {
		conf.Nodes = append(conf.Nodes,
			configv1alpha3.Node{
				Image: nodeImage,
				Role:  configv1alpha3.WorkerRole,
			})
	}

	conf.Networking.ServiceSubnet = "10.0.0.0/16"

	// create kind cluster
	log.Infof("kind: creating kind cluster %q", k.ctx.Name())
	if err := k.ctx.Create(create.WithV1Alpha3(conf)); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %s", err)
	}

	// generate rest config to kind cluster
	kubeconfig := k.ctx.KubeConfigPath()
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, k.errStop(fmt.Errorf("failed to build kind rest client: %s", err))
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, k.errStop(fmt.Errorf("failed to build kind kubernetes client: %s", err))
	}
	k.client = client

	if err := k.waitForNodesReady(); err != nil {
		return nil, k.errStop(fmt.Errorf("failed to wait for nodes to become ready: %s", err))
	}

	if err := k.waitForCoreDNSReady(); err != nil {
		return nil, k.errStop(fmt.Errorf("failed to wait for DNS pods to become ready: %s", err))
	}

	log.Infof("kind: cluster ready %q", k.ctx.Name())

	return k, nil
}

func (k *Kind) Stop() error {
	log.Infof("kind: stopping cluster %q", k.ctx.Name())
	if err := k.ctx.Delete(); err != nil {
		return fmt.Errorf("failed to delete kind cluster: %s", err)
	}

	log.Infof("kind: destroyed cluster %q", k.ctx.Name())

	return nil
}

func (k *Kind) errStop(err error) error {
	k.Stop()
	return err
}

func (k *Kind) waitForNodesReady() error {
	log.Infof("kind: waiting for all nodes to become ready...")

	return wait.PollImmediate(time.Second*5, time.Minute*10, func() (bool, error) {
		nodes, err := k.client.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		if len(nodes.Items) == 0 {
			log.Warn("kind: no nodes found - checking again...")
			return false, nil
		}

		var notReady []string
		for _, node := range nodes.Items {
			var ready bool
			for _, c := range node.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
					ready = true
					break
				}
			}

			if !ready {
				notReady = append(notReady, node.Name)
			}
		}

		if len(notReady) > 0 {
			log.Infof("kind: nodes not ready: %s",
				strings.Join(notReady, ", "))
			return false, nil
		}

		return true, nil
	})
}

func (k *Kind) waitForCoreDNSReady() error {
	log.Infof("kind: waiting for all DNS pods to become ready...")
	return k.waitForPodsReady("kube-system", "k8s-app=kube-dns")
}

func (k *Kind) waitForPodsReady(namespace, labelSelector string) error {
	return wait.PollImmediate(time.Second*5, time.Minute*10, func() (bool, error) {
		pods, err := k.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return false, err
		}

		if len(pods.Items) == 0 {
			log.Warnf("kind: no pods found in namespace %q with selector %q - checking again...",
				namespace, labelSelector)
			return false, nil
		}

		var notReady []string
		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				notReady = append(notReady, fmt.Sprintf("%s:%s (%s)",
					pod.Namespace, pod.Name, pod.Status.Phase))
			}
		}

		if len(notReady) > 0 {
			log.Infof("kind: pods not ready: %s",
				strings.Join(notReady, ", "))
			return false, nil
		}

		return true, nil
	})
}

func downloadFile(filepath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
