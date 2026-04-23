/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
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
	xdraw "golang.org/x/image/draw"
)

// compressCmd represents the compress command
var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf>",
	Short: "Compress a PDF to reduce file size",
	Long: `The compress command optimizes a PDF while preserving document usability.
Use it to make PDF files easier to store and share.`,
	Example: `  pdforge compress large-report.pdf
	pdforge compress archive.pdf -o archive_compressed.pdf`,
	Args: argsWithHelp(cobra.ExactArgs(1)),
	RunE: runCompress,
}

var compressOutput string
var compressImageMode string
var compressImageMaxDimension int
var compressImageJPEGQuality int

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

const (
	compressImageModeOff          = "off"
	compressImageModeReadable     = "readable"
	compressImageModeBalanced     = "balanced"
	compressImageModeAggressive   = "aggressive"
	compressImageModeExperimental = "experimental"
)

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

		enable, promptErr := promptYesNo(cmd, "Enable balanced image mode now? [y/N]: ")
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

	if err := ensureOutputDirectory(cmd, outDir); err != nil {
		return err
	}

	output = resolveOutputPath(output)

	bar = progressbar.Default(-1, "Compressing")
	if err := api.OptimizeFile(workingInput, output, nil); err != nil {
		return err
	}
	_ = bar.Finish()

	fmt.Fprintln(cmd.OutOrStdout(), "===== Compression Completed =====")
	report, err := GetFileInfo(output)
	if err != nil {
		return err
	}
	report.PrintReport(cmd.OutOrStdout())

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
func resolveImageModeProfile(mode string) imageModeProfile {
	switch mode {
	case compressImageModeAggressive:
		return imageModeProfile{
			maxDimension: 1400,
			jpegQuality:  70,
			warning:      "Warning: --image-mode aggressive may noticeably reduce visual quality and can increase file size for some PDFs.",
		}
	case compressImageModeReadable:
		return imageModeProfile{
			maxDimension: 2600,
			jpegQuality:  93,
			warning:      "Notice: --image-mode readable prioritizes quality over size reduction.",
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
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, b, xdraw.Over, nil)
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
	compressCmd.Flags().StringVar(&compressImageMode, "image-mode", compressImageModeOff, "Image-aware compression mode: off|readable|balanced|aggressive")
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

		if compressImageMode != compressImageModeOff && compressImageMode != compressImageModeReadable && compressImageMode != compressImageModeBalanced && compressImageMode != compressImageModeAggressive {
			return fmt.Errorf("invalid --image-mode '%s', expected: off|readable|balanced|aggressive", compressImageMode)
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
