package driver

import (
	"github.com/golang/glog"

	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	version = "0.0.1-alpha.0"
)

type Driver struct {
	endpoint string

	ids *identityServer
	ns  *nodeServer
}

func New(driverName, nodeID, endpoint, dataRoot string) *Driver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	return &Driver{
		endpoint: endpoint,
		ids:      NewIdentityServer(driverName, version),
		ns:       NewNodeServer(nodeID, dataRoot),
	}
}

func (d *Driver) Run() {
	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(d.endpoint, d.ids, nil, d.ns)
	s.Wait()
}
