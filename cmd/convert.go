/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert <image1> [image2 ...]",
	Short: "Convert image files into a single PDF",
	Long: `The convert command creates a PDF from one or more image files.
Use it to package scanned pages or image sets into one document.`,
	Example: `  pdforge convert scan1.jpg scan2.jpg
	pdforge convert page.png diagram.tiff -o converted.pdf`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "pdforge convert called with %d images, output: %s\n", len(args), convertOutput)
		return nil
	},
}

var convertOutput string

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "converted.pdf", "Output PDF file path")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// convertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// convertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
