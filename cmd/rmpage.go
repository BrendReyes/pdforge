/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// rmpageCmd represents the rmpage command
var rmpageCmd = &cobra.Command{
	Use:   "rmpage [flags] <input_path> [page_selector]",
	Short: "Remove one or more pages from a PDF",
	Long: `Remove pages from a PDF file. Accepts:
	  - Single page: pdforge rmpage input.pdf 8
	  - Flag form: pdforge rmpage input.pdf --page 8
	  - Range: pdforge rmpage input.pdf 1-3
	  - Combination: pdforge rmpage input.pdf 1,6-11,17

If --output is omitted, pdforge writes remove.pdf in the selected directory.
If the output file exists, pdforge auto-increments the filename (example: remove (1).pdf).`,
	Example: `  pdforge rmpage input.pdf 8
  pdforge rmpage input.pdf --page 1-3
  pdforge rmpage input.pdf 1-3 -o cleaned.pdf
  pdforge rmpage input.pdf 1,6-11,17 -d ./out`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		pageFlag, err := cmd.Flags().GetString("page")
		if err != nil {
			return err
		}

		positionalPage := ""
		if len(args) == 2 {
			positionalPage = args[1]
		}

		if positionalPage != "" && strings.TrimSpace(pageFlag) != "" {
			return fmt.Errorf("provide page selector either as positional argument or via --page, not both")
		}

		pageSpec := strings.TrimSpace(pageFlag)
		if pageSpec == "" {
			pageSpec = strings.TrimSpace(positionalPage)
		}
		if pageSpec == "" {
			return fmt.Errorf("missing page selector: provide [page_selector] or --page")
		}

		if strings.ToLower(filepath.Ext(input)) != ".pdf" {
			return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
		}

		if err := api.ValidateFile(input, nil); err != nil {
			return fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
		}

		_, err = parsePageSpecification(pageSpec)
		if err != nil {
			return fmt.Errorf("invalid page specification: %w", err)
		}

		selectedPages, err := api.ParsePageSelection(pageSpec)
		if err != nil {
			return fmt.Errorf("invalid page specification: %w", err)
		}

		outputFlag, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		dirFlag, err := cmd.Flags().GetString("dir")
		if err != nil {
			return err
		}

		output := outputFlag
		if output == "" {
			output = "remove.pdf"
		}

		if dirFlag == "" {
			dirFlag = filepath.Dir(input)
			if dirFlag == "" {
				dirFlag = "."
			}
		}

		if !filepath.IsAbs(output) {
			output = filepath.Join(dirFlag, output)
		}

		if strings.ToLower(filepath.Ext(output)) != ".pdf" {
			return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", output)
		}

		outDir := filepath.Dir(output)
		if outDir == "" {
			outDir = "."
		}

		if err := ensureOutputDirectory(cmd, outDir); err != nil {
			return err
		}

		output = resolveOutputPath(output)

		bar := progressbar.Default(-1, "Removing pages")
		err = api.RemovePagesFile(input, output, selectedPages, nil)
		if err != nil {
			return err
		}
		_ = bar.Finish()

		fmt.Fprintln(cmd.OutOrStdout(), "===== Page Removal Completed =====")
		report, err := GetFileInfo(output)
		if err != nil {
			return err
		}
		report.PrintReport(cmd.OutOrStdout())

		return nil
	},
}

var rmpageOutput string
var rmpageDir string
var rmpagePage string

// parsePageSpecification parses a page specification string and returns a sorted list of unique page numbers
func parsePageSpecification(spec string) ([]int, error) {
	// Check for empty input
	if strings.TrimSpace(spec) == "" {
		return nil, fmt.Errorf("page specification cannot be empty")
	}

	pageMap := make(map[int]bool)

	// Split by comma for multiple parts
	parts := strings.Split(spec, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for empty parts (from leading/trailing/consecutive commas)
		if part == "" {
			return nil, fmt.Errorf("invalid page specification: empty segment")
		}

		// Check if it's a range (contains a dash)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s (too many dashes)", part)
			}

			startStr := strings.TrimSpace(rangeParts[0])
			endStr := strings.TrimSpace(rangeParts[1])

			// Check for empty strings after trimming (e.g., from leading dash like "-5")
			if startStr == "" || endStr == "" {
				return nil, fmt.Errorf("invalid range format: %s (missing start or end)", part)
			}

			start, err := strconv.Atoi(startStr)
			if err != nil {
				return nil, fmt.Errorf("invalid start page in range: %s", startStr)
			}

			end, err := strconv.Atoi(endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid end page in range: %s", endStr)
			}

			// Validate page numbers are positive
			if start < 1 {
				return nil, fmt.Errorf("page number must be positive: %d", start)
			}
			if end < 1 {
				return nil, fmt.Errorf("page number must be positive: %d", end)
			}

			// Validate range order
			if start > end {
				return nil, fmt.Errorf("invalid range: start (%d) cannot be greater than end (%d)", start, end)
			}

			// Add all pages in range
			for i := start; i <= end; i++ {
				pageMap[i] = true
			}
		} else {
			// Single page number
			page, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", part)
			}

			if page < 1 {
				return nil, fmt.Errorf("page number must be positive: %d", page)
			}

			pageMap[page] = true
		}
	}

	// Check that we got at least one valid page
	if len(pageMap) == 0 {
		return nil, fmt.Errorf("no valid pages parsed")
	}

	// Convert map to sorted slice
	pages := make([]int, 0, len(pageMap))
	for page := range pageMap {
		pages = append(pages, page)
	}

	// Sort pages using simple bubble sort
	for i := 0; i < len(pages); i++ {
		for j := i + 1; j < len(pages); j++ {
			if pages[i] > pages[j] {
				pages[i], pages[j] = pages[j], pages[i]
			}
		}
	}

	return pages, nil
}

func init() {
	rootCmd.AddCommand(rmpageCmd)
	rmpageCmd.SetHelpTemplate(subHelpTemplate)
	rmpageCmd.Flags().StringVarP(&rmpagePage, "page", "p", "", "Page selector (example: 3, 1-4, 2,6-9)")
	rmpageCmd.Flags().StringVarP(&rmpageOutput, "output", "o", "", "Output PDF file path")
	rmpageCmd.Flags().StringVarP(&rmpageDir, "dir", "d", "", "Output directory (default: input PDF directory)")
}
