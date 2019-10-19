package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func BuildVolumeName(podName, volID string) string {
	return fmt.Sprintf("cert-manager-csi-%s-%s",
		podName, volID)
}

func BuildVolumeID(podUID, volSourceSpecName string) string {
	result := sha256.Sum256([]byte(fmt.Sprintf("%s%s", podUID, volSourceSpecName)))
	return fmt.Sprintf("csi-%x", result)
}

func WriteMetaDataFile(vol *csiapi.MetaData) error {
	b, err := json.Marshal(vol)
	if err != nil {
		return err
	}

	metaPath := filepath.Join(vol.Path, csiapi.MetaDataFileName)
	return WriteFile(metaPath, b, 0600)
}
