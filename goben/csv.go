package main

import (
	"fmt"
	"os"
)

func exportCsv(filename string, info *ExportInfo) error {

	out, errCreate := os.Create(filename)
	if errCreate != nil {
		return errCreate
	}
	defer out.Close()

	format := "%s,\"%v\",%v\n"

	if _, errHeader := out.Write([]byte(fmt.Sprintf(format, "DIRECTION", "TIME", "RATE"))); errHeader != nil {
		return errHeader
	}

	for i, x := range info.Input.XValues {
		y := info.Input.XValues[i]
		s := fmt.Sprintf(format, "input", timeFromFloat(x), y)
		if _, err := out.Write([]byte(s)); err != nil {
			return err
		}
	}

	for i, x := range info.Output.XValues {
		y := info.Output.XValues[i]
		s := fmt.Sprintf(format, "output", timeFromFloat(x), y)
		if _, err := out.Write([]byte(s)); err != nil {
			return err
		}
	}

	return nil
}
