package ai

import (
	"testing"
)

// ============ Error Types Tests ============

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrMissingAPIKey", ErrMissingAPIKey, "ai: missing API key"},
		{"ErrDisabled", ErrDisabled, "ai: service is disabled"},
		{"ErrInvalidRole", ErrInvalidRole, "ai: invalid message role"},
		{"ErrEmptyPrompt", ErrEmptyPrompt, "ai: empty prompt"},
		{"ErrImageTooLarge", ErrImageTooLarge, "ai: image exceeds maximum size"},
		{"ErrInvalidImage", ErrInvalidImage, "ai: invalid image data"},
		{"ErrEmptyTexts", ErrEmptyTexts, "ai: empty texts for embedding"},
		{"ErrPromptTooLong", ErrPromptTooLong, "ai: prompt exceeds maximum length"},
		{"ErrTooManyMessages", ErrTooManyMessages, "ai: too many messages in chat"},
		{"ErrTooManyTexts", ErrTooManyTexts, "ai: too many texts for embedding"},
		{"ErrContentBlocked", ErrContentBlocked, "ai: content blocked by safety filters"},
		{"ErrFunctionNotAllowed", ErrFunctionNotAllowed, "ai: function calling is not enabled"},
		{"ErrJSONModeNotAllowed", ErrJSONModeNotAllowed, "ai: JSON mode is not enabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

// ============ Role Constants Tests ============

func TestRoleConstants(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleUser, "user"},
		{RoleModel, "model"},
		{RoleSystem, "system"},
		{RoleAssistant, "assistant"},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("Role = %q, want %q", tt.role, tt.want)
			}
		})
	}
}

// ============ Message Structure Tests ============

func TestMessage_Creation(t *testing.T) {
	msg := Message{
		Role:    RoleUser,
		Content: "Hello, world!",
	}

	if msg.Role != RoleUser {
		t.Errorf("Message.Role = %q, want %q", msg.Role, RoleUser)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("Message.Content = %q, want %q", msg.Content, "Hello, world!")
	}
}

// ============ ImageInput Structure Tests ============

func TestImageInput_Creation(t *testing.T) {
	tests := []struct {
		name     string
		input    ImageInput
		wantData string
		wantMime string
		wantURL  string
	}{
		{
			name: "with base64 data",
			input: ImageInput{
				Data:     "base64encodeddata",
				MimeType: "image/jpeg",
			},
			wantData: "base64encodeddata",
			wantMime: "image/jpeg",
			wantURL:  "",
		},
		{
			name: "with URL",
			input: ImageInput{
				URL: "https://example.com/image.png",
			},
			wantData: "",
			wantMime: "",
			wantURL:  "https://example.com/image.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.Data != tt.wantData {
				t.Errorf("ImageInput.Data = %q, want %q", tt.input.Data, tt.wantData)
			}
			if tt.input.MimeType != tt.wantMime {
				t.Errorf("ImageInput.MimeType = %q, want %q", tt.input.MimeType, tt.wantMime)
			}
			if tt.input.URL != tt.wantURL {
				t.Errorf("ImageInput.URL = %q, want %q", tt.input.URL, tt.wantURL)
			}
		})
	}
}

// ============ GenerateOptions Tests ============

func TestGenerateOptions_Creation(t *testing.T) {
	temp := float32(0.7)
	maxTokens := 1024
	topP := float32(0.95)
	topK := 40

	opts := GenerateOptions{
		Temperature:  &temp,
		MaxTokens:    &maxTokens,
		TopP:         &topP,
		TopK:         &topK,
		StopSequence: []string{"STOP", "END"},
	}

	if *opts.Temperature != temp {
		t.Errorf("GenerateOptions.Temperature = %v, want %v", *opts.Temperature, temp)
	}

	if *opts.MaxTokens != maxTokens {
		t.Errorf("GenerateOptions.MaxTokens = %v, want %v", *opts.MaxTokens, maxTokens)
	}

	if len(opts.StopSequence) != 2 {
		t.Errorf("GenerateOptions.StopSequence length = %d, want 2", len(opts.StopSequence))
	}
}

// ============ Response Structure Tests ============

func TestResponse_Creation(t *testing.T) {
	resp := Response{
		Content: "Generated text",
		Model:   "gemini-2.0-flash",
		Usage: &Usage{
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
		},
	}

	if resp.Content != "Generated text" {
		t.Errorf("Response.Content = %q, want %q", resp.Content, "Generated text")
	}

	if resp.Model != "gemini-2.0-flash" {
		t.Errorf("Response.Model = %q, want %q", resp.Model, "gemini-2.0-flash")
	}

	if resp.Usage.TotalTokens != 150 {
		t.Errorf("Response.Usage.TotalTokens = %d, want 150", resp.Usage.TotalTokens)
	}
}

// ============ StreamChunk Tests ============

func TestStreamChunk_Creation(t *testing.T) {
	tests := []struct {
		name  string
		chunk StreamChunk
	}{
		{
			name: "token chunk",
			chunk: StreamChunk{
				Token: "Hello",
				Done:  false,
			},
		},
		{
			name: "done chunk",
			chunk: StreamChunk{
				Token: "",
				Done:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.chunk.Done && tt.chunk.Token != "" {
				t.Error("Done chunk should have empty token")
			}
		})
	}
}

// ============ FunctionDeclaration Tests ============

func TestFunctionDeclaration_Creation(t *testing.T) {
	fn := FunctionDeclaration{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "The city name",
				},
			},
			"required": []string{"location"},
		},
	}

	if fn.Name != "get_weather" {
		t.Errorf("FunctionDeclaration.Name = %q, want %q", fn.Name, "get_weather")
	}

	if fn.Parameters == nil {
		t.Error("FunctionDeclaration.Parameters should not be nil")
	}
}

// ============ FunctionCall Tests ============

func TestFunctionCall_Creation(t *testing.T) {
	call := FunctionCall{
		Name: "get_weather",
		Args: map[string]interface{}{
			"location": "San Francisco",
		},
	}

	if call.Name != "get_weather" {
		t.Errorf("FunctionCall.Name = %q, want %q", call.Name, "get_weather")
	}

	if call.Args["location"] != "San Francisco" {
		t.Errorf("FunctionCall.Args[location] = %v, want %q", call.Args["location"], "San Francisco")
	}
}

// ============ JSONSchema Tests ============

func TestJSONSchema_Creation(t *testing.T) {
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "integer",
			},
		},
		Required:    []string{"name"},
		Description: "A person object",
	}

	if schema.Type != "object" {
		t.Errorf("JSONSchema.Type = %q, want %q", schema.Type, "object")
	}

	if len(schema.Required) != 1 {
		t.Errorf("JSONSchema.Required length = %d, want 1", len(schema.Required))
	}
}
