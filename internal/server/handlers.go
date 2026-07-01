package server

import (
	"archive/zip"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brendreyes/pdforge/internal/pdfops"
)

// ── Result store ──────────────────────────────────────────────────────────────

type operationResult struct {
	files   []string
	tempDir string
}

type resultStore struct {
	mu   sync.Mutex
	data map[string]*operationResult
}

func newResultStore() *resultStore {
	return &resultStore{data: make(map[string]*operationResult)}
}

func (s *resultStore) set(token string, r *operationResult) {
	s.mu.Lock()
	s.data[token] = r
	s.mu.Unlock()
}

func (s *resultStore) get(token string) (*operationResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data[token]
	return r, ok
}

func (s *resultStore) delete(token string) {
	s.mu.Lock()
	delete(s.data, token)
	s.mu.Unlock()
}

// ── Operation runner ──────────────────────────────────────────────────────────

// runOp executes fn in the current goroutine, publishing SSE events and storing
// the result. Call it inside a go statement from each handler.
func (s *Server) runOp(token, tempDir string, fn func() ([]string, error)) {
	s.broker.publish(token, sseEvent{Type: "started"})

	files, err := fn()
	if err != nil {
		s.broker.publish(token, sseEvent{Type: "error", Message: err.Error()})
		s.broker.closeToken(token)
		os.RemoveAll(tempDir)
		return
	}

	s.store.set(token, &operationResult{files: files, tempDir: tempDir})
	s.broker.publish(token, sseEvent{Type: "done", Token: token})
	s.broker.closeToken(token)

	// Auto-cleanup if the result is never downloaded within 10 minutes.
	time.AfterFunc(10*time.Minute, func() {
		if _, ok := s.store.get(token); ok {
			os.RemoveAll(tempDir)
			s.store.delete(token)
		}
	})
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (s *Server) handleOptimize(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	fh, err := singleFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	password := r.FormValue("password")

	tempDir, inputPath, token, ok := prepareOp(w, fh, 0)
	if !ok {
		return
	}

	outputPath := filepath.Join(tempDir, "optimized_"+filepath.Base(fh.Filename))

	s.broker.register(token)
	writeJSON(w, map[string]string{"token": token})

	go s.runOp(token, tempDir, func() ([]string, error) {
		if _, err := pdfops.Optimize(inputPath, outputPath, password); err != nil {
			return nil, err
		}
		return []string{outputPath}, nil
	})
}

func (s *Server) handleMerge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	fhs, err := multipleFiles(r, "files")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(fhs) < 2 {
		writeError(w, http.StatusBadRequest, "merge requires at least 2 files")
		return
	}

	password := r.FormValue("password")

	tempDir, err := os.MkdirTemp("", "pdforge-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp dir")
		return
	}

	inputPaths, err := saveFiles(fhs, tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusInternalServerError, "failed to save uploads")
		return
	}

	outputPath := filepath.Join(tempDir, "merged.pdf")

	token, err := generateToken()
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	s.broker.register(token)
	writeJSON(w, map[string]string{"token": token})

	go s.runOp(token, tempDir, func() ([]string, error) {
		if _, err := pdfops.Merge(inputPaths, outputPath, password); err != nil {
			return nil, err
		}
		return []string{outputPath}, nil
	})
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	fhs, err := multipleFiles(r, "files")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tempDir, err := os.MkdirTemp("", "pdforge-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp dir")
		return
	}

	inputPaths, err := saveFiles(fhs, tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusInternalServerError, "failed to save uploads")
		return
	}

	outputPath := filepath.Join(tempDir, "converted.pdf")

	token, err := generateToken()
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	s.broker.register(token)
	writeJSON(w, map[string]string{"token": token})

	go s.runOp(token, tempDir, func() ([]string, error) {
		if _, err := pdfops.Convert(inputPaths, outputPath); err != nil {
			return nil, err
		}
		return []string{outputPath}, nil
	})
}

func (s *Server) handleRotate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	fh, err := singleFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	degreesStr := r.FormValue("degrees")
	degrees, convErr := strconv.Atoi(degreesStr)
	if convErr != nil || degrees == 0 || degrees%90 != 0 {
		writeError(w, http.StatusBadRequest, "degrees must be a non-zero multiple of 90")
		return
	}

	pageSpec := r.FormValue("page_spec")
	password := r.FormValue("password")

	tempDir, inputPath, token, ok := prepareOp(w, fh, 0)
	if !ok {
		return
	}

	outputPath := filepath.Join(tempDir, "rotated_"+filepath.Base(fh.Filename))

	s.broker.register(token)
	writeJSON(w, map[string]string{"token": token})

	go s.runOp(token, tempDir, func() ([]string, error) {
		if _, err := pdfops.Rotate(inputPath, outputPath, degrees, pageSpec, password); err != nil {
			return nil, err
		}
		return []string{outputPath}, nil
	})
}

