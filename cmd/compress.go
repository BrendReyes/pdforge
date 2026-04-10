/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf>",
	Short: "Compress a PDF to reduce file size",
	Long: `The compress command optimizes a PDF while preserving document usability.
Use it to make PDF files easier to store and share.`,
	Example: `  pdforge compress large-report.pdf
  pdforge compress archive.pdf`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pdforge compress called")
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
