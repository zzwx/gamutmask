package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"log"
	"strings"
	"syscall"

	"os"
	"os/signal"

	"github.com/fsnotify/fsnotify"

	"github.com/zzwx/gamutmask/lib"
)

const (
	inputDefault  = "./_input"
	outputDefault = "./_output"
)

// Usage prints how to use the command-line utility
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func isInputFileForProcessing(folderName, fileName string) bool {
	switch filepath.Ext(fileName) {
	case ".jpg", ".jpeg", ".png":
		return true
	}
	return false
}

func isOutputFileSanitizable(outputFolderName, outputFileName string) bool {
	return true
}

func beforeDelete(folderName, fileName string) bool {
	if fileName == ".gitignore" {
		return false
	}
	fmt.Printf("Deleting: %v\n", folderName+"/"+fileName)
	return true
}

// main() calls ProcessChangedFilesOnly periodically
func main() {
	flag.Usage = Usage
	var help bool
	flag.BoolVar(&help, "help", false, "Print this help")

	//var fresh bool
	//flag.BoolVar(&fresh, "fresh", false, "Start fresh by deleting all images from output")
	//
	var monitor bool
	flag.BoolVar(&monitor, "monitor", true, "Monitor input folder for new and updated files")

	var once bool
	flag.BoolVar(&once, "once", false, "Shortuct to monitor=false")

	var recursive bool
	flag.BoolVar(&recursive, "recursive", false, "Walk all subfolders of the input folder too recursively")

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

	//if fresh {
	//	lib.SanitizeOutputFolder(output, isOutputFileSanitizable, &lib.FileInfoList{})
	//}

	var settings = RunGamutSettings{
		Width:    width,
		Height:   height,
		PaddingX: paddingX,
		PaddingY: paddingY,
	}

	if monitor {
		fmt.Println("Monitoring:", input, "for new and updated images...")

		watcher := newWatcher(input, recursive)

		timer := time.NewTimer(time.Second * 1)
		fallbackTimer := time.NewTimer(time.Second * 10)
		restart := make(chan int)
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Name /*relativePath*/ != "" {
						if filepath.Base(event.Name) == "_list.json" {
							// Skip _list.json changes to avoid non-stopping changes
						} else {
							if recursive {
								restart <- 0
							} else {
								resetTimer(timer, time.Second*2)
							}
						}
					}
				case <-watcher.Errors:
					// Skip the errors
				case <-timer.C:
					executeProcess(recursive, input, output, settings)
				case <-fallbackTimer.C:
					resetTimer(timer, time.Nanosecond)
					resetTimer(fallbackTimer, time.Second*10)
				}
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			for {
				select {
				case <-c:
					// TODO Wait for operations to finish gracefully
					fmt.Println("\nStopped Monitoring:", input)
					os.Exit(0) // We consider it a valid termination
				case <-restart:
					watcher.Close()
					watcher = newWatcher(input, recursive)
					resetTimer(timer, time.Second*2)
				}
			}
		}()
		<-make(chan int) // Blocking main() forever

	} else {
		executeProcess(recursive, input, output, settings)
	}

}

func newWatcher(input string, isRecursive bool) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	addToWatcher(watcher, input, isRecursive)
	return watcher
}

func addToWatcher(watcher *fsnotify.Watcher, input string, isRecursive bool) {
	err := watcher.Add(input)
	if err != nil {
		log.Fatal("Watcher error: ", err)
	}
	if isRecursive {
		files, err := ioutil.ReadDir(input)
		if err != nil {
			log.Fatal("Adding folder tot watcher error: ", err)
		}
		for _, f := range files {
			if f.IsDir() {
				addToWatcher(watcher, input+"/"+f.Name(), isRecursive)
			}
		}
	}
}

var outputFileName = func(inputFileName string) string {
	return inputFileName + ".png" // Simply appending .png at the end
}

func executeProcess(recursive bool, input string, output string, settings RunGamutSettings) {
	if recursive {
		lib.ProcessChangedFilesOnlyRecursively(input,
			output,
			outputFileName,
			isInputFileForProcessing,
			func(inputFolderName string) string {
				return inputFolderName + "/_list.json"
			},
			RunGamutFuncGen(&settings),
			beforeDelete)
	} else {
		lib.ProcessChangedFilesOnly(
			input,
			output,
			outputFileName,
			isInputFileForProcessing,
			input+"/_list.json",
			RunGamutFuncGen(&settings),
			beforeDelete)
	}
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	if timer == nil {
		timer = time.NewTimer(duration)
	} else {
		if !timer.Stop() {
			//<-timer.C
		}
		timer.Reset(duration)
	}
}