func (s *Server) handleRmpage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	fh, err := singleFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pageSpec := strings.TrimSpace(r.FormValue("page_spec"))
	if pageSpec == "" {
		writeError(w, http.StatusBadRequest, "page_spec is required")
		return
	}

	password := r.FormValue("password")

	tempDir, inputPath, token, ok := prepareOp(w, fh, 0)
	if !ok {
		return
	}

	outputPath := filepath.Join(tempDir, "removed_"+filepath.Base(fh.Filename))

	s.broker.register(token)
	writeJSON(w, map[string]string{"token": token})

	go s.runOp(token, tempDir, func() ([]string, error) {
		if _, err := pdfops.RemovePages(inputPath, outputPath, pageSpec, password); err != nil {
			return nil, err
		}
		return []string{outputPath}, nil
	})
}

func (s *Server) handleSplit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	fh, err := singleFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	selector := strings.TrimSpace(r.FormValue("selector"))
	if selector == "" {
		writeError(w, http.StatusBadRequest, "selector is required")
		return
	}

	extract := r.FormValue("extract") == "true"
	odd := r.FormValue("odd") == "true"
	even := r.FormValue("even") == "true"
	password := r.FormValue("password")

	tempDir, inputPath, token, ok := prepareOp(w, fh, 0)
	if !ok {
		return
	}

	// Build split jobs eagerly so validation errors go back as HTTP errors,
	// not buried in an SSE error event.
	numPages, err := pdfops.PageCount(inputPath, password)
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	jobs, err := pdfops.BuildSplitJobs(selector, numPages, pdfops.SplitOptions{
		Extract: extract,
		Odd:     odd,
		Even:    even,
	})
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	outputPaths := make([]string, len(jobs))
	for i, job := range jobs {
		outputPaths[i] = filepath.Join(tempDir, fmt.Sprintf("split_%s.pdf", job.Label))
	}

	s.broker.register(token)
	writeJSON(w, map[string]string{"token": token})

	go s.runOp(token, tempDir, func() ([]string, error) {
		if _, err := pdfops.Split(inputPath, outputPaths, jobs, password); err != nil {
			return nil, err
		}
		return outputPaths, nil
	})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	res, ok := s.store.get(token)
	if !ok {
		http.Error(w, "result not found or already downloaded", http.StatusNotFound)
		return
	}

	var serveErr error
	switch len(res.files) {
	case 0:
		http.Error(w, "no output files", http.StatusInternalServerError)
		return
	case 1:
		serveErr = serveFile(w, r, res.files[0])
	default:
		serveErr = serveZip(w, res.files)
	}

	if serveErr != nil {
		// Headers may already be sent; log only.
		_ = serveErr
		return
	}

	// Clean up after the response is fully written.
	go func() {
		os.RemoveAll(res.tempDir)
		s.store.delete(token)
	}()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// prepareOp creates a temp dir, saves the uploaded file, and generates a token.
// Returns false and writes an error response on failure.
func prepareOp(w http.ResponseWriter, fh *multipart.FileHeader, idx int) (tempDir, inputPath, token string, ok bool) {
	var err error
	tempDir, err = os.MkdirTemp("", "pdforge-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp dir")
		return
	}

	inputPath, err = saveFile(fh, tempDir, idx)
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusInternalServerError, "failed to save upload")
		return
	}

	token, err = generateToken()
	if err != nil {
		os.RemoveAll(tempDir)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	ok = true
	return
}

func singleFile(r *http.Request, field string) (*multipart.FileHeader, error) {
	fhs, exists := r.MultipartForm.File[field]
	if !exists || len(fhs) == 0 {
		return nil, fmt.Errorf("missing required file field '%s'", field)
	}
	return fhs[0], nil
}

func multipleFiles(r *http.Request, field string) ([]*multipart.FileHeader, error) {
	fhs, exists := r.MultipartForm.File[field]
	if !exists || len(fhs) == 0 {
		return nil, fmt.Errorf("missing required file field '%s'", field)
	}
	return fhs, nil
}

// saveFile writes the uploaded file to dir/<idx>_<originalname>.
// The index prefix prevents collisions when multiple files share a name.
func saveFile(fh *multipart.FileHeader, dir string, idx int) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	filename := fmt.Sprintf("%d_%s", idx, filepath.Base(fh.Filename))
	dst := filepath.Join(dir, filename)

	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, src)
	return dst, err
}

func saveFiles(fhs []*multipart.FileHeader, dir string) ([]string, error) {
	paths := make([]string, 0, len(fhs))
	for i, fh := range fhs {
		path, err := saveFile(fh, dir, i)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(path)))
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), f)
	return nil
}

func serveZip(w http.ResponseWriter, files []string) error {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="pdforge_result.zip"`)

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, path := range files {
		fw, err := zw.Create(filepath.Base(path))
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(fw, f)
		f.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
