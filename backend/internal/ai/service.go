package ai

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

// Service defines the AI service interface
type Service interface {
	// Chat operations
	GenerateText(ctx context.Context, prompt string, opts *GenerateOptions) (*Response, error)
	Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*Response, error)
	StreamChat(ctx context.Context, messages []Message, opts *ChatOptions) (<-chan StreamChunk, error)

	// Advanced chat with function calling and JSON mode
	ChatAdvanced(ctx context.Context, messages []Message, opts *ChatOptionsAdvanced) (*AdvancedResponse, error)

	// Multi-modal operations
	AnalyzeImage(ctx context.Context, image ImageInput, prompt string) (*Response, error)
	GenerateWithImages(ctx context.Context, prompt string, images []ImageInput, opts *GenerateOptions) (*Response, error)

	// Embedding operations
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// Validation helpers
	ValidatePrompt(prompt string) error
	ValidateMessages(messages []Message) error
	ValidateTexts(texts []string) error

	// Configuration
	GetModel() string
	GetConfig() *Config
	IsAvailable() bool
}

// geminiService implements the Service interface using Google's Gemini API
type geminiService struct {
	config *Config
	client *genai.Client
}

var (
	instance Service
	once     sync.Once
	mu       sync.RWMutex
)

// Initialize sets up the AI service
func Initialize(config *Config) error {
	var initErr error

	once.Do(func() {
		if err := config.Validate(); err != nil {
			initErr = err
			return
		}

		if !config.Enabled {
			log.Info().Msg("ai service disabled")
			instance = &noOpService{}
			return
		}

		// Create Gemini client
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  config.APIKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			initErr = err
			return
		}

		instance = &geminiService{
			config: config,
			client: client,
		}

		log.Info().
			Str("model", config.Model).
			Str("embedding_model", config.EmbeddingModel).
			Msg("ai service initialized")
	})

	return initErr
}

// GetService returns the global AI service instance
func GetService() Service {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// IsAvailable returns true if the AI service is initialized and available
func IsAvailable() bool {
	svc := GetService()
	return svc != nil && svc.IsAvailable()
}

// GetModel returns the model name
func (s *geminiService) GetModel() string {
	return s.config.Model
}

// GetConfig returns the service configuration
func (s *geminiService) GetConfig() *Config {
	return s.config
}

// IsAvailable returns true if the service is enabled
func (s *geminiService) IsAvailable() bool {
	return s.config.Enabled
}

// ValidatePrompt checks if a prompt meets length requirements
func (s *geminiService) ValidatePrompt(prompt string) error {
	if len(prompt) > s.config.MaxPromptLength {
		return ErrPromptTooLong
	}
	return nil
}

// ValidateMessages checks if messages meet requirements
func (s *geminiService) ValidateMessages(messages []Message) error {
	if len(messages) > s.config.MaxMessagesPerChat {
		return ErrTooManyMessages
	}
	totalLength := 0
	for _, msg := range messages {
		totalLength += len(msg.Content)
	}
	if totalLength > s.config.MaxPromptLength {
		return ErrPromptTooLong
	}
	return nil
}

// ValidateTexts checks if texts meet embedding requirements
func (s *geminiService) ValidateTexts(texts []string) error {
	if len(texts) > s.config.MaxTextsPerEmbed {
		return ErrTooManyTexts
	}
	return nil
}

// noOpService is a no-op implementation when AI is disabled
type noOpService struct{}

func (n *noOpService) GenerateText(ctx context.Context, prompt string, opts *GenerateOptions) (*Response, error) {
	return nil, ErrDisabled
}

func (n *noOpService) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*Response, error) {
	return nil, ErrDisabled
}

func (n *noOpService) StreamChat(ctx context.Context, messages []Message, opts *ChatOptions) (<-chan StreamChunk, error) {
	return nil, ErrDisabled
}

func (n *noOpService) ChatAdvanced(ctx context.Context, messages []Message, opts *ChatOptionsAdvanced) (*AdvancedResponse, error) {
	return nil, ErrDisabled
}

func (n *noOpService) AnalyzeImage(ctx context.Context, image ImageInput, prompt string) (*Response, error) {
	return nil, ErrDisabled
}

func (n *noOpService) GenerateWithImages(ctx context.Context, prompt string, images []ImageInput, opts *GenerateOptions) (*Response, error) {
	return nil, ErrDisabled
}

func (n *noOpService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, ErrDisabled
}

func (n *noOpService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, ErrDisabled
}

func (n *noOpService) ValidatePrompt(prompt string) error {
	return ErrDisabled
}

func (n *noOpService) ValidateMessages(messages []Message) error {
	return ErrDisabled
}

func (n *noOpService) ValidateTexts(texts []string) error {
	return ErrDisabled
}

func (n *noOpService) GetModel() string {
	return ""
}

func (n *noOpService) GetConfig() *Config {
	return nil
}

func (n *noOpService) IsAvailable() bool {
	return false
}
