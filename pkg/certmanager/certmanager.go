package certmanager

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	cmclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/jetstack/cert-manager/pkg/util/pki"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/util"
)

type CertManager struct {
	cmClient cmclient.Interface
}

func New() (*CertManager, error) {
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
	}, nil
}

func (c *CertManager) CreateNewCertificate(vol *csiapi.MetaData, keyBundle *util.KeyBundle) (*x509.Certificate, error) {
	attr := vol.Attributes

	uris, err := util.ParseURIs(attr[csiapi.URISANsKey])
	if err != nil {
		return nil, err
	}

	ips := util.ParseIPAddresses(attr[csiapi.IPSANsKey])

	dnsNames := strings.Split(attr[csiapi.DNSNamesKey], ",")
	commonName := attr[csiapi.CommonNameKey]

	duration := cmapi.DefaultCertificateDuration
	if durStr, ok := attr[csiapi.DurationKey]; ok {
		duration, err = time.ParseDuration(durStr)
		if err != nil {
			return nil, err
		}
	}

	isCA := false
	if isCAStr, ok := attr[csiapi.IsCAKey]; ok {
		switch strings.ToLower(isCAStr) {
		case "true":
			isCA = true
		case "false":
			isCA = false
		}
	}

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
		return nil, err
	}

	namespace := attr[csiapi.CSIPodNamespaceKey]
	_, err = c.cmClient.CertmanagerV1alpha2().CertificateRequests(namespace).Get(vol.Name, metav1.GetOptions{})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return nil, err
		}
	} else {
		glog.Infof("cert-manager: deleting existing CertificateRequest %s", vol.Name)

		// exists so delete old
		// TODO (@joshvanl): change this to matches spec
		err = c.cmClient.CertmanagerV1alpha2().CertificateRequests(namespace).Delete(vol.Name, &metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	}

	cr := &cmapi.CertificateRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vol.Name,
			Namespace: namespace,
		},
		Spec: cmapi.CertificateRequestSpec{
			CSRPEM: csrPEM,
			IsCA:   isCA,
			Duration: &metav1.Duration{
				Duration: duration,
			},
			IssuerRef: cmmeta.ObjectReference{
				Name:  attr[csiapi.IssuerNameKey],
				Kind:  attr[csiapi.IssuerKindKey],
				Group: attr[csiapi.IssuerGroupKey],
			},
		},
	}

	glog.Infof("cert-manager: created CertificateRequest %s", vol.Name)
	_, err = c.cmClient.CertmanagerV1alpha2().CertificateRequests(namespace).Create(cr)
	if err != nil {
		return nil, err
	}

	glog.Infof("cert-manager: waiting for CertificateRequest to= become ready %s", vol.Name)
	cr, err = c.waitForCertificateRequestReady(cr.Name, namespace, time.Second*30)
	if err != nil {
		return nil, err
	}

	// Write metadata to file
	metaDataBytes, err := json.Marshal(vol)
	if err != nil {
		return nil, err
	}

	metaPath := filepath.Join(vol.Path, csiapi.MetaDataFileName)
	if err := ioutil.WriteFile(metaPath, metaDataBytes, 0600); err != nil {
		return nil, err
	}

	glog.V(4).Infof("cert-manager: metadata written to file %s", metaPath)

	certPath := util.CertPath(vol)

	if err := util.WriteFile(certPath, cr.Status.Certificate, 0600); err != nil {
		return nil, err
	}

	cert, err := pki.DecodeX509CertificateBytes(cr.Status.Certificate)
	if err != nil {
		return nil, err
	}

	glog.Infof("cert-manager: certificate written to file %s", certPath)

	keyPath := util.KeyPath(vol)
	if err := util.WriteFile(keyPath, keyBundle.PEM, 0600); err != nil {
		return nil, fmt.Errorf("faild to write key data to file: %s", err)
	}

	glog.Infof("cert-manager: private key written to file: %s", keyPath)

	return cert, nil
}

func (c *CertManager) RenewCertificate(vol *csiapi.MetaData) (*x509.Certificate, error) {
	var err error
	var keyBundle *util.KeyBundle

	glog.Infof("cert-manager: renewing certicate %s", vol.Name)

	if b, ok := vol.Attributes[csiapi.ReusePrivateKey]; !ok || b != "true" {
		keyBundle, err = util.NewRSAKey()
		if err != nil {
			return nil, err
		}

	} else {

		keyBytes, err := ioutil.ReadFile(util.KeyPath(vol))

		if err != nil {
			return nil, err
		}

		sk, err := pki.DecodePKCS1PrivateKeyBytes(keyBytes)
		if err != nil {
			return nil, err
		}

		keyBundle = &util.KeyBundle{
			PEM:                keyBytes,
			PrivateKey:         sk,
			SignatureAlgorithm: x509.SHA256WithRSA,
			PublicKeyAlgorithm: x509.RSA,
		}
	}

	cert, err := c.CreateNewCertificate(vol, keyBundle)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *CertManager) DeleteCertificateRequest(vol *csiapi.MetaData) error {
	namespace := vol.Attributes[csiapi.CSIPodNamespaceKey]
	name := vol.Attributes[csiapi.CSIPodNameKey]
	return c.cmClient.CertmanagerV1alpha2().CertificateRequests(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (c *CertManager) waitForCertificateRequestReady(name, ns string, timeout time.Duration) (*cmapi.CertificateRequest, error) {
	var cr *cmapi.CertificateRequest
	err := wait.PollImmediate(time.Second, timeout,
		func() (bool, error) {

			glog.V(4).Infof("cert-manager: polling CertificateRequest %s/%s for ready status", name, ns)

			var err error
			cr, err = c.cmClient.CertmanagerV1alpha2().CertificateRequests(ns).Get(name, metav1.GetOptions{})
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
