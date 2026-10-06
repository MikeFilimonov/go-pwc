package config

import "testing"

func TestLoad(t *testing.T) {

	cases := []struct {
		name          string
		args          []string
		expected      int
		errorExpected bool
	}{
		{"default", []string{"-maxDepth=3"}, 3, false},
		{"invalid", []string{"-maxDepth=abc"}, 0, true},
		{"zero", []string{"-maxDepth=0"}, 0, false},
	}

	for _, item := range cases {

		result, err := Load(item.args)

		if err != nil && !item.errorExpected || err == nil && item.errorExpected {
			t.Errorf("%s: unexpected error state: err=%v", item.name, err)
		}

		if !item.errorExpected && result.MaxDepth != item.expected {
			t.Errorf("%s: expected: %v, got: %v", item.name, item.expected, result.MaxDepth)
		}

	}

}
