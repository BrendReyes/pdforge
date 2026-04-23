package cmd

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type webServer struct {
	port         int
	tempRoot     string
	frontendRoot string
	jobs         map[string]webJob
	jobsMu       sync.RWMutex
	template     *template.Template
}

const (
	webUploadMaxBytes    int64 = 256 << 20
	webUploadMemoryBytes int64 = 8 << 20
	webCSRFCookieName          = "pdforge_csrf"
)

type webJob struct {
	ID      string
	Action  string
	Summary string
	Created time.Time
	Files   []webJobFile
}

type webJobFile struct {
	Path   string
	Report *FileInfoReport
}

type webPageData struct {
	Port      int
	Alert     string
	AlertKind string
	Result    *webResult
	CSRFToken string
	Now       string
}

type webResult struct {
	Action  string          `json:"action"`
	Summary string          `json:"summary"`
	Items   []webResultItem `json:"items"`
}

type webResultItem struct {
	Name        string `json:"name"`
	Size        string `json:"size"`
	Pages       int    `json:"pages"`
	DownloadURL string `json:"downloadURL"`
}

type webAPIResponse struct {
	OK     bool       `json:"ok"`
	Alert  string     `json:"alert,omitempty"`
	Error  string     `json:"error,omitempty"`
	Result *webResult `json:"result,omitempty"`
}

type webMergeForm struct {
	Output string
	Dir    string
}

type webSplitForm struct {
	Selector string
	PageFlag string
	Output   string
	Dir      string
	Extract  bool
	Odd      bool
	Even     bool
	Verbose  bool
}

type webRemoveForm struct {
	Page   string
	Output string
	Dir    string
}

type webCompressForm struct {
	Output           string
	Dir              string
	ImageMode        string
	ImageMaxDim      int
	ImageJPEGQuality int
}

func newWebServer(port int) (*webServer, error) {
	tempRoot, err := os.MkdirTemp("", "pdforge-web-")
	if err != nil {
		return nil, err
	}

	frontendRoot := detectFrontendRoot()

	tpl, err := template.New("pdforge-web").Parse(pdforgeWebTemplate)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, err
	}

	return &webServer{
		port:         port,
		tempRoot:     tempRoot,
		frontendRoot: frontendRoot,
		jobs:         make(map[string]webJob),
		template:     tpl,
	}, nil
}

func (s *webServer) Close() {
	_ = os.RemoveAll(s.tempRoot)
}

func (s *webServer) Run(cmdOut io.Writer) error {
	return s.run(cmdOut, false)
}

func (s *webServer) RunWithAutoOpen(cmdOut io.Writer) error {
	return s.run(cmdOut, true)
}

func (s *webServer) run(cmdOut io.Writer, autoOpen bool) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/csrf", s.handleCSRF)
	mux.HandleFunc("/download", s.handleDownload)
	mux.HandleFunc("/api/merge", s.handleMerge)
	mux.HandleFunc("/api/split", s.handleSplit)
	mux.HandleFunc("/api/remove", s.handleRemove)
	mux.HandleFunc("/api/compress", s.handleCompress)

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port))

	server := &http.Server{
		Addr:              addr,
		Handler:           s.withSecurityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve(ln)
	}()

	if err := waitForHTTPReady("http://127.0.0.1:"+strconv.Itoa(s.port)+"/healthz", 5*time.Second); err != nil {
		_ = server.Shutdown(context.Background())
		return err
	}

	fmt.Fprintf(cmdOut, "pdforge web UI running at http://localhost:%d\n", s.port)
	if autoOpen {
		if err := openBrowser("http://localhost:" + strconv.Itoa(s.port)); err != nil {
			fmt.Fprintf(cmdOut, "Browser auto-open skipped: %v\n", err)
		}
	}
	fmt.Fprintln(cmdOut, "Press Ctrl+C to stop.")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case sig := <-signalCh:
		fmt.Fprintf(cmdOut, "Shutting down after %s...\n", sig)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *webServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.applySecurityHeaders(w)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}

func (s *webServer) handleCSRF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.applySecurityHeaders(w)
	token := s.ensureCSRFToken(w, r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = fmt.Fprintf(w, `{"token":%q}`, token)
}

func (s *webServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if s.serveFrontendAsset(w, r) {
			return
		}
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.frontendRoot != "" {
		s.serveFrontendIndex(w, r)
		return
	}
	data := webPageData{
		Port:      s.port,
		CSRFToken: s.ensureCSRFToken(w, r),
		Now:       time.Now().Format(time.RFC1123),
	}
	s.renderPage(w, r, data)
}

func (s *webServer) handleMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, alert, err := s.processMerge(w, r)
	if wantsJSONResponse(r) {
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, webAPIResponse{OK: false, Alert: alert, Error: err.Error()})
			return
		}
		s.writeJSON(w, http.StatusOK, webAPIResponse{OK: true, Alert: alert, Result: result})
		return
	}

	s.renderPage(w, r, webPageData{
		Port:      s.port,
		Alert:     alert,
		AlertKind: alertKind(err),
		Result:    result,
		CSRFToken: s.ensureCSRFToken(w, r),
		Now:       time.Now().Format(time.RFC1123),
	})
}

