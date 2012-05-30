package main

import (
	"log"
)

const (
	batchSize = 50
)

func main() {
	var i = 0
	var hasData = true

	for ; hasData; i += batchSize {
		log.Println("Fetching batch", i)

		var batch, err = ReadBatch("tagi", i, batchSize)
		hasData = len(batch) > 0

		if err != nil {
			log.Fatal(err)
		}

		for _, a := range batch {
			var site, err = a.Site()

			if err == ErrNoData {
				continue
			}

			if err != nil {
				log.Fatal(err)
			}

			defer site.Close()
			text, err := ExtractText(site)

			if err != nil {
				log.Fatal(err)
			}

			if err := a.SetWebsite(text); err != nil {
				log.Fatal(err)
			}

			a.SiteData = nil
		}

		UpdateBatch("tagi", batch)
		log.Println("Pushed batch", i)
	}

	//web.Get("/(.*)", hello)
	//web.Run("0.0.0.0:9999")
}
