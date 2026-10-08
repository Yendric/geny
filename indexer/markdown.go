package indexer

import (
	"bytes"
	"html/template"

	"github.com/Yendric/geny/headings"
	"github.com/Yendric/geny/islands"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
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
		goldmark.WithParserOptions(parser.WithAutoHeadingID(), parser.WithHeadingAttribute()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

type parsedMarkdown struct {
	metaData map[string]interface{}
	html     template.HTML
	islands  []islands.Usage
	headings headings.Headings
}

func (i *Indexer) parseMdFile(reg *islands.Registry, mdFile []byte) (parsedMarkdown, error) {
	var buf bytes.Buffer
	context, collector := islands.NewParserContext(reg)
	doc := i.md.Parser().Parse(text.NewReader(mdFile), parser.WithContext(context))
	if err := i.md.Renderer().Render(&buf, mdFile, doc); err != nil {
		return parsedMarkdown{}, err
	}
	if err := collector.Err(); err != nil {
		return parsedMarkdown{}, err
	}

	pageHeadings, err := headings.Extract(doc, mdFile, i.md.Renderer())
	if err != nil {
		return parsedMarkdown{}, err
	}

	return parsedMarkdown{
		metaData: meta.Get(context),
		html:     template.HTML(buf.String()),
		islands:  collector.Usages(),
		headings: pageHeadings,
	}, nil
}
