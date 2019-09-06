package driver

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	cmclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	"github.com/joshvanl/cert-manager-csi/pkg/util"
)

const (
	issuerNameKey  = "csi.certmanager.k8s.io/issuer-name"
	issuerKindKey  = "csi.certmanager.k8s.io/issuer-kind"
	issuerGroupKey = "csi.certmanager.k8s.io/issuer-group"

	commonNameKey = "csi.certmanager.k8s.io/common-name"
	dnsNamesKey   = "csi.certmanager.k8s.io/dns-names"
	ipSANsKey     = "csi.certmanager.k8s.io/ip-sans"
	uriSANsKey    = "csi.certmanager.k8s.io/uri-sans"
	durationKey   = "csi.certmanager.k8s.io/duration"
	isCAKey       = "csi.certmanager.k8s.io/is-ca"

	certFileKey  = "csi.certmanager.k8s.io/certificate-file"
	keyFileKey   = "csi.certmanager.k8s.io/privatekey-file"
	namespaceKey = "csi.certmanager.k8s.io/namespace"
)

type certmanager struct {
	nodeID   string
	dataDir  string
	cmClient cmclient.Interface
}

func NewCertManager(nodeID, dataDir string) (*certmanager, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	cmClient, err := cmclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &certmanager{
		cmClient: cmClient,
		nodeID:   nodeID,
		dataDir:  dataDir,
	}, nil
}

func (c *certmanager) createKeyCertPair(vol *volume, attr map[string]string) error {
	uris, err := util.ParseURIs(attr[uriSANsKey])
	if err != nil {
		return err
	}

	ips := util.ParseIPAddresses(attr[ipSANsKey])

	dnsNames := strings.Split(attr[dnsNamesKey], ",")
	commonName := attr[commonNameKey]

	if len(commonName) == 0 {
		if len(dnsNames) == 0 {
			return errors.New("no common name or DNS names given")
		}

		commonName = dnsNames[0]
	}

	duration := cmapi.DefaultCertificateDuration
	if durStr, ok := attr[durationKey]; ok {
		duration, err = time.ParseDuration(durStr)
		if err != nil {
			return err
		}
	}

	isCA := false
	if isCAStr, ok := attr[isCAKey]; ok {
		switch strings.ToLower(isCAStr) {
		case "true":
			isCA = true
		case "false":
			isCA = false
		default:
			return fmt.Errorf("invalid is-ca value: %s", isCAStr)
		}
	}

	keyPath := attr[keyFileKey]
	if keyPath == "" {
		keyPath = "key.pem"
	}
	keyPath = filepath.Join(vol.Path, keyPath)

	keyBundle, err := util.NewRSAKey(keyPath)
	if err != nil {
		return err
	}

	glog.Infof("cert-manager: new private key written to file: %s", keyPath)

	csr := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames:           dnsNames,
		IPAddresses:        ips,
		URIs:               uris,
		PublicKey:          keyBundle.PrivateKey.Public(),
		PublicKeyAlgorithm: keyBundle.PublicKeyAlgorithm,
		SignatureAlgorithm: keyBundle.SignatureAlgorithm,
	}

	csrPEM, err := util.EncodeCSR(csr, keyBundle.PrivateKey)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("cert-manager-csi-%s-%s-%s",
		c.nodeID, vol.PodName, vol.ID)

	namespace := attr[namespaceKey]
	if len(namespace) == 0 {
		glog.V(4).Infof("certmanager: %s: no namespace specified for key %s so using pod namespace %s",
			vol.Name, namespaceKey, vol.PodNamespace)
		namespace = vol.PodNamespace
	}

	_, err = c.cmClient.CertmanagerV1alpha1().CertificateRequests(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
	} else {
		glog.Infof("cert-manager: deleting existing CertificateRequest %s", name)

		// exists so delete old
		err = c.cmClient.CertmanagerV1alpha1().CertificateRequests(namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	issuerKind := attr[issuerKindKey]
	if issuerKind == "" {
		issuerKind = cmapi.IssuerKind
	}

	issuerGroup := attr[issuerGroupKey]
	if issuerGroup == "" {
		issuerGroup = cm.GroupName
	}

	cr := &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cmapi.CertificateRequestSpec{
			CSRPEM: csrPEM,
			IsCA:   isCA,
			Duration: &metav1.Duration{
				Duration: duration,
			},
			IssuerRef: cmapi.ObjectReference{
				Name:  attr[issuerNameKey],
				Kind:  issuerKind,
				Group: issuerGroup,
			},
		},
	}

	glog.Infof("cert-manager: created CertificateRequest %s", name)
	_, err = c.cmClient.CertmanagerV1alpha1().CertificateRequests(namespace).Create(cr)
	if err != nil {
		return err
	}

	glog.Infof("cert-manager: waiting for CertificateRequest to= become ready %s", name)
	cr, err = c.waitForCertificateRequestReady(cr.Name, namespace, time.Second*30)
	if err != nil {
		return err
	}

	certPath := attr[certFileKey]
	if certPath == "" {
		certPath = "crt.pem"
	}
	certPath = filepath.Join(vol.Path, certPath)

	if err := util.WriteFile(certPath, cr.Status.Certificate, 0600); err != nil {
		return err
	}

	glog.Infof("cert-manager: certificate written to file %s", certPath)

	return nil
}

func (c *certmanager) waitForCertificateRequestReady(name, ns string, timeout time.Duration) (*cmapi.CertificateRequest, error) {
	var cr *cmapi.CertificateRequest
	err := wait.PollImmediate(time.Second, timeout,
		func() (bool, error) {

			glog.V(4).Infof("cert-manager: polling CertificateRequest %s/%s for ready status", name, ns)

			var err error
			cr, err = c.cmClient.CertmanagerV1alpha1().CertificateRequests(ns).Get(name, metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("error getting CertificateRequest %s: %v", name, err)
			}

			if reason, failed := util.CertificateRequestFailed(cr); failed {
				return false, fmt.Errorf("certificate request marked as failed: %s", reason)
			}

			if !util.CertificateRequestReady(cr) {
				return false, nil
			}

			return true, nil
		},
	)

	if err != nil {
		return cr, err
	}

	return cr, nil
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
		return fmt.Errorf("failed to validate volume attributes: %s",
			strings.Join(errs, ", "))
	}

	return nil
}
