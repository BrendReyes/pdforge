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
	pdforge split handbook.pdf -p 1-3,5 -o ./split_out`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "pdforge split called with input: %s, pages: %s, output dir: %s\n", args[0], splitPages, splitOutputDir)
		return nil
	},
}

var splitPages string
var splitOutputDir string

func init() {
	rootCmd.AddCommand(splitCmd)
	splitCmd.Flags().StringVarP(&splitPages, "pages", "p", "", "Page selection (example: 1-3,5,8-10)")
	splitCmd.Flags().StringVarP(&splitOutputDir, "output", "o", ".", "Output directory for split files")

	if err := splitCmd.MarkFlagRequired("pages"); err != nil {
		panic(err)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// splitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// splitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
