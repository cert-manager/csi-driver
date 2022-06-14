//go:build tools
// +build tools

// This file exists to force 'go mod' to fetch tool dependencies
// See: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package tools

import (
	_ "github.com/norwoodj/helm-docs/cmd/helm-docs"
	_ "helm.sh/helm/v3/cmd/helm"
	_ "sigs.k8s.io/kind"
)
