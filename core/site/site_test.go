package site

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/MikeFilimonov/go-pwc/core/links"
	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

func TestCrawl(t *testing.T) {

	metasite := map[string]string{
		"http://it.works/main": `<p>
			<a class="c" href="/main/news.html">News</a>
			<a class="c" href="/main/about.html">About</a>	
		</p>`,
		"http://it.works/main/news.html": `<p><a class="c" href="/main/news.html">News
				<p>
					<a class="c" href="/main/news/economy.html">Economy</a>
					<a class="c" href="/main/news/science.html">Science</a>
				</p></a>
		</p>`,
		"http://it.works/main/about.html":        `<p><a class="c" href="/main/about.html">About</a></p>`,
		"http://it.works/main/news/science.html": `<p><a class="c" href="/main/news/science.html">Science</a></p>`,
		"http://it.works/main/news/economy.html": `<p><a class="c" href="/main/news/economy.html">Economy</a></p>`,
	}

	cases := []struct {
		name     string
		maxDepth int
		input    string
		metasite map[string]string

		stats        pages.Stats
		expectedUrls []string
	}{
		{"one_level_deep", 1, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 1, PageCount: 1, TotalBytes: 0},
			[]string{"/main"}},
		{"a_level_deeper", 2, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 3, PageCount: 3, TotalBytes: 0},
			[]string{
				"/main",
				"/main/news.html",
				"/main/about.html",
			}},
		{"level_3", 3, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 5, PageCount: 5, TotalBytes: 0},
			[]string{
				"/main",
				"/main/news.html",
				"/main/about.html",
				"/main/news/science.html",
				"/main/news/economy.html",
			}},
	}

	for _, v := range cases {

		t.Run(v.name, func(t *testing.T) {

			input := v.input
			maxDepth := v.maxDepth

			for _, u := range v.expectedUrls {

				url, err := links.Complete(input, u)
				if err == nil {

					value, found := v.metasite[url.String()]
					if found {
						v.stats.TotalBytes += len([]byte(value))
					}

				}

			}

			handler := func(w http.ResponseWriter, r *http.Request) {

				if !slices.Contains(v.expectedUrls, r.URL.Path) {
					t.Errorf("unexpected request path: %q", r.URL.Path)
					http.NotFound(w, r)
					return
				}

				url, err := links.Complete(input, r.URL.Path)
				if err != nil {
					t.Errorf("failed to get the full url from the path: %v", r.URL.Path)
					return
				}
				contents := v.metasite[url.String()]
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

				if pages.PageCount != v.stats.PageCount {
					t.Errorf("unexpected page count: expected %v, discovered %v",
						v.stats.PageCount, pages.PageCount)
				}

				if pages.TotalBytes != v.stats.TotalBytes {
					t.Errorf("unexpected total size (bytes): expected %v, got %v",
						v.stats.TotalBytes, pages.TotalBytes)
				}

				if pages.LinksFound != v.stats.LinksFound {
					t.Errorf("unexpected found links count (items): expected %v, got %v",
						v.stats.LinksFound, pages.LinksFound)
				}

			}

		})

	}

}
