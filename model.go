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

type Webpage struct {
	template string

	data struct {
		Version  string
		Style    template.CSS
		Monitors []*Plane
		Ivalue   bool
		Rewind   int
		Script   template.JS
		Core     HyprCtlVersion
	}

	funcMap template.FuncMap
}

func (w *Webpage) Print(out io.Writer, i int) {
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

func (w Webpage) _Template() string {
	tmpl, staticError := WEBFOLDER.ReadFile("web/page.html.tmpl")
	if staticError != nil {
		log.Fatal(staticError)
	}
	return string(tmpl)
}
func (w Webpage) _Webview() string {
	tmpl, staticError := WEBFOLDER.ReadFile("web/webview.html.tmpl")
	if staticError != nil {
		log.Fatal(staticError)
	}
	return string(tmpl)
}
func webInit() Webpage {
	page := Webpage{}

	core, err := hyprCtlVersion()

	if err != nil {
		log.Fatal(err)
	}

	page.data.Core = core

	hist, _ := readHistory()

	funcMap := template.FuncMap{
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"plusOne": func(n int) int {
			return n + 1
		},
		"lessOne": func(n int) int {
			return n - 1
		},
		"inHistory": func(n int) bool {
			return n < len(hist)
		},
		"gtZero": func(n int) bool {
			return n > 0
		},
	}
	// default page values
	page.template = page._Template()
	page.funcMap = funcMap
	page.data.Version = VERSION
	return page
}
