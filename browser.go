package relwarc

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// Browser is a browser instance.
type Browser struct {
	ctx   context.Context
	first *Tab
	tabs  map[target.ID]*Tab
}

// NewTab opens a new tab.
func (b *Browser) NewTab() *Tab {
	ctx, cancel := chromedp.NewContext(b.ctx)

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

	b.tabs[tgt.TargetID] = &tab

	chromedp.ListenTarget(ctx, tab.onTargetEvent)

	return &tab
}

// Close ...
func (b *Browser) Close() {
	for _, tab := range b.tabs {
		tab.Close()
	}
	b.first.Close()
}
