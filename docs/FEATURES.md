# Feature Documentation

This document consolidates all feature documentation for the React-Golang Starter Kit.

## Table of Contents

- [JWT Authentication & Security](#jwt-authentication--security)
- [Rate Limiting](#rate-limiting)
- [Role-Based Access Control (RBAC)](#role-based-access-control-rbac)
- [File Upload System](#file-upload-system)

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
