package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/brendreyes/pdforge/internal/pdfops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:     "merge <file1.pdf> <file2.pdf> [more.pdf...]",
	Short:   "Merge two or more PDF files into one document",
	Long:    "The merge command combines multiple PDF files into a single output PDF. The input order is preserved in the merged document.",
	Example: "pdforge merge invoice-jan.pdf invoice-feb.pdf\npdforge merge file1.pdf file2.pdf -o result.pdf",
	RunE:    runMerge,
	Args:    cobra.MinimumNArgs(2),
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.SetHelpTemplate(subHelpTemplate)
	mergeCmd.Flags().StringP("output", "o", "", "Location with filename or filename only")
	mergeCmd.Flags().StringP("dir", "d", "", "directory (default: first input PDF directory)")
	mergeCmd.Flags().StringP("password", "P", "", "Password for protected PDFs (output stays protected)")
}

func runMerge(cmd *cobra.Command, args []string) error {
	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return err
	}

	if dir == "" {
		dir = filepath.Dir(args[0])
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
		output = "merged_" + time.Now().Format("20060102_150405") + ".pdf"
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

	bar := progressbar.Default(-1, "Merging")
	report, err := pdfops.Merge(args, output, password)
	if err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Merged Completed =====")
	report.PrintReport(cmd.OutOrStdout())

	return nil
}
