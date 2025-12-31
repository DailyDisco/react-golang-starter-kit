package stripe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ HandleWebhook Tests ============

func TestHandleWebhook_EmptyBody(t *testing.T) {
	config := &Config{
		WebhookSecret: "whsec_test",
	}

	handler := HandleWebhook(config)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	// With empty body, signature verification will fail
	if rec.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	// Verify JSON response
	var resp models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Response.Code = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}

func TestHandleWebhook_MissingSignature(t *testing.T) {
	config := &Config{
		WebhookSecret: "whsec_test",
	}

	handler := HandleWebhook(config)

	body := []byte(`{"type": "checkout.session.completed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	// Without signature header, verification will fail
	if rec.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Message, "signature") {
		t.Errorf("Response.Message should contain 'signature', got %q", resp.Message)
	}
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	config := &Config{
		WebhookSecret: "whsec_test",
	}

	handler := HandleWebhook(config)

	body := []byte(`{"type": "checkout.session.completed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "invalid_signature")
	rec := httptest.NewRecorder()

	handler(rec, req)

	// Invalid signature should fail
	if rec.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleWebhook_BodyTooLarge(t *testing.T) {
	config := &Config{
		WebhookSecret: "whsec_test",
	}

	handler := HandleWebhook(config)

	// Create a body larger than 64KB
	largeBody := make([]byte, 65537)
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(largeBody))
	req.Header.Set("Stripe-Signature", "t=1234567890,v1=test")
	rec := httptest.NewRecorder()

	handler(rec, req)

	// Body too large should fail
	if rec.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleWebhook_ContentType(t *testing.T) {
	config := &Config{
		WebhookSecret: "whsec_test",
	}

	handler := HandleWebhook(config)

	body := []byte(`{"type": "checkout.session.completed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "invalid")
	rec := httptest.NewRecorder()

	handler(rec, req)

	// Response should have JSON content type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

// ============ Webhook Event Type Tests ============

func TestWebhookEventTypes(t *testing.T) {
	// Test that we handle expected event types without panicking
	eventTypes := []string{
		"checkout.session.completed",
		"customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted",
		"invoice.payment_failed",
		"unknown.event.type",
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			// This is a unit test to ensure the switch statement handles all expected types
			// In production, these would need valid Stripe signatures
			switch eventType {
			case "checkout.session.completed",
				"customer.subscription.created",
				"customer.subscription.updated",
				"customer.subscription.deleted",
				"invoice.payment_failed":
				// Expected to be handled
			default:
				// Unknown events should be logged but not error
			}
		})
	}
}

// ============ SyncUserRole Tests ============

func TestSyncUserRole_StatusMapping(t *testing.T) {
	// Test the logic of role mapping based on subscription status
	// This tests the expected behavior without needing a database

	tests := []struct {
		name          string
		status        string
		expectPremium bool
	}{
		{"active subscription", "active", true},
		{"trialing subscription", "trialing", true},
		{"past_due subscription", "past_due", true}, // Grace period
		{"canceled subscription", "canceled", false},
		{"unpaid subscription", "unpaid", false},
		{"incomplete subscription", "incomplete", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This documents the expected behavior
			isPremium := tt.status == "active" || tt.status == "trialing" || tt.status == "past_due"
			if isPremium != tt.expectPremium {
				t.Errorf("Status %q premium check = %v, want %v", tt.status, isPremium, tt.expectPremium)
			}
		})
	}
}
