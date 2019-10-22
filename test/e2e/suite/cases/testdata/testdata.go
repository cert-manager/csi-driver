package testdata

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

type issuer struct {
	name, kind string
}

type TestData struct {
	*rand.Rand
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
		Rand:    rand.New(rand.NewSource(seed)),
		issuers: unIssuers,
	}, nil
}

func (t *TestData) RandomVolumeAttributes() map[string]string {
	var attr map[string]string

	issName, issKind, issGroup := t.Issuer()
	attr["issuer-name"] = issName
	if issKind == "ClusterIssuer" {
		attr["issuer-kind"] = issKind
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
	} {
		t.maybeAddAttribute(attr, a.k, a.v)
	}

	return attr
}

func (t *TestData) RandNameString() string {
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[t.Intn(len(letterRunes))]
	}

	return string(b)
}

func (t *TestData) maybeAddAttribute(attr map[string]string, k, v string) {
	if t.Int()%2 == 0 {
		return
	}

	attr["csi.cert-manager.io/"+k] = v
}

func (t *TestData) Issuer() (name, kind, group string) {
	n := t.Int() % len(t.issuers)
	name, kind = t.issuers[n].name, t.issuers[n].kind

	if kind == "Issuer" {
		if t.Int()%2 == 0 {
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
	if t.Int()%2 == 0 {
		return ""
	}

	return time.Duration(t.Int63()).String()
}

func (t *TestData) CommonName() string {
	cns := commonNameData()
	n := t.Int()%len(cns) + 1
	if n == len(cns) {
		return ""
	}

	return cns[n]
}

func (t *TestData) IsCA() string {
	switch t.Int() % 3 {
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
	n := t.Int() % len(set)

	for i := 0; i < n; i++ {
		r := t.Int() % len(set)
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
		"foo-bar",
		"bla.bla",
		"boo",
	}
}
