#!/bin/bash

# ============================================
# React + Go Starter Kit - Blue-Green Deployment
# ============================================
# Zero-downtime deployments using blue-green strategy
# with graceful shutdown, fast rollback, and auto-recovery
#
# Usage:
#   ./deploy-bluegreen.sh              # Auto-deploy to inactive environment
#   ./deploy-bluegreen.sh --switch     # Switch traffic without rebuild
#   ./deploy-bluegreen.sh --rollback   # Rollback to previous environment
#   ./deploy-bluegreen.sh --status     # Show current state

set -e

# ============================================
# Configuration
# ============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
STATE_FILE="${PROJECT_DIR}/.bluegreen-state"
ENV_FILE="${PROJECT_DIR}/.env.prod"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Health check settings (optimized for faster feedback)
HEALTH_CHECK_RETRIES=15
HEALTH_CHECK_INTERVAL=2
HEALTH_CHECK_TIMEOUT=5
DEEP_HEALTH_ENDPOINT="/health/ready"

# Graceful shutdown settings
GRACEFUL_SHUTDOWN_TIMEOUT=30

# Post-switch validation settings
POST_SWITCH_CHECKS=5
POST_SWITCH_INTERVAL=2
POST_SWITCH_THRESHOLD=3

# Container prefix (from PROJECT_NAME in .env.prod, loaded later via source)
# Will be set after sourcing env file
CONTAINER_PREFIX=""

# Image names for rollback caching
IMAGE_NAME="docker-backend"
ROLLBACK_TAG="rollback"

# ============================================
# Helper Functions
# ============================================

print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${CYAN}ℹ${NC} $1"; }

# Compose command with production files (blue-green deployment)
compose_cmd() {
    docker compose --env-file "$ENV_FILE" \
        -f "${PROJECT_DIR}/docker/compose.yml" \
        -f "${PROJECT_DIR}/docker/compose.prod.yml" \
        "$@"
}

# ============================================
# Graceful Shutdown
# ============================================

graceful_stop() {
    local env="$1"
    local container="${CONTAINER_PREFIX}-backend-${env}"

    if docker ps --filter "name=${container}" --filter "status=running" -q | grep -q .; then
        print_info "Gracefully stopping ${container} (${GRACEFUL_SHUTDOWN_TIMEOUT}s drain)..."
        docker stop --time=${GRACEFUL_SHUTDOWN_TIMEOUT} "$container" 2>/dev/null || true
        docker rm -f "$container" 2>/dev/null || true
        print_success "Backend ${env} stopped gracefully"
    else
        print_info "Backend ${env} not running, skipping stop"
    fi
}

# ============================================
# Image Retention for Fast Rollback
# ============================================

tag_current_image_for_rollback() {
    local current_image
    current_image=$(docker images -q "${IMAGE_NAME}:latest" 2>/dev/null || true)

    if [ -n "$current_image" ]; then
        print_info "Tagging current image for fast rollback..."
        docker tag "${IMAGE_NAME}:latest" "${IMAGE_NAME}:${ROLLBACK_TAG}" 2>/dev/null || true
        print_success "Rollback image cached: ${IMAGE_NAME}:${ROLLBACK_TAG}"
    fi
}

has_rollback_image() {
    docker image inspect "${IMAGE_NAME}:${ROLLBACK_TAG}" &>/dev/null
}

use_rollback_image() {
    if has_rollback_image; then
        print_info "Using cached rollback image (fast rollback)"
        docker tag "${IMAGE_NAME}:${ROLLBACK_TAG}" "${IMAGE_NAME}:latest"
        return 0
    fi
    return 1
}

# ============================================
# State Management
# ============================================

init_state() {
    if [ ! -f "$STATE_FILE" ]; then
        cat > "$STATE_FILE" << EOF
ACTIVE_ENV=blue
PREVIOUS_ENV=
LAST_DEPLOY=$(date -Iseconds)
EOF
        print_info "Initialized state file with blue as active"
    fi
}

read_state() {
    init_state
    source "$STATE_FILE"
}

write_state() {
    local active="$1"
    local previous="$2"
    cat > "$STATE_FILE" << EOF
ACTIVE_ENV=${active}
PREVIOUS_ENV=${previous}
LAST_DEPLOY=$(date -Iseconds)
EOF
}

