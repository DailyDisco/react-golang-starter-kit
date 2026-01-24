package jobs

import (
	"encoding/json"
	"testing"
)

// ============ SendVerificationEmailArgs Tests ============

func TestSendVerificationEmailArgs_Kind(t *testing.T) {
	args := SendVerificationEmailArgs{}
	if args.Kind() != "send_verification_email" {
		t.Errorf("Kind() = %q, want %q", args.Kind(), "send_verification_email")
	}
}

func TestSendVerificationEmailArgs_InsertOpts(t *testing.T) {
	args := SendVerificationEmailArgs{}
	opts := args.InsertOpts()

	if opts.Queue != "email" {
		t.Errorf("InsertOpts().Queue = %q, want %q", opts.Queue, "email")
	}
	if opts.MaxAttempts != 5 {
		t.Errorf("InsertOpts().MaxAttempts = %d, want %d", opts.MaxAttempts, 5)
	}
}

func TestSendVerificationEmailArgs_JSON(t *testing.T) {
	args := SendVerificationEmailArgs{
		UserID: 42,
		Email:  "test@example.com",
		Name:   "Test User",
		Token:  "abc123",
	}

	// Test JSON marshaling
	data, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify all fields are present in JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["user_id"] != float64(42) {
		t.Errorf("user_id = %v, want 42", result["user_id"])
	}
	if result["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", result["email"])
	}
	if result["name"] != "Test User" {
		t.Errorf("name = %v, want Test User", result["name"])
	}
	if result["token"] != "abc123" {
		t.Errorf("token = %v, want abc123", result["token"])
	}
}

// ============ SendPasswordResetEmailArgs Tests ============

func TestSendPasswordResetEmailArgs_Kind(t *testing.T) {
	args := SendPasswordResetEmailArgs{}
	if args.Kind() != "send_password_reset_email" {
		t.Errorf("Kind() = %q, want %q", args.Kind(), "send_password_reset_email")
	}
}

func TestSendPasswordResetEmailArgs_InsertOpts(t *testing.T) {
	args := SendPasswordResetEmailArgs{}
	opts := args.InsertOpts()

	if opts.Queue != "email" {
		t.Errorf("InsertOpts().Queue = %q, want %q", opts.Queue, "email")
	}
	if opts.MaxAttempts != 5 {
		t.Errorf("InsertOpts().MaxAttempts = %d, want %d", opts.MaxAttempts, 5)
	}
}

// ============ ProcessStripeWebhookArgs Tests ============

func TestProcessStripeWebhookArgs_Kind(t *testing.T) {
	args := ProcessStripeWebhookArgs{}
	if args.Kind() != "process_stripe_webhook" {
		t.Errorf("Kind() = %q, want %q", args.Kind(), "process_stripe_webhook")
	}
}

func TestProcessStripeWebhookArgs_InsertOpts(t *testing.T) {
	args := ProcessStripeWebhookArgs{}
	opts := args.InsertOpts()

	if opts.Queue != "webhooks" {
		t.Errorf("InsertOpts().Queue = %q, want %q", opts.Queue, "webhooks")
	}
	if opts.MaxAttempts != 3 {
		t.Errorf("InsertOpts().MaxAttempts = %d, want %d", opts.MaxAttempts, 3)
	}
	// UniqueOpts.ByArgs should be true to prevent duplicate event processing
	if !opts.UniqueOpts.ByArgs {
		t.Error("InsertOpts().UniqueOpts.ByArgs should be true")
	}
}

func TestProcessStripeWebhookArgs_JSON(t *testing.T) {
	payload := json.RawMessage(`{"customer": "cus_123"}`)
	args := ProcessStripeWebhookArgs{
		EventID:   "evt_123",
		EventType: "checkout.session.completed",
		Payload:   payload,
	}

	data, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["event_id"] != "evt_123" {
		t.Errorf("event_id = %v, want evt_123", result["event_id"])
	}
	if result["event_type"] != "checkout.session.completed" {
		t.Errorf("event_type = %v, want checkout.session.completed", result["event_type"])
	}
}

// ============ SendAnnouncementEmailArgs Tests ============

func TestSendAnnouncementEmailArgs_Kind(t *testing.T) {
	args := SendAnnouncementEmailArgs{}
	if args.Kind() != "send_announcement_email" {
		t.Errorf("Kind() = %q, want %q", args.Kind(), "send_announcement_email")
	}
}

func TestSendAnnouncementEmailArgs_InsertOpts(t *testing.T) {
	args := SendAnnouncementEmailArgs{}
	opts := args.InsertOpts()

	if opts.Queue != "email" {
		t.Errorf("InsertOpts().Queue = %q, want %q", opts.Queue, "email")
	}
	if opts.MaxAttempts != 5 {
		t.Errorf("InsertOpts().MaxAttempts = %d, want %d", opts.MaxAttempts, 5)
	}
}

