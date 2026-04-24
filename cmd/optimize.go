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

// optimizeCmd represents the optimize command
var optimizeCmd = &cobra.Command{
	Use:   "optimize <file.pdf>",
	Short: "Optimize a PDF to reduce file size",
	Long: `The optimize command optimizes a PDF while preserving document usability.
Use it to make PDF files easier to store and share.

Only accepts 1 file at a time.
Note: optimization focuses on PDF structure and text streams.
For best results on image-heavy PDFs, consider reducing image
resolution before optimizing.`,
	Example: `  pdforge optimize large-report.pdf
	pdforge optimize archive.pdf -o archive_optimized.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runOptimize,
}

func runOptimize(cmd *cobra.Command, args []string) error {

	inFile := args[0]
	
	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	// --dir validator
	err = ensureOutputDirectory(cmd, dir)
	if err != nil {
		return err
	}

	// Invalid args checker
	ftype := strings.ToLower(filepath.Ext(inFile))
	if ftype != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(inFile))
	}

	err = api.ValidateFile(inFile, nil)
	if err != nil {
		return fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(inFile), err)
	}


	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	// auto naming the file if no name is passed
	if output == "" {
		output = "optimized_" + time.Now().Format("20060102_150405") + ".pdf"
	}

	fileType := strings.ToLower(filepath.Ext(output))
	if fileType != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", output)
	}

	output = filepath.Join(dir, output)
	output = resolveOutputPath(output) // Duplicate overwrite function

	bar := progressbar.Default(-1, "Optimizing")
	err = api.OptimizeFile(inFile, output, nil)
	if err != nil {
		return err
	}
	bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Optimization Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		return err
	}
	report.PrintReport(cmd.OutOrStdout())

	return nil
}

func init() {
	
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	rootCmd.AddCommand(optimizeCmd)
	optimizeCmd.SetHelpTemplate(subHelpTemplate)
	optimizeCmd.Flags().StringP("output", "o", "", "Location with filename")
	optimizeCmd.Flags().StringP("dir", "d", currentDir, "directory")
}