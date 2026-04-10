package goadmin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestData_Set_And_Has(t *testing.T) {
	data := &Data{}

	assert.False(t, data.Has("key"))

	data.Set("key", "value")

	assert.True(t, data.Has("key"))
	assert.False(t, data.Has("missing"))
}

func TestData_JetVars(t *testing.T) {
	data := &Data{Title: "Test Page"}
	data.Set("custom", "val")

	vars := data.JetVars()

	assert.NotNil(t, vars)
	assert.Contains(t, vars, "title")
	assert.Contains(t, vars, "scripts")
	assert.Contains(t, vars, "styles")
	assert.Contains(t, vars, "custom")
}

func TestData_JetData(t *testing.T) {
	data := &Data{}

	// No user, no breadcrumbs
	result := data.JetData()
	assert.Empty(t, result)

	// With user
	data.User = &User{ID: 1, Name: "test"}

	result = data.JetData()
	assert.Contains(t, result, "User")

	// With breadcrumbs
	data.Breadcrumbs.Add("Home", "/", nil)

	result = data.JetData()
	assert.Contains(t, result, "Breadcrumbs")
}

func TestData_JetVars_InitializesVarMap(t *testing.T) {
	data := &Data{}

	// VarMap is nil initially
	vars := data.JetVars()
	assert.NotNil(t, vars)
}
