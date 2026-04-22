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

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert <image1> [image2 ...]",
	Short: "Convert image files into a single PDF",
	Long: "The convert command creates a PDF from one or more image files. Supported types: JPG, PNG, WEBP, TIFF",
	Example: "pdforge convert scan1.jpg scan2.jpg -o converted.pdf",
	RunE: runConvert,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringP("output", "o", "", "Location with filename")
	convertCmd.Flags().StringP("dir", "d", currentDir, "directory")

}

func runConvert(cmd *cobra.Command, args []string) error {

	dir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return err
	}

	// --dir validator
	if err := ensureOutputDirectory(cmd, dir) 
	err != nil {
		return err
	}

	// Invalid args checker
	bar := progressbar.Default(int64(len(args)), "Validating files")
	for _, item := range args {
		ftype := strings.ToLower(filepath.Ext(item))
		if ftype != ".png" && ftype != ".jpg" && ftype != ".webp" && ftype != ".tiff" {
			return fmt.Errorf("the file '%s' is invalid, must be supported image file (JPG, PNG, WEBP, TIFF)", filepath.Base(item))
		}
		
		_, err := os.Stat(item)
		if err != nil {
			return fmt.Errorf("invalid image '%s': \n%v", filepath.Base(item), err)
		}

		bar.Add(1)
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	// auto naming the file if no name is passed
	if output == "" {
		output = "convert_" + time.Now().Format("20060102_150405") + ".pdf"
	}

	fileType := strings.ToLower(filepath.Ext(output))
	if fileType != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", output)
	}

	output = filepath.Join(dir, output)
	output = resolveOutputPath(output) // Duplicate overwrite function

	bar = progressbar.Default(-1, "Converting")
	err = api.ImportImagesFile(args, output, nil, nil)
	if err != nil {
		return err
	}
	bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Conversion Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		return err
	}
	report.PrintReport(cmd.OutOrStdout())

	return nil
}
