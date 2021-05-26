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

package main

import (
	"fmt"
	"os"
	"sigs.k8s.io/kind/pkg/cluster"

	"github.com/jetstack/cert-manager-csi/test/e2e/environment"
)

func main() {
	if len(os.Args) != 2 {
		errExit(fmt.Errorf("expecting 2 arguments, got=%d",
			len(os.Args)))
	}

	switch os.Args[1] {
	case "create":
		create()
	case "destroy":
		destroy()
	default:
		errExit(fmt.Errorf("unexpected argument %q, expecting %q or %q",
			os.Args[1], "create", "destroy"))
	}
}

func create() {
	env, err := environment.Create(1, 1)
	errExit(err)

	fmt.Printf("dev environment created.\nexport KUBECONFIG=%s\n",
		env.KubeConfigPath())
}

func destroy() {
	errExit(cluster.NewProvider(cluster.ProviderWithDocker()).Delete("cert-manager-csi-e2e", ""))
	fmt.Printf("dev environment destroyed.\n")
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
