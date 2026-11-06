package site

import (
	"fmt"
	"net/url"
	"sync"
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

	crawler.fetcher.Init(nil)

	jobs := make(chan pages.Item)
	results := make(chan result)

	var waitGroup sync.WaitGroup

	for range workerPoolCount {
		waitGroup.Go(func() {
			worker(crawler.fetcher, results, jobs)
		})
	}

	crawler.handleURL(crawler.startURL, crawler.startURL, initialLevel)

	jobsInFlight := 0

	for len(crawler.buffer.Items) > 0 || jobsInFlight > 0 {

		item, success := crawler.buffer.Dequeue()
		if !success {
			continue
		}
		page, err := crawler.fetcher.Fetch(item.URL)
		crawler.processResult(result{item, page, err})

	}
	close(jobs)
	waitGroup.Wait()
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

func (c *Crawler) processResult(outcome result) {

	if outcome.error != nil {

		fmt.Printf("[%v]: fetch failed: url: %s, error: %v\n",
			time.Now(), outcome.job.URL, outcome.error)
		return
	}

	pageURLValue := outcome.result.URL

	isFromSameZone, err := links.IsFromSameZone(c.startURL, pageURLValue)
	if err != nil || !isFromSameZone {
		return
	}

	pageURL, err := url.Parse(pageURLValue)
	if err != nil {
		return
	}

	normalizedPageURL := links.Normalize(pageURL)

	isDuplicate := outcome.job.URL != normalizedPageURL &&
		c.seenSet.Has(normalizedPageURL)
	if isDuplicate {
		return
	}

	c.seenSet.Add(normalizedPageURL)

	c.stats.PageCount++
	c.stats.TotalBytes += len(outcome.result.Body)
	fmt.Printf("[%v]: url:%s, level: %v \n", time.Now(), pageURLValue,
		outcome.job.Level)

	if outcome.job.Level < c.maxDepth {
		hrefs, err := pages.Parse(outcome.result.Body)

		if err != nil {
			return
		}

		outcome.result = nil

		for _, element := range hrefs {
			c.handleURL(pageURLValue, element, outcome.job.Level)
		}

	}

}
