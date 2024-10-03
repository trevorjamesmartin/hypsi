package main

import (
	"fmt"
	"os"
	"strconv"
)

const VERSION = "0.5"

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
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {

		switch argsWithoutProg[0] {
		case "-listen":
			api()
		case "-json":
			fmt.Print(jsonText())
		case "-html":
			fmt.Print(hyperText())
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
		default:
			readFromCLI(argsWithoutProg)
		}

	} else {
		fmt.Printf(MESSAGE, VERSION)
	}

}
