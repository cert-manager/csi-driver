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

	var literalSubject = "CN=${POD_NAME}.${POD_NAMESPACE}.svc.cluster.local,OU=0:${POD_NAME}\\;1:${POD_NAMESPACE}\\;2:my-region\\;4:unittest,O=foo.bar.com"

	tests := map[string]struct {
		attr   map[string]string
		expErr field.ErrorList
	}{
		"attributes with no issuer name but DNS names should error": {
			attr: map[string]string{
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: field.ErrorList{
				field.Required(field.NewPath("volumeAttributes", "csi.cert-manager.io/issuer-name"), "issuer-name is a required field"),
			},
		},
		"attributes with common name but no issuer name or DNS names should error": {
			attr: map[string]string{
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
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
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
				csiapi.CommonNameKey:  "foo.bar",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"valid attributes with DNS names should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"valid attributes with one key usages should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:   "client auth",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"valid attributes with key usages extended key usages should return no error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.KeyUsagesKey:   "code signing  ,      email protection,    s/mime,ipsec end system",
				csiapi.KeyEncodingKey: "PKCS1",
			},
			expErr: nil,
		},
		"attributes with wrong key usages should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.CAFileKey:      "ca.crt",
				csiapi.CertFileKey:    "crt.tls",
				csiapi.KeyFileKey:     "key.tls",
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
				csiapi.CAFileKey:       "ca.crt",
				csiapi.CertFileKey:     "crt.tls",
				csiapi.KeyFileKey:      "key.tls",
				csiapi.DurationKey:     "bad-duration",
				csiapi.ReusePrivateKey: "FOO",
				csiapi.KeyEncodingKey:  "PKCS1",
			},
			expErr: field.ErrorList{
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/duration"), "bad-duration", `must be a valid duration string: time: invalid duration "bad-duration"`),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/reuse-private-key"), "FOO", `may only accept values of "true" or "false"`),
			},
		},
		"invalid PKCS12 options should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:             "test-issuer",
				csiapi.KeyEncodingKey:            "PKCS1",
				csiapi.CAFileKey:                 "ca.crt",
				csiapi.CertFileKey:               "crt.tls",
				csiapi.KeyFileKey:                "key.tls",
				csiapi.KeyStorePKCS12FileKey:     "../crt.p12",
				csiapi.KeyStorePKCS12PasswordKey: "password",
			},
			expErr: field.ErrorList{
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/pkcs12-filename"), "../crt.p12",
					`filename must not start with '..'`),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/pkcs12-filename"), "../crt.p12",
					`filename must not include '/'`),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/pkcs12-filename"), "../crt.p12",
					"cannot use attribute without \"csi.cert-manager.io/pkcs12-enable\" set to \"true\" or \"false\""),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/pkcs12-password"), "password",
					"cannot use attribute without \"csi.cert-manager.io/pkcs12-enable\" set to \"true\" or \"false\""),
			},
		},
		"setting output filenames which are duplicated should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:             "test-issuer",
				csiapi.KeyEncodingKey:            "PKCS1",
				csiapi.CAFileKey:                 "ca.crt",
				csiapi.CertFileKey:               "crt.tls",
				csiapi.KeyFileKey:                "ca.crt",
				csiapi.KeyStorePKCS12FileKey:     "crt.tls",
				csiapi.KeyStorePKCS12EnableKey:   "true",
				csiapi.KeyStorePKCS12PasswordKey: "password",
			},
			expErr: field.ErrorList{
				field.Duplicate(field.NewPath("volumeAttributes", "csi.cert-manager.io/ca-file"), "ca.crt"),
				field.Duplicate(field.NewPath("volumeAttributes", "csi.cert-manager.io/certificate-file"), "crt.tls"),
				field.Duplicate(field.NewPath("volumeAttributes", "csi.cert-manager.io/pkcs12-filename"), "crt.tls"),
				field.Duplicate(field.NewPath("volumeAttributes", "csi.cert-manager.io/privatekey-file"), "ca.crt"),
			},
		},
		"correct PKCS12 options should not error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:             "test-issuer",
				csiapi.KeyEncodingKey:            "PKCS1",
				csiapi.CAFileKey:                 "ca.crt",
				csiapi.CertFileKey:               "crt.tls",
				csiapi.KeyFileKey:                "key.tls",
				csiapi.KeyStorePKCS12EnableKey:   "true",
				csiapi.KeyStorePKCS12FileKey:     "crt.p12",
				csiapi.KeyStorePKCS12PasswordKey: "password",
			},
			expErr: nil,
		},
		"bad filenames for the certificate, key, and ca files should error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:  "test-issuer",
				csiapi.DNSNamesKey:    "foo.bar.com,car.bar.com",
				csiapi.CAFileKey:      "../foo/../bar",
				csiapi.KeyEncodingKey: "PKCS8",
				csiapi.CertFileKey:    "foofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoo",
				csiapi.KeyFileKey:     "/foobar",
			},
			expErr: field.ErrorList{
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/ca-file"), "../foo/../bar", "filename must not start with '..'"),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/ca-file"), "../foo/../bar", "filename must not include '/'"),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/certificate-file"), "foofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoo", "filename must be no longer than 255 characters"),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/privatekey-file"), "/foobar", "filename must not be an absolute path"),
				field.Invalid(field.NewPath("volumeAttributes", "csi.cert-manager.io/privatekey-file"), "/foobar", "filename must not include '/'"),
			},
		},
		"correct literal-subject should not error": {
			attr: map[string]string{
				csiapi.IssuerNameKey:     "test-issuer",
				csiapi.LiteralSubjectKey: literalSubject,
				csiapi.CAFileKey:         "ca.crt",
				csiapi.CertFileKey:       "crt.tls",
				csiapi.KeyFileKey:        "key.tls",
				csiapi.DNSNamesKey:       "foo.bar.com",
				csiapi.KeyEncodingKey:    "PKCS8",
			},
			expErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.EqualValues(t, test.expErr, ValidateAttributes(test.attr))
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

func Test_PKCS12Values(t *testing.T) {
	basePath := field.NewPath("root")

	tests := map[string]struct {
		attr   map[string]string
		expErr field.ErrorList
	}{
		"if no attributes, expect no error": {
			attr:   map[string]string{},
			expErr: nil,
		},
		"if key and password is defined, but enabled is not defined, expect error": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-filename": "my-file",
				"csi.cert-manager.io/pkcs12-password": "password",
			},
			expErr: field.ErrorList{
				field.Invalid(basePath.Child("csi.cert-manager.io/pkcs12-filename"), "my-file",
					"cannot use attribute without \"csi.cert-manager.io/pkcs12-enable\" set to \"true\" or \"false\""),
				field.Invalid(basePath.Child("csi.cert-manager.io/pkcs12-password"), "password",
					"cannot use attribute without \"csi.cert-manager.io/pkcs12-enable\" set to \"true\" or \"false\""),
			},
		},

		"if key and password is defined, and enabled is defined as false, expect no error": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "false",
				"csi.cert-manager.io/pkcs12-filename": "my-file",
				"csi.cert-manager.io/pkcs12-password": "password",
			},
			expErr: nil,
		},
		"if key and password is defined, but enabled is defined as foo, expect error": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "foo",
				"csi.cert-manager.io/pkcs12-filename": "my-file",
				"csi.cert-manager.io/pkcs12-password": "password",
			},
			expErr: field.ErrorList{
				field.NotSupported(basePath.Child("csi.cert-manager.io/pkcs12-enable"), "foo", []string{"true", "false"}),
			},
		},
		"if key and password is not defined, and enabled is defined as true, expect error": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-enable": "true",
			},
			expErr: field.ErrorList{
				field.Required(basePath.Child("csi.cert-manager.io/pkcs12-filename"), "required attribute when PKCS12 KeyStore is enabled"),
				field.Required(basePath.Child("csi.cert-manager.io/pkcs12-password"), "required attribute when PKCS12 KeyStore is enabled"),
			},
		},
		"if key and password is defined as empty string, and enabled is defined as true, expect error": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-filename": "",
				"csi.cert-manager.io/pkcs12-password": "",
			},
			expErr: field.ErrorList{
				field.Required(basePath.Child("csi.cert-manager.io/pkcs12-filename"), "required attribute when PKCS12 KeyStore is enabled"),
				field.Required(basePath.Child("csi.cert-manager.io/pkcs12-password"), "required attribute when PKCS12 KeyStore is enabled"),
			},
		},
		"if key and password is defined, and enabled is defined as true, expect no error": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-filename": "my-file",
				"csi.cert-manager.io/pkcs12-password": "password",
			},
			expErr: nil,
		},
		"if key and password is defined, and enabled is defined as true, expect no error foo": {
			attr: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "true",
				"csi.cert-manager.io/pkcs12-filename": "/my-file",
				"csi.cert-manager.io/pkcs12-password": "password",
			},
			expErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.EqualValues(t, test.expErr, pkcs12Values(basePath, test.attr))
		})
	}
}

