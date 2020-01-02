package main

import (
	"context"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Tab is a browser tab.
type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc
	target *chromedp.Target
}

func (t *Tab) executor() context.Context {
	return cdp.WithExecutor(t.ctx, t.target)
}

func (t *Tab) do(actions ...chromedp.Action) error {
	for _, action := range actions {
		if err := action.Do(t.executor()); err != nil {
			return err
		}
	}
	return nil
}

// Navigate navigates current page to the given URL.
//
// TODO arguments but URL are ignored.
func (t *Tab) Navigate(request *page.NavigateParams) error {
	if request == nil {
		request = &page.NavigateParams{
			URL: "about:blank",
		}
	}
	return t.do(chromedp.Navigate(request.URL))
}

// CaptureScreenshot capture page screenshot.
func (t *Tab) CaptureScreenshot(request *page.CaptureScreenshotParams) ([]byte, error) {
	if request == nil {
		request = &page.CaptureScreenshotParams{}
	}
	return request.Do(t.executor())
}
