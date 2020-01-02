package main

import (
	"fmt"
	"io/ioutil"

	"github.com/chromedp/cdproto/page"
)

func main() {
	relwarc := NewRelwarc()
	defer relwarc.Close()

	browser1, tab1 := relwarc.NewBrowserAndTab()

	_ = browser1

	tab1.Navigate(&page.NavigateParams{
		URL: "https://www.example.com",
	})

	png, err := tab1.CaptureScreenshot(nil)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("screenshot.png", png, 0666); err != nil {
		panic(err)
	}

	location, err := tab1.Location()
	if err != nil {
		panic(err)
	}
	fmt.Println(location)

	title, err := tab1.Title()
	if err != nil {
		panic(title)
	}
	fmt.Println(title)
}
