package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kind/pkg/cluster/nodes"

	csidefaults "github.com/jetstack/cert-manager-csi/pkg/apis/defaults"
	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/util"
)

func (h *Helper) MetaDataCertificateKeyExistInLocalPath(cr *cmapi.CertificateRequest,
	pod *corev1.Pod, attr map[string]string, podMountPath, dataDir string) error {
	volID := util.BuildVolumeID(string(pod.UID), podMountPath)
	volName := util.BuildVolumeName(pod.Name, volID)
	dirPath := filepath.Join(dataDir, volName)

	// set defaults and csi storage attrubutes from pod
	attr, err := csidefaults.SetDefaultAttributes(attr)
	if err != nil {
		return fmt.Errorf("failed to set default volume attributes: %s", err)
	}

	attr["csi.storage.k8s.io/ephemeral"] = "true"
	attr["csi.storage.k8s.io/pod.name"] = pod.Name
	attr["csi.storage.k8s.io/pod.namespace"] = pod.Namespace
	attr["csi.storage.k8s.io/pod.uid"] = string(pod.UID)
	attr["csi.storage.k8s.io/serviceAccount.name"] = pod.Spec.ServiceAccountName

	node, err := h.cfg.Environment.Node(pod.Spec.NodeName)
	if err != nil {
		return err
	}

	if err := h.matchFilePerm(node, dirPath, 700); err != nil {
		return err
	}

	metaPath := filepath.Join(dirPath, "metadata.json")
	if err := h.matchFilePerm(node, metaPath, 600); err != nil {
		return err
	}

	metaDataData, err := h.readFile(node, metaPath)
	if err != nil {
		return err
	}

	expMetaData := &csiapi.MetaData{
		ID:   volID,
		Name: volName,
		Size: 102400,
		Path: filepath.Join("/csi-data-dir", volName),
		TargetPath: fmt.Sprintf("/var/lib/kubelet/pods/%s/volumes/kubernetes.io~csi/%s/mount",
			string(pod.UID), podMountPath),
		Attributes: attr,
	}

	gotMetaData := new(csiapi.MetaData)
	if err := json.Unmarshal(metaDataData, gotMetaData); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %s", err)
	}

	if err := h.metaDataMatches(expMetaData, gotMetaData); err != nil {
		return fmt.Errorf("bad metadata at %q: %s", metaPath, err)
	}

	dataDirPath := filepath.Join(dirPath, "data")
	if err := h.matchFilePerm(node, dataDirPath, 744); err != nil {
		return err
	}

	certPath := attr[csiapi.CertFileKey]
	certPath = filepath.Join(dataDirPath, certPath)
	if err := h.matchFilePerm(node, certPath, 600); err != nil {
		return err
	}

	keyPath := attr[csiapi.KeyFileKey]
	keyPath = filepath.Join(dataDirPath, keyPath)
	if err := h.matchFilePerm(node, keyPath, 600); err != nil {
		return err
	}

	certData, err := h.readFile(node, certPath)
	if err != nil {
		return err
	}
	keyData, err := h.readFile(node, keyPath)
	if err != nil {
		return err
	}

	return h.certKeyMatch(cr, certData, keyData)
}

func (h *Helper) readFile(node *nodes.Node, path string) ([]byte, error) {
	// TODO (@joshvanl): use tar compression
	execOut, execErr := new(bytes.Buffer), new(bytes.Buffer)
	cmd := node.Command("cat", path)
	cmd.SetStdout(execOut)
	cmd.SetStderr(execErr)

	if err := cmd.Run(); err != nil {
		log.Errorf("helper: cat %q failed: %s", path, execErr.String())
		return nil, err
	}

	return execOut.Bytes(), nil
}

func (h *Helper) matchFilePerm(node *nodes.Node, path string, perm int) error {
	execOut, execErr := new(bytes.Buffer), new(bytes.Buffer)

	cmd := node.Command("stat", "-c", "\"%a\"", path)
	cmd.SetStdout(execOut)
	cmd.SetStderr(execErr)

	if err := cmd.Run(); err != nil {
		log.Errorf("helper: stat failed: %s", execErr.String())
		pathD := filepath.Dir(path)

		cmd := node.Command("ls", "-la", pathD)
		cmd.SetStdout(execOut)
		if lErr := cmd.Run(); lErr == nil {
			log.Infof("helper: ls -la %q: %s\n",
				pathD, execOut.String())
		}

		return fmt.Errorf("failed to get stat of file %q: %s",
			path, err)
	}

	uStr := strings.ReplaceAll(
		strings.TrimSpace(execOut.String()), `"`, "")
	u, err := strconv.ParseUint(uStr, 10, 32)
	if err != nil {
		return err
	}

	if uint32(u) != uint32(perm) {
		return fmt.Errorf("expected %q to have permissions %v but got %v",
			path, uint32(perm), uint32(u))
	}

	return nil
}

func (h *Helper) metaDataMatches(exp, got *csiapi.MetaData) error {
	var errs []string

	if exp.ID != got.ID {
		errs = append(errs, fmt.Sprintf("miss-match id, exp=%s got=%s",
			exp.ID, got.ID))
	}

	if exp.Name != got.Name {
		errs = append(errs, fmt.Sprintf("miss-match name, exp=%s got=%s",
			exp.Name, got.Name))
	}

	if exp.Path != got.Path {
		errs = append(errs, fmt.Sprintf("miss-match path, exp=%s got=%s",
			exp.Path, got.Path))
	}

	if exp.Size != got.Size {
		errs = append(errs, fmt.Sprintf("miss-match size, exp=%d got=%d",
			exp.Size, got.Size))
	}

	if exp.TargetPath != got.TargetPath {
		errs = append(errs, fmt.Sprintf("miss-match targetPath, exp=%s got=%s",
			exp.TargetPath, got.TargetPath))
	}

	if !reflect.DeepEqual(exp.Attributes, got.Attributes) {
		errs = append(errs, fmt.Sprintf("miss-match attributes, exp=%s got=%s",
			exp.Attributes, got.Attributes))
	}

	if len(errs) > 0 {
		return fmt.Errorf("unexpected metadata: %s",
			strings.Join(errs, ", "))
	}

	return nil
}
