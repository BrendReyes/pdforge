/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		output := compressOutput
		if output == "" {
			output = defaultCompressedOutput(args[0])
		}

		fmt.Fprintf(cmd.OutOrStdout(), "pdforge compress called with input: %s, output: %s\n", args[0], output)
		return nil
	},
}

var compressOutput string

func defaultCompressedOutput(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".pdf"
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	return filepath.Join(filepath.Dir(inputPath), base+"_compressed"+ext)
}

func init() {
	rootCmd.AddCommand(compressCmd)
	compressCmd.Flags().StringVarP(&compressOutput, "output", "o", "", "Output PDF file path (default: <input>_compressed.pdf)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
