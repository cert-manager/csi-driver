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

package validation

import (
	"strings"
	"time"

	cmapiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	csiapi "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
)

// ValidateAttributes validates that the attributes provided
func ValidateAttributes(attr map[string]string) field.ErrorList {
	var el field.ErrorList

	path := field.NewPath("volumeAttributes")

	if len(attr[csiapi.IssuerNameKey]) == 0 {
		el = append(el, field.Required(path.Child(csiapi.IssuerNameKey), "issuer-name is a required field"))
	}

	el = append(el, boolValue(path.Child(csiapi.IsCAKey), attr[csiapi.IsCAKey])...)

	el = append(el, durationParse(path.Child(csiapi.DurationKey), attr[csiapi.DurationKey])...)

	el = append(el, keyUsages(path.Child(csiapi.KeyUsagesKey), attr[csiapi.KeyUsagesKey])...)

	el = append(el, filepathBreakout(path.Child(csiapi.CAFileKey), attr[csiapi.CAFileKey])...)
	el = append(el, filepathBreakout(path.Child(csiapi.CertFileKey), attr[csiapi.CertFileKey])...)
	el = append(el, filepathBreakout(path.Child(csiapi.KeyFileKey), attr[csiapi.KeyFileKey])...)

	el = append(el, durationParse(path.Child(csiapi.RenewBeforeKey), attr[csiapi.RenewBeforeKey])...)
	el = append(el, boolValue(path.Child(csiapi.ReusePrivateKey), attr[csiapi.ReusePrivateKey])...)

	// If there are errors, then return not approved and the aggregated errors.
	if len(el) > 0 {
		return el
	}

	return nil
}

func keyUsages(path *field.Path, ss string) field.ErrorList {
	if len(ss) == 0 {
		return nil
	}

	usages := strings.Split(ss, ",")

	var el field.ErrorList
	for _, usage := range usages {
		trimedUsage := strings.TrimSpace(usage)
		if _, ok := cmapiutil.ExtKeyUsageType(cmapi.KeyUsage(trimedUsage)); !ok {
			if _, ok := cmapiutil.KeyUsageType(cmapi.KeyUsage(trimedUsage)); !ok {
				el = append(el, field.Invalid(path, trimedUsage, "not a valid key usage"))
			}
		}
	}

	return el
}

func filepathBreakout(path *field.Path, s string) field.ErrorList {
	if strings.Contains(s, "..") {
		return field.ErrorList{field.Invalid(path, s, `filepaths may not contain ".."`)}
	}
	return nil
}

func durationParse(path *field.Path, s string) field.ErrorList {
	if len(s) == 0 {
		return nil
	}
	if _, err := time.ParseDuration(s); err != nil {
		return field.ErrorList{field.Invalid(path, s, "must be a valid duration string: "+err.Error())}
	}
	return nil
}

func boolValue(path *field.Path, s string) field.ErrorList {
	if len(s) == 0 {
		return nil
	}
	if s != "false" && s != "true" {
		return field.ErrorList{field.Invalid(path, s, `may only accept values of "true" or "false"`)}
	}
	return nil
}
