package main

import (
	"crypto/md5"
	rss "github.com/akuendig/go-rss"
	"io"
	"time"
)

type LinkChooser func(item *rss.Item) string

func DefaultLink(item *rss.Item) string {
	return item.Link
}

func Hook(url string, linkChooser LinkChooser, close chan bool) chan *Article {
	out := make(chan *Article)

	go func() {
		fetch(url, linkChooser, out)

		for {
			select {
			case <-time.After(time.Hour * 1):
				fetch(url, linkChooser, out)
			case <-close:
				return
			}
		}
	}()

	return out
}

func fetch(url string, linkChooser LinkChooser, out chan *Article) {
	channel, err := rss.Read(url)

	if err != nil {
		return
	}

	for _, item := range channel.Item {
		link := linkChooser(item)

		h := md5.New()
		io.WriteString(h, link)

		article := new(Article)

		article.Id = string(h.Sum(nil))
		article.Title = item.Title
		article.Link = link
		article.PubDate, _ = time.Parse(time.RFC822, item.PubDate)
		article.Summary = item.Description

		out <- article
	}
}

func TagiFeeds() []string {
	return []string{
		"http://www.tagesanzeiger.ch/rss.html",
		"http://www.tagesanzeiger.ch/rss_ticker.html",
		"http://www.tagesanzeiger.ch/zuerich/rss.html",
		"http://www.tagesanzeiger.ch/schweiz/rss.html",
		"http://www.tagesanzeiger.ch/ausland/rss.html",
		"http://www.tagesanzeiger.ch/wirtschaft/rss.html",
		"http://www.tagesanzeiger.ch/sport/rss.html",
		"http://www.tagesanzeiger.ch/kultur/rss.html",
		"http://www.tagesanzeiger.ch/panorama/rss.html",
		"http://www.tagesanzeiger.ch/leben/rss.html",
		"http://www.tagesanzeiger.ch/auto/rss.html",
		"http://www.tagesanzeiger.ch/digital/rss.html",
		"http://www.tagesanzeiger.ch/wissen/rss.html",
		"http://www.tagesanzeiger.ch/dienste/RSS/story/rss.html",
	}
}

func BlickFeeds() []string {
	return []string{
		"http://www.blick.ch/news/rss.xml",                        // News
		"http://www.blick.ch/news/schweiz/rss.xml",                // News/Schweiz
		"http://www.blick.ch/news/schweiz/aargau/rss.xml",         // News/Schweiz/Aargau
		"http://www.blick.ch/news/schweiz/basel/rss.xml",          // News/Schweiz/Basel
		"http://www.blick.ch/news/schweiz/bern/rss.xml",           // News/Schweiz/Bern
		"http://www.blick.ch/news/schweiz/graubuenden/rss.xml",    // News/Schweiz/Graubuenden
		"http://www.blick.ch/news/schweiz/ostschweiz/rss.xml",     // News/Schweiz/Ostschweiz
		"http://www.blick.ch/news/schweiz/tessin/rss.xml",         // News/Schweiz/Tessin
		"http://www.blick.ch/news/schweiz/westschweiz/rss.xml",    // News/Schweiz/Westschweiz
		"http://www.blick.ch/news/schweiz/zentralschweiz/rss.xml", // News/Schweiz/Zentralschweiz
		"http://www.blick.ch/news/schweiz/zuerich/rss.xml",        // News/Schweiz/Zuerich
		"http://www.blick.ch/news/ausland/rss.xml",                // News/Ausland
		"http://www.blick.ch/news/wirtschaft/rss.xml",             // News/Wirtschaft
		"http://www.blick.ch/news/wissenschaftundtechnik/rss.xml", // Wissen
		"http://www.blick.ch/sport/rss.xml",                       // Sport
		"http://www.blick.ch/sport/fussball/rss.xml",              // Sport/Fussball
		"http://www.blick.ch/sport/eishockey/rss.xml",             // Sport/Eishockey
		"http://www.blick.ch/sport/ski/rss.xml",                   // Sport/Ski
		"http://www.blick.ch/sport/tennis/rss.xml",                // Sport/Tennis
		"http://www.blick.ch/sport/formel1/rss.xml",               // Sport/Formel 1
		"http://www.blick.ch/sport/rad/rss.xml",                   // Sport/Rad
		"http://www.blick.ch/people/rss.xml",                      // People
		"http://www.blick.ch/unterhaltung/rss.xml",                // Unterhaltung
		"http://www.blick.ch/life/rss.xml",                        // Life
		"http://www.blick.ch/life/mode/rss.xml",                   // Life/Mode & Beauty
		"http://www.blick.ch/life/gourmet/rss.xml",                // Life/Gourmet
		"http://www.blick.ch/life/digital/rss.xml",                // Life/Digital
	}
}

func MinutenFeeds() []string {
	return []string{
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=1",   // Front
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=4",   // News
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=3",    // Ausland
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=2",    // Schweiz
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=8",   // Wirtschaft
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=19",   // Zuerich
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=20",   // Bern
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=2087", // Mittelland
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=21",   // Basel
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=112",  // Zentralschweiz
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=126",  // Ostschweiz
		"http://www.20min.ch/rss/rss.tmpl?type=rubrik&get=13",   // Panorama
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=28",  // People
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=9",   // Sport
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=10",  // Digital
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=11",  // Auto
		"http://www.20min.ch/rss/rss.tmpl?type=channel&get=25",  // Life
	}
}
