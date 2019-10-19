package framework

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	crclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/jetstack/cert-manager-csi/test/e2e/framework/config"
	"github.com/jetstack/cert-manager-csi/test/e2e/framework/helper"
	"github.com/jetstack/cert-manager-csi/test/e2e/framework/util"
)

// DefaultConfig contains the default shared config the is likely parsed from
// command line arguments.
var DefaultConfig = &config.Config{}

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	BaseName string

	Config *config.Config

	// Kubernetes API clientsets
	KubeClientSet        kubernetes.Interface
	CertManagerClientSet crclientset.Interface

	// Namespace in which all test resources should reside
	Namespace *corev1.Namespace

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle

	// The self signed issuer to reference
	Issuer cmmeta.ObjectReference

	helper *helper.Helper
}

func NewDefaultFramework(baseName string) *Framework {
	return NewFramework(baseName, DefaultConfig)
}

func NewFramework(baseName string, cfg *config.Config) *Framework {
	f := &Framework{
		Config:   cfg,
		BaseName: baseName,
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

func (f *Framework) BeforeEach() {
	f.helper = helper.NewHelper(f.Config)
	f.cleanupHandle = AddCleanupAction(f.AfterEach)

	By("Creating a kubernetes client")
	kubeConfig, err := util.LoadConfig(f.Config.KubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	f.KubeClientSet, err = kubernetes.NewForConfig(kubeConfig)
	Expect(err).NotTo(HaveOccurred())

	By("Creating a cert manager client")
	f.CertManagerClientSet, err = crclientset.NewForConfig(kubeConfig)
	Expect(err).NotTo(HaveOccurred())

	By("Building a namespace api object")
	f.Namespace, err = f.CreateKubeNamespace(f.BaseName)
	Expect(err).NotTo(HaveOccurred())

	By("Using the namespace " + f.Namespace.Name)

	By("Creating CA Issuer")
	f.Issuer, err = f.CreateCAIssuer(f.Namespace.Name, f.BaseName)
	Expect(err).NotTo(HaveOccurred())

	f.helper.RestConfig = kubeConfig
	f.helper.CMClient = f.CertManagerClientSet
	f.helper.KubeClient = f.KubeClientSet
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	if !f.Config.Cleanup {
		return
	}

	By("Deleting test namespace")
	err := f.DeleteKubeNamespace(f.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())

	By("Waiting for test namespace to no longer exist")
	err = f.WaitForKubeNamespaceNotExist(f.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())
}

func (f *Framework) Helper() *helper.Helper {
	return f.helper
}

func CasesDescribe(text string, body func()) bool {
	return Describe("[TEST] "+text, body)
}
