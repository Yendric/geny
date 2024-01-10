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

func ParseMdFile(mdFile []byte) (map[string]interface{}, template.HTML) {
	md := goldmark.New(
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

	var buf bytes.Buffer
	context := parser.NewContext()
	err := md.Convert(mdFile, &buf, parser.WithContext(context))
	if err != nil {
		panic(err)
	}

	metaData := meta.Get(context)

	return metaData, template.HTML(buf.String())
}