func (s *webServer) handleSplit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, alert, err := s.processSplit(w, r)
	if wantsJSONResponse(r) {
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, webAPIResponse{OK: false, Alert: alert, Error: err.Error()})
			return
		}
		s.writeJSON(w, http.StatusOK, webAPIResponse{OK: true, Alert: alert, Result: result})
		return
	}

	s.renderPage(w, r, webPageData{
		Port:      s.port,
		Alert:     alert,
		AlertKind: alertKind(err),
		Result:    result,
		CSRFToken: s.ensureCSRFToken(w, r),
		Now:       time.Now().Format(time.RFC1123),
	})
}

func (s *webServer) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, alert, err := s.processRemove(w, r)
	if wantsJSONResponse(r) {
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, webAPIResponse{OK: false, Alert: alert, Error: err.Error()})
			return
		}
		s.writeJSON(w, http.StatusOK, webAPIResponse{OK: true, Alert: alert, Result: result})
		return
	}

	s.renderPage(w, r, webPageData{
		Port:      s.port,
		Alert:     alert,
		AlertKind: alertKind(err),
		Result:    result,
		CSRFToken: s.ensureCSRFToken(w, r),
		Now:       time.Now().Format(time.RFC1123),
	})
}

func (s *webServer) handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, alert, err := s.processCompress(w, r)
	if wantsJSONResponse(r) {
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, webAPIResponse{OK: false, Alert: alert, Error: err.Error()})
			return
		}
		s.writeJSON(w, http.StatusOK, webAPIResponse{OK: true, Alert: alert, Result: result})
		return
	}

	s.renderPage(w, r, webPageData{
		Port:      s.port,
		Alert:     alert,
		AlertKind: alertKind(err),
		Result:    result,
		CSRFToken: s.ensureCSRFToken(w, r),
		Now:       time.Now().Format(time.RFC1123),
	})
}

func (s *webServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	s.applySecurityHeaders(w)
	jobID := strings.TrimSpace(r.URL.Query().Get("job"))
	indexValue := strings.TrimSpace(r.URL.Query().Get("index"))
	if jobID == "" || indexValue == "" {
		http.Error(w, "missing job or index", http.StatusBadRequest)
		return
	}

	index, err := strconv.Atoi(indexValue)
	if err != nil || index < 0 {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	s.jobsMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobsMu.RUnlock()
	if !ok || index >= len(job.Files) {
		http.NotFound(w, r)
		return
	}

	file := job.Files[index]
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(file.Report.Name)))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, file.Path)
}

func (s *webServer) renderPage(w http.ResponseWriter, r *http.Request, data webPageData) {
	s.applySecurityHeaders(w)
	data.CSRFToken = s.ensureCSRFToken(w, r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.template.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *webServer) serveFrontendIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(s.frontendRoot, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, indexPath)
}

func (s *webServer) serveFrontendAsset(w http.ResponseWriter, r *http.Request) bool {
	if s.frontendRoot == "" {
		return false
	}

	cleaned := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if cleaned == "." || strings.HasPrefix(cleaned, "..") {
		return false
	}
	assetPath := filepath.Join(s.frontendRoot, cleaned)
	if _, err := os.Stat(assetPath); err != nil {
		return false
	}

	http.ServeFile(w, r, assetPath)
	return true
}

