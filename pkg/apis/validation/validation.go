package validation

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/joshvanl/cert-manager-csi/pkg/apis/v1alpha1"
)

func ValidateAttributes(attr v1alpha1.Attributes) error {
	var errs []string

	if len(attr[v1alpha1.IssuerNameKey]) == 0 {
		errs = append(errs, fmt.Sprintf("%s field required", v1alpha1.IssuerNameKey))
	}

	if len(attr[v1alpha1.CommonNameKey]) == 0 && len(attr[v1alpha1.DNSNamesKey]) == 0 {
		errs = append(errs, fmt.Sprintf("both %s and %s may not be empty",
			v1alpha1.CommonNameKey, v1alpha1.DNSNamesKey))
	}

	errs = boolValue(attr[v1alpha1.IsCAKey], v1alpha1.IsCAKey, errs)

	errs = durationParse(attr[v1alpha1.DurationKey], v1alpha1.DurationKey, errs)

	errs = filepathBreakout(attr[v1alpha1.CertFileKey], v1alpha1.CertFileKey, errs)
	errs = filepathBreakout(attr[v1alpha1.KeyFileKey], v1alpha1.KeyFileKey, errs)

	// TODO (@joshvanl): add better validation for renew before to ensure we
	// don't go into a crazy renew loop
	errs = durationParse(attr[v1alpha1.RenewBeforeKey], v1alpha1.RenewBeforeKey, errs)
	errs = boolValue(attr[v1alpha1.DisableAutoRenewKey], v1alpha1.DisableAutoRenewKey, errs)
	errs = boolValue(attr[v1alpha1.ReusePrivateKey], v1alpha1.ReusePrivateKey, errs)

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func filepathBreakout(s string, k v1alpha1.Attribute, errs []string) []string {
	if strings.Contains(s, "..") {
		errs = append(errs, fmt.Sprintf("%s filepaths may not contain '..'",
			k))
	}

	return errs
}

func durationParse(s string, k v1alpha1.Attribute, errs []string) []string {
	if len(s) == 0 {
		return errs
	}

	if _, err := time.ParseDuration(s); err != nil {
		errs = append(errs, fmt.Sprintf("%s must be a valid duration string: %s",
			k, err))
	}

	return errs
}

func boolValue(s string, k v1alpha1.Attribute, errs []string) []string {
	if len(s) == 0 {
		return errs
	}

	if s != "false" && s != "true" {
		errs = append(errs, fmt.Sprintf("%s may only be set to 'true' for 'false'",
			k))
	}

	return errs
}
