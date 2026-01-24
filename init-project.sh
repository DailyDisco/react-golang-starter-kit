#!/bin/bash

# ============================================
# React + Go Starter Kit - Project Initializer
# ============================================
# This script helps you quickly customize this template for a new project.
# It will:
# - Rename the project throughout the codebase
# - Generate secure secrets
# - Configure environment variables
# - Initialize git repository
# - Create initial commit

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Generate random string
generate_secret() {
    local length=${1:-32}
    if command -v openssl &> /dev/null; then
        openssl rand -hex "$length"
    else
        # Fallback to /dev/urandom
        LC_ALL=C tr -dc 'a-zA-Z0-9' </dev/urandom | head -c "$((length * 2))"
    fi
}

# Validate input
validate_project_name() {
    local name=$1
    if [[ ! $name =~ ^[a-z0-9_-]+$ ]]; then
        return 1
    fi
    return 0
}

validate_domain() {
    local domain=$1
    if [[ -z "$domain" ]]; then
        return 0  # Optional field
    fi
    if [[ ! $domain =~ ^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$ ]]; then
        return 1
    fi
    return 0
}

# Check if this is already initialized
check_if_initialized() {
    if [ -f ".initialized" ]; then
        print_error "This project has already been initialized!"
        print_info "If you want to re-initialize, delete the .initialized file first."
        exit 1
    fi
}

# Welcome message
print_header "Welcome to React + Go Starter Kit Initializer"

