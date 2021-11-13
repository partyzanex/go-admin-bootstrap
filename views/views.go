package views

import "embed"

//go:embed auth/* errors/* index/* layouts/* user/* widgets/*
var Sources embed.FS
