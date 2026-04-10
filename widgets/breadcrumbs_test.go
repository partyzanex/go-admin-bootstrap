package widgets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int {
	return &v
}

func TestBreadcrumbs_AddSingle(t *testing.T) {
	var b Breadcrumbs
	b.Add("Home", "/", nil)

	require.Len(t, b, 1)
	assert.Equal(t, "Home", b[0].Name)
	assert.Equal(t, "/", b[0].URL)
	assert.True(t, b[0].Active, "last added should be active")
	assert.Equal(t, 0, b[0].SortOrder, "nil sort order should default to current length (0)")
}

func TestBreadcrumbs_AddMultiple(t *testing.T) {
	var b Breadcrumbs
	b.Add("Home", "/", nil)
	b.Add("Users", "/users", nil)
	b.Add("Edit", "/users/1/edit", nil)

	require.Len(t, b, 3)

	// Only the last one should be active
	assert.False(t, b[0].Active, "Home should not be active")
	assert.False(t, b[1].Active, "Users should not be active")
	assert.True(t, b[2].Active, "Edit should be active (last added)")
}

func TestBreadcrumbs_AddNilSortOrder(t *testing.T) {
	var b Breadcrumbs
	b.Add("Home", "/", nil)
	b.Add("Users", "/users", nil)
	b.Add("Edit", "/users/1/edit", nil)

	// With nil sort order, items get sequential sort orders
	assert.Equal(t, 0, b[0].SortOrder)
	assert.Equal(t, 1, b[1].SortOrder)
	assert.Equal(t, 2, b[2].SortOrder)
}

func TestBreadcrumbs_AddCustomSortOrder(t *testing.T) {
	var b Breadcrumbs
	b.Add("Edit", "/users/1/edit", intPtr(30))
	b.Add("Home", "/", intPtr(10))
	b.Add("Users", "/users", intPtr(20))

	assert.Equal(t, 30, b[0].SortOrder)
	assert.Equal(t, 10, b[1].SortOrder)
	assert.Equal(t, 20, b[2].SortOrder)

	// Only the last one should be active
	assert.False(t, b[0].Active)
	assert.False(t, b[1].Active)
	assert.True(t, b[2].Active)
}

func TestBreadcrumbs_Sort(t *testing.T) {
	var b Breadcrumbs
	b.Add("Edit", "/users/1/edit", intPtr(30))
	b.Add("Home", "/", intPtr(10))
	b.Add("Users", "/users", intPtr(20))

	b.Sort()

	require.Len(t, b, 3)
	assert.Equal(t, "Home", b[0].Name)
	assert.Equal(t, "Users", b[1].Name)
	assert.Equal(t, "Edit", b[2].Name)
}

func TestBreadcrumbs_SortWithNilSortOrder(t *testing.T) {
	var b Breadcrumbs
	b.Add("Home", "/", nil)
	b.Add("Users", "/users", nil)
	b.Add("Edit", "/users/1/edit", nil)

	b.Sort()

	// Already in order (0, 1, 2), sort should preserve
	assert.Equal(t, "Home", b[0].Name)
	assert.Equal(t, "Users", b[1].Name)
	assert.Equal(t, "Edit", b[2].Name)
}

func TestBreadcrumbs_SortPreservesActiveFlag(t *testing.T) {
	var b Breadcrumbs
	b.Add("Z-Last", "/z", intPtr(100))
	b.Add("A-First", "/a", intPtr(1))

	b.Sort()

	assert.Equal(t, "A-First", b[0].Name)
	assert.True(t, b[0].Active, "A-First was last added, should remain active after sort")
	assert.Equal(t, "Z-Last", b[1].Name)
	assert.False(t, b[1].Active)
}

func TestBreadcrumbs_AddDeactivatesPreviousItems(t *testing.T) {
	var b Breadcrumbs
	b.Add("Home", "/", nil)

	assert.True(t, b[0].Active)

	b.Add("Users", "/users", nil)

	assert.False(t, b[0].Active, "Home should be deactivated after adding Users")
	assert.True(t, b[1].Active, "Users should be active")
}

func TestBreadcrumbs_Empty(t *testing.T) {
	var b Breadcrumbs

	assert.Len(t, b, 0)

	b.Sort() // should not panic on empty

	assert.Len(t, b, 0)
}

func TestBreadcrumbs_SortEqualOrder(t *testing.T) {
	var b Breadcrumbs
	b.Add("B", "/b", intPtr(1))
	b.Add("A", "/a", intPtr(1))

	b.Sort()

	// Both have sort order 1; stable sort not guaranteed by SortFunc,
	// but we just verify no panic and length is correct
	assert.Len(t, b, 2)
}
