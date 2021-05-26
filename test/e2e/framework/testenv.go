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

package framework

import (
	"context"
	"fmt"
	"time"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// Defines methods that help provision test environments

const (
	// How often to poll for conditions
	Poll = 2 * time.Second

	rootCert = `-----BEGIN CERTIFICATE-----
MIID4DCCAsigAwIBAgIJAJzTROInmDkQMA0GCSqGSIb3DQEBCwUAMFMxCzAJBgNV
BAYTAlVLMQswCQYDVQQIEwJOQTEVMBMGA1UEChMMY2VydC1tYW5hZ2VyMSAwHgYD
VQQDExdjZXJ0LW1hbmFnZXIgdGVzdGluZyBDQTAeFw0xNzA5MTAxODMzNDNaFw0y
NzA5MDgxODMzNDNaMFMxCzAJBgNVBAYTAlVLMQswCQYDVQQIEwJOQTEVMBMGA1UE
ChMMY2VydC1tYW5hZ2VyMSAwHgYDVQQDExdjZXJ0LW1hbmFnZXIgdGVzdGluZyBD
QTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAM+Q2AO4hARav0qwjk7I
4mEh5R201HS8s7HpaLOXBNvvh7qJ9yJz6jLqYg6EvP0K/bK56Cp2oe2igd7GOxpV
3YPOc3CG0CCqHMprEcvxj2xBKX00Rtcn4oVLhDPhAb0BV/R7NFLeWxzh+ggvPI1X
m1qLaWYqYZEJ5bBsYXD3tPdS4GGINRz8Zvih46f0Z2wVkCGoTpsbX8HO74sa2Day
UjzAsWGlO5bZGiMSHjDEnf9yek2TcjEyVoohoOLaQg/ng21T5RWzeZKTl1cznwuG
Vr9tZfHFqxQ5qeaId+1ICtxNvkEjbTnZl6Wy9Cthn0dxwOeS5TqMJ7SFNXy1gp4j
f/MCAwEAAaOBtjCBszAdBgNVHQ4EFgQUBtrjvWfbkLA0iX6sKVRhKUo864kwgYMG
A1UdIwR8MHqAFAba471n25CwNIl+rClUYSlKPOuJoVekVTBTMQswCQYDVQQGEwJV
SzELMAkGA1UECBMCTkExFTATBgNVBAoTDGNlcnQtbWFuYWdlcjEgMB4GA1UEAxMX
Y2VydC1tYW5hZ2VyIHRlc3RpbmcgQ0GCCQCc00TiJ5g5EDAMBgNVHRMEBTADAQH/
MA0GCSqGSIb3DQEBCwUAA4IBAQCR+jXhup5tCKwhAf8xgvp589BczQOjmotuZGEL
Dcint2y263ChEdsoLhyJfvFCAZfTSm+UT95Hl+ZKVuoVEcAS7udaFUFpC/gIYVOi
H4/uvJps4SpVCB7+T/orcTjZ2ewT23mQAQg+B+iwX9VCof+fadkYOg1XD9/eaj6E
9McXID3iuCXg02RmEOwVMrTggHPwHrOGAilSaZc58cJZHmMYlT5rGrJcWS/AyXnH
VOodKC004yjh7w9aSbCCbAL0tDEnhm4Jrb8cxt7pDWbdEVUeuk9LZRQtluYBnmJU
kQ7ALfUfUh/RUpCV4uI6sEI3NDX2YqQbOtsBD/hNaL1F85FA
-----END CERTIFICATE-----`

	rootKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAz5DYA7iEBFq/SrCOTsjiYSHlHbTUdLyzselos5cE2++Huon3
InPqMupiDoS8/Qr9srnoKnah7aKB3sY7GlXdg85zcIbQIKocymsRy/GPbEEpfTRG
1yfihUuEM+EBvQFX9Hs0Ut5bHOH6CC88jVebWotpZiphkQnlsGxhcPe091LgYYg1
HPxm+KHjp/RnbBWQIahOmxtfwc7vixrYNrJSPMCxYaU7ltkaIxIeMMSd/3J6TZNy
MTJWiiGg4tpCD+eDbVPlFbN5kpOXVzOfC4ZWv21l8cWrFDmp5oh37UgK3E2+QSNt
OdmXpbL0K2GfR3HA55LlOowntIU1fLWCniN/8wIDAQABAoIBAQCYvGvIKSG0FpbG
vi6pmLbEZO20s1jW4fiUxT2PUWR49sR4pocdahB/EOvA5TowNcNDnftSK+Ox+q/4
HwRkt6R+Fg/qULmcH7F53dnFqeYw8a42/J3YOvg7v7rzdfISg4eWVobFJ+wBz+Nt
3FyBYWLm+MlBLZSH5rGG5em59/zJNHWIhH+oQPfCxAkYEvd8tXOTUzjhqvEfjaJy
FZghnT9xto4MwDdNCPbtzdNjTMhiv0AHkcZGGtRJfkehXX2qhXOQ2UzzO9XrMZnv
5KgYf+bXKJsyS3SPl6TTl7vg2gKBciRvsdFhMy5I5GyIADrEDJnNNmXQRtiaFLfd
k/aqfPT5AoGBAPquMouZUbVS/Qh+qbls7G4zAuznfCiqdctcKmUGPRP4sTTjWdUp
fjI+UTt1e8hncmr4RY7Oa9kUV/kDwzS5spUZZ+u0PczS3XKxOwNOleoH00dfc9vt
cxctHdPdDTndRi8Z4k3m931jIX7jB/Pyx8qeNYB3pj0k3ThktwMbAVLnAoGBANP4
beI5zpbvtAdExJcuxx2mRDGF0lIdKC0bvQaeqM3Lwqnmc0Fz1dbP7KXDa+SdJWPd
res+NHPZoEPeEJuDTSngXOLNECZe4Ja9frn1TeY858vMJBwIkyc8zu+sgXxjQUM+
TWUlTUhtXyybkRnxAEny4OT2TTgmXITJaKOmV1UVAoGAHaXSlo4YitB42rNYUXTf
dZ0U4H30Qj7+1YFeBjq5qI4GL1IgQsS4hyq1osmfTTFm593bJCunt7HfQbU/NhIs
W9P4ZXkYwgvCYxkw+JAnzNkGFO/mHQG1Ve1hFLiVIt3XuiRejoYdiTfbM02YmDKD
jKQvgbUk9SBSBaRrvLNJ8csCgYAYnrZEnGo+ZcEHRxl+ZdSCwRkSl3SCTRiphJtD
9ZGttYj6quWgKJAhzyyxZC1X9FivbMQSmrsE6bYPq+9J4MpJnuGrBh5mFocHeyMI
/lD5+QEDTsay6twMpqdydxrjE7Q01zuuD9MWIn33dGo6FR/vduJgNatqZipA0hPx
ThS+sQKBgQDh0+cVo1mfYiCkp3IQPB8QYiJ/g2/UBk6pH8ZZDZ+A5td6NveiWO1y
wTEUWkX2qyz9SLxWDGOhdKqxNrLCUSYSOV/5/JQEtBm6K50ArFtrY40JP/T/5KvM
tSK2ayFX1wQ3PuEmewAogy/20tWo80cr556AXA62Utl2PzLK30Db8w==
-----END RSA PRIVATE KEY-----`
)

// CreateKubeNamespace creates a new Kubernetes Namespace for a test.
func (f *Framework) CreateKubeNamespace(baseName string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-cert-manager-csi-tests-%v-", baseName),
		},
	}

	return f.KubeClientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
}

// CreateCAIssuer creates a CA issuer used for creating certificates
func (f *Framework) CreateCAIssuer(namespace, baseName string) (cmmeta.ObjectReference, error) {
	sec, err := f.KubeClientSet.CoreV1().Secrets(namespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: baseName + "-",
			Namespace:    namespace,
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte(rootCert),
			corev1.TLSPrivateKeyKey: []byte(rootKey),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return cmmeta.ObjectReference{}, err
	}

	issuer, err := f.CertManagerClientSet.CertmanagerV1().Issuers(namespace).Create(context.TODO(), &cmapi.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ca-issuer-",
			Namespace:    namespace,
		},
		Spec: cmapi.IssuerSpec{
			IssuerConfig: cmapi.IssuerConfig{
				CA: &cmapi.CAIssuer{
					SecretName: sec.Name,
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return cmmeta.ObjectReference{}, err
	}

	return cmmeta.ObjectReference{
		Name:  issuer.Name,
		Kind:  issuer.Kind,
		Group: cm.GroupName,
	}, nil
}

// CreateCAClusterIssuer creates a CA cluster issuer used for creating certificates
func (f *Framework) CreateCAClusterIssuer(baseName string) (cmmeta.ObjectReference, error) {
	sec, err := f.KubeClientSet.CoreV1().Secrets("cert-manager").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: baseName + "-",
			Namespace:    "cert-manager",
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte(rootCert),
			corev1.TLSPrivateKeyKey: []byte(rootKey),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return cmmeta.ObjectReference{}, err
	}

	issuer, err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Create(context.TODO(), &cmapi.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ca-clusterissuer-",
		},
		Spec: cmapi.IssuerSpec{
			IssuerConfig: cmapi.IssuerConfig{
				CA: &cmapi.CAIssuer{
					SecretName: sec.Name,
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return cmmeta.ObjectReference{}, err
	}

	return cmmeta.ObjectReference{
		Name:  issuer.Name,
		Kind:  issuer.Kind,
		Group: cm.GroupName,
	}, nil
}

// DeleteKubeNamespace will delete a namespace resource
func (f *Framework) DeleteKubeNamespace(namespace string) error {
	return f.KubeClientSet.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
}

// WaitForKubeNamespaceNotExist will wait for the namespace with the given name
// to not exist for up to 2 minutes.
func (f *Framework) WaitForKubeNamespaceNotExist(namespace string) error {
	return wait.PollImmediate(Poll, time.Minute*2, namespaceNotExist(f.KubeClientSet, namespace))
}

func namespaceNotExist(c kubernetes.Interface, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}
}
