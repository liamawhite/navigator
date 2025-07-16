package ui

import (
	"io/fs"

	uiassets "github.com/liamawhite/navigator/ui"
)

// GetFileSystem returns a filesystem that can be used to serve the UI assets.
func GetFileSystem() (fs.FS, error) {
	return uiassets.GetFileSystem()
}
