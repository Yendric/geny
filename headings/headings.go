package headings

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Yendric/geny/islands"
	"github.com/Yendric/geny/util"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type Heading struct {
	Level    int
	ID       string
	Text     string
	HTML     template.HTML
	Children Headings
}

type Headings []Heading

func (h Headings) Between(from, to int) Headings {
	var flat []Heading
	h.walk(func(heading Heading) {
		if heading.Level >= from && heading.Level <= to {
			heading.Children = nil
			flat = append(flat, heading)
		}
	})
	return nest(flat)
}

func (h Headings) walk(fn func(Heading)) {
	for _, heading := range h {
		fn(heading)
		heading.Children.walk(fn)
	}
}

func nest(flat []Heading) Headings {
	var out Headings
	for i := 0; i < len(flat); {
		heading := flat[i]
		end := i + 1
		for end < len(flat) && flat[end].Level > heading.Level {
			end++
		}
		heading.Children = nest(flat[i+1 : end])
		out = append(out, heading)
		i = end
	}
	return out
}

func Extract(doc ast.Node, source []byte, r renderer.Renderer) (Headings, error) {
	var flat []Heading
	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if node.Kind() == islands.KindBlock {
			return ast.WalkSkipChildren, nil
		}
		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		var buf bytes.Buffer
		for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
			if err := r.Render(&buf, source, child); err != nil {
				return ast.WalkStop, err
			}
		}
		html := template.HTML(buf.String())

		id, _ := heading.AttributeString("id")
		idBytes, ok := id.([]byte)
		if !ok {
			return ast.WalkStop, fmt.Errorf("heading %q has no id", util.StripTags(html))
		}

		flat = append(flat, Heading{
			Level: heading.Level,
			ID:    string(idBytes),
			Text:  util.StripTags(html),
			HTML:  html,
		})
		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		return nil, err
	}
	return nest(flat), nil
}
