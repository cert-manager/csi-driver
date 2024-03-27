/*
Copyright 2021 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cases

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/cert-manager/csi-driver/test/e2e/framework"
)

func basePod(f *framework.Framework, csiAttributes map[string]string) (corev1.Volume, *corev1.Pod) {
	volume := corev1.Volume{
		Name: "tls",
		VolumeSource: corev1.VolumeSource{
			CSI: &corev1.CSIVolumeSource{
				Driver:           "csi.cert-manager.io",
				ReadOnly:         pointer.Bool(true),
				VolumeAttributes: csiAttributes,
			},
		},
	}

	return volume, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: f.BaseName + "-",
			Namespace:    f.Namespace.Name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "test-container-1",
					Image:   "busybox",
					Command: []string{"sleep", "10000"},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/tls",
							Name:      "tls",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				volume,
			},
		},
	}
}
