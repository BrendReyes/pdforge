package pdfops

import (
	"os"
	"path/filepath"
	"testing"
)

// testdata paths relative to this package directory
const (
	tdPDFs   = "../../cmd/testdata/pdfs"
	tdImages = "../../cmd/testdata/images"
)

func pdf(name string) string   { return filepath.Join(tdPDFs, name) }
func img(name string) string   { return filepath.Join(tdImages, name) }
func tmp(t *testing.T) string  {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "out.pdf")
}

// ── Optimize ─────────────────────────────────────────────────────────────────

func TestOptimize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		password  string
		expectErr bool
	}{
		{"valid pdf", pdf("sample1.pdf"), "", false},
		{"uppercase extension", pdf("sample5.PDF"), "", false},
		{"protected pdf correct password", pdf("sample_protected.pdf"), "samplefiles", false},
		{"protected pdf no password", pdf("sample_protected.pdf"), "", true},
		{"protected pdf wrong password", pdf("sample_protected.pdf"), "wrongpw", true},
		{"non-pdf input", pdf("sample.docx"), "", true},
		{"file does not exist", pdf("ghost.pdf"), "", true},
		{"image as input", img("sample.jpg"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tmp(t)
			_, err := Optimize(tt.input, out, tt.password)
			assertErr(t, err, tt.expectErr)
		})
	}
}

func TestOptimizeOutputExtension(t *testing.T) {
	out := filepath.Join(t.TempDir(), "result.docx")
	_, err := Optimize(pdf("sample1.pdf"), out, "")
	if err == nil {
		t.Error("expected error for non-.pdf output, got nil")
	}
}

// ── Merge ─────────────────────────────────────────────────────────────────────

func TestMerge(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		password  string
		expectErr bool
	}{
		{"two valid pdfs", []string{pdf("sample1.pdf"), pdf("sample2.pdf")}, "", false},
		{"three valid pdfs", []string{pdf("sample1.pdf"), pdf("sample2.pdf"), pdf("sample3.pdf")}, "", false},
		{"non-pdf in inputs", []string{pdf("sample1.pdf"), img("sample.jpg")}, "", true},
		{"file does not exist", []string{pdf("sample1.pdf"), pdf("ghost.pdf")}, "", true},
		{"invalid pdf content", []string{pdf("sample1.pdf"), pdf("sample.docx")}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tmp(t)
			_, err := Merge(tt.inputs, out, tt.password)
			assertErr(t, err, tt.expectErr)
		})
	}
}

func TestMergeOutputExtension(t *testing.T) {
	out := filepath.Join(t.TempDir(), "result.txt")
	_, err := Merge([]string{pdf("sample1.pdf"), pdf("sample2.pdf")}, out, "")
	if err == nil {
		t.Error("expected error for non-.pdf output, got nil")
	}
}

// ── Convert ───────────────────────────────────────────────────────────────────

func TestConvert(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		expectErr bool
	}{
		{"single jpg", []string{img("sample.jpg")}, false},
		{"single png", []string{img("sample.png")}, false},
		{"single webp", []string{img("sample.webp")}, false},
		{"single tiff", []string{img("sample.tiff")}, false},
		{"single tif", []string{img("sample.tif")}, false},
		{"multiple images", []string{img("sample.jpg"), img("sample.png"), img("sample.webp")}, false},
		{"unsupported format bmp", []string{img("sample.bmp")}, true},
		{"unsupported format gif", []string{img("sample.gif")}, true},
		{"file does not exist", []string{img("ghost.jpg")}, true},
		{"pdf as input", []string{pdf("sample1.pdf")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tmp(t)
			_, err := Convert(tt.inputs, out)
			assertErr(t, err, tt.expectErr)
		})
	}
}

func TestConvertOutputExtension(t *testing.T) {
	out := filepath.Join(t.TempDir(), "result.jpg")
	_, err := Convert([]string{img("sample.jpg")}, out)
	if err == nil {
		t.Error("expected error for non-.pdf output, got nil")
	}
}

// ── Rotate ────────────────────────────────────────────────────────────────────

func TestRotate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		degrees   int
		pageSpec  string
		password  string
		expectErr bool
	}{
		{"90 degrees all pages", pdf("sample1.pdf"), 90, "", "", false},
		{"180 degrees all pages", pdf("sample1.pdf"), 180, "", "", false},
		{"-90 degrees all pages", pdf("sample1.pdf"), -90, "", "", false},
		{"270 degrees all pages", pdf("sample1.pdf"), 270, "", "", false},
		{"select page", pdf("sample2.pdf"), 90, "1", "", false},
		{"invalid rotation zero", pdf("sample1.pdf"), 0, "", "", true},
		{"invalid rotation not multiple of 90", pdf("sample1.pdf"), 45, "", "", true},
		{"non-pdf input", img("sample.jpg"), 90, "", "", true},
		{"file does not exist", pdf("ghost.pdf"), 90, "", "", true},
		{"protected pdf correct password", pdf("sample_protected.pdf"), 90, "", "samplefiles", false},
		{"protected pdf no password", pdf("sample_protected.pdf"), 90, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tmp(t)
			_, err := Rotate(tt.input, out, tt.degrees, tt.pageSpec, tt.password)
			assertErr(t, err, tt.expectErr)
		})
	}
}

// ── RemovePages ───────────────────────────────────────────────────────────────

