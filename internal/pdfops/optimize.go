package pdfops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Optimize compresses input and writes the result to output.
// Pass an empty string for password if the file is not protected.
func Optimize(input, output, password string) (*FileInfo, error) {
	if strings.ToLower(filepath.Ext(input)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
	}

	if strings.ToLower(filepath.Ext(output)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(output))
	}

	conf := newConfig(password)

	if err := api.ValidateFile(input, conf); err != nil {
		return nil, fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
	}

	if err := api.OptimizeFile(input, output, conf); err != nil {
		return nil, err
	}

	return getFileInfo(output, conf)
}
