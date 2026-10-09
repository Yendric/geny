package islands

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

const RuntimeEntry = "virtual:geny/islands"

var extensions = []string{".tsx", ".jsx"}

type Directive string

const (
	Load    Directive = "load"
	Visible Directive = "visible"
)

func ParseDirective(s string) (Directive, error) {
	switch d := Directive(s); d {
	case Load, Visible:
		return d, nil
	}
	return "", fmt.Errorf("unknown island directive %q, expected %q or %q", s, Load, Visible)
}

type Prop struct {
	Key   string
	Value json.RawMessage
}

type Island struct {
	Name      string
	Props     []Prop
	Directive Directive
	Fallback  template.HTML
}

func New(name string) Island {
	return Island{Name: name, Directive: Load}
}

func (i Island) WithProp(key string, value json.RawMessage) (Island, error) {
	for _, p := range i.Props {
		if p.Key == key {
			return i, fmt.Errorf("duplicate prop %q", key)
		}
	}
	props := make([]Prop, len(i.Props), len(i.Props)+1)
	copy(props, i.Props)
	i.Props = append(props, Prop{Key: key, Value: value})
	return i, nil
}

func (i Island) PropsJSON() string {
	var b strings.Builder
	b.WriteByte('{')
	for n, p := range i.Props {
		if n > 0 {
			b.WriteByte(',')
		}
		key, _ := json.Marshal(p.Key)
		b.Write(key)
		b.WriteByte(':')
		b.Write(p.Value)
	}
	b.WriteByte('}')
	return b.String()
}

const closeTag = "</geny-island>"

func (i Island) openTag() string {
	return fmt.Sprintf(
		`<geny-island component="%s" client="%s" props="%s">`,
		html.EscapeString(i.Name),
		i.Directive,
		html.EscapeString(i.PropsJSON()),
	)
}

func (i Island) HTML() template.HTML {
	return template.HTML(i.openTag() + string(i.Fallback) + closeTag)
}

type Usage struct {
	Island Island
	Source string
	Line   int
}

type Registry struct {
	files      map[string]string
	components map[string]*template.Template
}

func Scan(dir string) (*Registry, error) {
	r := &Registry{files: map[string]string{}}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !hasIslandExtension(path) {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(strings.TrimSuffix(rel, filepath.Ext(rel)))

		if existing, ok := r.files[name]; ok {
			return fmt.Errorf("island %s is defined twice: %s and %s", name, existing, path)
		}
		if isHTMLElement(name) {
			return fmt.Errorf("island %s (%s) has the name of an HTML element, rename it", name, path)
		}
		r.files[name] = path
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning islands in %s: %w", dir, err)
	}
	return r, nil
}

func hasIslandExtension(path string) bool {
	ext := filepath.Ext(path)
	for _, e := range extensions {
		if ext == e {
			return true
		}
	}
	return false
}

func (r *Registry) Has(name string) bool {
	_, ok := r.files[name]
	return ok
}

func (r *Registry) SetComponents(components map[string]*template.Template) error {
	for name := range components {
		if r.Has(name) {
			return fmt.Errorf("%s is both an island and a component, rename one", name)
		}
	}
	r.components = components
	return nil
}

func (r *Registry) File(name string) (string, bool) {
	path, ok := r.files[name]
	return path, ok
}

func (r *Registry) SameNames(other *Registry) bool {
	if len(r.files) != len(other.files) {
		return false
	}
	for name := range r.files {
		if !other.Has(name) {
			return false
		}
	}
	return true
}
