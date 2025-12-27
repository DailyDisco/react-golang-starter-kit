package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRequest helps build test HTTP requests.
type TestRequest struct {
	method  string
	path    string
	body    interface{}
	headers map[string]string
}

// NewRequest creates a new test request builder.
func NewRequest(method, path string) *TestRequest {
	return &TestRequest{
		method:  method,
		path:    path,
		headers: make(map[string]string),
	}
}

// GET creates a new GET request.
func GET(path string) *TestRequest {
	return NewRequest(http.MethodGet, path)
}

// POST creates a new POST request.
func POST(path string) *TestRequest {
	return NewRequest(http.MethodPost, path)
}

// PUT creates a new PUT request.
func PUT(path string) *TestRequest {
	return NewRequest(http.MethodPut, path)
}

// PATCH creates a new PATCH request.
func PATCH(path string) *TestRequest {
	return NewRequest(http.MethodPatch, path)
}

// DELETE creates a new DELETE request.
func DELETE(path string) *TestRequest {
	return NewRequest(http.MethodDelete, path)
}

// WithJSON sets the request body as JSON.
func (r *TestRequest) WithJSON(body interface{}) *TestRequest {
	r.body = body
	r.headers["Content-Type"] = "application/json"
	return r
}

// WithBody sets the raw request body.
func (r *TestRequest) WithBody(body interface{}) *TestRequest {
	r.body = body
	return r
}

// WithHeader sets a request header.
func (r *TestRequest) WithHeader(key, value string) *TestRequest {
	r.headers[key] = value
	return r
}

// WithAuth sets the Authorization header with a Bearer token.
func (r *TestRequest) WithAuth(token string) *TestRequest {
	r.headers["Authorization"] = "Bearer " + token
	return r
}

// WithCSRF sets the CSRF token header.
func (r *TestRequest) WithCSRF(token string) *TestRequest {
	r.headers["X-CSRF-Token"] = token
	return r
}

// WithRequestID sets the request ID header.
func (r *TestRequest) WithRequestID(id string) *TestRequest {
	r.headers["X-Request-ID"] = id
	return r
}

// Build creates the http.Request from the builder.
func (r *TestRequest) Build(t *testing.T) *http.Request {
	t.Helper()

	var bodyReader io.Reader
	if r.body != nil {
		switch v := r.body.(type) {
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = bytes.NewReader([]byte(v))
		case io.Reader:
			bodyReader = v
		default:
			// Assume it's a struct that should be JSON encoded
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	req := httptest.NewRequest(r.method, r.path, bodyReader)
	for key, value := range r.headers {
		req.Header.Set(key, value)
	}

	return req
}

// ResponseAsserter helps validate HTTP responses.
type ResponseAsserter struct {
	t   *testing.T
	rec *httptest.ResponseRecorder
}

// AssertResponse creates a new response asserter.
func AssertResponse(t *testing.T, rec *httptest.ResponseRecorder) *ResponseAsserter {
	t.Helper()
	return &ResponseAsserter{
		t:   t,
		rec: rec,
	}
}

// Status asserts the response status code.
func (a *ResponseAsserter) Status(expected int) *ResponseAsserter {
	a.t.Helper()
	if a.rec.Code != expected {
		a.t.Errorf("expected status %d, got %d; body: %s", expected, a.rec.Code, a.rec.Body.String())
	}
	return a
}

// StatusOK asserts the response status is 200 OK.
func (a *ResponseAsserter) StatusOK() *ResponseAsserter {
	return a.Status(http.StatusOK)
}

// StatusCreated asserts the response status is 201 Created.
func (a *ResponseAsserter) StatusCreated() *ResponseAsserter {
	return a.Status(http.StatusCreated)
}

// StatusBadRequest asserts the response status is 400 Bad Request.
func (a *ResponseAsserter) StatusBadRequest() *ResponseAsserter {
	return a.Status(http.StatusBadRequest)
}

// StatusUnauthorized asserts the response status is 401 Unauthorized.
func (a *ResponseAsserter) StatusUnauthorized() *ResponseAsserter {
	return a.Status(http.StatusUnauthorized)
}

// StatusForbidden asserts the response status is 403 Forbidden.
func (a *ResponseAsserter) StatusForbidden() *ResponseAsserter {
	return a.Status(http.StatusForbidden)
}

// StatusNotFound asserts the response status is 404 Not Found.
func (a *ResponseAsserter) StatusNotFound() *ResponseAsserter {
	return a.Status(http.StatusNotFound)
}

// StatusInternalServerError asserts the response status is 500 Internal Server Error.
func (a *ResponseAsserter) StatusInternalServerError() *ResponseAsserter {
	return a.Status(http.StatusInternalServerError)
}

// HasJSON asserts the response has JSON content type.
func (a *ResponseAsserter) HasJSON() *ResponseAsserter {
	a.t.Helper()
	contentType := a.rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		a.t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
	return a
}

// HasHeader asserts a specific header is present with the expected value.
func (a *ResponseAsserter) HasHeader(key, expected string) *ResponseAsserter {
	a.t.Helper()
	actual := a.rec.Header().Get(key)
	if actual != expected {
		a.t.Errorf("expected header %s=%s, got %s", key, expected, actual)
	}
	return a
}

// HasHeaderContaining asserts a specific header contains the expected substring.
func (a *ResponseAsserter) HasHeaderContaining(key, substring string) *ResponseAsserter {
	a.t.Helper()
	actual := a.rec.Header().Get(key)
	if !bytes.Contains([]byte(actual), []byte(substring)) {
		a.t.Errorf("expected header %s to contain %s, got %s", key, substring, actual)
	}
	return a
}

// BodyContains asserts the response body contains the expected string.
func (a *ResponseAsserter) BodyContains(expected string) *ResponseAsserter {
	a.t.Helper()
	if !bytes.Contains(a.rec.Body.Bytes(), []byte(expected)) {
		a.t.Errorf("expected body to contain %q, got: %s", expected, a.rec.Body.String())
	}
	return a
}

// BodyNotContains asserts the response body does not contain the expected string.
func (a *ResponseAsserter) BodyNotContains(notExpected string) *ResponseAsserter {
	a.t.Helper()
	if bytes.Contains(a.rec.Body.Bytes(), []byte(notExpected)) {
		a.t.Errorf("expected body not to contain %q, but it does: %s", notExpected, a.rec.Body.String())
	}
	return a
}

// JSON unmarshals the response body into the destination.
func (a *ResponseAsserter) JSON(dest interface{}) *ResponseAsserter {
	a.t.Helper()
	if err := json.Unmarshal(a.rec.Body.Bytes(), dest); err != nil {
		a.t.Fatalf("failed to unmarshal response body: %v; body: %s", err, a.rec.Body.String())
	}
	return a
}

// Body returns the response body as a string.
func (a *ResponseAsserter) Body() string {
	return a.rec.Body.String()
}

// Recorder returns the underlying httptest.ResponseRecorder.
func (a *ResponseAsserter) Recorder() *httptest.ResponseRecorder {
	return a.rec
}

// NewRecorder creates a new httptest.ResponseRecorder.
func NewRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// ExecuteHandler executes an http.Handler with the given request and returns an asserter.
func ExecuteHandler(t *testing.T, handler http.Handler, req *http.Request) *ResponseAsserter {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return AssertResponse(t, rec)
}
