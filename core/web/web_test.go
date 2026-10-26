package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {

	// unreachable server

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	failard := httptest.NewServer(http.HandlerFunc(handler))
	url := failard.URL
	failard.Close()

	fetcher := Fetcher{}
	page, err := fetcher.Fetch(url)
	if !errors.Is(err, ErrRequestFailed) {
		t.Errorf("server unreachable: unexpected error state: \"%v\", got \"%v\"",
			ErrRequestFailed, err)
	}

	// active server, various cases

	cases := []struct {
		name          string
		input         string
		outputURL     string
		expectedError error
	}{
		{"basic", "/ok", "/ok", nil},
		{"not found", "/notfound", "/notfound", ErrBadStatus},
		{"not HTML", "/json", "/json", ErrUnparsableHTML},
		{"redirect", "/redirect", "/ok", nil},
	}

	handler = func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.Header().Add("Content-Type", "text/html")
			w.Write([]byte(`<p><a class="c" href="/y">Y</a></p>`))
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/json":
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		case "/redirect":
			http.Redirect(w, r, "/ok", http.StatusFound)
		default:
		}

	}

	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	for _, item := range cases {

		page, err = fetcher.Fetch(srv.URL + item.input)
		if !errors.Is(err, item.expectedError) {
			t.Errorf("%s: unexpected error, expected \"%v\", got \"%v\"",
				item.name, item.expectedError, err)
		}

		if page != nil && err != nil {
			t.Errorf("%s: error on fetching the page %v", item.name, err)
		}

		if page == nil {
			if err == nil {
				t.Errorf("%s: failed to get the page, returned nil", item.name)
			}
			continue
		}

		if err == nil && page.URL != srv.URL+item.outputURL {
			t.Errorf("%s: unexpected output URL, expected \"%v\", got \"%v\"",
				item.name, item.outputURL, page.URL)
		}

	}

}
