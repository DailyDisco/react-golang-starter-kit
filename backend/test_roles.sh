#!/bin/bash

# Test script for RBAC (Role-Based Access Control) system
# This script tests the complete user role and permission system

echo "🚀 Testing RBAC System Implementation"
echo "======================================"

BASE_URL="http://localhost:8080/api"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to make HTTP requests and check responses
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local auth_token=$4
    local expected_status=$5
    local description=$6

    echo -e "\n${BLUE}Testing: ${description}${NC}"
    echo "Method: $method"
    echo "URL: $url"

    # Build curl command
    local curl_cmd="curl -s -w \"\nHTTP_STATUS:%{http_code}\" -X $method"

    if [ -n "$auth_token" ]; then
        curl_cmd="$curl_cmd -H \"Authorization: Bearer $auth_token\""
    fi

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H \"Content-Type: application/json\" -d '$data'"
    fi

    curl_cmd="$curl_cmd $url"

    # Execute request
    response=$(eval $curl_cmd)
    status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_STATUS:/d')

    echo "Expected Status: $expected_status"
    echo "Actual Status: $status"

    if [ "$status" -eq "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "Response Body: $body"
    fi

    return $([ "$status" -eq "$expected_status" ])
}

# Test data
USER_EMAIL="testuser@example.com"
USER_PASSWORD="TestPass123!"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="AdminPass123!"

echo -e "\n${YELLOW}Step 1: Testing User Registration${NC}"
echo "======================================="

# Register a regular user
test_endpoint "POST" "$BASE_URL/users" "{
    \"name\": \"Test User\",
    \"email\": \"$USER_EMAIL\",
    \"password\": \"$USER_PASSWORD\"
}" "" 200 "User Registration"

# Register an admin user (we'll need to manually set this in the database)
test_endpoint "POST" "$BASE_URL/auth/register" "{
    \"name\": \"Admin User\",
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
}" "" 200 "Admin User Registration"

echo -e "\n${YELLOW}Step 2: Testing User Login${NC}"
echo "=============================="

# Login as regular user
echo "Logging in as regular user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$USER_EMAIL\", \"password\": \"$USER_PASSWORD\"}")

USER_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "User token obtained: ${USER_TOKEN:0:20}..."

# Login as admin user
echo "Logging in as admin user..."
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$ADMIN_EMAIL\", \"password\": \"$ADMIN_PASSWORD\"}")

ADMIN_TOKEN=$(echo $ADMIN_LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "Admin token obtained: ${ADMIN_TOKEN:0:20}..."

echo -e "\n${YELLOW}Step 3: Testing User Profile Access${NC}"
echo "========================================"

# Test user accessing their own profile
test_endpoint "GET" "$BASE_URL/users/me" "" "$USER_TOKEN" 200 "User accessing own profile"

# Test user accessing admin routes (should fail)
test_endpoint "GET" "$BASE_URL/users/admin" "" "$USER_TOKEN" 403 "Regular user accessing admin routes"

echo -e "\n${YELLOW}Step 4: Testing Premium Content Access${NC}"
echo "==========================================="

# Test user accessing premium content (should fail)
test_endpoint "GET" "$BASE_URL/premium/content" "" "$USER_TOKEN" 403 "Regular user accessing premium content"

echo -e "\n${YELLOW}Step 5: Testing Admin Access (Note: Need to set admin role manually)${NC}"
echo "======================================================================"

echo -e "\n${BLUE}Manual Setup Required:${NC}"
echo "1. Connect to your database"
echo "2. Find the admin user with email: $ADMIN_EMAIL"
echo "3. Update their role to 'admin' or 'super_admin'"
echo "4. Run this test again to verify admin permissions"

echo -e "\n${YELLOW}Step 6: Testing Public Endpoints${NC}"
echo "====================================="

# Test public endpoints (no auth required)
test_endpoint "GET" "$BASE_URL/health" "" "" 200 "Health check endpoint"
test_endpoint "POST" "$BASE_URL/auth/login" "{
    \"email\": \"$USER_EMAIL\",
    \"password\": \"$USER_PASSWORD\"
}" "" 200 "Login endpoint"

echo -e "\n${YELLOW}Step 7: Testing Authentication Middleware${NC}"
echo "==============================================="

# Test accessing protected endpoint without token
test_endpoint "GET" "$BASE_URL/users/me" "" "" 401 "Accessing protected endpoint without token"

# Test accessing protected endpoint with invalid token
test_endpoint "GET" "$BASE_URL/users/me" "" "invalid_token" 401 "Accessing protected endpoint with invalid token"

echo -e "\n${GREEN}🎉 RBAC Testing Complete!${NC}"
echo "============================"
echo ""
echo "Next steps:"
echo "1. Set up your database and start the server"
echo "2. Run this test script: ./test_roles.sh"
echo "3. Manually promote a user to admin role in the database"
echo "4. Re-run the tests to verify admin permissions work"
echo ""
echo "Expected Results:"
echo "- Regular users can access their own profile and public endpoints"
echo "- Regular users cannot access admin or premium content"
echo "- Admin users can access all endpoints including user management"
echo "- Premium users can access premium content"
