package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOutputPathForRmPage(t *testing.T) {
	t.Run("no conflict returns original", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "remove.pdf")

		got := resolveOutputPath(path)
		if got != path {
			t.Fatalf("resolveOutputPath() = %s, want %s", got, path)
		}
	})

	t.Run("single conflict returns (1)", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "remove.pdf")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("failed creating fixture file: %v", err)
		}

		got := resolveOutputPath(path)
		want := filepath.Join(dir, "remove (1).pdf")
		if got != want {
			t.Fatalf("resolveOutputPath() = %s, want %s", got, want)
		}
	})

	t.Run("multiple conflicts returns next index", func(t *testing.T) {
		dir := t.TempDir()
		fixtures := []string{
			filepath.Join(dir, "report.pdf"),
			filepath.Join(dir, "report (1).pdf"),
			filepath.Join(dir, "report (2).pdf"),
		}

		for _, f := range fixtures {
			if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
				t.Fatalf("failed creating fixture file %s: %v", f, err)
			}
		}

		got := resolveOutputPath(filepath.Join(dir, "report.pdf"))
		want := filepath.Join(dir, "report (3).pdf")
		if got != want {
			t.Fatalf("resolveOutputPath() = %s, want %s", got, want)
		}
	})
}