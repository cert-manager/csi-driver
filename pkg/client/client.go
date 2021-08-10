/*
Copyright 2021 The Jetstack cert-manager contributors.

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

package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cert-manager/csi-lib/manager"
	"github.com/cert-manager/csi-lib/metadata"
	cmclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"k8s.io/client-go/rest"
)

const (
	csiServiceAccountTokensKey = "csi.storage.k8s.io/serviceAccount.tokens"
)

// ClientForMetadataFunc will, if configured, return a csi-lib
// ClientForMetadataFunc that returns a cert-manager client which authenticates
// using the request token's credentials.
func ClientForMetadataFunc(restConfig *rest.Config, useRequestToken bool) manager.ClientForMetadataFunc {
	if !useRequestToken {
		return nil
	}

	return func(meta metadata.Metadata) (cmclient.Interface, error) {
		apiToken, err := apiTokenFromMetadata(meta)
		if err != nil {
			return nil, err
		}

		return cmclient.NewForConfig(&rest.Config{
			Host:            restConfig.Host,
			TLSClientConfig: restConfig.TLSClientConfig,
			BearerToken:     apiToken,
		})
	}
}

// apiTokenFromMetadata returns the empty audience service account token from
// the volume attributes contained in the metadata.
func apiTokenFromMetadata(meta metadata.Metadata) (string, error) {
	tokens := make(map[string]struct {
		Token string `json:"token"`
	})

	err := json.Unmarshal([]byte(meta.VolumeContext[csiServiceAccountTokensKey]), &tokens)
	if err != nil {
		return "", fmt.Errorf("failed to parse service account tokens from CSI volume context, driver driver likely doesn't have token requests enabled: %w",
			err)
	}

	apiToken, ok := tokens[""]
	if !ok || len(apiToken.Token) == 0 {
		return "", errors.New("empty or Kubernetes API audience service account token doesn't exist in CSI volume context, driver likely doesn't have an empty audience token request configured")
	}

	return apiToken.Token, nil
}
