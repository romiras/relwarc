package relwarc

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// Relwarc manages browsers.
type Relwarc struct {
	// The root context that holds the allocator.
	ctx context.Context

	// The canceller to cancel an allocator.
	//
	// Cancel an allocator will deallocate/destroy all browsers
	// allocated by this allocator.
	cancel context.CancelFunc
}

// NewRelwarc creates a new relwarc.
func NewRelwarc() *Relwarc {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), defaultExecAllocatorOptions...)
	return &Relwarc{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Close ...
func (r *Relwarc) Close() {
	r.cancel()
}

// NewBrowser creates a new browser.
func (r *Relwarc) NewBrowser() *Browser {
	ctx, cancel := chromedp.NewContext(r.ctx)

	// make sure a browser and its first tab are created.
	if err := chromedp.Run(ctx); err != nil {
		panic(err)
	}

	// enable network by default.
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		panic(err)
	}

	tgt := chromedp.FromContext(ctx).Target

	tab := Tab{
		ctx:        ctx,
		cancel:     cancel,
		target:     tgt,
		requestMap: map[network.RequestID]*Request{},
	}

	chromedp.ListenTarget(ctx, tab.onTargetEvent)

	browser := Browser{
		ctx:   ctx,
		first: &tab,
		tabs:  map[target.ID]*Tab{},
	}

	return &browser
}

var defaultExecAllocatorOptions = []chromedp.ExecAllocatorOption{
	//chromedp.Headless,
	chromedp.NoFirstRun,
	chromedp.NoDefaultBrowserCheck,
	chromedp.Flag("disable-background-networking", true),
	chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
	chromedp.Flag("disable-background-timer-throttling", true),
	chromedp.Flag("disable-backgrounding-occluded-windows", true),
	chromedp.Flag("disable-breakpad", true),
	chromedp.Flag("disable-client-side-phishing-detection", true),
	chromedp.Flag("disable-default-apps", true),
	chromedp.Flag("disable-dev-shm-usage", true),
	chromedp.Flag("disable-extensions", true),
	chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
	chromedp.Flag("disable-hang-monitor", true),
	chromedp.Flag("disable-ipc-flooding-protection", true),
	chromedp.Flag("disable-popup-blocking", true),
	chromedp.Flag("disable-prompt-on-repost", true),
	chromedp.Flag("disable-renderer-backgrounding", true),
	chromedp.Flag("disable-sync", true),
	chromedp.Flag("force-color-profile", "srgb"),
	chromedp.Flag("metrics-recording-only", true),
	chromedp.Flag("safebrowsing-disable-auto-update", true),
	chromedp.Flag("enable-automation", true),
	chromedp.Flag("password-store", "basic"),
	chromedp.Flag("use-mock-keychain", true),
	chromedp.Flag("blink-settings", "imagesEnabled=false"),
}
