package models

import (
	"testing"
	"time"
)

// ============ User Tests ============

func TestUser_ToUserResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:            1,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		Role:          RoleUser,
		OAuthProvider: "google",
		AvatarURL:     "https://example.com/avatar.jpg",
		Bio:           "Test bio",
		Location:      "New York",
		SocialLinks:   `{"twitter": "@john"}`,
	}

	response := user.ToUserResponse()

	if response.ID != user.ID {
		t.Errorf("ID = %d, want %d", response.ID, user.ID)
	}
	if response.Name != user.Name {
		t.Errorf("Name = %q, want %q", response.Name, user.Name)
	}
	if response.Email != user.Email {
		t.Errorf("Email = %q, want %q", response.Email, user.Email)
	}
	if response.EmailVerified != user.EmailVerified {
		t.Errorf("EmailVerified = %v, want %v", response.EmailVerified, user.EmailVerified)
	}
	if response.IsActive != user.IsActive {
		t.Errorf("IsActive = %v, want %v", response.IsActive, user.IsActive)
	}
	if response.Role != user.Role {
		t.Errorf("Role = %q, want %q", response.Role, user.Role)
	}
	if response.OAuthProvider != user.OAuthProvider {
		t.Errorf("OAuthProvider = %q, want %q", response.OAuthProvider, user.OAuthProvider)
	}
	if response.AvatarURL != user.AvatarURL {
		t.Errorf("AvatarURL = %q, want %q", response.AvatarURL, user.AvatarURL)
	}
	if response.Bio != user.Bio {
		t.Errorf("Bio = %q, want %q", response.Bio, user.Bio)
	}
	if response.Location != user.Location {
		t.Errorf("Location = %q, want %q", response.Location, user.Location)
	}
	if response.SocialLinks != user.SocialLinks {
		t.Errorf("SocialLinks = %q, want %q", response.SocialLinks, user.SocialLinks)
	}
}

func TestUser_ToUserResponse_TimeFormatting(t *testing.T) {
	now := time.Date(2023, 8, 27, 12, 0, 0, 0, time.UTC)
	user := &User{
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := user.ToUserResponse()

	expectedFormat := now.Format(time.RFC3339)
	if response.CreatedAt != expectedFormat {
		t.Errorf("CreatedAt = %q, want %q", response.CreatedAt, expectedFormat)
	}
	if response.UpdatedAt != expectedFormat {
		t.Errorf("UpdatedAt = %q, want %q", response.UpdatedAt, expectedFormat)
	}
}

func TestUser_ToUserResponse_EmptyOptionalFields(t *testing.T) {
	user := &User{
		ID:        1,
		Name:      "John",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	response := user.ToUserResponse()

	if response.OAuthProvider != "" {
		t.Errorf("OAuthProvider should be empty, got %q", response.OAuthProvider)
	}
	if response.AvatarURL != "" {
		t.Errorf("AvatarURL should be empty, got %q", response.AvatarURL)
	}
	if response.Bio != "" {
		t.Errorf("Bio should be empty, got %q", response.Bio)
	}
	if response.Location != "" {
		t.Errorf("Location should be empty, got %q", response.Location)
	}
}

// ============ TokenBlacklist Tests ============

func TestTokenBlacklist_TableName(t *testing.T) {
	blacklist := TokenBlacklist{}
	if got := blacklist.TableName(); got != "token_blacklist" {
		t.Errorf("TokenBlacklist.TableName() = %q, want %q", got, "token_blacklist")
	}
}

// ============ File Tests ============

func TestFile_ToFileResponse(t *testing.T) {
	file := &File{
		ID:          1,
		UserID:      42,
		FileName:    "document.pdf",
		ContentType: "application/pdf",
		FileSize:    1024,
		Location:    "https://s3.example.com/file.pdf",
		StorageType: "s3",
		CreatedAt:   "2023-08-27T12:00:00Z",
		UpdatedAt:   "2023-08-27T12:00:00Z",
	}

	response := file.ToFileResponse()

	if response.ID != file.ID {
		t.Errorf("ID = %d, want %d", response.ID, file.ID)
	}
	if response.UserID != file.UserID {
		t.Errorf("UserID = %d, want %d", response.UserID, file.UserID)
	}
	if response.FileName != file.FileName {
		t.Errorf("FileName = %q, want %q", response.FileName, file.FileName)
	}
	if response.ContentType != file.ContentType {
		t.Errorf("ContentType = %q, want %q", response.ContentType, file.ContentType)
	}
	if response.FileSize != file.FileSize {
		t.Errorf("FileSize = %d, want %d", response.FileSize, file.FileSize)
	}
	if response.Location != file.Location {
		t.Errorf("Location = %q, want %q", response.Location, file.Location)
	}
	if response.StorageType != file.StorageType {
		t.Errorf("StorageType = %q, want %q", response.StorageType, file.StorageType)
	}
	if response.CreatedAt != file.CreatedAt {
		t.Errorf("CreatedAt = %q, want %q", response.CreatedAt, file.CreatedAt)
	}
	if response.UpdatedAt != file.UpdatedAt {
		t.Errorf("UpdatedAt = %q, want %q", response.UpdatedAt, file.UpdatedAt)
	}
}

// ============ Role Constants Tests ============

func TestRoleConstants(t *testing.T) {
	tests := []struct {
		role string
		want string
	}{
		{RoleSuperAdmin, "super_admin"},
		{RoleAdmin, "admin"},
		{RolePremium, "premium"},
		{RoleUser, "user"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if tt.role != tt.want {
				t.Errorf("Role = %q, want %q", tt.role, tt.want)
			}
		})
	}
}

func TestRoleHierarchy(t *testing.T) {
	// Verify hierarchy levels are defined
	tests := []struct {
		role string
		want int
	}{
		{RoleSuperAdmin, 100},
		{RoleAdmin, 50},
		{RolePremium, 20},
		{RoleUser, 10},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if got := RoleHierarchy[tt.role]; got != tt.want {
				t.Errorf("RoleHierarchy[%q] = %d, want %d", tt.role, got, tt.want)
			}
		})
	}
}

