package main

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Tab is a browser tab.
type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc

	// The DevTool.
	target *chromedp.Target

	// The network requests.
	requests     []*Request
	requestMap   map[network.RequestID]*Request
	lockRequests sync.Mutex

	// If false, each time we navigate an URL,
	// the requests will be cleared.
	PreserveRequests bool
}

// Request represents a network request and its response statuses.
type Request struct {
	Request      *network.EventRequestWillBeSent
	Response     *network.EventResponseReceived
	Finished     *network.EventLoadingFinished
	Failed       *network.EventLoadingFailed
	DataReceived *network.EventDataReceived
	parent       *Request
}

func (t *Tab) executor() context.Context {
	return cdp.WithExecutor(t.ctx, t.target)
}

func (t *Tab) withLockedRequests(fn func()) {
	t.lockRequests.Lock()
	defer t.lockRequests.Unlock()
	fn()
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
	if !t.PreserveRequests {
		t.withLockedRequests(func() {
			t.requests = t.requests[:0]
			t.requestMap = make(map[network.RequestID]*Request)
		})
	}
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

func (t *Tab) onTargetEvent(ev interface{}) {
	fmt.Println(reflect.TypeOf(ev))
	switch event := ev.(type) {
	case *network.EventRequestWillBeSent:
		request := &Request{
			Request: event,
		}
		t.withLockedRequests(func() {
			// for redirected requests, they have the same request id.
			// here we chain them. But the request maps is updated to the latest
			// request in order to receive response events.
			if parent, ok := t.requestMap[event.RequestID]; ok {
				request.parent = parent
				t.requestMap[event.RequestID] = request
				t.requests = append(t.requests, request)
			} else {
				t.requestMap[event.RequestID] = request
				t.requests = append(t.requests, request)
			}
		})
	case *network.EventResponseReceived:
		t.withLockedRequests(func() {
			t.requestMap[event.RequestID].Response = event
		})
	case *network.EventLoadingFinished:
		t.withLockedRequests(func() {
			t.requestMap[event.RequestID].Finished = event
		})
	case *network.EventLoadingFailed:
		t.withLockedRequests(func() {
			t.requestMap[event.RequestID].Failed = event
		})
	case *network.EventDataReceived:
		t.withLockedRequests(func() {
			t.requestMap[event.RequestID].DataReceived = event
		})
	}
}
