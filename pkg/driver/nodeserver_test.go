package driver

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

//func TestNodePublishVolume(t *testing.T) {
//	type npvT struct {
//		req      csi.NodePublishVolumeRequest
//		expError error
//	}
//
//	dir, err := ioutil.TempDir(os.TempDir(),
//		"cert-manager-csi-NodePublishVolume")
//	if err != nil {
//		t.Error(err)
//		t.FailNow()
//	}
//
//	for name, test := range tests {
//		t.Run(name, func(t *testing.T) {
//			ns := new(NodeServer)
//			err := ns.validateAttributes(&test.req)
//			if test.expError == nil {
//				if err != nil {
//					t.Errorf("unexpected error, got=%s",
//						err)
//				}
//
//				return
//			}
//
//			if err == nil || err.Error() != test.expError.Error() {
//				t.Errorf("unexpected error, exp=%s got=%s",
//					test.expError, err)
//			}
//		})
//	}
//}

func TestValidateNodeServerAttributes(t *testing.T) {
	type vaT struct {
		req      csi.NodePublishVolumeRequest
		expError error
	}

	tests := map[string]vaT{
		"if ephemeral volumes are disabled then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "target-path",
				TargetPath: "test-namespace",
				VolumeContext: map[string]string{
					podNameKey:                     "test-pod",
					podNamespaceKey:                "test-pod",
					"csi.storage.k8s.io/ephemeral": "false",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: errors.New("publishing a non-ephemeral volume mount is not supported"),
		},
		"if not volume ID or target path then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeContext: map[string]string{
					podNameKey:      "test-pod",
					podNamespaceKey: "test-namespace",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: errors.New("volume ID missing, target path missing"),
		},
		"if no volume capability procided or pod Namespace then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					podNameKey: "test-pod",
				},
				VolumeCapability: nil,
			},
			expError: errors.New(
				"volume capability missing, expecting both csi.storage.k8s.io/pod.namespace and csi.storage.k8s.io/pod.name attributes to be set in context"),
		},
		"if block access support added then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					podNameKey:      "test-pod",
					podNamespaceKey: "test-namespace",
				},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
				},
			},
			expError: errors.New("block access type not supported"),
		},
		"a request with valid attributes and ephemeral attribute set to 'true' should not error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					podNameKey:      "test-pod",
					podNamespaceKey: "test-namespace",

					"csi.storage.k8s.io/ephemeral": "true",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: nil,
		},
		"a request with valid attributes and no ephemeral attribute should not error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					podNameKey:      "test-pod",
					podNamespaceKey: "test-namespace",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := new(NodeServer)
			err := ns.validateAttributes(&test.req)
			if test.expError == nil {
				if err != nil {
					t.Errorf("unexpected error, got=%s",
						err)
				}

				return
			}

			if err == nil || err.Error() != test.expError.Error() {
				t.Errorf("unexpected error, exp=%s got=%s",
					test.expError, err)
			}
		})
	}
}

func TestCreateDeleteVolume(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(),
		"cert-manager-csi-create-delete-volume")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}()

	ns := &NodeServer{
		dataRoot: dir,
		volumes:  make(map[string]volume),
	}

	id := "test-id"
	attr := map[string]string{
		podNameKey:      "test-pod",
		podNamespaceKey: "test-namespace",
	}

	vol, err := ns.createVolume(id, attr)
	if err != nil {
		t.Error(err)
		return
	}

	path := dir + "/test-id"

	f, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected directory to have been created: %s",
			err)
		return
	}

	if !f.IsDir() {
		t.Errorf("expected volume created to be a directory: %s",
			dir+"/test-id")
		return
	}

	if _, ok := ns.volumes[id]; !ok {
		t.Errorf("expected volume to exist in nodeserver map: %s",
			id)
	}

	if err := ns.deleteVolume(vol); err != nil {
		t.Error(err)
		return
	}

	_, err = os.Stat(path)
	if err == nil || !os.IsNotExist(err) {
		t.Errorf("expected is not exist error but got: %s",
			err)
		return
	}

	if _, ok := ns.volumes[id]; ok {
		t.Errorf("expected volume to not exist in nodeserver map: %s",
			id)
	}
}
