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
	defaultFetcherTimeout = 10 * time.Second
	contentTypeFieldName  = "Content-Type"
	htmlHeaderValue       = "text/html"
)

var (
	ErrBadStatus      = errors.New("bad response")
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

func (f *Fetcher) init() {
	if f.timeout == 0 {
		f.timeout = defaultFetcherTimeout
	}
	if f.client == nil {
		f.client = &http.Client{Timeout: f.timeout}
	}
}

func (f *Fetcher) Fetch(url string) (*Page, error) {

	f.init()
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrBadStatus, resp.StatusCode)
	}

	hdrValue := resp.Header.Get(contentTypeFieldName)
	if !strings.HasPrefix(hdrValue, htmlHeaderValue) {
		return nil, ErrUnparsableHTML
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(" failed to read the page : %w", err)
	}

	return &Page{URL: resp.Request.URL.String(), Body: data}, nil

}
