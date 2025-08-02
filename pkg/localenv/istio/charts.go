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

package istio

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ChartFiles contains the embedded Istio Helm chart tar files
//
//go:embed charts charts/* charts/*/*.tgz
var ChartFiles embed.FS

// GetChartFS returns the embedded chart filesystem
func GetChartFS() fs.FS {
	return ChartFiles
}

// ListVersions returns available Istio versions from embedded tars
func ListVersions() ([]string, error) {
	entries, err := fs.ReadDir(ChartFiles, "charts")
	if err != nil {
		return nil, fmt.Errorf("failed to read charts directory: %w", err)
	}

	var versions []string
	versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if versionRegex.MatchString(entry.Name()) {
			versions = append(versions, entry.Name())
		}
	}

	sort.Strings(versions)
	return versions, nil
}

// ListCharts returns available chart names for a specific version
func ListCharts(version string) ([]string, error) {
	versionDir := filepath.Join("charts", version)
	entries, err := fs.ReadDir(ChartFiles, versionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts directory for version %s: %w", version, err)
	}

	var charts []string
	suffix := fmt.Sprintf("-%s.tgz", version)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), suffix) {
			chartName := strings.TrimSuffix(entry.Name(), suffix)
			charts = append(charts, chartName)
		}
	}

	sort.Strings(charts)
	return charts, nil
}

// GetChartTar returns the raw tar.gz data for a specific chart
func GetChartTar(version, chartName string) ([]byte, error) {
	tarFileName := fmt.Sprintf("%s-%s.tgz", chartName, version)
	tarPath := filepath.Join("charts", version, tarFileName)

	data, err := fs.ReadFile(ChartFiles, tarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart tar %s: %w", tarFileName, err)
	}

	return data, nil
}

// ExtractChart extracts a chart tar to memory and returns filesystem
func ExtractChart(version, chartName string) (fs.FS, error) {
	tarData, err := GetChartTar(version, chartName)
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

// GetChartFile reads a specific file from within a chart tar
func GetChartFile(version, chartName, filePath string) ([]byte, error) {
	chartFS, err := ExtractChart(version, chartName)
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
