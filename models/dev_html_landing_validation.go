package models

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
)

func validateDevHTMLLanding(pageHTML string) error {
	doc, err := html.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return errors.New("HTML Landing Page: gecersiz HTML")
	}

	forms := map[string]bool{}
	var walk func(*html.Node, func(*html.Node))
	walk = func(n *html.Node, fn func(*html.Node)) {
		fn(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, fn)
		}
	}

	attr := func(n *html.Node, key string) string {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val
			}
		}
		return ""
	}

	hasAttr := func(n *html.Node, key string) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return true
			}
		}
		return false
	}

	walk(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			if id := attr(n, "id"); id != "" {
				forms[id] = true
			}
		}
	})

	found := false
	var check func(*html.Node, bool)
	check = func(n *html.Node, inForm bool) {
		if n.Type == html.ElementNode {
			if n.Data == "form" {
				inForm = true
			}

			if n.Data == "button" || n.Data == "a" {
				var label strings.Builder
				label.WriteString(attr(n, "title"))
				label.WriteString(" ")
				label.WriteString(attr(n, "aria-label"))

				walk(n, func(child *html.Node) {
					if child.Type == html.TextNode {
						label.WriteString(" ")
						label.WriteString(child.Data)
					}
				})

				s := strings.ToLower(label.String())
				isDownload := strings.Contains(s, "indir") ||
					strings.Contains(s, "i̇ndir") ||
					strings.Contains(s, "download")

				typ := strings.ToLower(attr(n, "type"))
				_, linkedForm := forms[attr(n, "form")]

				validButton := n.Data == "button" &&
					!hasAttr(n, "disabled") &&
					typ != "button" && typ != "reset" &&
					(inForm || (attr(n, "form") != "" && linkedForm))

				validLink := n.Data == "a" &&
					attr(n, "href") != ""

				if isDownload && (validButton || validLink) {
					found = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			check(c, inForm)
		}
	}
	check(doc, false)

	// Runtime binds download-intent forms and legacy <a id="download">.
	// Reject pages which only contain a visually labelled but unbindable
	// control; otherwise campaign creation succeeds but landing returns 422.
	legacyAnchor := false
	walk(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" &&
			attr(n, "id") == "download" && attr(n, "href") != "" {
			legacyAnchor = true
		}
	})
	if !found || (!forms["download-intent"] && !legacyAnchor) {
		return errors.New("File Landing Page: indir butonu download-intent formuna bagli olmali veya <a id=download> kullanilmali")
	}
	return nil
}
