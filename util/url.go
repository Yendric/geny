package util

import "strings"

func GenerateUrl(parts ...string) string {
	return StripEmpty("/"+strings.Join(parts, "/")) + "/"
}

func GenerateContentUrl(contentDir, path string) string {
	path = StripHidden(path)
	path = StripExtension(path)
	path = strings.ReplaceAll(path, contentDir, "")

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
