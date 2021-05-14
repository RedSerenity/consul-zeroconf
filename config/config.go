package config

import (
	"bytes"
	"io"
	"os"
	"text/template"
)

const ANON_TOKEN = "00000000-0000-0000-0000-000000000002"

func GetTemplate(name, tmplString string, vars interface{}) (string, error) {
	tmpl, err := template.New(name).Parse(tmplString)
	if err != nil {
		return "", err
	}

	compiledTemplate := &bytes.Buffer{}
	if err = tmpl.Execute(compiledTemplate, vars); err != nil {
		return "", err
	}

	return compiledTemplate.String(), nil
}

func SaveConfig(path, filename, contents string) error {
	file, err := os.Create(path + filename)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err = io.WriteString(file, contents); err != nil {
		return err
	}

	return file.Sync()
}
