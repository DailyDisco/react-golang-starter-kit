package ai

import (
	"context"
	"sync"
	"testing"
)

// ============ NoOp Service Tests ============

func TestNoOpService_IsAvailable(t *testing.T) {
	svc := &noOpService{}

	if svc.IsAvailable() {
		t.Error("noOpService.IsAvailable() = true, want false")
	}
}

func TestNoOpService_GetModel(t *testing.T) {
	svc := &noOpService{}

	if model := svc.GetModel(); model != "" {
		t.Errorf("noOpService.GetModel() = %q, want empty", model)
	}
}

func TestNoOpService_GetConfig(t *testing.T) {
	svc := &noOpService{}

	if config := svc.GetConfig(); config != nil {
		t.Error("noOpService.GetConfig() should return nil")
	}
}

func TestNoOpService_ValidatePrompt(t *testing.T) {
	svc := &noOpService{}

	err := svc.ValidatePrompt("test prompt")
	if err != ErrDisabled {
		t.Errorf("noOpService.ValidatePrompt() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_ValidateMessages(t *testing.T) {
	svc := &noOpService{}

	err := svc.ValidateMessages([]Message{{Role: RoleUser, Content: "test"}})
	if err != ErrDisabled {
		t.Errorf("noOpService.ValidateMessages() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_ValidateTexts(t *testing.T) {
	svc := &noOpService{}

	err := svc.ValidateTexts([]string{"test"})
	if err != ErrDisabled {
		t.Errorf("noOpService.ValidateTexts() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GenerateText(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GenerateText(context.Background(), "test", nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.GenerateText() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_Chat(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.Chat(context.Background(), nil, nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.Chat() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_StreamChat(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.StreamChat(context.Background(), nil, nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.StreamChat() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_ChatAdvanced(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.ChatAdvanced(context.Background(), nil, nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.ChatAdvanced() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_AnalyzeImage(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.AnalyzeImage(context.Background(), ImageInput{}, "describe")
	if err != ErrDisabled {
		t.Errorf("noOpService.AnalyzeImage() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GenerateWithImages(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GenerateWithImages(context.Background(), "prompt", nil, nil)
	if err != ErrDisabled {
		t.Errorf("noOpService.GenerateWithImages() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GenerateEmbedding(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GenerateEmbedding(context.Background(), "test")
	if err != ErrDisabled {
		t.Errorf("noOpService.GenerateEmbedding() error = %v, want %v", err, ErrDisabled)
	}
}

func TestNoOpService_GenerateEmbeddings(t *testing.T) {
	svc := &noOpService{}

	_, err := svc.GenerateEmbeddings(context.Background(), []string{"test"})
	if err != ErrDisabled {
		t.Errorf("noOpService.GenerateEmbeddings() error = %v, want %v", err, ErrDisabled)
	}
}

// ============ GeminiService Validation Tests ============

func TestGeminiService_ValidatePrompt(t *testing.T) {
	svc := &geminiService{
		config: &Config{
			MaxPromptLength: 100,
		},
	}

	tests := []struct {
		name    string
		prompt  string
		wantErr error
	}{
		{"valid short prompt", "Hello", nil},
		{"valid at limit", string(make([]byte, 100)), nil},
		{"too long prompt", string(make([]byte, 101)), ErrPromptTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePrompt(tt.prompt)
			if err != tt.wantErr {
				t.Errorf("ValidatePrompt() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeminiService_ValidateMessages(t *testing.T) {
	svc := &geminiService{
		config: &Config{
			MaxMessagesPerChat: 3,
			MaxPromptLength:    100,
		},
	}

	tests := []struct {
		name     string
		messages []Message
		wantErr  error
	}{
		{
			name:     "valid messages",
			messages: []Message{{Role: RoleUser, Content: "Hi"}},
			wantErr:  nil,
		},
		{
			name: "too many messages",
			messages: []Message{
				{Role: RoleUser, Content: "1"},
				{Role: RoleModel, Content: "2"},
				{Role: RoleUser, Content: "3"},
				{Role: RoleModel, Content: "4"},
			},
			wantErr: ErrTooManyMessages,
		},
		{
			name: "content too long",
			messages: []Message{
				{Role: RoleUser, Content: string(make([]byte, 101))},
			},
			wantErr: ErrPromptTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateMessages(tt.messages)
			if err != tt.wantErr {
				t.Errorf("ValidateMessages() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeminiService_ValidateTexts(t *testing.T) {
	svc := &geminiService{
		config: &Config{
			MaxTextsPerEmbed: 3,
		},
	}

	tests := []struct {
		name    string
		texts   []string
		wantErr error
	}{
		{"valid texts", []string{"a", "b"}, nil},
		{"at limit", []string{"a", "b", "c"}, nil},
		{"too many texts", []string{"a", "b", "c", "d"}, ErrTooManyTexts},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateTexts(tt.texts)
			if err != tt.wantErr {
				t.Errorf("ValidateTexts() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeminiService_GetModel(t *testing.T) {
	svc := &geminiService{
		config: &Config{
			Model: "gemini-2.0-flash",
		},
	}

	if model := svc.GetModel(); model != "gemini-2.0-flash" {
		t.Errorf("GetModel() = %q, want %q", model, "gemini-2.0-flash")
	}
}

func TestGeminiService_GetConfig(t *testing.T) {
	config := &Config{
		Model:   "gemini-2.0-flash",
		Enabled: true,
	}
	svc := &geminiService{config: config}

	if svc.GetConfig() != config {
		t.Error("GetConfig() should return the config")
	}
}

func TestGeminiService_IsAvailable(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &geminiService{
				config: &Config{Enabled: tt.enabled},
			}

			if got := svc.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============ Initialize Tests ============

func TestInitialize_Disabled(t *testing.T) {
	// Reset singleton for testing
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled: false,
	}

	err := Initialize(config)
	if err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}

	svc := GetService()
	if svc == nil {
		t.Fatal("GetService() returned nil after Initialize")
	}

	if svc.IsAvailable() {
		t.Error("Service should not be available when disabled")
	}
}

func TestInitialize_MissingAPIKey(t *testing.T) {
	// Reset singleton for testing
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled: true,
		APIKey:  "",
	}

	err := Initialize(config)
	if err != ErrMissingAPIKey {
		t.Errorf("Initialize() error = %v, want %v", err, ErrMissingAPIKey)
	}
}

// ============ GetService Tests ============

func TestGetService_BeforeInit(t *testing.T) {
	// Reset singleton
	once = sync.Once{}
	instance = nil

	svc := GetService()
	if svc != nil {
		t.Error("GetService() should return nil before initialization")
	}
}

// ============ IsAvailable Package Function Tests ============

func TestIsAvailable_NotInitialized(t *testing.T) {
	// Reset singleton
	once = sync.Once{}
	instance = nil

	if IsAvailable() {
		t.Error("IsAvailable() = true when not initialized, want false")
	}
}

func TestIsAvailable_Disabled(t *testing.T) {
	// Reset singleton and initialize with disabled config
	once = sync.Once{}
	instance = nil

	config := &Config{
		Enabled: false,
	}
	Initialize(config)

	if IsAvailable() {
		t.Error("IsAvailable() = true when disabled, want false")
	}
}
