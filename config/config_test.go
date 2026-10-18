package config

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {

	cases := []struct {
		name          string
		args          []string
		expected      int
		errorType     error
		errorExpected bool
		prompt        string
		initialURL    string
	}{
		{"default", []string{"-maxDepth=3", "https://a.com"}, 3, nil, false, "", "https://a.com"},
		{"forgotten url", []string{"-maxDepth=3"}, 3, nil, true, "please, provide the initial URL to start processing", "https://a.com"},
		{"invalid", []string{"-maxDepth=abc", "https://a.com"}, 0, nil, true, "", "https://a.com"},
		{"zero", []string{"-maxDepth=0", "https://a.com"}, 0, nil, true, "", "https://a.com"},
		{"help", []string{"-h", "https://a.com"}, 0, flag.ErrHelp, true, "-h", "https://a.com"},
		{"maxDepth", []string{"-maxDepth"}, 0, nil, true, "-maxDepth", ""},
		{"two URLS", []string{"-maxDepth=3", "https://a.com", "https://d.com"}, 3, nil, false, "", "https://a.com"},
	}

	for _, item := range cases {

		buffer := &bytes.Buffer{}
		result, err := Load(item.args, buffer)

		if !strings.Contains(buffer.String(), item.prompt) {
			t.Errorf("%s: expected prompt \"%s\", got \"%s\"", item.name, item.prompt, buffer.String())
		}

		if err != nil && !item.errorExpected || err == nil && item.errorExpected {
			t.Errorf("%s: unexpected error state: err=%v", item.name, err)
			continue
		}

		if !item.errorExpected && result.MaxDepth != item.expected {
			t.Errorf("%s: expected: %v, got: %v", item.name, item.expected, result.MaxDepth)
		}

		if item.errorExpected && item.errorType != nil && !errors.Is(err, item.errorType) {
			t.Errorf("%s: unexpected error: expected: \"%s\", got \"%s\"",
				item.name, item.errorType, err)
		}

		if !item.errorExpected && result.InitialURL != item.initialURL {
			t.Errorf("%s: initial url \"%s\"  mismatches the  expected one : \"%s\"",
				item.name, result.InitialURL, item.initialURL)
		}

	}

}
