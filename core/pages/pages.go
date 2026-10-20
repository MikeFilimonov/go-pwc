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
	Elements map[string]struct{} //[normalized url]

}
