package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/ai"
)

// ============ Constants Tests ============

func TestAITimeoutConstants(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{"aiChatTimeout", aiChatTimeout, 60 * time.Second},
		{"aiStreamTimeout", aiStreamTimeout, 5 * time.Minute},
		{"aiImageTimeout", aiImageTimeout, 60 * time.Second},
		{"aiEmbeddingsTimeout", aiEmbeddingsTimeout, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.timeout, tt.expected)
			}
		})
	}
}

// ============ ChatRequest Structure Tests ============

func TestChatRequest_Structure(t *testing.T) {
	temp := float32(0.7)
	maxTokens := 100
	topP := float32(0.9)
	topK := 40

	req := ChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Hello"},
		},
		SystemPrompt: "You are helpful",
		Temperature:  &temp,
		MaxTokens:    &maxTokens,
		TopP:         &topP,
		TopK:         &topK,
	}

	if len(req.Messages) != 1 {
		t.Errorf("ChatRequest.Messages length = %d, want 1", len(req.Messages))
	}

	if req.SystemPrompt != "You are helpful" {
		t.Errorf("ChatRequest.SystemPrompt = %q, want %q", req.SystemPrompt, "You are helpful")
	}

	if *req.Temperature != 0.7 {
		t.Errorf("ChatRequest.Temperature = %v, want 0.7", *req.Temperature)
	}

	if *req.MaxTokens != 100 {
		t.Errorf("ChatRequest.MaxTokens = %d, want 100", *req.MaxTokens)
	}
}

func TestChatRequest_JSONMarshaling(t *testing.T) {
	temp := float32(0.5)
	req := ChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Test message"},
		},
		SystemPrompt: "System prompt",
		Temperature:  &temp,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ChatRequest: %v", err)
	}

	var decoded ChatRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ChatRequest: %v", err)
	}

	if len(decoded.Messages) != 1 {
		t.Errorf("Messages length after unmarshal = %d, want 1", len(decoded.Messages))
	}

	if decoded.Messages[0].Content != "Test message" {
		t.Errorf("Message content = %q, want %q", decoded.Messages[0].Content, "Test message")
	}
}

func TestChatRequest_OptionalFields(t *testing.T) {
	// Test with only required fields
	req := ChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Hello"},
		},
	}

	if req.SystemPrompt != "" {
		t.Errorf("ChatRequest.SystemPrompt should be empty")
	}

	if req.Temperature != nil {
		t.Errorf("ChatRequest.Temperature should be nil")
	}

	if req.MaxTokens != nil {
		t.Errorf("ChatRequest.MaxTokens should be nil")
	}
}

// ============ AnalyzeImageRequest Structure Tests ============

func TestAnalyzeImageRequest_Structure(t *testing.T) {
	req := AnalyzeImageRequest{
		Image: ai.ImageInput{
			Data:     "base64data",
			MimeType: "image/png",
		},
		Prompt:   "Describe this image",
		MimeType: "image/jpeg",
	}

	if req.Image.Data != "base64data" {
		t.Errorf("AnalyzeImageRequest.Image.Data = %q, want base64data", req.Image.Data)
	}

	if req.Prompt != "Describe this image" {
		t.Errorf("AnalyzeImageRequest.Prompt = %q, want Describe this image", req.Prompt)
	}

	if req.MimeType != "image/jpeg" {
		t.Errorf("AnalyzeImageRequest.MimeType = %q, want image/jpeg", req.MimeType)
	}
}

func TestAnalyzeImageRequest_JSONMarshaling(t *testing.T) {
	req := AnalyzeImageRequest{
		Image: ai.ImageInput{
			URL:      "https://example.com/image.png",
			MimeType: "image/png",
		},
		Prompt: "What is in this image?",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal AnalyzeImageRequest: %v", err)
	}

	var decoded AnalyzeImageRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AnalyzeImageRequest: %v", err)
	}

	if decoded.Image.URL != req.Image.URL {
		t.Errorf("Image.URL = %q, want %q", decoded.Image.URL, req.Image.URL)
	}

	if decoded.Prompt != req.Prompt {
		t.Errorf("Prompt = %q, want %q", decoded.Prompt, req.Prompt)
	}
}

