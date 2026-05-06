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

package readinessgate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "k8s.io/client-go/kubernetes/fake"

	"github.com/cert-manager/csi-lib/metadata"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

func Test_Parse(t *testing.T) {
	tests := map[string]struct {
		specs    []string
		wantLen  int
		wantErr  bool
	}{
		"empty input returns empty gates": {
			specs:   []string{},
			wantLen: 0,
		},
		"valid pod-ip:any": {
			specs:   []string{"pod-ip:any"},
			wantLen: 1,
		},
		"valid pod-ip:ipv4": {
			specs:   []string{"pod-ip:ipv4"},
			wantLen: 1,
		},
		"valid pod-ip:ipv6": {
			specs:   []string{"pod-ip:ipv6"},
			wantLen: 1,
		},
		"valid pod-condition without explicit status defaults to True": {
			specs:   []string{"pod-condition:Ready"},
			wantLen: 1,
		},
		"valid pod-condition with explicit True": {
			specs:   []string{"pod-condition:Ready=True"},
			wantLen: 1,
		},
		"valid pod-condition with explicit False": {
			specs:   []string{"pod-condition:NetworkAttached=False"},
			wantLen: 1,
		},
		"valid pod-annotation": {
			specs:   []string{"pod-annotation:k8s.v1.cni.cncf.io/networks-status"},
			wantLen: 1,
		},
		"multiple valid specs": {
			specs:   []string{"pod-ip:ipv6", "pod-condition:Ready", "pod-annotation:my-key"},
			wantLen: 3,
		},
		"missing colon errors": {
			specs:   []string{"pod-ip"},
			wantErr: true,
		},
		"empty value after colon errors": {
			specs:   []string{"pod-ip:"},
			wantErr: true,
		},
		"unknown type errors": {
			specs:   []string{"pod-label:app=foo"},
			wantErr: true,
		},
		"pod-ip invalid family errors": {
			specs:   []string{"pod-ip:dual-stack"},
			wantErr: true,
		},
		"pod-condition empty type errors": {
			specs:   []string{"pod-condition:=True"},
			wantErr: true,
		},
		"second spec invalid errors after first is valid": {
			specs:   []string{"pod-ip:ipv6", "unknown:value"},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gates, err := Parse(tc.specs)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, gates, tc.wantLen)
		})
	}
}

func Test_podIPGate(t *testing.T) {
	tests := map[string]struct {
		family    string
		podIPs    []corev1.PodIP
		wantReady bool
		wantMsg   string
	}{
		"any: passes when IPv4 present": {
			family:    "any",
			podIPs:    []corev1.PodIP{{IP: "10.0.0.1"}},
			wantReady: true,
		},
		"any: passes when IPv6 present": {
			family:    "any",
			podIPs:    []corev1.PodIP{{IP: "fd00::1"}},
			wantReady: true,
		},
		"any: fails when no IPs": {
			family:    "any",
			podIPs:    nil,
			wantReady: false,
			wantMsg:   "pod has no any address yet",
		},
		"ipv4: passes with IPv4": {
			family:    "ipv4",
			podIPs:    []corev1.PodIP{{IP: "10.0.0.1"}},
			wantReady: true,
		},
		"ipv4: fails with only IPv6": {
			family:    "ipv4",
			podIPs:    []corev1.PodIP{{IP: "fd00::1"}},
			wantReady: false,
			wantMsg:   "pod has no ipv4 address yet",
		},
		"ipv4: passes when both IPv4 and IPv6 present": {
			family:    "ipv4",
			podIPs:    []corev1.PodIP{{IP: "fd00::1"}, {IP: "10.0.0.1"}},
			wantReady: true,
		},
		"ipv6: passes with IPv6": {
			family:    "ipv6",
			podIPs:    []corev1.PodIP{{IP: "fd00::1"}},
			wantReady: true,
		},
		"ipv6: fails with only IPv4": {
			family:    "ipv6",
			podIPs:    []corev1.PodIP{{IP: "10.0.0.1"}},
			wantReady: false,
			wantMsg:   "pod has no ipv6 address yet",
		},
		"ipv6: passes when both IPv4 and IPv6 present": {
			family:    "ipv6",
			podIPs:    []corev1.PodIP{{IP: "10.0.0.1"}, {IP: "fd00::1"}},
			wantReady: true,
		},
		"ipv6: fails when no IPs": {
			family:    "ipv6",
			podIPs:    nil,
			wantReady: false,
			wantMsg:   "pod has no ipv6 address yet",
		},
		"invalid IP string in PodIPs is skipped": {
			family:    "ipv4",
			podIPs:    []corev1.PodIP{{IP: "not-an-ip"}, {IP: "10.0.0.1"}},
			wantReady: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gate, err := podIPGate(tc.family)
			require.NoError(t, err)

			pod := &corev1.Pod{Status: corev1.PodStatus{PodIPs: tc.podIPs}}
			ready, reason := gate(pod)

			assert.Equal(t, tc.wantReady, ready)
			if !tc.wantReady {
				assert.Equal(t, tc.wantMsg, reason)
			}
		})
	}
}

