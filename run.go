package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/zzwx/gamutmask/lib"

	"runtime"

	"strconv"

	"gopkg.in/cheggaaa/pb.v1"
)

type RunGamutSettings struct {
	Width    int
	Height   int
	PaddingX int
	PaddingY int
}

var DefaultRunGamutSettings = RunGamutSettings{250, 250, 2, 2}

// RunGamutFuncGen generates a function that satisfies requirement of returned function signature while keeping reference
// of the settings and using it during actual call of the RunGamutFunc which requres settings
func RunGamutFuncGen(settings *RunGamutSettings) func(inputFileName string, outputFileName string) (exitCode int, err error) {
	if settings == nil {
		settings = &DefaultRunGamutSettings
	}
	return func(inputFileName string, outputFileName string) (exitCode int, err error) {
		return RunGamutFunc(inputFileName, outputFileName, settings)
	}
}

func ensureDir(dir string) error {
	err := os.Mkdir(dir, 0700)
	if err != nil {
		if !os.IsExist(err) { // We skip already existing dir error
			return fmt.Errorf("error creating directory: %w", err)
		}
	}
	return nil
}

// RunGamutFunc will execute GenerateGamutMask against inputFileName and generate outputFileName
// The file generated is going to be PNG so the outputFileName by convention
// is going to be <inputFileNameWithExtention>.png
func RunGamutFunc(inputFileName string, outputFileName string, settings *RunGamutSettings) (exitCode int, err error) {
	if settings == nil {
		settings = &DefaultRunGamutSettings
	}
	bar := newBar(4)

	exitCode = 0

	f, err := os.Open(inputFileName)
	if err != nil {
		return exitCode, nil // skip non-existing file
	}
	defer f.Close()

	start := time.Now()
	fmt.Printf("Generating: %v\n", inputFileName)

	bar.Start()
	bar.Increment()
	bar.Update()

	img, err := Decode(f)

	if err != nil {
		return 1, fmt.Errorf("image couldn't be read: %w", err)
	}

	bar.Increment()
	bar.Update()

	wheel := lib.GenerateGamutMask(img, settings.Width, settings.Height, settings.PaddingX, settings.PaddingY)
	bar.Increment()
	bar.Update()

	// Making sure directory exists
	if err := ensureDir(filepath.Dir(outputFileName)); err != nil {
		fmt.Errorf("error ensuring directory exists: %w", err)
	}

	out, err := os.Create(outputFileName)
	if err != nil {
		fmt.Errorf("error creating output file: %w", err)
	}
	bar.Increment()
	bar.Update()

	// Always using png to encode
	png.Encode(out, wheel)

	eraseLine()
	fmt.Printf("  %8.2fs (%vpx)\n", time.Since(start).Seconds(),
		comma(strconv.Itoa(img.Bounds().Dx()*img.Bounds().Dy())))

	return 0, nil
}

// Decode wraps logic of different images types into one function
func Decode(f *os.File) (image.Image, error) {
	var img image.Image
	var err error
	switch filepath.Ext(f.Name()) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("can't decode image: %w", err)
		}
	case ".png":
		img, err = png.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("can't decode image: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported image type: %v", filepath.Ext(f.Name()))
	}
	return img, nil
}

func eraseLine() {
	fmt.Printf("\r") // carriage return. Not always erasing the symbols
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Sending a special char sequence
		fmt.Printf("\033[2K")
	}
}

func newBar(count int) (bar *pb.ProgressBar) {
	bar = pb.New(count)
	bar.ShowPercent = false
	bar.ShowBar = true
	bar.ShowCounters = false
	bar.ShowTimeLeft = false
	bar.AlwaysUpdate = false
	bar.ManualUpdate = true
	bar.ForceWidth = false
	bar.Format("│\x00▒\x00▒\x00░\x00│")
	bar.ForceWidth = true
	bar.Width = count*2 + 2
	return
}

// comma inserts commas in a non-negative decimal integer string.
func comma(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	return comma(s[:n-3]) + "," + s[n-3:]
}
