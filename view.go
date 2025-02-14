package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	//webview "github.com/webview/webview_go"
	// ^ waiting for version bump to webkit2gtk-4.1
	webview "github.com/trevorjamesmartin/webview_go" // testing solution
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
		if preloaded := sources[p.Paper]; !preloaded {
			text += fmt.Sprintf("preload = %s\n", p.Paper)
			sources[p.Paper] = true
		}
	}

	for _, p := range activeplanes {
		switch p.Mode {
		case "contain":
			text += fmt.Sprintf("wallpaper = %s:%s,%s\n", p.Mode, p.Paper, p.Monitor)
		default:
			text += fmt.Sprintf("wallpaper = %s,%s\n", p.Monitor, p.Paper)
		}
	}
	text += "splash = false\n"
	return text
}

type thumbnail struct {
	Monitor string `json:"monitor"`
	Image   string `json:"image"`
	Mode    string `json:"mode,omitempty"`
}

type eventResp struct {
	Rewind   int         `json:"rewind,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Message  string      `json:"message,omitempty"`
	Monitors []thumbnail `json:"monitors,omitempty"`
}

type WebviewSubcriber struct {
	view webview.WebView
	home string
}

func (ws *WebviewSubcriber) receive(path, event string) {
	switch event {
	case "WRITE":
		ws.view.Navigate(ws.home)
	default:
		return
	}
}

func gtkView(pub Publisher) {
	var port string
	var w webview.WebView
	port = os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	webviewHome := fmt.Sprintf("http://127.0.0.1:%s/webview", port)

	allowInspector := os.Getenv("DEBUG")
	if len(allowInspector) > 0 {
		w = webview.New(true)
	} else {

		w = webview.New(false)
	}

	defer w.Destroy()
	w.SetTitle("Hypsi")
	w.SetSize(640, 380, webview.Hint(webview.HintMin))

	if pub != nil {
		webviewSub := &WebviewSubcriber{view: w, home: webviewHome}
		var sub Subscriber = webviewSub
		pub.register(&sub)
		fmt.Print("\n[ 👀 webview ]\n")
	}

	w.Bind("SaveLocalJSON", func(localStorage json.RawMessage) {
		var id int

		sqlData := openDatabase()
		defer sqlData.Close()

		var stmt string
		var data []byte
		row := sqlData.QueryRow(`select * from localstorage order by id desc limit 1`)

		if row.Scan(&id, &data) != nil {
			stmt = fmt.Sprintf(`insert into localstorage(id, data) values(%d, '%s');`, 0, localStorage)
		} else {
			stmt = fmt.Sprintf(`update localstorage set data='%s' where id=%d;`, localStorage, 0)
		}

		_, err := sqlData.Exec(stmt)
		if err != nil {
			fmt.Printf("%q: %s\n", err, stmt)
		}
	})

	w.Bind("GetLocalJSON", func() json.RawMessage {
		var data interface{}
		var id int

		sqlData := openDatabase()
		defer sqlData.Close()

		row := sqlData.QueryRow(`select * from localstorage order by id desc limit 1`)
		if row.Scan(&id, &data) != nil {
			fmt.Println("nothing stored")
			return nil
		}

		x, err := json.Marshal(data)

		if err != nil {
			fmt.Println("error reading JSON database record")
			fmt.Println(err)
		}

		return x
	})

	w.Bind("SetWallpaperMode", setWallpaperMode)

	w.Bind("MonitorFileName", monitorFilename) // returns filename (string)

	w.Bind("GetModeSetting", getModeSetting) // returns mode (string) cover|contain

	w.Bind("RollBack", func(n int) eventResp {

		if n < 0 {
			n = 0
			return eventResp{Rewind: 0, Message: "lower limit reached"}
		}

		good, limit := rewind(n)
		if !good {
			return eventResp{Message: "upper limit reached", Limit: limit}
		}

		HYPSI_STATE.SetRewind(n)

		monitors, errListing := listActive()
		if errListing != nil {
			HYPSI_STATE.SetMessage(errListing.Error())
			log.Fatal(errListing)
		}

		var thumbs []thumbnail

		for _, mon := range monitors {
			img, _ := mon.Thumb64()
			mode := getModeSetting(mon.Monitor, filepath.Base(mon.Paper))
			thumbs = append(thumbs, thumbnail{Monitor: mon.Monitor, Image: img, Mode: mode})
		}

		return eventResp{Rewind: n, Message: "ok", Monitors: thumbs, Limit: limit}
	})

	w.Navigate(webviewHome)
	w.Run()
}
