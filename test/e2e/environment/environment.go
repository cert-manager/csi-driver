package environment

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"

	"github.com/jetstack/cert-manager-csi/test/kind"
)

const (
	defaultNodeImage          = "1.16.1"
	defaultCertManagerVersion = "0.11.0"
	defaultRootPath           = "../../../."
)

type Environment struct {
	kind *kind.Kind
}

func Create(masterNodes, workerNodes int) (*Environment, error) {
	nodeImage := os.Getenv("CERT_MANAGER_CSI_K8S_VERSION")
	if nodeImage == "" {
		nodeImage = defaultNodeImage
	}
	nodeImage = fmt.Sprintf("kindest/node:v%s", nodeImage)

	certManagerVersion := os.Getenv("CERT_MANAGER_CSI_CERT_MANAGER_VERSION")
	if certManagerVersion == "" {
		certManagerVersion = defaultCertManagerVersion
	}

	rootPath := os.Getenv("CERT_MANAGER_CSI_ROOT_PATH")
	if rootPath == "" {
		rootPath = defaultRootPath
	}

	rPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path %q: %s",
			rootPath, err)
	}

	k, err := kind.New(rPath, nodeImage, masterNodes, workerNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to create kind cluster: %s", err)
	}

	if err := k.DeployCertManager(certManagerVersion); err != nil {
		return nil, fmt.Errorf("failed to deploy cert-manager: %s", err)
	}

	if err := k.DeployCSIDriver("canary"); err != nil {
		return nil, fmt.Errorf("failed to deploy cert-manager-csi driver: %s", err)
	}

	return &Environment{
		kind: k,
	}, nil
}

func (e *Environment) Destory() error {
	if err := e.kind.Stop(); err != nil {
		return err
	}

	return nil
}

func (e *Environment) KubeClient() *kubernetes.Clientset {
	return e.kind.KubeClient()
}
