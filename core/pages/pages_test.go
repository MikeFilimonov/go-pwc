package pages

import (
	"slices"
	"testing"
)

var (
	a = Item{URL: "a.html", Level: 1}
	b = Item{URL: "b.html", Level: 1}
	c = Item{URL: "c/a.html", Level: 2}
	d = Item{URL: "d/b.html", Level: 2}
	e = Item{URL: "d/b.html", Level: 3}
)

func TestParse(t *testing.T) {

	cases := []struct {
		name     string
		input    string
		expected []string
		err      error
	}{
		{"basic", `<p><a class="c" href="/y">Y</a></p>`, []string{"/y"}, nil},
		{"special tag", `<link rel="stylesheet" href="style.css">`, []string{}, nil},
		{"combo", `<p><a class="c" href="/y">Y</a><a class="c" href="/no">Y</a><a class="c" href="acme.html">Y</a></p>`,
			[]string{"/y", "/no", "acme.html"}, nil},
		{"empty", `<p><a/></p>`, []string{}, nil},
		{"broken HTML", `<p><a class="c" href="/y">Y<a></p>`, []string{"/y"}, nil},
	}

	for _, item := range cases {

		input := []byte(item.input)
		result, err := Parse(input)
		if err != item.err {
			t.Errorf("%s:unexpected error state, expected \"%v\", got \"%v\"\n", item.name, item.err, err)
		}

		if !slices.Equal(item.expected, result) {
			t.Errorf("%s: expected %v, got: %v", item.name, item.expected, result)
		}

	}

}

func TestEnqueue(t *testing.T) {

	cases := []struct {
		name     string
		added    Item
		buffer   Buffer
		expected []Item
	}{
		{"order", e, Buffer{Items: []Item{b, d}}, []Item{b, d, e}},
		{"empty item", Item{}, Buffer{Items: []Item{a, b}}, []Item{a, b, Item{}}},
	}

	for _, item := range cases {
		item.buffer.Enqueue(item.added)
		if !slices.Equal(item.buffer.Items, item.expected) {
			t.Errorf("%s, expected \"%v\" got \"%v\"", item.name, item.expected, item.buffer.Items)
		}
	}

}

func TestDequeue(t *testing.T) {

	cases := []struct {
		name         string
		buffer       Buffer
		expected     []Item
		toBeDequeued Item
		success      bool
	}{
		{"dequeue once", Buffer{Items: []Item{a, b, c, d}}, []Item{b, c, d}, a, true},
		{"empty", Buffer{Items: []Item{}}, []Item{}, Item{}, false},
	}

	for _, item := range cases {

		dequeued, success := item.buffer.Dequeue()

		if success != item.success {
			t.Errorf("%s: unexpected Dequeue completion state: expected \"%v\", got \"%v\" ",
				item.name, item.success, success)
		}

		if success && item.toBeDequeued != dequeued {
			t.Errorf("%s: dequeued the wrong item: expected \"%v\", got \"%v\"",
				item.name, item.toBeDequeued, dequeued)
		}

		if !slices.Equal(item.buffer.Items, item.expected) {
			t.Errorf("%s: expected \"%v\" got \"%v\"", item.name, item.expected, item.buffer.Items)
		}
	}

}

func TestQueueRoutinesCombo(t *testing.T) {

	var buf = Buffer{Items: []Item{a, b, c}}

	iterations := len(buf.Items)

	for i := 0; i < iterations; i++ {
		if _, success := buf.Dequeue(); !success {
			t.Error("dequeuing: failed to get the item on dequeuing")
		}
	}
	if len(buf.Items) > 0 {
		t.Error("draining: the buffer failed to get drained")
	}

	value, success := buf.Dequeue()
	if success {
		t.Errorf("post-drained: the buffer returned a value after getting empty: \"%v\"", value)
	}

	buf.Enqueue(d)
	value, success = buf.Dequeue()
	if !success {
		t.Errorf("failed to get the previously queued item \"%v\"", d)
	}

	if success && value != d {
		t.Errorf("unexpected return: expected \"%v\", got \"%v\"", d, value)
	}

}

func TestAdd(t *testing.T) {

	cases := []struct {
		name           string
		added          []string
		expectedLength int
	}{
		{"default", []string{"acme.sixseven"}, 1},
		{"clones", []string{"acme.sixseven", "acme.sixseven"}, 1},
	}

	for _, item := range cases {

		var buffer PagesSeen

		for _, v := range item.added {
			buffer.Add(v)
		}
		if buffer.Len() != item.expectedLength {
			t.Errorf("%s: expected seen count (%v) doesn't meet the real one (%v)",
				item.name, item.expectedLength, buffer.Len())
		}

	}

}

func TestHas(t *testing.T) {

	cases := []struct {
		name     string
		added    []string
		searched string
		found    bool
	}{
		{"default", []string{"acme.sixseven"}, "acme.sixseven", true},
		{"clones", []string{"acme.sixseven", "acme.sixseven"}, "acme.sixseven", true},
		{"inexistent", []string{"acme.sixseven"}, "rnd.generic", false},
		{"zero", []string{}, "", false},
	}

	for _, item := range cases {

		var buffer PagesSeen

		for _, v := range item.added {
			buffer.Add(v)
		}

		if buffer.Has(item.searched) != item.found {
			t.Errorf("%s, unexpected search result: expected %v got %v",
				item.name, item.found, buffer.Has(item.searched))
		}

	}

}
