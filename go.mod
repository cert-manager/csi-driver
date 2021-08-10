module github.com/jetstack/cert-manager-csi

go 1.17

require (
	github.com/cert-manager/csi-lib v0.0.0-20210809101349-dd8ae5d66f53
	github.com/jetstack/cert-manager v1.4.0
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
)
