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

// main() calls cli.ProcessChangedFilesOnly periodically
func main() {
	var isServer bool
	flag.BoolVar(&isServer, "server", false, "Use to start a server")

	var input string
	flag.StringVar(&input, "input", inputDefault, "Folder name where input files are located")
	var output string
	flag.StringVar(&output, "output", outputDefault, "Folder name where output files should be saved")

	flag.Parse()

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
				cli.ProcessChangedFilesOnly(input, output, cli.RunGamutFunc)
			}
		}
	}()

	//argsWithoutProg := os.Args[1:]
	//if len(argsWithoutProg) > 0 && strings.ToLower(argsWithoutProg[0]) == "server" {
	if isServer {
		fmt.Println("Starting a server...")
	} else {
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
