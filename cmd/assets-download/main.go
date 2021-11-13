package main

import (
	"github.com/spf13/pflag"
	"log"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

func main() {
	var (
		assetsPath = pflag.StringP("assets-path", "a", "./assets", "path fo assets sources")
		viewsPath  = pflag.StringP("views-path", "v", "./views", "path fo views sources")
	)

	pflag.Parse()

	admin := new(goadmin.Admin)
	admin.Config = new(goadmin.Config)
	admin.AssetsPath = *assetsPath
	admin.ViewsPath = *viewsPath

	if err := admin.LoadSources(); err != nil {
		log.Fatal(err)
	}
}
