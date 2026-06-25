package util

import (
	"path/filepath"
	"strings"
)

func GeneratePath(parts ...string) string {
	return StripEmpty(strings.Join(parts, "/"))
}

func StripHidden(path string) string {
	out := []string{}

	for _, part := range strings.Split(path, "/") {
		if !strings.HasPrefix(part, "_") {
			out = append(out, part)
		}
	}

	return strings.Join(out, "/")
}

func StripExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}
