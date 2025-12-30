#!/bin/bash

# ============================================
# React + Go Starter Kit - Railway Deployment
# ============================================
# Deploy backend to Railway

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

print_header "Railway Backend Deployment"

# ============================================
# Pre-flight Checks
# ============================================

print_info "Running pre-flight checks..."

# Check Railway CLI
if ! command -v railway &> /dev/null; then
    print_error "Railway CLI is not installed"
    echo ""
    echo "Install it with: npm i -g @railway/cli"
    echo "Then login with: railway login"
    exit 1
fi
print_success "Railway CLI is installed"

# Check if logged in
if ! railway whoami &> /dev/null; then
    print_error "Not logged in to Railway"
    echo ""
    echo "Login with: railway login"
    exit 1
fi
print_success "Logged in to Railway"

# ============================================
# Project Setup
# ============================================

print_header "Railway Project Setup"

# Check if linked to a project
if ! railway status &> /dev/null; then
    print_warning "Not linked to a Railway project"
    echo ""
    echo "Options:"
    echo "  1. Link existing project: railway link"
    echo "  2. Create new project: railway init"
    echo ""
    read -p "Create a new Railway project? [y/N]: " CREATE_PROJECT
    if [[ $CREATE_PROJECT =~ ^[Yy]$ ]]; then
        railway init
    else
        print_info "Please link to a project and run this script again"
        exit 0
    fi
fi

print_success "Linked to Railway project"

# ============================================
# Environment Variables
# ============================================

print_header "Environment Variables"

cat << EOF
Required environment variables for Railway:

Database (Railway PostgreSQL plugin):
  DATABASE_URL (automatically set by Railway PostgreSQL)

Application:
  JWT_SECRET          - Secret for JWT signing
  API_PORT            - Port for the API (default: 8080)
  GO_ENV              - Environment (production)
  CORS_ALLOWED_ORIGINS - Your frontend URL

Optional:
  STRIPE_SECRET_KEY   - Stripe API key
  STRIPE_WEBHOOK_SECRET - Stripe webhook signing secret
  SMTP_HOST           - Email server
  AWS_ACCESS_KEY_ID   - For S3 file storage
  AWS_SECRET_ACCESS_KEY
  AWS_BUCKET_NAME
  AWS_REGION

EOF

read -p "Would you like to set environment variables now? [y/N]: " SET_VARS
if [[ $SET_VARS =~ ^[Yy]$ ]]; then
    print_info "Opening Railway dashboard for variable configuration..."
    railway variables
fi

# ============================================
# Deploy
# ============================================

print_header "Deploying Backend"

print_info "Deploying backend to Railway..."
cd backend
railway up --detach
cd ..

print_success "Deployment initiated!"

# ============================================
# Summary
# ============================================

print_header "Deployment Complete"

echo "Next steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. Add PostgreSQL plugin in Railway dashboard"
echo "2. Set required environment variables"
echo "3. Get your API URL from Railway dashboard"
echo "4. Deploy frontend with: ./scripts/deploy-vercel.sh"
echo ""
print_info "View deployment: railway open"
print_info "View logs: railway logs"
print_info "Railway dashboard: https://railway.app/dashboard"

exit 0
