package pages

type Stats struct {
	PageCount  int
	TotalBytes int
	LinksFound int
}

type Item struct {
	Url   string
	Level int
}

type Buffer struct {
	Items []Item
}

func (b *Buffer) Enqueue(item Item) {
	b.Items = append(b.Items, item)
}

func (b *Buffer) Dequeue() (Item, bool) {

	if len(b.Items) == 0 {
		return Item{}, false
	}

	result := b.Items[0]
	b.Items[0] = Item{}
	b.Items = b.Items[1:]
	return result, true

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
