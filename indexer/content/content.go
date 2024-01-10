package content

import (
	goTemplate "html/template"

	"github.com/Yendric/geny/indexer/template"
)

type ContentFile struct {
	Content     goTemplate.HTML
	RawContent  []byte
	Path        string
	FileName    string
	Url         string
	MetaData    map[string]interface{}
	Template    *template.Template
	Collections *Collections
}
