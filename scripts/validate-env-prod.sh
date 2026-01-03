#!/usr/bin/env bash
# =============================================================================
# Production Environment Validation Script
# =============================================================================
# Validates that .env.prod (or specified env file) is properly configured
# for production deployment. Run before deploying to catch configuration issues.
#
# Usage:
#   ./scripts/validate-env-prod.sh              # Validates .env.prod
#   ./scripts/validate-env-prod.sh .env.staging # Validates custom file
#
# Exit codes:
#   0 - All validations passed
#   1 - Critical errors found (deployment blocked)
#   2 - Warnings found (review recommended)
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
ERRORS=0
WARNINGS=0

# Target file
ENV_FILE="${1:-.env.prod}"

# =============================================================================
# Helper Functions
# =============================================================================

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((ERRORS++))
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

# Get value from env file (handles quoted values)
get_env() {
    local key="$1"
    local value
    value=$(grep -E "^${key}=" "$ENV_FILE" 2>/dev/null | head -1 | cut -d '=' -f2- | sed 's/^["'"'"']//;s/["'"'"']$//')
    echo "$value"
}

# Check if variable exists and is not empty
require_set() {
    local key="$1"
    local desc="$2"
    local value
    value=$(get_env "$key")

    if [[ -z "$value" ]]; then
        error "$key is not set ($desc)"
        return 1
    fi
    return 0
}

# Check if variable equals expected value
require_value() {
    local key="$1"
    local expected="$2"
    local desc="$3"
    local value
    value=$(get_env "$key")

    if [[ "$value" != "$expected" ]]; then
        error "$key should be '$expected' for production ($desc) - found: '$value'"
        return 1
    fi
    success "$key=$expected"
    return 0
}

# Check if variable is NOT a specific value
require_not_value() {
    local key="$1"
    local bad_value="$2"
    local desc="$3"
    local value
    value=$(get_env "$key")

    if [[ "$value" == "$bad_value" ]]; then
        error "$key should not be '$bad_value' in production ($desc)"
        return 1
    fi
    return 0
}

