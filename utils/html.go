package utils

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func DomHtml(el *goquery.Selection) string {
	html, _ := el.Html()
	return html
}

func DomRemovedSpecialCharsText(node *goquery.Selection) string {
	str := DomSanitizedText(node)
	m := regexp.MustCompile(`[-\[\]\(\)【】（）：:]`)
	str = m.ReplaceAllString(str, " ")
	return str
}

func DomSanitizedText(el *goquery.Selection) string {
	return SanitizeText(el.Text())
}

/*
 * DIY 了几个选择器语法（附加在标准CSS选择器字符串末尾）
 * @text 用于选择某个 Element 里的第一个 TEXT_NODE
 * @after 用于选择某个 Element 后面的 TEXT_NODE
 */
func DomSelectorText(el *goquery.Selection, selector string) (text string) {
	isTextNode := int64(0)
	if strings.HasSuffix(selector, "@text") {
		isTextNode = 1
		selector = selector[:len(selector)-5]
	} else if strings.HasSuffix(selector, "@after") {
		isTextNode = 2
		selector = selector[:len(selector)-6]
	}
	el = el.Find(selector)
	if el.Length() == 0 {
		return
	}
	if isTextNode == 1 {
		elNode := el.Get(0)
		node := elNode.FirstChild
		for node != nil {
			if node.Type == html.TextNode {
				text += SanitizeText(node.Data)
				break
			}
			node = node.NextSibling
		}
	} else if isTextNode == 2 {
		elNode := el.Get(0).NextSibling
		if elNode != nil {
			text = SanitizeText(elNode.Data)
		}
	} else {
		text = DomSanitizedText(el)
	}
	return
}

func GetUrlDoc(url string, cookie string, client *http.Client) (*goquery.Document, error) {
	res, err := FetchUrl(url, cookie, client)
	if err != nil {
		return nil, fmt.Errorf("can not fetch site data %v", err)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse site page DOM, error: %v", err)
	}
	return doc, nil
}