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

package suite

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"

	"github.com/jetstack/cert-manager-csi/test/e2e/environment"
	"github.com/jetstack/cert-manager-csi/test/e2e/framework"
)

var (
	env *environment.Environment
	cfg = framework.DefaultConfig
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	env, err = environment.Create(1, 3)
	if err != nil {
		framework.Failf("Error provisioning environment: %v", err)
	}

	cfg.KubeConfigPath = env.KubeConfigPath()
	cfg.Kubectl = filepath.Join(env.RootPath(), "bin", "kubectl")
	cfg.RepoRoot = env.RootPath()
	cfg.Environment = env

	if err := framework.DefaultConfig.Validate(); err != nil {
		framework.Failf("Invalid test config: %v", err)
	}

	return nil
}, func([]byte) {
})

var globalLogs map[string]string

var _ = SynchronizedAfterSuite(func() {},
	func() {
		if env != nil {
			if err := env.Destory(); err != nil {
				framework.Failf("Failed to destory environment: %s", err)
			}
		}
	},
)
