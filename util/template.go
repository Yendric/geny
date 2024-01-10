package util

import (
	"html/template"
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
