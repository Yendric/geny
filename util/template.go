package util

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func Truncate(text string) string {
	if len(text) > 150 {
		return text[:150] + "..."
	}
	return text
}

func StripTags(htmlText template.HTML) string {
	htmlin := strings.NewReader(string(htmlText))
	doc, err := html.Parse(htmlin)
	if err != nil {
		return ""
	}
	skip := map[string]bool{
		"script":   true,
		"style":    true,
		"textarea": true,
		"title":    true,
	}
	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			if n.Parent.Type == html.ElementNode && !skip[strings.ToLower(n.Parent.Data)] {
				sb.WriteString(n.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return sb.String()
}

func GetCurrentYear() int {
	return time.Now().Year()
}

type FileInfo struct {
	Name  string
	Size  string
	Lines int
}

func GetFileInfo(path string) (FileInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileInfo{}, err
	}
	lines := bytes.Count(data, []byte{'\n'})
	if len(data) > 0 && data[len(data)-1] != '\n' {
		lines++
	}
	return FileInfo{Name: filepath.Base(path), Size: humanSize(len(data)), Lines: lines}, nil
}

func humanSize(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	size, units := float64(n)/1024, "KMGT"
	for size >= 1024 && len(units) > 1 {
		size /= 1024
		units = units[1:]
	}
	return fmt.Sprintf("%.1f %cB", size, units[0])
}