cat << "EOF"
   ____                  _     ____  _             _
  / __ \                | |   / ___|| |           | |
 | |  | |_   _  ___  ___| | __\_  \ | |_ __ _ _ __| |_ ___ _ __
 | |  | | | | |/ _ \/ __| |/ / ___| | __/ _` | '__| __/ _ \ '__|
 | |__| | |_| |  __/ (__|   < /\__ \ | || (_| | |  | ||  __/ |
  \___\_\\__,_|\___|\___|_|\_\\____/  \__\__,_|_|   \__\___|_|

EOF

print_info "This script will help you set up your new project quickly."
print_warning "Make sure you have a backup before proceeding!\n"

check_if_initialized

# Confirm continuation
read -p "Do you want to continue? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "Initialization cancelled."
    exit 0
fi

# ============================================
# Step 1: Gather Project Information
# ============================================
print_header "Step 1: Project Information"

# Project name
while true; do
    read -p "Enter project name (lowercase, hyphens/underscores only) [my-project]: " PROJECT_NAME
    PROJECT_NAME=${PROJECT_NAME:-my-project}

    if validate_project_name "$PROJECT_NAME"; then
        break
    else
        print_error "Invalid project name. Use only lowercase letters, numbers, hyphens, and underscores."
    fi
done

# Convert to different formats
PROJECT_NAME_SNAKE=$(echo "$PROJECT_NAME" | tr '-' '_')
PROJECT_NAME_CAMEL=$(echo "$PROJECT_NAME" | perl -pe 's/(^|-)(\w)/\U$2/g')
PROJECT_NAME_UPPER=$(echo "$PROJECT_NAME_SNAKE" | tr '[:lower:]' '[:upper:]')

print_success "Project name: $PROJECT_NAME"

# Project description
read -p "Enter project description [A full-stack application]: " PROJECT_DESC
PROJECT_DESC=${PROJECT_DESC:-A full-stack application}

# Production domain (optional)
while true; do
    read -p "Production domain (optional, e.g., example.com): " PROD_DOMAIN

    if validate_domain "$PROD_DOMAIN"; then
        break
    else
        print_error "Invalid domain format."
    fi
done

# GitHub username/org (optional)
read -p "GitHub username or organization (optional): " GITHUB_USER

# ============================================
# Step 2: Database Configuration
# ============================================
print_header "Step 2: Database Configuration"

read -p "Database name [$PROJECT_NAME_SNAKE]: " DB_NAME
DB_NAME=${DB_NAME:-$PROJECT_NAME_SNAKE}

read -p "Database user [$PROJECT_NAME_SNAKE]: " DB_USER
DB_USER=${DB_USER:-$PROJECT_NAME_SNAKE}

# Generate database password
DB_PASSWORD=$(generate_secret 16)
print_success "Generated secure database password"

# Database port
read -p "Database port [5432]: " DB_PORT
DB_PORT=${DB_PORT:-5432}

# ============================================
# Step 3: API Configuration
# ============================================
print_header "Step 3: API Configuration"

read -p "API port [8080]: " API_PORT
API_PORT=${API_PORT:-8080}

read -p "Frontend dev port [5173]: " FRONTEND_PORT
FRONTEND_PORT=${FRONTEND_PORT:-5173}

# ============================================
# Step 4: Security Configuration
# ============================================
print_header "Step 4: Security Configuration"

print_info "Generating secure secrets..."

JWT_SECRET=$(generate_secret 32)
print_success "Generated JWT secret"
print_info "Note: JWT secret generator also available at backend/scripts/generate-jwt-secret.sh"

# JWT expiration
read -p "JWT token expiration (hours) [24]: " JWT_EXPIRATION
JWT_EXPIRATION=${JWT_EXPIRATION:-24}

# ============================================
# Step 5: Confirmation
# ============================================
print_header "Step 5: Confirmation"

echo "Please review your configuration:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Project Name:       $PROJECT_NAME"
echo "Description:        $PROJECT_DESC"
echo "Database Name:      $DB_NAME"
echo "Database User:      $DB_USER"
echo "API Port:           $API_PORT"
echo "Frontend Port:      $FRONTEND_PORT"
echo "JWT Expiration:     ${JWT_EXPIRATION}h"
if [ -n "$PROD_DOMAIN" ]; then
    echo "Production Domain:  $PROD_DOMAIN"
fi
if [ -n "$GITHUB_USER" ]; then
    echo "GitHub User:        $GITHUB_USER"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

read -p "Is this correct? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_error "Initialization cancelled. Please run the script again."
    exit 0
fi

# ============================================
# Step 6: Apply Changes
# ============================================
print_header "Step 6: Applying Changes"

# Create backup
print_info "Creating backup..."
BACKUP_DIR=".init-backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r . "$BACKUP_DIR/" 2>/dev/null || true
print_success "Backup created: $BACKUP_DIR"

# Update package.json files
print_info "Updating package.json files..."

if [ -f "package.json" ]; then
    sed -i.bak "s/\"name\": \"react-golang-starter-kit\"/\"name\": \"$PROJECT_NAME\"/" package.json
    rm package.json.bak
    print_success "Updated root package.json"
fi

if [ -f "frontend/package.json" ]; then
    sed -i.bak "s/\"name\": \"frontend\"/\"name\": \"${PROJECT_NAME}-frontend\"/" frontend/package.json
    rm frontend/package.json.bak
    print_success "Updated frontend/package.json"
fi

# Update Go module name
print_info "Updating Go module..."

if [ -f "backend/go.mod" ]; then
    CURRENT_MODULE=$(grep "^module " backend/go.mod | awk '{print $2}')
    NEW_MODULE="github.com/${GITHUB_USER:-yourname}/${PROJECT_NAME}"

    # Update go.mod
    sed -i.bak "s|$CURRENT_MODULE|$NEW_MODULE|g" backend/go.mod
    rm backend/go.mod.bak

    # Update all .go files with imports
    find backend -name "*.go" -type f -exec sed -i.bak "s|$CURRENT_MODULE|$NEW_MODULE|g" {} \;
    find backend -name "*.bak" -type f -delete

    print_success "Updated Go module name"
fi

# Create .env.local file
print_info "Creating .env.local file..."

cat > .env.local << EOF
# ============================================
# ${PROJECT_NAME_UPPER} - Environment Configuration
# ============================================
# Generated: $(date)

# ============================================
# DATABASE CONFIGURATION
# ============================================
DB_HOST=localhost
DB_PORT=$DB_PORT
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME
DB_SSLMODE=disable

# ============================================
# AUTHENTICATION & SECURITY
# ============================================
JWT_SECRET=$JWT_SECRET
JWT_EXPIRATION_HOURS=$JWT_EXPIRATION

# ============================================
# API CONFIGURATION
# ============================================
API_PORT=$API_PORT
VITE_API_URL=http://localhost:$API_PORT

# CORS - comma-separated list of allowed origins
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:$FRONTEND_PORT,http://localhost:$API_PORT

# ============================================
# APPLICATION SETTINGS
# ============================================
GO_ENV=development
NODE_ENV=development
DEBUG=true

# Logging
LOG_LEVEL=info
LOG_PRETTY=true

# ============================================
# RATE LIMITING
# ============================================
RATE_LIMIT_ENABLED=true

# ============================================
# PRODUCTION NOTES
# ============================================
# For production deployment:
# 1. Set DB_SSLMODE=require
# 2. Update CORS_ALLOWED_ORIGINS with your production domain
# 3. Set DEBUG=false and LOG_LEVEL=info or warn
# 4. Review optional features below for additional capabilities
EOF

print_success "Created .env.local file with secure configuration"

# Update README.md
print_info "Updating README.md..."

if [ -f "README.md" ]; then
    # Update title
    sed -i.bak "s/# ✨ React-Golang Starter Kit/# $PROJECT_NAME_CAMEL/" README.md

    # Update description
    sed -i.bak "s/A modern, production-ready full-stack starter template.*/$PROJECT_DESC/" README.md

    # Update clone URLs if GitHub user provided
    if [ -n "$GITHUB_USER" ]; then
        sed -i.bak "s|https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git|https://github.com/$GITHUB_USER/$PROJECT_NAME.git|g" README.md
        sed -i.bak "s|github.com/YOUR_USERNAME/YOUR_REPO_NAME|github.com/$GITHUB_USER/$PROJECT_NAME|g" README.md
    fi

    # Update demo URL with domain if provided
    if [ -n "$PROD_DOMAIN" ]; then
        sed -i.bak "s|https://react-golang-starter-kit.vercel.app/|https://$PROD_DOMAIN/|g" README.md
    else
        # Remove demo URL line if no domain
        sed -i.bak '/🌐 \*\*\[Live Demo\]/d' README.md
    fi

    rm README.md.bak
    print_success "Updated README.md"
fi

# Update docker-compose files
print_info "Updating Docker configuration..."

for compose_file in docker/compose.yml docker/compose.prod.yml; do
    if [ -f "$compose_file" ]; then
        sed -i.bak "s/react-golang-/${PROJECT_NAME}-/g" "$compose_file"
        rm "${compose_file}.bak"
        print_success "Updated $compose_file"
    fi
done

# ============================================
# Step 7: Git Initialization
# ============================================
print_header "Step 7: Git Initialization"

read -p "Do you want to initialize a new git repository? (y/n): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Check if .git exists
    if [ -d ".git" ]; then
        read -p "Git repository already exists. Remove it and start fresh? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf .git
            print_success "Removed existing .git directory"
        else
            print_info "Keeping existing git repository"
        fi
    fi

    if [ ! -d ".git" ]; then
        git init
        print_success "Initialized new git repository"

        # Create initial commit
        git add .
        git commit -m "chore: initialize project from react-golang-starter-kit template

Project: $PROJECT_NAME
Description: $PROJECT_DESC

Initial setup includes:
- Configured environment variables
- Updated project names across codebase
- Generated secure secrets
- Customized for new project

Template: https://github.com/YOUR_USERNAME/react-golang-starter-kit"

        print_success "Created initial commit"

        # Set up remote if GitHub user provided
        if [ -n "$GITHUB_USER" ]; then
            git remote add origin "https://github.com/$GITHUB_USER/$PROJECT_NAME.git"
            print_success "Added git remote: https://github.com/$GITHUB_USER/$PROJECT_NAME.git"
            print_info "Don't forget to create the repository on GitHub and push:"
            echo "  git push -u origin master"
        fi
    fi
else
    print_info "Skipping git initialization"
fi

# ============================================
# Step 8: Install Dependencies
# ============================================
print_header "Step 8: Install Dependencies"

read -p "Do you want to install dependencies now? (y/n): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_info "Installing frontend dependencies..."
    if [ -d "frontend" ]; then
        (cd frontend && npm install)
        print_success "Frontend dependencies installed"
    fi

    print_info "Installing backend dependencies..."
    if [ -d "backend" ]; then
        (cd backend && go mod tidy)
        print_success "Backend dependencies installed"
    fi

    print_info "Installing root dependencies..."
    if [ -f "package.json" ]; then
        npm install
        print_success "Root dependencies installed"
    fi
else
    print_info "Skipping dependency installation"
    print_warning "Remember to run:"
    echo "  npm install"
    echo "  cd frontend && npm install"
    echo "  cd backend && go mod tidy"
fi

# Mark as initialized
touch .initialized
echo "$PROJECT_NAME" > .initialized

# Ensure .initialized is in .gitignore
if ! grep -q "^\.initialized$" .gitignore 2>/dev/null; then
    echo ".initialized" >> .gitignore
    print_success "Added .initialized to .gitignore"
fi

# ============================================
# Step 9: Summary and Next Steps
# ============================================
print_header "🎉 Initialization Complete!"

cat << EOF

Your project "$PROJECT_NAME" is ready!

${GREEN}Configuration Summary:${NC}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Project renamed throughout codebase
✓ Secure JWT secret generated
✓ Database password generated
✓ .env.local file created with your configuration
✓ Docker compose files updated
✓ README.md customized
✓ Go module name updated
$([ -d ".git" ] && echo "✓ Git repository initialized")
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

${BLUE}Important Security Information:${NC}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
${YELLOW}⚠${NC}  Your .env.local file contains sensitive information!
${YELLOW}⚠${NC}  JWT Secret: $JWT_SECRET
${YELLOW}⚠${NC}  DB Password: $DB_PASSWORD

${RED}DO NOT commit the .env.local file to git!${NC}
It's already in .gitignore, but be careful.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

${BLUE}Next Steps:${NC}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Review your .env.local file
2. Start development:
   ${GREEN}make dev${NC}  or  ${GREEN}docker compose up -d${NC}

3. Access your application:
   Frontend: ${GREEN}http://localhost:$FRONTEND_PORT${NC}
   Backend:  ${GREEN}http://localhost:$API_PORT${NC}
   Health:   ${GREEN}http://localhost:$API_PORT/health${NC}

4. Check service health:
   ${GREEN}make health${NC}

$([ -n "$GITHUB_USER" ] && echo "5. Create GitHub repository and push:
   ${GREEN}git push -u origin master${NC}")
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

${BLUE}Useful Commands:${NC}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  make help             Show all available commands
  make dev              Start development environment
  make prod             Start production environment
  make logs             View all service logs
  make clean            Clean up containers and volumes
  make health           Check service health
  make format-backend   Format Go backend code
  ./docker-build.sh     Build operations (see --help)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

${BLUE}Documentation:${NC}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  README.md                Main documentation
  docs/README.md          Documentation hub
  docs/DEPLOYMENT.md      Deployment guides
  docs/FEATURES.md        Feature guides
  docs/FRONTEND_GUIDE.md  Frontend development
  docs/DOCKER_SETUP.md    Docker configuration
  backend/README.md       Backend development
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

${GREEN}Happy coding! 🚀${NC}

EOF

# Save configuration for reference
cat > .project-info << EOF
Project Name: $PROJECT_NAME
Description: $PROJECT_DESC
Initialized: $(date)
Database: $DB_NAME
API Port: $API_PORT
Frontend Port: $FRONTEND_PORT
EOF

print_success "Project information saved to .project-info"

exit 0
