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

package requestgen

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	cmpki "github.com/cert-manager/cert-manager/pkg/util/pki"
	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"
	"github.com/stretchr/testify/assert"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

func Test_RequestForMetadata(t *testing.T) {
	t.Parallel()

	baseMetadataWith := func(meta metadata.Metadata) metadata.Metadata {
		if meta.VolumeContext == nil {
			meta.VolumeContext = make(map[string]string)
		}
		for k, v := range baseMetadata().VolumeContext {
			meta.VolumeContext[k] = v
		}
		return meta
	}

	mustParseURI := func(t *testing.T, uri string) *url.URL {
		puri, err := url.ParseRequestURI(uri)
		assert.NoError(t, err)
		return puri
	}

	var literalSubject = "CN=my-pod.my-namespace.svc.cluster.local,OU=0:my-pod\\;1:my-namespace\\;2:my-region\\;4:unittest,O=foo.bar.com"
	rdnSequence, err := cmpki.UnmarshalSubjectStringToRDNSequence(literalSubject)
	if err != nil {
		assert.NoError(t, err)
	}

	rawLiteralSubject, err := cmpki.MarshalRDNSequenceToRawDERBytes(rdnSequence)
	if err != nil {
		assert.NoError(t, err)
	}

	tests := map[string]struct {
		meta       metadata.Metadata
		expRequest *manager.CertificateRequestBundle
		expErr     bool
	}{
		"a metadata with no contents should return error": {
			meta:       baseMetadataWith(metadata.Metadata{}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with just issuer name set should return all defaults": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
			}}),
			expRequest: &manager.CertificateRequestBundle{
				Request: new(x509.CertificateRequest),
				IsCA:    false,
				Usages: []cmapi.KeyUsage{
					cmapi.KeyUsage("digital signature"),
					cmapi.KeyUsage("key encipherment"),
				},
				Namespace: "my-namespace",
				IssuerRef: cmmeta.ObjectReference{
					Name:  "my-issuer",
					Kind:  "Issuer",
					Group: "cert-manager.io",
				},
				Duration:    time.Hour * 24 * 90,
				Annotations: make(map[string]string),
			},
			expErr: false,
		},
		"a metadata with a bad duration should return an error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/duration":    "foo",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with a bad common name variable should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/common-name": "$Foo",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with a bad dnsName variable should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/dns-names":   "foo,$Foo",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with a bad uriNames variable should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/uri-sans":    "foo,$Foo",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with bad ips set should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/ip-sans":     "foo",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with usages set should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/key-usages":  "usage-1,usage-2",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with all options set should be returned": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name":  "my-issuer",
				"csi.cert-manager.io/issuer-kind":  "FooBar",
				"csi.cert-manager.io/issuer-group": "joshvanl.com",
				"csi.cert-manager.io/duration":     "1h",
				"csi.cert-manager.io/common-name":  "${POD_NAME}.$POD_NAMESPACE",
				"csi.cert-manager.io/dns-names":    "${POD_NAME}-my-dns-$POD_NAMESPACE-$POD_UID,$POD_NAME,\n             ${POD_NAME}.${POD_NAMESPACE},$POD_NAME.$POD_NAMESPACE.svc,$POD_UID",
				"csi.cert-manager.io/uri-sans":     "spiffe://foo.bar/${POD_NAMESPACE}/${POD_NAME}/$POD_UID,file://foo-bar,     foo://${POD_UID}",
				"csi.cert-manager.io/ip-sans":      "1.2.3.4,\n \t 5.6.7.8",
				"csi.cert-manager.io/is-ca":        "true",
				"csi.cert-manager.io/key-usages":   "server auth,client auth",
			}}),
			expRequest: &manager.CertificateRequestBundle{
				Request: &x509.CertificateRequest{
					Subject: pkix.Name{CommonName: "my-pod-name.my-namespace"},
					DNSNames: []string{
						"my-pod-name-my-dns-my-namespace-my-pod-uuid",
						"my-pod-name", "my-pod-name.my-namespace",
						"my-pod-name.my-namespace.svc", "my-pod-uuid",
					},
					IPAddresses: []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")},
					URIs: []*url.URL{
						mustParseURI(t, "spiffe://foo.bar/my-namespace/my-pod-name/my-pod-uuid"),
						mustParseURI(t, "file://foo-bar"),
						mustParseURI(t, "foo://my-pod-uuid"),
					},
				},
				IsCA: true,
				Usages: []cmapi.KeyUsage{
					cmapi.KeyUsage("server auth"),
					cmapi.KeyUsage("client auth"),
				},
				Namespace: "my-namespace",
				IssuerRef: cmmeta.ObjectReference{
					Name:  "my-issuer",
					Kind:  "FooBar",
					Group: "joshvanl.com",
				},
				Duration:    time.Hour,
				Annotations: make(map[string]string),
			},
			expErr: false,
		},
		"a metadata with literal subject set should be returned": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				csiapi.IssuerNameKey:     "my-issuer",
				csiapi.LiteralSubjectKey: literalSubject,
			}}),
			expRequest: &manager.CertificateRequestBundle{
				Request:   &x509.CertificateRequest{RawSubject: rawLiteralSubject},
				Usages:    cmapi.DefaultKeyUsages(),
				Namespace: "my-namespace",
				IssuerRef: cmmeta.ObjectReference{
					Name:  "my-issuer",
					Kind:  "Issuer",
					Group: "cert-manager.io",
				},
				Duration:    cmapi.DefaultCertificateDuration,
				Annotations: make(map[string]string),
			},
			expErr: false,
		},
		"a metadata with incorrect literal subject set should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				csiapi.IssuerNameKey:     "my-issuer",
				csiapi.LiteralSubjectKey: strings.ReplaceAll(literalSubject, ";", "&"),
			}}),
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			request, err := RequestForMetadata(test.meta)
			assert.Equalf(t, test.expErr, err != nil, "%v", err)
			assert.Equal(t, test.expRequest, request)
		})
	}
}

