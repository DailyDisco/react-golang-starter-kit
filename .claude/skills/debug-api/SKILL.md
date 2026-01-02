---
name: debug-api
description: Generate curl commands for testing API endpoints. Use when debugging API issues.
allowed-tools: Read, Grep, Glob, Bash(curl:*), AskUserQuestion
context-files:
  - backend/cmd/main.go
---

# API Debug Helper

Generate ready-to-use curl commands for testing this project's API endpoints.

## Hard Rules

1. Always include cookie jar for JWT auth (`-b cookies.txt -c cookies.txt`)
2. Include CSRF token for mutating requests (POST/PUT/PATCH/DELETE)
3. Use correct Content-Type headers
4. Show both the curl command and expected response shape
5. Default to localhost:8080 unless user specifies otherwise

## Authentication Flow

This API uses httpOnly cookies for JWT auth. The flow is:

1. **Login** → Sets `access_token` and `refresh_token` cookies
2. **Subsequent requests** → Send cookies automatically
3. **CSRF** → Required for mutations, get from `/api/auth/csrf`

### Login Command (Always Show First)

```bash
# Step 1: Get CSRF token
curl -c cookies.txt -b cookies.txt http://localhost:8080/api/auth/csrf

# Step 2: Login (saves JWT cookies)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: <token-from-step-1>" \
  -b cookies.txt -c cookies.txt \
  -d '{"email": "admin@example.com", "password": "admin123!"}'
```

## Common Endpoints

### Auth Endpoints

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt \
  -d '{"email": "user@example.com", "password": "Password123!", "name": "Test User"}'

# Get current user
curl http://localhost:8080/api/auth/me \
  -b cookies.txt -c cookies.txt

# Logout
curl -X POST http://localhost:8080/api/auth/logout \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt

# Refresh token
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt
```

### User Endpoints

```bash
# List users (admin only)
curl http://localhost:8080/api/v1/users \
  -b cookies.txt -c cookies.txt

# Get user by ID
curl http://localhost:8080/api/v1/users/1 \
  -b cookies.txt -c cookies.txt

# Update user
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt \
  -d '{"name": "Updated Name"}'
```

### Organization Endpoints

```bash
# Get organization by slug
curl http://localhost:8080/api/v1/orgs/my-org \
  -b cookies.txt -c cookies.txt

# List organization members
curl http://localhost:8080/api/v1/orgs/my-org/members \
  -b cookies.txt -c cookies.txt

# Update organization
curl -X PUT http://localhost:8080/api/v1/orgs/my-org \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt \
  -d '{"name": "New Org Name"}'
```

### Billing Endpoints

```bash
# Get subscription status
curl http://localhost:8080/api/v1/billing/subscription \
  -b cookies.txt -c cookies.txt

# Create checkout session
curl -X POST http://localhost:8080/api/v1/billing/checkout \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt \
  -d '{"priceId": "price_xxx"}'

# Get portal session
curl -X POST http://localhost:8080/api/v1/billing/portal \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt
```

### WebSocket Connection

```bash
# WebSocket connections use the auth cookie
websocat ws://localhost:8080/ws --header "Cookie: $(cat cookies.txt | grep access_token | awk '{print $6"="$7}')"
```

## Guided Workflow

When user asks for help debugging an endpoint:

### Step 1: Ask for Context

Ask:
1. Which endpoint are you testing? (path or description)
2. Are you authenticated? (have you logged in?)
3. What response are you getting? (status code, body)

### Step 2: Generate Commands

Based on the endpoint, generate:

1. **Setup commands** (CSRF + login if needed)
2. **The actual request**
3. **Expected successful response shape**
4. **Common error responses and fixes**

### Step 3: Troubleshooting

If the request fails:

| Status | Likely Cause | Fix |
|--------|--------------|-----|
| 401 | Not authenticated | Run login command first |
| 403 | Missing/invalid CSRF | Get fresh CSRF token |
| 403 | Insufficient permissions | Check user role |
| 404 | Wrong endpoint path | Verify route in main.go |
| 422 | Validation failed | Check request body format |
| 500 | Server error | Check backend logs |

## Output Format

```markdown
## Debug: {endpoint description}

### Prerequisites
- [ ] Backend running (`make dev` or `docker compose up`)
- [ ] Logged in (run login commands below if not)

### Commands

#### 1. Setup (run once per session)
\`\`\`bash
# Get CSRF token
CSRF=$(curl -s -c cookies.txt http://localhost:8080/api/auth/csrf | jq -r '.token')

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -b cookies.txt -c cookies.txt \
  -d '{"email": "admin@example.com", "password": "admin123!"}'
\`\`\`

#### 2. Your Request
\`\`\`bash
{the actual curl command}
\`\`\`

### Expected Response
\`\`\`json
{expected response shape}
\`\`\`

### Troubleshooting
{common issues for this endpoint}
```

## Tips

- Use `| jq .` to pretty-print JSON responses
- Use `-v` flag for verbose output (see headers)
- Use `-w "\n%{http_code}\n"` to see status code
- Check `docker compose logs backend -f` for server logs
