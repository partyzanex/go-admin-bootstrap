package goadmin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/partyzanex/go-admin-bootstrap/views"
)

func TestFSLoader_Open(t *testing.T) {
	loader := NewFSLoader(&views.Sources)

	assert.True(t, loader.Exists("auth/login.jet"))
	assert.False(t, loader.Exists("auth/any.jet"))

	r, err := loader.Open("auth/login.jet")
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestFSLoader_Exists_NonExistentDir(t *testing.T) {
	loader := NewFSLoader(&views.Sources)

	// Directory does not exist → ReadDir will return an error → Exists will return false
	assert.False(t, loader.Exists("nonexistent/template.jet"))
}

func TestFSLoader_Open_NonExistent(t *testing.T) {
	loader := NewFSLoader(&views.Sources)

	_, err := loader.Open("auth/nonexistent.jet")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot open")
}
