package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"react-golang-starter/internal/ai"
)

// AI operation timeouts
const (
	aiChatTimeout       = 60 * time.Second // Chat completions
	aiStreamTimeout     = 5 * time.Minute  // Streaming (longer for token-by-token)
	aiImageTimeout      = 60 * time.Second // Image analysis
	aiEmbeddingsTimeout = 30 * time.Second // Embeddings generation
)

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages     []ai.Message `json:"messages" validate:"required,min=1"`
	SystemPrompt string       `json:"systemPrompt,omitempty"`
	Temperature  *float32     `json:"temperature,omitempty"`
	MaxTokens    *int         `json:"maxTokens,omitempty"`
	TopP         *float32     `json:"topP,omitempty"`
	TopK         *int         `json:"topK,omitempty"`
}

// AnalyzeImageRequest represents an image analysis request
type AnalyzeImageRequest struct {
	Image    ai.ImageInput `json:"image" validate:"required"`
	Prompt   string        `json:"prompt" validate:"required"`
	MimeType string        `json:"mimeType,omitempty"`
}

// EmbeddingsRequest represents an embeddings generation request
type EmbeddingsRequest struct {
	Texts []string `json:"texts" validate:"required,min=1"`
}

// AIChat handles chat completion requests
// POST /api/ai/chat
func AIChat(w http.ResponseWriter, r *http.Request) {
	// Check if AI service is available
	if !ai.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service is not available")
		return
	}

	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		WriteBadRequest(w, r, "At least one message is required")
		return
	}

	// Validate messages using service validation
	if err := ai.GetService().ValidateMessages(req.Messages); err != nil {
		switch err {
		case ai.ErrTooManyMessages:
			WriteBadRequest(w, r, "Too many messages in chat history")
		case ai.ErrPromptTooLong:
			WriteBadRequest(w, r, "Message content exceeds maximum length")
		case ai.ErrInvalidRole:
			WriteBadRequest(w, r, "Invalid message role")
		default:
			log.Error().Err(err).Msg("unexpected error validating chat messages")
			WriteBadRequest(w, r, "Invalid message format")
		}
		return
	}

	// Build chat options
	opts := &ai.ChatOptions{
		SystemPrompt: req.SystemPrompt,
		GenerateOptions: ai.GenerateOptions{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			TopP:        req.TopP,
			TopK:        req.TopK,
		},
	}

	// Generate response with timeout
	ctx, cancel := context.WithTimeout(r.Context(), aiChatTimeout)
	defer cancel()

	resp, err := ai.GetService().Chat(ctx, req.Messages, opts)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			WriteError(w, r, http.StatusGatewayTimeout, "TIMEOUT", "AI request timed out")
			return
		}
		if err == ai.ErrContentBlocked {
			WriteError(w, r, http.StatusUnprocessableEntity, "CONTENT_BLOCKED", "Content was blocked by safety filters")
			return
		}
		log.Error().Err(err).Msg("failed to generate chat response")
		WriteInternalError(w, r, "Failed to generate response")
		return
	}

	WriteSuccess(w, "Chat response generated", resp)
}

