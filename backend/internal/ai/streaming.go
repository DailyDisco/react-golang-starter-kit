package ai

import (
	"context"

	"google.golang.org/genai"
)

// StreamChat handles streaming chat responses
func (s *geminiService) StreamChat(ctx context.Context, messages []Message, opts *ChatOptions) (<-chan StreamChunk, error) {
	if len(messages) == 0 {
		return nil, ErrEmptyPrompt
	}

	// Build generation config
	config := s.buildChatConfig(opts)

	// Convert messages to Gemini format
	contents := s.messagesToContents(messages)

	// Create output channel
	chunks := make(chan StreamChunk, 100)

	// Start streaming in goroutine
	go func() {
		defer close(chunks)

		// Stream content using Go 1.23+ range-over-func
		// GenerateContentStream returns iter.Seq2[*GenerateContentResponse, error]
		for resp, err := range s.client.Models.GenerateContentStream(ctx, s.config.Model, contents, config) {
			if err != nil {
				chunks <- StreamChunk{Error: err}
				return
			}

			// Extract text from response chunk
			text := s.extractTextFromResponse(resp)
			if text != "" {
				chunks <- StreamChunk{Token: text}
			}
		}

		// Stream completed
		chunks <- StreamChunk{Done: true}
	}()

	return chunks, nil
}

// extractTextFromResponse extracts text content from a GenerateContentResponse
func (s *geminiService) extractTextFromResponse(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return ""
	}

	var text string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			text += part.Text
		}
	}

	return text
}
