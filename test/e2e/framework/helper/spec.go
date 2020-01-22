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

package helper

import (
	"fmt"
	"reflect"
	"strings"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func (h *Helper) metaDataMatches(exp, got *csiapi.MetaData) error {
	var errs []string

	if exp.ID != got.ID {
		errs = append(errs, fmt.Sprintf("miss-match id, exp=%s got=%s",
			exp.ID, got.ID))
	}

	if exp.Name != got.Name {
		errs = append(errs, fmt.Sprintf("miss-match name, exp=%s got=%s",
			exp.Name, got.Name))
	}

	if exp.Path != got.Path {
		errs = append(errs, fmt.Sprintf("miss-match path, exp=%s got=%s",
			exp.Path, got.Path))
	}

	if exp.Size != got.Size {
		errs = append(errs, fmt.Sprintf("miss-match size, exp=%d got=%d",
			exp.Size, got.Size))
	}

	if exp.TargetPath != got.TargetPath {
		errs = append(errs, fmt.Sprintf("miss-match targetPath, exp=%s got=%s",
			exp.TargetPath, got.TargetPath))
	}

	if !reflect.DeepEqual(exp.Attributes, got.Attributes) {
		errs = append(errs, fmt.Sprintf("miss-match attributes, exp=%s got=%s",
			exp.Attributes, got.Attributes))
	}

	if len(errs) > 0 {
		return fmt.Errorf("unexpected metadata: %s",
			strings.Join(errs, ", "))
	}

	return nil
}
