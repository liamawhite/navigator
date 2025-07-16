//go:build !test && !integration && !lint

package ui

import (
	"embed"
	"io/fs"
)

// EmbeddedFiles contains the embedded UI assets
//
//go:embed dist dist/* dist/assets/*
var EmbeddedFiles embed.FS

// GetFileSystem returns a filesystem that can be used to serve the UI assets.
// It strips the "dist" prefix so files can be served from the root.
func GetFileSystem() (fs.FS, error) {
	return fs.Sub(EmbeddedFiles, "dist")
}
