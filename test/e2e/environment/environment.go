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

package environment

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Environment struct {
	kubeconfigPath string
	client         kubernetes.Interface
	clusterName    string
}

func Create(kubeconfigPath, clusterName string) (*Environment, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Environment{
		kubeconfigPath: kubeconfigPath,
		client:         client,
		clusterName:    clusterName,
	}, nil
}

func (e *Environment) KubeClient() kubernetes.Interface {
	return e.client
}

func (e *Environment) KubeConfigPath() string {
	return e.kubeconfigPath
}

func (e *Environment) ClusterName() string {
	return e.clusterName
}
