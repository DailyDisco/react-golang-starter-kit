#!/bin/bash

# ============================================
# React + Go Starter Kit - Vercel Deployment
# ============================================
# Deploy frontend to Vercel (backend should be on Railway/Render/VPS)

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
PROD=false
PROJECT_NAME=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --prod)
            PROD=true
            shift
            ;;
        --project)
            PROJECT_NAME="$2"
            shift 2
            ;;
        --help)
            echo "Usage: ./deploy-vercel.sh [options]"
            echo ""
            echo "Options:"
            echo "  --prod              Deploy to production (default: preview)"
            echo "  --project <name>    Vercel project name"
            echo "  --help              Show this help message"
            echo ""
            echo "Prerequisites:"
            echo "  1. Install Vercel CLI: npm i -g vercel"
            echo "  2. Login to Vercel: vercel login"
            echo "  3. Backend deployed and API_URL configured"
            echo ""
            echo "Examples:"
            echo "  ./deploy-vercel.sh                    # Preview deployment"
            echo "  ./deploy-vercel.sh --prod             # Production deployment"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_header "Vercel Frontend Deployment"

# ============================================
# Pre-flight Checks
# ============================================

print_info "Running pre-flight checks..."

# Check Vercel CLI
if ! command -v vercel &> /dev/null; then
    print_error "Vercel CLI is not installed"
    echo ""
    echo "Install it with: npm i -g vercel"
    echo "Then login with: vercel login"
    exit 1
fi
print_success "Vercel CLI is installed"

# Check if logged in
if ! vercel whoami &> /dev/null; then
    print_error "Not logged in to Vercel"
    echo ""
    echo "Login with: vercel login"
    exit 1
fi
print_success "Logged in to Vercel as: $(vercel whoami)"

# Check frontend directory
if [ ! -d "frontend" ]; then
    print_error "frontend directory not found"
    exit 1
fi
print_success "Frontend directory found"

# ============================================
# Environment Check
# ============================================

print_header "Environment Configuration"

# Check for API URL
if [ -z "$VITE_API_URL" ]; then
    print_warning "VITE_API_URL environment variable not set"
    read -p "Enter your production API URL (e.g., https://api.example.com): " VITE_API_URL
    if [ -z "$VITE_API_URL" ]; then
        print_error "API URL is required for deployment"
        exit 1
    fi
fi
print_success "API URL: $VITE_API_URL"

# ============================================
# Build Frontend
# ============================================

print_header "Building Frontend"

cd frontend

print_info "Installing dependencies..."
npm ci --silent
print_success "Dependencies installed"

print_info "Building for production..."
VITE_API_URL="$VITE_API_URL" npm run build
print_success "Build complete"

# ============================================
# Deploy to Vercel
# ============================================

print_header "Deploying to Vercel"

# Build deploy command
DEPLOY_CMD="vercel"

if [ "$PROD" = true ]; then
    DEPLOY_CMD="$DEPLOY_CMD --prod"
    print_info "Deploying to PRODUCTION"
else
    print_info "Deploying preview (use --prod for production)"
fi

if [ -n "$PROJECT_NAME" ]; then
    DEPLOY_CMD="$DEPLOY_CMD --name $PROJECT_NAME"
fi

# Set environment variables
DEPLOY_CMD="$DEPLOY_CMD -e VITE_API_URL=$VITE_API_URL"

# Deploy
print_info "Running: $DEPLOY_CMD"
DEPLOY_URL=$($DEPLOY_CMD --yes 2>&1 | tail -1)

cd ..

# ============================================
# Summary
# ============================================

print_header "Deployment Complete"

echo "Deployment URL: $DEPLOY_URL"
echo ""
if [ "$PROD" = true ]; then
    print_success "Production deployment successful!"
else
    print_success "Preview deployment successful!"
    print_info "Deploy to production with: ./deploy-vercel.sh --prod"
fi

echo ""
echo "Environment Variables to set in Vercel Dashboard:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  VITE_API_URL = $VITE_API_URL"
echo ""
print_info "View deployment: $DEPLOY_URL"
print_info "Vercel dashboard: https://vercel.com/dashboard"

exit 0
