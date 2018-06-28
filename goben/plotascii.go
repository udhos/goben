package main

import (
	"fmt"
	"log"

	"github.com/guptarohit/asciigraph"
)

func plotascii(info *ExportInfo) {
	if len(info.Input.YValues) > 0 {
		log.Println("input:")
		input := asciigraph.Plot(info.Input.YValues, asciigraph.Height(10), asciigraph.Width(60))
		fmt.Println(input)
	}

	if len(info.Output.YValues) > 0 {
		log.Println("output:")
		output := asciigraph.Plot(info.Output.YValues, asciigraph.Height(10), asciigraph.Width(60))
		fmt.Println(output)
	}
}
