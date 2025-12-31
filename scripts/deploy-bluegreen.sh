#!/bin/bash

# ============================================
# React + Go Starter Kit - Blue-Green Deployment
# ============================================
# Zero-downtime deployments using blue-green strategy
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

# Health check settings
HEALTH_CHECK_RETRIES=30
HEALTH_CHECK_INTERVAL=2

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
        -f "${PROJECT_DIR}/docker-compose.yml" \
        -f "${PROJECT_DIR}/docker-compose.prod.yml" \
        "$@"
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
    local container_name="react-golang-${service}"
    local retries=$HEALTH_CHECK_RETRIES

    print_info "Waiting for ${service} to be healthy..."

    while [ $retries -gt 0 ]; do
        local status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "not_found")

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

check_env_healthy() {
    local env="$1"
    wait_for_healthy "backend-${env}" && wait_for_healthy "frontend-${env}"
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
        if docker exec react-golang-postgres-prod pg_isready -U "${DB_USER:-devuser}" >/dev/null 2>&1; then
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
    print_header "Deploying ${env} environment"

    print_info "Building and starting ${env} services..."
    compose_cmd up -d --build "backend-${env}" "frontend-${env}"

    if ! check_env_healthy "$env"; then
        print_error "Deployment failed - ${env} environment is not healthy"
        print_info "Stopping failed ${env} services..."
        compose_cmd stop "backend-${env}" "frontend-${env}"
        compose_cmd rm -f "backend-${env}" "frontend-${env}"
        return 1
    fi

    print_success "${env} environment deployed successfully"
    return 0
}

switch_traffic() {
    local target_env="$1"
    print_header "Switching traffic to ${target_env}"

    # Update environment variables for Caddy
    export ACTIVE_BACKEND="backend-${target_env}"
    export ACTIVE_FRONTEND="frontend-${target_env}"

    print_info "Reloading Caddy with new upstreams..."

    # Recreate Caddy with new environment
    compose_cmd up -d --no-deps caddy

    # Verify Caddy is running
    sleep 3
    if docker ps --filter "name=react-golang-caddy-prod" --filter "status=running" | grep -q caddy; then
        print_success "Traffic switched to ${target_env}"
    else
        print_error "Caddy failed to reload"
        return 1
    fi
}

stop_env() {
    local env="$1"
    print_info "Stopping ${env} environment..."
    compose_cmd stop "backend-${env}" "frontend-${env}"
    compose_cmd rm -f "backend-${env}" "frontend-${env}"
    print_success "${env} environment stopped"
}

# ============================================
# Main Commands
# ============================================

cmd_deploy() {
    read_state
    local target_env=$(get_inactive_env)
    local current_env="$ACTIVE_ENV"

    print_header "Blue-Green Deployment"
    echo "Current active: ${CYAN}${current_env}${NC}"
    echo "Deploying to:   ${CYAN}${target_env}${NC}"
    echo ""

    # Start infrastructure
    start_infrastructure

    # Deploy to inactive environment
    if ! deploy_env "$target_env"; then
        print_error "Deployment to ${target_env} failed"
        print_warning "Keeping ${current_env} active"
        exit 1
    fi

    # Switch traffic
    if ! switch_traffic "$target_env"; then
        print_error "Traffic switch failed"
        print_warning "Attempting to keep ${current_env} active"
        stop_env "$target_env"
        exit 1
    fi

    # Update state
    write_state "$target_env" "$current_env"

    # Stop old environment
    if [ -n "$current_env" ] && [ "$current_env" != "$target_env" ]; then
        stop_env "$current_env"
    fi

    print_header "Deployment Complete"
    echo -e "Active environment: ${GREEN}${target_env}${NC}"
    echo -e "Previous environment: ${YELLOW}${current_env}${NC} (stopped)"
    echo ""
    print_info "Use 'make rollback' to revert if needed"
}

cmd_switch() {
    read_state
    local target_env=$(get_inactive_env)

    print_header "Switching Traffic"
    echo "Current: ${CYAN}${ACTIVE_ENV}${NC}"
    echo "Target:  ${CYAN}${target_env}${NC}"
    echo ""

    # Check if target is running
    if ! docker ps --filter "name=react-golang-backend-${target_env}" --filter "status=running" | grep -q backend; then
        print_error "${target_env} environment is not running"
        print_info "Use 'make deploy-${target_env}' to start it first"
        exit 1
    fi

    if ! switch_traffic "$target_env"; then
        exit 1
    fi

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
    echo "Current: ${CYAN}${ACTIVE_ENV}${NC}"
    echo "Rolling back to: ${CYAN}${PREVIOUS_ENV}${NC}"
    echo ""

    local rollback_env="$PREVIOUS_ENV"
    local current_env="$ACTIVE_ENV"

    # Start infrastructure
    start_infrastructure

    # Rebuild and start previous environment
    if ! deploy_env "$rollback_env"; then
        print_error "Rollback deployment failed"
        exit 1
    fi

    # Switch traffic
    if ! switch_traffic "$rollback_env"; then
        print_error "Traffic switch failed during rollback"
        exit 1
    fi

    # Update state
    write_state "$rollback_env" "$current_env"

    # Stop failed environment
    stop_env "$current_env"

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

    echo "Service Status:"
    echo "---------------"
    for env in blue green; do
        for svc in backend frontend; do
            local container="react-golang-${svc}-${env}"
            local status=$(docker inspect --format='{{.State.Status}}' "$container" 2>/dev/null || echo "not_found")
            local health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "n/a")

            if [ "$status" = "running" ]; then
                if [ "$health" = "healthy" ]; then
                    echo -e "  ${svc}-${env}: ${GREEN}running (healthy)${NC}"
                else
                    echo -e "  ${svc}-${env}: ${YELLOW}running (${health})${NC}"
                fi
            else
                echo -e "  ${svc}-${env}: ${RED}${status}${NC}"
            fi
        done
    done
}

cmd_help() {
    echo "Blue-Green Deployment Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  (none)      Deploy to inactive environment (full blue-green deploy)"
    echo "  --switch    Switch traffic to inactive environment (no rebuild)"
    echo "  --rollback  Rollback to previous environment"
    echo "  --status    Show current deployment status"
    echo "  --help      Show this help message"
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

if [ ! -f "$ENV_FILE" ]; then
    print_error "Environment file not found: $ENV_FILE"
    print_info "Create it from .env.prod.example or use --env-file flag"
    exit 1
fi

# Load environment for variable substitution
set -a
source "$ENV_FILE"
set +a

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
