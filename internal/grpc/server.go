package grpc

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/liamawhite/navigator/internal/ui"
	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/datastore"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/liamawhite/navigator/pkg/troubleshooting"
)

// Server wraps the gRPC server and HTTP gateway, providing methods to start and stop them.
type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	uiServer   *http.Server
	listener   net.Listener
	port       int
	serveUI    bool
}

// NewServer creates a new server with both gRPC and HTTP endpoints.
func NewServer(serviceDS datastore.ServiceDatastore, troubleshootingDS troubleshooting.ProxyDatastore, port int) (*Server, error) {
	return NewServerWithOptions(serviceDS, troubleshootingDS, port, false)
}

// NewServerWithUI creates a new server with gRPC, HTTP endpoints, and UI serving.
func NewServerWithUI(serviceDS datastore.ServiceDatastore, troubleshootingDS troubleshooting.ProxyDatastore, port int) (*Server, error) {
	return NewServerWithOptions(serviceDS, troubleshootingDS, port, true)
}

// NewServerWithOptions creates a new server with configurable UI serving.
func NewServerWithOptions(serviceDS datastore.ServiceDatastore, troubleshootingDS troubleshooting.ProxyDatastore, port int, serveUI bool) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	// Create gRPC server with logging interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor()),
		grpc.StreamInterceptor(logging.StreamServerInterceptor()),
	)

	// Register the ServiceRegistryService
	serviceRegistryServer := NewServiceRegistryServer(serviceDS)
	v1alpha1.RegisterServiceRegistryServiceServer(grpcServer, serviceRegistryServer)

	// Register the TroubleshootingService
	troubleshootingServer := NewTroubleshootingServer(troubleshootingDS)
	v1alpha1.RegisterTroubleshootingServiceServer(grpcServer, troubleshootingServer)

	// Enable reflection for easier debugging and testing
	reflection.Register(grpcServer)

	// Create HTTP gateway with logging middleware
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = v1alpha1.RegisterServiceRegistryServiceHandlerFromEndpoint(
		context.Background(), mux, fmt.Sprintf("localhost:%d", port), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register service registry gateway: %w", err)
	}

	err = v1alpha1.RegisterTroubleshootingServiceHandlerFromEndpoint(
		context.Background(), mux, fmt.Sprintf("localhost:%d", port), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register troubleshooting gateway: %w", err)
	}

	// Wrap the mux with logging middleware
	httpHandler := logging.HTTPMiddleware()(mux)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port+1),
		Handler:           httpHandler,
		ReadHeaderTimeout: 30 * time.Second,
	}

	var uiServer *http.Server
	if serveUI {
		uiFS, err := ui.GetFileSystem()
		if err != nil {
			return nil, fmt.Errorf("failed to get UI filesystem: %w", err)
		}

		uiHandler := createUIHandler(uiFS, port+1) // Pass API port for proxying
		uiServer = &http.Server{
			Addr:              fmt.Sprintf(":%d", port+2),
			Handler:           uiHandler,
			ReadHeaderTimeout: 30 * time.Second,
		}
	}

	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		uiServer:   uiServer,
		listener:   listener,
		port:       port,
		serveUI:    serveUI,
	}, nil
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
		defer file.Close()

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

// Start starts gRPC, HTTP, and optionally UI servers. This method blocks until the servers are stopped.
func (s *Server) Start() error {
	logger := logging.For(logging.ComponentServer)
	logger.Info("starting gRPC server", "port", s.port)
	logger.Info("starting HTTP gateway", "port", s.port+1)

	// Start HTTP server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Start UI server if enabled
	if s.serveUI && s.uiServer != nil {
		logger.Info("starting UI server", "port", s.port+2)
		go func() {
			if err := s.uiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("UI server error", "error", err)
			}
		}()
	}

	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops gRPC, HTTP, and UI servers.
func (s *Server) Stop() {
	logger := logging.For(logging.ComponentServer)
	logger.Info("stopping servers")
	s.grpcServer.GracefulStop()
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}
	if s.uiServer != nil {
		if err := s.uiServer.Shutdown(context.Background()); err != nil {
			logger.Error("UI server shutdown error", "error", err)
		}
	}
}

// Address returns the address the gRPC server is listening on.
func (s *Server) Address() string {
	return s.listener.Addr().String()
}

// HTTPAddress returns the address the HTTP gateway is listening on.
func (s *Server) HTTPAddress() string {
	return fmt.Sprintf(":%d", s.port+1)
}

// UIAddress returns the address the UI server is listening on.
func (s *Server) UIAddress() string {
	return fmt.Sprintf(":%d", s.port+2)
}

// UIEnabled returns whether the UI server is enabled.
func (s *Server) UIEnabled() bool {
	return s.serveUI
}
