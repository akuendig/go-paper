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

    go func () {
        for a := range f.Articles {
            log.Println("article", a.Title)
        }
    } ()

    // f.Once()
    article, _ := FirstArticle()
    article.ExtractText()

	//web.Get("/(.*)", hello)
	//web.Run("0.0.0.0:9999")
}
