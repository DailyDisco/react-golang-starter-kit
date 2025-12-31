package ai

import "errors"

// Common errors
var (
	ErrMissingAPIKey = errors.New("ai: missing API key")
	ErrDisabled      = errors.New("ai: service is disabled")
	ErrInvalidRole   = errors.New("ai: invalid message role")
	ErrEmptyPrompt   = errors.New("ai: empty prompt")
	ErrImageTooLarge = errors.New("ai: image exceeds maximum size")
	ErrInvalidImage  = errors.New("ai: invalid image data")
	ErrEmptyTexts    = errors.New("ai: empty texts for embedding")

	// Validation errors
	ErrPromptTooLong      = errors.New("ai: prompt exceeds maximum length")
	ErrTooManyMessages    = errors.New("ai: too many messages in chat")
	ErrTooManyTexts       = errors.New("ai: too many texts for embedding")
	ErrContentBlocked     = errors.New("ai: content blocked by safety filters")
	ErrFunctionNotAllowed = errors.New("ai: function calling is not enabled")
	ErrJSONModeNotAllowed = errors.New("ai: JSON mode is not enabled")
)

// Role represents a message role in a conversation
type Role string

const (
	RoleUser      Role = "user"
	RoleModel     Role = "model"
	RoleSystem    Role = "system"
	RoleAssistant Role = "assistant" // Alias for model
)

// Message represents a single message in a conversation
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ImageInput represents an image for multi-modal requests
type ImageInput struct {
	// Base64 encoded image data
	Data string `json:"data,omitempty"`
	// MIME type (e.g., "image/jpeg", "image/png")
	MimeType string `json:"mimeType,omitempty"`
	// URL of the image (alternative to Data)
	URL string `json:"url,omitempty"`
}

// GenerateOptions contains options for text generation
type GenerateOptions struct {
	Temperature  *float32 `json:"temperature,omitempty"`
	MaxTokens    *int     `json:"maxTokens,omitempty"`
	TopP         *float32 `json:"topP,omitempty"`
	TopK         *int     `json:"topK,omitempty"`
	StopSequence []string `json:"stopSequence,omitempty"`
}

// ChatOptions contains options for chat conversations
type ChatOptions struct {
	SystemPrompt string `json:"systemPrompt,omitempty"`
	GenerateOptions
}

// Response represents an AI generation response
type Response struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Usage   *Usage `json:"usage,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// StreamChunk represents a single chunk in a streaming response
type StreamChunk struct {
	Token string `json:"token,omitempty"`
	Done  bool   `json:"done,omitempty"`
	Error error  `json:"-"`
}

// EmbeddingResponse represents the response from an embedding request
type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Model      string      `json:"model"`
}

// FunctionDeclaration defines a function that the model can call
type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"` // JSON Schema
}

// FunctionCall represents a function call made by the model
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// FunctionResponse represents the result of a function call
type FunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// ToolConfig configures how the model uses tools
type ToolConfig struct {
	// FunctionCallingMode controls when functions can be called
	// "auto" - model decides (default)
	// "any" - model must call a function
	// "none" - model cannot call functions
	FunctionCallingMode string `json:"functionCallingMode,omitempty"`
}

// JSONSchema defines the expected structure for JSON mode output
type JSONSchema struct {
	Type        string                 `json:"type"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ChatOptionsAdvanced extends ChatOptions with function calling and JSON mode
type ChatOptionsAdvanced struct {
	ChatOptions

	// Function calling
	Functions  []FunctionDeclaration `json:"functions,omitempty"`
	ToolConfig *ToolConfig           `json:"toolConfig,omitempty"`

	// JSON mode - forces structured output
	JSONMode   bool        `json:"jsonMode,omitempty"`
	JSONSchema *JSONSchema `json:"jsonSchema,omitempty"`
}

// AdvancedResponse includes function calls in addition to text
type AdvancedResponse struct {
	Response
	FunctionCalls []FunctionCall `json:"functionCalls,omitempty"`
}
