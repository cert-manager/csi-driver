package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/jetstack/cert-manager-csi/test/kind"
)

const (
	defaultNodeImage = "1.16.1"
)

func main() {
	nodeImage := os.Getenv("CERT_MANAGER_CSI_K8S_VERSION")
	if nodeImage == "" {
		nodeImage = defaultNodeImage
	}

	nodeImage = fmt.Sprintf("kindest/node:v%s", nodeImage)

	k, err := kind.New(nodeImage, 1, 3)
	if err != nil {
		log.Fatal(err)
	}

	if err := k.DeplotCertManager("0.11.0"); err != nil {
		log.Errorf(err)
	}

	if err := k.Stop(); err != nil {
		log.Fatal(err)
	}
}
