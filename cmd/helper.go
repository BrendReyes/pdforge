package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/spf13/cobra"
)

type FileInfoReport struct {
	Name      string
	Bytes     int64
	PageCount int
	Location  string
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
		Bytes:     info.Size(),
		PageCount: pageCount,
		Location:  filepath.Dir(path),
	}, nil

}

func (r *FileInfoReport) PrintReport(w io.Writer) {
	var fileSize string
	sizeMB := float64(r.Bytes) / (1024 * 1024)
	sizeKB := float64(r.Bytes) / (1024)

	if sizeMB >= 1 {
		fileSize = fmt.Sprintf("%.2f MB", sizeMB)
	} else {
		fileSize = fmt.Sprintf("%.2f KB", sizeKB)
	}

	fmt.Fprintf(w, "Name: %s\nSize: %s\nPages: %d\nLocation: %s\n", r.Name, fileSize, r.PageCount, r.Location)
}

// this is to avoid overwriting a file with same name
// e.g. merged (1).pdf, merged (2).pdf
func resolveOutputPath(output string) string {
	_, err := os.Stat(output)
	if os.IsNotExist(err) {
		return output
	}

	ext := filepath.Ext(output)
	base := strings.TrimSuffix(output, ext)
	counter := 1

	for {
		candidate := fmt.Sprintf("%s (%d)%s", base, counter, ext)
		_, err := os.Stat(candidate)
		if os.IsNotExist(err) {
			return candidate
		}
		counter++
	}
}

func ensureOutputDirectory(cmd *cobra.Command, dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("'%s' is not a directory", dir)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("unable to access directory '%s': %w", dir, err)
	}

	create, promptErr := promptYesNo(cmd, fmt.Sprintf("Directory '%s' does not exist. Create it now? [y/N]: ", dir))
	if promptErr != nil {
		return promptErr
	}

	if !create {
		return fmt.Errorf("directory '%s' does not exist", dir)
	}

	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, mkErr)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created directory: %s\n", dir)
	return nil
}

func promptYesNo(cmd *cobra.Command, message string) (bool, error) {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false, nil
	}

	if (stdinInfo.Mode() & os.ModeCharDevice) == 0 {
		return false, nil
	}

	fmt.Fprint(cmd.OutOrStdout(), message)
	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

