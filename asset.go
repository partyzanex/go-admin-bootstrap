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
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/assets/css/style.css",
		},
	}
	views = []AssetSource{
		{
			Path: "layouts/nav.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/layouts/nav.jet",
		},
		{
			Path: "layouts/main.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/layouts/main.jet",
		},
		{
			Path: "widgets/breadcrumbs.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/widgets/breadcrumbs.jet",
		},
		{
			Path: "widgets/pagination.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/widgets/pagination.jet",
		},
		{
			Path: "index/dashboard.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/index/dashboard.jet",
		},
		{
			Path: "errors/error.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/errors/error.jet",
		},
		{
			Path: "auth/login.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/auth/login.jet",
		},
		{
			Path: "user/index.jet",
			URL:  "https://raw.githubusercontent.com/partyzanex/go-admin-bootstrap/" + Version + "/views/user/index.jet",
		},
	}
)

type AssetSource struct {
	Path string
	URL  string
}

func (admin *Admin) LoadSources() error {
	for _, source := range js {
		err := admin.loadSource(admin.AssetsPath, source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	for _, source := range css {
		err := admin.loadSource(admin.AssetsPath, source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	for _, source := range views {
		err := admin.loadSource(admin.ViewsPath, source)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	return nil
}

func (admin Admin) loadSource(path string, source AssetSource) error {
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
