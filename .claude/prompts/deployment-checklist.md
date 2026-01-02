# Deployment Checklist - React + Go Production

Use this checklist before deploying to production.

---

## Pre-Deployment

### 1. Environment Configuration

```
[ ] GO_ENV=production
[ ] DEBUG=false
[ ] LOG_PRETTY=false (use JSON logs for aggregation)
[ ] AUTO_SEED=false
```

### 2. Security Secrets

```
[ ] JWT_SECRET - Generated securely (openssl rand -hex 32)
[ ] JWT_SECRET - NOT the default dev value
[ ] TOTP_ENCRYPTION_KEY - Set for 2FA (if enabled)
[ ] Database passwords - Strong, rotated from dev
[ ] API keys - Production keys (not test/dev)
```

### 3. Database

```
[ ] DB_SSLMODE=require (or verify-full)
[ ] Connection pooling configured
[ ] Migrations applied (make migrate-up)
[ ] Indexes verified for production queries
[ ] Backups configured and tested
```

### 4. HTTPS & Security Headers

```
[ ] HTTPS enforced (redirect HTTP → HTTPS)
[ ] CSRF_COOKIE_SECURE=true
[ ] SECURITY_HSTS_ENABLED=true
[ ] SECURITY_HEADERS_ENABLED=true
[ ] Proper CSP configured
```

### 5. CORS

```
[ ] CORS_ALLOWED_ORIGINS - Specific production URLs only
[ ] No wildcard (*) origins
[ ] Credentials allowed only for known origins
```

---

## Application

### 6. Build Verification

```bash
# Backend
cd backend && go build -o app ./cmd
go test ./...
go vet ./...

# Frontend
cd frontend && npm run build
npm run typecheck
```

```
[ ] Backend builds without errors
[ ] Frontend builds without errors
[ ] All tests pass
[ ] No TypeScript errors
[ ] No linter errors
```

### 7. API Health

```
[ ] /health endpoint responds 200
[ ] /ready endpoint checks DB + cache
[ ] Graceful shutdown configured
[ ] Request timeouts set
```

### 8. Rate Limiting

```
[ ] RATE_LIMIT_ENABLED=true
[ ] Appropriate limits for auth endpoints
[ ] Appropriate limits for API endpoints
[ ] Trusted proxies configured (if behind LB)
```

---

## Third-Party Services

### 9. Stripe (if enabled)

```
[ ] STRIPE_SECRET_KEY - Live key (sk_live_*)
[ ] STRIPE_PUBLISHABLE_KEY - Live key (pk_live_*)
[ ] STRIPE_WEBHOOK_SECRET - Production webhook secret
[ ] Webhook endpoint registered in Stripe Dashboard
[ ] Checkout URLs point to production domain
[ ] Portal return URL points to production domain
```

**Test:**
```bash
# Verify webhook is receiving events
curl https://yourdomain.com/api/webhooks/stripe -X POST -d "{}" -I
# Should return 400 (invalid signature), not 404
```

### 10. OAuth (if enabled)

```
[ ] Google OAuth - Production client ID/secret
[ ] GitHub OAuth - Production client ID/secret
[ ] Redirect URIs updated to production domain
[ ] Callback URLs verified
```

### 11. Email (if enabled)

```
[ ] SMTP credentials - Production values
[ ] SMTP_FROM - Verified sender domain
[ ] Email templates tested
[ ] EMAIL_DEV_MODE=false
```

### 12. Monitoring

```
[ ] SENTRY_DSN - Production DSN configured
[ ] VITE_SENTRY_DSN - Frontend DSN configured
[ ] Log aggregation configured
[ ] Alerts set up for errors
[ ] Uptime monitoring active
```

---

## Infrastructure

### 13. Docker/Container

```
[ ] Multi-stage build (small image)
[ ] Non-root user in container
[ ] Health checks defined
[ ] Resource limits set
[ ] Secrets not baked into image
```

### 14. Load Balancer / Proxy

```
[ ] SSL termination configured
[ ] WebSocket connections supported
[ ] Proper proxy headers forwarded (X-Forwarded-For, etc.)
[ ] Connection draining on deploy
```

### 15. Cache (Redis/Dragonfly)

```
[ ] REDIS_PASSWORD set
[ ] Persistence configured (if needed)
[ ] Memory limits set
[ ] Connection to cache verified
```

---

## Post-Deployment

### 16. Smoke Tests

```
[ ] Homepage loads
[ ] Login works
[ ] Registration works (if public)
[ ] Protected routes redirect to login
[ ] API returns expected data
[ ] WebSocket connects
```

### 17. Monitoring Verification

```
[ ] Logs appearing in aggregator
[ ] Metrics being collected
[ ] No error spikes
[ ] Response times acceptable
```

### 18. Rollback Plan

```
[ ] Previous version tagged/available
[ ] Rollback procedure documented
[ ] Database migration rollback tested (if applicable)
[ ] Team knows rollback process
```

---

## Quick Checks

### Environment Diff

```bash
# Compare .env.example with production
diff .env.example .env.production

# Check for sensitive defaults
grep -E "(dev|test|example|change)" .env.production
```

### SSL Certificate

```bash
# Check SSL
curl -vI https://yourdomain.com 2>&1 | grep -E "(SSL|expire)"

# Check certificate expiry
echo | openssl s_client -connect yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates
```

### API Health

```bash
# Health check
curl https://yourdomain.com/health

# Ready check (DB + cache)
curl https://yourdomain.com/ready
```

---

## Output Format

When running this checklist, report:

```markdown
## Deployment Readiness: {Environment}

### Passed ✅
- Environment configuration
- Security secrets
- Build verification
...

### Failed ❌
- [ ] CORS_ALLOWED_ORIGINS still has localhost
- [ ] Stripe webhook secret not set
...

### Warnings ⚠️
- Consider enabling HSTS after initial deploy
- Sentry DSN not configured (optional but recommended)

### Ready to Deploy?
❌ No - Fix 2 critical issues first
```

---

## Related

- `/env-check` - Validate environment variables
- `security-audit.md` - Security review checklist
- `/full-stack-optimizer` - Performance optimization
