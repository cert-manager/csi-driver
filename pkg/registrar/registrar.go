package registrar

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/golang/glog"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"

	"github.com/joshvanl/cert-manager-csi/pkg/driver"
	registerapi "github.com/joshvanl/cert-manager-csi/pkg/registrar/v1"
)

type Registrar struct {
	driverName        string
	kubeletEndpoint   string
	supportedVersions []string

	ns *driver.NodeServer
}

func New(driverName, kubeletEndpoint string, ns *driver.NodeServer) *Registrar {
	return &Registrar{
		driverName:      driverName,
		kubeletEndpoint: kubeletEndpoint,
		// only CSI 1.0.0 version supported
		supportedVersions: []string{"1.0.0"},

		ns: ns,
	}
}

// Run the non-blocking registration server.
func (r *Registrar) Run() error {
	socketPath := fmt.Sprintf("/registration/%s-reg.sock", r.driverName)

	fi, err := os.Stat(socketPath)
	if err == nil && (fi.Mode()&os.ModeSocket) != 0 {
		// Remove any socket, stale or not, but fall through for other files
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove stale socket %s with error: %+v", socketPath, err)
		}
	}

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat the socket %s with error: %+v", socketPath, err)
	}

	// Default to only user accessible socket, caller can open up later if desired
	oldmask := unix.Umask(0077)

	glog.Infof("registrar: registration Server at: %s\n", socketPath)
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %s with error: %+v", socketPath, err)
	}

	unix.Umask(oldmask)
	glog.Infof("registrar: server started at: %s\n", socketPath)
	grpcServer := grpc.NewServer()

	// Registers kubelet plugin watcher api.
	registerapi.RegisterRegistrationServer(grpcServer, r)

	// Starts service
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			glog.Errorf("registrar: server stopped serving: %v", err)
			os.Exit(1)
		}
	}()

	return nil
}

// GetInfo is the RPC invoked by plugin watcher
func (r *Registrar) GetInfo(ctx context.Context, req *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
	glog.Infof("registrar: received GetInfo call: %+v", req)
	return &registerapi.PluginInfo{
		Type:              registerapi.CSIPlugin,
		Name:              r.driverName,
		Endpoint:          r.kubeletEndpoint,
		SupportedVersions: r.supportedVersions,
	}, nil
}

func (r *Registrar) NotifyRegistrationStatus(ctx context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
	glog.Infof("registrar: received NotifyRegistrationStatus call: %+v", status)
	if !status.PluginRegistered {
		glog.Errorf("registrar: registration process failed with error: %+v, restarting registration container.", status.Error)
		os.Exit(1)
	}

	return &registerapi.RegistrationStatusResponse{}, nil
}
