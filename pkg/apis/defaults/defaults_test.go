/*
Copyright 2022 The cert-manager Authors.

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

package defaults

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pkcs12Values(t *testing.T) {
	tests := map[string]struct {
		input     map[string]string
		expOutput map[string]string
	}{
		"if attributes are empty, expect no PKCS12 attributes": {
			input:     map[string]string{},
			expOutput: map[string]string{},
		},
		"if PKCS12 enable attribute present, expect PKCS12 attributes present": {
			input: map[string]string{
				"csi.cert-manager.io/pkcs12-enable": "foo",
			},
			expOutput: map[string]string{
				"csi.cert-manager.io/pkcs12-enable":   "foo",
				"csi.cert-manager.io/pkcs12-filename":     "keystore.p12",
				"csi.cert-manager.io/pkcs12-password": "",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			out := test.input
			setDefaultKeyStorePKCS12(out)
			assert.Equal(t, test.expOutput, out)
		})
	}
}
