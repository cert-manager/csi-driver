package driver

import (
	"github.com/golang/glog"
)

const (
	version = "0.0.1-alpha.0"
)

type Driver struct {
	endpoint string

	ids *identityServer
	ns  *nodeServer
}

func New(driverName, nodeID, endpoint, dataRoot string) (*Driver, error) {
	glog.Infof("Driver: %v version: %v", driverName, version)

	ns, err := NewNodeServer(nodeID, dataRoot)
	if err != nil {
		return nil, err
	}

	return &Driver{
		endpoint: endpoint,
		ids:      NewIdentityServer(driverName, version),
		ns:       ns,
	}, nil
}

func (d *Driver) Run() {
	s := NewNonBlockingGRPCServer()
	s.Start(d.endpoint, d.ids, nil, d.ns)
	s.Wait()
}
