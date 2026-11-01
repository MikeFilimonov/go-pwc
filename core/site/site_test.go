package site

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

func TestCrawl(t *testing.T) {

	input := "http://it.works/about.html"
	contents := `<p><a class="c" href="/y">Y</a></p>`
	maxDepth := 1

	expectedTotalBytes := len([]byte(contents))
	expectedPageCount := 1
	expectedLinksFound := 1 // only the seed enters the seen-set
	expectedURL := "/about.html"

	handler := func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/about.html" {
			t.Errorf("unexpected request path: expected %q, got %q",
				expectedURL, r.URL.Path)
			http.NotFound(w, r)
			return
		}

		w.Header().Add(web.ContentTypeFieldName, web.HtmlHeaderValue)
		w.Write([]byte(contents))

	}

	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()
	destination := srv.Listener.Addr().String()

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {

			dailer := net.Dialer{}
			return dailer.DialContext(ctx, network, destination)

		},
	}

	defer transport.CloseIdleConnections()

	config := &http.Client{
		Timeout:   web.DefaultFetcherTimeout,
		Transport: transport,
	}

	crawler := Crawler{
		buffer:   pages.Buffer{},
		maxDepth: maxDepth,
		startURL: input,
		seenSet:  pages.PagesSeen{},
		stats:    pages.Stats{},
		fetcher:  web.Fetcher{},
	}

	crawler.fetcher.Init(config)

	pages, err := crawler.crawl()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else {

		if pages.PageCount != expectedPageCount {
			t.Errorf("unexpected page count: expected %v, discovered %v",
				expectedPageCount, pages.PageCount)
		}

		if pages.TotalBytes != expectedTotalBytes {
			t.Errorf("unexpected total size (bytes): expected %v, got %v",
				expectedTotalBytes, pages.TotalBytes)
		}

		if pages.LinksFound != expectedLinksFound {
			t.Errorf("unexpected found links count (items): expected %v, got %v",
				expectedLinksFound, pages.LinksFound)
		}

	}

}
