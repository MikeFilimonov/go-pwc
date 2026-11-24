package site

import (
	"context"
	"sync"
	"time"

	"github.com/MikeFilimonov/go-pwc/core/links"
	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

const (
	initialLevel            = 0
	workerLimit             = 9
	gracefulShutdownTimeout = 5
)

type Crawler struct {
	buffer   pages.Buffer
	maxDepth int
	startURL string
	seenSet  pages.PagesSeen
	stats    pages.Stats
	fetcher  web.Fetcher
	mtx      sync.Mutex
}

func Crawl(ctx context.Context, input string, maxDepth int) (*pages.Stats, error) {

	crawler := Crawler{
		buffer:   pages.Buffer{},
		maxDepth: maxDepth,
		startURL: input,
		seenSet:  pages.PagesSeen{},
		stats:    pages.Stats{},
		fetcher:  web.Fetcher{},
	}

	return crawler.crawl(ctx)

}

func (crawler *Crawler) crawl(ctx context.Context) (*pages.Stats, error) {

	completed := make(chan (struct{}))
	graceDuration := time.Second * gracefulShutdownTimeout

	requestCtx := context.WithoutCancel(ctx)
	cancellationContext, cancelFunc := context.WithCancel(requestCtx)
	defer cancelFunc()
	source := make(chan (pages.Item))
	var workerTracker sync.WaitGroup
	var jobs sync.WaitGroup

	crawler.handleURL(crawler.startURL, crawler.startURL, initialLevel)
	crawler.fetcher.Init(nil)

	workerTracker.Go(func() {
		watcher(ctx, completed, graceDuration, cancelFunc)
	})

	for range workerLimit {

		workerTracker.Go(func() {

			for page := range source {
				crawler.processPage(cancellationContext, &page)
				jobs.Done()
			}
		})

	}

GracefulShutdown:

	for {

		crawler.mtx.Lock()
		itemsToHandle := len(crawler.buffer.Items)
		crawler.mtx.Unlock()
		if itemsToHandle == 0 {
			break
		}

		for range itemsToHandle {
			crawler.mtx.Lock()
			page, success := crawler.buffer.Dequeue()
			crawler.mtx.Unlock()
			if success {
				jobs.Add(1)
				select {
				case source <- page:
				case <-ctx.Done():
					jobs.Done()
					break GracefulShutdown
				}

			}
		}

		jobs.Wait()
	}
	close(source)
	jobs.Wait()
	close(completed)
	workerTracker.Wait()

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
	if err != nil || !isFromSameDomain {
		return
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.seenSet.Has(normalized) {
		return
	}

	c.seenSet.Add(normalized)
	item := pages.Item{URL: normalized, Level: parentLevel + 1}
	c.buffer.Enqueue(item)

}

func watcher(stopContext context.Context, completed <-chan struct{}, graceTimeout time.Duration, cancelCallback context.CancelFunc) {

	select {
	case <-completed:
		return
		//no default needed, waiting is intentional
	case <-stopContext.Done():
	}

	timer := time.NewTimer(graceTimeout)
	defer timer.Stop()

	select {
	case <-completed:
		return
	case <-timer.C:
		cancelCallback()
	}

}
