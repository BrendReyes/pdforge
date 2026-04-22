/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:     "merge <file1.pdf> <file2.pdf> *use quotation for filename or directory with spaces or special char* [more.pdf...]",
	Short:   "Merge two or more PDF files into one document",
	Long:    "The merge command combines multiple PDF files into a single output PDF. The input order is preserved in the merged document.",
	Example: "pdforge merge invoice-jan.pdf invoice-feb.pdf\npdforge merge file1.pdf file2.pdf -o result.pdf",
	RunE:    runMerge,
	Args:    argsWithHelp(cobra.MinimumNArgs(2)),
}

func init() {

	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringP("output", "o", "", "Location with filename")
	mergeCmd.Flags().StringP("dir", "d", currentDir, "directory")
}

func runMerge(cmd *cobra.Command, args []string) error {

	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
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

	output = filepath.Join(dir, output)
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
