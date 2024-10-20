package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	webview "github.com/webview/webview_go"
)

func jsonText() string {
	active, err := listActive()
	if err != nil {
		log.Fatal(err)
	}
	bs, err := json.Marshal(active)
	if err != nil {
		log.Fatal(err)
	}

	return string(bs)
}

func configText() string {
	var text string
	sources := make(map[string]bool)
	activeplanes, err := listActive()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range activeplanes {
		if preloaded, _ := sources[p.Paper]; !preloaded {
			text += fmt.Sprintf("preload = %s\n", p.Paper)
			sources[p.Paper] = true
		}
	}

	for _, p := range activeplanes {
		text += fmt.Sprintf("wallpaper = %s,%s\n", p.Monitor, p.Paper)
	}
	text += "splash = false\n"
	return text
}

type thumbnail struct {
	Monitor string `json:"monitor"`
	Image   string `json:"image"`
}

type eventResp struct {
	Rewind   int         `json:"rewind,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Message  string      `json:"message,omitempty"`
	Monitors []thumbnail `json:"monitors,omitempty"`
}

func gtkView() {
	var port string
	var w webview.WebView
	port = os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	gtkViewHome := fmt.Sprintf("http://localhost:%s/webview", port)

	allowInspector := os.Getenv("DEBUG")
	if len(allowInspector) > 0 {
		w = webview.New(true)
	} else {

		w = webview.New(false)
	}

	defer w.Destroy()
	w.SetTitle("Hypsi")
	w.SetSize(0, 0, webview.HintNone)

	w.Bind("RollBack", func(n int) eventResp {

		if n < 0 {
			n = 0
			return eventResp{Rewind: 0, Message: "lower limit reached"}
		}

		good, limit := rewind(n)
		if !good {
			return eventResp{Message: "upper limit reached", Limit: limit}
		}

		HYPSI_STATE.Rewind = n

		monitors, errListing := listActive()
		if errListing != nil {
			HYPSI_STATE.Message = errListing.Error()
			log.Fatal(errListing)
		}

		var thumbs []thumbnail

		for _, mon := range monitors {
			img, _ := mon.Thumb64()
			thumbs = append(thumbs, thumbnail{Monitor: mon.Monitor, Image: img})
		}

		return eventResp{Rewind: n, Message: "ok", Monitors: thumbs, Limit: limit}
	})

	w.Navigate(gtkViewHome)
	w.Run()
}
