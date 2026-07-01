package pdfops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Convert imports one or more image files into a single PDF written to output.
// Supported image formats: JPG, PNG, WEBP, TIFF, TIF.
func Convert(inputs []string, output string) (*FileInfo, error) {
	if strings.ToLower(filepath.Ext(output)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(output))
	}

	for _, item := range inputs {
		ftype := strings.ToLower(filepath.Ext(item))
		if ftype != ".png" && ftype != ".jpg" && ftype != ".webp" && ftype != ".tiff" && ftype != ".tif" {
			return nil, fmt.Errorf("the file '%s' is invalid, must be supported image file (JPG, PNG, WEBP, TIFF, TIF)", filepath.Base(item))
		}

		if _, err := os.Stat(item); err != nil {
			return nil, fmt.Errorf("invalid image '%s': \n%v", filepath.Base(item), err)
		}
	}

	if err := api.ImportImagesFile(inputs, output, nil, nil); err != nil {
		return nil, err
	}

	return getFileInfo(output, nil)
}
