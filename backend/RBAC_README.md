# Role-Based Access Control (RBAC) System

This document describes the comprehensive RBAC system implemented in the React Go Starter Kit.

## Overview

The RBAC system provides fine-grained access control with 4 user roles and a permission-based architecture that ensures users can only access resources they're authorized to use.

## User Roles

### 1. Super Admin (`super_admin`)

- **Purpose**: System-level administration
- **Permissions**:
  - All permissions across the entire system
  - User management (create, update, delete, view)
  - Role management
  - System administration
  - Premium content access
  - Content management

### 2. Admin (`admin`)

- **Purpose**: Content and user administration
- **Permissions**:
  - View all users
  - Update user information
  - Premium content access
  - Content management
  - Cannot modify user roles (security measure)

### 3. Premium (`premium`)

- **Purpose**: Paid subscribers with enhanced features
- **Permissions**:
  - Access to premium content
  - Basic user profile management

### 4. User (`user`)

- **Purpose**: Regular users (default role for new registrations)
- **Permissions**:
  - Basic user profile management
  - No access to premium or admin features

## Permission System

The system uses a permission-based architecture where roles are mapped to specific permissions:

```go
// Permission constants
const (
    PermViewUsers     Permission = "users:view"
    PermCreateUsers   Permission = "users:create"
    PermUpdateUsers   Permission = "users:update"
    PermDeleteUsers   Permission = "users:delete"
    PermManageRoles   Permission = "users:manage_roles"
    PermViewPremium   Permission = "content:premium"
    PermManageContent Permission = "content:manage"
    PermSystemAdmin   Permission = "system:admin"
)

// Role-to-permission mappings
var RolePermissions = map[string][]Permission{
    RoleSuperAdmin: {PermViewUsers, PermCreateUsers, /* ... all permissions */},
    RoleAdmin:      {PermViewUsers, PermUpdateUsers, /* ... admin permissions */},
    RolePremium:    {PermViewPremium},
    RoleUser:       {/* basic permissions only */},
}
```

## API Endpoints

### Public Endpoints (No Authentication Required)

```
GET  /health                    # Health check
POST /api/users                 # User registration
POST /api/auth/login            # User login
POST /api/auth/register         # User registration (alternative)
POST /api/auth/reset-password   # Password reset
```

### Authenticated Endpoints (Any Logged-in User)

```
GET  /api/users/me              # Get current user profile
PUT  /api/users/me              # Update current user profile
GET  /api/auth/me               # Get current user info (alternative)
```

### Premium Endpoints (Premium + Admin + Super Admin)

```
GET  /api/premium/content       # Access premium content
```

### Admin Endpoints (Admin + Super Admin Only)

```
GET  /api/users/admin           # List all users
GET  /api/users/admin/{id}      # Get specific user details
PUT  /api/users/admin/{id}      # Update user information
PUT  /api/users/admin/{id}/role # Update user role (Super Admin only)
DELETE /api/users/admin/{id}    # Delete user
```

## Usage Examples

### User Registration (Creates user with default "user" role)

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### User Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### Access User Profile (Authenticated)

```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Access Premium Content (Premium users only)

```bash
curl -X GET http://localhost:8080/api/premium/content \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### List All Users (Admin only)

```bash
curl -X GET http://localhost:8080/api/users/admin \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update User Role (Super Admin only)

```bash
curl -X PUT http://localhost:8080/api/users/admin/123/role \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role": "premium"}'
```

## Security Features

### 1. JWT Token Security

- Tokens include user role information
- Tokens are validated on every request
- Automatic token expiration (24 hours by default)

### 2. Permission-Based Middleware

- Each protected endpoint uses permission-based middleware
- Middleware checks user role against required permissions
- Clear error messages for unauthorized access

### 3. Role Management Security

- Users cannot modify their own roles
- Only Super Admins can assign roles
- Role validation prevents invalid role assignments

### 4. Database Security

- User roles are stored in the database
- Foreign key constraints ensure data integrity
- Role changes are logged and cached appropriately

## Database Schema

The `users` table includes a `role` column:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Testing

### Automated Testing

Run the automated test suite:

```bash
./test_roles.sh
```

### Manual Testing Setup

1. Start the environment:

```bash
./setup_test_env.sh
```

2. Test users created:
   - **Regular User**: `testuser@example.com` / `TestPass123!`
   - **Admin User**: `admin@example.com` / `TestPass123!`
   - **Premium User**: `premium@example.com` / `TestPass123!`
   - **Super Admin**: `superadmin@example.com` / `TestPass123!`

### Testing Scenarios

1. **Regular User Access**:

   - ✅ Can access own profile
   - ✅ Can update own profile
   - ❌ Cannot access premium content
   - ❌ Cannot access admin endpoints

2. **Premium User Access**:

   - ✅ Can access own profile
   - ✅ Can access premium content
   - ❌ Cannot access admin endpoints

3. **Admin User Access**:

   - ✅ Can access own profile
   - ✅ Can access premium content
   - ✅ Can list all users
   - ✅ Can update user information
   - ❌ Cannot modify user roles

4. **Super Admin Access**:
   - ✅ All permissions from Admin
   - ✅ Can modify user roles
   - ✅ Can delete users

## Error Handling

The system provides clear error messages:

- **401 Unauthorized**: Missing or invalid authentication token
- **403 Forbidden**: Insufficient permissions for the requested action
- **404 Not Found**: User or resource not found
- **409 Conflict**: Email already exists (registration)
- **500 Internal Server Error**: Server-side errors

## Best Practices

### 1. Role Assignment

- New users get the "user" role by default
- Role changes should be logged for audit purposes
- Use Super Admin sparingly for security

### 2. Permission Design

- Keep permissions granular and specific
- Use permission groups for related actions
- Regularly review and update permissions

### 3. Security Considerations

- Store JWT secrets securely
- Use HTTPS in production
- Implement rate limiting
- Log security events

### 4. Performance

- User roles are cached with the JWT token
- Database queries are optimized
- Use Redis for session management

## Extending the System

### Adding New Roles

1. Add the role constant to `models/models.go`
2. Add permissions to `RolePermissions` map
3. Update the role hierarchy if needed

### Adding New Permissions

1. Add permission constant to `permissions.go`
2. Map the permission to appropriate roles
3. Create middleware for the new permission
4. Update route handlers

### Adding New Protected Endpoints

1. Choose appropriate permission(s)
2. Apply permission middleware to the route
3. Handle authorization errors appropriately

## Troubleshooting

### Common Issues

1. **403 Forbidden Errors**:

   - Check user role in database
   - Verify JWT token contains correct role
   - Ensure middleware is applied correctly

2. **Database Migration Issues**:

   - Run migrations manually
   - Check database connection
   - Verify table schema

3. **JWT Token Issues**:
   - Check JWT_SECRET environment variable
   - Verify token expiration
   - Ensure proper token format

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
export DEBUG=true
```

### Database Queries

Check user roles directly:

```sql
SELECT id, name, email, role FROM users;
```

## Support

For issues or questions about the RBAC system:

1. Check the test output for error messages
2. Review the application logs
3. Verify database connectivity
4. Ensure proper environment configuration
