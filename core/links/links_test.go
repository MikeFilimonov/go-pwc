package links

import (
	"net/url"
	"testing"
)

func TestFixURL(t *testing.T) {

	cases := []struct {
		name          string
		input         string
		expected      string
		errorExpected error
	}{
		{"default", "HTTPS://go.dev", "HTTPS://go.dev", nil},
		{"nextPage", "go.dev/login?next=https://go.dev/doc", "https://go.dev/login?next=https://go.dev/doc", nil},
		{"empty", "", "", ErrEmptyHost},
		{"go.dev:port", "go.dev:8080", "https://go.dev:8080", nil},
		{"ftp", "ftp://go.dev", "", ErrUnsupportedScheme},
		{"localhost", "localhost", "https://localhost", nil},
		{"fake scheme", "://go.dev", "", ErrUnsupportedScheme},
		{"fake host", "https://", "", ErrEmptyHost},
		{"keep the case", "go.dev/Login?next=https://go.dev/doc", "https://go.dev/Login?next=https://go.dev/doc", nil},
	}

	for _, item := range cases {

		result, err := FixURL(item.input)
		if err != item.errorExpected {
			t.Errorf("%s: unexpected error state: err=%v", item.name, err)
			continue
		}

		if result != item.expected {
			t.Errorf("%s: expected \"%s\", got \"%s\"", item.name, item.expected, result)
		}

	}

}

func TestNormalize(t *testing.T) {

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"default", "hTTps://mRh.com:80/waylon.html", "https://mrh.com:80/waylon.html"},
		{"default_port", "https://marvel.co:443/semiconductors.htm", "https://marvel.co/semiconductors.htm"},
		{"fragments", "https://abc.com#memo", "https://abc.com"},
		{"surreal_protocol_with_trailing_slash", "sptth://com.ru.test.co/generandom/", "sptth://com.ru.test.co/generandom"},
		{"singletone", "/", "/"},
	}

	for _, item := range cases {

		parsed, _ := url.Parse(item.input)

		if Normalize(parsed) != item.expected {
			t.Errorf("%s: expected %s, got %s", item.name, item.expected, Normalize(parsed))
		}

	}

}

func TestComplete(t *testing.T) {

	cases := []struct {
		name           string
		baseUrl        string
		initialUrl     string
		resultsInError bool
		expected       string
	}{

		{"default", "https://generic.random", "boolship.html", false, "https://generic.random/boolship.html"},
		{"root", "https://single.tone", "/", false, "https://single.tone/"},
		{"empty_base", "", "/peace.html", false, "/peace.html"},
		{"erroneous", "http://a.com/\n", "smth.html", true, ""},
		{"go.dev", "go.dev", "go.dev", false, "/go.dev"},
	}

	for _, item := range cases {

		completed, err := Complete(item.baseUrl, item.initialUrl)
		if err != nil && !item.resultsInError ||
			err == nil && item.resultsInError {
			t.Errorf("%s: unexpected error state: err=%v", item.name, err)
			continue
		}

		if err == nil && completed.String() != item.expected {
			t.Errorf("%s: expected: %s, got: %s", item.name, item.expected, completed.String())
		}

	}

}

func TestIsFromSameZone(t *testing.T) {

	cases := []struct {
		name          string
		expected      bool
		source        string
		input         string
		errorExpected bool
	}{
		{"same domains", true, "https://future.life.com", "https://current.life.com", false},
		{"bare vs subdomain", true, "https://x.com", "https://y.x.com", false},
		{".co.name trap", false, "https://sixseven.co.uk", "https://schools.co.uk", false},
		{"parse failure", false, "http://a.com/\n", "http://tbd.a.com", true},
		{"+/-", false, "http://six.com", "http://seven.ru/\n", true},
		{"eTLD+1", false, "http://localhost", "http://generic.random", true},
		{"githubbyO", false, "https://x.github.io", "https://y.github.io", false},
		{"no_eTLD+1_input", false, "http://x.com", "http://127.0.0.1", true},
	}

	for _, item := range cases {

		result, err := IsFromSameZone(item.source, item.input)

		if (err != nil) != item.errorExpected {
			t.Errorf("%s unexpected error:  %v, expected: %v, got: %v",
				item.name, err, item.errorExpected, err != nil)
			continue
		}

		if result != item.expected {
			t.Errorf("%s: %v , expected: %v, got %v",
				item.name, err, item.expected, result)
		}

	}

}
