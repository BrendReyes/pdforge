package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/brendreyes/pdforge/internal/pdfops"
	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split [flags] <input.pdf> [selector]",
	Short: "Split a PDF by boundary or extract selected ranges",
	Long: `Default mode treats the selector as the last page of the first split.
For example, 'pdforge split input.pdf 8' creates pages 1-8 in the first file and 9-end in the second.

Use --page to pass the same boundary explicitly, or enable extract mode (-e/--extract) to write one output file per selected segment.

If --output is omitted, split uses the default naming pattern for the selected mode.`,
	Example: `  pdforge split input.pdf 8
  pdforge split input.pdf --page 8
  pdforge split -e input.pdf 6,8-10,11
  pdforge split input.pdf 8 -o section -d ./out`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSplit,
}

var splitExtract bool
var splitOdd bool
var splitEven bool
var splitVerbose bool
var splitOutput string
var splitDir string
var splitPage string
var splitPassword string

func runSplit(cmd *cobra.Command, args []string) error {
	if len(args) > 1 && looksLikeSelector(args[0]) && strings.ToLower(filepath.Ext(args[1])) == ".pdf" {
		return fmt.Errorf("invalid argument order: expected '<input.pdf> <selector>', got '<selector> <input.pdf>'")
	}

	input := args[0]

	pageFlag := strings.TrimSpace(splitPage)
	positionalSelector := ""
	if len(args) == 2 {
		positionalSelector = strings.TrimSpace(args[1])
	}

	if positionalSelector != "" && pageFlag != "" {
		return fmt.Errorf("provide selector either as positional argument or via --page, not both")
	}

	selector := pageFlag
	if selector == "" {
		selector = positionalSelector
	}
	if selector == "" {
		return fmt.Errorf("missing selector: provide [selector] or --page")
	}

	if splitOdd && splitEven {
		return fmt.Errorf("--odd and --even cannot be used together")
	}
	if !splitExtract && (splitOdd || splitEven) {
		return fmt.Errorf("--odd/--even can only be used with --extract")
	}

	numPages, err := pdfops.PageCount(input, splitPassword)
	if err != nil {
		return fmt.Errorf("failed to read page count: %w", err)
	}

	if numPages <= 1 {
		return fmt.Errorf("cannot split a single-page PDF")
	}

	jobs, err := pdfops.BuildSplitJobs(selector, numPages, pdfops.SplitOptions{
		Extract: splitExtract,
		Odd:     splitOdd,
		Even:    splitEven,
	})
	if err != nil {
		return err
	}

	outputPaths, err := resolveSplitOutputs(cmd, input, jobs)
	if err != nil {
		return err
	}

	results, err := pdfops.Split(input, outputPaths, jobs, splitPassword)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "===== Split Completed =====")
	for i, result := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "-- File %d --\n", i+1)
		result.PrintReport(cmd.OutOrStdout())
		if splitVerbose {
			fmt.Fprintf(cmd.OutOrStdout(), "Pages: %s\n", pagesDisplay(jobs[i].Pages))
		}
		fmt.Fprintln(cmd.OutOrStdout(), "")
	}

	return nil
}

func resolveSplitOutputs(cmd *cobra.Command, inputPath string, jobs []pdfops.SplitJob) ([]string, error) {
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no split jobs to process")
	}

	baseDir := splitDir
	if baseDir == "" {
		baseDir = filepath.Dir(inputPath)
		if baseDir == "" {
			baseDir = "."
		}
	}

	if err := ensureOutputDirectory(cmd, baseDir); err != nil {
		return nil, err
	}

	if len(jobs) == 1 && splitOutput != "" {
		out := splitOutput
		if strings.ToLower(filepath.Ext(out)) != ".pdf" {
			return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", out)
		}

		if !filepath.IsAbs(out) {
			out = filepath.Join(baseDir, out)
		}

		dir := filepath.Dir(out)
		if dir == "" {
			dir = "."
		}
		if err := ensureOutputDirectory(cmd, dir); err != nil {
			return nil, err
		}

		return []string{resolveOutputPath(out)}, nil
	}

	prefix := "split"
	if splitOutput != "" {
		prefix = strings.TrimSpace(splitOutput)
		if prefix == "" {
			return nil, fmt.Errorf("output prefix cannot be empty")
		}

		if strings.ToLower(filepath.Ext(prefix)) == ".pdf" {
			prefix = strings.TrimSuffix(prefix, filepath.Ext(prefix))
		}

		prefix = filepath.Base(prefix)
	}

	paths := make([]string, 0, len(jobs))
	for _, job := range jobs {
		name := fmt.Sprintf("%s_%s.pdf", prefix, job.Label)
		candidate := filepath.Join(baseDir, name)
		paths = append(paths, resolveOutputPath(candidate))
	}

	return paths, nil
}

func pagesDisplay(pages []int) string {
	parts := pdfops.PagesToSelectionTokens(pages)
	return strings.Join(parts, ",")
}

func looksLikeSelector(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r >= '0' && r <= '9') || r == ',' || r == '-' || r == ' ' {
			continue
		}
		return false
	}

	return true
}

func init() {
	rootCmd.AddCommand(splitCmd)
	splitCmd.SetHelpTemplate(subHelpTemplate)
	splitCmd.Flags().BoolVarP(&splitExtract, "extract", "e", false, "Extract selected page segments into separate files")
	splitCmd.Flags().BoolVar(&splitOdd, "odd", false, "Keep only odd pages in selected segments (extract mode)")
	splitCmd.Flags().BoolVarP(&splitEven, "even", "n", false, "Keep only even pages in selected segments (extract mode)")
	splitCmd.Flags().BoolVarP(&splitVerbose, "verbose", "v", false, "Print per-output page details")
	splitCmd.Flags().StringVarP(&splitPage, "page", "p", "", "Page selector (boundary or extract selector)")
	splitCmd.Flags().StringVarP(&splitOutput, "output", "o", "", "Output file name (single result) or prefix (multiple results)")
	splitCmd.Flags().StringVarP(&splitDir, "dir", "d", "", "Output directory (default: input PDF directory)")
	splitCmd.Flags().StringVarP(&splitPassword, "password", "P", "", "Password for protected PDFs (output stays protected)")
}