// AIChatStream handles streaming chat completion requests
// POST /api/ai/chat/stream
func AIChatStream(w http.ResponseWriter, r *http.Request) {
	// Check if AI service is available
	if !ai.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service is not available")
		return
	}

	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		WriteBadRequest(w, r, "At least one message is required")
		return
	}

	// Validate messages using service validation
	if err := ai.GetService().ValidateMessages(req.Messages); err != nil {
		switch err {
		case ai.ErrTooManyMessages:
			WriteBadRequest(w, r, "Too many messages in chat history")
		case ai.ErrPromptTooLong:
			WriteBadRequest(w, r, "Message content exceeds maximum length")
		case ai.ErrInvalidRole:
			WriteBadRequest(w, r, "Invalid message role")
		default:
			log.Error().Err(err).Msg("unexpected error validating chat messages")
			WriteBadRequest(w, r, "Invalid message format")
		}
		return
	}

	// Build chat options
	opts := &ai.ChatOptions{
		SystemPrompt: req.SystemPrompt,
		GenerateOptions: ai.GenerateOptions{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			TopP:        req.TopP,
			TopK:        req.TopK,
		},
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteInternalError(w, r, "Streaming not supported")
		return
	}

	// Start streaming with timeout
	ctx, cancel := context.WithTimeout(r.Context(), aiStreamTimeout)
	defer cancel()

	chunks, err := ai.GetService().StreamChat(ctx, req.Messages, opts)
	if err != nil {
		// Write error as SSE event
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Fprintf(w, "event: error\ndata: AI request timed out\n\n")
		} else {
			log.Error().Err(err).Msg("failed to start streaming chat")
			fmt.Fprintf(w, "event: error\ndata: Failed to generate response\n\n")
		}
		flusher.Flush()
		return
	}

	// Stream chunks to client
	for chunk := range chunks {
		if chunk.Error != nil {
			log.Error().Err(chunk.Error).Msg("error during streaming chat")
			fmt.Fprintf(w, "event: error\ndata: Stream interrupted\n\n")
			flusher.Flush()
			return
		}

		if chunk.Done {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			return
		}

		// Send token as JSON
		data, _ := json.Marshal(map[string]string{"token": chunk.Token})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}

// AIAnalyzeImage handles image analysis requests
// POST /api/ai/analyze-image
func AIAnalyzeImage(w http.ResponseWriter, r *http.Request) {
	// Check if AI service is available
	if !ai.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service is not available")
		return
	}

	// Parse request body
	var req AnalyzeImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request
	if req.Prompt == "" {
		WriteBadRequest(w, r, "Prompt is required")
		return
	}
	if req.Image.Data == "" && req.Image.URL == "" {
		WriteBadRequest(w, r, "Image data or URL is required")
		return
	}

	// Set MIME type if provided at top level
	if req.MimeType != "" && req.Image.MimeType == "" {
		req.Image.MimeType = req.MimeType
	}

	// Analyze image with timeout
	ctx, cancel := context.WithTimeout(r.Context(), aiImageTimeout)
	defer cancel()

	resp, err := ai.GetService().AnalyzeImage(ctx, req.Image, req.Prompt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			WriteError(w, r, http.StatusGatewayTimeout, "TIMEOUT", "Image analysis timed out")
			return
		}
		if err == ai.ErrImageTooLarge {
			WriteBadRequest(w, r, "Image exceeds maximum allowed size")
			return
		}
		if err == ai.ErrInvalidImage {
			WriteBadRequest(w, r, "Invalid image data")
			return
		}
		log.Error().Err(err).Msg("failed to analyze image")
		WriteInternalError(w, r, "Failed to analyze image")
		return
	}

	WriteSuccess(w, "Image analyzed successfully", resp)
}

// AIEmbeddings handles embeddings generation requests
// POST /api/ai/embeddings
func AIEmbeddings(w http.ResponseWriter, r *http.Request) {
	// Check if AI service is available
	if !ai.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service is not available")
		return
	}

	// Parse request body
	var req EmbeddingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request
	if len(req.Texts) == 0 {
		WriteBadRequest(w, r, "At least one text is required")
		return
	}

	// Validate texts using service validation
	if err := ai.GetService().ValidateTexts(req.Texts); err != nil {
		switch err {
		case ai.ErrTooManyTexts:
			WriteBadRequest(w, r, "Too many texts for embedding")
		case ai.ErrPromptTooLong:
			WriteBadRequest(w, r, "Text exceeds maximum length")
		default:
			log.Error().Err(err).Msg("unexpected error validating embeddings texts")
			WriteBadRequest(w, r, "Invalid text format")
		}
		return
	}

	// Generate embeddings with timeout
	ctx, cancel := context.WithTimeout(r.Context(), aiEmbeddingsTimeout)
	defer cancel()

	embeddings, err := ai.GetService().GenerateEmbeddings(ctx, req.Texts)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			WriteError(w, r, http.StatusGatewayTimeout, "TIMEOUT", "Embeddings generation timed out")
			return
		}
		log.Error().Err(err).Msg("failed to generate embeddings")
		WriteInternalError(w, r, "Failed to generate embeddings")
		return
	}

	// Get the embedding model name from service
	model := ai.GetService().GetModel()
	if svc, ok := ai.GetService().(interface{ GetEmbeddingModel() string }); ok {
		model = svc.GetEmbeddingModel()
	}

	WriteSuccess(w, "Embeddings generated successfully", map[string]interface{}{
		"embeddings": embeddings,
		"model":      model,
	})
}

