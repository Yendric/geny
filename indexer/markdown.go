package indexer

import (
	"bytes"
	"html/template"

	"github.com/Yendric/geny/islands"
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
			islands.Extension,
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

type parsedMarkdown struct {
	metaData map[string]interface{}
	html     template.HTML
	islands  []islands.Usage
}

func (i *Indexer) parseMdFile(reg *islands.Registry, mdFile []byte) (parsedMarkdown, error) {
	var buf bytes.Buffer
	context, collector := islands.NewParserContext(reg)
	if err := i.md.Convert(mdFile, &buf, parser.WithContext(context)); err != nil {
		return parsedMarkdown{}, err
	}
	if err := collector.Err(); err != nil {
		return parsedMarkdown{}, err
	}

	return parsedMarkdown{
		metaData: meta.Get(context),
		html:     template.HTML(buf.String()),
		islands:  collector.Usages(),
	}, nil
}
