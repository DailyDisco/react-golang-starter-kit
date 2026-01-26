# Feature Documentation

This document consolidates all feature documentation for the React-Golang Starter Kit.

## Table of Contents

- [JWT Authentication & Security](#jwt-authentication--security)
- [Rate Limiting](#rate-limiting)
- [Role-Based Access Control (RBAC)](#role-based-access-control-rbac)
- [File Upload System](#file-upload-system)
- [Multi-Tenancy (Organizations)](#multi-tenancy-organizations)
- [Background Jobs](#background-jobs)
- [Observability](#observability)

---

## JWT Authentication & Security

### Overview
This application uses JWT (JSON Web Tokens) for secure authentication with proper secret management.

### Security Best Practices

#### JWT Secret Management
- **Never use default values**: Always generate unique secrets for each environment
- **Use cryptographically secure generation**: Always use `openssl rand -hex 32` or equivalent
- **Store securely**: Use environment variables or secret management services (AWS Secrets Manager, Vault, etc.)
- **Rotate regularly**: Change secrets every 3-6 months
- **Different per environment**: Never reuse secrets across development, staging, and production

### Setup Instructions

#### For Development
1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Generate a secure JWT secret:
   ```bash
   ./generate-jwt-secret.sh
   ```

3. Add the generated secret to your `.env` file

#### For Production
1. Generate a new JWT secret:
   ```bash
   ./generate-jwt-secret.sh
   ```

2. Set the JWT_SECRET in your production environment variables (Railway, Vercel, etc.)

### Migration Guide

**Important**: Changing the JWT secret will invalidate all existing JWT tokens. Plan accordingly:

1. **Schedule maintenance window** for production deployment
2. **Notify users** about temporary service interruption
3. **Implement token refresh mechanism** if needed
4. **Consider gradual rollout** with both old and new secrets temporarily

### Troubleshooting

Common issues:
- **Application won't start**: Check if JWT_SECRET is set in environment
- **Users can't log in**: JWT secret changed, tokens invalidated
- **Docker container fails**: Ensure JWT_SECRET is passed to container environment

---

## Rate Limiting

### Overview
Rate limiting protects the API from abuse and ensures fair usage using Chi router's `httprate` middleware.

### Features
- **IP-based rate limiting**: Protects against abuse from individual IP addresses
- **User-based rate limiting**: Different limits for authenticated vs anonymous users
- **Endpoint-specific limits**: Stricter limits for authentication endpoints
- **Configurable via environment variables**: Easy to adjust limits without code changes

### Configuration

#### Global Settings
```bash
RATE_LIMIT_ENABLED=true
```

#### IP-Based Rate Limiting
```bash
RATE_LIMIT_IP_PER_MINUTE=60      # Requests per minute from a single IP
RATE_LIMIT_IP_PER_HOUR=1000      # Requests per hour from a single IP
```

#### User-Based Rate Limiting
```bash
RATE_LIMIT_USER_PER_MINUTE=120   # Requests per minute from authenticated user
RATE_LIMIT_USER_PER_HOUR=2000    # Requests per hour from authenticated user
```

#### Authentication Endpoints
```bash
RATE_LIMIT_AUTH_PER_MINUTE=5     # Strict limit for login/register
RATE_LIMIT_AUTH_PER_HOUR=20
```

### Rate Limit Response

When rate limited, the API returns:
- **Status Code**: `429 Too Many Requests`
- **Headers**:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Unix timestamp when the limit resets
  - `Retry-After`: Seconds to wait before retrying

### Best Practices
1. **Monitor Rate Limiting**: Log rate limit violations for analysis
2. **Adjust Limits Gradually**: Start with conservative limits and adjust based on usage
3. **Consider User Roles**: Implement role-based limits for premium users
4. **Use Redis in Production**: For multi-instance deployments

---

## Role-Based Access Control (RBAC)

### Overview
The RBAC system provides fine-grained access control with 4 user roles and a permission-based architecture.

### User Roles

#### 1. Super Admin (`super_admin`)
- **Purpose**: System-level administration
- **Permissions**: All permissions (user management, role management, premium content, etc.)

#### 2. Admin (`admin`)
- **Purpose**: Content and user administration
- **Permissions**: View/update users, premium content access, content management
- **Restrictions**: Cannot modify user roles

#### 3. Premium (`premium`)
- **Purpose**: Paid subscribers with enhanced features
- **Permissions**: Access to premium content, basic profile management

#### 4. User (`user`)
- **Purpose**: Regular users (default role for new registrations)
- **Permissions**: Basic profile management only

### API Endpoints

#### Public (No Authentication)
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/reset-password` - Password reset

#### Authenticated (Any Logged-in User)
- `GET /api/users/me` - Get current user profile
- `PUT /api/users/me` - Update current user profile

#### Premium (Premium + Admin + Super Admin)
- `GET /api/premium/content` - Access premium content

#### Admin (Admin + Super Admin)
- `GET /api/users/admin` - List all users
- `GET /api/users/admin/{id}` - Get specific user
- `PUT /api/users/admin/{id}` - Update user information
- `DELETE /api/users/admin/{id}` - Delete user

#### Super Admin Only
- `PUT /api/users/admin/{id}/role` - Update user role

### Security Features
- JWT tokens include user role information
- Tokens validated on every request
- Permission-based middleware on all protected endpoints
- Only Super Admins can assign roles
- Users cannot modify their own roles

### Testing
```bash
# Run automated test suite
./test_roles.sh
```

---

## File Upload System

### Overview
Comprehensive file upload system with dual storage backend: AWS S3 (preferred) or PostgreSQL database (fallback).

### Features
- **Dual Storage Backend**: Automatically uses S3 when configured, falls back to database
- **Secure Upload**: Requires authentication
- **File Management**: Complete CRUD operations
- **Public Downloads**: Downloads are publicly accessible
- **Rate Limiting**: Built-in protection against abuse
- **File Size Limits**: Configurable (default 10MB)

### API Endpoints

#### Upload File
```
POST /api/files/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>
```

#### Download File
```
GET /api/files/{id}/download
```

#### Get File Information
```
GET /api/files/{id}
Authorization: Bearer <token>
```

#### List Files
```
GET /api/files?limit=10&offset=0
Authorization: Bearer <token>
```

#### Delete File
```
DELETE /api/files/{id}
Authorization: Bearer <token>
```

#### Storage Status
```
GET /api/files/storage/status
```

### Configuration

#### Environment Variables
```bash
# AWS S3 Configuration (Optional)
AWS_ACCESS_KEY_ID=your-aws-access-key-id
AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key
AWS_REGION=us-east-1
AWS_S3_BUCKET=your-s3-bucket-name

# File Upload Configuration
MAX_FILE_SIZE_MB=10
```

### AWS S3 Setup

1. **Create S3 Bucket**: Create a bucket in AWS S3 Console
2. **Create IAM User**: Create user with programmatic access
3. **Attach Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
      "Resource": "arn:aws:s3:::your-bucket-name/*"
    }
  ]
}
```

### Storage Backend Selection

The system automatically chooses storage based on configuration:

1. **S3 Storage** (Preferred):
   - Used when all AWS credentials are provided
   - Files uploaded to S3, metadata in database
   - Better for large files and high traffic

2. **Database Storage** (Fallback):
   - Used when AWS credentials are missing
   - Files stored as BLOB in PostgreSQL
   - Suitable for development/small scale use

### Usage Example (JavaScript)

```javascript
// Upload file
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const response = await fetch('/api/files/upload', {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${token}`,
  },
  body: formData,
});

const result = await response.json();
console.log('Uploaded file ID:', result.data.id);

// Download file
const urlResponse = await fetch(`/api/files/${fileId}/url`);
const urlData = await urlResponse.json();
window.open(urlData.data, '_blank');
```

### Security Considerations
- File uploads require JWT authentication
- 10MB default file size limit (configurable)
- Minimal IAM permissions for S3 access
- Built-in rate limiting

### Troubleshooting

**S3 Upload Fails**:
- Check AWS credentials
- Verify S3 bucket permissions
- Check bucket region configuration

**Database Storage Issues**:
- Check PostgreSQL connection
- Verify BYTEA column size limits
- Monitor database storage space

---

## Multi-Tenancy (Organizations)

### Overview

The organization system provides multi-tenancy support, allowing users to belong to multiple organizations with different roles.

### Organization Model

```
Organization
├── id, name, slug (unique URL-friendly identifier)
├── owner_id (user who created the organization)
├── seat_limit (max members, null = unlimited)
├── subscription_id (linked billing)
└── members[] (OrganizationMember)
    ├── user_id
    ├── role (owner, admin, member)
    └── joined_at
```

### Organization Roles

| Role | Permissions |
|------|-------------|
| **Owner** | Full control, billing, delete org, transfer ownership |
| **Admin** | Manage members, invite users, manage settings |
| **Member** | Access org resources, view members |

### API Endpoints

```
# Organization CRUD
GET    /api/v1/organizations              # List user's orgs
POST   /api/v1/organizations              # Create org
GET    /api/v1/organizations/:slug        # Get org details
PUT    /api/v1/organizations/:slug        # Update org
DELETE /api/v1/organizations/:slug        # Delete org (owner only)

# Member Management
GET    /api/v1/organizations/:slug/members           # List members
PUT    /api/v1/organizations/:slug/members/:id/role  # Change role
DELETE /api/v1/organizations/:slug/members/:id       # Remove member

# Invitations
POST   /api/v1/organizations/:slug/invitations       # Invite user
GET    /api/v1/organizations/:slug/invitations       # List pending
DELETE /api/v1/organizations/:slug/invitations/:id   # Cancel invitation
POST   /api/v1/invitations/:token/accept             # Accept invitation
```

### Invitation Flow

1. Admin invites user by email
2. System creates invitation with unique token (expires in 7 days)
3. User receives email with invitation link
4. User clicks link → creates account or logs in → accepts invitation
5. User becomes member of organization

### Data Isolation

Organization data is isolated via `org_id` foreign keys:
- Each resource query includes `WHERE org_id = ?`
- Middleware validates user's org membership before allowing access
- Cross-org access is prevented at the service layer

### Seat Limits

Organizations can have seat limits based on subscription tier:
- `seat_limit = NULL`: Unlimited members
- `seat_limit = 5`: Maximum 5 members (invitations fail beyond limit)

---

## Background Jobs

### Overview

Background jobs are powered by [River](https://github.com/riverqueue/river), a PostgreSQL-backed job queue for Go.

### Why River?

- **PostgreSQL-native**: Uses the same database, no Redis needed
- **Transactional**: Jobs can be enqueued within database transactions
- **Reliable**: At-least-once delivery with automatic retries
- **Observable**: Built-in job status tracking and metrics

### Available Jobs

| Job | Purpose |
|-----|---------|
| `data_export` | Export user data (GDPR compliance) |
| `cleanup` | Clean up expired sessions, tokens |
| `retention` | Apply data retention policies |
| `email` | Send transactional emails |

### Configuration

```bash
# Environment variables
RIVER_WORKERS=5              # Number of worker goroutines
RIVER_POLL_INTERVAL=1s       # How often to check for jobs
RIVER_MAX_ATTEMPTS=3         # Retry attempts before marking failed
```

### Enqueueing Jobs

```go
// In a handler or service
import "github.com/riverqueue/river"

// Simple job
_, err := riverClient.Insert(ctx, workers.DataExportArgs{
    UserID: userID,
    Format: "json",
}, nil)

// Scheduled job (run in 1 hour)
_, err := riverClient.Insert(ctx, workers.CleanupArgs{}, &river.InsertOpts{
    ScheduledAt: time.Now().Add(time.Hour),
})

// With transaction (job only created if transaction commits)
tx := db.Begin()
_, err := riverClient.InsertTx(ctx, tx, args, nil)
tx.Commit()
```

### Job Status

Jobs progress through these states:
1. `available` → Ready to be picked up
2. `running` → Currently executing
3. `completed` → Successfully finished
4. `retryable` → Failed, will retry
5. `discarded` → Failed, max attempts reached

### Monitoring

```sql
-- View pending jobs
SELECT * FROM river_job WHERE state = 'available';

-- View failed jobs
SELECT * FROM river_job WHERE state = 'discarded';

-- Job stats
SELECT state, COUNT(*) FROM river_job GROUP BY state;
```

---

## Observability

### Overview

The application includes comprehensive observability features: structured logging, metrics, health checks, and error tracking.

### Structured Logging

Uses [zerolog](https://github.com/rs/zerolog) for JSON-structured logs:

```go
import "github.com/rs/zerolog/log"

log.Info().
    Str("user_id", userID).
    Str("action", "login").
    Msg("User logged in")

// Output:
// {"level":"info","user_id":"123","action":"login","message":"User logged in","time":"..."}
```

**Log Levels:**
- `debug` - Detailed debugging (development only)
- `info` - Normal operations
- `warn` - Potential issues
- `error` - Errors requiring attention

**Configuration:**
```bash
LOG_LEVEL=info        # Minimum level to output
LOG_FORMAT=json       # json or console
```

### Prometheus Metrics

Metrics are exposed at `/metrics` in Prometheus format.

**Available Metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `http_request_duration_seconds` | Histogram | Request latency distribution |
| `db_queries_total` | Counter | Database queries by operation |
| `db_query_duration_seconds` | Histogram | Query latency |
| `websocket_connections` | Gauge | Active WebSocket connections |
| `auth_attempts_total` | Counter | Login attempts by success/failure |
| `river_jobs_total` | Counter | Background jobs by state |

**Grafana Dashboard:**

Import the provided dashboard from `docker/grafana/dashboards/app-dashboard.json`.

### Health Checks

```
GET /health          # Basic health (returns 200 OK)
GET /health/ready    # Readiness (checks DB, Redis connections)
GET /health/live     # Liveness (always returns 200)
```

**Readiness Response:**
```json
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "cache": "ok"
  },
  "version": "1.0.0"
}
```

### Error Tracking (Sentry)

Sentry integration captures and reports errors.

**Configuration:**
```bash
SENTRY_DSN=https://xxx@sentry.io/xxx
SENTRY_ENVIRONMENT=production
SENTRY_TRACES_SAMPLE_RATE=0.1    # 10% of requests traced
```

**Features:**
- Automatic panic recovery and reporting
- Request context (user ID, URL, headers)
- Distributed tracing with correlation IDs
- Performance monitoring

### Correlation IDs

Every request receives a unique `X-Request-ID` header for tracing:

1. Request arrives → middleware assigns ID
2. ID propagated through context
3. All logs include the request ID
4. Response includes `X-Request-ID` header
5. Client can use ID for support requests

```go
// Access in handlers/services
requestID := middleware.GetReqID(ctx)
log.Info().Str("request_id", requestID).Msg("Processing request")
```

### Observability Stack (Docker)

Start the full observability stack:

```bash
docker compose -f compose.observability.yml up -d
```

**Services:**
- **Prometheus** (`:9090`) - Metrics collection
- **Grafana** (`:3001`) - Dashboards and visualization
- **Loki** (optional) - Log aggregation

### Best Practices

1. **Always include context** - Add relevant fields to log entries
2. **Use correlation IDs** - Include request_id in all logs
3. **Don't log sensitive data** - Mask passwords, tokens, PII
4. **Monitor error rates** - Alert on sudden increases
5. **Set up dashboards** - Visualize key metrics
