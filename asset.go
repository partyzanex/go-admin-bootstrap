package goadmin

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Asset struct {
	Path string
	URL  string
}

func (a *Admin) LoadSources() error {
	for _, source := range js {
		err := a.loadSource(a.AssetsPath, source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	for _, source := range css {
		err := a.loadSource(a.AssetsPath, source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	for _, source := range views {
		err := a.loadSource(a.ViewsPath, source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	return nil
}

func (Admin) loadSource(path string, source Asset) error {
	sourcePath := filepath.Join(path, source.Path)

	_, err := os.Stat(sourcePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "loading asset %s (%s) source failed",
			source.Path, source.URL,
		)
	}
	if os.IsNotExist(err) {
		sourceDir := filepath.Dir(sourcePath)
		err = os.MkdirAll(sourceDir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "make assets dir %s failed",
				sourceDir,
			)
		}

		resp, err := http.Get(source.URL)
		if err != nil {
			return errors.Wrapf(err, "loading asset source from url %s failed", source.URL)
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// todo: wrap error
			return err
		}

		err = ioutil.WriteFile(sourcePath, b, os.ModePerm)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	return nil
}
