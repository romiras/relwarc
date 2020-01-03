package main

import (
	"fmt"
	"reflect"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
)

func main() {
	relwarc := NewRelwarc()
	defer relwarc.Close()

	browser1, tab1 := relwarc.NewBrowserAndTab()

	_ = browser1

	tab1.Navigate(&page.NavigateParams{
		URL: "https://www.example.com",
	})

	var out interface{}

	tab1.EvaluateAsDevTools(&runtime.EvaluateParams{
		Expression: "window.close()",
	}, &out)

	fmt.Println(out, reflect.TypeOf(out))

	//time.Sleep(time.Hour)

	tab1.cancel()
}
