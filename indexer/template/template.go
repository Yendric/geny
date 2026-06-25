package template

import (
	"errors"
	"os"

	"github.com/Yendric/geny/util"
)

type Template struct {
	Name   string
	GoHtml string
}

var templates map[string]Template = make(map[string]Template)

func GetByName(templatesDir, name string) (*Template, error) {
	template, ok := templates[name]

	if ok {
		return &template, nil
	}

	html, err := os.ReadFile(util.GeneratePath(templatesDir, name+".html"))
	if err != nil {
		return nil, errors.New("template not found: " + name)
	}
	templates[name] = Template{
		Name:   name,
		GoHtml: string(html),
	}

	template = templates[name]

	return &template, nil
}
