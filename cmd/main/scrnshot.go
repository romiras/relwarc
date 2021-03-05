package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/movsb/relwarc"
)

func main() {
	var u string
	args := os.Args
	if len(args) == 1 {
		// os.Exit(1)
		u = "https://finance.yahoo.com/quote/AAPL/"
	} else {
		u = args[1]
	}

	crawler := relwarc.NewRelwarc()
	defer crawler.Close()

	browser1 := crawler.NewBrowser()

	tab1 := browser1.NewTab()

	if err := tab1.Navigate(&page.NavigateParams{
		URL: u,
	}, true); err != nil {
		panic(err)
	}

	png, err := tab1.CaptureScreenshot(nil)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("screenshot.png", png, 0666); err != nil {
		panic(err)
	}

	title, err := tab1.Title()
	if err != nil {
		panic(title)
	}
	fmt.Println(title)

	tab1.Close()
}
