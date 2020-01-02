package main

import "github.com/chromedp/cdproto/target"

// Browser is a browser instance.
type Browser struct {
	tabs map[target.ID]*Tab
}
