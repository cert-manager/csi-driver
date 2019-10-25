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
