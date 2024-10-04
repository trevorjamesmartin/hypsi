package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