get_inactive_env() {
    if [ "$ACTIVE_ENV" = "blue" ]; then
        echo "green"
    else
        echo "blue"
    fi
}

# ============================================
# Health Checks
# ============================================

wait_for_healthy() {
    local service="$1"
    local container_name="${CONTAINER_PREFIX}-${service}"
    local retries=$HEALTH_CHECK_RETRIES

    print_info "Waiting for ${service} to be healthy..."

    while [ $retries -gt 0 ]; do
        local status
        status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "not_found")

        if [ "$status" = "healthy" ]; then
            print_success "${service} is healthy"
            return 0
        elif [ "$status" = "unhealthy" ]; then
            print_error "${service} is unhealthy"
            return 1
        fi

        echo -n "."
        sleep $HEALTH_CHECK_INTERVAL
        retries=$((retries - 1))
    done

    echo ""
    print_error "${service} health check timed out"
    return 1
}

# Deep health check using /health/ready endpoint (verifies DB + cache)
deep_health_check() {
    local env="$1"
    local container="${CONTAINER_PREFIX}-backend-${env}"
    local retries=$HEALTH_CHECK_RETRIES

    # Get the port the container is listening on
    local port
    port=$(docker port "$container" 8080 2>/dev/null | head -1 | cut -d: -f2 || echo "")

    if [ -z "$port" ]; then
        # Container might be using internal networking, try the standard port
        port="${BACKEND_PORT:-8080}"
    fi

    print_info "Running deep health check on port ${port}..."

    while [ $retries -gt 0 ]; do
        local response
        response=$(curl -sf --max-time $HEALTH_CHECK_TIMEOUT \
            "http://localhost:${port}${DEEP_HEALTH_ENDPOINT}" 2>/dev/null || echo "")

        if echo "$response" | grep -q '"status":"healthy"'; then
            print_success "Deep health check passed (DB + cache verified)"
            return 0
        elif echo "$response" | grep -q '"status":"degraded"'; then
            print_warning "Deep health check: degraded (cache unavailable)"
            return 0  # Degraded is acceptable - app works without cache
        fi

        echo -n "."
        sleep $HEALTH_CHECK_INTERVAL
        retries=$((retries - 1))
    done

    echo ""
    print_error "Deep health check failed"
    return 1
}

check_env_healthy() {
    local env="$1"

    # First: wait for Docker health check to pass
    if ! wait_for_healthy "backend-${env}"; then
        return 1
    fi

    # Second: run deep health check (DB + cache connectivity)
    if ! deep_health_check "$env"; then
        return 1
    fi

    return 0
}

# Post-switch validation (runs multiple checks after traffic switch)
post_switch_validation() {
    local port="${BACKEND_PORT:-8080}"
    local checks_passed=0

    print_info "Post-switch health validation (${POST_SWITCH_CHECKS} checks)..."

    for i in $(seq 1 $POST_SWITCH_CHECKS); do
        sleep $POST_SWITCH_INTERVAL
        local response
        response=$(curl -sf --max-time 3 "http://localhost:${port}${DEEP_HEALTH_ENDPOINT}" 2>/dev/null || echo "")

        if echo "$response" | grep -qE '"status":"(healthy|degraded)"'; then
            checks_passed=$((checks_passed + 1))
            echo -n "${GREEN}✓${NC}"
        else
            echo -n "${RED}✗${NC}"
        fi
    done
    echo ""

    if [ $checks_passed -lt $POST_SWITCH_THRESHOLD ]; then
        print_error "Post-switch validation failed (${checks_passed}/${POST_SWITCH_CHECKS} checks passed, need ${POST_SWITCH_THRESHOLD})"
        return 1
    fi

    print_success "Post-switch validation passed (${checks_passed}/${POST_SWITCH_CHECKS} checks)"
    return 0
}

# ============================================
# Deployment Functions
# ============================================

start_infrastructure() {
    print_info "Starting shared infrastructure (postgres, dragonfly)..."
    compose_cmd up -d postgres dragonfly

    # Wait for postgres
    local retries=30
    while [ $retries -gt 0 ]; do
        if docker exec ${CONTAINER_PREFIX}-postgres-prod pg_isready -U "${DB_USER:-devuser}" >/dev/null 2>&1; then
            print_success "PostgreSQL is ready"
            break
        fi
        sleep 2
        retries=$((retries - 1))
    done

    if [ $retries -eq 0 ]; then
        print_error "PostgreSQL failed to start"
        exit 1
    fi
}