func Test_filename(t *testing.T) {
	basePath := field.NewPath("root")

	tests := map[string]struct {
		filename string
		expErr   field.ErrorList
	}{
		"a filename which is absolute should error": {
			filename: "/foo/bar",
			expErr: field.ErrorList{
				field.Invalid(basePath, "/foo/bar", "filename must not be an absolute path"),
				field.Invalid(basePath, "/foo/bar", "filename must not include '/'"),
			},
		},
		"a filename which has a prefix of '..' should error": {
			filename: "...foobar",
			expErr: field.ErrorList{
				field.Invalid(basePath, "...foobar", "filename must not start with '..'"),
			},
		},
		"a filename which includes '/' should error": {
			filename: "foo/bar",
			expErr: field.ErrorList{
				field.Invalid(basePath, "foo/bar", "filename must not include '/'"),
			},
		},
		"a filename which includes ' ' at the beginning should error": {
			filename: "  foo",
			expErr: field.ErrorList{
				field.Invalid(basePath, "  foo", "filename must not include leading or trailing spaces"),
			},
		},
		"a filename which includes ' ' at the end should error": {
			filename: "foo  ",
			expErr: field.ErrorList{
				field.Invalid(basePath, "foo  ", "filename must not include leading or trailing spaces"),
			},
		},
		"a filename which includes ' ' at the end or beginning should error": {
			filename: " foo ",
			expErr: field.ErrorList{
				field.Invalid(basePath, " foo ", "filename must not include leading or trailing spaces"),
			},
		},
		"a filename which is too long should error": {
			filename: "foofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoo",
			expErr: field.ErrorList{
				field.Invalid(basePath, "foofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoofoo", "filename must be no longer than 255 characters"),
			},
		},
		"a valid filename should not error": {
			filename: "foo.bar",
			expErr:   nil,
		},
		"a filename which includes '..' in the middle should not error": {
			filename: "foo...bar",
			expErr:   nil,
		},
		"a filename which includes '..' at the end should not error": {
			filename: "foobar...",
			expErr:   nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expErr, filename(basePath, test.filename))
		})
	}
}

