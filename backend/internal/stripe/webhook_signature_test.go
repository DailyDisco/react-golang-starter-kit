package stripe

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Webhook Signature Verification Tests
// These test the actual HTTP handler signature verification

// generateTestSignature creates a valid Stripe webhook signature for testing
func generateTestSignature(payload []byte, secret string, timestamp int64) string {
	// Stripe signature format: t=<timestamp>,v1=<signature>
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("t=%d,v1=%s", timestamp, signature)
}

func TestWebhookSignatureVerification(t *testing.T) {
	// Test webhook secret
	testSecret := "whsec_test_secret_12345"

	// Create a test config
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}

	handler := HandleWebhook(config)

	t.Run("valid signature accepted", func(t *testing.T) {
		// Create a test event payload with API version (required by Stripe SDK)
		// Use "ping" event type which doesn't require database operations
		event := map[string]any{
			"id":               "evt_test_valid",
			"type":             "ping", // Use ping - doesn't trigger DB operations
			"api_version":      "2023-10-16", // Must match Stripe SDK expected version
			"created":          time.Now().Unix(),
			"livemode":         false,
			"pending_webhooks": 1,
			"request": map[string]any{
				"id": "req_test_123",
			},
			"data": map[string]any{
				"object": map[string]any{},
			},
		}
		payload, _ := json.Marshal(event)

		timestamp := time.Now().Unix()
		signature := generateTestSignature(payload, testSecret, timestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got: %d, body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("expired signature rejected (timestamp tolerance)", func(t *testing.T) {
		event := map[string]any{
			"id":      "evt_test_expired",
			"type":    "ping",
			"created": time.Now().Unix(),
		}
		payload, _ := json.Marshal(event)

		// Use a timestamp from 10 minutes ago (Stripe's default tolerance is 5 minutes)
		expiredTimestamp := time.Now().Add(-10 * time.Minute).Unix()
		signature := generateTestSignature(payload, testSecret, expiredTimestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for expired signature, got: %d", rr.Code)
		}
	})

	t.Run("tampered payload rejected", func(t *testing.T) {
		event := map[string]any{
			"id":      "evt_test_original",
			"type":    "ping",
			"created": time.Now().Unix(),
		}
		payload, _ := json.Marshal(event)

		timestamp := time.Now().Unix()
		signature := generateTestSignature(payload, testSecret, timestamp)

		// Tamper with the payload after signing
		tamperedEvent := map[string]any{
			"id":      "evt_test_tampered", // Changed ID
			"type":    "ping",
			"created": time.Now().Unix(),
		}
		tamperedPayload, _ := json.Marshal(tamperedEvent)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(tamperedPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature) // Original signature

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for tampered payload, got: %d", rr.Code)
		}
	})

	t.Run("missing signature header rejected", func(t *testing.T) {
		event := map[string]any{
			"id":      "evt_test_nosig",
			"type":    "ping",
			"created": time.Now().Unix(),
		}
		payload, _ := json.Marshal(event)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		// No Stripe-Signature header

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing signature, got: %d", rr.Code)
		}
	})

	t.Run("invalid signature format rejected", func(t *testing.T) {
		event := map[string]any{
			"id":      "evt_test_badsig",
			"type":    "ping",
			"created": time.Now().Unix(),
		}
		payload, _ := json.Marshal(event)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", "invalid_signature_format")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid signature format, got: %d", rr.Code)
		}
	})

	t.Run("wrong secret rejected", func(t *testing.T) {
		event := map[string]any{
			"id":      "evt_test_wrongsecret",
			"type":    "ping",
			"created": time.Now().Unix(),
		}
		payload, _ := json.Marshal(event)

		timestamp := time.Now().Unix()
		// Sign with wrong secret
		signature := generateTestSignature(payload, "whsec_wrong_secret", timestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for wrong secret, got: %d", rr.Code)
		}
	})

	t.Run("body too large rejected", func(t *testing.T) {
		// Create a payload larger than 65KB
		largeData := make([]byte, 70000)
		for i := range largeData {
			largeData[i] = 'x'
		}

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(largeData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", "t=123,v1=abc")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for oversized body, got: %d", rr.Code)
		}
	})

	t.Run("empty body rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", "t=123,v1=abc")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty body, got: %d", rr.Code)
		}
	})
}

func TestWebhookSuccessResponse(t *testing.T) {
	testSecret := "whsec_test_secret_response"
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}

	handler := HandleWebhook(config)

	t.Run("returns proper JSON success response", func(t *testing.T) {
		event := map[string]any{
			"id":              "evt_test_response",
			"type":            "ping", // Unhandled type, but still returns success
			"api_version":     "2023-10-16",
			"created":         time.Now().Unix(),
			"livemode":        false,
			"pending_webhooks": 1,
			"request": map[string]any{
				"id": "req_test_456",
			},
			"data": map[string]any{
				"object": map[string]any{},
			},
		}
		payload, _ := json.Marshal(event)

		timestamp := time.Now().Unix()
		signature := generateTestSignature(payload, testSecret, timestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got: %d", rr.Code)
		}

		// Verify response is JSON
		if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got: %s", contentType)
		}

		// Verify response body
		var response map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Expected valid JSON response: %v", err)
		}

		if success, ok := response["success"].(bool); !ok || !success {
			t.Error("Expected success: true in response")
		}
	})

	t.Run("returns proper JSON error response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		// Invalid signature
		req.Header.Set("Stripe-Signature", "t=123,v1=invalid")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got: %d", rr.Code)
		}

		// Verify response is JSON
		if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got: %s", contentType)
		}

		// Verify error response structure
		var response map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Expected valid JSON error response: %v", err)
		}

		if _, ok := response["error"]; !ok {
			t.Error("Expected 'error' field in error response")
		}
		if _, ok := response["message"]; !ok {
			t.Error("Expected 'message' field in error response")
		}
	})
}
