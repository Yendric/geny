package generator

import "github.com/Yendric/geny/indexer/content"

func generateCollections(contentFiles []content.ContentFile) *content.Collections {
	var collections = content.Collections{}

	for _, contentFile := range contentFiles {
		collections[contentFile.Template.Name] = append(collections[contentFile.Template.Name], contentFile)
		contentFile.Collections = &collections
	}

	return &collections
}
