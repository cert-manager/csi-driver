package driver

import (
	"errors"
	"testing"
)

func TestValidateCertManagerAttributes(t *testing.T) {
	type vaT struct {
		attr     map[string]string
		expError error
	}

	tests := map[string]vaT{
		"attributes with no issuer name or common name/dns names should error": {
			attr: map[string]string{},
			expError: errors.New(
				"csi.certmanager.k8s.io/issuer-name field required, both csi.certmanager.k8s.io/common-name and csi.certmanager.k8s.io/dns-names may not be empty"),
		},
		"attributes with no issuer name but common name": {
			attr: map[string]string{
				issuerNameKey: "test-issuer",
			},
			expError: errors.New(
				"both csi.certmanager.k8s.io/common-name and csi.certmanager.k8s.io/dns-names may not be empty"),
		},
		"attributes with no issuer name but DNS names should error": {
			attr: map[string]string{
				dnsNamesKey: "foo.bar.com,car.bar.com",
			},
			expError: errors.New(
				"csi.certmanager.k8s.io/issuer-name field required"),
		},
		"attributes with common name but no issuer name or DNS names should error": {
			attr: map[string]string{
				commonNameKey: "foo.bar",
			},
			expError: errors.New(
				"csi.certmanager.k8s.io/issuer-name field required"),
		},
		"valid attributes with common name should return no error": {
			attr: map[string]string{
				issuerNameKey: "test-issuer",
				commonNameKey: "foo.bar",
			},
			expError: nil,
		},
		"valid attributes with DNS names should return no error": {
			attr: map[string]string{
				issuerNameKey: "test-issuer",
				dnsNamesKey:   "foo.bar.com,car.bar.com",
			},
			expError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := new(certmanager)
			err := c.validateAttributes(test.attr)
			if test.expError == nil {
				if err != nil {
					t.Errorf("unexpected error, got=%s",
						err)
				}

				return
			}

			if err == nil || err.Error() != test.expError.Error() {
				t.Errorf("unexpected error, exp=%s got=%s",
					test.expError, err)
			}
		})
	}
}
