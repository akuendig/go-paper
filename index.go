package main

import (
	"log"

//"github.com/hoisie/web"
// "fmt"
)

// func hello(ctx *web.Context, val string) {
// 	for k, v := range ctx.Params {
// 		println(k, v)
// 	}
// }

func main() {
	f := NewFetch(TagiFeeds(), DefaultLink)

	// go func () {
	//     for a := range f.Articles {
	//         log.Println("article", a.Title)
	//     }
	// } ()

	go f.Once()

	article := <-f.Articles
	err := article.DownloadWebsite()

	if err != nil {
		log.Fatal(err)
	}

	log.Println(article.ExtractText())
	// site, _ := article.Site()

	// io.Copy(os.Stdout, site)
	// site.Close()

	//web.Get("/(.*)", hello)
	//web.Run("0.0.0.0:9999")
}
