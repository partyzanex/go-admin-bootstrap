package assets

import "embed"

var (
	//go:embed css/*.css plugins/bootstrap/css/*
	CSS embed.FS

	//go:embed plugins/bootstrap/js/* plugins/jquery/* plugins/popper/*
	JS embed.FS
)
