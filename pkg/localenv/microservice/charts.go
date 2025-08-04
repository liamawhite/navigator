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

package microservice

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// ChartFiles contains the embedded microservice Helm chart
//
//go:embed chart/charts/microservice.tgz
var ChartFiles embed.FS

// GetChartTar returns the raw tar.gz data for the microservice chart
func GetChartTar() ([]byte, error) {
	data, err := fs.ReadFile(ChartFiles, "chart/charts/microservice.tgz")
	if err != nil {
		return nil, fmt.Errorf("failed to read microservice chart: %w", err)
	}

	return data, nil
}

// ExtractChart extracts the microservice chart tar to memory and returns filesystem
func ExtractChart() (fs.FS, error) {
	tarData, err := GetChartTar()
	if err != nil {
		return nil, err
	}

	// Create a memory filesystem from the tar
	memFS := make(map[string][]byte)

	if err := extractTarToMemory(tarData, memFS); err != nil {
		return nil, fmt.Errorf("failed to extract chart: %w", err)
	}

	return &memoryFS{files: memFS}, nil
}

// GetChartFile reads a specific file from within the microservice chart tar
func GetChartFile(filePath string) ([]byte, error) {
	chartFS, err := ExtractChart()
	if err != nil {
		return nil, err
	}

	data, err := fs.ReadFile(chartFS, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s from chart: %w", filePath, err)
	}

	return data, nil
}

// extractTarToMemory extracts a tar.gz to an in-memory map
func extractTarToMemory(tarData []byte, memFS map[string][]byte) error {
	// Create gzip reader
	gzReader, err := gzip.NewReader(strings.NewReader(string(tarData)))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			// Log error but continue execution - error ignored intentionally
			_ = closeErr
		}
	}()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Track the root directory to strip it
	var rootDir string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Get the root directory name from the first entry
		if rootDir == "" {
			parts := strings.Split(header.Name, "/")
			if len(parts) > 0 {
				rootDir = parts[0] + "/"
			}
		}

		// Strip the root directory from the path
		relativePath := strings.TrimPrefix(header.Name, rootDir)
		if relativePath == "" {
			continue // Skip the root directory itself
		}

		// Only handle regular files
		if header.Typeflag == tar.TypeReg {
			// Read file content
			content, err := io.ReadAll(tarReader)
			if err != nil {
				return err
			}
			memFS[relativePath] = content
		}
	}

	return nil
}

// memoryFS implements fs.FS for in-memory files
type memoryFS struct {
	files map[string][]byte
}

func (m *memoryFS) Open(name string) (fs.File, error) {
	if name == "." {
		return &memoryDir{fs: m, path: "."}, nil
	}

	content, exists := m.files[name]
	if !exists {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return &memoryFile{
		name:    filepath.Base(name),
		content: content,
		reader:  strings.NewReader(string(content)),
	}, nil
}

// memoryFile implements fs.File for in-memory files
type memoryFile struct {
	name    string
	content []byte
	reader  *strings.Reader
}

func (f *memoryFile) Stat() (fs.FileInfo, error) {
	return &memoryFileInfo{name: f.name, size: int64(len(f.content))}, nil
}

func (f *memoryFile) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f *memoryFile) Close() error {
	return nil
}

// memoryDir implements fs.File for directories
type memoryDir struct {
	fs   *memoryFS
	path string
}

func (d *memoryDir) Stat() (fs.FileInfo, error) {
	return &memoryFileInfo{name: filepath.Base(d.path), size: 0, isDir: true}, nil
}

func (d *memoryDir) Read(p []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.path, Err: fs.ErrInvalid}
}

func (d *memoryDir) Close() error {
	return nil
}

// memoryFileInfo implements fs.FileInfo
type memoryFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (i *memoryFileInfo) Name() string       { return i.name }
func (i *memoryFileInfo) Size() int64        { return i.size }
func (i *memoryFileInfo) Mode() fs.FileMode  { return 0644 }
func (i *memoryFileInfo) ModTime() time.Time { return time.Time{} }
func (i *memoryFileInfo) IsDir() bool        { return i.isDir }
func (i *memoryFileInfo) Sys() interface{}   { return nil }
