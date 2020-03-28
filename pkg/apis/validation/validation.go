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

package validation

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cmapiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func ValidateAttributes(attr map[string]string) error {
	var errs []string

	if len(attr[csiapi.IssuerNameKey]) == 0 {
		errs = append(errs, fmt.Sprintf("%s field required", csiapi.IssuerNameKey))
	}

	errs = boolValue(attr[csiapi.IsCAKey], csiapi.IsCAKey, errs)

	errs = durationParse(attr[csiapi.DurationKey], csiapi.DurationKey, errs)

	errs = keyUsages(attr[csiapi.KeyUsagesKey], errs)

	errs = filepathBreakout(attr[csiapi.CAFileKey], csiapi.CAFileKey, errs)
	errs = filepathBreakout(attr[csiapi.CertFileKey], csiapi.CertFileKey, errs)
	errs = filepathBreakout(attr[csiapi.KeyFileKey], csiapi.KeyFileKey, errs)

	errs = durationParse(attr[csiapi.RenewBeforeKey], csiapi.RenewBeforeKey, errs)
	errs = boolValue(attr[csiapi.DisableAutoRenewKey], csiapi.DisableAutoRenewKey, errs)
	errs = boolValue(attr[csiapi.ReusePrivateKey], csiapi.ReusePrivateKey, errs)

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func keyUsages(ss string, errs []string) []string {
	if len(ss) == 0 {
		return errs
	}

	usages := strings.Split(ss, ",")

	for _, usage := range usages {
		if _, ok := cmapiutil.ExtKeyUsageType(cmapi.KeyUsage(usage)); !ok {
			if _, ok := cmapiutil.KeyUsageType(cmapi.KeyUsage(usage)); !ok {
				errs = append(errs, fmt.Sprintf("%q is not a valid key usage", usage))
			}
		}
	}

	return errs
}

func filepathBreakout(s, k string, errs []string) []string {
	if strings.Contains(s, "..") {
		errs = append(errs, fmt.Sprintf("%s filepaths may not contain '..'",
			k))
	}

	return errs
}

func durationParse(s, k string, errs []string) []string {
	if len(s) == 0 {
		return errs
	}

	if _, err := time.ParseDuration(s); err != nil {
		errs = append(errs, fmt.Sprintf("%s must be a valid duration string: %s",
			k, err))
	}

	return errs
}

func boolValue(s, k string, errs []string) []string {
	if len(s) == 0 {
		return errs
	}

	if s != "false" && s != "true" {
		errs = append(errs, fmt.Sprintf("%s may only be set to 'true' for 'false'",
			k))
	}

	return errs
}
