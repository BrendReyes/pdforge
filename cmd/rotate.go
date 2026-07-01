package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/brendreyes/pdforge/internal/pdfops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <input.pdf> <degrees>",
	Short: "Rotate pages of a PDF clockwise",
	Long: `Rotate pages of a PDF by a multiple of 90 degrees.

Degrees are applied clockwise; use a negative value to rotate counter-clockwise.
By default every page is rotated. Use --page to rotate only selected pages.

  - All pages:      pdforge rotate input.pdf 90
  - Counter-clockwise: pdforge rotate input.pdf -90
  - Selected pages: pdforge rotate input.pdf 180 --page 1,3-5`,
	Example: `  pdforge rotate input.pdf 90
  pdforge rotate input.pdf -90 -o rotated.pdf
  pdforge rotate input.pdf 180 --page 2,6-9 -d ./out`,
	Args: cobra.ExactArgs(2),
	RunE: runRotate,
}

func init() {
	rootCmd.AddCommand(rotateCmd)
	rotateCmd.SetHelpTemplate(subHelpTemplate)
	rotateCmd.Flags().StringP("page", "p", "", "Page selector (example: 3, 1-4, 2,6-9). Default: all pages")
	rotateCmd.Flags().StringP("output", "o", "", "Location with filename or filename only")
	rotateCmd.Flags().StringP("dir", "d", "", "Output directory (default: input PDF directory)")
	rotateCmd.Flags().StringP("password", "P", "", "Password for protected PDFs (output stays protected)")
}

func runRotate(cmd *cobra.Command, args []string) error {
	input := args[0]

	rotation, err := strconv.Atoi(strings.TrimSpace(args[1]))
	if err != nil {
		return fmt.Errorf("invalid rotation '%s': expected a multiple of 90 (e.g. 90, 180, -90)", args[1])
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return err
	}

	pageFlag, err := cmd.Flags().GetString("page")
	if err != nil {
		return err
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
		output = "rotated_" + time.Now().Format("20060102_150405") + ".pdf"
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

	outDir := filepath.Dir(output)
	if outDir == "" {
		outDir = "."
	}

	if err := ensureOutputDirectory(cmd, outDir); err != nil {
		return err
	}

	output = resolveOutputPath(output)

	bar := progressbar.Default(-1, "Rotating")
	report, err := pdfops.Rotate(input, output, rotation, pageFlag, password)
	if err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Rotation Completed =====")
	report.PrintReport(cmd.OutOrStdout())

	return nil
}
