package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"time"
    "net/http"
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

func (a *Article) Site() (string, error) {
	var data []byte

	if a.SiteData.Data == nil {
		return "", nil
	}

	if a.SiteData.Compressed {
		decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(a.SiteData.Data))
		reader, err := zlib.NewReader(decoder)

		if err != nil {
			return "", err
		}

		defer reader.Close()
		data, err = ioutil.ReadAll(reader)

		if err != nil {
			return "", err
		}
	} else {
		data = a.SiteData.Data
	}

	return string(data), nil
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

func (a Article) String() string {
	return fmt.Sprintf(
		`id: %s
        title: %s
        summary: %s
        published: %s
        link: %s
        compressed: %t`, a.Id, a.Title, a.Summary, a.PubDate, a.Link, a.SiteData.Compressed)
}
