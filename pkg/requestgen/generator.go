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
	"fmt"
	"net"
	"net/url"
	"strings"
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

	namespace := attrs["csi.storage.k8s.io/pod.namespace"]

	uris, err := parseURIs(attrs[csiapi.URISANsKey])
	if err != nil {
		return nil, fmt.Errorf("invalid URI provided in %q attribute: %w", csiapi.URISANsKey, err)
	}

	ips := parseIPAddresses(attrs[csiapi.IPSANsKey])

	dnsNames := strings.Split(attrs[csiapi.DNSNamesKey], ",")
	commonName := attrs[csiapi.CommonNameKey]

	duration := cmapi.DefaultCertificateDuration
	if durStr, ok := attrs[csiapi.DurationKey]; ok {
		duration, err = time.ParseDuration(durStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q attribute: %w", csiapi.DurationKey, err)
		}
	}

	return &manager.CertificateRequestBundle{
		Request: &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: commonName,
			},
			DNSNames:    dnsNames,
			IPAddresses: ips,
			URIs:        uris,
		},
		IsCA:      strings.ToLower(attrs[csiapi.IsCAKey]) == "true",
		Namespace: namespace,
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

func parseIPAddresses(ips string) []net.IP {
	if len(ips) == 0 {
		return nil
	}

	ipsS := strings.Split(ips, ",")

	var ipAddresses []net.IP

	for _, ipName := range ipsS {
		ip := net.ParseIP(ipName)
		if ip != nil {
			ipAddresses = append(ipAddresses, ip)
		}
	}

	return ipAddresses
}

func parseURIs(uris string) ([]*url.URL, error) {
	if len(uris) == 0 {
		return nil, nil
	}

	urisS := strings.Split(uris, ",")

	var urisURL []*url.URL

	for _, uriS := range urisS {
		uri, err := url.Parse(uriS)
		if err != nil {
			return nil, err
		}

		urisURL = append(urisURL, uri)
	}

	return urisURL, nil
}

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
