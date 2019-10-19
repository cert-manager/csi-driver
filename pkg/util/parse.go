package util

import (
	"net"
	"net/url"
	"strings"
)

func ParseDNSNames(dnsNames string) []string {
	return strings.Split(dnsNames, ",")
}

func ParseIPAddresses(ips string) []net.IP {
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

func ParseURIs(uris string) ([]*url.URL, error) {
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

func IPAddressesMatch(a, b []net.IP) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}

	return true
}

func URIsMatch(a, b []*url.URL) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}

	return true
}
