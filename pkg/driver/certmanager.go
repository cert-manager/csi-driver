package driver

import (
	"fmt"
	"strings"

	cmclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	//"k8s.io/client-go/rest"
)

const (
	issuerNameKey = "certmanager.k8s.io/issuer-name"
	issuerKindKey = "certmanager.k8s.io/issuer-kind"
	commonNameKey = "certmanager.k8s.io/common-name"
	dnsNamesKey   = "certmanager.k8s.io/dns-names"
	ipSANsKey     = "certmanager.k8s.io/ip-sans"
)

type certmanager struct {
	cmClient cmclient.Interface
}

func NewCertManager() (*certmanager, error) {
	//restConfig, err := rest.InClusterConfig()
	//if err != nil {
	//	return nil, err
	//}

	//cmClient, err := cmclient.NewForConfig(restConfig)
	//if err != nil {
	//	return nil, err
	//}

	return &certmanager{
		//cmClient: cmClient,
	}, nil
}

func (c *certmanager) validateAttributes(attr map[string]string) error {
	var errs []string

	if len(attr[issuerNameKey]) == 0 {
		errs = append(errs, fmt.Sprintf("%s field required", issuerNameKey))
	}

	if len(attr[dnsNamesKey]) == 0 && len(attr[ipSANsKey]) == 0 {
		errs = append(errs, fmt.Sprintf("both %s and %s may not be empty",
			commonNameKey, dnsNamesKey))
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to validate volume attributes: \n* %s",
			strings.Join(errs, "\n* "))
	}

	return nil
}
