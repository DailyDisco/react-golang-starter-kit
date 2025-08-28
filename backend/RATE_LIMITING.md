# Rate Limiting Implementation

This document describes the rate limiting implementation for the React-Golang Starter Kit API.

## Overview

Rate limiting is implemented using the Chi router's `httprate` middleware to protect the API from abuse and ensure fair usage. The implementation supports different rate limits for different types of endpoints and users.

## Features

- **IP-based rate limiting**: Protects against abuse from individual IP addresses
- **User-based rate limiting**: Different limits for authenticated vs anonymous users
- **Endpoint-specific limits**: Stricter limits for authentication endpoints
- **Configurable via environment variables**: Easy to adjust limits without code changes
- **Multiple time windows**: Per-minute and per-hour limits
- **Graceful degradation**: Returns appropriate HTTP status codes and headers

## Configuration

Rate limiting is configured via environment variables. All settings are optional and have sensible defaults.

### Global Settings

```bash
# Enable/disable rate limiting globally
RATE_LIMIT_ENABLED=true
```

### IP-Based Rate Limiting

These limits apply to all requests based on IP address:

```bash
# Requests per minute from a single IP
RATE_LIMIT_IP_PER_MINUTE=60

# Requests per hour from a single IP
RATE_LIMIT_IP_PER_HOUR=1000

# Burst size - allows short bursts above the per-minute limit
RATE_LIMIT_IP_BURST_SIZE=10
```

### User-Based Rate Limiting

These limits apply to authenticated users (less restrictive):

```bash
# Requests per minute from an authenticated user
RATE_LIMIT_USER_PER_MINUTE=120

# Requests per hour from an authenticated user
RATE_LIMIT_USER_PER_HOUR=2000

# Burst size for authenticated users
RATE_LIMIT_USER_BURST_SIZE=20
```

### Authentication Endpoints

Stricter limits for auth endpoints to prevent brute force attacks:

```bash
# Requests per minute for auth endpoints (login, register, etc.)
RATE_LIMIT_AUTH_PER_MINUTE=5

# Requests per hour for auth endpoints
RATE_LIMIT_AUTH_PER_HOUR=20

# Burst size for auth endpoints
RATE_LIMIT_AUTH_BURST_SIZE=2
```

### General API Endpoints

Balanced limits for most API endpoints:

```bash
# Requests per minute for general API endpoints
RATE_LIMIT_API_PER_MINUTE=100

# Requests per hour for general API endpoints
RATE_LIMIT_API_PER_HOUR=1500

# Burst size for API endpoints
RATE_LIMIT_API_BURST_SIZE=15
```

## How It Works

### Rate Limiting Strategy

1. **Global IP Limiting**: All requests are first checked against IP-based limits
2. **Endpoint-Specific Limiting**: Different endpoints have different rate limiting strategies:
   - **Auth endpoints** (`/api/auth/*`): Strict IP-based limiting
   - **Protected endpoints**: User-based limiting (authenticated users get higher limits)
   - **Public API endpoints**: Mixed IP/user-based limiting

### Rate Limit Headers

When a request is rate limited, the API returns:

- **Status Code**: `429 Too Many Requests`
- **Headers**:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Unix timestamp when the limit resets
  - `Retry-After`: Seconds to wait before retrying

### Example Response

```http
HTTP/1.1 429 Too Many Requests
Content-Type: text/plain; charset=utf-8
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1693526400
Retry-After: 60

Rate limit exceeded. Too many requests from this IP address.
```

## Implementation Details

### Middleware Functions

The rate limiting is implemented through several middleware functions:

- `NewIPRateLimitMiddleware()`: Global IP-based rate limiting
- `NewAuthRateLimitMiddleware()`: Strict rate limiting for auth endpoints
- `NewUserRateLimitMiddleware()`: User-based rate limiting for authenticated users
- `NewAPIRateLimitMiddleware()`: General API rate limiting

### Key Generation

Different strategies use different keys:

- **IP-based**: Uses client IP address
- **User-based**: Uses user ID from JWT context, falls back to IP
- **Auth endpoints**: Uses IP address (prevents brute force attacks)

### Storage

Rate limiting data is stored in memory using httprate's default local counter. For production deployments with multiple instances, consider using Redis with `WithLimitCounter()`.

## Usage Examples

### Development (Relaxed Limits)

```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_IP_PER_MINUTE=100
RATE_LIMIT_USER_PER_MINUTE=200
RATE_LIMIT_AUTH_PER_MINUTE=10
```

### Production (Strict Limits)

```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_IP_PER_MINUTE=30
RATE_LIMIT_USER_PER_MINUTE=100
RATE_LIMIT_AUTH_PER_MINUTE=3
```

### Disable Rate Limiting

```bash
RATE_LIMIT_ENABLED=false
```

## Best Practices

1. **Monitor Rate Limiting**: Log rate limit violations for analysis
2. **Adjust Limits Gradually**: Start with conservative limits and adjust based on usage
3. **Consider User Roles**: Implement role-based limits for premium users
4. **Use Redis in Production**: For multi-instance deployments
5. **Set Appropriate Retry-After**: Help clients implement proper backoff

## Troubleshooting

### Common Issues

1. **Rate Limited During Development**: Increase limits or disable rate limiting
2. **Inconsistent Limits**: Check that all instances use the same Redis instance
3. **Proxy Issues**: Ensure proper IP forwarding headers are configured

### Debugging

Enable debug logging to see rate limiting decisions:

```bash
# The middleware logs when rate limits are exceeded
# Check your application logs for rate limiting events
```

## Future Enhancements

- **Redis Storage**: For distributed rate limiting
- **Role-Based Limits**: Different limits based on user roles
- **Dynamic Limits**: Adjust limits based on server load
- **Rate Limit Analytics**: Dashboard for monitoring usage patterns
- **Whitelist/Blacklist**: IP-based allow/deny lists
