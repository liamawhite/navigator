// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build test || integration || lint

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
