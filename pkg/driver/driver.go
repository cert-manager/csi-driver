package driver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
	"github.com/jetstack/cert-manager-csi/pkg/util"
	"github.com/jetstack/cert-manager-csi/pkg/webhook"
)

const (
	Version = "0.1.0-alpha.1"
)

type Driver struct {
	endpoint string

	wh *webhook.Webhook

	ids *identityServer
	cs  *ControllerServer
	ns  *NodeServer
}

func New(driverID *csiapi.DriverID, endpoint,
	dataRoot, tmpfsSize string, wh *webhook.Webhook) (*Driver, error) {
	glog.Infof("driver: %v version: %v", driverID.DriverName, Version)

	mntPoint, err := util.IsLikelyMountPoint(dataRoot)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dataRoot, 0700); err != nil {
			return nil, status.Error(codes.Internal,
				fmt.Sprintf("failed to create data root directory %s: %s", dataRoot, err))
		}

		mntPoint = false
	}

	if !mntPoint {
		execErr := new(bytes.Buffer)
		cmd := exec.Command("mount", "-F", "tmpfs", "-o", "size="+tmpfsSize+"m", "swap", dataRoot)
		cmd.Stderr = execErr

		if err := cmd.Run(); err != nil {
			glog.Errorf("node: failed to mount data root (%s): %s",
				execErr.String(), err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	ns, err := NewNodeServer(driverID, dataRoot, tmpfsSize, wh)
	if err != nil {
		return nil, err
	}

	return &Driver{
		endpoint: endpoint,
		wh:       wh,
		ids:      NewIdentityServer(driverID.DriverName, Version),
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
