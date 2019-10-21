/*
Copyright 2019 The Jetstack cert-manager contributors.

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

package config

import (
	"errors"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/jetstack/cert-manager-csi/test/e2e/environment"
)

type Config struct {
	KubeConfigPath string
	Kubectl        string

	Ginkgo Ginkgo

	// If Cleanup is true, addons will be cleaned up both before and after provisioning
	Cleanup bool

	// RepoRoot is used as the base path for any parts of the framework that
	// require access to repo files, such as Helm charts and test fixtures.
	RepoRoot string

	Environment *environment.Environment
}

func (c *Config) Validate() error {
	var errs []error
	if c.KubeConfigPath == "" {
		errs = append(errs, errors.New("kubeconfig path not defined"))
	}
	if c.RepoRoot == "" {
		errs = append(errs, errors.New("repo root not defined"))
	}

	errs = append(errs, c.Ginkgo.Validate()...)

	return utilerrors.NewAggregate(errs)
}
