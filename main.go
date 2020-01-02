package main

import "time"

func main() {
	relwarc := NewRelwarc()
	defer relwarc.Close()

	browser1, tab1 := relwarc.NewBrowserAndTab()

	_ = browser1

	tab1.Navigate("https://www.baidu.com")

	time.Sleep(time.Minute)
}
