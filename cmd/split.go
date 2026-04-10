/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// splitCmd represents the split command
var splitCmd = &cobra.Command{
	Use:   "split <file.pdf>",
	Short: "Split a PDF into selected pages or ranges",
	Long: `The split command extracts pages from a source PDF based on page selections.
Use it to isolate specific pages or ranges from larger documents.`,
	Example: `  pdforge split report.pdf
  pdforge split handbook.pdf`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pdforge split called")
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// splitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// splitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
