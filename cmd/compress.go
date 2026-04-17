/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/image/draw"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf>",
	Short: "Compress a PDF to reduce file size",
	Long: `The compress command optimizes a PDF while preserving document usability.
Use it to make PDF files easier to store and share.`,
	Example: `  pdforge compress large-report.pdf
	pdforge compress archive.pdf -o archive_compressed.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runCompress,
}

var compressOutput string
var compressImageMode string
var compressImageMaxDimension int
var compressImageJPEGQuality int

const (
	compressImageModeOff          = "off"
	compressImageModeBalanced     = "balanced"
	compressImageModeAggressive   = "aggressive"
	compressImageModeExperimental = "experimental" // Backward-compatible alias for aggressive.
)

type imageStats struct {
	count      int
	totalBytes int64
	maxWidth   int
	maxHeight  int
}

type imageModeProfile struct {
	maxDimension int
	jpegQuality  int
	warning      string
}

func defaultCompressedOutput(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".pdf"
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	return filepath.Join(filepath.Dir(inputPath), base+"_compressed"+ext)
}

func runCompress(cmd *cobra.Command, args []string) error {
	input := args[0]

	if strings.ToLower(filepath.Ext(input)) != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", filepath.Base(input))
	}

	bar := progressbar.Default(1, "Validating file")
	if err := api.ValidateFile(input, nil); err != nil {
		return fmt.Errorf("invalid PDF '%s': \n%v", filepath.Base(input), err)
	}
	_ = bar.Add(1)

	stats, imageHeavy, err := detectImageHeavyPDF(input)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: unable to analyze embedded images: %v\n", err)
	}

	if compressImageMode == compressImageModeOff && imageHeavy {
		fmt.Fprintf(cmd.OutOrStdout(), "Detected image-heavy PDF (%d images, %.2f MB embedded image streams).\n", stats.count, float64(stats.totalBytes)/(1024*1024))
		fmt.Fprintln(cmd.OutOrStdout(), "Suggestion: try --image-mode balanced for safer image-aware compression.")

		enable, promptErr := promptEnableExperimentalImageMode(cmd)
		if promptErr != nil {
			return promptErr
		}
		if enable {
			compressImageMode = compressImageModeBalanced
			fmt.Fprintln(cmd.OutOrStdout(), "Using --image-mode balanced for this run.")
		}
	}

	workingInput := input
	var cleanup func()
	if compressImageMode != compressImageModeOff {
		profile := resolveImageModeProfile(compressImageMode)
		maxDimension := profile.maxDimension
		jpegQuality := profile.jpegQuality

		if cmd.Flags().Changed("image-max-dimension") {
			maxDimension = compressImageMaxDimension
		}
		if cmd.Flags().Changed("image-jpeg-quality") {
			jpegQuality = compressImageJPEGQuality
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Image mode '%s' is enabled. Re-encoding/downsampling embedded images before optimization.\n", compressImageMode)
		processedInput, replacedCount, cleanFn, processErr := preprocessPDFImages(input, maxDimension, jpegQuality)
		if processErr != nil {
			return processErr
		}
		cleanup = cleanFn
		workingInput = processedInput
		fmt.Fprintf(cmd.OutOrStdout(), "Experimental image preprocessing replaced %d image(s).\n", replacedCount)
	}
	if cleanup != nil {
		defer cleanup()
	}

	output := compressOutput
	if output == "" {
		output = defaultCompressedOutput(input)
	}

	if strings.ToLower(filepath.Ext(output)) != ".pdf" {
		return fmt.Errorf("the file '%s' is invalid, must be '.pdf'", output)
	}

	outDir := filepath.Dir(output)
	if outDir == "" {
		outDir = "."
	}

	dirInfo, err := os.Stat(outDir)
	if err != nil {
		return fmt.Errorf("directory '%s' does not exist", outDir)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", outDir)
	}

	output = resolveOutputPath(output)

	bar = progressbar.Default(-1, "Compressing")
	if err := api.OptimizeFile(workingInput, output, nil); err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Println("===== Compression Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		return err
	}
	report.PrintReport()

	return nil
}

func detectImageHeavyPDF(input string) (imageStats, bool, error) {
	var stats imageStats

	f, err := os.Open(input)
	if err != nil {
		return stats, false, err
	}
	defer f.Close()

	imagesPerPage, err := api.Images(f, nil, nil)
	if err != nil {
		return stats, false, err
	}

	for _, pageImages := range imagesPerPage {
		for _, img := range pageImages {
			stats.count++
			stats.totalBytes += img.Size
			if img.Width > stats.maxWidth {
				stats.maxWidth = img.Width
			}
			if img.Height > stats.maxHeight {
				stats.maxHeight = img.Height
			}
		}
	}

	const imageCountThreshold = 8
	const imageBytesThreshold = int64(8 * 1024 * 1024)

	imageHeavy := stats.count >= imageCountThreshold || stats.totalBytes >= imageBytesThreshold
	return stats, imageHeavy, nil
}

func promptEnableExperimentalImageMode(cmd *cobra.Command) (bool, error) {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false, nil
	}

	if (stdinInfo.Mode() & os.ModeCharDevice) == 0 {
		return false, nil
	}

	fmt.Fprint(cmd.OutOrStdout(), "Enable balanced image mode now? [y/N]: ")
	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func resolveImageModeProfile(mode string) imageModeProfile {
	switch mode {
	case compressImageModeAggressive:
		return imageModeProfile{
			maxDimension: 1400,
			jpegQuality:  70,
			warning:      "Warning: --image-mode aggressive may noticeably reduce visual quality and can increase file size for some PDFs.",
		}
	case compressImageModeBalanced:
		fallthrough
	default:
		return imageModeProfile{
			maxDimension: 2200,
			jpegQuality:  88,
			warning:      "Notice: --image-mode balanced is experimental and may slightly affect visual quality.",
		}
	}
}

func preprocessPDFImages(input string, maxDimension, jpegQuality int) (string, int, func(), error) {
	tmpDir, err := os.MkdirTemp("", "pdforge-image-mode-")
	if err != nil {
		return "", 0, nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	f, err := os.Open(input)
	if err != nil {
		cleanup()
		return "", 0, nil, err
	}

	imagesPerPage, err := api.ExtractImagesRaw(f, nil, nil)
	_ = f.Close()
	if err != nil {
		cleanup()
		return "", 0, nil, err
	}

	workingPDF := input
	replaced := 0
	step := 0

	for _, pageImages := range imagesPerPage {
		for _, img := range pageImages {
			if img.ObjNr <= 0 {
				continue
			}

			imgData, err := io.ReadAll(img.Reader)
			if err != nil {
				continue
			}

			decoded, format, err := image.Decode(bytes.NewReader(imgData))
			if err != nil {
				continue
			}

			newImg := resizeImage(decoded, maxDimension)

			ext := ".jpg"
			writePNG := hasAlpha(newImg) || strings.EqualFold(format, "png")
			if writePNG {
				ext = ".png"
			}

			replacementPath := filepath.Join(tmpDir, fmt.Sprintf("img-replacement-%d%s", step, ext))
			if err := writeReplacementImage(replacementPath, newImg, writePNG, jpegQuality); err != nil {
				continue
			}

			nextPDF := filepath.Join(tmpDir, fmt.Sprintf("step-%d.pdf", step))
			if err := api.UpdateImagesFile(workingPDF, replacementPath, nextPDF, img.ObjNr, 0, "", nil); err != nil {
				continue
			}

			workingPDF = nextPDF
			replaced++
			step++
		}
	}

	if replaced == 0 {
		cleanup()
		return input, 0, nil, nil
	}

	return workingPDF, replaced, cleanup, nil
}

func resizeImage(src image.Image, maxDimension int) image.Image {
	b := src.Bounds()
	width := b.Dx()
	height := b.Dy()

	if width <= 0 || height <= 0 || maxDimension <= 0 {
		return src
	}

	if width <= maxDimension && height <= maxDimension {
		return src
	}

	var newW, newH int
	if width >= height {
		newW = maxDimension
		newH = int(float64(height) * (float64(maxDimension) / float64(width)))
	} else {
		newH = maxDimension
		newW = int(float64(width) * (float64(maxDimension) / float64(height)))
	}

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

func hasAlpha(img image.Image) bool {
	switch img.(type) {
	case *image.Alpha, *image.Alpha16, *image.RGBA, *image.RGBA64, *image.NRGBA, *image.NRGBA64:
		return true
	default:
		return false
	}
}

func writeReplacementImage(path string, img image.Image, writePNG bool, jpegQuality int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if writePNG {
		return png.Encode(f, img)
	}

	if jpegQuality < 1 {
		jpegQuality = 1
	}
	if jpegQuality > 100 {
		jpegQuality = 100
	}

	return jpeg.Encode(f, img, &jpeg.Options{Quality: jpegQuality})
}

func init() {
	rootCmd.AddCommand(compressCmd)
	compressCmd.Flags().StringVarP(&compressOutput, "output", "o", "", "Output PDF file path (default: <input>_compressed.pdf)")
	compressCmd.Flags().StringVar(&compressImageMode, "image-mode", compressImageModeOff, "Image-aware compression mode: off|balanced|aggressive (experimental)")
	compressCmd.Flags().IntVar(&compressImageMaxDimension, "image-max-dimension", 0, "Advanced: override max image dimension in pixels for image mode")
	compressCmd.Flags().IntVar(&compressImageJPEGQuality, "image-jpeg-quality", 0, "Advanced: override JPEG quality for image mode (1-100)")

	if err := compressCmd.Flags().MarkHidden("image-max-dimension"); err != nil {
		panic(err)
	}
	if err := compressCmd.Flags().MarkHidden("image-jpeg-quality"); err != nil {
		panic(err)
	}

	compressCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		compressImageMode = strings.ToLower(strings.TrimSpace(compressImageMode))
		if compressImageMode == compressImageModeExperimental {
			compressImageMode = compressImageModeAggressive
		}

		if compressImageMode != compressImageModeOff && compressImageMode != compressImageModeBalanced && compressImageMode != compressImageModeAggressive {
			return fmt.Errorf("invalid --image-mode '%s', expected: off|balanced|aggressive", compressImageMode)
		}

		if cmd.Flags().Changed("image-max-dimension") && compressImageMaxDimension < 1 {
			return fmt.Errorf("--image-max-dimension must be >= 1")
		}

		if cmd.Flags().Changed("image-jpeg-quality") && (compressImageJPEGQuality < 1 || compressImageJPEGQuality > 100) {
			return fmt.Errorf("--image-jpeg-quality must be between 1 and 100")
		}

		if compressImageMode != compressImageModeOff {
			profile := resolveImageModeProfile(compressImageMode)
			fmt.Fprintln(cmd.OutOrStdout(), profile.warning)
		}

		return nil
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
