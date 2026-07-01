package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/brendreyes/pdforge/internal/pdfops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert <image1> [image2 ...]",
	Short: "Convert image files into a single PDF",
	Long: `The convert command creates a PDF from one or more image files.
Use it to package scanned pages or image sets into one document.

Supported image file: JPG, PNG, WEBP, TIFF, TIF`,
	Example: `pdforge convert scan1.jpg scan2.jpg
pdforge convert page.png diagram.tiff -o converted.pdf`,
	RunE: runConvert,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.SetHelpTemplate(subHelpTemplate)
	convertCmd.Flags().StringP("output", "o", "", "Location with filename or filename only")
	convertCmd.Flags().StringP("dir", "d", "", "directory (default: input image directory)")
}

func runConvert(cmd *cobra.Command, args []string) error {
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

	if err := ensureOutputDirectory(cmd, dir); err != nil {
		return err
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	if output == "" {
		output = "converted_" + time.Now().Format("20060102_150405") + ".pdf"
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

	bar := progressbar.Default(-1, "Converting")
	report, err := pdfops.Convert(args, output)
	if err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Conversion Completed =====")
	report.PrintReport(cmd.OutOrStdout())

	return nil
}
