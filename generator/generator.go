package generator

import (
	"html/template"
	"os"
	"strings"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/indexer/content"
	"github.com/Yendric/geny/util"
)

var funcMap = template.FuncMap{
	"stripTags":      util.StripTags,
	"truncate":       util.Truncate,
	"getCurrentYear": util.GetCurrentYear,
}

func GenerateFiles(contentFiles []content.ContentFile) error {
	collections := generateCollections(contentFiles)

	for _, contentFile := range contentFiles {
		contentFile.Collections = collections

		err := generateFile(contentFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateFile(contentFile content.ContentFile) error {
	template, err := template.New(contentFile.Template.Name + ".html").Funcs(funcMap).ParseGlob(common.TEMPLATES_DIR + "/*.html")
	if err != nil {
		return err
	}
	template, err = template.ParseGlob(common.TEMPLATES_DIR + "/**/*.html")
	if err != nil {
		return err
	}

	whereTo := util.StripHidden(contentFile.Path)
	whereTo = strings.ReplaceAll(whereTo, common.CONTENT_DIR, common.BUILD_DIR)
	whereTo = util.StripExtension(whereTo)
	whereTo = util.StripEmpty(whereTo)

	var buildFile *os.File
	if contentFile.FileName == "index.md" || contentFile.FileName == "404.md" {
		buildFile, err = os.Create(util.GeneratePath(whereTo + ".html"))
		if err != nil {
			return err
		}
	} else {
		err = os.MkdirAll(whereTo, os.ModePerm)
		if err != nil {
			return err
		}

		buildFile, err = os.Create(util.GeneratePath(whereTo, "index.html"))
		if err != nil {
			return err
		}
	}

	err = template.Execute(buildFile, contentFile)
	if err != nil {
		return err
	}

	buildFile.Close()

	return nil
}
