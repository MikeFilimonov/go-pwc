package site

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
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
		"http://it.works/main/news.html": `<p>
					<a class="c" href="news/economy.html">Economy</a>
					<a class="c" href="news/science.html">Science</a>
					<a class="c" href="/home.html">Home</a>
			</p>`,
		"http://it.works/main/about.html": `<p>
			<a class="c" href="/main/about.html">About</a>
			<a class="c" href="/home.html">Home</a>
			</p>`,
		"http://it.works/main/news/science.html": `<p><a class="c" href="/main/news/science.html">Science</a></p>`,
		"http://it.works/main/news/economy.html": `<p><a class="c" href="/main/news/economy.html">Economy</a></p>`,
		"http://it.works/home.html": `<p>
			<a class="c" href="/home.html">Home</a>
			<a class="c" href="/main"> Main </a>		
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
		{"one_level_deep", 1, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 0, PageCount: 0, TotalBytes: 0},
			[]string{"/main"}},
		{"a_level_deeper", 2, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 0, PageCount: 0, TotalBytes: 0},
			[]string{
				"/main",
				"/main/news.html",
				"/main/about.html",
			}},
		{"level_3", 3, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 0, PageCount: 0, TotalBytes: 0},
			[]string{
				"/main",
				"/main/news.html",
				"/main/about.html",
				"/main/news/science.html",
				"/main/news/economy.html",
				"/home.html",
			}},
		{"crosslinks", 4, "http://it.works/main",
			metasite, pages.Stats{LinksFound: 0, PageCount: 0, TotalBytes: 0},
			[]string{
				"/main",
				"/main/news.html",
				"/main/about.html",
				"/main/news/science.html",
				"/main/news/economy.html",
				"/home.html",
			}},
	}

	for _, v := range cases {

		t.Run(v.name, func(t *testing.T) {

			seenSet := make(map[string]struct{})
			input := v.input
			maxDepth := v.maxDepth

			v.stats.TotalBytes = totalBytes(t, v.expectedUrls, input, v.metasite)
			v.stats.LinksFound = len(v.expectedUrls)
			v.stats.PageCount = len(v.expectedUrls)
			handler := handler(t, input, v.expectedUrls, v.metasite, seenSet)

			srv := httptest.NewServer(http.HandlerFunc(handler))
			defer srv.Close()
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
				maxDepth: maxDepth,
				startURL: input,
				seenSet:  pages.PagesSeen{},
				stats:    pages.Stats{},
				fetcher:  web.Fetcher{},
			}

			crawler.fetcher.Init(config)

			results, err := crawler.crawl()

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				detectCrawlErrors(t, results, v.stats)
			}

		})

	}

}

func totalBytes(t *testing.T, expectedUrls []string, input string, metasite map[string]string) int {

	t.Helper()

	result := 0
	for _, u := range expectedUrls {

		url, err := links.Complete(input, u)
		if err != nil {
			t.Fatalf("unable to complete the url: %v (%v)", u, err)
		}

		value, found := metasite[url.String()]
		if found {
			result += len([]byte(value))
		} else {
			t.Errorf("missing page contents for url \"%s\"", value)
		}

	}

	return result

}

func handler(t *testing.T, input string, expectedUrls []string, metasite map[string]string, seenSet map[string]struct{}) func(http.ResponseWriter, *http.Request) {

	var seenSetLock sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {

		if !slices.Contains(expectedUrls, r.URL.Path) {
			t.Errorf("unexpected request path: %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		url, err := links.Complete(input, r.URL.Path)
		if err != nil {
			t.Errorf("unable to complete full url from the path: %v", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		fullUrl := url.String()

		seenSetLock.Lock()
		_, isDuplicate := seenSet[fullUrl]
		seenSet[fullUrl] = struct{}{}
		seenSetLock.Unlock()

		if isDuplicate {
			t.Errorf("duplicate path detected: %v", r.URL.Path)
		}

		contents, found := metasite[fullUrl]
		if !found {
			t.Errorf("missing contents for %v", fullUrl)
			http.NotFound(w, r)
			return
		}

		w.Header().Add(web.ContentTypeFieldName, web.HtmlHeaderValue)
		w.Write([]byte(contents))

	}

}

func detectCrawlErrors(t *testing.T, results *pages.Stats, expectedMetrics pages.Stats) {

	t.Helper()

	if results.PageCount != expectedMetrics.PageCount {
		t.Errorf("unexpected page count: expected %v, discovered %v",
			expectedMetrics.PageCount, results.PageCount)
	}

	if results.TotalBytes != expectedMetrics.TotalBytes {
		t.Errorf("unexpected total size (bytes): expected %v, got %v",
			expectedMetrics.TotalBytes, results.TotalBytes)
	}

	if results.LinksFound != expectedMetrics.LinksFound {
		t.Errorf("unexpected found links count (items): expected %v, got %v",
			expectedMetrics.LinksFound, results.LinksFound)
	}

}
