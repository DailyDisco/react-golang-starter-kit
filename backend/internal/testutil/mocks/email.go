package mocks

import (
	"context"
	"sync"
	"time"
)

// CapturedEmail represents a captured email for testing.
type CapturedEmail struct {
	To           string
	Subject      string
	TemplateName string
	Data         map[string]interface{}
	PlainText    string
	HTMLBody     string
	SentAt       time.Time
}

// MockEmailProvider captures emails for testing instead of sending them.
type MockEmailProvider struct {
	mu       sync.RWMutex
	emails   []CapturedEmail
	sendErr  error
	disabled bool

	// Track calls for assertions
	SendCalls      int
	SendBatchCalls int
}

// NewMockEmailProvider creates a new mock email provider.
func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		emails: make([]CapturedEmail, 0),
	}
}

// SendParams matches the parameters for sending an email.
type SendParams struct {
	To           string
	Subject      string
	TemplateName string
	Data         map[string]interface{}
	PlainText    string
	HTMLBody     string
}

// Send captures an email instead of sending it.
func (m *MockEmailProvider) Send(ctx context.Context, params SendParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SendCalls++

	if m.sendErr != nil {
		return m.sendErr
	}

	m.emails = append(m.emails, CapturedEmail{
		To:           params.To,
		Subject:      params.Subject,
		TemplateName: params.TemplateName,
		Data:         params.Data,
		PlainText:    params.PlainText,
		HTMLBody:     params.HTMLBody,
		SentAt:       time.Now(),
	})

	return nil
}

// SendBatch captures multiple emails.
func (m *MockEmailProvider) SendBatch(ctx context.Context, params []SendParams) error {
	m.mu.Lock()
	m.SendBatchCalls++
	m.mu.Unlock()

	for _, p := range params {
		if err := m.Send(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// IsAvailable returns true if the mock is not disabled.
func (m *MockEmailProvider) IsAvailable() bool {
	return !m.disabled
}

// Close is a no-op for the mock.
func (m *MockEmailProvider) Close() error {
	return nil
}

// GetEmails returns all captured emails.
func (m *MockEmailProvider) GetEmails() []CapturedEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]CapturedEmail, len(m.emails))
	copy(result, m.emails)
	return result
}

// GetLastEmail returns the most recent email.
func (m *MockEmailProvider) GetLastEmail() *CapturedEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.emails) == 0 {
		return nil
	}
	email := m.emails[len(m.emails)-1]
	return &email
}

// GetEmailsTo returns emails sent to a specific address.
func (m *MockEmailProvider) GetEmailsTo(to string) []CapturedEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []CapturedEmail
	for _, e := range m.emails {
		if e.To == to {
			result = append(result, e)
		}
	}
	return result
}

// GetEmailsByTemplate returns emails using a specific template.
func (m *MockEmailProvider) GetEmailsByTemplate(template string) []CapturedEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []CapturedEmail
	for _, e := range m.emails {
		if e.TemplateName == template {
			result = append(result, e)
		}
	}
	return result
}

// GetEmailsBySubject returns emails with a specific subject.
func (m *MockEmailProvider) GetEmailsBySubject(subject string) []CapturedEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []CapturedEmail
	for _, e := range m.emails {
		if e.Subject == subject {
			result = append(result, e)
		}
	}
	return result
}

// EmailCount returns the number of captured emails.
func (m *MockEmailProvider) EmailCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.emails)
}

// Clear removes all captured emails.
func (m *MockEmailProvider) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emails = make([]CapturedEmail, 0)
}

// SetSendError sets an error to return on send.
func (m *MockEmailProvider) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErr = err
}

// ClearSendError clears the send error.
func (m *MockEmailProvider) ClearSendError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErr = nil
}

// SetDisabled sets the availability.
func (m *MockEmailProvider) SetDisabled(disabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disabled = disabled
}

// Reset clears all captured emails and resets state.
func (m *MockEmailProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.emails = make([]CapturedEmail, 0)
	m.sendErr = nil
	m.disabled = false
	m.SendCalls = 0
	m.SendBatchCalls = 0
}

// AssertEmailSent checks if an email was sent to the given address.
// Returns true if at least one email was sent to that address.
func (m *MockEmailProvider) AssertEmailSent(to string) bool {
	emails := m.GetEmailsTo(to)
	return len(emails) > 0
}

// AssertEmailSentWithSubject checks if an email with the given subject was sent.
func (m *MockEmailProvider) AssertEmailSentWithSubject(subject string) bool {
	emails := m.GetEmailsBySubject(subject)
	return len(emails) > 0
}

// AssertNoEmailsSent checks that no emails were sent.
func (m *MockEmailProvider) AssertNoEmailsSent() bool {
	return m.EmailCount() == 0
}

// AssertEmailCount checks the number of emails sent.
func (m *MockEmailProvider) AssertEmailCount(expected int) bool {
	return m.EmailCount() == expected
}
