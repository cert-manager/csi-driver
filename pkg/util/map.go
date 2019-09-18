package util

import (
	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

// TODO (@joshvanl): fix this...
func MapStringToAttributes(a map[string]string) v1alpha1.Attributes {
	b := make(v1alpha1.Attributes)

	for n, v := range a {
		b[v1alpha1.Attribute(n)] = v
	}

	return b
}
