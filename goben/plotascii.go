package goben

import (
	"fmt"
	"log"
	"os"

	"github.com/guptarohit/asciigraph"
)

func plotasciiToFile(filename string, info *ExportInfo, remote string, index int) {

	height := 10
	width := 70

	var buf string

	if len(info.Input.YValues) > 0 {
		caption := fmt.Sprintf("Input Mbps: %s Connection %d", remote, index)
		log.Printf("%s input:", remote)
		input := asciigraph.Plot(info.Input.YValues, asciigraph.Caption(caption), asciigraph.Height(height), asciigraph.Width(width))
		fmt.Println(input)
		buf += input + "\n"
	}

	if len(info.Output.YValues) > 0 {
		caption := fmt.Sprintf("Output Mbps: %s Connection %d", remote, index)
		log.Printf("%s output:", remote)
		output := asciigraph.Plot(info.Output.YValues, asciigraph.Caption(caption), asciigraph.Height(height), asciigraph.Width(width))
		fmt.Println(output)
		buf += output + "\n"
	}

	if filename != "" && buf != "" {
		if err := os.WriteFile(filename, []byte(buf), 0644); err != nil {
			log.Printf("plotascii: write file %s: %v", filename, err)
		}
	}
}
