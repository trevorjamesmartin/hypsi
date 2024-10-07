package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const VERSION = "0.9"

const MESSAGE = `
hyprPaperPlanes %s

usage: hyprPaperPlanes [ <file> | <args> ]

   <file>	To set the desktop wallpaper of your focused monitor, simply provide the absolute path to your desired image file.

alternatively by sending <args>, you can:

   -listen	Start a local web server, listening on port 3000
   
   -json	Show the current configuration in JSON format

   -html	Render HTML without starting a web server

   -rewind	rewind config via logfile 

`

func main() {
	paperPath := fmt.Sprintf("%s/wallpaper", os.Getenv("HOME"))
	// ensure the "upload" folder exists
	if _, err := os.Stat(paperPath); os.IsNotExist(err) {
		// create with 0755 permissions (read, write, and execute for owner, read and execute for group and others)
		err := os.MkdirAll(paperPath, 0755)
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
			hyperText(os.Stdout, -1)
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
			writeConfig()

		case "-free":
			// free memory
			// (undocumented dev feature atm)
			unloadWallpaper("all")

		default:
			readFromCLI(argsWithoutProg)
		}

	} else {
		fmt.Printf(MESSAGE, VERSION)
	}

}
