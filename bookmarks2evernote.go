// bookmarks2evernote will read an exported html file of bookmarks and create an evernote file (.enex)
// that can be imported into evernote. Ibe note is created for each bookmark. It is best to import
// the resulting evernote file into its own notebook. Tags and bookmark description is preserved along
// with the title and url of the bookmark. This program currently works best when using an export file from
// Delicious, or otherwise when the list of bookmarks is a flat list (no folders).
package main

import (
	"fmt"
	"golang.org/x/net/html"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
	"bytes"
	"io"
)

const enexRawTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE en-export SYSTEM "http://xml.evernote.com/pub/evernote-export3.dtd">
<en-export application="Evernote">%s</en-export>`

const noteRawTmpl = `<note>
	<title><![CDATA[{{.Title}}]]></title>
	<content><![CDATA[<?xml version="1.0" encoding="utf-8"?><!DOCTYPE en-note SYSTEM "http://xml.evernote.com/pub/enml2.dtd"><en-note><a href="{{.Url}}">{{.Description}}</a></en-note>]]></content>
	{{.Tags}}
	<created>{{.Date}}</created>
	<updated>{{.Date}}</updated>
	<note-attributes>
		<source-url>{{.Url}}</source-url>
	</note-attributes>
</note>`

var noteTmpl = template.Must(template.New("note").Parse(noteRawTmpl))

type note struct {
	Title string
	Tags  string
	Date  string
	Url   string
	Description string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <input file> <output file>\n", os.Args[0])
		os.Exit(1)
	}

	input, output := os.Args[1], os.Args[2]

	inputFile, err := os.Open(input)
	if err != nil {
		panic(err)
	}

	// Parse input as
	doc, err := html.Parse(inputFile)
	inputFile.Close()
	if err != nil {
		panic(err)
	}

	// Find bookmark list
	for doc = doc.FirstChild; doc != nil; {
		if doc.Data == "html" {
			doc = doc.FirstChild
		} else if doc.Data == "body" {
			doc = doc.FirstChild
		} else if doc.Data == "dl" {
			if doc.FirstChild.Data == "p" {
				doc = doc.FirstChild.NextSibling
			} else {
				doc = doc.FirstChild
			}
			break
		} else {
			doc = doc.NextSibling
		}
	}

	if doc == nil {
		panic("Could not find list of bookmarks")
	}

	// navigate xml and generate note strings (using a note template)
	bookmarks := []note{}
	for doc != nil {
		doc, bookmarks = processBookmark(doc, bookmarks)
	}

	// concatenate notes,
	noteStrings := []string{}
	for _, bookmark := range bookmarks {
		noteString := bytes.NewBufferString("")
		noteTmpl.Execute(noteString, bookmark)
		noteStrings = append(noteStrings, noteString.String())
	}

	// and generate .enex file using a template
	outputFile, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	_, err = io.WriteString(outputFile, fmt.Sprintf(enexRawTmpl, strings.Join(noteStrings, "")))
	if err != nil {
		panic(err)
	}

	outputFile.Close()
}

func processBookmark(doc *html.Node, bookmarks []note) (*html.Node, []note) {
	if doc == nil {
		return nil, bookmarks
	}

	//fmt.Printf("%#v\n", doc)

	if doc.Type == html.ElementNode && doc.Data == "dt" {
		n := note{}
		link := doc.FirstChild

		n.Title = link.FirstChild.Data
		for _, attr := range link.Attr {
			switch attr.Key {
			case "href":
				n.Url = attr.Val
			case "add_date":
				epochSecs, err := strconv.ParseInt(attr.Val, 10, 64)
				if err != nil {
					panic(err)
				}

				date := time.Unix(epochSecs, 0)
				n.Date = fmt.Sprintf(
					"%d%02d%02dT%02d%02d%02dZ",
					date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(),
				)

			case "tags":
				if attr.Val != "" {
					n.Tags = "<tag>" + strings.Join(strings.Split(attr.Val, ","), "</tag><tag>") + "</tag>"
				}
			}
		}
		n.Description = n.Url
		bookmarks = append(bookmarks, n)

		nextNode := doc.NextSibling
		if nextNode != nil && nextNode.Type == html.ElementNode && nextNode.Data == "dd" {
			n.Description = nextNode.FirstChild.Data
			return nextNode.NextSibling, bookmarks
		}
	}

	return doc.NextSibling, bookmarks
}
