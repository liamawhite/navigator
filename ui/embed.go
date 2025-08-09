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

//go:build !test && !integration && !lint && !docs

package ui

//go:generate sh -c "npm ci || npm install"
//go:generate npm run build

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
