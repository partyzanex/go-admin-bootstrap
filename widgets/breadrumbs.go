package widgets

import "sort"

type breadcrumb struct {
	Name      string
	URL       string
	Active    bool
	SortOrder int
}

type Breadcrumbs []breadcrumb

func (b *Breadcrumbs) Add(name, url string, sortOrder *int) {
	items := *b
	for i := range items {
		items[i].Active = false
	}

	var order int

	if sortOrder == nil {
		order = len(items)
	} else {
		order = *sortOrder
	}

	items = append(items, breadcrumb{
		Name:      name,
		URL:       url,
		Active:    true,
		SortOrder: order,
	})

	*b = items
}

func (b *Breadcrumbs) Sort() {
	items := *b

	sort.Slice(items, func(i, j int) bool {
		return items[i].SortOrder < items[j].SortOrder
	})

	*b = items
}
