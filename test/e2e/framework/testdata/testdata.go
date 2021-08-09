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

package testdata

import (
	"errors"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

type issuer struct {
	name, kind string
}

type TestData struct {
	r       *rand.Rand
	issuers []issuer
}

func New(seed int64, issuers, clusterIssuers []string) (*TestData, error) {
	if len(issuers) == 0 && len(clusterIssuers) == 0 {
		return nil, errors.New(
			"expecting at least a one issuer or ClusterIssuer")
	}

	var unIssuers []issuer
	for _, i := range issuers {
		unIssuers = append(unIssuers, issuer{i, "Issuer"})
	}
	for _, i := range clusterIssuers {
		unIssuers = append(unIssuers, issuer{i, "ClusterIssuer"})
	}

	return &TestData{
		r:       rand.New(rand.NewSource(seed)),
		issuers: unIssuers,
	}, nil
}

func (t *TestData) RandomVolumeAttributes() map[string]string {
	attr := make(map[string]string)

	issName, issKind, issGroup := t.Issuer()
	attr["csi.cert-manager.io/issuer-name"] = issName
	// cluster issuers have to have the kind set since "Issuer" is the default
	if issKind == "ClusterIssuer" {
		attr["csi.cert-manager.io/issuer-kind"] = issKind
	} else {
		t.maybeAddAttribute(attr, "issuer-kind", issKind)
	}

	for _, a := range []struct {
		k, v string
	}{
		{"issuer-group", issGroup},
		{"dns-names", t.DNSNames()},
		{"uri-sans", t.URISANs()},
		{"ip-sans", t.IPSANs()},
		{"ip-duration", t.Duration()},
		{"is-ca", t.IsCA()},
		{"common-name", t.CommonName()},
		{"certificate-file", filepath.Join(t.RandomDirPath(), t.RandomName()+".pem")},
		{"privatekey-file", filepath.Join(t.RandomDirPath(), t.RandomName()+".pem")},
	} {
		t.maybeAddAttribute(attr, a.k, a.v)
	}

	return attr
}

func (t *TestData) RandomDirPath() string {
	dirs := make([]string, (t.r.Int()%5)+1)
	for i := range dirs {
		dirs[i] = t.RandomName()
	}

	return filepath.Join(dirs...)
}

func (t *TestData) RandomName() string {
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[t.r.Intn(len(letterRunes))]
	}

	return string(b)
}

func (t *TestData) Int(n int) int {
	return t.r.Int() % n
}

func (t *TestData) maybeAddAttribute(attr map[string]string, k, v string) {
	if t.r.Int()%4 == 0 {
		return
	}

	attr["csi.cert-manager.io/"+k] = v
}

func (t *TestData) Issuer() (name, kind, group string) {
	n := t.r.Int() % len(t.issuers)
	name, kind = t.issuers[n].name, t.issuers[n].kind

	if kind == "Issuer" {
		if t.r.Int()%2 == 0 {
			kind = ""
		}
	}

	return name, kind, "cert-manager.io"
}

func (t *TestData) DNSNames() string {
	return t.randSubset(dnsNamesData())
}

func (t *TestData) URISANs() string {
	return t.randSubset(uriSANSData())
}

func (t *TestData) IPSANs() string {
	return t.randSubset(ipSANsData())
}

func (t *TestData) Duration() string {
	if t.r.Int()%2 == 0 {
		return ""
	}

	return time.Duration(t.r.Int63()).String()
}

func (t *TestData) CommonName() string {
	cns := commonNameData()
	return cns[t.r.Int()%len(cns)]
}

func (t *TestData) IsCA() string {
	switch t.r.Int() % 3 {
	case 0:
		return ""
	case 1:
		return "true"
	default:
		return "false"
	}
}

// Return random subset of given list. Can be empty.
func (t *TestData) randSubset(set []string) string {
	var out []string
	n := t.r.Int() % len(set)

	for i := 0; i < n; i++ {
		r := t.r.Int() % len(set)
		out = append(out, set[r])
		set = append(set[:r], set[i+1:]...)
	}

	return strings.Join(out, ",")
}

func dnsNamesData() []string {
	return []string{
		"a.exmaple.com",
		"b.example.com",
		"c.example.com",
	}
}

func uriSANSData() []string {
	return []string{
		"spiffe://my-service.sandbox.cluster.local",
		"http://foo.bar",
		"spiffe://foo.bar.local",
	}
}

func ipSANsData() []string {
	return []string{
		"192.168.0.1",
		"123.4.5.6",
		"8.8.8.8",
	}
}

func commonNameData() []string {
	return []string{
		"",
		"foo-bar",
		"bla.bla",
		"boo",
	}
}
