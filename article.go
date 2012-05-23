package main

import (
	"bytes"
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
	ErrNoData = errors.New("No Data stored in SiteData")
	ErrNoToken = errors.New("Not on a start tag")
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

func (a *Article) SetSite(reader io.Reader) error {
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	writer := zlib.NewWriter(encoder)

	_, err := io.Copy(writer, reader)

	if err != nil {
		return err
	}

	err = writer.Flush()

	if err != nil {
		return err
	}

	a.SiteData.Data = buffer.Bytes()
	a.SiteData.Compressed = true

	return nil
}

func (a *Article) Site() (io.ReadCloser, error) {
	if a.SiteData.Data == nil {
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

	var query = query{nizer: html.NewTokenizer(site) }
	err = query.moveAttr("id", "singleLeft")

	if err != nil {
		return "", err
	}

	err = query.moveTag("p")

	if err != nil {
		return "", err
	}

	return fmt.Sprint(query.current, string(query.nizer.Text())), nil
}

type query struct {
	nizer *html.Tokenizer
	current *html.Token
}

func (q *query) moveTag(tagName string) error {
	var tnizer = q.nizer

	for {
		var tt = tnizer.Next()

		if tt == html.ErrorToken {
			q.current = nil
			return tnizer.Err()
		}

		if tt == html.EndTagToken {
			continue
		}

		var attrs []html.Attribute
		name, moreAttr := tnizer.TagName()

		if string(name) != tagName {
			continue
		}

		for moreAttr {
			var key, val []byte

			key, val, moreAttr = tnizer.TagAttr()

			attrs = append(attrs, html.Attribute{"", string(key), string(val)})
		}

		q.current = &html.Token{tt, string(name), attrs}
		return nil
	}

	return tnizer.Err()
}

func (q *query) moveAttr(key, val string) error {
	var tnizer = q.nizer

	for {
		var tt = tnizer.Next()

		if tt == html.ErrorToken {
			q.current = nil
			return tnizer.Err()
		}

		if tt == html.EndTagToken {
			continue
		}

		var attrs []html.Attribute
		name, moreAttr := tnizer.TagName()

		for moreAttr {
			var key, val []byte
			key, val, moreAttr = tnizer.TagAttr()

			attrs = append(attrs, html.Attribute{"", string(key), string(val)})
		}

		for _, attr := range attrs {
			if attr.Key == key && attr.Val == val {
				q.current = &html.Token{tt, string(name), attrs}
				return nil
			}
		}
	}

	return tnizer.Err()
}

func (q *query) node() (*Node, error) {
	if q.current == nil || q.current.Type != html.StartTagToken {
		return nil, ErrNoToken
	}

	var token html.Token
	token = *q.current

	var tnizer = q.nizer
	var root = &Node{token.Data, "", token.Attr, nil, nil}

	stack := new(nodeStack)
	stack.push(root)

	for stack.count() > 0 {
		token = tnizer.Token()

		switch token.Type {
		case html.ErrorToken:
			q.current = nil
			return tnizer.Err()
		case html.TextToken:
			var cur = stack.peek()
			cur.Text = token.Data
		case html.CommentToken, html.DoctypeToken:
			// TODO skipped
		case html.StartTagToken:
			var par = stack.peek()
			var child = &Node{token.Data, "", token.Attr, nil}

			par.Children = append(par.Children, child)
			stack.push(child)
		case html.SelfClosingTagToken:
			var par = stack.peek()
			var child = &Node{token.Data, "", token.Attr, nil}

			par.Children = append(par.Children, child)
		case html.EndTagToken:
			var par = stack.pop()

			if par.Name != token.Data {
				fmt.Print("nonmaching end token", token, "Should be", par.Name)
			}
		default:
			fmt.Print("unrecognized token:", token)
		}
	}

	q.current = &token
}

type Node struct {
	Name, Text string
	Attributes []html.Attribute
	Children []*Node
}

type nodeStack struct {
	stack []*Node
}

func (s *nodeStack) count() int {
	return len(s.stack)
}

func (s *nodeStack) clear() {
	s.stack = nil
}

func (s *nodeStack) push(node *Node) {
	s.stack = append(s.stack, node)
}

func (s *nodeStack) peek() *Node {
	var l = len(s.stack)

	if l == 0 {
		return nil
	}

	return s.stack[l - 1]
}

func (s *nodeStack) pop() *Node {
	var l = len(s.stack)

	if l == 0 {
		return nil
	}

	var res = s.stack[l-1]
	s.stack = s.stack[:l-1]

	return res
}
