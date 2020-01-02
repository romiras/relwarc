package main

import "time"

func main() {
	relwarc := NewRelwarc()
	defer relwarc.Close()

	browser1, tab1 := relwarc.NewBrowserAndTab()
	browser2, tab2 := relwarc.NewBrowserAndTab()

	_, _, _, _ = browser1, browser2, tab1, tab2

	time.Sleep(time.Minute)
}
