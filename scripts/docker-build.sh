#!/bin/bash

# Docker Build Script with Optimization
# Automatically enables BuildKit and provides common build operations

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Enable BuildKit
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
export BUILDKIT_PROGRESS=auto

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Function to check BuildKit support
check_buildkit() {
    if ! docker buildx version > /dev/null 2>&1; then
        print_warning "BuildKit not available. Install with: docker buildx install"
        return 1
    fi
    print_success "BuildKit is enabled"
    return 0
}

# Function to show disk usage
show_disk_usage() {
    print_info "Docker disk usage:"
    docker system df
    echo ""
    print_info "Build cache usage:"
    docker buildx du 2>/dev/null || docker builder prune --help > /dev/null 2>&1
}

# Compose file combinations
COMPOSE_DEV="-f docker/compose.yml -f docker/compose.dev.yml"
COMPOSE_PROD="-f docker/compose.yml -f docker/compose.prod.yml"

# Function to build development environment
build_dev() {
    print_info "Building development environment..."
    docker compose $COMPOSE_DEV build "$@"
    print_success "Development build complete"
}

# Function to build production environment
build_prod() {
    print_info "Building production environment..."
    docker compose $COMPOSE_PROD build "$@"
    print_success "Production build complete"
}

# Function to rebuild without cache
build_no_cache() {
    print_warning "Building without cache (this will be slow)..."
    docker compose $COMPOSE_DEV build --no-cache "$@"
    print_success "No-cache build complete"
}

# Function to clean Docker resources
clean_docker() {
    print_warning "This will remove unused Docker resources"
    read -p "Continue? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Removing dangling images..."
        docker image prune -f

        print_info "Removing build cache (keeping recent)..."
        docker builder prune -f --keep-storage 10GB

        print_info "Removing stopped containers..."
        docker container prune -f

        print_success "Cleanup complete"
        show_disk_usage
    fi
}

# Function to clean everything (aggressive)
clean_all() {
    print_error "WARNING: This will remove ALL unused Docker resources including volumes!"
    read -p "Are you absolutely sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Removing all unused resources..."
        docker system prune -a --volumes -f
        docker builder prune -a -f
        print_success "Aggressive cleanup complete"
        show_disk_usage
    fi
}

# Function to show help
show_help() {
    cat << EOF
Docker Build Script - Optimized builds with BuildKit

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  dev              Build development environment (default)
  prod             Build production environment
  no-cache         Build without cache (slow, clean build)
  clean            Clean unused Docker resources (safe)
  clean-all        Remove ALL Docker resources (dangerous!)
  stats            Show Docker disk usage statistics
  help             Show this help message

Options:
  [service]        Specify service to build (backend, frontend, postgres)

Examples:
  $0                    # Build development environment
  $0 dev backend        # Build only backend service (dev)
  $0 prod               # Build production environment
  $0 no-cache frontend  # Rebuild frontend without cache
  $0 clean              # Clean up unused resources
  $0 stats              # Show disk usage

Environment:
  DOCKER_BUILDKIT=1    (enabled automatically)
  COMPOSE_DOCKER_CLI_BUILD=1    (enabled automatically)

For more information, see DOCKER_OPTIMIZATION.md
EOF
}

# Main script logic
main() {
    check_docker
    check_buildkit

    # Parse command
    COMMAND=${1:-dev}
    shift || true  # Remove first argument, continue if no more args

    case "$COMMAND" in
        dev|development)
            build_dev "$@"
            ;;
        prod|production)
            build_prod "$@"
            ;;
        no-cache|clean-build)
            build_no_cache "$@"
            ;;
        clean|cleanup)
            clean_docker
            ;;
        clean-all|purge)
            clean_all
            ;;
        stats|disk|usage)
            show_disk_usage
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
