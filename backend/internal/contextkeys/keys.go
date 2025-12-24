// Package contextkeys provides shared context key definitions to avoid import cycles.
package contextkeys

// RequestIDKeyType is the context key type for request ID
type RequestIDKeyType string

// RequestIDKey is the context key for storing the request ID
const RequestIDKey RequestIDKeyType = "request_id"

// RequestIDHeader is the HTTP header name for request ID
const RequestIDHeader = "X-Request-ID"
