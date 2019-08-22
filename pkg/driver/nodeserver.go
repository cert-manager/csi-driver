package driver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/joshvanl/cert-manager-csi/pkg/util"
)

const (
	kib int64 = 1024
	mib int64 = kib * 1024

	maxStorageCapacity = 10 * mib

	deviceID = "deviceID"

	podNameKey      = "csi.storage.k8s.io/pod.name"
	podNamespaceKey = "csi.storage.k8s.io/pod.namespace"
)

type NodeServer struct {
	nodeID   string
	dataRoot string

	cm      *certmanager
	volumes map[string]volume
}

type volume struct {
	Name string
	ID   string
	Size int64
	Path string

	PodName      string
	PodNamespace string
}

func NewNodeServer(nodeID, dataRoot string) (*NodeServer, error) {
	cm, err := NewCertManager(nodeID, dataRoot)
	if err != nil {
		return nil, err
	}

	return &NodeServer{
		nodeID:   nodeID,
		dataRoot: dataRoot,

		cm:      cm,
		volumes: make(map[string]volume),
	}, nil
}

func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	attr := req.GetVolumeContext()
	targetPath := req.GetTargetPath()
	readOnly := req.GetReadonly()

	// Kubernetes 1.15 doesn't have csi.storage.k8s.io/ephemeral.
	ephemeralVolume := attr["csi.storage.k8s.io/ephemeral"] == "true" || attr["csi.storage.k8s.io/ephemeral"] == ""
	if !ephemeralVolume {
		return nil, status.Error(codes.InvalidArgument, "publishing a non-ephemeral volume mount is not supported")
	}

	if err := ns.cm.validateAttributes(attr); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume ID missing in request")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "target path missing in request")
	}

	if req.GetVolumeCapability().GetBlock() != nil {
		return nil, status.Error(codes.InvalidArgument, "block access type not supported")
	}

	volID := req.GetVolumeId()
	vol, err := ns.createVolume(volID, attr)
	if err != nil && !os.IsExist(err) {
		glog.Error("node: failed to create volume: ", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.Infof("node: created volume: %s", vol.Path)

	glog.Infof("node: creating key/cert pair with cert-manager: %s", vol.Path)
	if err := ns.cm.createKeyCertPair(vol, attr); err != nil {
		return nil, err
	}

	mntPoint, err := util.IsLikelyMountPoint(targetPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(targetPath, 0750); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		mntPoint = false
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if mntPoint {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	deviceId := ""
	if req.GetPublishContext() != nil {
		deviceId = req.GetPublishContext()[deviceID]
	}

	glog.V(4).Infof("node: target %v\ndevice %v\nreadonly %v\nvolumeId %v\nattributes %v\n",
		targetPath, deviceId, readOnly, volID, attr)

	var options []string
	if readOnly {
		options = append(options, "ro")
	}

	if err := util.Mount(vol.Path, targetPath, options); err != nil {
		if rmErr := os.RemoveAll(vol.Path); rmErr != nil && !os.IsNotExist(rmErr) {
			err = fmt.Errorf("%s,%s", err, rmErr)
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	// Check arguments
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume ID missing in request")
	}

	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "target path missing in request")
	}

	vol, ok := ns.volumes[volumeID]
	if !ok {
		return nil, status.Error(codes.NotFound,
			fmt.Sprintf("volume id %s does not exit in the volumes list", volumeID))
	}

	// Unmounting the image
	err := util.Unmount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.V(4).Infof("node: volume %s/%s has been unmounted.", targetPath, volumeID)

	glog.V(4).Infof("node: deleting volume %s", volumeID)
	if err := ns.deleteVolume(&vol); err != nil && !os.IsNotExist(err) {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete volume: %s", err))
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// createVolume create the directory for the volume. It returns the volume
// path or err if one occurs.
func (ns *NodeServer) createVolume(id string, attr map[string]string) (*volume, error) {
	path := filepath.Join(ns.dataRoot, id)

	err := os.MkdirAll(path, 0777)
	if err != nil {
		return nil, err
	}

	vol := volume{
		ID:           id,
		Name:         fmt.Sprintf("cert-manager-csi-%s", id),
		Size:         maxStorageCapacity,
		Path:         path,
		PodName:      attr[podNameKey],
		PodNamespace: attr[podNamespaceKey],
	}

	ns.volumes[id] = vol
	return &vol, nil
}

func (ns *NodeServer) deleteVolume(vol *volume) error {
	glog.V(4).Infof("node: deleting volume: %s", vol.ID)

	if err := os.RemoveAll(vol.Path); err != nil && !os.IsNotExist(err) {
		return err
	}

	delete(ns.volumes, vol.ID)

	return nil
}

func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	glog.Info("node: getting default node info")

	return &csi.NodeGetInfoResponse{
		NodeId: ns.nodeID,
	}, nil
}

func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *NodeServer) NodeExpandVolume(ctx context.Context, in *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
