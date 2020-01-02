package main

import (
	"context"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// Tab is a browser tab.
type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc
	target *chromedp.Target
}

func (t *Tab) do(actions ...chromedp.Action) error {
	ctx := cdp.WithExecutor(t.ctx, t.target)
	for _, action := range actions {
		if err := action.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Navigate navigates to the URL at urlstr.
func (t *Tab) Navigate(urlstr string) error {
	return t.do(chromedp.Navigate(urlstr))
}
