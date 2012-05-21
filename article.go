package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"errors"
)

type Article struct {
	Id       string
	Title    string
	Summary  string
	PubDate  time.Time "pubDate"
	Link     string
	SiteData struct {
		Data       []byte
		Compressed bool
	} "site"
}

func (a *Article) SetSite(value io.Reader) error {
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	writer := zlib.NewWriter(encoder)

	_, err := io.Copy(writer, value)

	if err != nil {
		return err
	}

	a.SiteData.Data = buffer.Bytes()
	a.SiteData.Compressed = true

	return nil
}

func (a *Article) Site() (io.ReadCloser, error) {
	if a.SiteData.Data == nil {
		return ioutil.NopCloser(new(bytes.Buffer)), nil
	}

	if a.SiteData.Compressed {
		decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(a.SiteData.Data))
		return zlib.NewReader(decoder)
	}

	return ioutil.NopCloser(bytes.NewReader(a.SiteData.Data)), nil
}

func (a *Article) DownloadWebsite() error {
	res, err := http.Get(a.Link)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	a.SetSite(res.Body)

	return nil
}

const selector = "html body div#mainWrapper div#mailLeftWrapper div#mainContainer div#singlePage div#singleLeft p"

type xmlDecoder xml.Decoder

func xmlExtract(reader io.Reader) string {
	doc := xml.NewDecoder(reader)
	doc.Strict = false

	readUntilElement("body", doc)

	return ""
}

func readUntilElement(name string, doc *xml.Decoder) (*xml.StartElement, error) {
	for i := 0; i < 10; i++ {
		token, err := doc.Token()

		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case *xml.StartElement:
			log.Println("StartElement", t.Name)
			if t.Name.Local == name {
				return t, nil
			}
		case xml.EndElement:
			// log.Println("EndElement")
		case xml.Directive:
			// log.Println("Directive", string(t))
		case xml.CharData:
			// log.Println("CharData", string(t))
		case xml.Comment:
			// log.Println("Comment", string(t))
		case xml.ProcInst:
			// log.Println("ProcInst", t)
		}
	}

	return nil, errors.New(fmt.Sprint("Element", name, "not found"))
}

func (a *Article) ExtractText() string {
	site, err := a.Site()

	if err != nil {
		log.Println(err)
		return ""
	}

	defer site.Close()
	xmlExtract(site)
 //    node, err := transform.NewDocFromReader(site)

	// if err != nil {
	// 	log.Println(err)
	// 	return ""
	// }

	// selector := transform.NewSelectorQuery("div#singleLeft", "p", "p")
	// matches := selector.Apply(node)

	// if len(matches) < 1 {
	// 	log.Println("No nodes found")
	// 	return ""
	// }

	// for _, paragraph := range matches {
	// 	log.Println(paragraph)
	// }

	return ""
}

func (a Article) String() string {
	return fmt.Sprintf(
		`id: %s
        title: %s
        summary: %s
        published: %s
        link: %s
        compressed: %t`, a.Id, a.Title, a.Summary, a.PubDate, a.Link, a.SiteData.Compressed)
}
