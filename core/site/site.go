package site

import (
	"fmt"
	"net/url"
	"time"

	"github.com/MikeFilimonov/go-pwc/core/links"
	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

const initialLevel = 0

type Crawler struct {
	buffer   pages.Buffer
	maxDepth int
	startURL string
	seenSet  pages.PagesSeen
	stats    pages.Stats
	fetcher  web.Fetcher
}

func Crawl(input string, maxDepth int) (*pages.Stats, error) {

	crawler := Crawler{
		buffer:   pages.Buffer{},
		maxDepth: maxDepth,
		startURL: input,
		seenSet:  pages.PagesSeen{},
		stats:    pages.Stats{},
		fetcher:  web.Fetcher{},
	}

	return crawler.crawl()

}

func (crawler *Crawler) crawl() (*pages.Stats, error) {

	crawler.handleURL(crawler.startURL, crawler.startURL, initialLevel)
	for len(crawler.buffer.Items) > 0 {

		item, success := crawler.buffer.Dequeue()
		if !success {
			continue
		}
		page, err := crawler.fetcher.Fetch(item.URL)
		if err != nil {
			continue
		}

		pageURLValue := page.URL

		isFromSameZone, err := links.IsFromSameZone(crawler.startURL, pageURLValue)
		if err != nil || !isFromSameZone {
			continue
		}

		pageURL, err := url.Parse(pageURLValue)
		if err != nil {
			continue
		}

		normalizedPageURL := links.Normalize(pageURL)

		isDuplicate := item.URL != normalizedPageURL &&
			crawler.seenSet.Has(normalizedPageURL)
		if isDuplicate {
			continue
		}

		crawler.seenSet.Add(normalizedPageURL)

		crawler.stats.PageCount++
		crawler.stats.TotalBytes += len(page.Body)
		fmt.Printf("[%v]: url:%s, level: %v \n", time.Now(), pageURLValue, item.Level)

		if item.Level < crawler.maxDepth {
			hrefs, err := pages.Parse(page.Body)

			if err != nil {
				continue
			}

			page = nil

			for _, element := range hrefs {
				crawler.handleURL(pageURLValue, element, item.Level)
			}

		}

	}

	crawler.stats.LinksFound = crawler.seenSet.Len()

	return &crawler.stats, nil

}

// handleURL resolves, filters and enqueues one link found on a page
func (c *Crawler) handleURL(baseURL, href string, parentLevel int) {

	completed, err := links.Complete(baseURL, href)
	if err != nil {
		return
	}

	normalized := links.Normalize(completed)
	if len(normalized) == 0 {
		return
	}
	isFromSameDomain, err := links.IsFromSameZone(c.startURL, normalized)
	if err != nil || !isFromSameDomain || c.seenSet.Has(normalized) {
		return
	}

	c.seenSet.Add(normalized)
	item := pages.Item{URL: normalized, Level: parentLevel + 1}
	c.buffer.Enqueue(item)

}