func detectFrontendRoot() string {
	candidates := []string{
		filepath.Join("web", "dist"),
		filepath.Join(".", "web", "dist"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil && !info.IsDir() {
			return candidate
		}
	}

	return ""
}

func (s *webServer) processMerge(w http.ResponseWriter, r *http.Request) (*webResult, string, error) {
	cleanup, err := s.prepareMultipartForm(w, r)
	if err != nil {
		return nil, "Unable to parse uploaded files.", err
	}
	defer cleanup()

	if err := s.requireCSRFToken(r); err != nil {
		return nil, "Unable to parse uploaded files.", err
	}

	files := r.MultipartForm.File["files"]
	if len(files) < 2 {
		return nil, "Merge needs at least two PDF files.", fmt.Errorf("merge requires at least two uploaded PDF files")
	}

	jobDir := filepath.Join(s.tempRoot, "merge-"+randomToken())
	inputDir := filepath.Join(jobDir, "inputs")
	outputDir := formValue(r, "dir")
	if outputDir == "" {
		outputDir = filepath.Join(jobDir, "output")
	}

	if err := ensureOutputDirectoryWeb(outputDir); err != nil {
		return nil, "Unable to prepare output directory.", err
	}

	inputs := make([]string, 0, len(files))
	for _, fh := range files {
		path, err := saveUploadedPDF(inputDir, fh)
		if err != nil {
			return nil, "Invalid merge input.", err
		}
		inputs = append(inputs, path)
	}

	outputName := strings.TrimSpace(formValue(r, "output"))
	if outputName == "" {
		outputName = "merged_" + time.Now().Format("20060102_150405") + ".pdf"
	}
	if strings.ToLower(filepath.Ext(outputName)) != ".pdf" {
		return nil, "Output file must end in .pdf.", fmt.Errorf("the file '%s' is invalid, must be '.pdf'", outputName)
	}

	outputPath := filepath.Join(outputDir, outputName)
	outputPath = resolveOutputPath(outputPath)
	if err := api.MergeCreateFile(inputs, outputPath, false, nil); err != nil {
		return nil, "Merge failed.", err
	}

	result, err := s.registerResult("Merge", "Combined uploaded PDFs into one file.", []string{outputPath})
	if err != nil {
		return nil, "Merge completed, but result registration failed.", err
	}
	return result, "Merge completed successfully.", nil
}

func (s *webServer) processSplit(w http.ResponseWriter, r *http.Request) (*webResult, string, error) {
	cleanup, err := s.prepareMultipartForm(w, r)
	if err != nil {
		return nil, "Unable to parse uploaded file.", err
	}
	defer cleanup()

	if err := s.requireCSRFToken(r); err != nil {
		return nil, "Invalid request token.", err
	}

	fileHeader := firstFile(r, "file")
	if fileHeader == nil {
		return nil, "Split needs one PDF file.", fmt.Errorf("missing uploaded PDF file")
	}

	jobDir := filepath.Join(s.tempRoot, "split-"+randomToken())
	inputPath, err := saveUploadedPDF(filepath.Join(jobDir, "input"), fileHeader)
	if err != nil {
		return nil, "Invalid split input.", err
	}

	pageFlag := strings.TrimSpace(formValue(r, "page"))
	positionalSelector := strings.TrimSpace(formValue(r, "selector"))
	if pageFlag != "" && positionalSelector != "" {
		return nil, "Use either selector or --page, not both.", fmt.Errorf("provide selector either as positional argument or via --page, not both")
	}

	selector := pageFlag
	if selector == "" {
		selector = positionalSelector
	}
	if selector == "" {
		return nil, "Split needs a selector.", fmt.Errorf("missing selector: provide [selector] or --page")
	}

	if strings.ToLower(filepath.Ext(inputPath)) != ".pdf" {
		return nil, "Split requires a PDF file.", fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(inputPath))
	}

	if err := api.ValidateFile(inputPath, nil); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "password") || strings.Contains(strings.ToLower(err.Error()), "encrypt") {
			return nil, "Encrypted PDFs are not supported yet.", fmt.Errorf("encrypted PDFs are not supported yet")
		}
		return nil, "Invalid PDF file.", fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(inputPath), err)
	}

	pageCount, err := api.PageCountFile(inputPath)
	if err != nil {
		return nil, "Unable to read page count.", err
	}

	extract := formChecked(r, "extract")
	odd := formChecked(r, "odd")
	even := formChecked(r, "even")
	verbose := formChecked(r, "verbose")

	if odd && even {
		return nil, "Odd and even cannot be used together.", fmt.Errorf("--odd and --even cannot be used together")
	}
	if !extract && (odd || even) {
		return nil, "Odd/even flags require extract mode.", fmt.Errorf("--odd/--even can only be used with --extract")
	}

	jobs, err := buildWebSplitJobs(selector, pageCount, extract, odd, even)
	if err != nil {
		return nil, "Unable to build split jobs.", err
	}

	outputDir := strings.TrimSpace(formValue(r, "dir"))
	if outputDir == "" {
		outputDir = filepath.Join(jobDir, "output")
	}

	outputName := strings.TrimSpace(formValue(r, "output"))
	paths, err := resolveWebSplitOutputs(inputPath, outputName, outputDir, jobs)
	if err != nil {
		return nil, "Unable to prepare split outputs.", err
	}

	for i, job := range jobs {
		selectedPages := pagesToSelectionTokens(job.Pages)
		if err := api.TrimFile(inputPath, paths[i], selectedPages, nil); err != nil {
			return nil, fmt.Sprintf("Failed writing %s.", filepath.Base(paths[i])), err
		}
	}

	summary := fmt.Sprintf("Created %d split file(s).", len(paths))
	if verbose {
		var lines []string
		for i, job := range jobs {
			lines = append(lines, fmt.Sprintf("%s -> pages %s", filepath.Base(paths[i]), pagesDisplay(job.Pages)))
		}
		summary = summary + " " + strings.Join(lines, " | ")
	}

	result, err := s.registerResult("Split", summary, paths)
	if err != nil {
		return nil, "Split completed, but result registration failed.", err
	}
	return result, "Split completed successfully.", nil
}

