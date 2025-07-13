#!/bin/bash

# dev-full-stack.sh - Start Navigator development environment with cluster + frontend
# This script coordinates starting Navigator with Kind cluster and the Vite frontend

set -euo pipefail

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common utilities
source "${SCRIPT_DIR}/lib/common.sh"

COMPONENT="dev-full-stack"

# Track background processes for cleanup
NAVIGATOR_PID=""
FRONTEND_PID=""

# Cleanup function for graceful shutdown
cleanup() {
    log "$COMPONENT" "Shutting down development environment..."
    
    if [ -n "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
        log "$COMPONENT" "Stopping frontend (PID: $FRONTEND_PID)..."
        kill -TERM "$FRONTEND_PID" 2>/dev/null || true
        wait "$FRONTEND_PID" 2>/dev/null || true
    fi
    
    if [ -n "$NAVIGATOR_PID" ] && kill -0 "$NAVIGATOR_PID" 2>/dev/null; then
        log "$COMPONENT" "Stopping Navigator server (PID: $NAVIGATOR_PID)..."
        kill -TERM "$NAVIGATOR_PID" 2>/dev/null || true
        wait "$NAVIGATOR_PID" 2>/dev/null || true
    fi
    
    cleanup_kubeconfig
    success "$COMPONENT" "Development environment stopped"
}

# Set up signal handlers
trap cleanup EXIT INT TERM

# Check if cluster exists (wrapper for common function)
check_cluster() {
    if ! cluster_exists; then
        show_cluster_missing_error "$COMPONENT"
        exit 1
    fi
    success "$COMPONENT" "Found Kind cluster '${CLUSTER_NAME}'"
}

# Start Navigator server in background
start_navigator() {
    log "$COMPONENT" "Starting Navigator server..."
    
    # Run the dev-with-cluster script in background
    ./tools/dev-with-cluster.sh &
    NAVIGATOR_PID=$!
    
    # Wait for Navigator to be ready
    if ! wait_for_service "Navigator" "$NAVIGATOR_HTTP_PORT" 30 "$COMPONENT"; then
        error "$COMPONENT" "Navigator failed to start"
        return 1
    fi
    
    success "$COMPONENT" "Navigator server running (PID: $NAVIGATOR_PID)"
    info "$COMPONENT" "  gRPC API: http://localhost:$NAVIGATOR_PORT"
    info "$COMPONENT" "  HTTP API: http://localhost:$NAVIGATOR_HTTP_PORT"
}

# Start frontend in background
start_frontend() {
    log "$COMPONENT" "Starting frontend development server..."
    
    # Change to UI directory and start Vite
    cd ui
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules/.package-json.hash" ]; then
        log "$COMPONENT" "Installing frontend dependencies..."
        npm install
        # Create a hash file to track when we last installed
        echo "$(sha256sum package.json)" > "node_modules/.package-json.hash"
    fi
    
    # Start Vite dev server in background
    npm run dev &
    FRONTEND_PID=$!
    
    # Go back to root directory
    cd ..
    
    # Wait for frontend to be ready
    if ! wait_for_service "Frontend" "$FRONTEND_PORT" 30 "$COMPONENT"; then
        error "$COMPONENT" "Frontend failed to start"
        return 1
    fi
    
    success "$COMPONENT" "Frontend server running (PID: $FRONTEND_PID)"
    info "$COMPONENT" "  Frontend: http://localhost:$FRONTEND_PORT"
}

# Display final information
show_ready_message() {
    echo
    echo "================================================================"
    success "$COMPONENT" "🚀 Full-stack development environment is ready!"
    echo "================================================================"
    echo
    info "$COMPONENT" "Services:"
    info "$COMPONENT" "  📱 Frontend:    http://localhost:$FRONTEND_PORT"
    info "$COMPONENT" "  🔧 Navigator:   http://localhost:$NAVIGATOR_HTTP_PORT"
    info "$COMPONENT" "  🔌 gRPC API:    localhost:$NAVIGATOR_PORT"
    echo
    info "$COMPONENT" "Cluster: $CLUSTER_NAME"
    
    # Show deployed services
    local services
    services=$(kubectl get services -n demo --no-headers 2>/dev/null | wc -l || echo "0")
    if [ "$services" -gt 0 ]; then
        info "$COMPONENT" "Services in demo namespace: $services"
    fi
    
    echo
    warn "$COMPONENT" "Press Ctrl+C to stop all services and exit"
    echo "================================================================"
}

# Main execution
main() {
    log "$COMPONENT" "Starting full-stack Navigator development environment"
    
    # Check dependencies and cluster
    check_cluster
    
    # Start services
    start_navigator
    start_frontend
    
    # Show ready message
    show_ready_message
    
    # Wait for user to stop (Ctrl+C will trigger cleanup)
    wait
}

# Run main function
main "$@"