func Test_podIPGate_invalidFamily(t *testing.T) {
	_, err := podIPGate("dual-stack")
	assert.Error(t, err)
}

func Test_podConditionGate(t *testing.T) {
	tests := map[string]struct {
		spec       string
		conditions []corev1.PodCondition
		wantReady  bool
		wantMsg    string
	}{
		"passes when condition type matches and status is True (default)": {
			spec: "Ready",
			conditions: []corev1.PodCondition{
				{Type: "Ready", Status: corev1.ConditionTrue},
			},
			wantReady: true,
		},
		"passes when explicit True status matches": {
			spec: "Ready=True",
			conditions: []corev1.PodCondition{
				{Type: "Ready", Status: corev1.ConditionTrue},
			},
			wantReady: true,
		},
		"passes when explicit False status matches": {
			spec: "Degraded=False",
			conditions: []corev1.PodCondition{
				{Type: "Degraded", Status: corev1.ConditionFalse},
			},
			wantReady: true,
		},
		"fails when condition type matches but status differs": {
			spec: "Ready",
			conditions: []corev1.PodCondition{
				{Type: "Ready", Status: corev1.ConditionFalse},
			},
			wantReady: false,
			wantMsg:   `pod condition Ready is "False", want "True"`,
		},
		"fails when condition type not present": {
			spec:       "NetworkAttached",
			conditions: nil,
			wantReady:  false,
			wantMsg:    `pod condition "NetworkAttached" not yet present`,
		},
		"fails when conditions list has unrelated entries only": {
			spec: "NetworkAttached",
			conditions: []corev1.PodCondition{
				{Type: "Ready", Status: corev1.ConditionTrue},
			},
			wantReady: false,
			wantMsg:   `pod condition "NetworkAttached" not yet present`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gate, err := podConditionGate(tc.spec)
			require.NoError(t, err)

			pod := &corev1.Pod{Status: corev1.PodStatus{Conditions: tc.conditions}}
			ready, reason := gate(pod)

			assert.Equal(t, tc.wantReady, ready)
			if !tc.wantReady {
				assert.Equal(t, tc.wantMsg, reason)
			}
		})
	}
}

func Test_podConditionGate_emptyType(t *testing.T) {
	_, err := podConditionGate("=True")
	assert.Error(t, err)
}

// Status values from the flag must be matched case-insensitively against
// corev1.ConditionStatus ("True", "False", "Unknown"). A user passing =true
// or =TRUE should find the gate behaves identically to =True.
func Test_podConditionGate_caseInsensitiveStatus(t *testing.T) {
	tests := map[string]struct {
		spec string
	}{
		"lowercase true should match corev1.ConditionTrue":  {spec: "Ready=true"},
		"uppercase TRUE should match corev1.ConditionTrue":  {spec: "Ready=TRUE"},
		"lowercase false should match corev1.ConditionFalse": {spec: "Degraded=false"},
	}

	conditions := []corev1.PodCondition{
		{Type: "Ready", Status: corev1.ConditionTrue},
		{Type: "Degraded", Status: corev1.ConditionFalse},
	}
	pod := &corev1.Pod{Status: corev1.PodStatus{Conditions: conditions}}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gate, err := podConditionGate(tc.spec)
			require.NoError(t, err)
			ready, reason := gate(pod)
			assert.True(t, ready, "expected gate to pass but got reason: %s", reason)
		})
	}
}

func Test_podAnnotationGate(t *testing.T) {
	tests := map[string]struct {
		key        string
		annotations map[string]string
		wantReady  bool
		wantMsg    string
	}{
		"passes when annotation key is present": {
			key:         "k8s.v1.cni.cncf.io/networks-status",
			annotations: map[string]string{"k8s.v1.cni.cncf.io/networks-status": "[...]"},
			wantReady:   true,
		},
		"fails when annotation key is present but value is empty": {
			key:         "my-key",
			annotations: map[string]string{"my-key": ""},
			wantReady:   false,
			wantMsg:     `pod does not yet have annotation "my-key" with a non-empty value`,
		},
		"fails when annotation key is absent": {
			key:         "my-key",
			annotations: map[string]string{"other-key": "value"},
			wantReady:   false,
			wantMsg:     `pod does not yet have annotation "my-key" with a non-empty value`,
		},
		"fails when pod has no annotations": {
			key:         "my-key",
			annotations: nil,
			wantReady:   false,
			wantMsg:     `pod does not yet have annotation "my-key" with a non-empty value`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gate, err := podAnnotationGate(tc.key)
			require.NoError(t, err)

			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: tc.annotations}}
			ready, reason := gate(pod)

			assert.Equal(t, tc.wantReady, ready)
			if !tc.wantReady {
				assert.Equal(t, tc.wantMsg, reason)
			}
		})
	}
}

