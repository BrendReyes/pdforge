/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge <file1.pdf> <file2.pdf> [more.pdf...]",
	Short: "Merge two or more PDF files into one document",
	Long: `The merge command combines multiple PDF files into a single output PDF.
The input order is preserved in the merged document.`,
	Example: `  pdforge merge invoice-jan.pdf invoice-feb.pdf
  pdforge merge part1.pdf part2.pdf appendix.pdf`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pdforge merge called")
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mergeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mergeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
