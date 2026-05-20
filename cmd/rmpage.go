package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// rmpageCmd represents the rmpage command
var rmpageCmd = &cobra.Command{
	Use:   "rmpage <input_path> [page_selector]",
	Short: "Remove one or more pages from a PDF",
	Long: `Remove pages from a PDF file. Accepts:
	  - Single page: pdforge rmpage input.pdf 8
	  - Flag form: pdforge rmpage input.pdf --page 8
	  - Range: pdforge rmpage input.pdf 1-3
	  - Combination: pdforge rmpage input.pdf 1,6-11,17`,
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
			output = "removed_" + time.Now().Format("20060102_150405") + ".pdf"
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

func init() {
	rootCmd.AddCommand(rmpageCmd)
	rmpageCmd.SetHelpTemplate(subHelpTemplate)
	rmpageCmd.Flags().StringP("page", "p", "", "Page selector (example: 3, 1-4, 2,6-9)")
	rmpageCmd.Flags().StringP("output", "o", "", "Location with filename or filename only")
	rmpageCmd.Flags().StringP("dir", "d", "", "Output directory (default: input PDF directory)")
}