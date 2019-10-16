package kind

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	e2eImage = "cert-manager-csi-e2e:%s"
)

func (k *Kind) DeployCSIDriver(version string) error {
	log.Info("kind: building in tree CSI driver")

	csiPath := filepath.Join(k.rootPath, "./bin/cert-manager-csi")
	cmdPath := filepath.Join(k.rootPath, "./cmd/.")

	err := k.runCmd("go", "build", "-v", "-o", csiPath, cmdPath)
	if err != nil {
		return err
	}

	image := fmt.Sprintf(e2eImage, version)
	log.Infof("kind: building CSI driver image %q", image)
	err = k.runCmd("docker", "build", "-t", image, k.rootPath)
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), "cert-manager-csi-e2e")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	imageArchive := filepath.Join(tmpDir, "cert-manager-csi-e2e.tar")
	log.Infof("kind: saving image to archive %q", imageArchive)
	err = k.runCmd("docker", "save", "--output="+imageArchive, image)
	if err != nil {
		return err
	}

	nodes, err := k.ctx.ListNodes()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(imageArchive)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		log.Infof("kind: loading image %q to node %q", image, node.Name())
		r := bytes.NewBuffer(b)
		if err := node.LoadImageArchive(r); err != nil {
			return err
		}
	}

	manifests, err := ioutil.ReadFile(
		filepath.Join(k.rootPath, "deploy", "cert-manager-csi-driver.yaml"))
	if err != nil {
		return err
	}

	manifests = bytes.ReplaceAll(manifests,
		[]byte("gcr.io/jetstack-josh/cert-manager-csi:v0.1.0-alpha.1"), []byte(image))

	deployPath := filepath.Join(tmpDir, "cert-manager-csi-driver.yaml")
	if err := ioutil.WriteFile(deployPath, manifests, 0644); err != nil {
		return err
	}

	if err := k.kubectlApplyF(deployPath); err != nil {
		return err
	}

	if err := k.waitForPodsReady("cert-manager", "app=cert-manager-csi"); err != nil {
		return err
	}

	return nil
}
