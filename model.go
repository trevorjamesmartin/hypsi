package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Plane struct {
	Monitor string
	Paper   string
}

func (p Plane) MarshallJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Plane) UnMarshallJSON(data []byte) error {
	var pln Plane
	if err := json.Unmarshal(data, &pln); err != nil {
		return err
	}
	*p = pln
	return nil
}

func (p *Plane) ToBase64() (string, error) {
	bts, err := os.ReadFile(p.Paper)

	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bts), base64.StdEncoding.EncodeToString(bts))

	return result, nil
}

func (p *Plane) Thumb64() (string, error) {
	fileName := filepath.Base(p.Paper)
	thumbFile := fmt.Sprintf("thumb__%s", fileName)
	thumbPath := filepath.Join(os.Getenv("HOME"), "wallpaper", thumbFile)

	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		makeThumbNail(p.Paper, thumbPath)
	}

	bts, err := os.ReadFile(thumbPath)

	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bts), base64.StdEncoding.EncodeToString(bts))

	return result, nil
}

type History struct {
	dt   string
	data string
}

func (h *History) unfold() []Plane {
	if h.data[0] == '{' {
		// single monitor
		h.data = fmt.Sprintf("[%s]", h.data)
	}
	var target []Plane

	grief := json.Unmarshal([]byte(h.data), &target)
	if grief != nil {
		log.Fatalf("Unable to marshal JSON due to %s", grief)
	}

	return target
}

func writeHistory() {
	// log the current wallpaper(s)
	historyfile := fmt.Sprintf("%s/wallpaper/hyprpaperplanes.log", os.Getenv("HOME"))
	file, grief := os.OpenFile(historyfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if grief != nil {
		log.Fatal(grief)
	}

	defer file.Close()
	defer fmt.Println("logged")

	log.SetOutput(file)

	log.Println(jsonText())
}

func readHistory() ([]History, error) {
	var past []History

	historyfile := fmt.Sprintf("%s/wallpaper/hyprpaperplanes.log", os.Getenv("HOME"))
	file, grief := os.Open(historyfile)

	if grief != nil {
		return past, grief
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		line := scanner.Text()
		if len(line) > 0 {
			idx := strings.IndexRune(line, '[')
			if idx == -1 {
				// catch for single monitor
				idx = strings.IndexRune(line, '{')
			}

			if idx >= 0 {
				past = append(past, History{dt: line[:idx], data: line[idx:]})
			}
		}

	}
	return past, nil
}

type Webview struct {
	template string

	data struct {
		Version  string
		Style    template.CSS
		Monitors []*Plane
		Ivalue   bool
		Rewind   int
		Script   template.JS
	}

	funcMap template.FuncMap
}

func (w *Webview) Print(out io.Writer, i int) {
	monitors, errListing := listActive()
	if errListing != nil {
		log.Fatal(errListing)
	}

	// these values change
	w.data.Rewind = i
	w.data.Ivalue = i >= 0
	w.data.Monitors = monitors

	// 'write => out' the resulting template.HTML
	template.Must(template.New("webpage").Funcs(w.funcMap).Parse(w.template)).Execute(out, w.data)
}

func webInit() Webview {
	page := Webview{}

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	}
	// load once, these values never change at runtime
	page.template = page._Template()
	page.funcMap = funcMap
	page.data.Version = VERSION
	page.data.Style = page._CSS()
	page.data.Script = page._JS()

	return page
}

func (w *Webview) _JS() template.JS {
	bts, err := os.ReadFile("./web/script.js")
	if err != nil {
		return ""
	}
	return template.JS(bts)
}

func (w *Webview) _CSS() template.CSS {
	bts, err := os.ReadFile("./web/style.css")
	if err != nil {
		return ""
	}
	return template.CSS(bts)

}

func (w Webview) _Template() string {
	bts, err := os.ReadFile("./web/page.html.tmpl")
	if err != nil {
		return ""
	}
	return string(bts)
}