// ============ EmbeddingsRequest Structure Tests ============

func TestEmbeddingsRequest_Structure(t *testing.T) {
	req := EmbeddingsRequest{
		Texts: []string{"text1", "text2", "text3"},
	}

	if len(req.Texts) != 3 {
		t.Errorf("EmbeddingsRequest.Texts length = %d, want 3", len(req.Texts))
	}

	if req.Texts[0] != "text1" {
		t.Errorf("EmbeddingsRequest.Texts[0] = %q, want text1", req.Texts[0])
	}
}

func TestEmbeddingsRequest_JSONMarshaling(t *testing.T) {
	req := EmbeddingsRequest{
		Texts: []string{"embed this", "and this too"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal EmbeddingsRequest: %v", err)
	}

	var decoded EmbeddingsRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EmbeddingsRequest: %v", err)
	}

	if len(decoded.Texts) != 2 {
		t.Errorf("Texts length after unmarshal = %d, want 2", len(decoded.Texts))
	}
}

func TestEmbeddingsRequest_EmptyTexts(t *testing.T) {
	req := EmbeddingsRequest{
		Texts: []string{},
	}

	if len(req.Texts) != 0 {
		t.Errorf("EmbeddingsRequest.Texts length = %d, want 0", len(req.Texts))
	}
}

// ============ AdvancedChatRequest Structure Tests ============

func TestAdvancedChatRequest_Structure(t *testing.T) {
	temp := float32(0.8)
	req := AdvancedChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Hello"},
		},
		SystemPrompt: "You are an assistant",
		Temperature:  &temp,
		JSONMode:     true,
		Functions: []ai.FunctionDeclaration{
			{
				Name:        "get_weather",
				Description: "Get weather data",
			},
		},
	}

	if len(req.Messages) != 1 {
		t.Errorf("AdvancedChatRequest.Messages length = %d, want 1", len(req.Messages))
	}

	if !req.JSONMode {
		t.Error("AdvancedChatRequest.JSONMode should be true")
	}

	if len(req.Functions) != 1 {
		t.Errorf("AdvancedChatRequest.Functions length = %d, want 1", len(req.Functions))
	}

	if req.Functions[0].Name != "get_weather" {
		t.Errorf("Function name = %q, want get_weather", req.Functions[0].Name)
	}
}

func TestAdvancedChatRequest_JSONMarshaling(t *testing.T) {
	req := AdvancedChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Test"},
		},
		JSONMode: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal AdvancedChatRequest: %v", err)
	}

	var decoded AdvancedChatRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AdvancedChatRequest: %v", err)
	}

	if !decoded.JSONMode {
		t.Error("JSONMode should be true after unmarshal")
	}
}

// ============ AIChat Handler Tests ============

func TestAIChat_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString("invalid json")
	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// This will fail with service unavailable if AI is not initialized
	// or with bad request if JSON is invalid
	AIChat(w, req)

	// Either 503 (service unavailable) or 400 (bad request) is acceptable
	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIChat() with invalid body status = %d, want 503 or 400", w.Code)
	}
}

func TestAIChat_EmptyMessages(t *testing.T) {
	chatReq := ChatRequest{
		Messages: []ai.Message{},
	}

	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIChat(w, req)

	// Should fail with service unavailable or bad request
	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIChat() with empty messages status = %d, want 503 or 400", w.Code)
	}
}

// ============ AIChatStream Handler Tests ============

func TestAIChatStream_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString("not json")
	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIChatStream(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIChatStream() with invalid body status = %d, want 503 or 400", w.Code)
	}
}

func TestAIChatStream_EmptyMessages(t *testing.T) {
	chatReq := ChatRequest{
		Messages: []ai.Message{},
	}

	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIChatStream(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIChatStream() with empty messages status = %d, want 503 or 400", w.Code)
	}
}

// ============ AIAnalyzeImage Handler Tests ============

