package infrastructure

import (
	"strings"

	"golang.org/x/net/html"
)

// Descriptions predate this portal and may contain rich HTML. Public cards
// receive text only; scripts/styles/templates must not become visible prose.
func plainDescription(value string) string {
	doc, err := html.Parse(strings.NewReader(value))
	if err != nil {
		return ""
	}
	var out strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "head", "script", "style", "template", "noscript":
				return
			}
		}
		if node.Type == html.TextNode {
			out.WriteString(node.Data)
		}
		block := false
		if node.Type == html.ElementNode {
			switch node.Data {
			case "p", "div", "br", "li", "ul", "ol", "table", "tr", "td", "th", "h1", "h2", "h3", "h4", "h5", "h6", "section", "article", "blockquote", "pre", "hr":
				block = true
			}
		}
		if block {
			out.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if block {
			out.WriteByte(' ')
		}
	}
	walk(doc)
	return strings.Join(strings.Fields(out.String()), " ")
}
