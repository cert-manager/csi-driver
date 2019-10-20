package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	corev1 "k8s.io/api/core/v1"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/util"
)

func (h *Helper) MetaDataCertificateKeyExistInLocalPath(cr *cmapi.CertificateRequest,
	pod *corev1.Pod, attr map[string]string, dataDir string) error {
	volID := util.BuildVolumeName(pod.Name, string(pod.UID))
	name := util.BuildVolumeName(pod.Name, volID)
	dirPath := filepath.Join(dataDir, name)

	if err := h.matchDirPerm(dirPath, true, 0700); err != nil {
		return err
	}

	metaPath := filepath.Join(dirPath, "metadata.json")
	if err := h.matchDirPerm(metaPath, false, 0600); err != nil {
		return err
	}

	metaDataData, err := ioutil.ReadFile(metaPath)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", metaDataData)

	// TODO: check that both attributes and metadata file match <<<<

	//metaData := new(csiapi.MetaData)
	//if err := json.Unmarshal(metaDataData, metaData); err != nil {
	//	return fmt.Errorf("failed to unmarshal metadata file for %q: %s", metaPath, err)
	//}

	dataDirPath := filepath.Join(dirPath, "data")
	if err := h.matchDirPerm(dataDirPath, true, 0700); err != nil {
		return err
	}

	certPath, ok := attr[csiapi.CertFileKey]
	if !ok {
		certPath = "crt.pem"
	}
	certPath = filepath.Join(dataDirPath, certPath)
	if err := h.matchDirPerm(certPath, false, 0600); err != nil {
		return err
	}

	keyPath, ok := attr[csiapi.KeyFileKey]
	if !ok {
		keyPath = "key.pem"
	}
	keyPath = filepath.Join(dataDirPath, keyPath)
	if err := h.matchDirPerm(keyPath, false, 0600); err != nil {
		return err
	}

	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}
	keyData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	return h.certKeyMatch(cr, certData, keyData)
}

func (h *Helper) matchDirPerm(path string, isDir bool, perm os.FileMode) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get stat of %q: %s", path, err)
	}

	if stat.IsDir() != isDir {
		return fmt.Errorf("expected is directory %q == %t, got %t", path, stat.IsDir(), isDir)
	}
	if stat.Mode() != perm {
		return fmt.Errorf("expected %q to have permissions %s but got %s",
			path, perm, stat.Mode())
	}

	return nil
}
