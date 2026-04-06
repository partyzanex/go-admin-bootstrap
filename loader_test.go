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