func Test_parseDNSNames(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		csv         string
		expDNSNames []string
		expErr      error
	}{
		"an empty csv should return an empty list": {
			csv:         "",
			expDNSNames: nil,
			expErr:      nil,
		},
		"a csv with single entry should expect that entry returned": {
			csv:         "my-dns",
			expDNSNames: []string{"my-dns"},
			expErr:      nil,
		},
		"a csv with multiple entries should expect those entries returned": {
			csv:         "my-dns,my-second-dns,my-third-dns",
			expDNSNames: []string{"my-dns", "my-second-dns", "my-third-dns"},
			expErr:      nil,
		},
		"a single csv which uses should be substituted correctly": {
			csv:         `$POD_NAME-my-dns-$POD_NAMESPACE-${POD_UID}`,
			expDNSNames: []string{"my-pod-name-my-dns-my-namespace-my-pod-uuid"},
			expErr:      nil,
		},
		"if references a variable that doesn't exist, error": {
			csv:         `$POD_NAME-my-dns-${POD_NAMESPACE}-$POD_UID-$Foo`,
			expDNSNames: nil,
			expErr:      errors.New(`undefined variable "Foo", known variables: [POD_NAME POD_NAMESPACE POD_UID SERVICE_ACCOUNT_NAME]`),
		},
		"a csv containing multiple entries which uses should be substituted correctly": {
			csv:         `$POD_NAME-my-dns-${POD_NAMESPACE}-$POD_UID,$POD_NAME,$POD_NAME.$POD_NAMESPACE,$POD_NAME.$POD_NAMESPACE.svc,$POD_UID`,
			expDNSNames: []string{"my-pod-name-my-dns-my-namespace-my-pod-uuid", "my-pod-name", "my-pod-name.my-namespace", "my-pod-name.my-namespace.svc", "my-pod-uuid"},
			expErr:      nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseDNSNames(baseMetadata(), test.csv)
			assert.Equal(t, test.expErr, err)
			assert.ElementsMatch(t, test.expDNSNames, got)
		})
	}
}

