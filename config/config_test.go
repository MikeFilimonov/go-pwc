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
	}{
		{"default", []string{"-maxDepth=3"}, 3, nil, false, ""},
		{"invalid", []string{"-maxDepth=abc"}, 0, nil, true, ""},
		{"zero", []string{"-maxDepth=0"}, 0, nil, true, ""},
		{"help", []string{"-h"}, 0, flag.ErrHelp, true, "-h"},
		{"maxDepth", []string{"-maxDepth"}, 0, nil, true, "-maxDepth"},
	}

	for _, item := range cases {

		buffer := &bytes.Buffer{}
		result, err := Load(item.args, buffer)

		if !strings.Contains(buffer.String(), item.prompt) {
			t.Errorf("%s: expected prompt %s, got %s", item.name, item.prompt, buffer.String())
		}

		if err != nil && !item.errorExpected || err == nil && item.errorExpected {
			t.Errorf("%s: unexpected error state: err=%v", item.name, err)
		}

		if !item.errorExpected && result.MaxDepth != item.expected {
			t.Errorf("%s: expected: %v, got: %v", item.name, item.expected, result.MaxDepth)
		}

		if item.errorExpected && item.errorType != nil && !errors.Is(err, item.errorType) {
			t.Errorf("%s: unexpected error: expected: %s, got %s",
				item.name, item.errorType, err)
		}

	}

}
