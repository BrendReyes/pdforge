package cmd

import (
	"os"
	"fmt"	
	"path/filepath"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func printFileInfo(output string) error {
	info, err := os.Stat(output) 
	if err != nil {
		return fmt.Errorf("Getting file info error: %w", err)
	}
	megabytes := float64(info.Size()) / (1024 * 1024)

	path, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("error in getting path, err: %w", err)
	}

	pageCount, err := api.PageCountFile(output)
	if err != nil {
		return fmt.Errorf("Page count error: %w", err)
	}

	fmt.Printf("Name: %s\nSize: %.2f MB\nTotal Pages: %d\nLocation: %s\n", info.Name(), megabytes, pageCount, filepath.Dir(path))

	return nil
}