func TestRoleHierarchy_Order(t *testing.T) {
	// Verify super_admin > admin > premium > user
	if RoleHierarchy[RoleSuperAdmin] <= RoleHierarchy[RoleAdmin] {
		t.Error("SuperAdmin should be higher than Admin")
	}
	if RoleHierarchy[RoleAdmin] <= RoleHierarchy[RolePremium] {
		t.Error("Admin should be higher than Premium")
	}
	if RoleHierarchy[RolePremium] <= RoleHierarchy[RoleUser] {
		t.Error("Premium should be higher than User")
	}
}

func TestRoleHierarchy_AllRolesDefined(t *testing.T) {
	roles := []string{RoleSuperAdmin, RoleAdmin, RolePremium, RoleUser}

	for _, role := range roles {
		if _, exists := RoleHierarchy[role]; !exists {
			t.Errorf("Role %q is not defined in RoleHierarchy", role)
		}
	}
}

// ============ Subscription Status Constants Tests ============

func TestSubscriptionStatusConstants(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{SubscriptionStatusActive, "active"},
		{SubscriptionStatusPastDue, "past_due"},
		{SubscriptionStatusCanceled, "canceled"},
		{SubscriptionStatusTrialing, "trialing"},
		{SubscriptionStatusUnpaid, "unpaid"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if tt.status != tt.want {
				t.Errorf("Status = %q, want %q", tt.status, tt.want)
			}
		})
	}
}

func TestSubscriptionStatusConstants_Unique(t *testing.T) {
	statuses := []string{
		SubscriptionStatusActive,
		SubscriptionStatusPastDue,
		SubscriptionStatusCanceled,
		SubscriptionStatusTrialing,
		SubscriptionStatusUnpaid,
	}

	seen := make(map[string]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("Duplicate subscription status: %q", status)
		}
		seen[status] = true
	}

	if len(seen) != 5 {
		t.Errorf("Expected 5 unique statuses, got %d", len(seen))
	}
}

// ============ Struct Field Tests ============

func TestUserResponse_Fields(t *testing.T) {
	response := UserResponse{
		ID:            1,
		Name:          "John",
		Email:         "john@example.com",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     "2023-08-27T12:00:00Z",
		UpdatedAt:     "2023-08-27T12:00:00Z",
		Role:          RoleAdmin,
	}

	if response.ID != 1 {
		t.Errorf("ID = %d, want 1", response.ID)
	}
	if response.Name != "John" {
		t.Errorf("Name = %q, want 'John'", response.Name)
	}
	if response.Email != "john@example.com" {
		t.Errorf("Email = %q, want 'john@example.com'", response.Email)
	}
	if !response.EmailVerified {
		t.Error("EmailVerified should be true")
	}
	if !response.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestAuthResponse_Fields(t *testing.T) {
	response := AuthResponse{
		User:         UserResponse{ID: 1},
		Token:        "jwt-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}

	if response.User.ID != 1 {
		t.Errorf("User.ID = %d, want 1", response.User.ID)
	}
	if response.Token != "jwt-token" {
		t.Errorf("Token = %q, want 'jwt-token'", response.Token)
	}
	if response.RefreshToken != "refresh-token" {
		t.Errorf("RefreshToken = %q, want 'refresh-token'", response.RefreshToken)
	}
	if response.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", response.ExpiresIn)
	}
}

func TestErrorResponse_Fields(t *testing.T) {
	response := ErrorResponse{
		Error:     "Bad Request",
		Message:   "Invalid email format",
		Code:      400,
		RequestID: "req-123",
		Details: []FieldError{
			{Field: "email", Message: "Invalid format"},
		},
	}

	if response.Error != "Bad Request" {
		t.Errorf("Error = %q, want 'Bad Request'", response.Error)
	}
	if response.Message != "Invalid email format" {
		t.Errorf("Message = %q, want 'Invalid email format'", response.Message)
	}
	if response.Code != 400 {
		t.Errorf("Code = %d, want 400", response.Code)
	}
	if response.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want 'req-123'", response.RequestID)
	}
	if len(response.Details) != 1 {
		t.Errorf("Details length = %d, want 1", len(response.Details))
	}
}

