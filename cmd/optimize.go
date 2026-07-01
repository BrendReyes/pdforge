package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/brendreyes/pdforge/internal/pdfops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize <file.pdf>",
	Short: "Optimize a PDF to reduce file size",
	Long: `The optimize command optimizes a PDF while preserving document usability.
Use it to make PDF files easier to store and share.

Only accepts 1 file at a time.
Note: optimization focuses on PDF structure and text streams.
For best results on image-heavy PDFs, consider reducing image
resolution before optimizing.`,
	Example: `pdforge optimize large-report.pdf
pdforge optimize archive.pdf -o archive_optimized.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runOptimize,
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
	optimizeCmd.SetHelpTemplate(subHelpTemplate)
	optimizeCmd.Flags().StringP("output", "o", "", "Location with filename or filename only")
	optimizeCmd.Flags().StringP("dir", "d", "", "directory (default: input PDF directory)")
	optimizeCmd.Flags().StringP("password", "P", "", "Password for protected PDFs (output stays protected)")
}

func runOptimize(cmd *cobra.Command, args []string) error {
	inFile := args[0]

	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return err
	}

	if dir == "" {
		dir = filepath.Dir(inFile)
		if dir == "" {
			dir = "."
		}
	}

	if err := ensureOutputDirectory(cmd, dir); err != nil {
		return err
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	if output == "" {
		output = "optimized_" + time.Now().Format("20060102_150405") + ".pdf"
	}

	if filepath.IsAbs(output) || filepath.Dir(output) != "." {
		outputDir := filepath.Dir(output)
		if err := ensureOutputDirectory(cmd, outputDir); err != nil {
			return err
		}
	} else {
		output = filepath.Join(dir, output)
	}
	output = resolveOutputPath(output)

	bar := progressbar.Default(-1, "Optimizing")
	report, err := pdfops.Optimize(inFile, output, password)
	if err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Optimization Completed =====")
	report.PrintReport(cmd.OutOrStdout())

	return nil
}