func (s *webServer) processRemove(w http.ResponseWriter, r *http.Request) (*webResult, string, error) {
	cleanup, err := s.prepareMultipartForm(w, r)
	if err != nil {
		return nil, "Unable to parse uploaded file.", err
	}
	defer cleanup()

	if err := s.requireCSRFToken(r); err != nil {
		return nil, "Invalid request token.", err
	}

	fileHeader := firstFile(r, "file")
	if fileHeader == nil {
		return nil, "Remove pages needs one PDF file.", fmt.Errorf("missing uploaded PDF file")
	}

	jobDir := filepath.Join(s.tempRoot, "remove-"+randomToken())
	inputPath, err := saveUploadedPDF(filepath.Join(jobDir, "input"), fileHeader)
	if err != nil {
		return nil, "Invalid remove-pages input.", err
	}

	pageSpec := strings.TrimSpace(formValue(r, "page"))
	if pageSpec == "" {
		pageSpec = strings.TrimSpace(formValue(r, "selector"))
	}
	if pageSpec == "" {
		return nil, "Remove pages needs a page selector.", fmt.Errorf("missing page selector: provide [page_selector] or --page")
	}

	if _, err := parsePageSpecification(pageSpec); err != nil {
		return nil, "Invalid page selector.", fmt.Errorf("invalid page specification: %w", err)
	}

	if _, err := api.ParsePageSelection(pageSpec); err != nil {
		return nil, "Invalid page selector.", fmt.Errorf("invalid page specification: %w", err)
	}
	selectedPages, err := api.ParsePageSelection(pageSpec)
	if err != nil {
		return nil, "Invalid page selector.", fmt.Errorf("invalid page specification: %w", err)
	}

	outputName := strings.TrimSpace(formValue(r, "output"))
	if outputName == "" {
		outputName = "remove.pdf"
	}
	if strings.ToLower(filepath.Ext(outputName)) != ".pdf" {
		return nil, "Output file must end in .pdf.", fmt.Errorf("the file '%s' is invalid, must be '.pdf'", outputName)
	}

	outputDir := strings.TrimSpace(formValue(r, "dir"))
	if outputDir == "" {
		outputDir = filepath.Join(jobDir, "output")
	}
	if err := ensureOutputDirectoryWeb(outputDir); err != nil {
		return nil, "Unable to prepare output directory.", err
	}

	outputPath := filepath.Join(outputDir, outputName)
	outputPath = resolveOutputPath(outputPath)
	if err := api.RemovePagesFile(inputPath, outputPath, selectedPages, nil); err != nil {
		return nil, "Page removal failed.", err
	}

	result, err := s.registerResult("Remove Pages", "Removed the selected pages from the uploaded PDF.", []string{outputPath})
	if err != nil {
		return nil, "Page removal completed, but result registration failed.", err
	}
	return result, "Page removal completed successfully.", nil
}

func (s *webServer) processCompress(w http.ResponseWriter, r *http.Request) (*webResult, string, error) {
	formCleanup, err := s.prepareMultipartForm(w, r)
	if err != nil {
		return nil, "Unable to parse uploaded file.", err
	}
	defer formCleanup()

	if err := s.requireCSRFToken(r); err != nil {
		return nil, "Invalid request token.", err
	}

	fileHeader := firstFile(r, "file")
	if fileHeader == nil {
		return nil, "Compression needs one PDF file.", fmt.Errorf("missing uploaded PDF file")
	}

	jobDir := filepath.Join(s.tempRoot, "compress-"+randomToken())
	inputPath, err := saveUploadedPDF(filepath.Join(jobDir, "input"), fileHeader)
	if err != nil {
		return nil, "Invalid compression input.", err
	}

	if strings.ToLower(filepath.Ext(inputPath)) != ".pdf" {
		return nil, "Compression requires a PDF file.", fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(inputPath))
	}

	if err := api.ValidateFile(inputPath, nil); err != nil {
		return nil, "Invalid PDF file.", fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(inputPath), err)
	}

	outputName := strings.TrimSpace(formValue(r, "output"))
	if outputName == "" {
		outputName = defaultCompressedOutput(inputPath)
		outputName = filepath.Base(outputName)
	}
	if strings.ToLower(filepath.Ext(outputName)) != ".pdf" {
		return nil, "Output file must end in .pdf.", fmt.Errorf("the file '%s' is invalid, must be '.pdf'", outputName)
	}

	outputDir := strings.TrimSpace(formValue(r, "dir"))
	if outputDir == "" {
		outputDir = filepath.Join(jobDir, "output")
	}
	if err := ensureOutputDirectoryWeb(outputDir); err != nil {
		return nil, "Unable to prepare output directory.", err
	}

	mode, err := normalizeWebCompressImageMode(formValue(r, "image_mode"))
	if err != nil {
		return nil, "Invalid image mode.", err
	}

	maxDim := parseIntFormValue(r, "image_max_dimension")
	jpegQuality := parseIntFormValue(r, "image_jpeg_quality")

	processedInput := inputPath
	var cleanup func()
	if mode != compressImageModeOff {
		profile := resolveImageModeProfile(mode)
		if maxDim <= 0 {
			maxDim = profile.maxDimension
		}
		if jpegQuality <= 0 {
			jpegQuality = profile.jpegQuality
		}

		processedInput, _, cleanup, err = preprocessPDFImages(inputPath, maxDim, jpegQuality)
		if err != nil {
			return nil, "Image preprocessing failed.", err
		}
		if cleanup != nil {
			defer cleanup()
		}
	}

	outputPath := filepath.Join(outputDir, outputName)
	outputPath = resolveOutputPath(outputPath)
	if err := api.OptimizeFile(processedInput, outputPath, nil); err != nil {
		return nil, "Compression failed.", err
	}

	result, err := s.registerResult("Compress", "Reduced the PDF file size and generated a downloadable output.", []string{outputPath})
	if err != nil {
		return nil, "Compression completed, but result registration failed.", err
	}
	return result, "Compression completed successfully.", nil
}

func (s *webServer) registerResult(action, summary string, paths []string) (*webResult, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no output files produced")
	}

	jobID := randomToken()
	stored := webJob{
		ID:      jobID,
		Action:  action,
		Summary: summary,
		Created: time.Now(),
	}

	result := &webResult{
		Action:  action,
		Summary: summary,
		Items:   make([]webResultItem, 0, len(paths)),
	}

	for i, path := range paths {
		report, err := GetFileInfo(path)
		if err != nil {
			return nil, err
		}
		stored.Files = append(stored.Files, webJobFile{Path: path, Report: report})
		result.Items = append(result.Items, webResultItem{
			Name:        report.Name,
			Size:        formatBytes(report.Bytes),
			Pages:       report.PageCount,
			DownloadURL: fmt.Sprintf("/download?job=%s&index=%d", url.QueryEscape(jobID), i),
		})
	}

	s.jobsMu.Lock()
	s.jobs[jobID] = stored
	s.jobsMu.Unlock()

	return result, nil
}

