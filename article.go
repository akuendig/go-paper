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
	Id       string
	Title    string
	Summary  string
	PubDate  time.Time "pubDate"
	Link     string
	Website  []byte
	SiteData *struct {
		Data       []byte
		Compressed bool
	} "site"
}

func (a *Article) SetSite(reader io.Reader) error {
	buffer := new(bytes.Buffer)
	writer, _ := flate.NewWriter(buffer, flate.BestCompression)

	if _, err := io.Copy(writer, reader); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	a.Website = buffer.Bytes()

	return nil
}

func (a *Article) Site() (io.ReadCloser, error) {
	if a.Website != nil {
		return flate.NewReader(bytes.NewReader(a.Website)), nil
	}

	if a.SiteData == nil || a.SiteData.Data == nil {
		return nil, ErrNoData
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

	err = a.SetSite(res.Body)
	res.Body.Close()

	return err
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

const selector = "html body div#mainWrapper div#mailLeftWrapper div#mainContainer div#singlePage div#singleLeft p"

func (a *Article) ExtractText() (string, error) {
	site, err := a.Site()

	if err != nil {
		return "", err
	}

	defer site.Close()

	root, err := html.Parse(site)

	if err != nil {
		return "", err
	}

	var r = toNode(root)
	var sL = r.getId("singleLeft")
	var p = sL.descendants(func(n *node) bool {
		return n.isTag("p")
	})

	return fmt.Sprint(p), nil
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

type predicate func(n *node) bool

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

func (n *node) where(pred predicate) *node {
	if pred(n) {
		return n
	}

	for _, child := range n.Child {
		var res = toNode(child).where(pred)

		if res != nil {
			return res
		}
	}

	return nil
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

func (n *node) getId(id string) *node {
	return n.where(func(n *node) bool {
		return n.hasAttribute("id", id)
	})
}

func (n *node) nextTag(tag string) *node {
	return n.where(func(n *node) bool {
		return n.isTag(tag)
	})
}

func (n *node) nextClass(class string) *node {
	return n.where(func(n *node) bool {
		return n.hasAttribute("class", class)
	})
}

type nodeVector struct {
	arr []*node
}

func (s *nodeVector) push(val *node) {
	s.arr = append(s.arr, val)
}