func TestSuccessResponse_Fields(t *testing.T) {
	response := SuccessResponse{
		Success: true,
		Message: "Operation completed",
		Data:    map[string]string{"key": "value"},
	}

	if !response.Success {
		t.Error("Success should be true")
	}
	if response.Message != "Operation completed" {
		t.Errorf("Message = %q, want 'Operation completed'", response.Message)
	}
	if response.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestFieldError_Fields(t *testing.T) {
	fieldErr := FieldError{
		Field:   "email",
		Message: "Invalid email format",
		Code:    "invalid_email",
		Value:   "not-an-email",
	}

	if fieldErr.Field != "email" {
		t.Errorf("Field = %q, want 'email'", fieldErr.Field)
	}
	if fieldErr.Message != "Invalid email format" {
		t.Errorf("Message = %q, want 'Invalid email format'", fieldErr.Message)
	}
	if fieldErr.Code != "invalid_email" {
		t.Errorf("Code = %q, want 'invalid_email'", fieldErr.Code)
	}
}

func TestHealthStatus_Fields(t *testing.T) {
	status := HealthStatus{
		OverallStatus: "healthy",
		Timestamp:     "2023-08-27T12:00:00Z",
		Uptime:        "24h",
		Version: VersionInfo{
			Version:   "1.0.0",
			BuildTime: "2023-08-27",
			GitCommit: "abc123",
		},
		Components: []ComponentStatus{
			{Name: "database", Status: "healthy"},
		},
	}

	if status.OverallStatus != "healthy" {
		t.Errorf("OverallStatus = %q, want 'healthy'", status.OverallStatus)
	}
	if status.Uptime != "24h" {
		t.Errorf("Uptime = %q, want '24h'", status.Uptime)
	}
	if status.Version.Version != "1.0.0" {
		t.Errorf("Version.Version = %q, want '1.0.0'", status.Version.Version)
	}
	if len(status.Components) != 1 {
		t.Errorf("Components length = %d, want 1", len(status.Components))
	}
}

func TestPaginationQuery_Fields(t *testing.T) {
	query := PaginationQuery{
		Page:  2,
		Limit: 25,
	}

	if query.Page != 2 {
		t.Errorf("Page = %d, want 2", query.Page)
	}
	if query.Limit != 25 {
		t.Errorf("Limit = %d, want 25", query.Limit)
	}
}

// ============ Request/Response Type Tests ============

func TestLoginRequest_Fields(t *testing.T) {
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	if req.Email != "test@example.com" {
		t.Errorf("Email = %q, want 'test@example.com'", req.Email)
	}
	if req.Password != "password123" {
		t.Errorf("Password = %q, want 'password123'", req.Password)
	}
}

func TestRegisterRequest_Fields(t *testing.T) {
	req := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	if req.Name != "John Doe" {
		t.Errorf("Name = %q, want 'John Doe'", req.Name)
	}
	if req.Email != "john@example.com" {
		t.Errorf("Email = %q, want 'john@example.com'", req.Email)
	}
	if req.Password != "SecurePass123!" {
		t.Errorf("Password = %q, want 'SecurePass123!'", req.Password)
	}
}

// ============ Subscription Tests ============

func TestSubscription_ToSubscriptionResponse(t *testing.T) {
	orgID := uint(10)
	sub := &Subscription{
		ID:                   1,
		UserID:               42,
		OrganizationID:       &orgID,
		StripeSubscriptionID: "sub_123",
		StripePriceID:        "price_123",
		Status:               SubscriptionStatusActive,
		CurrentPeriodStart:   "2023-08-01T00:00:00Z",
		CurrentPeriodEnd:     "2023-09-01T00:00:00Z",
		CancelAtPeriodEnd:    false,
		CanceledAt:           "",
		CreatedAt:            "2023-08-01T00:00:00Z",
		UpdatedAt:            "2023-08-15T00:00:00Z",
	}

	response := sub.ToSubscriptionResponse()

	if response.ID != sub.ID {
		t.Errorf("ID = %d, want %d", response.ID, sub.ID)
	}
	if response.UserID != sub.UserID {
		t.Errorf("UserID = %d, want %d", response.UserID, sub.UserID)
	}
	if response.OrganizationID != sub.OrganizationID {
		t.Errorf("OrganizationID = %v, want %v", response.OrganizationID, sub.OrganizationID)
	}
	if response.Status != sub.Status {
		t.Errorf("Status = %q, want %q", response.Status, sub.Status)
	}
	if response.StripePriceID != sub.StripePriceID {
		t.Errorf("StripePriceID = %q, want %q", response.StripePriceID, sub.StripePriceID)
	}
	if response.CurrentPeriodStart != sub.CurrentPeriodStart {
		t.Errorf("CurrentPeriodStart = %q, want %q", response.CurrentPeriodStart, sub.CurrentPeriodStart)
	}
	if response.CurrentPeriodEnd != sub.CurrentPeriodEnd {
		t.Errorf("CurrentPeriodEnd = %q, want %q", response.CurrentPeriodEnd, sub.CurrentPeriodEnd)
	}
	if response.CancelAtPeriodEnd != sub.CancelAtPeriodEnd {
		t.Errorf("CancelAtPeriodEnd = %v, want %v", response.CancelAtPeriodEnd, sub.CancelAtPeriodEnd)
	}
}

func TestSubscription_IsActiveSubscription(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"active status", SubscriptionStatusActive, true},
		{"trialing status", SubscriptionStatusTrialing, true},
		{"past_due status", SubscriptionStatusPastDue, false},
		{"canceled status", SubscriptionStatusCanceled, false},
		{"unpaid status", SubscriptionStatusUnpaid, false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{Status: tt.status}
			if got := sub.IsActiveSubscription(); got != tt.want {
				t.Errorf("IsActiveSubscription() with status %q = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// ============ AuditLog Tests ============

func TestAuditLog_ToAuditLogResponse(t *testing.T) {
	userID := uint(42)
	targetID := uint(10)
	log := &AuditLog{
		ID:         1,
		UserID:     &userID,
		TargetType: AuditTargetUser,
		TargetID:   &targetID,
		Action:     "update",
		Changes:    `{"name": "New Name"}`,
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		CreatedAt:  "2023-08-27T12:00:00Z",
	}

	response := log.ToAuditLogResponse()

	if response.ID != log.ID {
		t.Errorf("ID = %d, want %d", response.ID, log.ID)
	}
	if response.UserID != log.UserID {
		t.Errorf("UserID = %v, want %v", response.UserID, log.UserID)
	}
	if response.TargetType != log.TargetType {
		t.Errorf("TargetType = %q, want %q", response.TargetType, log.TargetType)
	}
	if response.TargetID != log.TargetID {
		t.Errorf("TargetID = %v, want %v", response.TargetID, log.TargetID)
	}
	if response.Action != log.Action {
		t.Errorf("Action = %q, want %q", response.Action, log.Action)
	}
	// Note: ToAuditLogResponse() doesn't copy Changes field - verified by reading the method
	if response.IPAddress != log.IPAddress {
		t.Errorf("IPAddress = %q, want %q", response.IPAddress, log.IPAddress)
	}
	if response.UserAgent != log.UserAgent {
		t.Errorf("UserAgent = %q, want %q", response.UserAgent, log.UserAgent)
	}
	if response.CreatedAt != log.CreatedAt {
		t.Errorf("CreatedAt = %q, want %q", response.CreatedAt, log.CreatedAt)
	}
}

func TestAuditLog_ToAuditLogResponse_WithUser(t *testing.T) {
	userID := uint(42)
	log := &AuditLog{
		ID:     1,
		UserID: &userID,
		User: &User{
			ID:    42,
			Name:  "Test User",
			Email: "test@example.com",
		},
		TargetType: AuditTargetUser,
		Action:     "create",
		CreatedAt:  "2023-08-27T12:00:00Z",
	}

	response := log.ToAuditLogResponse()

	if response.UserName != log.User.Name {
		t.Errorf("UserName = %q, want %q", response.UserName, log.User.Name)
	}
	if response.UserEmail != log.User.Email {
		t.Errorf("UserEmail = %q, want %q", response.UserEmail, log.User.Email)
	}
}

// ============ AuditTargetType Constants Tests ============

func TestAuditTargetTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"user target", AuditTargetUser, "user"},
		{"subscription target", AuditTargetSubscription, "subscription"},
		{"file target", AuditTargetFile, "file"},
		{"settings target", AuditTargetSettings, "settings"},
		{"feature_flag target", AuditTargetFeatureFlag, "feature_flag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("AuditTarget constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ============ UserAPIKey Tests ============

func TestUserAPIKey_ToUserAPIKeyResponse(t *testing.T) {
	lastUsed := "2023-08-27T12:00:00Z"
	key := &UserAPIKey{
		ID:           1,
		UserID:       42,
		Provider:     "openai",
		Name:         "Test API Key",
		KeyHash:      "hash123abc",
		KeyEncrypted: "encrypted_key_data",
		KeyPreview:   "...xyz1234",
		IsActive:     true,
		LastUsedAt:   &lastUsed,
		UsageCount:   10,
		CreatedAt:    "2023-08-01T00:00:00Z",
		UpdatedAt:    "2023-08-15T00:00:00Z",
	}

	response := key.ToUserAPIKeyResponse()

	if response.ID != key.ID {
		t.Errorf("ID = %d, want %d", response.ID, key.ID)
	}
	if response.Provider != key.Provider {
		t.Errorf("Provider = %q, want %q", response.Provider, key.Provider)
	}
	if response.Name != key.Name {
		t.Errorf("Name = %q, want %q", response.Name, key.Name)
	}
	if response.KeyPreview != key.KeyPreview {
		t.Errorf("KeyPreview = %q, want %q", response.KeyPreview, key.KeyPreview)
	}
	if response.IsActive != key.IsActive {
		t.Errorf("IsActive = %v, want %v", response.IsActive, key.IsActive)
	}
	if response.LastUsedAt == nil || *response.LastUsedAt != *key.LastUsedAt {
		t.Errorf("LastUsedAt = %v, want %v", response.LastUsedAt, key.LastUsedAt)
	}
	if response.UsageCount != key.UsageCount {
		t.Errorf("UsageCount = %d, want %d", response.UsageCount, key.UsageCount)
	}
}

func TestUserAPIKey_TableName(t *testing.T) {
	key := &UserAPIKey{}
	if got := key.TableName(); got != "user_api_keys" {
		t.Errorf("UserAPIKey.TableName() = %q, want 'user_api_keys'", got)
	}
}

// ============ UserAPIKey TableName Test ============

// ============ BillingPlan Tests ============

func TestBillingPlan_Fields(t *testing.T) {
	plan := BillingPlan{
		ID:          "premium",
		Name:        "Premium Plan",
		Description: "Best value for professionals",
		PriceID:     "price_premium_123",
		Amount:      1999, // $19.99
		Currency:    "usd",
		Interval:    "month",
		Features:    []string{"Unlimited projects", "Priority support", "Advanced analytics"},
	}

	if plan.ID != "premium" {
		t.Errorf("ID = %q, want 'premium'", plan.ID)
	}
	if plan.Amount != 1999 {
		t.Errorf("Amount = %d, want 1999", plan.Amount)
	}
	if plan.Currency != "usd" {
		t.Errorf("Currency = %q, want 'usd'", plan.Currency)
	}
	if len(plan.Features) != 3 {
		t.Errorf("Features length = %d, want 3", len(plan.Features))
	}
}

// ============ CheckoutRequest Tests ============

func TestCreateCheckoutRequest_Fields(t *testing.T) {
	req := CreateCheckoutRequest{
		PriceID: "price_123",
	}

	if req.PriceID != "price_123" {
		t.Errorf("PriceID = %q, want 'price_123'", req.PriceID)
	}
}

// ============ Session Response Tests ============

func TestCheckoutSessionResponse_Fields(t *testing.T) {
	resp := CheckoutSessionResponse{
		SessionID: "cs_test_123",
		URL:       "https://checkout.stripe.com/pay/cs_test_123",
	}

	if resp.SessionID != "cs_test_123" {
		t.Errorf("SessionID = %q, want 'cs_test_123'", resp.SessionID)
	}
	if resp.URL != "https://checkout.stripe.com/pay/cs_test_123" {
		t.Errorf("URL = %q, want 'https://checkout.stripe.com/pay/cs_test_123'", resp.URL)
	}
}

func TestPortalSessionResponse_Fields(t *testing.T) {
	resp := PortalSessionResponse{
		URL: "https://billing.stripe.com/session/123",
	}

	if resp.URL != "https://billing.stripe.com/session/123" {
		t.Errorf("URL = %q, want 'https://billing.stripe.com/session/123'", resp.URL)
	}
}

// ============ Settings Models ToResponse Tests ============

func TestSystemSetting_ToResponse_NonSensitive(t *testing.T) {
	setting := &SystemSetting{
		ID:          1,
		Key:         "site_name",
		Value:       []byte(`"My Site"`),
		Category:    "site",
		Description: "The site name",
		IsSensitive: false,
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}

	resp := setting.ToResponse()

	if resp.ID != setting.ID {
		t.Errorf("ID = %d, want %d", resp.ID, setting.ID)
	}
	if resp.Key != setting.Key {
		t.Errorf("Key = %q, want %q", resp.Key, setting.Key)
	}
	if resp.Value != "My Site" {
		t.Errorf("Value = %v, want 'My Site'", resp.Value)
	}
	if resp.Category != setting.Category {
		t.Errorf("Category = %q, want %q", resp.Category, setting.Category)
	}
	if resp.IsSensitive != setting.IsSensitive {
		t.Errorf("IsSensitive = %v, want %v", resp.IsSensitive, setting.IsSensitive)
	}
}

func TestSystemSetting_ToResponse_Sensitive(t *testing.T) {
	setting := &SystemSetting{
		ID:          2,
		Key:         "smtp_password",
		Value:       []byte(`"secret123"`),
		Category:    "email",
		IsSensitive: true,
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}

	resp := setting.ToResponse()

	// Sensitive values should be masked
	if resp.Value != "********" {
		t.Errorf("Value should be masked, got %v", resp.Value)
	}
}

func TestUserPreferences_ToResponse(t *testing.T) {
	prefs := &UserPreferences{
		ID:                 1,
		UserID:             42,
		Theme:              "dark",
		Timezone:           "America/New_York",
		Language:           "en",
		DateFormat:         "YYYY-MM-DD",
		TimeFormat:         "24h",
		EmailNotifications: []byte(`{"marketing":true,"security":true,"updates":false,"weekly_digest":true}`),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	resp := prefs.ToResponse()

	if resp.Theme != prefs.Theme {
		t.Errorf("Theme = %q, want %q", resp.Theme, prefs.Theme)
	}
	if resp.Timezone != prefs.Timezone {
		t.Errorf("Timezone = %q, want %q", resp.Timezone, prefs.Timezone)
	}
	if resp.Language != prefs.Language {
		t.Errorf("Language = %q, want %q", resp.Language, prefs.Language)
	}
	if resp.DateFormat != prefs.DateFormat {
		t.Errorf("DateFormat = %q, want %q", resp.DateFormat, prefs.DateFormat)
	}
	if resp.TimeFormat != prefs.TimeFormat {
		t.Errorf("TimeFormat = %q, want %q", resp.TimeFormat, prefs.TimeFormat)
	}
	if !resp.EmailNotifications.Marketing {
		t.Error("EmailNotifications.Marketing should be true")
	}
	if !resp.EmailNotifications.Security {
		t.Error("EmailNotifications.Security should be true")
	}
	if resp.EmailNotifications.Updates {
		t.Error("EmailNotifications.Updates should be false")
	}
	if !resp.EmailNotifications.WeeklyDigest {
		t.Error("EmailNotifications.WeeklyDigest should be true")
	}
}

func TestUserPreferences_ToResponse_NilEmailNotifications(t *testing.T) {
	prefs := &UserPreferences{
		Theme:              "light",
		EmailNotifications: nil,
	}

	resp := prefs.ToResponse()

	// Should have zero values for email notifications
	if resp.EmailNotifications.Marketing {
		t.Error("EmailNotifications.Marketing should be false when nil")
	}
}

func TestUserSession_ToResponse(t *testing.T) {
	now := time.Now()
	session := &UserSession{
		ID:           1,
		UserID:       42,
		DeviceInfo:   []byte(`{"browser":"Chrome","browser_version":"120","os":"macOS","device_type":"desktop"}`),
		IPAddress:    "192.168.1.100",
		Location:     []byte(`{"country":"United States","country_code":"US","city":"New York"}`),
		IsCurrent:    true,
		LastActiveAt: now,
		CreatedAt:    now,
	}

	resp := session.ToResponse()

	if resp.ID != session.ID {
		t.Errorf("ID = %d, want %d", resp.ID, session.ID)
	}
	if resp.IPAddress != session.IPAddress {
		t.Errorf("IPAddress = %q, want %q", resp.IPAddress, session.IPAddress)
	}
	if !resp.IsCurrent {
		t.Error("IsCurrent should be true")
	}
	if resp.DeviceInfo.Browser != "Chrome" {
		t.Errorf("DeviceInfo.Browser = %q, want 'Chrome'", resp.DeviceInfo.Browser)
	}
	if resp.DeviceInfo.OS != "macOS" {
		t.Errorf("DeviceInfo.OS = %q, want 'macOS'", resp.DeviceInfo.OS)
	}
	if resp.Location.Country != "United States" {
		t.Errorf("Location.Country = %q, want 'United States'", resp.Location.Country)
	}
	if resp.Location.City != "New York" {
		t.Errorf("Location.City = %q, want 'New York'", resp.Location.City)
	}
}

func TestUserSession_ToResponse_NilFields(t *testing.T) {
	now := time.Now()
	session := &UserSession{
		ID:           1,
		DeviceInfo:   nil,
		Location:     nil,
		LastActiveAt: now,
		CreatedAt:    now,
	}

	resp := session.ToResponse()

	// Should have empty DeviceInfo and Location
	if resp.DeviceInfo.Browser != "" {
		t.Errorf("DeviceInfo.Browser should be empty, got %q", resp.DeviceInfo.Browser)
	}
	if resp.Location.Country != "" {
		t.Errorf("Location.Country should be empty, got %q", resp.Location.Country)
	}
}

func TestAnnouncementBanner_ToResponse(t *testing.T) {
	startsAt := "2025-01-01T00:00:00Z"
	endsAt := "2025-12-31T23:59:59Z"
	publishedAt := "2025-01-01T00:00:00Z"

	banner := &AnnouncementBanner{
		ID:            1,
		Title:         "System Update",
		Message:       "We have a scheduled maintenance",
		Type:          "info",
		DisplayType:   "banner",
		Category:      "update",
		LinkURL:       "https://example.com/details",
		LinkText:      "Learn more",
		IsActive:      true,
		IsDismissible: true,
		Priority:      10,
		StartsAt:      &startsAt,
		EndsAt:        &endsAt,
		PublishedAt:   &publishedAt,
		TargetRoles:   []string{"user", "admin"},
	}

	resp := banner.ToResponse()

	if resp.ID != banner.ID {
		t.Errorf("ID = %d, want %d", resp.ID, banner.ID)
	}
	if resp.Title != banner.Title {
		t.Errorf("Title = %q, want %q", resp.Title, banner.Title)
	}
	if resp.Message != banner.Message {
		t.Errorf("Message = %q, want %q", resp.Message, banner.Message)
	}
	if resp.Type != banner.Type {
		t.Errorf("Type = %q, want %q", resp.Type, banner.Type)
	}
	if resp.StartsAt != startsAt {
		t.Errorf("StartsAt = %q, want %q", resp.StartsAt, startsAt)
	}
	if resp.EndsAt != endsAt {
		t.Errorf("EndsAt = %q, want %q", resp.EndsAt, endsAt)
	}
	if resp.PublishedAt != publishedAt {
		t.Errorf("PublishedAt = %q, want %q", resp.PublishedAt, publishedAt)
	}
	if len(resp.TargetRoles) != 2 {
		t.Errorf("TargetRoles length = %d, want 2", len(resp.TargetRoles))
	}
}

func TestAnnouncementBanner_ToResponse_NilOptionalFields(t *testing.T) {
	banner := &AnnouncementBanner{
		ID:          1,
		Title:       "Test",
		Message:     "Test message",
		StartsAt:    nil,
		EndsAt:      nil,
		PublishedAt: nil,
	}

	resp := banner.ToResponse()

	if resp.StartsAt != "" {
		t.Errorf("StartsAt should be empty, got %q", resp.StartsAt)
	}
	if resp.EndsAt != "" {
		t.Errorf("EndsAt should be empty, got %q", resp.EndsAt)
	}
	if resp.PublishedAt != "" {
		t.Errorf("PublishedAt should be empty, got %q", resp.PublishedAt)
	}
}

func TestLoginHistory_ToResponse(t *testing.T) {
	now := time.Now()
	history := &LoginHistory{
		ID:            1,
		UserID:        42,
		Success:       true,
		FailureReason: "",
		IPAddress:     "192.168.1.100",
		DeviceInfo:    []byte(`{"browser":"Firefox","os":"Windows"}`),
		Location:      []byte(`{"country":"Germany","city":"Berlin"}`),
		AuthMethod:    AuthMethodPassword,
		CreatedAt:     now,
	}

	resp := history.ToResponse()

	if resp.ID != history.ID {
		t.Errorf("ID = %d, want %d", resp.ID, history.ID)
	}
	if !resp.Success {
		t.Error("Success should be true")
	}
	if resp.IPAddress != history.IPAddress {
		t.Errorf("IPAddress = %q, want %q", resp.IPAddress, history.IPAddress)
	}
	if resp.DeviceInfo.Browser != "Firefox" {
		t.Errorf("DeviceInfo.Browser = %q, want 'Firefox'", resp.DeviceInfo.Browser)
	}
	if resp.Location.Country != "Germany" {
		t.Errorf("Location.Country = %q, want 'Germany'", resp.Location.Country)
	}
	if resp.AuthMethod != AuthMethodPassword {
		t.Errorf("AuthMethod = %q, want %q", resp.AuthMethod, AuthMethodPassword)
	}
}

func TestLoginHistory_ToResponse_FailedLogin(t *testing.T) {
	now := time.Now()
	history := &LoginHistory{
		ID:            2,
		Success:       false,
		FailureReason: LoginFailureInvalidPassword,
		AuthMethod:    AuthMethodPassword,
		CreatedAt:     now,
	}

	resp := history.ToResponse()

	if resp.Success {
		t.Error("Success should be false")
	}
	if resp.FailureReason != LoginFailureInvalidPassword {
		t.Errorf("FailureReason = %q, want %q", resp.FailureReason, LoginFailureInvalidPassword)
	}
}

func TestLoginHistory_TableName(t *testing.T) {
	history := LoginHistory{}
	tableName := history.TableName()

	if tableName != "login_history" {
		t.Errorf("TableName() = %q, want 'login_history'", tableName)
	}
}

func TestEmailTemplate_ToResponse(t *testing.T) {
	template := &EmailTemplate{
		ID:                 1,
		Key:                "welcome_email",
		Name:               "Welcome Email",
		Description:        "Sent to new users",
		Subject:            "Welcome to our platform!",
		BodyHTML:           "<h1>Welcome {{name}}</h1>",
		BodyText:           "Welcome {{name}}",
		AvailableVariables: []byte(`[{"name":"name","description":"User's name"},{"name":"email","description":"User's email"}]`),
		IsActive:           true,
		IsSystem:           true,
		SendCount:          100,
		UpdatedAt:          "2025-01-01T00:00:00Z",
	}

	resp := template.ToResponse()

	if resp.ID != template.ID {
		t.Errorf("ID = %d, want %d", resp.ID, template.ID)
	}
	if resp.Key != template.Key {
		t.Errorf("Key = %q, want %q", resp.Key, template.Key)
	}
	if resp.Name != template.Name {
		t.Errorf("Name = %q, want %q", resp.Name, template.Name)
	}
	if resp.Subject != template.Subject {
		t.Errorf("Subject = %q, want %q", resp.Subject, template.Subject)
	}
	if !resp.IsActive {
		t.Error("IsActive should be true")
	}
	if !resp.IsSystem {
		t.Error("IsSystem should be true")
	}
	if len(resp.AvailableVariables) != 2 {
		t.Errorf("AvailableVariables length = %d, want 2", len(resp.AvailableVariables))
	}
	if resp.AvailableVariables[0].Name != "name" {
		t.Errorf("AvailableVariables[0].Name = %q, want 'name'", resp.AvailableVariables[0].Name)
	}
}

func TestEmailTemplate_ToResponse_NilVariables(t *testing.T) {
	template := &EmailTemplate{
		ID:                 1,
		Key:                "simple_email",
		AvailableVariables: nil,
	}

	resp := template.ToResponse()

	if len(resp.AvailableVariables) != 0 {
		t.Errorf("AvailableVariables should be empty, got %d items", len(resp.AvailableVariables))
	}
}

func TestDataExport_ToResponse(t *testing.T) {
	downloadURL := "https://example.com/exports/123.zip"
	expiresAt := "2025-02-01T00:00:00Z"

	export := &DataExport{
		ID:          1,
		UserID:      42,
		Status:      ExportStatusCompleted,
		DownloadURL: &downloadURL,
		FileSize:    1024000,
		RequestedAt: "2025-01-15T00:00:00Z",
		ExpiresAt:   &expiresAt,
	}

	resp := export.ToResponse()

	if resp.ID != export.ID {
		t.Errorf("ID = %d, want %d", resp.ID, export.ID)
	}
	if resp.Status != ExportStatusCompleted {
		t.Errorf("Status = %q, want %q", resp.Status, ExportStatusCompleted)
	}
	if resp.DownloadURL != downloadURL {
		t.Errorf("DownloadURL = %q, want %q", resp.DownloadURL, downloadURL)
	}
	if resp.ExpiresAt != expiresAt {
		t.Errorf("ExpiresAt = %q, want %q", resp.ExpiresAt, expiresAt)
	}
	if resp.FileSize != export.FileSize {
		t.Errorf("FileSize = %d, want %d", resp.FileSize, export.FileSize)
	}
}

func TestDataExport_ToResponse_Pending(t *testing.T) {
	export := &DataExport{
		ID:          2,
		Status:      ExportStatusPending,
		DownloadURL: nil,
		ExpiresAt:   nil,
		RequestedAt: "2025-01-15T00:00:00Z",
	}

	resp := export.ToResponse()

	if resp.Status != ExportStatusPending {
		t.Errorf("Status = %q, want %q", resp.Status, ExportStatusPending)
	}
	if resp.DownloadURL != "" {
		t.Errorf("DownloadURL should be empty, got %q", resp.DownloadURL)
	}
	if resp.ExpiresAt != "" {
		t.Errorf("ExpiresAt should be empty, got %q", resp.ExpiresAt)
	}
}

// ============ Login Failure Reason Constants Tests ============

func TestLoginFailureReasonConstants(t *testing.T) {
	tests := []struct {
		constant string
		expected string
	}{
		{LoginFailureInvalidPassword, "invalid_password"},
		{LoginFailureAccountLocked, "account_locked"},
		{LoginFailure2FAFailed, "2fa_failed"},
		{LoginFailureAccountInactive, "account_inactive"},
		{LoginFailureEmailNotFound, "email_not_found"},
	}

	for _, tt := range tests {
		if tt.constant != tt.expected {
			t.Errorf("constant = %q, want %q", tt.constant, tt.expected)
		}
	}
}

// ============ Auth Method Constants Tests ============

func TestAuthMethodConstants(t *testing.T) {
	tests := []struct {
		constant string
		expected string
	}{
		{AuthMethodPassword, "password"},
		{AuthMethodOAuthGoogle, "oauth_google"},
		{AuthMethodOAuthGitHub, "oauth_github"},
		{AuthMethodRefreshToken, "refresh_token"},
		{AuthMethod2FA, "2fa"},
	}

	for _, tt := range tests {
		if tt.constant != tt.expected {
			t.Errorf("constant = %q, want %q", tt.constant, tt.expected)
		}
	}
}

// ============ Export Status Constants Tests ============

func TestExportStatusConstants(t *testing.T) {
	tests := []struct {
		constant string
		expected string
	}{
		{ExportStatusPending, "pending"},
		{ExportStatusProcessing, "processing"},
		{ExportStatusCompleted, "completed"},
		{ExportStatusFailed, "failed"},
		{ExportStatusExpired, "expired"},
	}

	for _, tt := range tests {
		if tt.constant != tt.expected {
			t.Errorf("constant = %q, want %q", tt.constant, tt.expected)
		}
	}
}
