package site

import (
	"github.com/MikeFilimonov/go-pwc/core/pages"
	"github.com/MikeFilimonov/go-pwc/core/web"
)

const (
	workerPoolCount = 2
)

type result struct {
	job    pages.Item
	result *web.Page
	error  error
}

func worker(fetcher web.Fetcher, results chan<- result, jobs <-chan pages.Item) {

	for item := range jobs {

		page, err := fetcher.Fetch(item.URL)

		results <- result{
			job:    item,
			result: page,
			error:  err,
		}

	}

}