deploy_env() {
    local env="$1"
    local use_cache="${2:-false}"

    print_header "Deploying ${env} backend"

    if [ "$use_cache" = "true" ] && use_rollback_image; then
        print_info "Starting backend-${env} from cached image..."
        compose_cmd up -d "backend-${env}"
    else
        print_info "Building and starting backend-${env}..."
        compose_cmd up -d --build "backend-${env}"
    fi

    if ! check_env_healthy "$env"; then
        print_error "Deployment failed - backend-${env} is not healthy"
        print_info "Stopping failed backend-${env}..."
        graceful_stop "$env"
        return 1
    fi

    print_success "Backend ${env} deployed successfully"
    return 0
}

switch_traffic() {
    local target_env="$1"
    local current_env
    current_env=$([ "$target_env" = "blue" ] && echo "green" || echo "blue")

    print_header "Switching traffic to ${target_env}"

    # Gracefully stop current backend (traffic will go to target which has exposed port)
    print_info "Gracefully stopping backend-${current_env}..."
    graceful_stop "$current_env"

    print_success "Traffic switched to ${target_env}"
    print_info "Backend available at port ${BACKEND_PORT:-8080}"
}

# ============================================
# Auto-Rollback
# ============================================

auto_rollback() {
    local failed_env="$1"
    local rollback_env="$2"

    print_warning "AUTO-ROLLBACK: Reverting to ${rollback_env}..."

    # Try to use cached image first
    if has_rollback_image; then
        use_rollback_image
    fi

    # Start the rollback environment
    compose_cmd up -d "backend-${rollback_env}"

    if wait_for_healthy "backend-${rollback_env}"; then
        graceful_stop "$failed_env"
        write_state "$rollback_env" "$failed_env"
        print_success "Auto-rollback complete - ${rollback_env} is now active"
        return 0
    else
        print_error "Auto-rollback failed - manual intervention required"
        return 1
    fi
}

# ============================================
# Main Commands
# ============================================

cmd_deploy() {
    read_state
    local target_env
    target_env=$(get_inactive_env)
    local current_env="$ACTIVE_ENV"

    print_header "Blue-Green Deployment"
    echo -e "Current active: ${CYAN}${current_env}${NC}"
    echo -e "Deploying to:   ${CYAN}${target_env}${NC}"
    echo ""

    # Tag current image for fast rollback
    tag_current_image_for_rollback

    # Start infrastructure
    start_infrastructure

    # Deploy to inactive environment
    if ! deploy_env "$target_env"; then
        print_error "Deployment to ${target_env} failed"
        print_warning "Keeping ${current_env} active"
        exit 1
    fi

    # Switch traffic
    switch_traffic "$target_env"

    # Update state
    write_state "$target_env" "$current_env"

    # Post-switch validation with auto-rollback
    if ! post_switch_validation; then
        print_error "Post-switch validation failed"
        if auto_rollback "$target_env" "$current_env"; then
            exit 1
        else
            print_error "Auto-rollback also failed - system may be in inconsistent state"
            exit 2
        fi
    fi

    print_header "Deployment Complete"
    echo -e "Active environment: ${GREEN}${target_env}${NC}"
    echo -e "Previous environment: ${YELLOW}${current_env}${NC} (stopped)"
    echo ""
    print_info "Use 'make rollback' to revert if needed"
}

cmd_switch() {
    read_state
    local target_env
    target_env=$(get_inactive_env)

    print_header "Switching Traffic"
    echo -e "Current: ${CYAN}${ACTIVE_ENV}${NC}"
    echo -e "Target:  ${CYAN}${target_env}${NC}"
    echo ""

    # Check if target is running
    if ! docker ps --filter "name=${CONTAINER_PREFIX}-backend-${target_env}" --filter "status=running" | grep -q backend; then
        print_error "backend-${target_env} is not running"
        print_info "Use 'make prod' to deploy first"
        exit 1
    fi

    switch_traffic "$target_env"
    write_state "$target_env" "$ACTIVE_ENV"
    print_success "Traffic switched to ${target_env}"
}

