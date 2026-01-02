---
name: env-check
description: Validate environment setup before starting development. Use to catch config issues early.
allowed-tools: Read, Bash(docker:*), Bash(cat:*), Bash(grep:*), Bash(test:*)
---

# Environment Check

Validate that the development environment is properly configured before starting.

## Hard Rules

1. NEVER expose or log actual secret values
2. Only report presence/format issues, not actual values
3. Check Docker containers are running
4. Provide specific fix instructions for each issue

## Process

### Phase 1: Check .env File Exists

```bash
if [ ! -f .env ]; then
  echo "ERROR: .env file not found"
  echo "FIX: cp .env.example .env"
  exit 1
fi
```

### Phase 2: Validate Required Variables

Check these critical variables are set (non-empty):

| Variable | Validation | Error Message |
|----------|------------|---------------|
| `JWT_SECRET` | Length >= 32 chars, not default | "JWT_SECRET too short or using default. Run: openssl rand -hex 32" |
| `DB_HOST` | Non-empty | "DB_HOST not set" |
| `DB_PORT` | Numeric | "DB_PORT must be a number" |
| `DB_USER` | Non-empty | "DB_USER not set" |
| `DB_PASSWORD` | Non-empty | "DB_PASSWORD not set" |
| `DB_NAME` | Non-empty | "DB_NAME not set" |

### Phase 3: Check Docker Containers

Verify required containers are running:

```bash
docker compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null
```

Expected containers:
- `postgres` or `db` - Database
- `dragonfly` or `redis` - Cache
- `backend` - API server (optional for local Go dev)
- `frontend` - Vite dev server (optional for local npm dev)

### Phase 4: Test Database Connection

```bash
# Check if postgres is accessible
docker compose exec -T db pg_isready -U $DB_USER -d $DB_NAME 2>/dev/null
```

### Phase 5: Test Cache Connection

```bash
# Check if dragonfly/redis is accessible
docker compose exec -T dragonfly redis-cli ping 2>/dev/null
```

## Output Format

```markdown
## Environment Check Results

### Configuration File
✅ .env file exists

### Required Variables
✅ JWT_SECRET - Set (64 chars)
✅ DB_HOST - localhost
✅ DB_PORT - 5432
✅ DB_USER - Set
✅ DB_PASSWORD - Set
✅ DB_NAME - starter_kit_db
❌ STRIPE_SECRET_KEY - Using default test key

### Docker Containers
✅ postgres - running (healthy)
✅ dragonfly - running
⚠️ backend - not running (OK if running locally)
⚠️ frontend - not running (OK if running locally)

### Connectivity
✅ Database - accepting connections
✅ Cache - responding to ping

### Issues Found: 1
1. STRIPE_SECRET_KEY is using a placeholder. Set a real key for payment testing.

### Ready to Develop?
✅ Yes - Core services are running. Run `make dev` to start.
```

## Common Issues and Fixes

| Issue | Fix |
|-------|-----|
| `.env` missing | `cp .env.example .env` |
| JWT_SECRET too short | `echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env` |
| Docker not running | `docker compose up -d` |
| Database connection refused | `docker compose restart db` |
| Port already in use | Check `.env` for port conflicts, or `lsof -i :5432` |

## Validation Rules

### JWT_SECRET
- Must be at least 32 characters
- Must NOT be the default value `dev-jwt-secret-key-change-in-production`
- Recommended: 64 hex characters from `openssl rand -hex 32`

### Database Variables
- DB_PORT must be numeric (default: 5432)
- DB_SSLMODE should be `disable` for local, `require` for production

### Stripe (if enabled)
- STRIPE_SECRET_KEY must start with `sk_test_` or `sk_live_`
- STRIPE_PUBLISHABLE_KEY must start with `pk_test_` or `pk_live_`
- STRIPE_WEBHOOK_SECRET must start with `whsec_`
