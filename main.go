package main

import (
	"flag"
	"fmt"
	"time"

	"log"
	"strings"
	"syscall"

	"os"
	"os/signal"

	"github.com/zzwx/gamutmask/internal/cli"

	"github.com/fsnotify/fsnotify"
)

const (
	inputDefault  = "./_input"
	outputDefault = "./_output"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

// main() calls cli.ProcessChangedFilesOnly periodically
func main() {
	flag.Usage = Usage
	var help bool
	flag.BoolVar(&help, "help", false, "Print this help")

	var fresh bool
	flag.BoolVar(&fresh, "fresh", false, "Start fresh by deleting all images from output")

	var monitor bool
	flag.BoolVar(&monitor, "monitor", true, "Monitor input folder for new and updated files")

	var once bool
	flag.BoolVar(&once, "once", false, "Shortuct to monitor=false")

	var input string
	flag.StringVar(&input, "input", inputDefault, "Folder name where input files are located")
	var output string
	flag.StringVar(&output, "output", outputDefault, "Folder name where output files should be saved")

	var width int
	var height int
	var paddingX int
	var paddingY int
	flag.IntVar(&width, "width", 250, "Widgth of the resulting gamut image")
	flag.IntVar(&height, "height", 250, "Height of the resulting gamut image")
	flag.IntVar(&paddingX, "paddingX", 2, "Widgth of the resulting gamut image")
	flag.IntVar(&paddingY, "paddingY", 2, "Widgth of the resulting gamut image")

	flag.Parse()

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 && strings.ToLower(argsWithoutProg[0]) == "help" {
		help = true
	}

	// monitorFlag := flag.Lookup("monitor")
	// onceFlag := flag.Lookup("once")
	// if monitorFlag != nil && onceFlag != nil {
	// 	fmt.Fprintf(os.Stderr, "once and monitor flags can'be use together\n")
	// 	Usage()
	// 	os.Exit(2)
	// }

	if help {
		Usage()
		os.Exit(2)
	}

	if once {
		monitor = false
	}

	fmt.Println("Input folder:", input)
	fmt.Println("Output folder:", output)

	if _, err := os.Stat(input); os.IsNotExist(err) {
		fmt.Printf("Error: Input folder \"%s\" not found\n", input)
		os.Exit(1)
	}
	if _, err := os.Stat(output); os.IsNotExist(err) {
		fmt.Printf("Error: Output folder \"%s\" not found\n", output)
		os.Exit(1)
	}

	if fresh {
		cli.SanitizeOutputFolder(output, cli.FileInfoList{})
	}

	var settings = cli.RunGamutSettings{
		Width:    width,
		Height:   height,
		PaddingX: paddingX,
		PaddingY: paddingY,
	}

	if monitor {
		fmt.Println("Monitoring:", input, "for new and updated images...")

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		err = watcher.Add(input)
		if err != nil {
			fmt.Println("Watcher error: ")
			log.Fatal(err)
		}

		timer := time.NewTimer(time.Second * 1)
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Name != "" {
						if strings.HasSuffix(event.Name, "/_list.json") {
							// Skip _list.json changes to avoid non-stopping changes
						} else {
							resetTimer(timer)
						}
					}
				case <-watcher.Errors:
					// Skip the errors
				case <-timer.C:
					cli.ProcessChangedFilesOnly(input, output, cli.RunGamutFunc, &settings)
				}
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			for range c {
				// TODO Wait for operations to finish
				fmt.Println("\nStopped Monitoring:", input)
				os.Exit(0) // We consider it a valid termination
			}
		}()
		<-make(chan int) // Blocking main() forever

	} else {
		cli.ProcessChangedFilesOnly(input, output, cli.RunGamutFunc, &settings)
	}

}

func resetTimer(timer *time.Timer) {
	if timer == nil {
		timer = time.NewTimer(time.Second * 2)
	} else {
		if !timer.Stop() {
			//<-timer.C
		}
		timer.Reset(time.Second * 2)
	}
}
