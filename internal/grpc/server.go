package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/datastore"
)

// Server wraps the gRPC server and HTTP gateway, providing methods to start and stop them.
type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	listener   net.Listener
	port       int
}

// NewServer creates a new server with both gRPC and HTTP endpoints.
func NewServer(ds datastore.ServiceDatastore, port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	// Register the ServiceRegistryService
	serviceRegistryServer := NewServiceRegistryServer(ds)
	v1alpha1.RegisterServiceRegistryServiceServer(grpcServer, serviceRegistryServer)

	// Enable reflection for easier debugging and testing
	reflection.Register(grpcServer)

	// Create HTTP gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = v1alpha1.RegisterServiceRegistryServiceHandlerFromEndpoint(
		context.Background(), mux, fmt.Sprintf("localhost:%d", port), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port+1),
		Handler: mux,
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
	log.Printf("Starting gRPC server on port %d", s.port)
	log.Printf("Starting HTTP gateway on port %d", s.port+1)

	// Start HTTP server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops both gRPC and HTTP servers.
func (s *Server) Stop() {
	log.Println("Stopping servers")
	s.grpcServer.GracefulStop()
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
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
