/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/spf13/cobra"
)

// splitCmd represents the split command
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

type splitJob struct {
	Label string
	Pages []int
}

var splitExtract bool
var splitOdd bool
var splitEven bool
var splitVerbose bool
var splitOutput string
var splitDir string
var splitPage string

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

	if strings.ToLower(filepath.Ext(input)) != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
	}

	if err := api.ValidateFile(input, nil); err != nil {
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "password") || strings.Contains(errText, "encrypt") {
			return fmt.Errorf("encrypted PDFs are not supported yet")
		}
		return fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
	}

	pageCount, err := api.PageCountFile(input)
	if err != nil {
		return fmt.Errorf("failed to read page count: %w", err)
	}

	if pageCount <= 1 {
		return fmt.Errorf("cannot split a single-page PDF")
	}

	if splitOdd && splitEven {
		return fmt.Errorf("--odd and --even cannot be used together")
	}
	if !splitExtract && (splitOdd || splitEven) {
		return fmt.Errorf("--odd/--even can only be used with --extract")
	}

	jobs, err := buildSplitJobs(selector, pageCount)
	if err != nil {
		return err
	}

	outputPaths, err := resolveSplitOutputs(cmd, input, jobs)
	if err != nil {
		return err
	}

	for i, job := range jobs {
		selectedPages := pagesToSelectionTokens(job.Pages)
		if trimErr := api.TrimFile(input, outputPaths[i], selectedPages, nil); trimErr != nil {
			return fmt.Errorf("failed writing '%s': %w", outputPaths[i], trimErr)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "===== Split Completed =====")
	for i, path := range outputPaths {
		fmt.Fprintf(cmd.OutOrStdout(), "-- File %d --\n", i+1)
		report, err := GetFileInfo(path)
		if err != nil {
			return err
		}
		report.PrintReport(cmd.OutOrStdout())
		if splitVerbose {
			fmt.Fprintf(cmd.OutOrStdout(), "Pages: %s\n", pagesDisplay(jobs[i].Pages))
		}
		fmt.Fprintln(cmd.OutOrStdout(), "")
	}


	return nil
}

func buildSplitJobs(selector string, pageCount int) ([]splitJob, error) {
	if splitExtract {
		return buildExtractJobs(selector, pageCount)
	}

	splitAt, err := strconv.Atoi(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid split page '%s': expected a page number", selector)
	}

	if splitAt < 1 || splitAt >= pageCount {
		return nil, fmt.Errorf("split page out of bounds: %d (valid range: 1-%d, and split point must be before last page)", splitAt, pageCount)
	}

	pad := pagePadWidth(pageCount)
	leftPages := buildRange(1, splitAt)
	rightPages := buildRange(splitAt+1, pageCount)

	leftLabel := fmt.Sprintf("%0*d-%0*d", pad, 1, pad, splitAt)
	rightLabel := fmt.Sprintf("%0*d-%0*d", pad, splitAt+1, pad, pageCount)

	return []splitJob{
		{Label: leftLabel, Pages: leftPages},
		{Label: rightLabel, Pages: rightPages},
	}, nil
}

func buildExtractJobs(selector string, pageCount int) ([]splitJob, error) {
	if strings.TrimSpace(selector) == "" {
		return nil, fmt.Errorf("page selection cannot be empty")
	}

	tokens := strings.Split(selector, ",")
	if len(tokens) == 0 {
		return nil, fmt.Errorf("page selection cannot be empty")
	}

	pad := pagePadWidth(pageCount)
	jobs := make([]splitJob, 0, len(tokens))

	for _, raw := range tokens {
		token := strings.TrimSpace(raw)
		if token == "" {
			return nil, fmt.Errorf("invalid page selection: empty segment")
		}

		pages, err := parseSegment(token, pageCount)
		if err != nil {
			return nil, err
		}

		pages = filterParity(pages, splitOdd, splitEven)
		if len(pages) == 0 {
			return nil, fmt.Errorf("page selection '%s' has no pages after applying odd/even filter", token)
		}

		label := formatSegmentLabel(pages, pad)
		jobs = append(jobs, splitJob{Label: label, Pages: pages})
	}

	return jobs, nil
}

func parseSegment(token string, pageCount int) ([]int, error) {
	if !strings.Contains(token, "-") {
		page, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid page '%s'", token)
		}
		if page < 1 || page > pageCount {
			return nil, fmt.Errorf("page out of bounds: %d (valid range: 1-%d)", page, pageCount)
		}
		return []int{page}, nil
	}

	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range '%s'", token)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	if left == "" && right == "" {
		return nil, fmt.Errorf("invalid range '%s'", token)
	}

	start := 1
	end := pageCount
	var err error

	if left != "" {
		start, err = strconv.Atoi(left)
		if err != nil {
			return nil, fmt.Errorf("invalid range start '%s'", left)
		}
	}

	if right != "" {
		end, err = strconv.Atoi(right)
		if err != nil {
			return nil, fmt.Errorf("invalid range end '%s'", right)
		}
	}

	if start < 1 || start > pageCount {
		return nil, fmt.Errorf("range start out of bounds: %d (valid range: 1-%d)", start, pageCount)
	}
	if end < 1 || end > pageCount {
		return nil, fmt.Errorf("range end out of bounds: %d (valid range: 1-%d)", end, pageCount)
	}
	if start > end {
		return nil, fmt.Errorf("invalid range '%s': start (%d) cannot be greater than end (%d)", token, start, end)
	}

	return buildRange(start, end), nil
}

func resolveSplitOutputs(cmd *cobra.Command, inputPath string, jobs []splitJob) ([]string, error) {
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

func pagePadWidth(pageCount int) int {
	width := len(strconv.Itoa(pageCount))
	if width < 3 {
		return 3
	}
	return width
}

func buildRange(start, end int) []int {
	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}

func filterParity(pages []int, odd bool, even bool) []int {
	if !odd && !even {
		return pages
	}

	filtered := make([]int, 0, len(pages))
	for _, p := range pages {
		if odd && p%2 == 1 {
			filtered = append(filtered, p)
		}
		if even && p%2 == 0 {
			filtered = append(filtered, p)
		}
	}

	return filtered
}

func formatSegmentLabel(pages []int, pad int) string {
	if len(pages) == 1 {
		return fmt.Sprintf("%0*d", pad, pages[0])
	}

	return fmt.Sprintf("%0*d-%0*d", pad, pages[0], pad, pages[len(pages)-1])
}

func pagesDisplay(pages []int) string {
	parts := pagesToSelectionTokens(pages)
	return strings.Join(parts, ",")
}

func pagesToSelectionTokens(pages []int) []string {
	if len(pages) == 0 {
		return nil
	}

	tokens := make([]string, 0)
	start := pages[0]
	prev := pages[0]

	for i := 1; i < len(pages); i++ {
		if pages[i] == prev+1 {
			prev = pages[i]
			continue
		}

		tokens = append(tokens, formatRangeToken(start, prev))
		start = pages[i]
		prev = pages[i]
	}

	tokens = append(tokens, formatRangeToken(start, prev))
	return tokens
}

func formatRangeToken(start, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d-%d", start, end)
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
}