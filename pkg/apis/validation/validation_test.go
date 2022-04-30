/*
Copyright 2021 The cert-manager Authors.

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
	"testing"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation/field"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

func Test_ValidateAttributes(t *testing.T) {
	type vaT struct {
		attr     map[string]string
		expError error
	}

	tests := map[string]struct {
		attr   map[string]string
		expErr field.ErrorList
	}{
		"attributes with no issuer name but DNS names should error": {
			attr: map[string]string{
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: field.ErrorList{
				field.Required(field.NewPath("volumeAttributes", "csi.cert-manager.io/issuer-name"), "issuer-name is a required field"),
			},
		},
		"attributes with common name but no issuer name or DNS names should error": {
			attr: map[string]string{
				csiapi.CommonNameKey:  "foo.bar",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: field.ErrorList{
				field.Required(field.NewPath("volumeAttributes", "csi.cert-manager.io/issuer-name"), "issuer-name is a required field"),
			},
		},
		"valid attributes with common name should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.CommonNameKey:  "foo.bar",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"valid attributes with DNS names should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"valid attributes with one key usages should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:   "client auth",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"valid attributes with key usages extended key usages should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:   "code signing  ,      email protection,    s/mime,ipsec end system",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"attributes with wrong key usages should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:   "foo,bar,hello world",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: field.ErrorList{
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/key-usages"), "foo", "not a valid key usage"),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/key-usages"), "bar", "not a valid key usage"),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/key-usages"), "hello world", "not a valid key usage"),
			},
		},
		"bad duration and a bad bool value should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:   "test-issuer",
				csiapi.DurationKey:     "bad-duration",
				csiapi.ReusePrivateKey: "FOO",
				csiapi.KeyEncodingKey:  "PKCS1",
			},
			expErr: field.ErrorList{
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/duration"), "bad-duration", `must be a valid duration string: time: invalid duration "bad-duration"`),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/reuse-private-key"), "FOO", `may only accept values of "true" or "false"`),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expErr, ValidateAttributes(test.attr))
		})
	}
}

func Test_filepathBreakOut(t *testing.T) {
	for name, test := range map[string]struct {
		s      string
		expErr field.ErrorList
	}{
		"normal filepath should not errors": {
			s:      "foo/bar",
			expErr: nil,
		},
		"no filepath shouldn't error": {
			s:      "",
			expErr: nil,
		},
		"single dot should not error": {
			s:      "foo/./bar",
			expErr: nil,
		},
		"two dots should error in middle": {
			s:      "foo/../bar",
			expErr: field.ErrorList{field.Invalid(field.NewPath("my-path"), "foo/../bar", `filepaths may not contain ".."`)},
		},
		"two dots should error": {
			s:      "..",
			expErr: field.ErrorList{field.Invalid(field.NewPath("my-path"), "..", `filepaths may not contain ".."`)},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expErr, filepathBreakout(field.NewPath("my-path"), test.s))
		})
	}
}

func Test_durationParse(t *testing.T) {
	for name, test := range map[string]struct {
		s      string
		expErr field.ErrorList
	}{
		"no duration should not error": {
			s:      "",
			expErr: nil,
		},
		"a good duation should parse": {
			s:      "30h",
			expErr: nil,
		},
		"a bad duration should error": {
			s:      "20days",
			expErr: field.ErrorList{field.Invalid(field.NewPath("my-duration"), "20days", `must be a valid duration string: time: unknown unit "days" in duration "20days"`)},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expErr, durationParse(field.NewPath("my-duration"), test.s))
		})
	}
}

func Test_boolValue(t *testing.T) {
	for name, test := range map[string]struct {
		s      string
		expErr field.ErrorList
	}{
		"no value should not error": {
			s:      "",
			expErr: nil,
		},
		"a 'true' value should not error": {
			s:      "true",
			expErr: nil,
		},
		"a 'false' value should not error": {
			s:      "false",
			expErr: nil,
		},
		"a camel case True should error": {
			s:      "True",
			expErr: field.ErrorList{field.Invalid(field.NewPath("my-bool"), "True", `may only accept values of "true" or "false"`)},
		},
		"an uppercase FALSE should error": {
			s:      "FALSE",
			expErr: field.ErrorList{field.Invalid(field.NewPath("my-bool"), "FALSE", `may only accept values of "true" or "false"`)},
		},
		"a bad string should error": {
			s:      "foo",
			expErr: field.ErrorList{field.Invalid(field.NewPath("my-bool"), "foo", `may only accept values of "true" or "false"`)},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expErr, boolValue(field.NewPath("my-bool"), test.s))
		})
	}
}

func Test_keyEncodingValue(t *testing.T) {
	for name, test := range map[string]struct {
		s      string
		expErr field.ErrorList
	}{
		"PKCS1 should not error": {
			s:      "PKCS1",
			expErr: nil,
		},
		"PKCS8 should not error": {
			s:      "PKCS8",
			expErr: nil,
		},
		"an unknown value should error": {
			s:      "foo",
			expErr: field.ErrorList{field.NotSupported(field.NewPath("my-pkcs"), "foo", []string{string(cmapi.PKCS1), string(cmapi.PKCS8)})},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expErr, keyEncodingValue(field.NewPath("my-pkcs"), test.s))
		})
	}
}
