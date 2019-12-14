package widgets

type breadcrumb struct {
	Name   string
	URL    string
	Active bool
}

type Breadcrumbs []breadcrumb

func (b *Breadcrumbs) Add(name, url string) {
	items := *b
	for i := range items {
		items[i].Active = false
	}

	items = append(items, breadcrumb{
		Name:   name,
		URL:    url,
		Active: true,
	})

	*b = items
}
