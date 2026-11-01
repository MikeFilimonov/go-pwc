package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultFetcherTimeout = 10 * time.Second
	ContentTypeFieldName  = "Content-Type"
	HtmlHeaderValue       = "text/html"
)

var (
	ErrBadStatus      = errors.New("bad response")
	ErrRequestFailed  = errors.New("request failed")
	ErrUnparsableHTML = errors.New("not a parseble HTML")
)

type Fetcher struct {
	timeout time.Duration
	client  *http.Client
}

type Page struct {
	URL  string
	Body []byte
}

func (f *Fetcher) Init(presets *http.Client) {

	if f.timeout == 0 {
		f.timeout = DefaultFetcherTimeout
	}

	if f.client == nil {
		f.client = &http.Client{Timeout: f.timeout}
	}

	if presets != nil {
		f.client = presets
	}

}

func (f *Fetcher) Fetch(url string) (*Page, error) {

	f.Init(nil)
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrBadStatus, resp.StatusCode)
	}

	hdrValue := resp.Header.Get(ContentTypeFieldName)
	if !strings.HasPrefix(hdrValue, HtmlHeaderValue) {
		return nil, ErrUnparsableHTML
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(" failed to read the page : %w", err)
	}

	return &Page{URL: resp.Request.URL.String(), Body: data}, nil

}
