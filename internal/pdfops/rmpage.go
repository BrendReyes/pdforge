package pdfops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// RemovePages removes the pages identified by pageSpec from input and writes
// the result to output. pageSpec accepts single pages, ranges, and combinations
// (e.g. "3", "1-4", "2,6-9").
// Pass an empty string for password if the file is not protected.
func RemovePages(input, output, pageSpec, password string) (*FileInfo, error) {
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

	selectedPages, err := api.ParsePageSelection(pageSpec)
	if err != nil {
		return nil, fmt.Errorf("invalid page specification: %w", err)
	}

	if err := api.RemovePagesFile(input, output, selectedPages, conf); err != nil {
		return nil, err
	}

	return getFileInfo(output, conf)
}
