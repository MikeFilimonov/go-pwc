package site

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MikeFilimonov/go-pwc/core/links"
	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

func BenchmarkCrawl(b *testing.B) {

	metasite := map[string]string{
		"http://it.works/main": `<p>
			<a class="c" href="/main/news.html">News</a>
			<a class="c" href="/main/about.html">About</a>	
		</p>`,
		"http://it.works/main/news.html": `<p>
					<a class="c" href="news/economy.html">Economy</a>
					<a class="c" href="news/science.html">Science</a>
					<a class="c" href="/home.html">Home</a>
			</p>`,
		"http://it.works/main/about.html": `<p>
			<a class="c" href="/main/about.html">About</a>
			<a class="c" href="/home.html">Home</a>
			</p>`}

	setup := struct {
		maxDepth     int
		input        string
		metasite     map[string]string
		stats        pages.Stats
		expectedUrls []string
	}{2, "http://it.works/main",
		metasite, pages.Stats{LinksFound: 0, PageCount: 0, TotalBytes: 0},
		[]string{
			"/main",
			"/main/news.html",
			"/main/about.html",
		}}

	stats := pages.Stats{
		LinksFound: len(setup.expectedUrls),
		TotalBytes: totalBytes(b, setup.expectedUrls, setup.input, metasite),
		PageCount:  len(setup.expectedUrls),
	}

	handler := func(w http.ResponseWriter, r *http.Request) {

		fullURL, err := links.Complete(setup.input, r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		data, found := metasite[fullURL.String()]

		if !found {
			http.NotFound(w, r)
			return
		}

		w.Header().Add(web.ContentTypeFieldName, web.HtmlHeaderValue)
		w.Write([]byte(data))
	}

	srv := httptest.NewServer(http.HandlerFunc(handler))
	destination := srv.Listener.Addr().String()
	// no ip support results into manual mapping for pseudo-dns
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {

			dailer := net.Dialer{}
			return dailer.DialContext(ctx, network, destination)

		},
	}

	b.Cleanup(func() {
		srv.Close()
		transport.CloseIdleConnections()
	})

	config := &http.Client{
		Timeout:   web.DefaultFetcherTimeout,
		Transport: transport,
	}

	b.ReportAllocs()

	for b.Loop() {

		crawler := Crawler{
			buffer:   pages.Buffer{},
			maxDepth: setup.maxDepth,
			startURL: setup.input,
			seenSet:  pages.PagesSeen{},
			stats:    pages.Stats{},
			fetcher:  web.Fetcher{},
		}

		crawler.fetcher.Init(config)

		results, err := crawler.crawl()

		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		} else {
			detectCrawlErrors(b, results, stats)
		}

	}

}
