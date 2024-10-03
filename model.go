package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type plane struct {
	monitor string
	paper   string
}

func (p *plane) json() string {
	return fmt.Sprintf(`{ "monitor": "%s", "paper": "%s" }`, p.monitor, p.paper)
}

type history struct {
	dt   string
	data string
}

func (h *history) unfold() []plane {
	if h.data[0] == '{' {
		// single monitor
		h.data = fmt.Sprintf("[%s]", h.data)
	}
	var target []map[string]string
	// basic
	grief := json.Unmarshal([]byte(h.data), &target)
	if grief != nil {
		log.Fatalf("Unable to marshal JSON due to %s", grief)
	}

	result := []plane{}

	for _, data := range target {
		paper := data["paper"]
		monitor := data["monitor"]
		if len(monitor) > 0 && len(paper) > 0 {
			p := plane{monitor: monitor, paper: paper}
			result = append(result, p)
		}
	}

	return result
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

func readHistory() ([]history, error) {
	var past []history

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
				past = append(past, history{dt: line[:idx], data: line[idx:]})
			}
		}

	}
	return past, nil
}
