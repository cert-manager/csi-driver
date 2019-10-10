package util

import (
	"encoding/json"
	"path/filepath"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func WriteMetaDataFile(vol *v1alpha1.MetaData) error {
	b, err := json.Marshal(vol)
	if err != nil {
		return err
	}

	metaPath := filepath.Join(vol.Path, v1alpha1.MetaDataFileName)
	return WriteFile(metaPath, b, 0600)
}
