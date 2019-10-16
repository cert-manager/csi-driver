package driver

import (
	"github.com/golang/glog"
)

const (
	Version = "0.1.0-alpha.1"
)

type Driver struct {
	endpoint string

	ids *identityServer
	cs  *ControllerServer
	ns  *NodeServer
}

func New(driverName, nodeID, endpoint, dataRoot string) (*Driver, error) {
	glog.Infof("driver: %v version: %v", driverName, Version)

	ns, err := NewNodeServer(nodeID, dataRoot)
	if err != nil {
		return nil, err
	}

	return &Driver{
		endpoint: endpoint,
		ids:      NewIdentityServer(driverName, Version),
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
