package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"react-golang-starter/internal/email"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

// ============================================
// Email Job Workers
// ============================================

// SendVerificationEmailArgs contains the job arguments for verification emails
type SendVerificationEmailArgs struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Token  string `json:"token"`
}

// Kind returns the job type identifier
func (SendVerificationEmailArgs) Kind() string {
	return "send_verification_email"
}

// InsertOpts returns default insert options for this job type
func (SendVerificationEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       "email",
		MaxAttempts: 5,
	}
}

// SendVerificationEmailWorker processes verification email jobs
type SendVerificationEmailWorker struct {
	river.WorkerDefaults[SendVerificationEmailArgs]
}

// Work executes the verification email job
func (w *SendVerificationEmailWorker) Work(ctx context.Context, job *river.Job[SendVerificationEmailArgs]) error {
	args := job.Args

	log.Info().
		Uint("user_id", args.UserID).
		Str("email", args.Email).
		Msg("sending verification email")

	// Build verification URL
	frontendURL := email.GetFrontendURL()
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, args.Token)

	// Send email using email service
	err := email.Send(ctx, email.SendParams{
		To:           args.Email,
		TemplateName: "verification",
		Data: map[string]interface{}{
			"Name":            args.Name,
			"VerificationURL": verificationURL,
		},
	})

	if err != nil {
		log.Error().Err(err).Str("email", args.Email).Msg("failed to send verification email")
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	log.Info().
		Uint("user_id", args.UserID).
		Str("email", args.Email).
		Msg("verification email sent successfully")

	return nil
}

// SendPasswordResetEmailArgs contains the job arguments for password reset emails
type SendPasswordResetEmailArgs struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Token  string `json:"token"`
}

// Kind returns the job type identifier
func (SendPasswordResetEmailArgs) Kind() string {
	return "send_password_reset_email"
}

// InsertOpts returns default insert options for this job type
func (SendPasswordResetEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       "email",
		MaxAttempts: 5,
	}
}

// SendPasswordResetEmailWorker processes password reset email jobs
type SendPasswordResetEmailWorker struct {
	river.WorkerDefaults[SendPasswordResetEmailArgs]
}

// Work executes the password reset email job
func (w *SendPasswordResetEmailWorker) Work(ctx context.Context, job *river.Job[SendPasswordResetEmailArgs]) error {
	args := job.Args

	log.Info().
		Uint("user_id", args.UserID).
		Str("email", args.Email).
		Msg("sending password reset email")

	// Build reset URL
	frontendURL := email.GetFrontendURL()
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, args.Token)

	// Send email using email service
	err := email.Send(ctx, email.SendParams{
		To:           args.Email,
		TemplateName: "password_reset",
		Data: map[string]interface{}{
			"Name":     args.Name,
			"ResetURL": resetURL,
		},
	})

	if err != nil {
		log.Error().Err(err).Str("email", args.Email).Msg("failed to send password reset email")
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	log.Info().
		Uint("user_id", args.UserID).
		Str("email", args.Email).
		Msg("password reset email sent successfully")

	return nil
}

// ============================================
// Stripe Webhook Worker
// ============================================

// ProcessStripeWebhookArgs contains Stripe webhook data
type ProcessStripeWebhookArgs struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

// Kind returns the job type identifier
func (ProcessStripeWebhookArgs) Kind() string {
	return "process_stripe_webhook"
}

// InsertOpts returns default insert options for this job type
func (ProcessStripeWebhookArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       "webhooks",
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true, // Prevent duplicate processing of same event
		},
	}
}

// ProcessStripeWebhookWorker processes Stripe webhook events
type ProcessStripeWebhookWorker struct {
	river.WorkerDefaults[ProcessStripeWebhookArgs]
}

// Work executes the Stripe webhook processing job
func (w *ProcessStripeWebhookWorker) Work(ctx context.Context, job *river.Job[ProcessStripeWebhookArgs]) error {
	args := job.Args

	log.Info().
		Str("event_id", args.EventID).
		Str("event_type", args.EventType).
		Msg("processing Stripe webhook")

	switch args.EventType {
	case "checkout.session.completed":
		return w.handleCheckoutCompleted(ctx, args.Payload)
	case "customer.subscription.created":
		return w.handleSubscriptionCreated(ctx, args.Payload)
	case "customer.subscription.updated":
		return w.handleSubscriptionUpdated(ctx, args.Payload)
	case "customer.subscription.deleted":
		return w.handleSubscriptionDeleted(ctx, args.Payload)
	case "invoice.payment_succeeded":
		return w.handlePaymentSucceeded(ctx, args.Payload)
	case "invoice.payment_failed":
		return w.handlePaymentFailed(ctx, args.Payload)
	default:
		log.Warn().Str("event_type", args.EventType).Msg("unhandled Stripe event type")
		return nil // Don't retry for unknown events
	}
}

func (w *ProcessStripeWebhookWorker) handleCheckoutCompleted(ctx context.Context, payload json.RawMessage) error {
	// Stripe service will implement the actual logic
	log.Info().Msg("handling checkout.session.completed")
	return nil
}

func (w *ProcessStripeWebhookWorker) handleSubscriptionCreated(ctx context.Context, payload json.RawMessage) error {
	log.Info().Msg("handling customer.subscription.created")
	return nil
}

func (w *ProcessStripeWebhookWorker) handleSubscriptionUpdated(ctx context.Context, payload json.RawMessage) error {
	log.Info().Msg("handling customer.subscription.updated")
	return nil
}

func (w *ProcessStripeWebhookWorker) handleSubscriptionDeleted(ctx context.Context, payload json.RawMessage) error {
	log.Info().Msg("handling customer.subscription.deleted")
	return nil
}

func (w *ProcessStripeWebhookWorker) handlePaymentSucceeded(ctx context.Context, payload json.RawMessage) error {
	log.Info().Msg("handling invoice.payment_succeeded")
	return nil
}

func (w *ProcessStripeWebhookWorker) handlePaymentFailed(ctx context.Context, payload json.RawMessage) error {
	log.Info().Msg("handling invoice.payment_failed")
	return nil
}

// ============================================
// Helper Functions for Job Insertion
// ============================================

// EnqueueVerificationEmail queues a verification email job
func EnqueueVerificationEmail(ctx context.Context, userID uint, email, name, token string) error {
	if !IsAvailable() {
		return fmt.Errorf("job system not available")
	}

	return Insert(ctx, SendVerificationEmailArgs{
		UserID: userID,
		Email:  email,
		Name:   name,
		Token:  token,
	}, nil)
}

// EnqueuePasswordResetEmail queues a password reset email job
func EnqueuePasswordResetEmail(ctx context.Context, userID uint, email, name, token string) error {
	if !IsAvailable() {
		return fmt.Errorf("job system not available")
	}

	return Insert(ctx, SendPasswordResetEmailArgs{
		UserID: userID,
		Email:  email,
		Name:   name,
		Token:  token,
	}, nil)
}

// EnqueueStripeWebhook queues a Stripe webhook for processing
func EnqueueStripeWebhook(ctx context.Context, eventID, eventType string, payload json.RawMessage) error {
	if !IsAvailable() {
		return fmt.Errorf("job system not available")
	}

	return Insert(ctx, ProcessStripeWebhookArgs{
		EventID:   eventID,
		EventType: eventType,
		Payload:   payload,
	}, nil)
}
