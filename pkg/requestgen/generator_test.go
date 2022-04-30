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
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
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
		puri, err := url.Parse(uri)
		assert.NoError(t, err)
		return puri
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
				Duration: time.Hour * 24 * 90,
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
		"a metadata with a bad common name template variable should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/common-name": "{{.Foo}}",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with a bad dnsName template variable should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/dns-names":   "foo,{{.Foo}}",
			}}),
			expRequest: nil,
			expErr:     true,
		},
		"a metadata with a bad uriNames template variable should error": {
			meta: baseMetadataWith(metadata.Metadata{VolumeContext: map[string]string{
				"csi.cert-manager.io/issuer-name": "my-issuer",
				"csi.cert-manager.io/uri-sans":    "foo,{{.Foo}}",
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
				"csi.cert-manager.io/common-name":  "{{.PodName}}.{{.PodNamespace}}",
				"csi.cert-manager.io/dns-names":    "{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}},{{.PodName}},{{.PodName}}.{{.PodNamespace}},{{.PodName}}.{{.PodNamespace}}.svc,{{.PodUID}}",
				"csi.cert-manager.io/uri-sans":     "spiffe://foo.bar/{{.PodNamespace}}/{{.PodName}}/{{.PodUID}},file://foo-bar,{{.PodUID}}",
				"csi.cert-manager.io/ip-sans":      "1.2.3.4,5.6.7.8",
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
						mustParseURI(t, "my-pod-uuid"),
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
				Duration: time.Hour,
			},
			expErr: false,
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
		expErr      bool
	}{
		"an empty csv should return an empty list": {
			csv:         "",
			expDNSNames: nil,
			expErr:      false,
		},
		"a csv with single entry should expect that entry returned": {
			csv:         "my-dns",
			expDNSNames: []string{"my-dns"},
			expErr:      false,
		},
		"a csv with multiple entries should expect those entries returned": {
			csv:         "my-dns,my-second-dns,my-third-dns",
			expDNSNames: []string{"my-dns", "my-second-dns", "my-third-dns"},
			expErr:      false,
		},
		"a single csv which uses templates should be substituted correctly": {
			csv:         `{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}}`,
			expDNSNames: []string{"my-pod-name-my-dns-my-namespace-my-pod-uuid"},
			expErr:      false,
		},
		"if template references a variable that doesn't exist, error": {
			csv:         `{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}}-{{.Foo}}`,
			expDNSNames: nil,
			expErr:      true,
		},
		"a csv containing multiple entries which uses templates should be substituted correctly": {
			csv:         `{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}},{{.PodName}},{{.PodName}}.{{.PodNamespace}},{{.PodName}}.{{.PodNamespace}}.svc,{{.PodUID}}`,
			expDNSNames: []string{"my-pod-name-my-dns-my-namespace-my-pod-uuid", "my-pod-name", "my-pod-name.my-namespace", "my-pod-name.my-namespace.svc", "my-pod-uuid"},
			expErr:      false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseDNSNames(baseMetadata(), test.csv)
			assert.Equalf(t, test.expErr, err != nil, "%v", err)
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
		expErr  bool
	}{
		"an empty csv should return an empty list": {
			csv:     "",
			expURIs: nil,
			expErr:  false,
		},
		"a csv with single entry should expect that entry returned": {
			csv: "spiffe://foo.bar",
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "spiffe://foo.bar"),
				}
			},
			expErr: false,
		},
		"a csv with multiple entries should expect those entries returned": {
			csv: "spiffe://foo.bar,file://hello-world/1234,1234",
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "spiffe://foo.bar"),
					mustParse(t, "file://hello-world/1234"),
					mustParse(t, "1234"),
				}
			},
			expErr: false,
		},
		"a csv with a bad URI should return an error": {
			csv:     "spiffe://foo.bar,\n,file://hello-world/1234,1234",
			expURIs: nil,
			expErr:  true,
		},
		"a single csv which uses templates should be substituted correctly": {
			csv: `{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}}`,
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "my-pod-name-my-dns-my-namespace-my-pod-uuid"),
				}
			},
			expErr: false,
		},
		"if template references a variable that doesn't exist, error": {
			csv:     `{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}}-{{.Foo}}`,
			expURIs: nil,
			expErr:  true,
		},
		"a csv containing multiple entries which uses templates should be substituted correctly": {
			csv: `spiffe://{{.PodName}}-my-dns-{{.PodNamespace}}-{{.PodUID}},spiffe://{{.PodName}},file://{{.PodName}}.{{.PodNamespace}},{{.PodName}}.{{.PodNamespace}}.svc,spiffe://{{.PodUID}}`,
			expURIs: func(t *testing.T) []*url.URL {
				return []*url.URL{
					mustParse(t, "spiffe://my-pod-name-my-dns-my-namespace-my-pod-uuid"),
					mustParse(t, "spiffe://my-pod-name"),
					mustParse(t, "file://my-pod-name.my-namespace"),
					mustParse(t, "my-pod-name.my-namespace.svc"),
					mustParse(t, "spiffe://my-pod-uuid"),
				}
			},
			expErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseURIs(baseMetadata(), test.csv)
			assert.Equalf(t, test.expErr, err != nil, "%v", err)
			var expURIs []*url.URL
			if test.expURIs != nil {
				expURIs = test.expURIs(t)
			}
			assert.ElementsMatch(t, expURIs, got)
		})
	}
}

func Test_executeTemplate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		expOutput string
		expErr    bool
	}{
		"if no input given, expect empty output": {
			input:     "",
			expOutput: "",
			expErr:    false,
		},
		"if using templates, expect to be substituted": {
			input:     "foo-{{.PodName}}-,,{{.PodNamespace}},{{.PodUID}}",
			expOutput: "foo-my-pod-name-,,my-namespace,my-pod-uuid",
			expErr:    false,
		},
		"if reference a template variable that does not exist, expect error": {
			input:     "foo-{{.PodName}}-,,{{.PodNamespace}},{{.PodUID}}.{{.Foo}}",
			expOutput: "",
			expErr:    true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := executeTemplate(baseMetadata(), test.input)
			assert.Equalf(t, test.expErr, err != nil, "%v", err)
			assert.Equal(t, test.expOutput, output)
		})
	}
}

func baseMetadata() metadata.Metadata {
	return metadata.Metadata{
		VolumeContext: map[string]string{
			"csi.storage.k8s.io/pod.name":      "my-pod-name",
			"csi.storage.k8s.io/pod.namespace": "my-namespace",
			"csi.storage.k8s.io/pod.uid":       "my-pod-uuid",
			"csi.storage.k8s.io/ephemeral":     "true",
		},
	}
}
