package kind

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

const (
	certManagerManifestPath = "https://github.com/jetstack/cert-manager/releases/download/v%s/cert-manager.yaml"
)

func (k *Kind) DeployCertManager(version string) error {
	log.Infof("kind: deploying cert-manager version %q", version)

	path := fmt.Sprintf(certManagerManifestPath, version)
	if err := k.kubectlApplyF(path); err != nil {
		return err
	}

	if err := k.waitForPodsReady(
		"cert-manager", "app.kubernetes.io/instance=cert-manager"); err != nil {
		return err
	}

	return nil
}

func (k *Kind) ensureKubectl() error {
	binPath := filepath.Join(k.rootPath, "bin")
	kubectlPath := filepath.Join(binPath, "kubectl")
	log.Debugf("kind: ensuring kubectl is present at %q", kubectlPath)

	err := os.MkdirAll(binPath, 0744)
	if err != nil {
		return err
	}

	s, err := os.Stat(kubectlPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		urlPath := fmt.Sprintf(kubectlURL, runtime.GOOS)
		log.Infof("kind: downloading kubectl from %q", urlPath)
		if err := downloadFile(kubectlPath, urlPath); err != nil {
			return err
		}

	} else {
		if s.IsDir() {
			return fmt.Errorf("kubectl filepath is directory %q", kubectlPath)
		}
	}

	if err := os.Chmod(kubectlPath, 0744); err != nil {
		return err
	}

	f, err := os.Open(kubectlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	fileHash := fmt.Sprintf("%x", h.Sum(nil))

	switch runtime.GOOS {
	case "linux":
		if fileHash != kubectlSHALinux {
			return fmt.Errorf("file hash did not match expected %s != (%q) %s",
				kubectlSHALinux, kubectlPath, fileHash)
		}

	case "darwin":
		if fileHash != kubectlSHADarwin {
			return fmt.Errorf("file hash did not match expected %s != (%q) %s",
				kubectlSHADarwin, kubectlPath, fileHash)
		}

	default:
		return fmt.Errorf("unsupported OS %q", runtime.GOOS)
	}

	return nil
}

func (k *Kind) kubectlApplyF(manifestPath string) error {
	log.Infof("kind: applying manifests %s", manifestPath)

	if err := k.ensureKubectl(); err != nil {
		return err
	}

	kubectlPath := filepath.Join(k.rootPath, "bin", "kubectl")
	err := k.runCmd(kubectlPath,
		"--kubeconfig="+k.ctx.KubeConfigPath(),
		"apply",
		"-f",
		manifestPath)
	if err != nil {
		return err
	}

	return nil
}

func (k *Kind) runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env,
		"GO111MODULE=on", "CGO_ENABLED=0", "HOME="+os.Getenv("HOME"))

	return cmd.Run()
}