func Test_podAnnotationGate_emptyKey(t *testing.T) {
	_, err := podAnnotationGate("")
	assert.Error(t, err)
}

// Multus and some CNI plugins write the annotation key immediately with an
// empty value and fill it in asynchronously once the interface is attached.
// The gate must not pass until the annotation has a non-empty value.
func Test_podAnnotationGate_emptyValueShouldNotPass(t *testing.T) {
	gate, err := podAnnotationGate("k8s.v1.cni.cncf.io/networks-status")
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks-status": "", // key present, value empty
			},
		},
	}

	ready, reason := gate(pod)
	assert.False(t, ready, "gate should not pass when annotation value is empty, but got reason: %s", reason)
}

func Test_NewReadyToRequestFunc(t *testing.T) {
	const (
		podName      = "my-pod"
		podNamespace = "my-namespace"
	)

	validMeta := metadata.Metadata{
		VolumeContext: map[string]string{
			csiapi.K8sVolumeContextKeyPodName:      podName,
			csiapi.K8sVolumeContextKeyPodNamespace: podNamespace,
		},
	}

	podWithIPv6 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: podNamespace},
		Status:     corev1.PodStatus{PodIPs: []corev1.PodIP{{IP: "10.0.0.1"}, {IP: "fd00::1"}}},
	}
	podNoIPs := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: podNamespace},
		Status:     corev1.PodStatus{},
	}

	tests := map[string]struct {
		meta       metadata.Metadata
		pod        *corev1.Pod
		specs      []string
		wantReady  bool
		wantReason string
	}{
		"missing pod name in VolumeContext returns false": {
			meta: metadata.Metadata{
				VolumeContext: map[string]string{
					csiapi.K8sVolumeContextKeyPodNamespace: podNamespace,
				},
			},
			pod:       podWithIPv6,
			specs:     []string{"pod-ip:ipv6"},
			wantReady: false,
		},
		"missing pod namespace in VolumeContext returns false": {
			meta: metadata.Metadata{
				VolumeContext: map[string]string{
					csiapi.K8sVolumeContextKeyPodName: podName,
				},
			},
			pod:       podWithIPv6,
			specs:     []string{"pod-ip:ipv6"},
			wantReady: false,
		},
		"pod not found returns false": {
			meta:      validMeta,
			pod:       nil, // not registered in fake client
			specs:     []string{"pod-ip:ipv6"},
			wantReady: false,
		},
		"single gate passes": {
			meta:      validMeta,
			pod:       podWithIPv6,
			specs:     []string{"pod-ip:ipv6"},
			wantReady: true,
		},
		"single gate fails": {
			meta:      validMeta,
			pod:       podNoIPs,
			specs:     []string{"pod-ip:ipv6"},
			wantReady: false,
		},
		"all gates pass": {
			meta: metadata.Metadata{
				VolumeContext: map[string]string{
					csiapi.K8sVolumeContextKeyPodName:      podName,
					csiapi.K8sVolumeContextKeyPodNamespace: podNamespace,
				},
			},
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        podName,
					Namespace:   podNamespace,
					Annotations: map[string]string{"my-annotation": "present"},
				},
				Status: corev1.PodStatus{
					PodIPs: []corev1.PodIP{{IP: "fd00::1"}},
				},
			},
			specs:     []string{"pod-ip:ipv6", "pod-annotation:my-annotation"},
			wantReady: true,
		},
		"first gate fails, second would pass": {
			meta:      validMeta,
			pod:       podNoIPs,
			specs:     []string{"pod-ip:ipv6", "pod-annotation:my-annotation"},
			wantReady: false,
		},
		"first gate passes, second gate fails": {
			meta: metadata.Metadata{
				VolumeContext: map[string]string{
					csiapi.K8sVolumeContextKeyPodName:      podName,
					csiapi.K8sVolumeContextKeyPodNamespace: podNamespace,
				},
			},
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: podNamespace},
				Status:     corev1.PodStatus{PodIPs: []corev1.PodIP{{IP: "fd00::1"}}},
			},
			specs:     []string{"pod-ip:ipv6", "pod-annotation:missing-annotation"},
			wantReady: false,
		},
		"all gates fail returns all reasons": {
			meta:  validMeta,
			pod:   &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: podNamespace}},
			specs: []string{"pod-ip:ipv6", "pod-annotation:missing-annotation"},
			wantReady:  false,
			wantReason: `pod has no ipv6 address yet; pod does not yet have annotation "missing-annotation" with a non-empty value`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var client *fakeclient.Clientset
			if tc.pod != nil {
				client = fakeclient.NewSimpleClientset(tc.pod)
			} else {
				client = fakeclient.NewSimpleClientset()
			}

			gates, err := Parse(tc.specs)
			require.NoError(t, err)

			fn := NewReadyToRequestFunc(client, gates)
			ready, reason := fn(tc.meta)
			assert.Equal(t, tc.wantReady, ready)
			if tc.wantReason != "" {
				assert.Equal(t, tc.wantReason, reason)
			}
		})
	}
}
