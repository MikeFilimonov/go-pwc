package pages

type Stats struct {
	PageCount  int
	TotalBytes int
	LinksFound []string
}

type Item struct {
	Url   string
	Level int
}

type Buffer struct {
	Items []Item
}

type PagesSeen struct {
	Elements map[string]struct{}
}
