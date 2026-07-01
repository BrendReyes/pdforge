package pdfops

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// BuildSplitJobs builds the list of output segments for a split operation.
// selector is either a boundary page number (boundary mode) or a comma-separated
// page selection (extract mode). numPages is the total page count of the input PDF.
func BuildSplitJobs(selector string, numPages int, opts SplitOptions) ([]SplitJob, error) {
	if opts.Extract {
		return buildExtractJobs(selector, numPages, opts)
	}

	splitAt, err := strconv.Atoi(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid split page '%s': expected a page number", selector)
	}

	if splitAt < 1 || splitAt >= numPages {
		return nil, fmt.Errorf("split page out of bounds: %d (valid range: 1-%d, and split point must be before last page)", splitAt, numPages)
	}

	pad := pagePadWidth(numPages)
	leftPages := buildRange(1, splitAt)
	rightPages := buildRange(splitAt+1, numPages)

	leftLabel := fmt.Sprintf("%0*d-%0*d", pad, 1, pad, splitAt)
	rightLabel := fmt.Sprintf("%0*d-%0*d", pad, splitAt+1, pad, numPages)

	return []SplitJob{
		{Label: leftLabel, Pages: leftPages},
		{Label: rightLabel, Pages: rightPages},
	}, nil
}

// Split writes each job's pages from input into the corresponding outputPaths entry.
// outputPaths must be the same length as jobs. Pass an empty string for password
// if the file is not protected.
func Split(input string, outputPaths []string, jobs []SplitJob, password string) ([]FileInfo, error) {
	if strings.ToLower(filepath.Ext(input)) != ".pdf" {
		return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
	}

	conf := newConfig(password)

	if err := api.ValidateFile(input, conf); err != nil {
		errText := strings.ToLower(err.Error())
		if conf == nil && (strings.Contains(errText, "password") || strings.Contains(errText, "encrypt")) {
			return nil, fmt.Errorf("'%s' is password protected: provide the password with --password", filepath.Base(input))
		}
		return nil, fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
	}

	results := make([]FileInfo, 0, len(jobs))

	for i, job := range jobs {
		selectedPages := PagesToSelectionTokens(job.Pages)
		if err := api.TrimFile(input, outputPaths[i], selectedPages, conf); err != nil {
			return nil, fmt.Errorf("failed writing '%s': %w", outputPaths[i], err)
		}

		info, err := getFileInfo(outputPaths[i], conf)
		if err != nil {
			return nil, err
		}
		results = append(results, *info)
	}

	return results, nil
}

// PagesToSelectionTokens converts a sorted list of page numbers into the
// compact range-token strings that pdfcpu expects (e.g. [1,2,3,5] → ["1-3","5"]).
func PagesToSelectionTokens(pages []int) []string {
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

func buildExtractJobs(selector string, numPages int, opts SplitOptions) ([]SplitJob, error) {
	if strings.TrimSpace(selector) == "" {
		return nil, fmt.Errorf("page selection cannot be empty")
	}

	tokens := strings.Split(selector, ",")
	if len(tokens) == 0 {
		return nil, fmt.Errorf("page selection cannot be empty")
	}

	pad := pagePadWidth(numPages)
	jobs := make([]SplitJob, 0, len(tokens))

	for _, raw := range tokens {
		token := strings.TrimSpace(raw)
		if token == "" {
			return nil, fmt.Errorf("invalid page selection: empty segment")
		}

		pages, err := parseSegment(token, numPages)
		if err != nil {
			return nil, err
		}

		pages = filterParity(pages, opts.Odd, opts.Even)
		if len(pages) == 0 {
			return nil, fmt.Errorf("page selection '%s' has no pages after applying odd/even filter", token)
		}

		label := formatSegmentLabel(pages, pad)
		jobs = append(jobs, SplitJob{Label: label, Pages: pages})
	}

	return jobs, nil
}

func parseSegment(token string, numPages int) ([]int, error) {
	if !strings.Contains(token, "-") {
		page, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid page '%s'", token)
		}
		if page < 1 || page > numPages {
			return nil, fmt.Errorf("page out of bounds: %d (valid range: 1-%d)", page, numPages)
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
	end := numPages
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

	if start < 1 || start > numPages {
		return nil, fmt.Errorf("range start out of bounds: %d (valid range: 1-%d)", start, numPages)
	}
	if end < 1 || end > numPages {
		return nil, fmt.Errorf("range end out of bounds: %d (valid range: 1-%d)", end, numPages)
	}
	if start > end {
		return nil, fmt.Errorf("invalid range '%s': start (%d) cannot be greater than end (%d)", token, start, end)
	}

	return buildRange(start, end), nil
}

func buildRange(start, end int) []int {
	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}

func filterParity(pages []int, odd, even bool) []int {
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

func pagePadWidth(numPages int) int {
	width := len(strconv.Itoa(numPages))
	if width < 3 {
		return 3
	}
	return width
}

func formatRangeToken(start, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}
