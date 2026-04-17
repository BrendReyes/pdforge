/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf>",
	Short: "Compress a PDF to reduce file size",
	Long: `The compress command optimizes a PDF while preserving document usability.
Use it to make PDF files easier to store and share.`,
	Example: `  pdforge compress large-report.pdf
	pdforge compress archive.pdf -o archive_compressed.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runCompress,
}

var compressOutput string
var compressImageMode string
var compressImageMaxDimension int
var compressImageJPEGQuality int

const (
	compressImageModeOff          = "off"
	compressImageModeReadable     = "readable"
	compressImageModeBalanced     = "balanced"
	compressImageModeAggressive   = "aggressive"
	compressImageModeExperimental = "experimental"
)

func defaultCompressedOutput(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".pdf"
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	return filepath.Join(filepath.Dir(inputPath), base+"_compressed"+ext)
}

func runCompress(cmd *cobra.Command, args []string) error {
	input := args[0]

	if strings.ToLower(filepath.Ext(input)) != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
	}

	bar := progressbar.Default(1, "Validating file")
	if err := api.ValidateFile(input, nil); err != nil {
		return fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
	}
	_ = bar.Add(1)

	output := compressOutput
	if output == "" {
		output = defaultCompressedOutput(input)
	}

	if strings.ToLower(filepath.Ext(output)) != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", output)
	}

	outDir := filepath.Dir(output)
	if outDir == "" {
		outDir = "."
	}

	dirInfo, err := os.Stat(outDir)
	if err != nil {
		return fmt.Errorf("directory '%s' does not exist", outDir)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", outDir)
	}

	output = resolveOutputPath(output)

	bar = progressbar.Default(-1, "Compressing")
	if err := api.OptimizeFile(input, output, nil); err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Println("===== Compression Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		return err
	}
	report.PrintReport()

	return nil
}

func init() {
	rootCmd.AddCommand(compressCmd)
	compressCmd.Flags().StringVarP(&compressOutput, "output", "o", "", "Output PDF file path (default: <input>_compressed.pdf)")
	compressCmd.Flags().StringVar(&compressImageMode, "image-mode", compressImageModeOff, "Image-aware compression mode: off|readable|balanced|aggressive")
	compressCmd.Flags().IntVar(&compressImageMaxDimension, "image-max-dimension", 0, "Advanced: override max image dimension in pixels for image mode")
	compressCmd.Flags().IntVar(&compressImageJPEGQuality, "image-jpeg-quality", 0, "Advanced: override JPEG quality for image mode (1-100)")

	if err := compressCmd.Flags().MarkHidden("image-max-dimension"); err != nil {
		panic(err)
	}
	if err := compressCmd.Flags().MarkHidden("image-jpeg-quality"); err != nil {
		panic(err)
	}

	compressCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		compressImageMode = strings.ToLower(strings.TrimSpace(compressImageMode))
		if compressImageMode == compressImageModeExperimental {
			compressImageMode = compressImageModeAggressive
		}

		if compressImageMode != compressImageModeOff && compressImageMode != compressImageModeReadable && compressImageMode != compressImageModeBalanced && compressImageMode != compressImageModeAggressive {
			return fmt.Errorf("invalid --image-mode '%s', expected: off|readable|balanced|aggressive", compressImageMode)
		}

		if compressImageMode != compressImageModeOff {
			fmt.Fprintln(cmd.OutOrStdout(), "Image mode flags are currently inactive; proceeding with standard PDF optimization only.")
		}

		return nil
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