func TestRemovePages(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pageSpec  string
		password  string
		expectErr bool
	}{
		{"remove single page", pdf("sample2.pdf"), "1", "", false},
		{"remove range", pdf("sample4.pdf"), "1-2", "", false},
		{"remove combination", pdf("sample4.pdf"), "1,3", "", false},
		{"non-pdf input", img("sample.jpg"), "1", "", true},
		{"file does not exist", pdf("ghost.pdf"), "1", "", true},
		{"invalid page spec", pdf("sample1.pdf"), "abc", "", true},
		{"protected pdf correct password", pdf("sample_protected.pdf"), "1", "samplefiles", false},
		{"protected pdf no password", pdf("sample_protected.pdf"), "1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tmp(t)
			_, err := RemovePages(tt.input, out, tt.pageSpec, tt.password)
			assertErr(t, err, tt.expectErr)
		})
	}
}

// ── Split / BuildSplitJobs ────────────────────────────────────────────────────

func TestBuildSplitJobs(t *testing.T) {
	tests := []struct {
		name      string
		selector  string
		numPages  int
		opts      SplitOptions
		wantJobs  int
		expectErr bool
	}{
		{"boundary mode page 1", "1", 3, SplitOptions{}, 2, false},
		{"boundary mode mid", "2", 4, SplitOptions{}, 2, false},
		{"boundary out of bounds", "5", 4, SplitOptions{}, 0, true},
		{"boundary zero", "0", 4, SplitOptions{}, 0, true},
		{"extract single page", "2", 4, SplitOptions{Extract: true}, 1, false},
		{"extract range", "1-2", 4, SplitOptions{Extract: true}, 1, false},
		{"extract multiple segments", "1,3", 4, SplitOptions{Extract: true}, 2, false},
		{"extract odd pages", "1-4", 4, SplitOptions{Extract: true, Odd: true}, 1, false},
		{"extract even pages", "1-4", 4, SplitOptions{Extract: true, Even: true}, 1, false},
		{"invalid selector in boundary mode", "abc", 4, SplitOptions{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs, err := BuildSplitJobs(tt.selector, tt.numPages, tt.opts)
			assertErr(t, err, tt.expectErr)
			if err == nil && len(jobs) != tt.wantJobs {
				t.Errorf("expected %d jobs, got %d", tt.wantJobs, len(jobs))
			}
		})
	}
}

func TestSplit(t *testing.T) {
	t.Run("boundary split", func(t *testing.T) {
		input := pdf("sample2.pdf")
		n, err := PageCount(input, "")
		if err != nil {
			t.Fatal(err)
		}

		jobs, err := BuildSplitJobs("1", n, SplitOptions{})
		if err != nil {
			t.Fatal(err)
		}

		dir := t.TempDir()
		paths := []string{
			filepath.Join(dir, "left.pdf"),
			filepath.Join(dir, "right.pdf"),
		}

		results, err := Split(input, paths, jobs, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("extract mode", func(t *testing.T) {
		input := pdf("sample4.pdf")
		n, err := PageCount(input, "")
		if err != nil {
			t.Fatal(err)
		}

		jobs, err := BuildSplitJobs("1,3", n, SplitOptions{Extract: true})
		if err != nil {
			t.Fatal(err)
		}

		dir := t.TempDir()
		paths := []string{
			filepath.Join(dir, "seg1.pdf"),
			filepath.Join(dir, "seg2.pdf"),
		}

		results, err := Split(input, paths, jobs, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("non-pdf input", func(t *testing.T) {
		jobs := []SplitJob{{Label: "001-001", Pages: []int{1}}}
		dir := t.TempDir()
		_, err := Split(img("sample.jpg"), []string{filepath.Join(dir, "out.pdf")}, jobs, "")
		if err == nil {
			t.Error("expected error for non-pdf input, got nil")
		}
	})
}

// ── PagesToSelectionTokens ────────────────────────────────────────────────────

func TestPagesToSelectionTokens(t *testing.T) {
	tests := []struct {
		name  string
		pages []int
		want  []string
	}{
		{"empty", []int{}, nil},
		{"single page", []int{3}, []string{"3"}},
		{"contiguous range", []int{1, 2, 3}, []string{"1-3"}},
		{"two separate pages", []int{1, 3}, []string{"1", "3"}},
		{"mixed", []int{1, 2, 4, 5, 7}, []string{"1-2", "4-5", "7"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PagesToSelectionTokens(tt.pages)
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// ── GetFileInfo ───────────────────────────────────────────────────────────────

func TestGetFileInfo(t *testing.T) {
	t.Run("valid pdf", func(t *testing.T) {
		info, err := GetFileInfo(pdf("sample1.pdf"), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.PageCount < 1 {
			t.Errorf("expected at least 1 page, got %d", info.PageCount)
		}
		if info.Bytes == 0 {
			t.Error("expected non-zero file size")
		}
		if info.Name == "" {
			t.Error("expected non-empty name")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := GetFileInfo(pdf("ghost.pdf"), "")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

// ── PageCount ─────────────────────────────────────────────────────────────────

func TestPageCount(t *testing.T) {
	t.Run("valid pdf", func(t *testing.T) {
		n, err := PageCount(pdf("sample1.pdf"), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n < 1 {
			t.Errorf("expected at least 1 page, got %d", n)
		}
	})

	t.Run("protected pdf", func(t *testing.T) {
		n, err := PageCount(pdf("sample_protected.pdf"), "samplefiles")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n < 1 {
			t.Errorf("expected at least 1 page, got %d", n)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := PageCount(pdf("ghost.pdf"), "")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func assertErr(t *testing.T, err error, expectErr bool) {
	t.Helper()
	if expectErr && err == nil {
		t.Error("expected error but got nil")
	}
	if !expectErr && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
