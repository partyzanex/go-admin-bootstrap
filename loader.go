package goadmin

import (
	"embed"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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
		return nil, errors.Wrapf(err, "cannot open %q file", templatePath)
	}

	return r, nil
}
