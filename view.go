package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
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

func hyperText(w io.Writer, i int) {
	tmpl := LoadTemplate()

	monitors, errListing := listActive()

	if errListing != nil {
		log.Fatal(errListing)
	}

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	}

	data := struct {
		Version  string
		Style    template.CSS
		Monitors []*Plane
		Ivalue   bool
		Rewind   int
		Script   template.JS
	}{VERSION, LoadCSS(), monitors, i >= 0, i, LoadJS()}

	template.Must(template.New("webpage").Funcs(funcMap).Parse(tmpl)).Execute(w, data)
}
