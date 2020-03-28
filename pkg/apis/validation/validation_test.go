/*
Copyright 2019 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	"errors"
	"strings"
	"testing"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func TestValidateCertManagerAttributes(t *testing.T) {
	type vaT struct {
		attr     map[string]string
		expError error
	}

	tests := map[string]vaT{
		"attributes with no issuer name but DNS names should error": {
			attr: map[string]string{
				csiapi.DNSNamesKey: "foo.bar.com,car.bar.com",
			},
			expError: errors.New(
				"csi.cert-manager.io/issuer-name field required"),
		},
		"attributes with common name but no issuer name or DNS names should error": {
			attr: map[string]string{
				csiapi.CommonNameKey: "foo.bar",
			},
			expError: errors.New(
				"csi.cert-manager.io/issuer-name field required"),
		},
		"valid attributes with common name should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey: "test-issuer",
				csiapi.CommonNameKey: "foo.bar",
			},
			expError: nil,
		},
		"valid attributes with DNS names should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey: "test-issuer",
				csiapi.DNSNamesKey:   "foo.bar.com,car.bar.com",
			},
			expError: nil,
		},
		"valid attributes with one key usages should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey: "test-issuer",
				csiapi.DNSNamesKey:   "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:  "client auth",
			},
			expError: nil,
		},
		"valid attributes with key usages extended key usages should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey: "test-issuer",
				csiapi.DNSNamesKey:   "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:  "code signing,email protection,s/mime,ipsec end system",
			},
			expError: nil,
		},
		"attributes with wrong key usages should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey: "test-issuer",
				csiapi.DNSNamesKey:   "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:  "foo,bar,hello world",
			},
			expError: errors.New(
				`"foo" is not a valid key usage, "bar" is not a valid key usage, "hello world" is not a valid key usage`,
			),
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
