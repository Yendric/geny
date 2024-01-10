package util

import (
	"strings"

	"github.com/Yendric/geny/common"
)

func GenerateUrl(parts ...string) string {
	return StripEmpty("/"+strings.Join(parts, "/")) + "/"
}

func GenerateContentUrl(path string) string {
	path = StripHidden(path)
	path = StripExtension(path)
	path = strings.ReplaceAll(path, common.CONTENT_DIR, "")

	return "/" + StripEmpty(path) + "/"
}

func StripEmpty(in string) string {
	slice := strings.Split(in, "/")

	var out []string
	for _, part := range slice {
		if part != "" {
			out = append(out, part)
		}
	}

	return strings.Join(out, "/")
}