func Test_uniqueFilePaths(t *testing.T) {
	basePath := field.NewPath("root")

	tests := map[string]struct {
		paths  map[string]string
		expErr field.ErrorList
	}{
		"if no paths, expect no error": {
			paths:  map[string]string{},
			expErr: nil,
		},
		"if all paths are unique, expect no error": {
			paths: map[string]string{
				"a": "1", "b": "2", "c": "3",
			},
			expErr: nil,
		},
		"if some paths have duplicates, expect error": {
			paths: map[string]string{
				"a": "1", "b": "2", "c": "2", "d": "4",
			},
			expErr: field.ErrorList{
				field.Duplicate(basePath.Child("b"), "2"),
				field.Duplicate(basePath.Child("c"), "2"),
			},
		},
		"if some other paths have duplicates, expect error": {
			paths: map[string]string{
				"a": "1", "b": "2", "c": "3", "d": "1",
			},
			expErr: field.ErrorList{
				field.Duplicate(basePath.Child("a"), "1"),
				field.Duplicate(basePath.Child("d"), "1"),
			},
		},
		"if all paths have duplicates, error": {
			paths: map[string]string{
				"a": "1", "b": "2", "c": "2", "d": "1",
			},
			expErr: field.ErrorList{
				field.Duplicate(basePath.Child("a"), "1"),
				field.Duplicate(basePath.Child("b"), "2"),
				field.Duplicate(basePath.Child("c"), "2"),
				field.Duplicate(basePath.Child("d"), "1"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.EqualValues(t, test.expErr, uniqueFilePaths(basePath, test.paths))
		})
	}
}
