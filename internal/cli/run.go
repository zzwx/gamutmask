package cli

import (
	"fmt"
	"image/jpeg"
	"os"
	"time"

	"runtime"

	"strconv"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/zzwx/gamutmask/internal/lib"
)

// Run will execute GenerateGamutMask against inputFileName and generate outputFileName
func Run(outputFileName string, inputFileName string) (exitCode int) {
	bar := newBar(4)

	exitCode = 0

	f, err := os.Open(inputFileName)
	if err != nil {
		return // skip non-existing file
	}
	defer f.Close()

	start := time.Now()
	fmt.Printf("Generating: %v\n", inputFileName)

	bar.Start()
	bar.Increment()
	bar.Update()

	img, err := jpeg.Decode(f)
	if err != nil {
		panic(err)
	}
	bar.Increment()
	bar.Update()

	wheel := lib.GenerateGamutMask(img, 150, 150)
	bar.Increment()
	bar.Update()

	out, err := os.Create(outputFileName)
	if err != nil {
		panic(err)
	}
	bar.Increment()
	bar.Update()

	jpeg.Encode(out, wheel, &jpeg.Options{Quality: 100})

	eraseLine()
	fmt.Printf("  %8.2fs (%vpx)\n", time.Since(start).Seconds(),
		comma(strconv.Itoa(img.Bounds().Dx()*img.Bounds().Dy())))
	return
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
