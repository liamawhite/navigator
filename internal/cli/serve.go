package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"github.com/liamawhite/navigator/internal/grpc"
	"github.com/liamawhite/navigator/pkg/datastore/kubeconfig"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Navigator gRPC server",
	Long: `Start the Navigator gRPC server that provides APIs for service discovery 
and Envoy configuration analysis.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		kubeconfigPath, _ := cmd.Flags().GetString("kubeconfig")

		// Create datastore
		ds, err := kubeconfig.New(kubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to create datastore: %w", err)
		}

		// Create gRPC server
		server, err := grpc.NewServer(ds, port)
		if err != nil {
			return fmt.Errorf("failed to create gRPC server: %w", err)
		}

		// Start server in a goroutine
		serverErrChan := make(chan error, 1)
		go func() {
			serverErrChan <- server.Start()
		}()

		// Wait for interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		fmt.Printf("Navigator gRPC server listening on %s\n", server.Address())
		fmt.Printf("Navigator HTTP gateway listening on %s\n", server.HTTPAddress())

		select {
		case err := <-serverErrChan:
			return fmt.Errorf("server error: %w", err)
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
			server.Stop()
			return nil
		}
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")

	// Default kubeconfig path
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}
	serveCmd.Flags().StringP("kubeconfig", "k", defaultKubeconfig, "Path to kubeconfig file")
}
