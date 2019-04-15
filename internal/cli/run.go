package cli

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"runtime"

	"strconv"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/zzwx/gamutmask/internal/lib"
)

type RunGamutSettings struct {
	Width    int
	Height   int
	PaddingX int
	PaddingY int
}

var DefaultRunGamutSettings = RunGamutSettings{250, 250, 2, 2}

// RunGamutFunc will execute GenerateGamutMask against inputFileName and generate outputFileName
// The file generated is going to be PNG so the outputFileName by convention
// is going to be <inputFileNameWithExtention>.png
func RunGamutFunc(outputFileName string, inputFileName string, settings *RunGamutSettings) (exitCode int, err error) {
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

	var img image.Image
	switch filepath.Ext(f.Name()) {
	case ".jpg":
		img, err = jpeg.Decode(f)
		if err != nil {
			panic(err)
		}
	case ".png":
		img, err = png.Decode(f)
		if err != nil {
			panic(err)
		}
	default:
		panic(filepath.Ext(f.Name()) + " Unsupported image type")
	}

	if img == nil {
		return 1, fmt.Errorf("Image couldn't be read: %s", f.Name())
	}

	bar.Increment()
	bar.Update()

	wheel := lib.GenerateGamutMask(img, settings.Width, settings.Height, settings.PaddingX, settings.PaddingY)
	bar.Increment()
	bar.Update()

	out, err := os.Create(outputFileName)
	if err != nil {
		panic(err)
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
