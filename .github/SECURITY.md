# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < main  | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Private Vulnerability Reporting** (preferred)
   - Go to the Security tab of this repository
   - Click "Report a vulnerability"
   - Fill out the form with details

2. **Email**
   - Contact the maintainer directly via their GitHub profile

### What to Include

When reporting a vulnerability, please include:

- Type of vulnerability (e.g., XSS, SQL injection, authentication bypass)
- Full paths of affected source files
- Location of the affected code (file, line number if known)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact assessment and potential attack scenarios

### Response Timeline

- **Initial response**: Within 48 hours
- **Status update**: Within 7 days
- **Fix timeline**: Depends on severity
  - Critical: 7 days
  - High: 30 days
  - Medium: 90 days
  - Low: Next release

### Disclosure Policy

We follow coordinated disclosure. Please allow up to 90 days before public disclosure to give us time to develop and release a fix.

## Security Best Practices

This project follows security best practices including:

- **Authentication**: JWT with httpOnly cookies, token rotation
- **Authorization**: Role-based access control (RBAC)
- **Input Validation**: Zod (frontend), go-playground/validator (backend)
- **SQL Injection Prevention**: Parameterized queries via GORM
- **XSS Prevention**: React's built-in escaping, CSP headers
- **CSRF Protection**: SameSite cookies, CSRF tokens for mutations
- **Dependency Scanning**: Dependabot, govulncheck, npm audit
- **Container Scanning**: Trivy in CI/CD pipeline

## Security-Related Configuration

### Environment Variables

Never commit secrets. Required security-related environment variables:

- `JWT_SECRET` - Must be at least 32 characters
- `DB_PASSWORD` - Database password
- `REDIS_URL` - Redis connection (if using auth)

### Recommended Headers

The application sets the following security headers:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` (in production)
