package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// resolveOutputPath avoids overwriting an existing file by appending (1), (2), …
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

	if mkErr := os.MkdirAll(dir, 0o750); mkErr != nil {
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
