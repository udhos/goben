package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// CSV fields
const (
	Dir  = 0 // Direction
	Time = 1 // Timestamp
	Rate = 2 // Rate
)

func exportCsv(filename string, info *ExportInfo) error {

	out, errCreate := os.Create(filename)
	if errCreate != nil {
		return errCreate
	}

	w := csv.NewWriter(out)

	entry := []string{"DIRECTION", "TIME", "RATE"}

	if errHeader := w.Write(entry); errHeader != nil {
		return errHeader
	}

	entry[Dir] = "input"
	for i, x := range info.Input.XValues {
		entry[Time] = x.String()
		entry[Rate] = fmt.Sprintf("%v", info.Input.YValues[i])
		if err := w.Write(entry); err != nil {
			return err
		}
	}

	entry[Dir] = "output"
	for i, x := range info.Output.XValues {
		entry[Time] = x.String()
		entry[Rate] = fmt.Sprintf("%v", info.Output.YValues[i])
		if err := w.Write(entry); err != nil {
			return err
		}
	}

	w.Flush()

	return out.Close()
}
