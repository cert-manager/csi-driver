package validation

import (
	"errors"
	"strings"
	"testing"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func TestValidateCertManagerAttributes(t *testing.T) {
	type vaT struct {
		attr     map[string]string
		expError error
	}

	tests := map[string]vaT{
		"attributes with no issuer name but DNS names should error": {
			attr: map[string]string{
				v1alpha1.DNSNamesKey: "foo.bar.com,car.bar.com",
			},
			expError: errors.New(
				"csi.cert-manager.io/issuer-name field required"),
		},
		"attributes with common name but no issuer name or DNS names should error": {
			attr: map[string]string{
				v1alpha1.CommonNameKey: "foo.bar",
			},
			expError: errors.New(
				"csi.cert-manager.io/issuer-name field required"),
		},
		"valid attributes with common name should return no error": {
			attr: map[string]string{
				v1alpha1.IssuerNameKey: "test-issuer",
				v1alpha1.CommonNameKey: "foo.bar",
			},
			expError: nil,
		},
		"valid attributes with DNS names should return no error": {
			attr: map[string]string{
				v1alpha1.IssuerNameKey: "test-issuer",
				v1alpha1.DNSNamesKey:   "foo.bar.com,car.bar.com",
			},
			expError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateAttributes(test.attr)
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

func TestFilePathBreakOut(t *testing.T) {
	for name, test := range map[string]struct {
		s       string
		expErrs string
	}{
		"normal filepath should not errors": {
			"foo/bar",
			"",
		},
		"no filepath shouldn't error": {
			"",
			"",
		},
		"single dot should not error": {
			"foo/./bar",
			"",
		},
		"two dots should error in middle": {
			"foo/../bar",
			"T filepaths may not contain '..'",
		},
		"two dots should error": {
			"..",
			"T filepaths may not contain '..'",
		},
	} {
		t.Run(name, func(t *testing.T) {
			errs := filepathBreakout(test.s, "T", nil)

			if test.expErrs != strings.Join(errs, "") {
				t.Errorf("unexpected error returned, exp=%s got=%s",
					test.expErrs, errs)
			}
		})
	}
}

func TestDurationParse(t *testing.T) {
	for name, test := range map[string]struct {
		s       string
		expErrs string
	}{
		"no duration should not error": {
			"",
			"",
		},
		"a good duation should parse": {
			"30h",
			"",
		},
		"a bad duration should error": {
			"20days",
			"T must be a valid duration string: time: unknown unit days in duration 20days",
		},
	} {
		t.Run(name, func(t *testing.T) {
			errs := durationParse(test.s, "T", nil)

			if test.expErrs != strings.Join(errs, "") {
				t.Errorf("unexpected error returned, exp=%s got=%s",
					test.expErrs, errs)
			}
		})
	}
}

func TestBoolValue(t *testing.T) {
	for name, test := range map[string]struct {
		s       string
		expErrs string
	}{
		"no value should not error": {
			"",
			"",
		},
		"a 'true' value should not error": {
			"true",
			"",
		},
		"a 'false' value should not error": {
			"false",
			"",
		},
		"a camel case True should error": {
			"True",
			"T may only be set to 'true' for 'false'",
		},
		"an uppercase FALSE should error": {
			"FALSE",
			"T may only be set to 'true' for 'false'",
		},
		"a bad string should error": {
			"foo",
			"T may only be set to 'true' for 'false'",
		},
	} {
		t.Run(name, func(t *testing.T) {
			errs := boolValue(test.s, "T", nil)

			if test.expErrs != strings.Join(errs, "") {
				t.Errorf("unexpected error returned, exp=%s got=%s",
					test.expErrs, errs)
			}
		})
	}
}
