package ai

import (
	"context"
	"strings"

	"google.golang.org/genai"
)

// GenerateText generates text from a simple prompt
func (s *geminiService) GenerateText(ctx context.Context, prompt string, opts *GenerateOptions) (*Response, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, ErrEmptyPrompt
	}

	// Create generation config
	config := s.buildGenerateConfig(opts)

	// Create content from text
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}

	// Generate content
	resp, err := s.client.Models.GenerateContent(ctx, s.config.Model, contents, config)
	if err != nil {
		return nil, err
	}

	return s.parseResponse(resp), nil
}

// Chat handles multi-turn conversations
func (s *geminiService) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*Response, error) {
	if len(messages) == 0 {
		return nil, ErrEmptyPrompt
	}

	// Build generation config
	config := s.buildChatConfig(opts)

	// Convert messages to Gemini format
	contents := s.messagesToContents(messages)

	// Generate content
	resp, err := s.client.Models.GenerateContent(ctx, s.config.Model, contents, config)
	if err != nil {
		return nil, err
	}

	return s.parseResponse(resp), nil
}

// buildGenerateConfig creates a GenerateContentConfig from options
func (s *geminiService) buildGenerateConfig(opts *GenerateOptions) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: int32(s.config.MaxTokens),
	}

	if opts == nil {
		return config
	}

	if opts.MaxTokens != nil {
		config.MaxOutputTokens = int32(*opts.MaxTokens)
	}

	if opts.Temperature != nil {
		config.Temperature = opts.Temperature
	}

	if opts.TopP != nil {
		config.TopP = opts.TopP
	}

	if opts.TopK != nil {
		tk := float32(*opts.TopK)
		config.TopK = &tk
	}

	if len(opts.StopSequence) > 0 {
		config.StopSequences = opts.StopSequence
	}

	return config
}

// buildChatConfig creates a GenerateContentConfig for chat with system prompt support
func (s *geminiService) buildChatConfig(opts *ChatOptions) *genai.GenerateContentConfig {
	var baseOpts *GenerateOptions
	if opts != nil {
		baseOpts = &opts.GenerateOptions
	}

	config := s.buildGenerateConfig(baseOpts)

	// Add system instruction if provided
	if opts != nil && opts.SystemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(opts.SystemPrompt, genai.RoleUser)
	}

	return config
}

// messagesToContents converts our Message types to Gemini Content types
func (s *geminiService) messagesToContents(messages []Message) []*genai.Content {
	contents := make([]*genai.Content, 0, len(messages))

	for _, msg := range messages {
		// Skip system messages as they're handled via SystemInstruction
		if msg.Role == RoleSystem {
			continue
		}

		role := s.convertRole(msg.Role)
		contents = append(contents, genai.NewContentFromText(msg.Content, role))
	}

	return contents
}

// convertRole converts our Role type to Gemini's Role type
func (s *geminiService) convertRole(role Role) genai.Role {
	switch role {
	case RoleUser:
		return genai.RoleUser
	case RoleModel, RoleAssistant:
		return genai.RoleModel
	default:
		return genai.RoleUser
	}
}

// parseResponse extracts our Response from Gemini's GenerateContentResponse
func (s *geminiService) parseResponse(resp *genai.GenerateContentResponse) *Response {
	if resp == nil {
		return &Response{
			Content: "",
			Model:   s.config.Model,
		}
	}

	// Extract text from response
	var content string
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				content += part.Text
			}
		}
	}

	// Extract usage metadata
	var usage *Usage
	if resp.UsageMetadata != nil {
		usage = &Usage{
			InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
			OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:  int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return &Response{
		Content: content,
		Model:   s.config.Model,
		Usage:   usage,
	}
}
