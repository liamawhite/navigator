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

package ui

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/liamawhite/navigator/pkg/ui"
)

// Server represents a UI server that serves the Navigator web interface
type Server struct {
	server *http.Server
	port   int
}

// NewServer creates a new UI server
func NewServer(port int, apiPort int) (*Server, error) {
	// Get UI filesystem
	uiFS, err := ui.GetFileSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to get UI filesystem: %w", err)
	}

	// Create UI handler
	handler := createUIHandler(uiFS, apiPort)

	// Create HTTP server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 30 * time.Second,
	}

	return &Server{
		server: server,
		port:   port,
	}, nil
}

// Start starts the UI server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop gracefully stops the UI server
func (s *Server) Stop() error {
	return s.server.Shutdown(context.TODO())
}

// Address returns the address the UI server is listening on
func (s *Server) Address() string {
	return fmt.Sprintf(":%d", s.port)
}

// createUIHandler creates an HTTP handler for serving the embedded UI files and proxying API requests
func createUIHandler(uiFS fs.FS, apiPort int) http.Handler {
	// Create reverse proxy for API requests
	apiURL, _ := url.Parse(fmt.Sprintf("http://localhost:%d", apiPort))
	proxy := httputil.NewSingleHostReverseProxy(apiURL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proxy API requests to the HTTP gateway
		if strings.HasPrefix(r.URL.Path, "/api/") {
			proxy.ServeHTTP(w, r)
			return
		}

		// Serve UI files
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to serve the requested file
		file, err := uiFS.Open(path)
		if err != nil {
			// If file not found, serve index.html for SPA routing
			file, err = uiFS.Open("index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
		}
		defer func() {
			if err := file.Close(); err != nil {
				// Log error but don't fail the request
				logging.For("navctl-ui").Warn("failed to close file", "error", err)
			}
		}()

		// Set appropriate content type
		if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		// Copy file contents to response
		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
}
