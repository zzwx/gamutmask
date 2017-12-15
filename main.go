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

	"github.com/fsnotify/fsnotify"
	"github.com/kataras/iris"
	"github.com/zzwx/gamutmask/src/cli"
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

	fmt.Println("Input folder:", input, "\nOutput folder:", output, "\nMonitoring...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	err = watcher.Add(input)
	if err != nil {
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
				cli.ProcessChangedFilesOnly(input, output, cli.Run)
			}
		}
	}()

	//argsWithoutProg := os.Args[1:]
	//if len(argsWithoutProg) > 0 && strings.ToLower(argsWithoutProg[0]) == "server" {
	if isServer {
		fmt.Println("Starting a server...")
		//ir := iris.New()
		//ir.Get("/hi", func(ctx *iris.Context) {
		//	ctx.Writef("Hi %s", "iris")
		//})
		//ir.Use
		//ir.Get("/", index)
		//ir.Listen(":8080")
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

func index(ctx *iris.Context) {
	ctx.Render("index.html", struct{ Name string }{Name: "iris"})
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
