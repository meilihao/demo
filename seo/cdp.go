// 没有使用start更多的chromedp是因为[没有seo的例子](https://github.com/chromedp/examples/issues/13)
// 参考了[rendora](github.com/rendora/rendora)的实现和[React 服务端渲染方案完美的解决方案](https://juejin.im/post/5bf3cb59f265da612b1336e2)思路
// 还是有报错, 原因不明
// 不推荐使用:  Chrome DevTools Protocol的api难用
// 推荐使用[rendertron](https://github.com/GoogleChrome/rendertron)或[puppeteer](https://github.com/GoogleChrome/puppeteer),两者均由Google Chrome 官方团队提供.
package main

import (
	"context"
	"fmt"
	"time"
	"net/http"
	"html/template"
	"encoding/json"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/rpcc"
)

var c *cdp.Client

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, template.HTML(getHTML()))
}

func main() {
	generateClient()

	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe(":8000", nil)
}

func generateClient() {
	ctx := context.Background()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			panic(err)
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		panic(err)
	}
	//defer conn.Close() // Leaving connections open will leak memory.

	c = cdp.NewClient(conn)

	// Open a DOMContentEventFired client to buffer this event.
	//domContent, err := c.Page.DOMContentEventFired(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = c.Page.Enable(ctx); err != nil {
		panic(err)
	}

	err = c.Network.Enable(ctx, nil)
	if err != nil {
		panic(err)
	}

	headers := map[string]string{
		"X-Rendora-Type": "RENDER",
	}

	headersStr, err := json.Marshal(headers)
	if err != nil {
		panic(err)
	}

	err = c.Network.SetExtraHTTPHeaders(ctx, network.NewSetExtraHTTPHeadersArgs(headersStr))
	if err != nil {
		panic(err)
	}

	blockedURLs := network.NewSetBlockedURLsArgs([]string{
		"*.png", "*.jpg", "*.jpeg", "*.webp", "*.gif", "*.css", "*.woff2", "*.svg", "*.woff", "*.ttf", "*.ico",
		"https://www.youtube.com/*", "https://www.google-analytics.com/*",
		"https://fonts.googleapis.com/*",
	})

	err = c.Network.SetBlockedURLs(ctx, blockedURLs)
	if err != nil {
		panic(err)
	}
}

func getHTML() string {
	fmt.Println("---start")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	networkResponse, err := c.Network.ResponseReceived(ctx)
	if err != nil {
		panic(err)
	}

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs("https://justineo.github.io/vue-awesome/demo/")
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		panic(err)
	}

	responseReply, err := networkResponse.Recv()
	if err != nil {
		panic(err)
	}
	fmt.Println(responseReply)

	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		panic(err)
	}
	defer domContent.Close()

	waitUntil := 0
	if waitUntil > 0 {
		time.Sleep(time.Duration(waitUntil) * time.Millisecond)
	}


	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		panic(err)
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		panic(err)
	}

	// Get the outer HTML for the page.
	result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("HTML: %s\n", result.OuterHTML)

	return result.OuterHTML
}