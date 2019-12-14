package goadmin

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var (
	js = []AssetSource{
		{
			Path: "plugins/jquery/jquery-3.4.1.min.js",
			URL:  "https://code.jquery.com/jquery-3.4.1.min.js",
		},
		{
			Path: "plugins/bootstrap/js/bootstrap.min.js",
			URL:  "https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js",
		},
	}
	css = []AssetSource{
		{
			Path: "plugins/bootstrap/css/bootstrap.min.css",
			URL:  "https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css",
		},
		{
			Path: "css/style.css",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/v0.0.1/assets/css/style.css",
		},
	}
)

type AssetSource struct {
	Path string
	URL  string
}

func (admin *Admin) LoadAssets() error {
	for _, source := range js {
		err := admin.loadAsset(source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	for _, source := range css {
		err := admin.loadAsset(source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	return nil
}

func (admin Admin) loadAsset(source AssetSource) error {
	assetPath := filepath.Join(admin.AssetsPath, source.Path)

	_, err := os.Stat(assetPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "loading asset %s (%s) source failed",
			source.Path, source.URL,
		)
	}
	if os.IsNotExist(err) {
		assetDir := filepath.Dir(assetPath)
		err = os.MkdirAll(assetDir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "make assets dir %s failed",
				assetDir,
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

		err = ioutil.WriteFile(assetPath, b, os.ModePerm)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	return nil
}
