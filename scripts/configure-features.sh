#!/bin/bash

# ============================================
# React + Go Starter Kit - Feature Configurator
# ============================================
# Interactive wizard to enable/disable features in your project.
# Run this after init-project.sh or anytime you want to change features.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${CYAN}ℹ${NC} $1"; }

# Feature selection function
select_feature() {
    local feature_name=$1
    local feature_desc=$2
    local default=$3

    if [ "$default" = "y" ]; then
        read -p "  Enable ${feature_name}? (${feature_desc}) [Y/n]: " response
        response=${response:-Y}
    else
        read -p "  Enable ${feature_name}? (${feature_desc}) [y/N]: " response
        response=${response:-N}
    fi

    if [[ $response =~ ^[Yy]$ ]]; then
        return 0  # enabled
    else
        return 1  # disabled
    fi
}

print_header "Feature Configuration Wizard"

cat << "EOF"
This wizard helps you configure which features to include in your project.
Disabled features will have their code commented out but not removed,
making it easy to re-enable them later.

EOF

# ============================================
# Feature Selection
# ============================================

echo -e "${CYAN}Core Features:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Stripe/Payments
if select_feature "Stripe Payments" "Subscriptions, checkout, webhooks" "y"; then
    ENABLE_STRIPE=true
    print_success "Stripe payments enabled"
else
    ENABLE_STRIPE=false
    print_info "Stripe payments disabled"
fi

# OAuth
if select_feature "OAuth Login" "Google and GitHub social login" "y"; then
    ENABLE_OAUTH=true
    print_success "OAuth enabled"
else
    ENABLE_OAUTH=false
    print_info "OAuth disabled"
fi

# 2FA
if select_feature "Two-Factor Auth" "TOTP-based 2FA" "y"; then
    ENABLE_2FA=true
    print_success "2FA enabled"
else
    ENABLE_2FA=false
    print_info "2FA disabled"
fi

# File Uploads
if select_feature "File Uploads" "S3 and database file storage" "y"; then
    ENABLE_FILES=true
    print_success "File uploads enabled"
else
    ENABLE_FILES=false
    print_info "File uploads disabled"
fi

echo ""
echo -e "${CYAN}Advanced Features:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Organizations (multi-tenancy)
if select_feature "Organizations" "Multi-tenant workspaces" "n"; then
    ENABLE_ORGS=true
    print_success "Organizations enabled"
else
    ENABLE_ORGS=false
    print_info "Organizations disabled"
fi

# WebSockets
if select_feature "WebSockets" "Real-time notifications" "n"; then
    ENABLE_WEBSOCKETS=true
    print_success "WebSockets enabled"
else
    ENABLE_WEBSOCKETS=false
    print_info "WebSockets disabled"
fi

# i18n
if select_feature "Internationalization" "Multi-language support" "n"; then
    ENABLE_I18N=true
    print_success "i18n enabled"
else
    ENABLE_I18N=false
    print_info "i18n disabled"
fi

# Feature Flags
if select_feature "Feature Flags" "Admin-controlled feature toggles" "n"; then
    ENABLE_FEATURE_FLAGS=true
    print_success "Feature flags enabled"
else
    ENABLE_FEATURE_FLAGS=false
    print_info "Feature flags disabled"
fi

# ============================================
# Generate Configuration
# ============================================

print_header "Generating Configuration"

# Create features config file
cat > .features.json << EOF
{
  "stripe": $ENABLE_STRIPE,
  "oauth": $ENABLE_OAUTH,
  "twoFactorAuth": $ENABLE_2FA,
  "fileUploads": $ENABLE_FILES,
  "organizations": $ENABLE_ORGS,
  "websockets": $ENABLE_WEBSOCKETS,
  "i18n": $ENABLE_I18N,
  "featureFlags": $ENABLE_FEATURE_FLAGS,
  "configuredAt": "$(date -Iseconds)"
}
EOF

print_success "Created .features.json"

# Update .env.local with feature flags
if [ -f ".env.local" ]; then
    # Remove old feature flags section if exists
    sed -i '/# FEATURE FLAGS/,/# END FEATURE FLAGS/d' .env.local 2>/dev/null || true

    # Add new feature flags section
    cat >> .env.local << EOF

# FEATURE FLAGS (configured by configure-features.sh)
ENABLE_STRIPE=$ENABLE_STRIPE
ENABLE_OAUTH=$ENABLE_OAUTH
ENABLE_2FA=$ENABLE_2FA
ENABLE_FILE_UPLOADS=$ENABLE_FILES
ENABLE_ORGANIZATIONS=$ENABLE_ORGS
ENABLE_WEBSOCKETS=$ENABLE_WEBSOCKETS
ENABLE_I18N=$ENABLE_I18N
ENABLE_FEATURE_FLAGS=$ENABLE_FEATURE_FLAGS
# END FEATURE FLAGS
EOF
    print_success "Updated .env.local with feature flags"
fi

# ============================================
# Update docker-compose based on features
# ============================================

if [ -f "docker/compose.yml" ]; then
    print_info "Feature-specific Docker services can be enabled/disabled in docker/compose.yml"
fi

# ============================================
# Summary
# ============================================

print_header "Configuration Complete"

echo "Enabled Features:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
[ "$ENABLE_STRIPE" = "true" ] && echo -e "  ${GREEN}✓${NC} Stripe Payments"
[ "$ENABLE_OAUTH" = "true" ] && echo -e "  ${GREEN}✓${NC} OAuth Login"
[ "$ENABLE_2FA" = "true" ] && echo -e "  ${GREEN}✓${NC} Two-Factor Auth"
[ "$ENABLE_FILES" = "true" ] && echo -e "  ${GREEN}✓${NC} File Uploads"
[ "$ENABLE_ORGS" = "true" ] && echo -e "  ${GREEN}✓${NC} Organizations"
[ "$ENABLE_WEBSOCKETS" = "true" ] && echo -e "  ${GREEN}✓${NC} WebSockets"
[ "$ENABLE_I18N" = "true" ] && echo -e "  ${GREEN}✓${NC} Internationalization"
[ "$ENABLE_FEATURE_FLAGS" = "true" ] && echo -e "  ${GREEN}✓${NC} Feature Flags"

echo ""
echo "Disabled Features:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
[ "$ENABLE_STRIPE" = "false" ] && echo -e "  ${YELLOW}○${NC} Stripe Payments"
[ "$ENABLE_OAUTH" = "false" ] && echo -e "  ${YELLOW}○${NC} OAuth Login"
[ "$ENABLE_2FA" = "false" ] && echo -e "  ${YELLOW}○${NC} Two-Factor Auth"
[ "$ENABLE_FILES" = "false" ] && echo -e "  ${YELLOW}○${NC} File Uploads"
[ "$ENABLE_ORGS" = "false" ] && echo -e "  ${YELLOW}○${NC} Organizations"
[ "$ENABLE_WEBSOCKETS" = "false" ] && echo -e "  ${YELLOW}○${NC} WebSockets"
[ "$ENABLE_I18N" = "false" ] && echo -e "  ${YELLOW}○${NC} Internationalization"
[ "$ENABLE_FEATURE_FLAGS" = "false" ] && echo -e "  ${YELLOW}○${NC} Feature Flags"

echo ""
print_info "Configuration saved to .features.json"
print_info "Environment variables updated in .env.local"
print_warning "Re-run 'make dev' to apply changes"

exit 0