func Test_URIs(t *testing.T) {
	t.Parallel()

	mustParse := func(t *testing.T, uri string) *url.URL {
		puri, err := url.Parse(uri)
		assert.NoError(t, err)
		return puri
	}

	tests := map[string]struct {
		csv     string
		expURIs func(t *testing.T) []*url.URL
		expErr  error
	}{
		"an empty csv should return an empty list": {
			csv:     "",
			expURIs: nil,
			expErr:  nil,
		},
		"a csv with single entry should expect that entry returned": {
			csv: "spiffe://foo.bar",
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "spiffe://foo.bar"),
				}
			},
			expErr: nil,
		},
		"a csv with multiple entries should expect those entries returned": {
			csv: "spiffe://foo.bar,file://hello-world/1234,foo://1234",
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "spiffe://foo.bar"),
					mustParse(t, "file://hello-world/1234"),
					mustParse(t, "foo://1234"),
				}
			},
			expErr: nil,
		},
		"a csv with a bad URI should return an error": {
			csv:     "spiffe://foo.bar,\n\nx\n,foo://foo\nbar,file://hello-world/1234,1234",
			expURIs: nil,
			expErr:  errors.New(`parse "x": invalid URI for request, parse "foo://foo\nbar": net/url: invalid control character in URL, parse "1234": invalid URI for request`),
		},
		"a single csv which uses variables should be substituted correctly": {
			csv: `foo://$POD_NAME-my-dns-${POD_NAMESPACE}-${POD_UID}`,
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "foo://my-pod-name-my-dns-my-namespace-my-pod-uuid"),
				}
			},
			expErr: nil,
		},
		"if variables references a variable that doesn't exist, error": {
			csv:     `$POD_NAME-my-dns-${POD_NAMESPACE}-$POD_UID-${Foo}`,
			expURIs: nil,
			expErr:  errors.New(`undefined variable "Foo", known variables: [POD_NAME POD_NAMESPACE POD_UID SERVICE_ACCOUNT_NAME]`),
		},
		"a csv containing multiple entries which uses variables should be substituted correctly": {
			csv: `spiffe://$POD_NAME-my-dns-${POD_NAMESPACE}-$POD_UID,spiffe://$POD_NAME,file://${POD_NAME}.$POD_NAMESPACE,foo://$POD_NAME.$POD_NAMESPACE.svc,spiffe://$POD_UID`,
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "spiffe://my-pod-name-my-dns-my-namespace-my-pod-uuid"),
					mustParse(t, "spiffe://my-pod-name"),
					mustParse(t, "file://my-pod-name.my-namespace"),
					mustParse(t, "foo://my-pod-name.my-namespace.svc"),
					mustParse(t, "spiffe://my-pod-uuid"),
				}
			},
			expErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseURIs(baseMetadata(), test.csv)
			assert.Equal(t, test.expErr, err)
			var expURIs []*url.URL
			if test.expURIs != nil {
				expURIs = test.expURIs(t)
			}
			assert.ElementsMatch(t, expURIs, got)
		})
	}
}

func Test_expand(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expOutput string
		expErr    error
	}{
		"if no input given, expect empty output": {
			input:     "",
			expOutput: "",
			expErr:    nil,
		},
		"if using variables, expect to be substituted": {
			input:     "foo-$POD_NAME-,,${POD_NAMESPACE},$POD_UID,${SERVICE_ACCOUNT_NAME}",
			expOutput: "foo-my-pod-name-,,my-namespace,my-pod-uuid,my-service-account",
			expErr:    nil,
		},
		"if reference a variable that does not exist, expect error": {
			input:     "foo-${POD_NAME}-,,$POD_NAMESPACE,${POD_UID}.$Foo${Bar}",
			expOutput: "",
			expErr:    errors.New(`undefined variable "Foo", undefined variable "Bar", known variables: [POD_NAME POD_NAMESPACE POD_UID SERVICE_ACCOUNT_NAME]`),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := expand(baseMetadata(), test.input)
			assert.Equal(t, test.expErr, err)
			assert.Equal(t, test.expOutput, output)
		})
	}
}

func baseMetadata() metadata.Metadata {
	return metadata.Metadata{
		VolumeContext: map[string]string{
			"csi.storage.k8s.io/pod.name":            "my-pod-name",
			"csi.storage.k8s.io/pod.namespace":       "my-namespace",
			"csi.storage.k8s.io/pod.uid":             "my-pod-uuid",
			"csi.storage.k8s.io/serviceAccount.name": "my-service-account",
			"csi.storage.k8s.io/ephemeral":           "true",
		},
	}
}
