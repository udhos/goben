package main

import (
	"fmt"
	"log"

	"github.com/guptarohit/asciigraph"
)

func plotascii(info *ExportInfo) {

	height := 10
	width := 70

	if len(info.Input.YValues) > 0 {
		log.Println("input:")
		input := asciigraph.Plot(info.Input.YValues, asciigraph.Caption("Input Mbps"), asciigraph.Height(height), asciigraph.Width(width))
		fmt.Println(input)
	}

	if len(info.Output.YValues) > 0 {
		log.Println("output:")
		output := asciigraph.Plot(info.Output.YValues, asciigraph.Caption("Output Mbps"), asciigraph.Height(height), asciigraph.Width(width))
		fmt.Println(output)
	}
}