# Check minimum length
require_min_length() {
    local key="$1"
    local min_len="$2"
    local desc="$3"
    local value
    value=$(get_env "$key")

    if [[ ${#value} -lt $min_len ]]; then
        error "$key must be at least $min_len characters ($desc) - found: ${#value} chars"
        return 1
    fi
    success "$key has sufficient length (${#value} chars)"
    return 0
}

# Check if URL starts with https
require_https() {
    local key="$1"
    local value
    value=$(get_env "$key")

    if [[ -n "$value" && ! "$value" =~ ^https:// ]]; then
        warn "$key should use HTTPS in production - found: $value"
        return 1
    fi
    return 0
}

# Check for development/test patterns
check_no_dev_values() {
    local key="$1"
    local value
    value=$(get_env "$key")

    if [[ "$value" =~ (localhost|127\.0\.0\.1|dev|test|example|changeme|password123|secret123) ]]; then
        warn "$key appears to contain development values: $value"
        return 1
    fi
    return 0
}

# =============================================================================
# Main Validation
# =============================================================================

echo ""
echo "=============================================="
echo " Production Environment Validation"
echo "=============================================="
echo " File: $ENV_FILE"
echo "=============================================="
echo ""

# Check file exists
if [[ ! -f "$ENV_FILE" ]]; then
    echo -e "${RED}ERROR: $ENV_FILE not found${NC}"
    echo ""
    echo "Create from template:"
    echo "  cp .env.example $ENV_FILE"
    echo ""
    exit 1
fi

# -----------------------------------------------------------------------------
# 1. Critical Security Settings
# -----------------------------------------------------------------------------
echo "--- Security Settings ---"

# GO_ENV must be production
require_value "GO_ENV" "production" "environment mode"

# DEBUG must be false
require_value "DEBUG" "false" "debug mode must be off"

# JWT_SECRET minimum length
require_set "JWT_SECRET" "authentication will fail" && \
    require_min_length "JWT_SECRET" 32 "use: openssl rand -hex 32" && \
    require_not_value "JWT_SECRET" "dev-jwt-secret-key-change-in-production" "default dev value"

# TOTP encryption key for 2FA
if [[ -n "$(get_env 'TOTP_ENCRYPTION_KEY')" ]]; then
    require_min_length "TOTP_ENCRYPTION_KEY" 32 "2FA encryption key"
fi

echo ""

# -----------------------------------------------------------------------------
# 2. Database Configuration
# -----------------------------------------------------------------------------
echo "--- Database Configuration ---"

# Check for Railway PG* vars OR DB_* vars
PG_HOST=$(get_env "PGHOST")
DB_HOST=$(get_env "DB_HOST")

if [[ -z "$PG_HOST" && -z "$DB_HOST" ]]; then
    error "No database host configured (PGHOST or DB_HOST required)"
elif [[ -n "$PG_HOST" ]]; then
    info "Using Railway PostgreSQL (PGHOST)"
    require_set "PGPASSWORD" "database authentication"
else
    info "Using standard PostgreSQL (DB_HOST)"
    require_set "DB_PASSWORD" "database authentication"
    check_no_dev_values "DB_PASSWORD"
    check_no_dev_values "DB_HOST"
fi

# SSL mode for production
DB_SSL=$(get_env "DB_SSLMODE")
if [[ "$DB_SSL" == "disable" ]]; then
    warn "DB_SSLMODE=disable - consider 'require' for production"
fi

echo ""

# -----------------------------------------------------------------------------
# 3. Cookie & CSRF Settings
# -----------------------------------------------------------------------------
echo "--- Cookie & CSRF Settings ---"

CSRF_SECURE=$(get_env "CSRF_COOKIE_SECURE")
if [[ "$CSRF_SECURE" != "true" ]]; then
    warn "CSRF_COOKIE_SECURE should be 'true' for HTTPS deployments"
fi

HSTS=$(get_env "SECURITY_HSTS_ENABLED")
if [[ "$HSTS" != "true" ]]; then
    warn "SECURITY_HSTS_ENABLED should be 'true' for HTTPS deployments"
fi

echo ""

# -----------------------------------------------------------------------------
# 4. URLs and CORS
# -----------------------------------------------------------------------------
echo "--- URLs and CORS ---"

# CORS origins should not have localhost
CORS=$(get_env "CORS_ALLOWED_ORIGINS")
if [[ "$CORS" =~ localhost ]]; then
    warn "CORS_ALLOWED_ORIGINS contains 'localhost' - remove for production"
fi

# Frontend URL should be HTTPS
require_https "FRONTEND_URL"

# Stripe URLs should be HTTPS
require_https "STRIPE_SUCCESS_URL"
require_https "STRIPE_CANCEL_URL"
require_https "STRIPE_PORTAL_RETURN_URL"

echo ""

# -----------------------------------------------------------------------------
# 5. API Keys and Secrets
# -----------------------------------------------------------------------------
echo "--- API Keys and Secrets ---"

# Stripe - check for test keys
STRIPE_SECRET=$(get_env "STRIPE_SECRET_KEY")
if [[ "$STRIPE_SECRET" =~ ^sk_test_ ]]; then
    warn "STRIPE_SECRET_KEY is a test key (sk_test_*) - use live key for production"
fi

STRIPE_PUB=$(get_env "STRIPE_PUBLISHABLE_KEY")
if [[ "$STRIPE_PUB" =~ ^pk_test_ ]]; then
    warn "STRIPE_PUBLISHABLE_KEY is a test key (pk_test_*) - use live key for production"
fi

# Redis password in production
REDIS_URL=$(get_env "REDIS_URL")
REDIS_PASS=$(get_env "REDIS_PASSWORD")
if [[ -n "$REDIS_URL" && -z "$REDIS_PASS" && ! "$REDIS_URL" =~ : ]]; then
    warn "REDIS_PASSWORD not set - recommended for production Redis"
fi

echo ""

# -----------------------------------------------------------------------------
# 6. Logging and Debug
# -----------------------------------------------------------------------------
echo "--- Logging Configuration ---"

# Pretty logging off in production (for JSON parsing)
LOG_PRETTY=$(get_env "LOG_PRETTY")
if [[ "$LOG_PRETTY" == "true" ]]; then
    warn "LOG_PRETTY=true - use 'false' for structured JSON logs in production"
fi

# Auto-seed must be off
AUTO_SEED=$(get_env "AUTO_SEED")
if [[ "$AUTO_SEED" == "true" ]]; then
    error "AUTO_SEED=true - must be 'false' in production"
fi

echo ""

# -----------------------------------------------------------------------------
# 7. Optional Feature Checks
# -----------------------------------------------------------------------------
echo "--- Optional Features ---"

# Email configuration
EMAIL_ENABLED=$(get_env "EMAIL_ENABLED")
if [[ "$EMAIL_ENABLED" == "true" ]]; then
    require_set "SMTP_HOST" "email sending"
    require_set "SMTP_USER" "email authentication"
    require_set "SMTP_PASSWORD" "email authentication"

    EMAIL_DEV=$(get_env "EMAIL_DEV_MODE")
    if [[ "$EMAIL_DEV" == "true" ]]; then
        warn "EMAIL_DEV_MODE=true - emails won't actually send"
    fi
fi

# Sentry for error tracking
SENTRY=$(get_env "SENTRY_DSN")
if [[ -z "$SENTRY" ]]; then
    info "SENTRY_DSN not configured - consider adding error tracking"
fi

echo ""

# =============================================================================
# Summary
# =============================================================================

echo "=============================================="
echo " Validation Summary"
echo "=============================================="

if [[ $ERRORS -gt 0 ]]; then
    echo -e " ${RED}Errors:   $ERRORS (must fix before deploy)${NC}"
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo -e " ${YELLOW}Warnings: $WARNINGS (review recommended)${NC}"
fi

if [[ $ERRORS -eq 0 && $WARNINGS -eq 0 ]]; then
    echo -e " ${GREEN}All checks passed!${NC}"
fi

echo "=============================================="
echo ""

# Exit with appropriate code
if [[ $ERRORS -gt 0 ]]; then
    echo -e "${RED}Deployment blocked: Fix errors above${NC}"
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}Review warnings before deploying${NC}"
    exit 2
else
    echo -e "${GREEN}Ready for production deployment${NC}"
    exit 0
fi
