package main

import (
	"crypto/md5"
	rss "github.com/akuendig/go-rss"
	"io"
	"time"
)

type LinkChooser func(item *rss.Item) string

type Fetch struct {
	Urls        []string
	Articles    chan *Article
	LinkChooser LinkChooser

	stopChannel    chan bool
	stoppedChannel chan bool
}

func NewFetch(urls []string, linkChooser LinkChooser) *Fetch {
	var fetch = new(Fetch)

	fetch.Urls = urls
	fetch.LinkChooser = linkChooser
	fetch.Articles = make(chan *Article)

	return fetch
}

func (f *Fetch) Once() {
	finished := make(chan bool)

	for _, url := range f.Urls {
		go func(u string) {
			f.fetch(u)
			finished <- true
		}(url)
	}

	for i := 0; i < len(f.Urls); i++ {
		<-finished
	}
}

func (f *Fetch) Again(interval time.Duration) {
	if f.stopChannel != nil {
		f.Stop()
	}

	f.stopChannel = make(chan bool)
	f.stoppedChannel = make(chan bool)

	go func() {
		for {
			select {
			case <-time.After(interval):
				f.Once()
			case <-f.stopChannel:
				f.stoppedChannel <- true
				return
			}
		}
	}()
}

func (f *Fetch) Stop() {
	if f.stopChannel == nil {
		return
	}

	f.stopChannel <- true
	<-f.stoppedChannel

	f.stopChannel = nil
	f.stoppedChannel = nil
}

func DefaultLink(item *rss.Item) string {
	return item.Link
}

func (f *Fetch) fetch(url string) {
	channel, err := rss.Read(url)

	if err != nil {
		return
	}

	for _, item := range channel.Item {
		link := f.LinkChooser(item)

		h := md5.New()
		io.WriteString(h, link)

		article := new(Article)

		article.Id = string(h.Sum(nil))
		article.Title = item.Title
		article.Link = link
		article.PubDate, _ = time.Parse(time.RFC822, item.PubDate)
		article.Summary = item.Description

		f.Articles <- article
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
