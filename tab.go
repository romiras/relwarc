package relwarc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
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

	// MainDocumentLoaded is the first request sent.
	mainDocumentLoaded chan *Request

	MainRequest *Request
}

// Close closes this tab.
func (t *Tab) Close() {
	t.cancel()
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
func (t *Tab) Navigate(request *page.NavigateParams, waitMainRequest bool) error {
	t.withLockedRequests(func() {
		t.requests = t.requests[:0]
		t.requestMap = make(map[network.RequestID]*Request)
	})

	t.mainDocumentLoaded = make(chan *Request, 1)

	if request == nil {
		request = &page.NavigateParams{
			URL: "about:blank",
		}
	}

	if err := t.do(chromedp.Navigate(request.URL)); err != nil {
		return err
	}

	if waitMainRequest {
		t.MainRequest = <-t.mainDocumentLoaded
	}

	return nil
}

// NavigateBack navigates the current frame backwards in its history.
func (t *Tab) NavigateBack() error {
	return chromedp.NavigateBack().Do(t.executor())
}

// NavigateForward navigates the current frame forwards in its history.
func (t *Tab) NavigateForward() error {
	return chromedp.NavigateForward().Do(t.executor())
}

// Reload reloads the current page.
func (t *Tab) Reload() error {
	return chromedp.Reload().Do(t.executor())
}

// Stop stops all navigation and pending resource retrieval.
func (t *Tab) Stop() error {
	return chromedp.Stop().Do(t.executor())
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

// Evaluate is an action to evaluate the Javascript expression, unmarshaling
// the result of the script evaluation to out.
//
// When out is a type other than *[]byte, or **chromedp/cdproto/runtime.RemoteObject,
// then the result of the script evaluation will be returned "by value" (ie,
// JSON-encoded), and subsequently an attempt will be made to json.Unmarshal
// the script result to out.
//
// Otherwise, when out is a *[]byte, the raw JSON-encoded value of the script
// result will be placed in out. Similarly, if out is a *runtime.RemoteObject,
// then out will be set to the low-level protocol type, and no attempt will be
// made to convert the result.
//
// If out is nil, the return value is ignored.
//
// Note: any exception encountered will be returned as an error.
func (t *Tab) Evaluate(request *runtime.EvaluateParams, out interface{}) error {
	if request == nil {
		request = &runtime.EvaluateParams{}
	}
	if out == nil {
		out = &out
	}

	switch out.(type) {
	case **runtime.RemoteObject:
	default:
		request.ReturnByValue = true
	}

	object, exception, err := request.Do(t.executor())
	if err != nil {
		return err
	}
	if exception != nil {
		return exception
	}

	switch x := out.(type) {
	case **runtime.RemoteObject:
		*x = object
		return nil
	case *[]byte:
		*x = []byte(object.Value)
		return nil
	}

	if object.Type == "undefined" {
		// The unmarshal above would fail with the cryptic
		// "unexpected end of JSON input" error, so try to give
		// a better one here.
		return fmt.Errorf("encountered an undefined value")
	}

	// unmarshal
	return json.Unmarshal(object.Value, out)
}

// EvaluateAsDevTools is an action that evaluates a Javascript expression as
// Chrome DevTools would, evaluating the expression in the "console" context,
// and making the Command Line API available to the script.
//
// See Evaluate for more information on how script expressions are evaluated.
//
// Note: this should not be used with untrusted Javascript.
func (t *Tab) EvaluateAsDevTools(request *runtime.EvaluateParams, out interface{}) error {
	if request == nil {
		request = &runtime.EvaluateParams{}
	}
	request.ObjectGroup = "console"
	request.IncludeCommandLineAPI = true
	return t.Evaluate(request, out)
}

func (t *Tab) onTargetEvent(ev interface{}) {
	switch ev.(type) {
	case *target.EventTargetInfoChanged:
		return
	case *cdproto.Message:
		return
	case *target.EventTargetDestroyed:
		return
	}
	switch event := ev.(type) {
	case *network.EventRequestWillBeSent:
		request := &Request{
			Request: event,
		}
		t.withLockedRequests(func() {
			// for redirected requests, they have the same request id.
			// here we chain them. But the request maps is updated to the latest
			// request in order to receive response events correctly.
			if parent, ok := t.requestMap[event.RequestID]; ok {
				request.parent = parent
			}
			t.requestMap[event.RequestID] = request
			t.requests = append(t.requests, request)
		})
	case *network.EventResponseReceived:
		t.withLockedRequests(func() {
			if request, ok := t.requestMap[event.RequestID]; ok {
				request.Response = event
			}

			if len(t.requests) == 1 {
				if event.RequestID != t.requests[0].Request.RequestID {
					panic("received: found no match request")
				}
				t.mainDocumentLoaded <- t.requests[0]
				close(t.mainDocumentLoaded)
			}
		})
	case *network.EventDataReceived:
		t.withLockedRequests(func() {
			if request, ok := t.requestMap[event.RequestID]; ok {
				request.DataReceived = event
			}
		})
	case *network.EventLoadingFinished:
		t.withLockedRequests(func() {
			if request, ok := t.requestMap[event.RequestID]; ok {
				request.Finished = event
			}
		})
	case *network.EventLoadingFailed:
		t.withLockedRequests(func() {
			// Some requests like network failed don't have a request sent,
			// but response events received. Before record what we received,
			// do check if the request exists.
			if request, ok := t.requestMap[event.RequestID]; ok {
				request.Failed = event
			}

			// check if it is the docuemnt request (the first one).
			if len(t.requests) == 1 {
				if event.RequestID != t.requests[0].Request.RequestID {
					panic("failed: found no match request")
				}
				t.mainDocumentLoaded <- t.requests[0]
				close(t.mainDocumentLoaded)
			}
		})
	}
}
