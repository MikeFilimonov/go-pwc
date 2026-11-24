package site

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

const timeout = 5 * time.Second

func TestCrawlConcurrentRequests(t *testing.T) {

	metasite := map[string]string{
		"http://it.works/main": `<p>
			<a class="c" href="/main/a.html">A</a>
			<a class="c" href="/main/b.html">B</a>	
		</p>`,
		"http://it.works/main/a.html": `<p>
			<a class="c" href="http://it.works/main/a.html">A</a>
		</p>`,
		"http://it.works/main/b.html": `<p>
			<a class="c" href="http://it.works/main/b.html">B</a>	
		</p>`,
	}

	cases := []struct {
		name         string
		maxDepth     int
		input        string
		metasite     map[string]string
		stats        pages.Stats
		expectedUrls []string
	}{
		{"one_level_deep", 2, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 0, PageCount: 0, TotalBytes: 0},
			[]string{
				"/main",
				"/main/a.html",
				"/main/b.html"}},
	}

	for _, item := range cases {

		t.Run(item.name, func(t *testing.T) {

			seenSet := make(map[string]struct{})
			release := make(chan struct{})
			arrivals := make(chan string, 2)
			crawlery := make(chan struct{})
			stats := &pages.Stats{}
			var failure error

			item.stats.PageCount = len(item.expectedUrls)
			item.stats.LinksFound = len(item.expectedUrls)
			item.stats.TotalBytes = totalBytes(t, item.expectedUrls, item.input, item.metasite)

			pendingChildren, children := make(map[string]struct{}), make(map[string]struct{})

			parsedInput, err := url.Parse(item.input)
			if err != nil {
				t.Fatalf("error parsing root url(%s): %v", item.input, err)
			}
			for _, path := range item.expectedUrls {
				if path != parsedInput.Path {
					pendingChildren[path] = struct{}{}
					children[path] = struct{}{}
				}
			}

			var singletone sync.Once

			handler := handler(t, item.input, item.expectedUrls, item.metasite, seenSet)

			wrapper := func(w http.ResponseWriter, r *http.Request) {

				// handling children only
				if _, found := children[r.URL.Path]; found {

					select {
					case arrivals <- r.URL.Path:
					case <-release:
						//Allow cleanup to interrupt a blocked arrival notification
					}
					<-release
				}

				handler(w, r)
			}

			srv := httptest.NewServer(http.HandlerFunc(wrapper))
			defer func() {
				releaseDescendants(&singletone, release)
				srv.Close()
				select {
				case <-crawlery:
				case <-time.After(timeout):
					t.Fatalf("crawl failed to complete in %v", timeout)
				}
			}()
			destination := srv.Listener.Addr().String()

			// no ip support results into manual mapping for pseudo-dns
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
				maxDepth: item.maxDepth,
				startURL: item.input,
				seenSet:  pages.PagesSeen{},
				stats:    pages.Stats{},
				fetcher:  web.Fetcher{},
			}

			crawler.fetcher.Init(config)

			timer := time.After(timeout)

			go func() {
				stats, failure = crawler.crawl()
				close(crawlery)
			}()

			for len(pendingChildren) > 0 {
				select {
				case path := <-arrivals:
					delete(pendingChildren, path)
				case <-timer:
					t.Fatalf("children requests failed to arrive in %v", timeout)

				}
			}

			releaseDescendants(&singletone, release)

			select {
			case <-crawlery:

				if failure != nil {
					t.Fatalf("unexpected error: %v", failure)
				}
				detectCrawlErrors(t, stats, item.stats)

			case <-time.After(timeout):
				t.Fatalf("crawl failed to finish after releasing handlers during %v", timeout)

			}

		})

	}

}

func releaseDescendants(singletone *sync.Once, output chan<- struct{}) {

	singletone.Do(func() {
		close(output)
	})
}
