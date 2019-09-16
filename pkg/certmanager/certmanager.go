package certmanager

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	cmclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/joshvanl/cert-manager-csi/pkg/util"
)

type CertManager struct {
	nodeID   string
	dataDir  string
	cmClient cmclient.Interface
}

func New(nodeID, dataDir string) (*CertManager, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	cmClient, err := cmclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &CertManager{
		cmClient: cmClient,
		nodeID:   nodeID,
		dataDir:  dataDir,
	}, nil
}

func (c *CertManager) CreateKeyCertPair(vol *v1alpha1.Volume, attr v1alpha1.Attributes) error {
	uris, err := util.ParseURIs(attr[v1alpha1.URISANsKey])
	if err != nil {
		return err
	}

	ips := util.ParseIPAddresses(attr[v1alpha1.IPSANsKey])

	dnsNames := strings.Split(attr[v1alpha1.DNSNamesKey], ",")
	commonName := attr[v1alpha1.CommonNameKey]

	if len(commonName) == 0 {
		commonName = dnsNames[0]
	}

	duration := cmapi.DefaultCertificateDuration
	if durStr, ok := attr[v1alpha1.DurationKey]; ok {
		duration, err = time.ParseDuration(durStr)
		if err != nil {
			return err
		}
	}

	isCA := false
	if isCAStr, ok := attr[v1alpha1.IsCAKey]; ok {
		switch strings.ToLower(isCAStr) {
		case "true":
			isCA = true
		case "false":
			isCA = false
		}
	}

	keyPath := filepath.Join(vol.Path, attr[v1alpha1.KeyFileKey])

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

	namespace := attr[v1alpha1.NamespaceKey]

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
				Name:  attr[v1alpha1.IssuerNameKey],
				Kind:  attr[v1alpha1.IssuerKindKey],
				Group: attr[v1alpha1.IssuerGroupKey],
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

	certPath := filepath.Join(vol.Path, attr[v1alpha1.CertFileKey])

	if err := util.WriteFile(certPath, cr.Status.Certificate, 0600); err != nil {
		return err
	}

	glog.Infof("cert-manager: certificate written to file %s", certPath)

	return nil
}

func (c *CertManager) waitForCertificateRequestReady(name, ns string, timeout time.Duration) (*cmapi.CertificateRequest, error) {
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
