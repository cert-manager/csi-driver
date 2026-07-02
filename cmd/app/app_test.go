/*
Copyright 2026 The cert-manager Authors.

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

package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cert-manager/csi-driver/cmd/app/options"
)

func TestValidateGateBackoff(t *testing.T) {
	tests := map[string]struct {
		opts    options.Options
		wantErr string
	}{
		"valid defaults": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   0.5,
				GateBackoffCap:      10 * time.Second,
			},
		},
		"zero duration": {
			opts: options.Options{
				GateBackoffCap: time.Second,
			},
			wantErr: "--gate-backoff-duration must be > 0",
		},
		"negative duration": {
			opts: options.Options{
				GateBackoffDuration: -time.Second,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-duration must be > 0",
		},
		"factor below 1": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   0.5,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-factor must be >= 1",
		},
		"jitter negative": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   -0.1,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-jitter must be in [0, 1]",
		},
		"jitter above 1": {
			opts: options.Options{
				GateBackoffDuration: time.Second,
				GateBackoffFactor:   2,
				GateBackoffJitter:   1.1,
				GateBackoffCap:      time.Second,
			},
			wantErr: "--gate-backoff-jitter must be in [0, 1]",
		},
		"cap below duration": {
			opts: options.Options{
				GateBackoffDuration: 30 * time.Second,
				GateBackoffFactor:   1,
				GateBackoffCap:      10 * time.Second,
			},
			wantErr: "--gate-backoff-cap (10s) must be >= --gate-backoff-duration (30s)",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateGateBackoff(&test.opts)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}
