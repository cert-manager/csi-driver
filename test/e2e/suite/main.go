package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/jetstack/cert-manager-csi/test/kind"
)

const (
	defaultNodeImage          = "1.16.1"
	defaultCertManagerVersion = "0.11.0"
	defaultRootPath           = "../../../."
)

func main() {
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

	bPath, err := filepath.Abs(rootPath)
	if err != nil {
		log.Fatalf("failed to get absolute path %q: %s",
			rootPath, err)
	}

	k, err := kind.New(bPath, nodeImage, 1, 3)
	if err != nil {
		log.Fatal(err)
	}

	if err := k.DeployCertManager(certManagerVersion); err != nil {
		log.Errorf(err.Error())
	}

	if err := k.DeployCSIDriver("canary"); err != nil {
		log.Errorf(err.Error())
	}

	if err := k.Stop(); err != nil {
		log.Fatal(err)
	}
}
