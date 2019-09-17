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
	cs  *ControllerServer
	ns  *NodeServer
}

func New(driverName, nodeID, endpoint, dataRoot string) (*Driver, error) {
	glog.Infof("driver: %v version: %v", driverName, version)

	ns, err := NewNodeServer(nodeID, dataRoot)
	if err != nil {
		return nil, err
	}

	return &Driver{
		endpoint: endpoint,
		ids:      NewIdentityServer(driverName, version),
		cs:       NewControllerServer(),
		ns:       ns,
	}, nil
}

func (d *Driver) Run() {
	s := NewNonBlockingGRPCServer()
	s.Start(d.endpoint, d.ids, d.cs, d.ns)
	s.Wait()
}

func (d *Driver) NodeServer() *NodeServer {
	return d.ns
}
