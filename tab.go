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

// Location is an action that retrieves the document location.
func (t *Tab) Location() (location string, err error) {
	err = chromedp.EvaluateAsDevTools(`document.location.toString()`, &location).Do(t.executor())
	return
}

// Title is an action that retrieves the document title.
func (t *Tab) Title() (title string, err error) {
	err = chromedp.EvaluateAsDevTools(`document.title`, &title).Do(t.executor())
	return
}

// WaitReady is an element query action that waits until the element matching
// the selector is ready (ie, has been "loaded").
func (t *Tab) WaitReady(sel interface{}, opts ...chromedp.QueryOption) error {
	return chromedp.WaitReady(sel, opts...).Do(t.executor())
}

// EvaluateAsDevTools is an action that evaluates a Javascript expression as
// Chrome DevTools would, evaluating the expression in the "console" context,
// and making the Command Line API available to the script.
//
// See Evaluate for more information on how script expressions are evaluated.
//
// Note: this should not be used with untrusted Javascript.
func (t *Tab) EvaluateAsDevTools(expression string, res interface{}, opts ...chromedp.EvaluateOption) error {
	return chromedp.EvaluateAsDevTools(expression, res, opts...).Do(t.executor())
}
