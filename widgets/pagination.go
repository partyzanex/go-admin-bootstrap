package widgets

import (
	"math"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	defaultLimit    = 10
	defaultNumLinks = 8
)

type Pagination struct {
	Ctx echo.Context

	Total       int64
	Limit, View int
	Page, Pages int
	NumLinks    int

	URLTemplate string
	PageParam   string

	Previous, Next PaginationItem
	First, Last    PaginationItem
}

type PaginationItem struct {
	PageNum int
	URL     string
	Current bool
}

func (item PaginationItem) Enable() bool {
	return item.PageNum > 0
}

func (p *Pagination) Exists() bool {
	p.ParsePage()

	if p.Limit == 0 {
		p.Limit = defaultLimit
	}

	pages := int(math.Ceil(float64(p.Total) / float64(p.Limit)))

	return pages > 1
}

func (p *Pagination) ParsePage() {
	page := p.Ctx.QueryParam(p.PageParam)
	p.Page, _ = strconv.Atoi(page)

	if p.Page == 0 {
		p.Page = 1
	}

	p.View = p.Page * p.Limit
	if p.View > int(p.Total) {
		p.View = int(p.Total)
	}

	if p.Page > 1 {
		p.Previous = PaginationItem{
			PageNum: p.Page - 1,
			URL:     p.url(p.Page - 1),
		}
	}

	pages := int(math.Ceil(float64(p.Total) / float64(p.Limit)))
	if p.Page < pages {
		p.Next = PaginationItem{
			PageNum: p.Page + 1,
			URL:     p.url(p.Page + 1),
		}
	}

	if p.Page+1 < pages {
		p.Last = PaginationItem{
			PageNum: pages,
			URL:     p.url(pages),
		}
	}

	if p.Page-1 > 0 {
		p.First = PaginationItem{
			PageNum: 1,
			URL:     p.url(1),
		}
	}
}

func (p *Pagination) Items() []PaginationItem {
	p.ParsePage()

	if p.Page < 1 {
		p.Page = 1
	}

	if p.Limit == 0 {
		p.Limit = defaultLimit
	}

	if p.NumLinks == 0 {
		p.NumLinks = defaultNumLinks
	}

	pages := int(math.Ceil(float64(p.Total) / float64(p.Limit)))

	if pages <= 1 {
		return nil
	}

	start := 1
	end := pages

	if pages > p.NumLinks {
		part := int(math.Floor(float64(p.NumLinks) / 2))
		start = p.Page - part
		end = p.Page + part
	}

	if start < 1 {
		end += int(math.Abs(float64(start)))
		start = 1
	}

	if end > pages {
		start -= end - pages
		end = pages
	}

	items := make([]PaginationItem, end-start+1)

	j := 0

	for i := start; i <= end; i++ {
		items[j] = PaginationItem{
			PageNum: i,
			URL:     p.url(i),
			Current: p.Page == i,
		}

		j++
	}

	return items
}

func (p Pagination) url(pageNum int) string {
	return strings.Replace(p.URLTemplate, "{page}", strconv.Itoa(pageNum), 1)
}
