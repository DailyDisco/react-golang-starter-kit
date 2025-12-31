package ai

import (
	"context"
	"strings"

	"google.golang.org/genai"
)

// GenerateEmbedding generates an embedding for a single text
func (s *geminiService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyPrompt
	}

	embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, ErrEmptyPrompt
	}

	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (s *geminiService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, ErrEmptyTexts
	}

	// Filter out empty texts
	validTexts := make([]string, 0, len(texts))
	for _, t := range texts {
		if strings.TrimSpace(t) != "" {
			validTexts = append(validTexts, t)
		}
	}

	if len(validTexts) == 0 {
		return nil, ErrEmptyTexts
	}

	// Build content for embedding
	contents := make([]*genai.Content, len(validTexts))
	for i, text := range validTexts {
		contents[i] = &genai.Content{
			Parts: []*genai.Part{
				{Text: text},
			},
		}
	}

	// Generate embeddings
	resp, err := s.client.Models.EmbedContent(ctx, s.config.EmbeddingModel, contents, nil)
	if err != nil {
		return nil, err
	}

	// Extract embeddings from response
	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		embeddings[i] = emb.Values
	}

	return embeddings, nil
}
