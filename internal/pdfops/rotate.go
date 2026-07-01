package pdfops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Rotate rotates pages of input by degrees (must be a non-zero multiple of 90)
// and writes the result to output. pageSpec selects which pages to rotate;
// pass an empty string to rotate all pages.
// Pass an empty string for password if the file is not protected.
func Rotate(input, output string, degrees int, pageSpec, password string) (*FileInfo, error) {
	if degrees == 0 || degrees%90 != 0 {
		return nil, fmt.Errorf("invalid rotation %d: must be a non-zero multiple of 90", degrees)
	}

	if strings.ToLower(filepath.Ext(input)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
	}

	if strings.ToLower(filepath.Ext(output)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(output))
	}

	conf := newConfig(password)

	if err := api.ValidateFile(input, conf); err != nil {
		errText := strings.ToLower(err.Error())
		if conf == nil && (strings.Contains(errText, "password") || strings.Contains(errText, "encrypt")) {
			return nil, fmt.Errorf("'%s' is password protected: provide the password with --password", filepath.Base(input))
		}
		return nil, fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
	}

	var selectedPages []string
	if spec := strings.TrimSpace(pageSpec); spec != "" {
		var err error
		selectedPages, err = api.ParsePageSelection(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid page specification: %w", err)
		}
	}

	if err := api.RotateFile(input, output, degrees, selectedPages, conf); err != nil {
		return nil, err
	}

	return getFileInfo(output, conf)
}
