package indexer

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			highlighting.NewHighlighting(
				highlighting.WithStyle("vulcan"),
			),
			meta.Meta,
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

func (i *Indexer) parseMdFile(mdFile []byte) (map[string]interface{}, template.HTML, error) {
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := i.md.Convert(mdFile, &buf, parser.WithContext(context)); err != nil {
		return nil, "", err
	}

	return meta.Get(context), template.HTML(buf.String()), nil
}
