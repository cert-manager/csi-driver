package util

import (
	"encoding/json"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func WriteMetaDataFile(vol *v1alpha1.MetaData) error {
	b, err := json.Marshal(vol)
	if err != nil {
		return err
	}

	return WriteFile(MountPath(vol), b, 0600)
}
