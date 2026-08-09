package template

import (
	"errors"
	"os"

	"github.com/Yendric/geny/util"
)

type Template struct {
	Name string
}

type Registry struct {
	dir       string
	templates map[string]*Template
}

func NewRegistry(dir string) *Registry {
	return &Registry{
		dir:       dir,
		templates: make(map[string]*Template),
	}
}

// resolves template names, caching them in the registry
func (r *Registry) GetByName(name string) (*Template, error) {
	if template, ok := r.templates[name]; ok {
		return template, nil
	}

	if _, err := os.Stat(util.GeneratePath(r.dir, name+".html")); err != nil {
		return nil, errors.New("template not found: " + name)
	}

	template := &Template{Name: name}
	r.templates[name] = template
	return template, nil
}
