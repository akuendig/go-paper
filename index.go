package main

import (
	"log"
	"regexp"
	"net/http"
	_ "net/http/pprof"
)

const (
	batchSize = 100
)

func main() {
	var iter, close = Articles("blick")
	var a Article

	defer close()

	if !iter.Next(&a) {
		log.Fatal(iter.Err())
	}
	if !iter.Next(&a) {
		log.Fatal(iter.Err())
	}

	go CompactBlick()

	http.ListenAndServe(":6060", nil)

	//web.Get("/(.*)", hello)
	//web.Run("0.0.0.0:9999")
}

var oldLinkRex = regexp.MustCompile(`(.+)-(\d+)$`)

func HasOldBlickLink(l string) bool {
	return oldLinkRex.MatchString(l)
}

func CompactBlick() {
	var i = 0
	var hasData = true

	for ; hasData; i += batchSize {
		log.Println("Fetching batch", i)

		var batch, err = ReadOldBatch("blick", i, batchSize)
		hasData = len(batch) > 0

		if err != nil {
			log.Fatal(err)
		}

		for _, a := range batch {
			if a.WebsiteRaw != nil {
				a.SiteData = nil
				continue
			}

			var site, err = a.Site()

			if err == ErrNoData {
				continue
			}

			if err != nil {
				log.Println("Error at id", a.Id, err)
				continue
			}

			defer site.Close()

			if !HasOldBlickLink(a.Link) {
				continue
			}

			text, err := ExtractBlickOld(site)

			if err != nil {
				log.Println("Error at id", a.Id, err)
				continue
			}

			if err := a.SetWebsite(text); err != nil {
				log.Println("Error at id", a.Id, err)
				continue
			}

			a.SiteData = nil
		}

		UpdateWebsiteBatch("blick", batch)
		log.Println("Pushed batch", i)
	}
}

func CompactTagi() {
	var i = 0
	var hasData = true

	for ; hasData; i += batchSize {
		log.Println("Fetching batch", i)

		var batch, err = ReadOldBatch("tagi", i, batchSize)
		hasData = len(batch) > 0

		if err != nil {
			log.Fatal(err)
		}

		for _, a := range batch {
			if a.WebsiteRaw != nil {
				a.SiteData = nil
				continue
			}

			var site, err = a.Site()

			if err == ErrNoData {
				continue
			}

			if err != nil {
				log.Println("Error at id", a.Id, err)
				continue
			}

			defer site.Close()
			text, err := ExtractTagi(site)

			if err != nil {
				log.Println("Error at id", a.Id, err)
				continue
			}

			if err := a.SetWebsite(text); err != nil {
				log.Println("Error at id", a.Id, err)
				continue
			}

			a.SiteData = nil
		}

		UpdateWebsiteBatch("tagi", batch)
		log.Println("Pushed batch", i)
	}
}