func (s *webServer) prepareMultipartForm(w http.ResponseWriter, r *http.Request) (func(), error) {
	r.Body = http.MaxBytesReader(w, r.Body, webUploadMaxBytes)
	if err := r.ParseMultipartForm(webUploadMemoryBytes); err != nil {
		return func() {}, err
	}

	cleanup := func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}

	return cleanup, nil
}

func (s *webServer) ensureCSRFToken(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(webCSRFCookieName); err == nil && len(cookie.Value) >= 16 {
		return cookie.Value
	}

	token := randomToken()
	http.SetCookie(w, &http.Cookie{
		Name:     webCSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
	return token
}

func (s *webServer) writeJSON(w http.ResponseWriter, status int, payload webAPIResponse) {
	s.applySecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *webServer) requireCSRFToken(r *http.Request) error {
	cookie, err := r.Cookie(webCSRFCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return fmt.Errorf("missing request token")
	}

	formToken := strings.TrimSpace(r.FormValue("csrf_token"))
	if formToken == "" {
		return fmt.Errorf("missing request token")
	}

	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(formToken)) != 1 {
		return fmt.Errorf("invalid request token")
	}

	return nil
}

func (s *webServer) applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'; img-src 'self' data:; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
}

func wantsJSONResponse(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/json")
}

func (s *webServer) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.applySecurityHeaders(w)
		next.ServeHTTP(w, r)
	})
}

func waitForHTTPReady(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 500 * time.Millisecond}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready within %s", timeout)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func buildWebSplitJobs(selector string, pageCount int, extract, odd, even bool) ([]splitJob, error) {
	if extract {
		return buildWebExtractJobs(selector, pageCount, odd, even)
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

func buildWebExtractJobs(selector string, pageCount int, odd bool, even bool) ([]splitJob, error) {
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

		pages = filterParity(pages, odd, even)
		if len(pages) == 0 {
			return nil, fmt.Errorf("page selection '%s' has no pages after applying odd/even filter", token)
		}

		label := formatSegmentLabel(pages, pad)
		jobs = append(jobs, splitJob{Label: label, Pages: pages})
	}

	return jobs, nil
}

func resolveWebSplitOutputs(inputPath string, outputName string, outputDir string, jobs []splitJob) ([]string, error) {
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no split jobs to process")
	}

	if err := ensureOutputDirectoryWeb(outputDir); err != nil {
		return nil, err
	}

	if len(jobs) == 1 && outputName != "" {
		out := outputName
		if strings.ToLower(filepath.Ext(out)) != ".pdf" {
			return nil, fmt.Errorf("the file '%s' is invalid, must be '.pdf'", out)
		}
		if !filepath.IsAbs(out) {
			out = filepath.Join(outputDir, out)
		}
		if err := ensureOutputDirectoryWeb(filepath.Dir(out)); err != nil {
			return nil, err
		}
		return []string{resolveOutputPath(out)}, nil
	}

	prefix := "split"
	if outputName != "" {
		prefix = strings.TrimSpace(outputName)
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
		candidate := filepath.Join(outputDir, name)
		paths = append(paths, resolveOutputPath(candidate))
	}

	return paths, nil
}

func saveUploadedPDF(destDir string, fh *multipart.FileHeader) (string, error) {
	if fh == nil {
		return "", fmt.Errorf("missing uploaded PDF")
	}
	if strings.ToLower(filepath.Ext(fh.Filename)) != ".pdf" {
		return "", fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(fh.Filename))
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	cleanName := filepath.Base(fh.Filename)
	if cleanName == "." || cleanName == string(os.PathSeparator) {
		cleanName = "input.pdf"
	}

	destPath := filepath.Join(destDir, cleanName)
	if err := copyReaderToFile(destPath, src); err != nil {
		return "", err
	}

	if err := api.ValidateFile(destPath, nil); err != nil {
		return "", fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(destPath), err)
	}

	return destPath, nil
}

func copyReaderToFile(path string, reader io.Reader) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

func firstFile(r *http.Request, field string) *multipart.FileHeader {
	if r.MultipartForm == nil {
		return nil
	}
	files := r.MultipartForm.File[field]
	if len(files) == 0 {
		return nil
	}
	return files[0]
}

func formValue(r *http.Request, key string) string {
	return strings.TrimSpace(r.FormValue(key))
}

func formChecked(r *http.Request, key string) bool {
	v := strings.ToLower(strings.TrimSpace(r.FormValue(key)))
	return v == "on" || v == "true" || v == "1" || v == "yes"
}

func parseIntFormValue(r *http.Request, key string) int {
	v := strings.TrimSpace(r.FormValue(key))
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func ensureOutputDirectoryWeb(dir string) error {
	if dir == "" {
		return nil
	}
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("'%s' is not a directory", dir)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("unable to access directory '%s': %w", dir, err)
	}
	return os.MkdirAll(dir, 0o755)
}

func randomToken() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(b[:]), "=")
}

