package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
)

const VERSION = "0.9.9"

const MESSAGE = `
hypsi %s

usage: hypsi [ <file> | <args> ]

   <file>	To set the desktop wallpaper of your focused monitor, simply provide the absolute path to your desired image file.

alternatively by sending <args>, you can:

   -listen	Start a local web server, listening on port 3000
   
   -json	Show the current configuration in JSON format

   -html	Render HTML without starting a web server

   -rewind	rewind config via logfile

   -webview	open with webkitgtk
`

type APPLICATION_STATE struct {
	Rewind  int    `json:"rewind"`
	Message string `json:"message,omitempty"`
}

var HYPSI_STATE APPLICATION_STATE

func main() {
	var watcher Publisher
	HYPSI_FILE := fmt.Sprintf("%s/.hypsi", os.Getenv("HOME"))
	HYPSI_STATE.Message = "ok"

	data, err := os.ReadFile(HYPSI_FILE)
	if err == nil {
		json.Unmarshal(data, &HYPSI_STATE)
	} else {
		HYPSI_STATE.Message = err.Error()
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	defer func() {
		f, err := os.Create(HYPSI_FILE)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		app_state, _ := json.Marshal(HYPSI_STATE)
		fmt.Fprintf(f, string(app_state)) // save hypsi state
		writeConfig(false)                // save hyprpaper state
		signal.Stop(c)                    // stop the channel
		cancel()                          // cancel the context
		if HYPSI_STATE.Message != "ok" {
			// show any unexpected messages
			fmt.Println(HYPSI_STATE.Message)
		}
		unloadWallpaper("all")		  // free memory
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	UPLOADS := fmt.Sprintf("%s/wallpaper", os.Getenv("HOME"))
	// ensure the "upload" folder exists
	if _, err := os.Stat(UPLOADS); os.IsNotExist(err) {
		// create with 0755 permissions (read, write, and execute for owner, read and execute for group and others)
		err := os.MkdirAll(UPLOADS, 0755)
		if err != nil {
			log.Fatal(err) // Handle the error appropriately
		}
	}
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {

		switch argsWithoutProg[0] {
		case "-listen":
			api()
		case "-json":
			fmt.Print(jsonText())
		case "-html":
			page := webInit()
			page.Print(os.Stdout, -1)
		case "-rewind":
			if len(argsWithoutProg) > 1 {
				i, err := strconv.Atoi(argsWithoutProg[1])
				if err != nil {
					fmt.Println("argument must be a number")
					return
				}
				rewind(i)
			} else {
				rewind(1)
			}
		case "-write":
			// log changes & write hyprpaper.config
			// (undocumented dev feature atm)
			writeConfig(true)

		case "-free":
			// free memory
			// (undocumented dev feature atm)
			unloadWallpaper("all")

		case "-webview":
			go api()
			gtkView(watcher)

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
					fmt.Fprintf(f, string(data))
					fmt.Println(localFile)
				}
			}

		case "-watch":
			var watchfolder string

			if len(argsWithoutProg) > 1 {
				watchfolder = argsWithoutProg[1]
				if _, err := os.Stat(watchfolder); os.IsNotExist(err) {
					log.Fatalf("Cannot watch %s, the path does not exist", watchfolder)
				}
			} else {
				fmt.Println("... no folder specified, watch working directory")
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
			readFromCLI(argsWithoutProg)
		}

	} else {
		fmt.Printf(MESSAGE, VERSION)
	}

}
