package pdfops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Merge combines inputs into a single PDF written to output.
// Pass an empty string for password if the files are not protected.
func Merge(inputs []string, output, password string) (*FileInfo, error) {
	if strings.ToLower(filepath.Ext(output)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(output))
	}

	conf := newConfig(password)

	for _, item := range inputs {
		if strings.ToLower(filepath.Ext(item)) != ".pdf" {
			return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(item))
		}

		if err := api.ValidateFile(item, conf); err != nil {
			return nil, fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(item), err)
		}
	}

	if err := api.MergeCreateFile(inputs, output, false, conf); err != nil {
		return nil, err
	}

	return getFileInfo(output, conf)
}
