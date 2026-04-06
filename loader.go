package goadmin

import (
	"embed"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type FSLoader struct {
	fs *embed.FS
}

func NewFSLoader(fs *embed.FS) *FSLoader {
	return &FSLoader{
		fs: fs,
	}
}

func (l *FSLoader) Exists(templatePath string) bool {
	templatePath = strings.TrimLeft(templatePath, "/")

	entries, err := l.fs.ReadDir(filepath.Dir(templatePath))
	if err != nil {
		return false
	}

	fileName := filepath.Base(templatePath)

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() == fileName {
			return true
		}
	}

	return false
}

func (l *FSLoader) Open(templatePath string) (io.ReadCloser, error) {
	templatePath = strings.TrimLeft(templatePath, "/")

	r, err := l.fs.Open(templatePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open %q file: %w", templatePath, err)
	}

	return r, nil
}