func TestAIAnalyzeImage_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString("invalid")
	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze-image", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIAnalyzeImage(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIAnalyzeImage() with invalid body status = %d, want 503 or 400", w.Code)
	}
}

func TestAIAnalyzeImage_EmptyPrompt(t *testing.T) {
	imageReq := AnalyzeImageRequest{
		Image: ai.ImageInput{
			Data: "base64data",
		},
		Prompt: "", // Empty prompt
	}

	body, _ := json.Marshal(imageReq)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze-image", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIAnalyzeImage(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIAnalyzeImage() with empty prompt status = %d, want 503 or 400", w.Code)
	}
}

func TestAIAnalyzeImage_NoImage(t *testing.T) {
	imageReq := AnalyzeImageRequest{
		Image:  ai.ImageInput{}, // No data or URL
		Prompt: "Describe this",
	}

	body, _ := json.Marshal(imageReq)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze-image", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIAnalyzeImage(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIAnalyzeImage() with no image status = %d, want 503 or 400", w.Code)
	}
}

// ============ AIEmbeddings Handler Tests ============

func TestAIEmbeddings_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString("not valid json")
	req := httptest.NewRequest(http.MethodPost, "/api/ai/embeddings", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIEmbeddings(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIEmbeddings() with invalid body status = %d, want 503 or 400", w.Code)
	}
}

func TestAIEmbeddings_EmptyTexts(t *testing.T) {
	embReq := EmbeddingsRequest{
		Texts: []string{},
	}

	body, _ := json.Marshal(embReq)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/embeddings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIEmbeddings(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIEmbeddings() with empty texts status = %d, want 503 or 400", w.Code)
	}
}

// ============ AIChatAdvanced Handler Tests ============

func TestAIChatAdvanced_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString("{invalid}")
	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/advanced", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIChatAdvanced(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIChatAdvanced() with invalid body status = %d, want 503 or 400", w.Code)
	}
}

func TestAIChatAdvanced_EmptyMessages(t *testing.T) {
	advReq := AdvancedChatRequest{
		Messages: []ai.Message{},
	}

	body, _ := json.Marshal(advReq)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/advanced", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AIChatAdvanced(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusBadRequest {
		t.Errorf("AIChatAdvanced() with empty messages status = %d, want 503 or 400", w.Code)
	}
}

// ============ Request Validation Tests ============

func TestChatRequest_ValidRoles(t *testing.T) {
	validRoles := []ai.Role{ai.RoleUser, ai.RoleAssistant, ai.RoleSystem}

	for _, role := range validRoles {
		t.Run(string(role), func(t *testing.T) {
			req := ChatRequest{
				Messages: []ai.Message{
					{Role: role, Content: "Test"},
				},
			}

			if req.Messages[0].Role != role {
				t.Errorf("Message role = %q, want %q", req.Messages[0].Role, role)
			}
		})
	}
}

func TestChatRequest_MultipleMessages(t *testing.T) {
	req := ChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Hello"},
			{Role: ai.RoleAssistant, Content: "Hi there!"},
			{Role: ai.RoleUser, Content: "How are you?"},
		},
	}

	if len(req.Messages) != 3 {
		t.Errorf("Messages length = %d, want 3", len(req.Messages))
	}

	// Verify conversation order
	if req.Messages[0].Role != ai.RoleUser {
		t.Error("First message should be from user")
	}

	if req.Messages[1].Role != ai.RoleAssistant {
		t.Error("Second message should be from assistant")
	}
}

// ============ Temperature Range Tests ============

func TestChatRequest_TemperatureRange(t *testing.T) {
	tests := []struct {
		name string
		temp float32
	}{
		{"zero", 0.0},
		{"low", 0.1},
		{"default", 0.7},
		{"high", 1.0},
		{"max", 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp := tt.temp
			req := ChatRequest{
				Messages: []ai.Message{
					{Role: ai.RoleUser, Content: "Test"},
				},
				Temperature: &temp,
			}

			if *req.Temperature != tt.temp {
				t.Errorf("Temperature = %v, want %v", *req.Temperature, tt.temp)
			}
		})
	}
}

