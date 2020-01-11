package main

import (
	"github.com/chromedp/cdproto/page"
	"github.com/movsb/relwarc"
)

func main() {
	crawler := relwarc.NewRelwarc()
	defer crawler.Close()

	browser1 := crawler.NewBrowser()

	tab1 := browser1.NewTab()
	tab2 := browser1.NewTab()

	if err := tab1.Navigate(&page.NavigateParams{
		// URL: "https://www.ex111ample.com",
		URL: "https://blog.twofei.com",
		// URL: "https://blog.twofei.com/style.css",
	}, true); err != nil {
		panic(err)
	}

	if err := tab2.Navigate(&page.NavigateParams{
		// URL: "https://www.ex111ample.com",
		URL: "https://blog.twofei.com",
		// URL: "https://blog.twofei.com/style.css",
	}, true); err != nil {
		panic(err)
	}

	tab1.Close()
	tab2.Close()
}
