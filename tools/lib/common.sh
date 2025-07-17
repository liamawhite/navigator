#!/bin/bash

# Copyright 2025 Navigator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# common.sh - Shared utilities for Navigator development scripts

# Configuration
CLUSTER_NAME="navigator-demo"
KUBECONFIG_FILE="/tmp/navigator-demo-kubeconfig"
NAVIGATOR_PORT=8080
NAVIGATOR_HTTP_PORT=8081
FRONTEND_PORT=5173

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${BLUE}[$1]${NC} $2"
}

error() {
    echo -e "${RED}[$1 ERROR]${NC} $2" >&2
}

success() {
    echo -e "${GREEN}[$1]${NC} $2"
}

warn() {
    echo -e "${YELLOW}[$1]${NC} $2"
}

info() {
    echo -e "${CYAN}[$1]${NC} $2"
}

# Check if required tools are available
check_dependencies() {
    local missing_deps=()
    
    if ! command -v kind &> /dev/null; then
        missing_deps+=("kind")
    fi
    
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi
    
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        error "common" "Missing required dependencies: ${missing_deps[*]}"
        error "common" "Please install them and try again"
        return 1
    fi
    return 0
}

# Check if cluster exists and is running
cluster_exists() {
    kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"
}

# Get kubeconfig for the cluster
get_kubeconfig() {
    local component="${1:-common}"
    
    log "$component" "Getting kubeconfig for cluster ${CLUSTER_NAME}..."
    if ! kind get kubeconfig --name "${CLUSTER_NAME}" > "${KUBECONFIG_FILE}" 2>/dev/null; then
        error "$component" "Failed to get kubeconfig for cluster ${CLUSTER_NAME}"
        return 1
    fi
    success "$component" "Kubeconfig saved to ${KUBECONFIG_FILE}"
    return 0
}

# Verify cluster connectivity
verify_cluster_connectivity() {
    local component="${1:-common}"
    
    log "$component" "Verifying cluster connectivity..."
    if ! kubectl --kubeconfig "${KUBECONFIG_FILE}" cluster-info >/dev/null 2>&1; then
        error "$component" "Cannot connect to cluster - is it running?"
        return 1
    fi
    
    log "$component" "Connected to cluster '${CLUSTER_NAME}'"
    return 0
}

# Show cluster information
show_cluster_info() {
    local component="${1:-common}"
    
    # Show deployed services
    local services
    services=$(kubectl --kubeconfig "${KUBECONFIG_FILE}" get services -n demo --no-headers 2>/dev/null | wc -l || echo "0")
    if [ "$services" -gt 0 ]; then
        success "$component" "Found $services services in demo namespace"
        kubectl --kubeconfig "${KUBECONFIG_FILE}" get services -n demo
    else
        warn "$component" "No services found in demo namespace - cluster may be empty"
    fi
}

# Wait for a service to be ready
wait_for_service() {
    local service_name="$1"
    local port="$2"
    local timeout="${3:-30}"
    local component="${4:-common}"
    
    info "$component" "Waiting for $service_name on port $port..."
    
    local count=0
    while [ $count -lt $timeout ]; do
        if curl -s "http://localhost:$port" >/dev/null 2>&1; then
            success "$component" "$service_name is ready!"
            return 0
        fi
        sleep 1
        count=$((count + 1))
        if [ $((count % 5)) -eq 0 ]; then
            info "$component" "Still waiting for $service_name... ($count/${timeout}s)"
        fi
    done
    
    error "$component" "$service_name failed to start within ${timeout}s"
    return 1
}

# Build Navigator if needed
build_navigator_if_needed() {
    local component="${1:-common}"
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Check if UI needs to be built (if ui/dist doesn't exist or UI source files are newer)
    local ui_needs_build=false
    if [ ! -d "ui/dist" ]; then
        ui_needs_build=true
    else
        # Check if any UI source files are newer than dist
        if [ -n "$(find ui/src -name '*.tsx' -o -name '*.ts' -o -name '*.jsx' -o -name '*.js' -o -name '*.css' -o -name '*.html' -newer ui/dist 2>/dev/null)" ]; then
            ui_needs_build=true
        fi
        # Also check package.json and vite.config.ts
        if [ "ui/package.json" -nt "ui/dist" ] || [ "ui/vite.config.ts" -nt "ui/dist" ]; then
            ui_needs_build=true
        fi
    fi
    
    # Build UI if needed
    if [ "$ui_needs_build" = true ]; then
        log "$component" "Building UI..."
        if ! (cd ui && npm ci && npm run build); then
            error "$component" "Failed to build UI"
            return 1
        fi
        success "$component" "UI built successfully"
    fi
    
    # Check if Navigator binary needs to be built
    if [ ! -f "bin/navigator" ] || [ "bin/navigator" -ot "cmd/navigator/main.go" ] || [ "$ui_needs_build" = true ]; then
        log "$component" "Building Navigator..."
        if ! go build -o bin/navigator cmd/navigator/main.go; then
            error "$component" "Failed to build Navigator"
            return 1
        fi
        success "$component" "Navigator built successfully"
    fi
    return 0
}

# Clean up kubeconfig file
cleanup_kubeconfig() {
    if [ -f "${KUBECONFIG_FILE}" ]; then
        rm -f "${KUBECONFIG_FILE}"
    fi
}

# Show error message for missing cluster
show_cluster_missing_error() {
    local component="${1:-common}"
    
    error "$component" "Kind cluster '${CLUSTER_NAME}' does not exist!"
    error "$component" "Please run one of the following first:"
    error "$component" "  make demo                    # Basic demo setup"
    error "$component" "  ./navigator demo --scenario <scenario>  # Specific scenario"
    error "$component" ""
    error "$component" "Available scenarios:"
    if [ -f "bin/navigator" ]; then
        bin/navigator demo --list 2>/dev/null | grep -E "^  [a-z-]+" | head -6
    fi
}

# Auto-setup cluster if it doesn't exist (for make dev)
auto_setup_cluster_if_needed() {
    local component="${1:-common}"
    
    if ! cluster_exists; then
        warn "$component" "Kind cluster '${CLUSTER_NAME}' does not exist, creating it..."
        
        # Build navigator if needed
        if ! build_navigator_if_needed "$component"; then
            return 1
        fi
        
        # Create cluster with basic scenario
        log "$component" "Setting up demo cluster with microservice topology..."
        if ! bin/navigator demo --scenario microservice-topology --cleanup-on-exit=false; then
            error "$component" "Failed to create demo cluster"
            return 1
        fi
        
        success "$component" "Demo cluster created successfully!"
    else
        success "$component" "Found existing Kind cluster '${CLUSTER_NAME}'"
    fi
    return 0
}
