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

// Package readinessgate provides concrete implementations of
// manager.ReadyToRequestFunc that defer certificate issuance until specific
// pod-level conditions are met. Gate implementations cover the three main
// sources of async pod state: IP assignment (pod-ip), status conditions
// (pod-condition), and annotations (pod-annotation). Multiple gates are
// combined with AND semantics via NewReadyToRequestFunc.
//
// These implementations are intended to be upstreamed into csi-lib once
// stabilised.
package readinessgate

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

// Gate tests a single condition on a pod. Returns (true, "") when satisfied,
// or (false, reason) when the condition is not yet met.
type Gate func(pod *corev1.Pod) (ready bool, reason string)

// Parse parses gate specs of the form "<type>:<value>" into Gate functions.
//
// Each spec must be one of:
//
//	pod-ip:<family>                     family: any | ipv4 | ipv6
//	pod-condition:<Type>[=<Status>]     Status defaults to "True"
//	pod-annotation:<key>                annotation key must be present
//
// Returns an error if any spec is malformed or uses an unsupported type.
func Parse(specs []string) ([]Gate, error) {
	gates := make([]Gate, 0, len(specs))
	for _, spec := range specs {
		gate, err := parse(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid --pod-readiness-gate %q: %w", spec, err)
		}
		gates = append(gates, gate)
	}
	return gates, nil
}

// NewReadyToRequestFunc builds a manager.ReadyToRequestFunc that fetches the
// pod owning the volume and evaluates all gates against it. All gates must
// pass (AND semantics). Intended to be paired with --continue-on-not-ready=true
// so that NodePublishVolume succeeds immediately and cert issuance is retried
// asynchronously until all gates pass.
func NewReadyToRequestFunc(client kubernetes.Interface, gates []Gate) manager.ReadyToRequestFunc {
	return func(meta metadata.Metadata) (bool, string) {
		podName := meta.VolumeContext[csiapi.K8sVolumeContextKeyPodName]
		podNamespace := meta.VolumeContext[csiapi.K8sVolumeContextKeyPodNamespace]
		if podName == "" || podNamespace == "" {
			return false, "pod name or namespace not present in volume context"
		}

		// TODO: replace this live Get with a pod informer scoped to the local node
		// (spec.nodeName field selector). The renewal loop fires every second per
		// managed volume, so this generates one API call per second per pending volume.
		// On a node with many pods awaiting certificates this can exhaust the client's
		// default rate limit (5 QPS) and slow down gate evaluation.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pod, err := client.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("failed to get pod %s/%s: %v", podNamespace, podName, err)
		}

		var reasons []string
		for _, gate := range gates {
			if ok, reason := gate(pod); !ok {
				reasons = append(reasons, reason)
			}
		}
		if len(reasons) > 0 {
			return false, strings.Join(reasons, "; ")
		}
		return true, ""
	}
}

func parse(spec string) (Gate, error) {
	kind, value, ok := strings.Cut(spec, ":")
	if !ok || value == "" {
		return nil, fmt.Errorf("expected <type>:<value>, got %q", spec)
	}
	switch kind {
	case "pod-ip":
		return podIPGate(value)
	case "pod-condition":
		return podConditionGate(value)
	case "pod-annotation":
		return podAnnotationGate(value)
	default:
		return nil, fmt.Errorf("unknown type %q; supported types: pod-ip, pod-condition, pod-annotation", kind)
	}
}

// podIPGate defers issuance until pod.Status.PodIPs contains an address of the
// requested family. Reads the CNI-populated field directly — no custom
// controller needs to write anything.
func podIPGate(family string) (Gate, error) {
	switch family {
	case "any", "ipv4", "ipv6":
	default:
		return nil, fmt.Errorf("pod-ip: unsupported family %q; use any, ipv4, or ipv6", family)
	}
	return func(pod *corev1.Pod) (bool, string) {
		for _, podIP := range pod.Status.PodIPs {
			if ipMatchesFamily(podIP.IP, family) {
				return true, ""
			}
		}
		return false, fmt.Sprintf("pod has no %s address yet", family)
	}, nil
}

// podConditionGate defers issuance until pod.Status.Conditions contains an
// entry with the given type and status. Useful when an external controller
// explicitly signals readiness via a pod condition.
func podConditionGate(value string) (Gate, error) {
	condType, wantStatus, _ := strings.Cut(value, "=")
	if condType == "" {
		return nil, fmt.Errorf("pod-condition: condition type must not be empty")
	}
	if wantStatus == "" {
		wantStatus = string(corev1.ConditionTrue)
	} else {
		// Normalise to canonical casing so that "true", "TRUE", "True" all work.
		switch strings.ToTitle(wantStatus[:1]) + strings.ToLower(wantStatus[1:]) {
		case string(corev1.ConditionTrue):
			wantStatus = string(corev1.ConditionTrue)
		case string(corev1.ConditionFalse):
			wantStatus = string(corev1.ConditionFalse)
		case string(corev1.ConditionUnknown):
			wantStatus = string(corev1.ConditionUnknown)
		default:
			return nil, fmt.Errorf("pod-condition: invalid status %q; must be True, False, or Unknown", wantStatus)
		}
	}
	return func(pod *corev1.Pod) (bool, string) {
		for _, c := range pod.Status.Conditions {
			if string(c.Type) == condType {
				if string(c.Status) == wantStatus {
					return true, ""
				}
				return false, fmt.Sprintf("pod condition %s is %q, want %q", condType, c.Status, wantStatus)
			}
		}
		return false, fmt.Sprintf("pod condition %q not yet present", condType)
	}, nil
}

// podAnnotationGate defers issuance until a specific annotation key is present
// on the pod. Useful for CNI plugins (e.g. Multus) that write network status
// into pod annotations after attaching secondary interfaces.
func podAnnotationGate(key string) (Gate, error) {
	if key == "" {
		return nil, fmt.Errorf("pod-annotation: annotation key must not be empty")
	}
	return func(pod *corev1.Pod) (bool, string) {
		if val, ok := pod.Annotations[key]; ok && val != "" {
			return true, ""
		}
		return false, fmt.Sprintf("pod does not yet have annotation %q with a non-empty value", key)
	}, nil
}

func ipMatchesFamily(ip, family string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	switch family {
	case "any":
		return true
	case "ipv4":
		return parsed.To4() != nil
	case "ipv6":
		return parsed.To4() == nil && parsed.To16() != nil
	default:
		return false
	}
}