// AdvancedChatRequest represents an advanced chat request with function calling and JSON mode
type AdvancedChatRequest struct {
	Messages     []ai.Message             `json:"messages" validate:"required,min=1"`
	SystemPrompt string                   `json:"systemPrompt,omitempty"`
	Temperature  *float32                 `json:"temperature,omitempty"`
	MaxTokens    *int                     `json:"maxTokens,omitempty"`
	TopP         *float32                 `json:"topP,omitempty"`
	TopK         *int                     `json:"topK,omitempty"`
	Functions    []ai.FunctionDeclaration `json:"functions,omitempty"`
	ToolConfig   *ai.ToolConfig           `json:"toolConfig,omitempty"`
	JSONMode     bool                     `json:"jsonMode,omitempty"`
	JSONSchema   *ai.JSONSchema           `json:"jsonSchema,omitempty"`
}

// AIChatAdvanced handles advanced chat requests with function calling and JSON mode
// POST /api/ai/chat/advanced
func AIChatAdvanced(w http.ResponseWriter, r *http.Request) {
	// Check if AI service is available
	if !ai.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service is not available")
		return
	}

	// Parse request body
	var req AdvancedChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		WriteBadRequest(w, r, "At least one message is required")
		return
	}

	// Validate messages using service validation
	if err := ai.GetService().ValidateMessages(req.Messages); err != nil {
		switch err {
		case ai.ErrTooManyMessages:
			WriteBadRequest(w, r, "Too many messages in chat history")
		case ai.ErrPromptTooLong:
			WriteBadRequest(w, r, "Message content exceeds maximum length")
		case ai.ErrInvalidRole:
			WriteBadRequest(w, r, "Invalid message role")
		default:
			log.Error().Err(err).Msg("unexpected error validating advanced chat messages")
			WriteBadRequest(w, r, "Invalid message format")
		}
		return
	}

	// Build advanced chat options
	opts := &ai.ChatOptionsAdvanced{
		ChatOptions: ai.ChatOptions{
			SystemPrompt: req.SystemPrompt,
			GenerateOptions: ai.GenerateOptions{
				Temperature: req.Temperature,
				MaxTokens:   req.MaxTokens,
				TopP:        req.TopP,
				TopK:        req.TopK,
			},
		},
		Functions:  req.Functions,
		ToolConfig: req.ToolConfig,
		JSONMode:   req.JSONMode,
		JSONSchema: req.JSONSchema,
	}

	// Generate response with timeout
	ctx, cancel := context.WithTimeout(r.Context(), aiChatTimeout)
	defer cancel()

	resp, err := ai.GetService().ChatAdvanced(ctx, req.Messages, opts)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			WriteError(w, r, http.StatusGatewayTimeout, "TIMEOUT", "AI request timed out")
			return
		}
		switch err {
		case ai.ErrFunctionNotAllowed:
			WriteError(w, r, http.StatusForbidden, "FUNCTION_CALLING_DISABLED", "Function calling is not enabled")
		case ai.ErrJSONModeNotAllowed:
			WriteError(w, r, http.StatusForbidden, "JSON_MODE_DISABLED", "JSON mode is not enabled")
		case ai.ErrContentBlocked:
			WriteError(w, r, http.StatusUnprocessableEntity, "CONTENT_BLOCKED", "Content was blocked by safety filters")
		default:
			log.Error().Err(err).Msg("failed to generate advanced chat response")
			WriteInternalError(w, r, "Failed to generate response")
		}
		return
	}

	WriteSuccess(w, "Advanced chat response generated", resp)
}
