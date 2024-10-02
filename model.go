package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	var target []map[string]string
	// basic
	err := json.Unmarshal([]byte(h.data), &target)
	if err != nil {
		log.Fatalf("Unable to marshal JSON due to %s", err)
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
