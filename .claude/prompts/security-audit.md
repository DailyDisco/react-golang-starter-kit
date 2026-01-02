# Security Audit Prompt - React + Go Stack

Use this prompt when reviewing code for security vulnerabilities specific to this stack.

---

## Stack-Specific Security Concerns

### Authentication (JWT + Cookies)

```
[ ] JWT stored in httpOnly cookie (not localStorage)
[ ] Refresh token rotation enabled
[ ] Token expiration is reasonable (15-60 min access, 7-30 days refresh)
[ ] Token blacklist checked on sensitive operations
[ ] CSRF protection enabled (SameSite + CSRF token)
[ ] Secure flag set on cookies in production
```

### Authorization (RBAC)

```
[ ] Every protected endpoint checks authentication
[ ] Role checks happen at handler level
[ ] Organization membership verified for org routes
[ ] Owner-only operations properly restricted
[ ] No privilege escalation paths
```

### Input Validation

```
[ ] All user input validated (Go: validator, TS: Zod)
[ ] SQL injection prevented (GORM parameterized queries)
[ ] XSS prevented (React escapes by default, but check dangerouslySetInnerHTML)
[ ] Path traversal prevented (file uploads)
[ ] Request size limits enforced
```

### Data Protection

```
[ ] Passwords hashed with bcrypt (cost ≥ 10)
[ ] Sensitive data not logged (passwords, tokens, PII)
[ ] PII encrypted at rest (if required)
[ ] HTTPS enforced in production
[ ] Database connections use SSL
```

---

## OWASP Top 10 Checklist

### 1. Injection

**Go/GORM:**
```go
// VULNERABLE - string concatenation
db.Where("email = '" + email + "'").First(&user)

// SAFE - parameterized
db.Where("email = ?", email).First(&user)
```

**Check:**
```
[ ] No raw SQL with string concatenation
[ ] GORM methods used correctly
[ ] No shell command injection (exec.Command)
```

### 2. Broken Authentication

**Check:**
```
[ ] Brute force protection (rate limiting, account lockout)
[ ] Password complexity enforced
[ ] Session invalidation on logout
[ ] Session invalidation on password change
[ ] 2FA available for sensitive accounts
```

### 3. Sensitive Data Exposure

**Check:**
```
[ ] API keys not in source code
[ ] Secrets in environment variables
[ ] No sensitive data in URLs (use POST body)
[ ] Logs don't contain PII
[ ] Error messages don't leak internals
```

### 4. XML External Entities (XXE)

Generally not applicable (JSON API), but check:
```
[ ] No XML parsing (or if so, external entities disabled)
```

### 5. Broken Access Control

**Check:**
```
[ ] Can user A access user B's data?
[ ] Can member access admin endpoints?
[ ] Can non-owner delete organization?
[ ] File access checks ownership
[ ] API keys scoped to correct user
```

### 6. Security Misconfiguration

**Check:**
```
[ ] Debug mode disabled in production
[ ] Stack traces not exposed to users
[ ] Default credentials changed
[ ] CORS configured correctly (not *)
[ ] Security headers set (CSP, X-Frame-Options, etc.)
```

### 7. Cross-Site Scripting (XSS)

**Check:**
```
[ ] No dangerouslySetInnerHTML with user content
[ ] User content sanitized before display
[ ] CSP header configured
[ ] HttpOnly cookies (JS can't access tokens)
```

### 8. Insecure Deserialization

**Check:**
```
[ ] JSON unmarshaling to strict types (not interface{})
[ ] No pickle/gob with untrusted data
[ ] Request body size limits
```

### 9. Using Components with Known Vulnerabilities

**Check:**
```
[ ] npm audit shows no high/critical
[ ] go mod tidy && govulncheck shows no issues
[ ] Dependencies recently updated
```

### 10. Insufficient Logging & Monitoring

**Check:**
```
[ ] Authentication failures logged
[ ] Authorization failures logged
[ ] Input validation failures logged
[ ] Logs include correlation ID
[ ] Alerts on anomalies
```

---

## Code Patterns to Flag

### Go

```go
// FLAG: SQL injection risk
db.Exec("DELETE FROM users WHERE id = " + id)

// FLAG: Command injection
exec.Command("sh", "-c", userInput)

// FLAG: Path traversal
filepath.Join(baseDir, userInput) // if userInput contains ../

// FLAG: Sensitive data in logs
log.Info().Str("password", password).Msg("user login")

// FLAG: Hardcoded secrets
const apiKey = "sk-..."
```

### TypeScript/React

```typescript
// FLAG: XSS risk
<div dangerouslySetInnerHTML={{ __html: userContent }} />

// FLAG: Sensitive data in localStorage
localStorage.setItem('token', jwt)

// FLAG: Missing CSRF
fetch('/api/transfer', { method: 'POST', body })  // no credentials/CSRF

// FLAG: Eval with user input
eval(userCode)
```

---

## Security Headers Checklist

```
Content-Security-Policy: default-src 'self'; script-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

---

## Output Format

```markdown
## Security Audit: {Area/Feature}

### Critical Issues
- {Issue with immediate risk}

### High Priority
- {Issue that should be fixed soon}

### Medium Priority
- {Issue to address in normal development}

### Recommendations
1. {Specific fix}
2. {Another fix}

### Passed Checks
- {What looks good}
```

---

## Quick Commands

```bash
# Check Go vulnerabilities
govulncheck ./...

# Check npm vulnerabilities
npm audit

# Check for secrets in code
git secrets --scan  # if git-secrets installed

# Check security headers
curl -I https://yoursite.com
```
