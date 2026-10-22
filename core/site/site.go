package site

import (
	"fmt"
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

	crawler.handleURL(input, input, initialLevel)
	for len(crawler.buffer.Items) > 0 {

		item, success := crawler.buffer.Dequeue()
		if !success {
			continue
		}
		page, err := crawler.fetcher.Fetch(item.URL)
		if err != nil {
			continue
		}
		isFromSameZone, err := links.IsFromSameZone(crawler.startURL, page.URL)
		if !isFromSameZone {
			continue
		}

		crawler.stats.PageCount++
		crawler.stats.TotalBytes += len(page.Body)
		fmt.Printf("[%v]: url:%s, level: %v \n", time.Now(), page.URL, item.Level)

		if item.Level < maxDepth {
			hrefs, err := pages.Parse(page.Body)

			if err != nil {
				continue
			}

			for _, element := range hrefs {
				crawler.handleURL(page.URL, element, item.Level)
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
