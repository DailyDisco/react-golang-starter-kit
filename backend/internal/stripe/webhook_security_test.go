package stripe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Additional security-focused tests for Stripe webhooks
// Complements webhook_signature_test.go with boundary and edge cases

func TestWebhookSecurity_TimestampBoundaries(t *testing.T) {
	testSecret := "whsec_test_boundaries"
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}
	handler := HandleWebhook(config)

	t.Run("timestamp 4 minutes ago - within tolerance", func(t *testing.T) {
		event := createTestEvent("evt_4min")
		payload, _ := json.Marshal(event)

		// 4 minutes ago - should be within Stripe's 5 minute tolerance
		timestamp := time.Now().Add(-4 * time.Minute).Unix()
		signature := generateTestSignature(payload, testSecret, timestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "4 minute old signature should be valid")
	})

	t.Run("timestamp 6 minutes ago - outside tolerance", func(t *testing.T) {
		event := createTestEvent("evt_6min")
		payload, _ := json.Marshal(event)

		// 6 minutes ago - should be outside Stripe's default tolerance
		timestamp := time.Now().Add(-6 * time.Minute).Unix()
		signature := generateTestSignature(payload, testSecret, timestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "6 minute old signature should be rejected")
	})

	t.Run("timestamp slightly in future - accepted within tolerance", func(t *testing.T) {
		event := createTestEvent("evt_future")
		payload, _ := json.Marshal(event)

		// Stripe's tolerance window is ±5 minutes, so future timestamps within
		// tolerance are accepted. This is expected behavior to handle clock skew.
		futureTimestamp := time.Now().Add(2 * time.Minute).Unix()
		signature := generateTestSignature(payload, testSecret, futureTimestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Stripe tolerates timestamps within ±5 minutes for clock skew
		assert.Equal(t, http.StatusOK, rr.Code,
			"near-future timestamp should be accepted (clock skew tolerance)")
	})
}

func TestWebhookSecurity_MalformedHeaders(t *testing.T) {
	testSecret := "whsec_test_malformed"
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}
	handler := HandleWebhook(config)

	tests := []struct {
		name           string
		signatureValue string
	}{
		{"missing timestamp", "v1=abc123"},
		{"missing signature", "t=1234567890"},
		{"malformed timestamp", "t=notanumber,v1=abc123"},
		{"empty signature value", "t=1234567890,v1="},
		{"empty timestamp value", "t=,v1=abc123"},
		{"no equals signs", "t1234567890v1abc123"},
		{"wrong delimiter", "t:1234567890;v1:abc123"},
		{"extra spaces", "t = 1234567890 , v1 = abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := createTestEvent("evt_" + tt.name)
			payload, _ := json.Marshal(event)

			req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Stripe-Signature", tt.signatureValue)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code,
				"malformed signature header should be rejected: %s", tt.name)
		})
	}
}

func TestWebhookSecurity_BodySizeBoundaries(t *testing.T) {
	testSecret := "whsec_test_bodysize"
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}
	handler := HandleWebhook(config)

	t.Run("body exactly at max size (65536 bytes)", func(t *testing.T) {
		// Create a payload that's exactly 65536 bytes
		// This is at the edge - MaxBytesReader is set to 65536
		largeData := make([]byte, 65536)
		for i := range largeData {
			largeData[i] = 'x'
		}

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(largeData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", "t=123,v1=abc")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Should fail on signature verification (but not body read)
		// MaxBytesReader with 65536 allows exactly 65536 bytes
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("body one byte over max (65537 bytes)", func(t *testing.T) {
		largeData := make([]byte, 65537)
		for i := range largeData {
			largeData[i] = 'x'
		}

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(largeData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", "t=123,v1=abc")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "body over 65536 bytes should be rejected")
	})
}

func TestWebhookSecurity_ContentTypeHandling(t *testing.T) {
	testSecret := "whsec_test_contenttype"
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}
	handler := HandleWebhook(config)

	t.Run("non-JSON content type still processes valid signature", func(t *testing.T) {
		event := createTestEvent("evt_contenttype")
		payload, _ := json.Marshal(event)

		timestamp := time.Now().Unix()
		signature := generateTestSignature(payload, testSecret, timestamp)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "text/plain") // Wrong content type
		req.Header.Set("Stripe-Signature", signature)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Should still work - Stripe cares about payload and signature, not Content-Type
		assert.Equal(t, http.StatusOK, rr.Code, "valid signature should work regardless of content-type")
	})

	t.Run("response always has application/json content type on error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", "invalid")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"),
			"error responses should be JSON")
	})
}

func TestWebhookSecurity_MultipleSignatures(t *testing.T) {
	testSecret := "whsec_test_multisig"
	config := &Config{
		SecretKey:     "sk_test_123",
		WebhookSecret: testSecret,
	}
	handler := HandleWebhook(config)

	t.Run("multiple v1 signatures - one valid", func(t *testing.T) {
		event := createTestEvent("evt_multisig")
		payload, _ := json.Marshal(event)

		timestamp := time.Now().Unix()
		validSig := generateTestSignature(payload, testSecret, timestamp)

		// Stripe can send multiple signatures during secret rotation
		// Format: t=timestamp,v1=sig1,v1=sig2
		// Extract just the signature from our generated header
		sigParts := strings.Split(validSig, ",")
		var v1Sig string
		for _, part := range sigParts {
			if strings.HasPrefix(part, "v1=") {
				v1Sig = strings.TrimPrefix(part, "v1=")
				break
			}
		}

		multiSigHeader := strings.ReplaceAll(validSig, "v1="+v1Sig, "v1=invalid123,v1="+v1Sig)

		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", multiSigHeader)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Stripe library should accept if any v1 signature is valid
		assert.Equal(t, http.StatusOK, rr.Code, "should accept when at least one signature is valid")
	})
}

// createTestEvent creates a minimal valid Stripe event for testing
func createTestEvent(id string) map[string]any {
	return map[string]any{
		"id":               id,
		"type":             "ping",
		"api_version":      "2023-10-16",
		"created":          time.Now().Unix(),
		"livemode":         false,
		"pending_webhooks": 1,
		"request": map[string]any{
			"id": "req_test_" + id,
		},
		"data": map[string]any{
			"object": map[string]any{},
		},
	}
}
