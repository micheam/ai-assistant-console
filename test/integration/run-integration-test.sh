#!/bin/bash
#
# Integration Test Runner for AICO Installation
# Tests the installation script in a container (Docker or Apple Container)
#
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Test configuration
IMAGE_NAME="aico-integration-test"
CONTAINER_NAME="aico-test-$$"

# Container runtime (will be auto-detected or set via CONTAINER_RUNTIME env var)
CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-}"

# Logging functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Detect container runtime
detect_container_runtime() {
    if [[ -n "${CONTAINER_RUNTIME}" ]]; then
        log_debug "Using container runtime from environment: ${CONTAINER_RUNTIME}"
        return 0
    fi

    # Try Apple Container first (macOS specific)
    if command -v container &> /dev/null; then
        CONTAINER_RUNTIME="container"
        log_info "Detected Apple Container"
        return 0
    fi

    # Fall back to Docker
    if command -v docker &> /dev/null; then
        CONTAINER_RUNTIME="docker"
        log_info "Detected Docker"
        return 0
    fi

    log_error "No container runtime found"
    log_error "Please install one of the following:"
    log_error "  - Apple Container (macOS): https://github.com/apple/container"
    log_error "  - Docker: https://www.docker.com/"
    return 1
}

# Build container image
build_image() {
    log_info "Building container image..."
    if ! ${CONTAINER_RUNTIME} build -t "${IMAGE_NAME}" -f "${SCRIPT_DIR}/Dockerfile" "${PROJECT_ROOT}"; then
        log_error "Failed to build container image"
        return 1
    fi
    log_success "Container image built successfully"
}

# Start container
start_container() {
    log_info "Starting container..."
    if ! ${CONTAINER_RUNTIME} run -d --name "${CONTAINER_NAME}" "${IMAGE_NAME}" sleep infinity; then
        log_error "Failed to start container"
        return 1
    fi
    log_success "Container started: ${CONTAINER_NAME}"
}

# Execute command in container
exec_in_container() {
    ${CONTAINER_RUNTIME} exec "${CONTAINER_NAME}" "$@"
}

# Get container logs
get_container_logs() {
    ${CONTAINER_RUNTIME} logs "${CONTAINER_NAME}"
}

# Check if container exists
container_exists() {
    if [[ "${CONTAINER_RUNTIME}" == "docker" ]]; then
        docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
    else
        # Apple Container uses 'container list'
        container list 2>/dev/null | grep -q "${CONTAINER_NAME}"
    fi
}

# Remove container
remove_container() {
    if [[ "${CONTAINER_RUNTIME}" == "docker" ]]; then
        docker rm -f "${CONTAINER_NAME}" > /dev/null 2>&1 || true
    else
        # Apple Container uses 'container delete'
        container delete "${CONTAINER_NAME}" > /dev/null 2>&1 || true
    fi
}

# Cleanup function
cleanup() {
    if container_exists; then
        log_info "Cleaning up container ${CONTAINER_NAME}..."
        remove_container
    fi
}

# Register cleanup on exit
trap cleanup EXIT

# Main test function
run_integration_test() {
    log_info "Starting AICO integration test..."
    log_info "Container runtime: ${CONTAINER_RUNTIME}"
    echo ""

    # Build container image
    if ! build_image; then
        return 1
    fi
    echo ""

    # Run installation test
    log_info "Running installation test in container..."

    # Start container
    if ! start_container; then
        return 1
    fi

    # Execute installation script
    log_info "Executing install.sh..."
    if ! exec_in_container bash -c "bash /home/testuser/install.sh"; then
        log_error "Installation script failed"
        get_container_logs
        return 1
    fi
    log_success "Installation completed successfully"
    echo ""

    # Test 1: Check if binary exists
    log_info "Test 1: Checking if binary exists..."
    if ! exec_in_container test -f /home/testuser/.local/bin/aico; then
        log_error "Binary not found at /home/testuser/.local/bin/aico"
        return 1
    fi
    log_success "Binary exists"

    # Test 2: Check if binary is executable
    log_info "Test 2: Checking if binary is executable..."
    if ! exec_in_container test -x /home/testuser/.local/bin/aico; then
        log_error "Binary is not executable"
        return 1
    fi
    log_success "Binary is executable"

    # Test 3: Run version command
    log_info "Test 3: Running 'aico --version'..."
    version_output=$(exec_in_container /home/testuser/.local/bin/aico --version)
    echo "  Version output: ${version_output}"

    if [[ -z "${version_output}" ]]; then
        log_error "Version command returned empty output"
        return 1
    fi
    log_success "Version command succeeded"

    # Test 4: Run help command
    log_info "Test 4: Running 'aico --help'..."
    if ! exec_in_container /home/testuser/.local/bin/aico --help > /dev/null 2>&1; then
        log_error "Help command failed"
        return 1
    fi
    log_success "Help command succeeded"

    echo ""
    log_success "All integration tests passed!"
    return 0
}

# Detect container runtime first
if ! detect_container_runtime; then
    exit 1
fi
echo ""

# Run the test
if run_integration_test; then
    exit 0
else
    log_error "Integration test failed"
    exit 1
fi
