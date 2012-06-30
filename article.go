package main

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"exp/html"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ErrNoData  = errors.New("No Data stored in SiteData")
	ErrNoToken = errors.New("Not on a start tag")
)

type Article struct {
	Id         string
	Title      string
	Summary    string
	PubDate    time.Time "pubDate"
	Link       string
	WebsiteRaw []byte
	SiteData   *struct {
		Data       []byte
		Compressed bool
	} "site"
}

func (a *Article) Website() io.ReadCloser {
	if a.WebsiteRaw != nil {
		return flate.NewReader(bytes.NewReader(a.WebsiteRaw))
	}

	return nil
}

func (a *Article) SetWebsite(reader io.Reader) error {
	buffer := new(bytes.Buffer)
	writer, _ := flate.NewWriter(buffer, flate.BestCompression)

	if _, err := io.Copy(writer, reader); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	a.WebsiteRaw = buffer.Bytes()

	return nil
}

func (a *Article) Site() (io.ReadCloser, error) {
	if a.SiteData == nil || a.SiteData.Data == nil {
		return nil, ErrNoData
	}

	if a.SiteData.Compressed {
		decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(a.SiteData.Data))
		return zlib.NewReader(decoder)
	}

	return ioutil.NopCloser(bytes.NewReader(a.SiteData.Data)), nil
}

func (a *Article) DownloadWebsite() (io.ReadCloser, error) {
	var response, err = http.Get(a.Link)

	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func (a *Article) String() string {
	return fmt.Sprintf(
		`id: %s
        title: %s
        summary: %s
        published: %s
        link: %s`, a.Id, a.Title, a.Summary, a.PubDate, a.Link)
}

func ExtractTagi(reader io.Reader) (io.Reader, error) {
	root, err := html.Parse(reader)

	if err != nil {
		return nil, err
	}

	var r = toNode(root)

	defer r.Dispose()

	var sp = r.descendant(Id("singlePage"))

	if sp == nil {
		return nil, errors.New("singlePage not found \n" + r.String())
	}

	var p = sp.descendants(Tag("p"))

	if p == nil {
		return nil, errors.New("p's not found \n" + r.String())
	}

	var buffer = new(bytes.Buffer)

	for _, node := range p {
		html.Render(buffer, node.toNode())
		buffer.WriteByte('\n')
	}

	return buffer, nil
}

func ExtractBlickOld(reader io.Reader) (io.Reader, error) {
	root, err := html.Parse(reader)

	if err != nil {
		return nil, err
	}

	var r = toNode(root)
	defer r.Dispose()

	var art = r.descendant(Class("article"))
	if art == nil {
		return nil, errors.New("article not found \n" + r.String())
	}

	var buffer = new(bytes.Buffer)
	html.Render(buffer, art.toNode())

	return buffer, nil
}

type node html.Node

func toNode(n *html.Node) *node {
	return (*node)(n)
}

func (n *node) toNode() *html.Node {
	return (*html.Node)(n)
}

func (n *node) String() string {
	var buffer = new(bytes.Buffer)

	html.Render(buffer, n.toNode())

	return string(buffer.Bytes())
}

// Set all child references recursively to null such that there are no cycles in memory
func (n *node) Dispose() {
	for _, child := range n.Child {
		toNode(child).Dispose()
	}

	n.Parent = nil
	n.Child = nil
}

type predicate func(n *node) bool

func Id(id string) predicate {
	return func(n *node) bool {
		return n.hasAttribute("id", id)
	}
}

func Tag(tag string) predicate {
	return func(n *node) bool {
		return n.isTag(tag)
	}
}

func Class(class string) predicate {
	return func(n *node) bool {
		return n.hasAttribute("class", class)
	}
}

func (n *node) isTag(tag string) bool {
	return n.Type == html.ElementNode && n.Data == tag
}

func (n *node) hasAttribute(key, val string) bool {
	for _, a := range n.Attr {
		if a.Key == key && a.Val == val {
			return true
		}
	}

	return false
}

func (n *node) hasAttributes(attr map[string]string) bool {
	for _, a := range n.Attr {
		if attr[a.Key] == a.Val {
			delete(attr, a.Key)
		}
	}

	return len(attr) == 0
}

func (n *node) descendants(pred predicate) []*node {
	var res nodeVector

	descendantsRec(n, pred, &res)

	return res.arr
}

func descendantsRec(n *node, pred predicate, res *nodeVector) {
	if pred(n) {
		res.push(n)
	}

	for _, child := range n.Child {
		descendantsRec(toNode(child), pred, res)
	}
}

func (n *node) descendant(pred predicate) *node {
	if pred(n) {
		return n
	}

	for _, child := range n.Child {
		var found = toNode(child).descendant(pred)

		if found != nil {
			return found
		}
	}

	return nil
}

type nodeVector struct {
	arr []*node
}

func (s *nodeVector) push(val *node) {
	s.arr = append(s.arr, val)
}