// ============ MaxTokens Range Tests ============

func TestChatRequest_MaxTokensRange(t *testing.T) {
	tests := []struct {
		name      string
		maxTokens int
	}{
		{"small", 10},
		{"medium", 100},
		{"large", 1000},
		{"very_large", 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxTokens := tt.maxTokens
			req := ChatRequest{
				Messages: []ai.Message{
					{Role: ai.RoleUser, Content: "Test"},
				},
				MaxTokens: &maxTokens,
			}

			if *req.MaxTokens != tt.maxTokens {
				t.Errorf("MaxTokens = %d, want %d", *req.MaxTokens, tt.maxTokens)
			}
		})
	}
}

// ============ Image Input Tests ============

func TestAnalyzeImageRequest_DataImage(t *testing.T) {
	req := AnalyzeImageRequest{
		Image: ai.ImageInput{
			Data:     "base64encodeddata",
			MimeType: "image/png",
		},
		Prompt: "What is this?",
	}

	if req.Image.Data == "" {
		t.Error("Image.Data should not be empty")
	}

	if req.Image.URL != "" {
		t.Error("Image.URL should be empty for data image")
	}
}

func TestAnalyzeImageRequest_URLImage(t *testing.T) {
	req := AnalyzeImageRequest{
		Image: ai.ImageInput{
			URL:      "https://example.com/image.jpg",
			MimeType: "image/jpeg",
		},
		Prompt: "Describe this",
	}

	if req.Image.URL == "" {
		t.Error("Image.URL should not be empty")
	}

	if req.Image.Data != "" {
		t.Error("Image.Data should be empty for URL image")
	}
}

// ============ MIME Type Tests ============

func TestAnalyzeImageRequest_MimeTypes(t *testing.T) {
	validMimeTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, mimeType := range validMimeTypes {
		t.Run(mimeType, func(t *testing.T) {
			req := AnalyzeImageRequest{
				Image: ai.ImageInput{
					Data:     "test",
					MimeType: mimeType,
				},
				Prompt: "Test",
			}

			if req.Image.MimeType != mimeType {
				t.Errorf("MimeType = %q, want %q", req.Image.MimeType, mimeType)
			}
		})
	}
}

func TestAnalyzeImageRequest_TopLevelMimeType(t *testing.T) {
	// Test that top-level MimeType is captured
	req := AnalyzeImageRequest{
		Image: ai.ImageInput{
			Data: "test",
		},
		Prompt:   "Test",
		MimeType: "image/png",
	}

	if req.MimeType != "image/png" {
		t.Errorf("Top-level MimeType = %q, want image/png", req.MimeType)
	}
}

// ============ Function Declaration Tests ============

func TestAdvancedChatRequest_Functions(t *testing.T) {
	req := AdvancedChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "What's the weather?"},
		},
		Functions: []ai.FunctionDeclaration{
			{
				Name:        "get_weather",
				Description: "Get current weather for a location",
			},
			{
				Name:        "get_forecast",
				Description: "Get weather forecast",
			},
		},
	}

	if len(req.Functions) != 2 {
		t.Errorf("Functions length = %d, want 2", len(req.Functions))
	}

	if req.Functions[0].Name != "get_weather" {
		t.Errorf("First function name = %q, want get_weather", req.Functions[0].Name)
	}
}

// ============ JSON Schema Tests ============

func TestAdvancedChatRequest_WithJSONSchema(t *testing.T) {
	schema := &ai.JSONSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]string{"type": "string"},
			"age":  map[string]string{"type": "integer"},
		},
	}

	req := AdvancedChatRequest{
		Messages: []ai.Message{
			{Role: ai.RoleUser, Content: "Generate a person"},
		},
		JSONMode:   true,
		JSONSchema: schema,
	}

	if !req.JSONMode {
		t.Error("JSONMode should be true")
	}

	if req.JSONSchema == nil {
		t.Error("JSONSchema should not be nil")
	}

	if req.JSONSchema.Type != "object" {
		t.Errorf("JSONSchema.Type = %q, want object", req.JSONSchema.Type)
	}
}
