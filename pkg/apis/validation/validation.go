package validation

import (
	"errors"
	"fmt"
	"strings"
	"time"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func ValidateAttributes(attr map[string]string) error {
	var errs []string

	if len(attr[csiapi.IssuerNameKey]) == 0 {
		errs = append(errs, fmt.Sprintf("%s field required", csiapi.IssuerNameKey))
	}

	errs = boolValue(attr[csiapi.IsCAKey], csiapi.IsCAKey, errs)

	errs = durationParse(attr[csiapi.DurationKey], csiapi.DurationKey, errs)

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
