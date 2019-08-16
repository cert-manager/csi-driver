package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
)

const (
	kib int64 = 1024
	mib int64 = kib * 1024

	maxStorageCapacity = 10 * mib

	deviceID = "deviceID"
)

type nodeServer struct {
	nodeID   string
	dataRoot string

	volumes map[string]volume
}

type volume struct {
	Name string `json:"volName"`
	ID   string `json:"volID"`
	Size int64  `json:"volSize"`
	Path string `json:"volPath"`
}

func NewNodeServer(nodeId, dataRoot string) *nodeServer {
	return &nodeServer{
		nodeID:   nodeId,
		dataRoot: dataRoot,

		volumes: make(map[string]volume),
	}
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	ephemeralVolume := req.GetVolumeContext()["csi.storage.k8s.io/ephemeral"] == "true" ||
		req.GetVolumeContext()["csi.storage.k8s.io/ephemeral"] == "" // Kubernetes 1.15 doesn't have csi.storage.k8s.io/ephemeral.

	if !ephemeralVolume {
		return nil, status.Error(codes.InvalidArgument, "publishing a non-ephemeral volume mount is not supported")
	}

	targetPath := req.GetTargetPath()

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
	if !req.GetReadonly() {
		return nil, status.Error(codes.InvalidArgument, "volume must be in read only mode")
	}

	volID := req.GetVolumeId()
	volName := fmt.Sprintf("cert-manager-csi-%s", volID)
	vol, err := ns.createVolume(req.GetVolumeId(), volName, maxStorageCapacity)
	if err != nil && !os.IsExist(err) {
		glog.Error("failed to create volume: ", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.Infof("created volume: %s", vol.Path)

	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(targetPath, 0750); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		notMnt = true
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	deviceId := ""
	if req.GetPublishContext() != nil {
		deviceId = req.GetPublishContext()[deviceID]
	}

	volumeId := req.GetVolumeId()
	attrib := req.GetVolumeContext()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()

	glog.V(4).Infof("target %v\nfstype %v\ndevice %v\nreadonly %v\nvolumeId %v\nattributes %v\nmountflags %v\n",
		targetPath, fsType, deviceId, true, volumeId, attrib, mountFlags)

	options := []string{"bind"}
	//if readOnly {
	options = append(options, "ro")
	//}
	mounter := mount.New("")

	if err := mounter.Mount(vol.Path, targetPath, "", options); err != nil {
		var errList strings.Builder
		errList.WriteString(err.Error())
		if rmErr := os.RemoveAll(vol.Path); rmErr != nil && !os.IsNotExist(rmErr) {
			errList.WriteString(fmt.Sprintf(" :%s", rmErr.Error()))
		}

		return nil, status.Error(codes.Internal, errList.String())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	// Check arguments
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	vol, ok := ns.volumes[volumeID]
	if !ok {
		return nil, status.Error(codes.NotFound,
			fmt.Sprintf("volume id %s does not exit in the volumes list", volumeID))
	}

	// Unmounting the image
	err := mount.New("").Unmount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.V(4).Infof("hostpath: volume %s/%s has been unmounted.", targetPath, volumeID)

	glog.V(4).Infof("deleting volume %s", volumeID)
	if err := ns.deleteHostpathVolume(&vol); err != nil && !os.IsNotExist(err) {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete volume: %s", err))
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// createVolume create the directory for the hostpath volume.
// It returns the volume path or err if one occurs.
func (ns *nodeServer) createVolume(id, name string, cap int64) (*volume, error) {
	path := filepath.Join(ns.dataRoot, id)

	err := os.MkdirAll(path, 0777)
	if err != nil {
		return nil, err
	}

	vol := volume{
		ID:   id,
		Name: name,
		Size: cap,
		Path: path,
	}

	ns.volumes[id] = vol
	return &vol, nil
}

// deleteVolume deletes the directory for the hostpath volume.
func (ns *nodeServer) deleteHostpathVolume(vol *volume) error {
	glog.V(4).Infof("deleting hostpath volume: %s", vol.ID)

	if err := os.RemoveAll(vol.Path); err != nil && !os.IsNotExist(err) {
		return err
	}

	delete(ns.volumes, vol.ID)

	return nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, in *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
