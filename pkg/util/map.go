package util

import (
	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func MapStringToAttributes(a interface{}) v1alpha1.Attributes {
	return a.(v1alpha1.Attributes)
}
