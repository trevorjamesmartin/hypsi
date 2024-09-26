package main

import "fmt"

type plane struct {
	monitor string
	paper   string
}

func (p *plane) json() string {
	return fmt.Sprintf(`{ "monitor": "%s", "paper": "%s" }`, p.monitor, p.paper)
}
