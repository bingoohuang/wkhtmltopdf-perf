package main

import (
	"bytes"
	"github.com/gocolly/colly/v2"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

func main() {
	c := colly.NewCollector(colly.MaxDepth(1))

	// Before making a request print "Visiting ..."
	// Handler the saving of search_index.json in this function
	// since according to my several tests when visiting the json link
	// the OnRequest function sometimes was not been called
	c.OnRequest(func(r *colly.Request) {
	})

	var page []byte

	c.OnScraped(func(response *colly.Response) {
		// create a file of the given name and in the given path
		f, _ := os.Create("zz.html")
		f.Write(page)
		f.Close()
	})

	// mainly handle html pages
	c.OnResponse(func(r *colly.Response) {
		url := r.Request.URL.String()
		log.Printf("url: %s", url)
		page = GetContent(url)
	})

	// Handle CSS files
	c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Printf("link: %s", link)
		log.Printf("url: %s", e.Request.AbsoluteURL(link))
		// <link href="/license-admin/resources/1.0.0/css/bootstrap.min.css" type="text/css" rel="stylesheet" />
		linkRegexp := regexp.MustCompile(`<.*?="` + regexp.QuoteMeta(link) + `.*?>`)
		css := GetContent(e.Request.AbsoluteURL(link))
		subs := linkRegexp.FindAllSubmatchIndex(page, -1)
		var b bytes.Buffer
		last := 0
		for i, sub := range subs {
			b.Write(page[last:sub[0]])
			if i == 0 {
				b.Write([]byte("<style>\n"))
				b.Write(css)
				b.Write([]byte("</style>\n"))
			}
			last = sub[1]
		}

		if last < len(page) {
			b.Write(page[last:])
		}

		page = b.Bytes()
	})

	// Handle JavaScript files
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		log.Printf("src: %s", link)
		log.Printf("url: %s", e.Request.AbsoluteURL(link))
		// <script src="/license-admin/resources/1.0.0/js/jquery-1.12.4.min.js" type="text/javascript"></script>
		// <link href="/license-admin/resources/1.0.0/css/bootstrap.min.css" type="text/css" rel="stylesheet" />
		linkRegexp := regexp.MustCompile(`<.*?="` + regexp.QuoteMeta(link) + `.*?>`)
		css := GetContent(e.Request.AbsoluteURL(link))
		subs := linkRegexp.FindAllSubmatchIndex(page, -1)
		var b bytes.Buffer
		last := 0
		for i, sub := range subs {
			b.Write(page[last:sub[0]])
			if i == 0 {
				b.Write([]byte("<script>\n"))
				b.Write(css)
			}
			last = sub[1]
		}

		if last < len(page) {
			b.Write(page[last:])
		}

		page = b.Bytes()
	})

	// Handle images
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		log.Printf("src: %s", link)
		log.Printf("url: %s", e.Request.AbsoluteURL(link))
		// <img src='/license-admin/resources/1.0.0/img/license/'>
		linkRegexp := regexp.MustCompile(`<.*?="` + regexp.QuoteMeta(link) + `.*?>`)
		css := GetContent(e.Request.AbsoluteURL(link))
		subs := linkRegexp.FindAllSubmatchIndex(page, -1)
		var b bytes.Buffer
		last := 0
		for i, sub := range subs {
			b.Write(page[last:sub[0]])
			if i == 0 {
				b.Write([]byte("<script>\n"))
				b.Write(css)
			}
			last = sub[1]
		}

		if last < len(page) {
			b.Write(page[last:])
		}

		page = b.Bytes()
	})

	// visit the url of the web to save most of the stuff
	c.Visit(os.Args[1])
}

func GetContent(url string) []byte {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var data bytes.Buffer
	io.Copy(&data, resp.Body)
	return data.Bytes()
}