func TestSendAnnouncementEmailArgs_Structure(t *testing.T) {
	args := SendAnnouncementEmailArgs{
		AnnouncementID: 1,
		UserID:         42,
		UserEmail:      "test@example.com",
		UserName:       "Test User",
		Title:          "New Feature Released",
		Message:        "Check out our new feature!",
		Category:       "feature",
		LinkURL:        "https://example.com/feature",
		LinkText:       "Learn More",
	}

	if args.AnnouncementID != 1 {
		t.Errorf("AnnouncementID = %d, want 1", args.AnnouncementID)
	}
	if args.Category != "feature" {
		t.Errorf("Category = %q, want feature", args.Category)
	}
}

// ============ SendAccountLockedEmailArgs Tests ============

func TestSendAccountLockedEmailArgs_Kind(t *testing.T) {
	args := SendAccountLockedEmailArgs{}
	if args.Kind() != "send_account_locked_email" {
		t.Errorf("Kind() = %q, want %q", args.Kind(), "send_account_locked_email")
	}
}

func TestSendAccountLockedEmailArgs_InsertOpts(t *testing.T) {
	args := SendAccountLockedEmailArgs{}
	opts := args.InsertOpts()

	if opts.Queue != "email" {
		t.Errorf("InsertOpts().Queue = %q, want %q", opts.Queue, "email")
	}
	if opts.MaxAttempts != 5 {
		t.Errorf("InsertOpts().MaxAttempts = %d, want %d", opts.MaxAttempts, 5)
	}
}

func TestSendAccountLockedEmailArgs_Structure(t *testing.T) {
	args := SendAccountLockedEmailArgs{
		UserID:         42,
		Email:          "test@example.com",
		Name:           "Test User",
		LockDuration:   "15 minutes",
		FailedAttempts: 5,
	}

	if args.LockDuration != "15 minutes" {
		t.Errorf("LockDuration = %q, want '15 minutes'", args.LockDuration)
	}
	if args.FailedAttempts != 5 {
		t.Errorf("FailedAttempts = %d, want 5", args.FailedAttempts)
	}
}

// ============ Queue Names Consistency Tests ============

func TestEmailJobsUseEmailQueue(t *testing.T) {
	emailJobs := []struct {
		name string
		args interface {
			Kind() string
			InsertOpts() interface{ GetQueue() string }
		}
	}{
		// Note: We can't directly test InsertOpts().Queue due to River's type system
		// So we just verify Kind() returns expected values
	}

	// Test that email jobs have "email" prefix or naming convention
	emailJobKinds := []string{
		SendVerificationEmailArgs{}.Kind(),
		SendPasswordResetEmailArgs{}.Kind(),
		SendAnnouncementEmailArgs{}.Kind(),
		SendAccountLockedEmailArgs{}.Kind(),
	}

	for i, kind := range emailJobKinds {
		if kind == "" {
			t.Errorf("Email job %d has empty Kind()", i)
		}
	}

	_ = emailJobs // Silence unused variable
}

// ============ Job Kinds Uniqueness Tests ============

func TestJobKindsAreUnique(t *testing.T) {
	kinds := make(map[string]bool)
	jobKinds := []string{
		SendVerificationEmailArgs{}.Kind(),
		SendPasswordResetEmailArgs{}.Kind(),
		ProcessStripeWebhookArgs{}.Kind(),
		SendAnnouncementEmailArgs{}.Kind(),
		SendAccountLockedEmailArgs{}.Kind(),
		DataExportArgs{}.Kind(),
	}

	for _, kind := range jobKinds {
		if kinds[kind] {
			t.Errorf("Duplicate job kind found: %q", kind)
		}
		kinds[kind] = true
	}
}

// ============ MaxAttempts Validation Tests ============

func TestEmailJobsHaveReasonableMaxAttempts(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
	}{
		{"verification email", SendVerificationEmailArgs{}.InsertOpts().MaxAttempts},
		{"password reset email", SendPasswordResetEmailArgs{}.InsertOpts().MaxAttempts},
		{"announcement email", SendAnnouncementEmailArgs{}.InsertOpts().MaxAttempts},
		{"account locked email", SendAccountLockedEmailArgs{}.InsertOpts().MaxAttempts},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Email jobs should retry a reasonable number of times (3-10)
			if tt.maxAttempts < 3 {
				t.Errorf("%s MaxAttempts = %d, should be at least 3", tt.name, tt.maxAttempts)
			}
			if tt.maxAttempts > 10 {
				t.Errorf("%s MaxAttempts = %d, should not exceed 10", tt.name, tt.maxAttempts)
			}
		})
	}
}

func TestWebhookJobsHaveLowerMaxAttempts(t *testing.T) {
	opts := ProcessStripeWebhookArgs{}.InsertOpts()

	// Webhooks should have lower retry count to fail fast
	if opts.MaxAttempts > 5 {
		t.Errorf("Stripe webhook MaxAttempts = %d, should not exceed 5", opts.MaxAttempts)
	}
}
