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
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	"github.com/cert-manager/csi-driver/pkg/apis/defaults"
	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	"github.com/cert-manager/csi-driver/pkg/apis/validation"
)

// RequestForMetadata returns a csi-lib CertificateRequestBundle built using
// the volume attributed contained within the passed metadata.
func RequestForMetadata(meta metadata.Metadata) (*manager.CertificateRequestBundle, error) {
	attrs, err := defaults.SetDefaultAttributes(meta.VolumeContext)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateAttributes(attrs); err != nil {
		return nil, err.ToAggregate()
	}

	duration := cmapi.DefaultCertificateDuration
	if durStr, ok := attrs[csiapi.DurationKey]; ok {
		duration, err = time.ParseDuration(durStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q attribute: %w", csiapi.DurationKey, err)
		}
	}

	commonName, err := executeTemplate(meta, attrs[csiapi.CommonNameKey])
	if err != nil {
		return nil, err
	}
	dns, err := parseDNSNames(meta, attrs[csiapi.DNSNamesKey])
	if err != nil {
		return nil, err
	}
	uris, err := parseURIs(meta, attrs[csiapi.URISANsKey])
	if err != nil {
		return nil, fmt.Errorf("invalid URI provided in %q attribute: %w", csiapi.URISANsKey, err)
	}
	ips, err := parseIPAddresses(attrs[csiapi.IPSANsKey])
	if err != nil {
		return nil, err
	}

	return &manager.CertificateRequestBundle{
		Request: &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: commonName,
			},
			DNSNames:    dns,
			IPAddresses: ips,
			URIs:        uris,
		},
		IsCA:      strings.ToLower(attrs[csiapi.IsCAKey]) == "true",
		Namespace: attrs[csiapi.K8sVolumeContextKeyPodNamespace],
		Duration:  duration,
		Usages:    keyUsagesFromAttributes(attrs[csiapi.KeyUsagesKey]),
		IssuerRef: cmmeta.ObjectReference{
			Name:  attrs[csiapi.IssuerNameKey],
			Kind:  attrs[csiapi.IssuerKindKey],
			Group: attrs[csiapi.IssuerGroupKey],
		},
		Annotations: nil,
	}, nil
}

// parseDNSNames parses a csi.cert-manager.io/dns-names value, and returns the
// set of DNS names to be requested. Executes metadata template on string.
func parseDNSNames(meta metadata.Metadata, dnsNames string) ([]string, error) {
	if len(dnsNames) == 0 {
		return nil, nil
	}

	csv, err := executeTemplate(meta, dnsNames)
	if err != nil {
		return nil, err
	}

	return strings.Split(csv, ","), nil
}

// parseIPAddresses parses a csi.cert-manager.io/ip-sans value, and returns the
// set IP addresses to be requested for.
func parseIPAddresses(ipCSV string) ([]net.IP, error) {
	if len(ipCSV) == 0 {
		return nil, nil
	}

	var ips []net.IP
	var errs []string
	for _, ipStr := range strings.Split(ipCSV, ",") {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			errs = append(errs, ipStr)
			continue
		}
		ips = append(ips, ip)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf(`failed to parse IP address: ["%s"]`, strings.Join(errs, `","`))
	}

	return ips, nil
}

// parseIPAddresses parses a csi.cert-manager.io/uri-sans value, and returns
// the set of URI SANs to be requested. Executes metadata template on string.
func parseURIs(meta metadata.Metadata, uriCSV string) ([]*url.URL, error) {
	if len(uriCSV) == 0 {
		return nil, nil
	}

	csv, err := executeTemplate(meta, uriCSV)
	if err != nil {
		return nil, err
	}

	var uris []*url.URL
	var errs []string
	for _, uriS := range strings.Split(csv, ",") {
		uri, err := url.Parse(uriS)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		uris = append(uris, uri)
	}

	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", "))
	}

	return uris, nil
}

// keyUsagesFromAttributes returns the set of key usages from the given CSV.
func keyUsagesFromAttributes(usagesCSV string) []cmapi.KeyUsage {
	if len(usagesCSV) == 0 {
		return nil
	}

	var keyUsages []cmapi.KeyUsage
	for _, usage := range strings.Split(usagesCSV, ",") {
		keyUsages = append(keyUsages, cmapi.KeyUsage(strings.TrimSpace(usage)))
	}

	return keyUsages
}

// executeTemplate executes the template on the given csv with volume context
// provided by the metadata.
func executeTemplate(meta metadata.Metadata, csv string) (string, error) {
	ptmpl, err := template.New("").Parse(csv)
	if err != nil {
		return "", fmt.Errorf("failed to parse dnsNames for templating: %w", err)
	}

	var buf bytes.Buffer
	if err := ptmpl.Execute(&buf, struct {
		PodName      string
		PodNamespace string
		PodUID       string
	}{
		PodName:      meta.VolumeContext[csiapi.K8sVolumeContextKeyPodName],
		PodNamespace: meta.VolumeContext[csiapi.K8sVolumeContextKeyPodNamespace],
		PodUID:       meta.VolumeContext[csiapi.K8sVolumeContextKeyPodUID],
	}); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
