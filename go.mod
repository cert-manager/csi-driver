module github.com/cert-manager/csi-driver

go 1.16

require (
	github.com/cert-manager/csi-lib v0.1.0
	github.com/go-logr/logr v0.4.0
	github.com/jetstack/cert-manager v1.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.3
	k8s.io/component-base v0.21.3
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubectl v0.21.0
	k8s.io/utils v0.0.0-20210722164352-7f3ee0f31471
	sigs.k8s.io/controller-runtime v0.9.5
)
