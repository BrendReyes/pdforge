/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"time"
	"log"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/spf13/cobra"
)
/**
TODO:
[] Remove the overwriting a same file name
[] Implement ability to choose a directory to store the file
[x] Auto naming
[x] Summary output
[] Check if file already exist
**/

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge <file1.pdf> <file2.pdf> [more.pdf...]",
	Short: "Merge two or more PDF files into one document",
	Long: "The merge command combines multiple PDF files into a single output PDF. The input order is preserved in the merged document.",
	Example: "pdforge merge invoice-jan.pdf invoice-feb.pdf\npdforge merge file1.pdf file2.pdf -o result.pdf",
	RunE: runMerge,
	Args: cobra.MinimumNArgs(2),
}

func init() {
	currentTime := time.Now()
	
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringP("output", "o", "merged_" + currentTime.Format("20060102_150405") + ".pdf", "Output file name") // (flag name, shortcut, default naming, name set)
	//mergeCmd.Flags().StringP("append", "a", "append_merged.pdf", "Output file name")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mergeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mergeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runMerge(cmd *cobra.Command, args []string) error {

	output, err := cmd.Flags().GetString("output")
    if err != nil {
        return err
    }


	err = api.MergeCreateFile(args, output, false, nil)
	if err != nil {
		return err
	}

	fmt.Println("===== Merged Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		log.Fatal(err)
	}
	report.PrintReport() 

	return nil
}