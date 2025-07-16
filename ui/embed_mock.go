//go:build test || integration

package ui

import (
	"io/fs"
	"testing/fstest"
)

// GetFileSystem returns a mock filesystem for CI/testing.
// This provides a simple in-memory filesystem that satisfies the same interface
// as the production embedded filesystem, allowing tests to run without requiring
// the UI build artifacts.
func GetFileSystem() (fs.FS, error) {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Navigator</title>
</head>
<body>
    <div id="root">Navigator UI Mock for Testing</div>
</body>
</html>`),
		},
		"navigator.svg": &fstest.MapFile{
			Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`),
		},
		"assets/index.js": &fstest.MapFile{
			Data: []byte(`console.log("Mock Navigator UI");`),
		},
		"assets/index.css": &fstest.MapFile{
			Data: []byte(`body { font-family: sans-serif; }`),
		},
	}, nil
}
