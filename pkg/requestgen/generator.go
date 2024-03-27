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
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	cmpki "github.com/cert-manager/cert-manager/pkg/util/pki"

	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"

	"github.com/cert-manager/csi-driver/pkg/apis"
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
			return nil, fmt.Errorf("%q: %w", csiapi.DurationKey, err)
		}
	}

	var request = &x509.CertificateRequest{}
	if lSubjStr, ok := attrs[csiapi.LiteralSubjectKey]; ok && len(lSubjStr) > 0 {
		lSubjStr, err = expand(meta, lSubjStr)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", csiapi.LiteralSubjectKey, err)
		}
		request.RawSubject, err = cmpki.ParseSubjectStringToRawDerBytes(lSubjStr)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", csiapi.LiteralSubjectKey, err)
		}
	} else {
		request.Subject = pkix.Name{}
		request.Subject.CommonName, err = expand(meta, attrs[csiapi.CommonNameKey])
		if err != nil {
			return nil, fmt.Errorf("%q: %w", csiapi.CommonNameKey, err)
		}
		if len(attrs[csiapi.SerialNumberKey]) > 0 {
			request.Subject.SerialNumber = attrs[csiapi.SerialNumberKey]
		}
		for k, v := range map[*[]string]string{
			&request.Subject.Organization:       csiapi.OrganizationsKey,
			&request.Subject.OrganizationalUnit: csiapi.OrganizationalUnitsKey,
			&request.Subject.Country:            csiapi.CountriesKey,
			&request.Subject.Province:           csiapi.ProvincesKey,
			&request.Subject.Locality:           csiapi.LocalitiesKey,
			&request.Subject.StreetAddress:      csiapi.StreetAddressesKey,
			&request.Subject.PostalCode:         csiapi.PostalCodesKey,
		} {
			if len(attrs[v]) > 0 {
				var e, err = expand(meta, attrs[v])
				if err != nil {
					return nil, fmt.Errorf("%q: %w", v, err)
				}
				*k = strings.Split(e, ",")
			}
		}
	}
	request.DNSNames, err = parseDNSNames(meta, attrs[csiapi.DNSNamesKey])
	if err != nil {
		return nil, fmt.Errorf("%q: %w", csiapi.DNSNamesKey, err)
	}
	request.IPAddresses, err = parseIPAddresses(attrs[csiapi.IPSANsKey])
	if err != nil {
		return nil, fmt.Errorf("%q: %w", csiapi.IPSANsKey, err)
	}
	request.URIs, err = parseURIs(meta, attrs[csiapi.URISANsKey])
	if err != nil {
		return nil, fmt.Errorf("%q: %w", csiapi.URISANsKey, err)
	}

	annotations := make(map[string]string)
	for key, val := range attrs {
		group, _, found := strings.Cut(key, "/")
		if !found {
			continue
		}

		if group != apis.GroupName &&
			group != "csi.storage.k8s.io" {
			annotations[key] = val
		}
	}

	return &manager.CertificateRequestBundle{
		Request:   request,
		IsCA:      strings.ToLower(attrs[csiapi.IsCAKey]) == "true",
		Namespace: attrs[csiapi.K8sVolumeContextKeyPodNamespace],
		Duration:  duration,
		Usages:    keyUsagesFromAttributes(attrs[csiapi.KeyUsagesKey]),
		IssuerRef: cmmeta.ObjectReference{
			Name:  attrs[csiapi.IssuerNameKey],
			Kind:  attrs[csiapi.IssuerKindKey],
			Group: attrs[csiapi.IssuerGroupKey],
		},
		Annotations: annotations,
	}, nil
}

// parseDNSNames parses a csi.cert-manager.io/dns-names value, and returns the
// set of DNS names to be requested. Executes metadata expand on string.
func parseDNSNames(meta metadata.Metadata, dnsNames string) ([]string, error) {
	if len(dnsNames) == 0 {
		return nil, nil
	}
	dns, err := expand(meta, dnsNames)
	if err != nil {
		return nil, err
	}
	return splitList(dns), nil
}

// parseIPAddresses parses a csi.cert-manager.io/ip-sans value, and returns the
// set IP addresses to be requested for.
func parseIPAddresses(ipCSV string) ([]net.IP, error) {
	if len(ipCSV) == 0 {
		return nil, nil
	}

	var ips []net.IP
	var errs []string
	for _, ipStr := range splitList(ipCSV) {
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
// the set of URI SANs to be requested. Executes metadata expand on string.
func parseURIs(meta metadata.Metadata, uriCSV string) ([]*url.URL, error) {
	if len(uriCSV) == 0 {
		return nil, nil
	}

	csv, err := expand(meta, uriCSV)
	if err != nil {
		return nil, err
	}

	var uris []*url.URL
	var errs []string
	for _, uriS := range splitList(csv) {
		uri, err := url.ParseRequestURI(uriS)
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
	for _, usage := range splitList(usagesCSV) {
		keyUsages = append(keyUsages, cmapi.KeyUsage(usage))
	}

	return keyUsages
}

// expand executes os.Expand on the given csv with volume context variables
// provided by the metadata.
func expand(meta metadata.Metadata, csv string) (string, error) {
	vars := map[string]string{
		"POD_NAME":             meta.VolumeContext[csiapi.K8sVolumeContextKeyPodName],
		"POD_NAMESPACE":        meta.VolumeContext[csiapi.K8sVolumeContextKeyPodNamespace],
		"POD_UID":              meta.VolumeContext[csiapi.K8sVolumeContextKeyPodUID],
		"SERVICE_ACCOUNT_NAME": meta.VolumeContext[csiapi.K8sVolumeContextKeyServiceAccountName],
	}

	var errs []string
	exp := os.Expand(csv, func(s string) string {
		v, ok := vars[s]
		if !ok {
			errs = append(errs, fmt.Sprintf("undefined variable %q", s))
		}
		return v
	})

	if len(errs) > 0 {
		return "", fmt.Errorf("%v, known variables: %v",
			strings.Join(errs, ", "),
			[]string{"POD_NAME", "POD_NAMESPACE", "POD_UID", "SERVICE_ACCOUNT_NAME"},
		)
	}

	return exp, nil
}

// splitList returns the given csv as a slice. Trims space of each element.
func splitList(csv string) []string {
	var list []string
	for _, s := range strings.Split(csv, ",") {
		list = append(list, strings.TrimSpace(s))
	}
	return list
}
