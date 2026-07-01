package pdfops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// FileInfo holds metadata about a processed PDF output file.
type FileInfo struct {
	Name      string
	Bytes     int64
	PageCount int
	Location  string
}

// SplitJob represents one output segment of a split operation.
type SplitJob struct {
	Label string
	Pages []int
}

// SplitOptions configures extract and parity behaviour for Split.
type SplitOptions struct {
	Extract bool
	Odd     bool
	Even    bool
}

func (r *FileInfo) PrintReport(w io.Writer) {
	var fileSize string
	sizeMB := float64(r.Bytes) / (1024 * 1024)
	sizeKB := float64(r.Bytes) / 1024

	if sizeMB >= 1 {
		fileSize = fmt.Sprintf("%.2f MB", sizeMB)
	} else {
		fileSize = fmt.Sprintf("%.2f KB", sizeKB)
	}

	fmt.Fprintf(w, "Name: %s\nSize: %s\nPages: %d\nLocation: %s\n", r.Name, fileSize, r.PageCount, r.Location)
}

func newConfig(password string) *model.Configuration {
	if password == "" {
		return nil
	}
	conf := model.NewDefaultConfiguration()
	conf.UserPW = password
	conf.OwnerPW = password
	return conf
}

func pageCount(path string, conf *model.Configuration) (int, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return 0, err
	}
	defer root.Close()

	f, err := root.Open(filepath.Base(path))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return api.PageCount(f, conf)
}

// PageCount returns the number of pages in the PDF at path.
// Pass an empty string for password if the file is not protected.
func PageCount(path, password string) (int, error) {
	return pageCount(path, newConfig(password))
}

func getFileInfo(output string, conf *model.Configuration) (*FileInfo, error) {
	info, err := os.Stat(output)
	if err != nil {
		return nil, fmt.Errorf("getting file info error: %w", err)
	}

	path, err := filepath.Abs(output)
	if err != nil {
		return nil, fmt.Errorf("error in getting path: %w", err)
	}

	count, err := pageCount(output, conf)
	if err != nil {
		return nil, fmt.Errorf("page count error: %w", err)
	}

	return &FileInfo{
		Name:      info.Name(),
		Bytes:     info.Size(),
		PageCount: count,
		Location:  filepath.Dir(path),
	}, nil
}

// GetFileInfo returns metadata for the PDF at output.
// Pass an empty string for password if the file is not protected.
func GetFileInfo(output, password string) (*FileInfo, error) {
	return getFileInfo(output, newConfig(password))
}