func formatBytes(bytesValue int64) string {
	if bytesValue <= 0 {
		return "0.00 KB"
	}
	mb := float64(bytesValue) / (1024 * 1024)
	if mb >= 1 {
		return fmt.Sprintf("%.2f MB", mb)
	}
	kb := float64(bytesValue) / (1024)
	return fmt.Sprintf("%.2f KB", kb)
}

func alertKind(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func normalizeWebCompressImageMode(raw string) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "" || mode == compressImageModeOff {
		return compressImageModeOff, nil
	}

	if mode == compressImageModeExperimental {
		mode = compressImageModeAggressive
	}

	switch mode {
	case compressImageModeReadable, compressImageModeBalanced, compressImageModeAggressive:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid --image-mode '%s', expected: off|readable|balanced|aggressive", raw)
	}
}

const pdforgeWebTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>pdforge web</title>
  <style>
    :root {
      --bg: #f8fafc;
      --panel: rgba(255,255,255,.8);
      --card: #ffffff;
      --ink: #0f172a;
      --muted: #64748b;
      --line: rgba(148,163,184,.25);
      --brand: #2563eb;
      --brand-soft: #eff6ff;
      --ok: #16a34a;
      --bad: #dc2626;
      --radius-xl: 24px;
      --radius-lg: 18px;
      --radius-md: 12px;
      --shadow: 0 18px 42px rgba(15,23,42,.08);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(1100px 500px at 12% -10%, #dbeafe 0%, transparent 60%),
        radial-gradient(900px 420px at 100% -10%, #d5f3ff 0%, transparent 62%),
        linear-gradient(180deg, #ffffff 0%, var(--bg) 100%);
      min-height: 100vh;
    }
    .wrap { max-width: 1280px; margin: 0 auto; padding: 28px 18px 56px; }
    .shell {
      display: grid;
      grid-template-columns: 1.25fr .75fr;
      gap: 16px;
      margin-top: 18px;
    }
    .topbar, .card, .panel {
      background: var(--panel);
      backdrop-filter: blur(16px);
      border: 1px solid var(--line);
      border-radius: var(--radius-xl);
      box-shadow: var(--shadow);
    }
    .topbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 14px 16px;
      position: sticky;
      top: 14px;
      z-index: 10;
    }
    .brand { display: flex; align-items: center; gap: 10px; font-weight: 800; }
    .logo {
      width: 32px; height: 32px; border-radius: 11px; display: grid; place-items: center;
      color: #fff; font-weight: 900; background: linear-gradient(145deg, #0ea5e9, #2563eb);
    }
    .chips { display: flex; gap: 8px; flex-wrap: wrap; }
    .chip {
      padding: 7px 10px; border-radius: 999px; border: 1px solid var(--line);
      background: rgba(255,255,255,.9); color: var(--muted); font-size: 13px;
    }
    .hero { display: grid; grid-template-columns: 1.1fr .9fr; gap: 16px; }
    .card, .panel { padding: 18px; }
    h1 { margin: 10px 0 10px; font-size: clamp(2rem, 4vw, 3.4rem); line-height: 1.04; letter-spacing: -.03em; }
    h2 { margin: 0; font-size: 18px; }
    p { margin: 0; color: var(--muted); line-height: 1.6; }
    .label {
      display: inline-flex; align-items: center; gap: 7px; width: fit-content;
      padding: 6px 10px; border-radius: 999px; font-size: 12px; font-weight: 700;
      color: #1d4ed8; background: var(--brand-soft); border: 1px solid rgba(37,99,235,.16);
    }
    .cta { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 18px; }
    .btn {
      appearance: none; border: 1px solid var(--line); background: #fff; color: var(--ink);
      border-radius: 14px; padding: 11px 14px; font-weight: 700; cursor: pointer; text-decoration: none;
      display: inline-flex; align-items: center; justify-content: center; gap: 8px;
    }
    .btn.primary { background: linear-gradient(145deg, #0ea5e9, #2563eb); color: #fff; border-color: transparent; }
    .btn.ghost { background: rgba(255,255,255,.8); }
    .btn:hover { transform: translateY(-1px); }
    .tabs { margin-top: 16px; }
    .tab-list {
      display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; padding: 8px; border-radius: 18px;
      background: rgba(255,255,255,.65); border: 1px solid var(--line);
    }
    .tab-list button {
      border: 0; background: transparent; border-radius: 12px; padding: 11px 10px; font-weight: 700; color: var(--muted); cursor: pointer;
    }
    .tab-list button.active { background: #fff; color: var(--ink); box-shadow: 0 8px 18px rgba(15,23,42,.08); }
    .tab-panel { display: none; margin-top: 12px; }
    .tab-panel.active { display: block; }
    .form-grid { display: grid; gap: 12px; }
    .field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
    .field, .field-wide { display: grid; gap: 6px; }
    .field-wide { grid-column: 1 / -1; }
    label { font-size: 13px; font-weight: 700; color: #334155; }
    .hint { font-size: 12px; color: var(--muted); }
    input[type="text"], input[type="number"], select, textarea {
      width: 100%; border: 1px solid var(--line); border-radius: 12px; background: #fff; color: var(--ink);
      padding: 11px 12px; font: inherit;
    }
    textarea { min-height: 86px; resize: vertical; }
    input[type="file"] { width: 100%; }
    .checks { display: flex; flex-wrap: wrap; gap: 10px; }
    .check {
      display: inline-flex; align-items: center; gap: 8px; border: 1px solid var(--line); background: #fff;
      border-radius: 999px; padding: 8px 10px; font-size: 13px;
    }
    .check input { accent-color: var(--brand); }
    .submit-row { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; margin-top: 4px; }
    .badge {
      display: inline-flex; align-items: center; gap: 7px; border-radius: 999px; padding: 6px 10px; font-size: 12px; font-weight: 700;
    }
    .badge.success { color: var(--ok); background: #f0fdf4; border: 1px solid rgba(22,163,74,.18); }
    .badge.error { color: var(--bad); background: #fef2f2; border: 1px solid rgba(220,38,38,.18); }
    .outputs { display: grid; gap: 10px; margin-top: 12px; }
    .output {
      padding: 12px; border-radius: 14px; border: 1px solid var(--line); background: #fff;
      display: flex; align-items: center; justify-content: space-between; gap: 12px;
    }
    .output strong { display: block; margin-bottom: 4px; }
    .side { display: grid; gap: 12px; align-content: start; }
    .mini { display: grid; gap: 10px; }
    .mini-card {
      padding: 14px; border-radius: 16px; border: 1px solid var(--line); background: rgba(255,255,255,.9);
    }
    .flag-list { display: grid; gap: 8px; margin-top: 10px; }
    .flag-item {
      display: flex; justify-content: space-between; gap: 12px; padding: 10px 12px; border-radius: 12px;
      background: #fff; border: 1px solid var(--line); font-size: 13px;
    }
    .footer { margin-top: 16px; color: var(--muted); font-size: 13px; text-align: center; }
    @media (max-width: 980px) {
      .shell, .hero { grid-template-columns: 1fr; }
      .tab-list { grid-template-columns: repeat(2, 1fr); }
    }
    @media (max-width: 620px) {
      .wrap { padding: 16px 12px 42px; }
      .topbar { border-radius: 18px; }
      .field-grid { grid-template-columns: 1fr; }
      .tab-list { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <header class="topbar">
      <div class="brand"><span class="logo">PF</span><span>pdforge web</span></div>
      <div class="chips">
        <span class="chip">shadcn-inspired controls</span>
        <span class="chip">local PDF processing</span>
        <span class="chip">port {{.Port}}</span>
      </div>
    </header>

    {{if .Alert}}
    <div class="card" style="margin-top:16px; padding:14px 16px;">
      <span class="badge {{.AlertKind}}">{{if eq .AlertKind "success"}}Success{{else}}Error{{end}}</span>
      <p style="margin-top:8px;">{{.Alert}}</p>
    </div>
    {{end}}

    {{if .Result}}
    <section class="card" style="margin-top:16px;">
      <h2>{{.Result.Action}} complete</h2>
      <p style="margin-top:6px;">{{.Result.Summary}}</p>
      <div class="outputs">
        {{range .Result.Items}}
        <div class="output">
          <div>
            <strong>{{.Name}}</strong>
            <p>{{.Size}} • {{.Pages}} page(s)</p>
          </div>
          <a class="btn primary" href="{{.DownloadURL}}">Download</a>
        </div>
        {{end}}
      </div>
    </section>
    {{end}}

    <section class="shell">
      <article class="card">
        <span class="label">Web MVP Platform</span>
        <h1>All pdforge workflows in one shadcn-style control panel.</h1>
        <p>Merge, split, remove pages, and compress PDFs directly from the browser. Every flag exposed in the CLI is mapped into a form control below, and files are handled locally on this machine.</p>
        <div class="cta">
          <a class="btn primary" href="#workflows">Open workflows</a>
          <a class="btn ghost" href="/healthz">Health check</a>
        </div>

        <div class="tabs" id="workflows">
          <div class="tab-list">
            <button type="button" class="active" data-tab="merge">Merge</button>
            <button type="button" data-tab="split">Split</button>
            <button type="button" data-tab="remove">Remove Pages</button>
            <button type="button" data-tab="compress">Compress</button>
          </div>

          <div class="tab-panel active" data-panel="merge">
            <form class="form-grid" action="/api/merge" method="post" enctype="multipart/form-data">
							<input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
              <div class="field-wide">
                <label>PDF files</label>
                <input type="file" name="files" accept=".pdf" multiple required />
                <div class="hint">Matches <strong>merge &lt;file1&gt; &lt;file2&gt; ...</strong></div>
              </div>
              <div class="field-grid">
                <div class="field">
                  <label>Output file (-o / --output)</label>
                  <input type="text" name="output" placeholder="merged.pdf" />
                </div>
                <div class="field">
                  <label>Directory (-d / --dir)</label>
                  <input type="text" name="dir" placeholder="optional output directory" />
                </div>
              </div>
              <div class="submit-row">
                <button class="btn primary" type="submit">Run merge</button>
                <span class="hint">Output defaults to a timestamped PDF when omitted.</span>
              </div>
            </form>
          </div>

          <div class="tab-panel" data-panel="split">
            <form class="form-grid" action="/api/split" method="post" enctype="multipart/form-data">
							<input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
              <div class="field-wide">
                <label>PDF file</label>
                <input type="file" name="file" accept=".pdf" required />
              </div>
              <div class="field-grid">
                <div class="field">
                  <label>Positional selector</label>
                  <input type="text" name="selector" placeholder="8 or 6,8-10,11" />
                </div>
                <div class="field">
                  <label>--page flag value</label>
                  <input type="text" name="page" placeholder="same as selector, but via flag" />
                </div>
                <div class="field">
                  <label>Output (-o / --output)</label>
                  <input type="text" name="output" placeholder="section or section.pdf" />
                </div>
                <div class="field">
                  <label>Directory (-d / --dir)</label>
                  <input type="text" name="dir" placeholder="optional output directory" />
                </div>
              </div>
              <div class="checks">
                <label class="check"><input type="checkbox" name="extract" /> Extract mode (-e)</label>
                <label class="check"><input type="checkbox" name="odd" /> Odd pages (--odd)</label>
                <label class="check"><input type="checkbox" name="even" /> Even pages (--even)</label>
                <label class="check"><input type="checkbox" name="verbose" /> Verbose output (--verbose)</label>
              </div>
              <div class="submit-row">
                <button class="btn primary" type="submit">Run split</button>
                <span class="hint">If output is omitted, split uses the default naming pattern for the selected mode.</span>
              </div>
            </form>
          </div>

          <div class="tab-panel" data-panel="remove">
            <form class="form-grid" action="/api/remove" method="post" enctype="multipart/form-data">
							<input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
              <div class="field-wide">
                <label>PDF file</label>
                <input type="file" name="file" accept=".pdf" required />
              </div>
              <div class="field-grid">
                <div class="field">
                  <label>Page selector (--page or positional)</label>
                  <input type="text" name="page" placeholder="1,6-11,17" />
                </div>
                <div class="field">
                  <label>Output (-o / --output)</label>
                  <input type="text" name="output" placeholder="remove.pdf" />
                </div>
                <div class="field">
                  <label>Directory (-d / --dir)</label>
                  <input type="text" name="dir" placeholder="optional output directory" />
                </div>
              </div>
              <div class="submit-row">
                <button class="btn primary" type="submit">Run page removal</button>
                <span class="hint">The web form uses the same page syntax as the CLI.</span>
              </div>
            </form>
          </div>

          <div class="tab-panel" data-panel="compress">
            <form class="form-grid" action="/api/compress" method="post" enctype="multipart/form-data">
							<input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
              <div class="field-wide">
                <label>PDF file</label>
                <input type="file" name="file" accept=".pdf" required />
              </div>
              <div class="field-grid">
                <div class="field">
                  <label>Output (-o / --output)</label>
                  <input type="text" name="output" placeholder="archive_compressed.pdf" />
                </div>
                <div class="field">
                  <label>Directory (-d / --dir)</label>
                  <input type="text" name="dir" placeholder="optional output directory" />
                </div>
                <div class="field">
                  <label>Image mode (--image-mode)</label>
                  <select name="image_mode">
                    <option value="off">off</option>
                    <option value="readable">readable</option>
                    <option value="balanced">balanced</option>
                    <option value="aggressive">aggressive</option>
                    <option value="experimental">experimental</option>
                  </select>
                </div>
                <div class="field">
                  <label>Image max dimension (--image-max-dimension)</label>
                  <input type="number" name="image_max_dimension" min="0" placeholder="0 = auto" />
                </div>
                <div class="field">
                  <label>JPEG quality (--image-jpeg-quality)</label>
                  <input type="number" name="image_jpeg_quality" min="0" max="100" placeholder="0 = auto" />
                </div>
              </div>
              <div class="submit-row">
                <button class="btn primary" type="submit">Run compression</button>
                <span class="hint">Image mode maps directly to the CLI flags. Experiment with readable, balanced, and aggressive modes.</span>
              </div>
            </form>
          </div>
        </div>
      </article>

      <aside class="side">
        <section class="panel mini">
          <h2>Supported flags</h2>
          <div class="flag-list">
            <div class="flag-item"><span>merge</span><span>-o, -d</span></div>
            <div class="flag-item"><span>split</span><span>-e, --odd, --even, -v, -p, -o, -d</span></div>
            <div class="flag-item"><span>remove pages</span><span>-p, -o, -d</span></div>
            <div class="flag-item"><span>compress</span><span>-o, --image-mode, --image-max-dimension, --image-jpeg-quality</span></div>
          </div>
        </section>

        <section class="panel mini">
          <h2>Local runtime</h2>
          <div class="mini-card">
            <strong>Browser workflow</strong>
            <p style="margin-top:6px;">Uploads stay on this machine, files are written to temp output directories, and results are downloadable from a local job store.</p>
          </div>
          <div class="mini-card">
            <strong>Health</strong>
            <p style="margin-top:6px;">Use <a href="/healthz">/healthz</a> to verify the server is running.</p>
          </div>
        </section>
      </aside>
    </section>

    <div class="footer">{{.Now}}</div>
  </div>

  <script>
    const buttons = document.querySelectorAll('[data-tab]');
    const panels = document.querySelectorAll('[data-panel]');
    buttons.forEach((button) => {
      button.addEventListener('click', () => {
        buttons.forEach((item) => item.classList.remove('active'));
        panels.forEach((panel) => panel.classList.remove('active'));
        button.classList.add('active');
        document.querySelector('[data-panel="' + button.dataset.tab + '"]').classList.add('active');
      });
    });
  </script>
</body>
</html>`
