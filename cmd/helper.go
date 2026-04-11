package cmd

import (
	"os"
	"fmt"	
	"path/filepath"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type FileInfoReport struct {
	Name 		string
	Bytes 		int64
	PageCount 	int
	Location 	string
}

func GetFileInfo(output string) (*FileInfoReport, error) {
	info, err := os.Stat(output) 
	if err != nil {
		return nil, fmt.Errorf("Getting file info error: %w", err)
	}

	path, err := filepath.Abs(output)
	if err != nil {
		return nil, fmt.Errorf("error in getting path, err: %w", err)
	}

	pageCount, err := api.PageCountFile(output)
	if err != nil {
		return nil, fmt.Errorf("Page count error: %w", err)
	}

	return &FileInfoReport{
        Name:      info.Name(),
		Bytes: 	   info.Size(),
        PageCount: pageCount,
        Location:  filepath.Dir(path),
    }, nil

}

func (r *FileInfoReport) PrintReport() {
	var fileSize string
	sizeMB := float64(r.Bytes) / (1024 * 1024)
	sizeKB := float64(r.Bytes) / (1024)

	if sizeMB >= 1 {
		fileSize = fmt.Sprintf("%.2f MB", sizeMB)
	} else {
		fileSize = fmt.Sprintf("%.2f KB", sizeKB)
	}

	fmt.Printf("Name: %s\nSize: %s\nPages: %d\nLocation: %s\n", r.Name, fileSize, r.PageCount, r.Location)
}