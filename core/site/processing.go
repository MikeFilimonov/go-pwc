package site

import (
	"fmt"
	"net/url"
	"time"

	"github.com/MikeFilimonov/go-pwc/core/links"
	"github.com/MikeFilimonov/go-pwc/core/pages"
)

func (crawler *Crawler) processPage(item *pages.Item) {

	page, err := crawler.fetcher.Fetch(item.URL)
	if err != nil {
		return
	}

	pageURLValue := page.URL

	isFromSameZone, err := links.IsFromSameZone(crawler.startURL, pageURLValue)
	if err != nil || !isFromSameZone {
		return
	}

	pageURL, err := url.Parse(pageURLValue)
	if err != nil {
		return
	}

	normalizedPageURL := links.Normalize(pageURL)

	crawler.mtx.Lock()
	isDuplicate := item.URL != normalizedPageURL &&
		crawler.seenSet.Has(normalizedPageURL)
	if isDuplicate {
		crawler.mtx.Unlock()
		return
	}

	crawler.seenSet.Add(normalizedPageURL)

	crawler.stats.PageCount++
	crawler.stats.TotalBytes += len(page.Body)
	crawler.mtx.Unlock()
	fmt.Printf("[%v]: url:%s, level: %v \n", time.Now(), pageURLValue, item.Level)

	if item.Level < crawler.maxDepth {
		hrefs, err := pages.Parse(page.Body)

		if err != nil {
			return
		}

		page = nil

		for _, element := range hrefs {
			crawler.handleURL(pageURLValue, element, item.Level)
		}

	}

}
