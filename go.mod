module github.com/jetstack/cert-manager-csi

go 1.12

require (
	github.com/container-storage-interface/spec v1.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/jetstack/cert-manager v0.11.0
	github.com/kubernetes-csi/csi-lib-utils v0.6.1
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	google.golang.org/grpc v1.23.0
	k8s.io/api v0.0.0-20191010143144-fbf594f18f80
	k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery v0.0.0-20191006235458-f9f2f3f8ab02
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kube-aggregator v0.0.0-20190222095010-0b78038fe9e5
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
	sigs.k8s.io/kind v0.5.1
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5
