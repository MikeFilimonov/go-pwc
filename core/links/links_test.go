package links

import (
	"net/url"
	"testing"
)

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
