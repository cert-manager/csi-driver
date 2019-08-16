module github.com/joshvanl/cert-manager-csi

go 1.12

require (
	github.com/container-storage-interface/spec v1.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.6.1
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	google.golang.org/genproto v0.0.0-20181109154231-b5d43981345b // indirect
	google.golang.org/grpc v1.23.0
	k8s.io/apimachinery v0.0.0-20181110190943-2a7c93004028 // indirect
	k8s.io/kubernetes v1.12.2
	k8s.io/utils v0.0.0-20181102055113-1bd4f387aa67 // indirect
)
