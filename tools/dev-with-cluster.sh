#!/bin/bash

# dev-with-cluster.sh - Start Navigator development environment with Kind cluster
# This script coordinates setting up a Kind cluster, deploying services, and starting Navigator

set -euo pipefail

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common utilities
source "${SCRIPT_DIR}/lib/common.sh"

COMPONENT="dev-cluster"

# Start Navigator server
start_navigator() {
    log "$COMPONENT" "Starting Navigator server..."
    info "$COMPONENT" "  gRPC API: localhost:${NAVIGATOR_PORT}"
    info "$COMPONENT" "  HTTP API: localhost:${NAVIGATOR_HTTP_PORT}"
    
    # Build Navigator if needed
    if ! build_navigator_if_needed "$COMPONENT"; then
        exit 1
    fi
    
    # Start Navigator server
    exec bin/navigator serve --kubeconfig "${KUBECONFIG_FILE}" --port "${NAVIGATOR_PORT}"
}

# Cleanup function for graceful shutdown
cleanup() {
    log "$COMPONENT" "Cleaning up..."
    cleanup_kubeconfig
}

# Set up signal handlers
trap cleanup EXIT INT TERM

# Main execution
main() {
    log "$COMPONENT" "Starting Navigator development environment with Kind cluster"
    
    # Check dependencies
    if ! check_dependencies; then
        exit 1
    fi
    
    # Check if cluster exists
    if ! cluster_exists; then
        show_cluster_missing_error "$COMPONENT"
        exit 1
    fi
    
    # Get kubeconfig
    if ! get_kubeconfig "$COMPONENT"; then
        error "$COMPONENT" "Failed to get kubeconfig - is the cluster running?"
        exit 1
    fi
    
    # Verify cluster is accessible
    if ! verify_cluster_connectivity "$COMPONENT"; then
        exit 1
    fi
    
    # Show cluster info
    show_cluster_info "$COMPONENT"
    
    echo
    success "$COMPONENT" "🚀 Starting Navigator server..."
    
    # Start Navigator (this will exec, so script ends here)
    start_navigator
}

# Run main function
main "$@"