cmd_rollback() {
    read_state

    if [ -z "$PREVIOUS_ENV" ]; then
        print_error "No previous environment to rollback to"
        exit 1
    fi

    print_header "Rollback to ${PREVIOUS_ENV}"
    echo -e "Current: ${CYAN}${ACTIVE_ENV}${NC}"
    echo -e "Rolling back to: ${CYAN}${PREVIOUS_ENV}${NC}"
    echo ""

    local rollback_env="$PREVIOUS_ENV"
    local current_env="$ACTIVE_ENV"

    # Start infrastructure
    start_infrastructure

    # Try fast rollback with cached image, fall back to rebuild
    local use_cache="false"
    if has_rollback_image; then
        print_info "Cached rollback image found - using fast rollback"
        use_cache="true"
    else
        print_warning "No cached image found - rebuilding (slower)"
    fi

    if ! deploy_env "$rollback_env" "$use_cache"; then
        print_error "Rollback deployment failed"
        exit 1
    fi

    # Switch traffic
    switch_traffic "$rollback_env"

    # Update state
    write_state "$rollback_env" "$current_env"

    print_header "Rollback Complete"
    echo -e "Active environment: ${GREEN}${rollback_env}${NC}"
}

cmd_status() {
    read_state

    print_header "Blue-Green Status"
    echo -e "Active environment:   ${GREEN}${ACTIVE_ENV}${NC}"
    echo -e "Previous environment: ${YELLOW}${PREVIOUS_ENV:-none}${NC}"
    echo -e "Last deploy:          ${LAST_DEPLOY}"
    echo ""

    echo "Backend Status:"
    echo "---------------"
    for env in blue green; do
        local container="${CONTAINER_PREFIX}-backend-${env}"
        local status
        status=$(docker inspect --format='{{.State.Status}}' "$container" 2>/dev/null || echo "not_found")
        local health
        health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "n/a")

        if [ "$status" = "running" ]; then
            if [ "$health" = "healthy" ]; then
                echo -e "  backend-${env}: ${GREEN}running (healthy)${NC}"
            else
                echo -e "  backend-${env}: ${YELLOW}running (${health})${NC}"
            fi
        else
            echo -e "  backend-${env}: ${RED}${status}${NC}"
        fi
    done

    echo ""
    echo "Rollback Image:"
    if has_rollback_image; then
        echo -e "  ${GREEN}Available${NC} (fast rollback enabled)"
    else
        echo -e "  ${YELLOW}Not available${NC} (rollback will require rebuild)"
    fi

    echo ""
    echo "Note: Frontend should be deployed separately to Vercel, Cloudflare Pages, or CDN"
}

cmd_help() {
    echo "Blue-Green Deployment Script (Optimized)"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  (none)      Deploy to inactive environment (full blue-green deploy)"
    echo "  --switch    Switch traffic to inactive environment (no rebuild)"
    echo "  --rollback  Rollback to previous environment (fast if image cached)"
    echo "  --status    Show current deployment status"
    echo "  --help      Show this help message"
    echo ""
    echo "Features:"
    echo "  - Graceful shutdown (${GRACEFUL_SHUTDOWN_TIMEOUT}s connection drain)"
    echo "  - Deep health checks (DB + cache verification)"
    echo "  - Fast rollback (cached images)"
    echo "  - Auto-rollback on post-switch failure"
    echo ""
    echo "Examples:"
    echo "  $0                 # Deploy new version"
    echo "  $0 --status        # Check current state"
    echo "  $0 --rollback      # Revert to previous version"
}

# ============================================
# Main
# ============================================

cd "$PROJECT_DIR"

# Check prerequisites
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    print_error "curl is not installed (required for health checks)"
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    print_error "Environment file not found: $ENV_FILE"
    print_info "Create it from .env.prod.example or use --env-file flag"
    exit 1
fi

# Load environment for variable substitution
set -a
source "$ENV_FILE"
set +a

# Set container prefix from PROJECT_NAME (must be after sourcing env)
CONTAINER_PREFIX="${PROJECT_NAME:-starter-kit}"

case "${1:-}" in
    --switch)
        cmd_switch
        ;;
    --rollback)
        cmd_rollback
        ;;
    --status)
        cmd_status
        ;;
    --help|-h)
        cmd_help
        ;;
    "")
        cmd_deploy
        ;;
    *)
        print_error "Unknown command: $1"
        cmd_help
        exit 1
        ;;
esac
