package pages

import (
	"bytes"

	"golang.org/x/net/html"
)

const (
	linkTagElement             = "a"
	refAttributeName           = "href"
	populationDensityThreshold = 0.2
)

func Parse(input []byte) ([]string, error) {

	result := make([]string, 0)

	tokenizer := html.NewTokenizer(bytes.NewBuffer(input))

	for {

		token := tokenizer.Next()

		if token == html.ErrorToken {
			return result, nil
		}

		if token == html.StartTagToken {
			name, hasAttributes := tokenizer.TagName()
			if string(name) != linkTagElement {
				continue
			}

			for hasAttributes {

				k, v, hasMore := tokenizer.TagAttr()
				hasAttributes = hasMore

				if string(k) == refAttributeName {
					result = append(result, string(v))
					break
				}

			}

		}

	}

}

type Stats struct {
	PageCount  int
	TotalBytes int
	LinksFound int
}

type Item struct {
	URL   string
	Level int
}

type Buffer struct {
	Items          []Item
	allocationSize int
}

func (b *Buffer) Enqueue(item Item) {

	if b.allocationSize == 0 {
		b.allocationSize = cap(b.Items)
	}

	currentCapacity := cap(b.Items)
	b.Items = append(b.Items, item)
	if cap(b.Items) != currentCapacity {
		b.allocationSize = cap(b.Items)
	}
}

func (b *Buffer) Dequeue() (Item, bool) {

	if b.allocationSize == 0 {
		b.allocationSize = cap(b.Items)
	}

	if len(b.Items) == 0 {

		b.allocationSize = 0
		b.Items = nil
		return Item{}, false
	}

	result := b.Items[0]
	b.Items[0] = Item{}
	b.Items = b.Items[1:]

	if len(b.Items) == 0 {
		b.allocationSize = 0
		b.Items = nil
		return result, true
	}

	if float32(len(b.Items))/float32(b.allocationSize) <= populationDensityThreshold {
		b.optimizeAllocation()
	}

	return result, true

}

func (b *Buffer) optimizeAllocation() {

	optimized := make([]Item, len(b.Items), 2*len(b.Items))
	copy(optimized, b.Items)
	b.Items = optimized
	b.allocationSize = cap(optimized)

}

type PagesSeen struct {
	elements map[string]struct{} //[normalized url]

}

func (p *PagesSeen) Add(url string) {

	if p.elements == nil {
		p.elements = make(map[string]struct{})
	}
	p.elements[url] = struct{}{}

}

func (p *PagesSeen) Has(url string) bool {

	_, success := p.elements[url]
	return success

}

func (p *PagesSeen) Len() int {

	if p.elements == nil {
		return 0
	}

	return len(p.elements)

}
