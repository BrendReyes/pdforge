package cmd

import (
	"fmt"
	//"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
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
}

func runMerge(cmd *cobra.Command, args []string) error {

	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	if dir == "" {
		dir = filepath.Dir(args[0])
		if dir == "" {
			dir = "."
		}
	}

	// --dir validator
	if err := ensureOutputDirectory(cmd, dir); err != nil {
		return err
	}

	// Invalid args checker
	bar := progressbar.Default(int64(len(args)), "Validating files")
	for _, item := range args {
		ftype := strings.ToLower(filepath.Ext(item))
		if ftype != ".pdf" {
			return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(item))
		}

		err := api.ValidateFile(item, nil)
		if err != nil {
			return fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(item), err)
		}

		bar.Add(1)
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	// auto naming the file if no name is passed
	if output == "" {
		output = "merged_" + time.Now().Format("20060102_150405") + ".pdf"
	}

	fileType := strings.ToLower(filepath.Ext(output))
	if fileType != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", output)
	}

	// Route output path correctly
	if filepath.IsAbs(output) || filepath.Dir(output) != "." {
		outputDir := filepath.Dir(output)
		if err := ensureOutputDirectory(cmd, outputDir); err != nil {
			return err
		}
	} else {
		output = filepath.Join(dir, output)
	}

	output = resolveOutputPath(output) // Duplicate overwrite function

	bar = progressbar.Default(-1, "Merging")
	err = api.MergeCreateFile(args, output, false, nil)
	if err != nil {
		return err
	}
	bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Merged Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		return err
	}
	report.PrintReport(cmd.OutOrStdout())

	return nil
}