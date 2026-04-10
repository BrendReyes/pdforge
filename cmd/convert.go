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
  pdforge convert page.png diagram.tiff`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pdforge convert called")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// convertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// convertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
