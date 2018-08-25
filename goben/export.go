package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

func export(filename string, info *ExportInfo) error {

	out, errCreate := os.Create(filename)
	if errCreate != nil {
		return errCreate
	}
	defer out.Close()

	b, errMarshall := yaml.Marshal(*info)
	if errMarshall != nil {
		return errMarshall
	}

	_, errWrite := out.Write(b)

	return errWrite
}
