package main

import (
	"fmt"
	"log"

	"github.com/guptarohit/asciigraph"
)

func plotascii(info *ExportInfo, remote string, index int) {

	height := 10
	width := 70

	if len(info.Input.YValues) > 0 {
		caption := fmt.Sprintf("Input Mbps: %s Connection %d", remote, index)
		log.Printf("%s input:", remote)
		input := asciigraph.Plot(info.Input.YValues, asciigraph.Caption(caption), asciigraph.Height(height), asciigraph.Width(width))
		fmt.Println(input)
	}

	if len(info.Output.YValues) > 0 {
		caption := fmt.Sprintf("Output Mbps: %s Connection %d", remote, index)
		log.Printf("%s output:", remote)
		output := asciigraph.Plot(info.Output.YValues, asciigraph.Caption(caption), asciigraph.Height(height), asciigraph.Width(width))
		fmt.Println(output)
	}
}
