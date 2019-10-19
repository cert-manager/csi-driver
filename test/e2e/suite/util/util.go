package util

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/cert-manager-csi/pkg/apis"
)

func ConstructCSIVolume(name string, attributes map[string]string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			CSI: &corev1.CSIVolumeSource{
				Driver:           apis.GroupName,
				VolumeAttributes: attributes,
			},
		},
	}
}
