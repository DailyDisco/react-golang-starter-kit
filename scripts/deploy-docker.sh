#!/bin/bash

# ============================================
# React + Go Starter Kit - Docker Deployment
# ============================================
# Deploy production containers to a Docker host

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${BLUE}ℹ${NC} $1"; }

# Default values
REGISTRY=""
TAG="latest"
PUSH=false
ENV_FILE=".env.prod"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --tag)
            TAG="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: ./deploy-docker.sh [options]"
            echo ""
            echo "Options:"
            echo "  --registry <url>   Docker registry URL (e.g., ghcr.io/username)"
            echo "  --tag <tag>        Image tag (default: latest)"
            echo "  --push             Push images to registry after build"
            echo "  --env-file <file>  Environment file to use (default: .env.prod)"
            echo "  --help             Show this help message"
            echo ""
            echo "Examples:"
            echo "  ./deploy-docker.sh --tag v1.0.0"
            echo "  ./deploy-docker.sh --registry ghcr.io/myorg --push"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_header "Docker Production Deployment"

# ============================================
# Pre-flight Checks
# ============================================

print_info "Running pre-flight checks..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi
print_success "Docker is installed"

# Check Docker Compose
if ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not available"
    exit 1
fi
print_success "Docker Compose is available"

# Check environment file
if [ ! -f "$ENV_FILE" ]; then
    print_warning "Environment file $ENV_FILE not found"
    if [ -f ".env.example" ]; then
        print_info "Creating $ENV_FILE from .env.example"
        cp .env.example "$ENV_FILE"
        print_warning "Please edit $ENV_FILE with production values before deploying!"
        exit 1
    else
        print_error "No .env.example found to create production config"
        exit 1
    fi
fi
print_success "Environment file found: $ENV_FILE"

# ============================================
# Build Images
# ============================================

print_header "Building Production Images"

# Determine image names
if [ -n "$REGISTRY" ]; then
    BACKEND_IMAGE="${REGISTRY}/backend:${TAG}"
    FRONTEND_IMAGE="${REGISTRY}/frontend:${TAG}"
else
    BACKEND_IMAGE="app-backend:${TAG}"
    FRONTEND_IMAGE="app-frontend:${TAG}"
fi

print_info "Building backend image: $BACKEND_IMAGE"
docker build -t "$BACKEND_IMAGE" -f backend/Dockerfile.prod backend/
print_success "Backend image built"

print_info "Building frontend image: $FRONTEND_IMAGE"
docker build -t "$FRONTEND_IMAGE" -f frontend/Dockerfile.prod frontend/
print_success "Frontend image built"

# ============================================
# Push to Registry (if requested)
# ============================================

if [ "$PUSH" = true ]; then
    if [ -z "$REGISTRY" ]; then
        print_error "Cannot push without --registry specified"
        exit 1
    fi

    print_header "Pushing Images to Registry"

    print_info "Pushing $BACKEND_IMAGE"
    docker push "$BACKEND_IMAGE"
    print_success "Backend image pushed"

    print_info "Pushing $FRONTEND_IMAGE"
    docker push "$FRONTEND_IMAGE"
    print_success "Frontend image pushed"
fi

# ============================================
# Deploy with Docker Compose
# ============================================

print_header "Deploying Services"

# Export image names for docker-compose
export BACKEND_IMAGE
export FRONTEND_IMAGE

print_info "Starting services..."
docker compose --env-file "$ENV_FILE" -f docker-compose.prod.yml up -d

# Wait for services to be healthy
print_info "Waiting for services to be healthy..."
sleep 10

# Check service status
if docker compose --env-file "$ENV_FILE" -f docker-compose.prod.yml ps | grep -q "healthy\|running"; then
    print_success "Services are running"
else
    print_error "Some services may not be healthy"
    docker compose --env-file "$ENV_FILE" -f docker-compose.prod.yml ps
fi

# ============================================
# Summary
# ============================================

print_header "Deployment Complete"

echo "Images:"
echo "  Backend:  $BACKEND_IMAGE"
echo "  Frontend: $FRONTEND_IMAGE"
echo ""
echo "Services:"
docker compose --env-file "$ENV_FILE" -f docker-compose.prod.yml ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""
print_info "View logs: docker compose -f docker-compose.prod.yml logs -f"
print_info "Stop services: docker compose -f docker-compose.prod.yml down"

exit 0
