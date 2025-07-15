package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/datastore"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/liamawhite/navigator/pkg/troubleshooting"
)

// Server wraps the gRPC server and HTTP gateway, providing methods to start and stop them.
type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	listener   net.Listener
	port       int
}

// NewServer creates a new server with both gRPC and HTTP endpoints.
func NewServer(serviceDS datastore.ServiceDatastore, troubleshootingDS troubleshooting.ProxyDatastore, port int) (*Server, error) {
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

	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		listener:   listener,
		port:       port,
	}, nil
}

// Start starts both gRPC and HTTP servers. This method blocks until the servers are stopped.
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

	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops both gRPC and HTTP servers.
func (s *Server) Stop() {
	logger := logging.For(logging.ComponentServer)
	logger.Info("stopping servers")
	s.grpcServer.GracefulStop()
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
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
