package widgets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext(pageParam, pageValue string) echo.Context {
	e := echo.New()
	target := "/"
	if pageValue != "" {
		target = "/?" + pageParam + "=" + pageValue
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestParsePage_ValidPage(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "3"),
		Total:       100,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 3, p.Page)
	assert.Equal(t, 30, p.View)
}

func TestParsePage_PageZero(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "0"),
		Total:       50,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 1, p.Page, "page=0 should be normalized to 1")
}

func TestParsePage_NegativePage(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "-5"),
		Total:       50,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 1, p.Page, "negative page should be normalized to 1")
}

func TestParsePage_NonNumeric(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "abc"),
		Total:       50,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 1, p.Page, "non-numeric page should default to 1")
}

func TestParsePage_EmptyQueryParam(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", ""),
		Total:       50,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 1, p.Page, "empty page should default to 1")
}

func TestParsePage_ViewClampedToTotal(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "10"),
		Total:       25,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.LessOrEqual(t, p.View, int(p.Total), "view should not exceed total")
}

func TestParsePage_PreviousAndNext(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "3"),
		Total:       100,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 2, p.Previous.PageNum)
	assert.Equal(t, "/items?page=2", p.Previous.URL)
	assert.Equal(t, 4, p.Next.PageNum)
	assert.Equal(t, "/items?page=4", p.Next.URL)
}

func TestParsePage_FirstPage_NoPrevious(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       100,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 0, p.Previous.PageNum, "first page should have no previous")
	assert.False(t, p.Previous.Enable())
}

func TestParsePage_LastPage_NoNext(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "10"),
		Total:       100,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 0, p.Next.PageNum, "last page should have no next")
	assert.False(t, p.Next.Enable())
}

func TestParsePage_FirstAndLastLinks(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "5"),
		Total:       100,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}
	p.ParsePage()

	assert.Equal(t, 1, p.First.PageNum)
	assert.Equal(t, "/items?page=1", p.First.URL)
	assert.Equal(t, 10, p.Last.PageNum)
	assert.Equal(t, "/items?page=10", p.Last.URL)
}

func TestExists_MultiplePages(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       25,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	assert.True(t, p.Exists(), "25 items / 10 per page = 3 pages, should exist")
}

func TestExists_SinglePage(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       5,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	assert.False(t, p.Exists(), "5 items / 10 per page = 1 page, should not exist")
}

func TestExists_ExactlyOnePage(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       10,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	assert.False(t, p.Exists(), "exactly 1 page should not exist")
}

func TestExists_ZeroTotal(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       0,
		Limit:       10,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	assert.False(t, p.Exists(), "zero total should not exist")
}

func TestExists_DefaultLimit(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       15,
		Limit:       0, // should use defaultLimit=10
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	assert.True(t, p.Exists(), "15 items with default limit 10 should be 2 pages")
}

func TestPageRange_MultiplePages(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       100,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	start, end := p.pageRange()

	assert.GreaterOrEqual(t, start, 1)
	assert.LessOrEqual(t, end, 10)
	assert.LessOrEqual(t, end-start+1, p.NumLinks+1)
}

func TestPageRange_SinglePage(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       5,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	start, end := p.pageRange()

	assert.Equal(t, 0, start, "single page range should be empty")
	assert.Equal(t, -1, end, "single page range should be empty")
}

func TestPageRange_ZeroItems(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       0,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	start, end := p.pageRange()

	assert.Equal(t, 0, start)
	assert.Equal(t, -1, end)
}

func TestPageRange_DefaultNumLinks(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "5"),
		Total:       200,
		Limit:       10,
		NumLinks:    0, // should default to 8
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	start, end := p.pageRange()

	assert.GreaterOrEqual(t, start, 1)
	assert.LessOrEqual(t, end, 20)
}

func TestPageRange_FewerPagesThanNumLinks(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       30,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	start, end := p.pageRange()

	assert.Equal(t, 1, start)
	assert.Equal(t, 3, end, "3 pages, all should be included")
}

func TestAll_IteratorBehavior(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "2"),
		Total:       50,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	var items []PaginationItem
	for item := range p.All() {
		items = append(items, item)
	}

	require.NotEmpty(t, items)

	// Verify current page is marked
	foundCurrent := false
	for _, item := range items {
		if item.PageNum == 2 {
			assert.True(t, item.Current, "page 2 should be marked as current")
			foundCurrent = true
		} else {
			assert.False(t, item.Current, "page %d should not be current", item.PageNum)
		}
	}
	assert.True(t, foundCurrent, "current page should be in the items")
}

func TestAll_SinglePageReturnsNoItems(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       5,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	var items []PaginationItem
	for item := range p.All() {
		items = append(items, item)
	}

	assert.Empty(t, items, "single page should yield no items")
}

func TestAll_URLsAreCorrect(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       30,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	var items []PaginationItem
	for item := range p.All() {
		items = append(items, item)
	}

	require.Len(t, items, 3)
	assert.Equal(t, "/items?page=1", items[0].URL)
	assert.Equal(t, "/items?page=2", items[1].URL)
	assert.Equal(t, "/items?page=3", items[2].URL)
}

func TestAll_EarlyBreak(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       100,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	count := 0
	for range p.All() {
		count++
		if count == 3 {
			break
		}
	}

	assert.Equal(t, 3, count, "should be able to break early from iterator")
}

func TestItems_ReturnsSeqRanger(t *testing.T) {
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       30,
		Limit:       10,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	ranger := p.Items()
	require.NotNil(t, ranger)
	assert.True(t, ranger.ProvidesIndex())

	// Iterate through the ranger
	var count int
	for {
		idx, val, end := ranger.Range()
		if end {
			break
		}
		assert.Equal(t, count, idx.Interface().(int))
		item := val.Interface().(PaginationItem)
		assert.Equal(t, count+1, item.PageNum)
		count++
	}
	assert.Equal(t, 3, count)
}

func TestURL_TemplateSubstitution(t *testing.T) {
	p := &Pagination{
		URLTemplate: "/users?page={page}&sort=name",
	}

	assert.Equal(t, "/users?page=5&sort=name", p.url(5))
	assert.Equal(t, "/users?page=1&sort=name", p.url(1))
}

func TestURL_MultiplePagePlaceholders(t *testing.T) {
	p := &Pagination{
		URLTemplate: "/items?page={page}&ref={page}",
	}

	assert.Equal(t, "/items?page=3&ref=3", p.url(3))
}

func TestPaginationItem_Enable(t *testing.T) {
	assert.True(t, PaginationItem{PageNum: 1}.Enable())
	assert.True(t, PaginationItem{PageNum: 5}.Enable())
	assert.False(t, PaginationItem{PageNum: 0}.Enable())
	assert.False(t, PaginationItem{PageNum: -1}.Enable())
}

func TestPageRange_DefaultLimit(t *testing.T) {
	// Limit=0 → defaultLimit should be set inside pageRange
	p := &Pagination{
		Ctx:         newTestContext("page", "1"),
		Total:       0,
		Limit:       0,
		NumLinks:    8,
		PageParam:   "page",
		URLTemplate: "/items?page={page}",
	}

	start, end := p.pageRange()

	// Total=0 → pages=0 → empty range
	assert.Equal(t, 0, start)
	assert.Equal(t, -1, end)
	// defaultLimit should have been set
	assert.Equal(t, defaultLimit, p.Limit)
}
