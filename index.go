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
	close := make(chan bool)
	articles := Hook("http://www.tagesanzeiger.ch/rss.html", DefaultLink, close)

	for a := range articles {
		log.Println("article", a.Title)
	}
	//web.Get("/(.*)", hello)
	//web.Run("0.0.0.0:9999")
}
