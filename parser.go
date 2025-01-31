package gohtml

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// parse parses a stirng and converts it into an html.
func parse(r io.Reader) *htmlDocument {
	htmlDoc := &htmlDocument{}
	tokenizer := html.NewTokenizer(r)
	for {
		if errorToken, _, _ := parseToken(tokenizer, htmlDoc, nil); errorToken {
			break
		}
	}
	return htmlDoc
}

func parseToken(tokenizer *html.Tokenizer, htmlDoc *htmlDocument, parent *tagElement) (bool, bool, string) {
	tokenType := tokenizer.Next()
	switch tokenType {
	case html.ErrorToken:
		return true, false, ""
	case html.CommentToken:
		// Remove all comments except Internet Explorer Conditionals
		raw := string(tokenizer.Raw())
		if strings.Contains(raw, "[if ") {
			textElement := &textElement{text: raw}
			appendElement(htmlDoc, parent, textElement)
		}
	case html.TextToken:
		text := string(tokenizer.Raw())
		text = strings.TrimSpace(text)
		if text == "" {
			break
		}
		textElement := &textElement{text: text}
		appendElement(htmlDoc, parent, textElement)
	case html.StartTagToken:
		tagElement := &tagElement{tagName: getTagName(tokenizer), startTagRaw: string(tokenizer.Raw())}
		appendElement(htmlDoc, parent, tagElement)
		for {
			errorToken, parentEnded, unsetEndTag := parseToken(tokenizer, htmlDoc, tagElement)
			if errorToken {
				return true, false, ""
			}
			if parentEnded {
				if unsetEndTag != "" {
					return false, false, unsetEndTag
				}
				break
			}
			if unsetEndTag != "" {
				return false, false, setEndTagRaw(tokenizer, tagElement, unsetEndTag)
			}
		}
	case html.EndTagToken:
		return false, true, setEndTagRaw(tokenizer, parent, getTagName(tokenizer))
	case html.DoctypeToken, html.SelfClosingTagToken:
		tagElement := &tagElement{tagName: getTagName(tokenizer), startTagRaw: string(tokenizer.Raw())}
		appendElement(htmlDoc, parent, tagElement)
	}
	return false, false, ""
}

// appendElement appends the element to the htmlDocument or parent tagElement.
func appendElement(htmlDoc *htmlDocument, parent *tagElement, e element) {
	if parent != nil {
		parent.appendChild(e)
	} else {
		htmlDoc.append(e)
	}
}

// getTagName gets a tagName from tokenizer.
func getTagName(tokenizer *html.Tokenizer) string {
	tagName, _ := tokenizer.TagName()
	return string(tagName)
}

// setEndTagRaw sets an endTagRaw to the parent.
func setEndTagRaw(tokenizer *html.Tokenizer, parent *tagElement, tagName string) string {
	if parent != nil && parent.tagName == tagName {
		parent.endTagRaw = string(tokenizer.Raw())
		return ""
	}
	return tagName
}
