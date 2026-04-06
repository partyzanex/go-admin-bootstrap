package widgets

import (
	"cmp"
	"slices"
)

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

	slices.SortFunc(items, func(a, c breadcrumb) int {
		return cmp.Compare(a.SortOrder, c.SortOrder)
	})

	*b = items
}
