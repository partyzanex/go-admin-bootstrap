package cmd

import "errors"

var (
	ErrBinNotFound = errors.New("binary not found")
	ErrTimeOut     = errors.New("process timeout exceeded")
)
