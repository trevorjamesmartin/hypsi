package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"
)

const VERSION = "1.0.5"

const MESSAGE = `
hypsi %s

usage: hypsi [ <file> | <args> ]

   <file>		To set the desktop wallpaper of your focused monitor, simply provide the absolute path to your desired image file.

			^ download the image file, when given a web link

* alternatively by sending <args>, you can:

   -json		Show the current configuration in JSON format

   -rewind <N>		rewind to a previously set wallpaper, <N> (default: 1)
   
   -mode <mode>		set the hyprpaper <mode> (default: cover), of your focused monitor

   -webview		open with webkitgtk
`

var HYPSI_STATE AppState

func main() {
	var watcher Publisher

	var factory StateFactory

	HYPSI_STATE = factory.Create()

	HYPSI_STATE.Load()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	defer func() {
		HYPSI_STATE.Save() // save application state
		writeConfig(false) // write config file
		signal.Stop(c)     // stop the channel
		cancel()           // cancel the context

		if msg := HYPSI_STATE.GetMessage(); msg != "ok" {
			// show any unexpected messages
			fmt.Println(msg)
		}

		time.Sleep(300 * time.Millisecond) // delay freeing memory

		unloadWallpaper("all") // free memory
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	args := os.Args[1:]

	if len(args) > 0 {
		// nextCommand := args[0]
		switch args[0] {
		case "-webview":
			if len(args) > 1 {
				readInput(args[1:])
			}
			go api()
			gtkView(watcher)

		case "-json":
			fmt.Print(jsonText())

		case "-rewind":
			if len(args) > 1 {
				i, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Println("argument must be a number")
					return
				}
				rewind(i)
			} else {
				rewind(1)
			}

		case "-mode":
			if len(args) < 2 {
				setWallpaperMode(activeMonitor(), "cover")
			} else {
				setWallpaperMode(activeMonitor(), args[1])
			}

		case "-develop":
			CWD, _ := os.Getwd()
			files := []string{"webview.html.tmpl", "page.html.tmpl"}
			for _, filename := range files {
				localFile := filepath.Join(CWD, filename)
				if _, err := os.Stat(localFile); os.IsNotExist(err) {
					data, _ := WEBFOLDER.ReadFile(fmt.Sprintf("web/%s", filename))
					f, err := os.Create(localFile)
					if err != nil {
						log.Fatal(err)
					}
					defer f.Close()
					f.Write(data)
					fmt.Println(localFile)
				}
			}

		case "-watch":
			var watchfolder string

			if len(args) > 1 {
				watchfolder = args[1]
				if _, err := os.Stat(watchfolder); os.IsNotExist(err) {
					log.Fatalf("Cannot watch %s, the path does not exist", watchfolder)
				}
			} else {
				watchfolder, _ = os.Getwd()
			}
			os.Setenv("HYPSI_WEBVIEW", filepath.Join(watchfolder, "webview.html.tmpl"))
			os.Setenv("HYPSI_WEBPAGE", filepath.Join(watchfolder, "page.html.tmpl"))
			fmt.Printf("\n\n[ 👀 watching %s]\n", watchfolder)
			watcher = NewPathWatcher(watchfolder)
			go api()
			go watcher.observe()
			gtkView(watcher)

		default:
			readInput(args)
		}

	} else {
		fmt.Printf(MESSAGE, VERSION)
	}

}
