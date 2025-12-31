package ai

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/genai"
)

// AnalyzeImage analyzes an image with a prompt
func (s *geminiService) AnalyzeImage(ctx context.Context, image ImageInput, prompt string) (*Response, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, ErrEmptyPrompt
	}

	// Validate and prepare image
	imgPart, err := s.prepareImagePart(image)
	if err != nil {
		return nil, err
	}

	// Build parts: image + text prompt
	parts := []*genai.Part{
		imgPart,
		genai.NewPartFromText(prompt),
	}

	// Create content
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	// Generate config
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: int32(s.config.MaxTokens),
	}

	// Generate content
	resp, err := s.client.Models.GenerateContent(ctx, s.config.Model, contents, config)
	if err != nil {
		return nil, err
	}

	return s.parseResponse(resp), nil
}

// GenerateWithImages generates content from a prompt and multiple images
func (s *geminiService) GenerateWithImages(ctx context.Context, prompt string, images []ImageInput, opts *GenerateOptions) (*Response, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, ErrEmptyPrompt
	}

	// Build parts starting with images
	parts := make([]*genai.Part, 0, len(images)+1)

	for _, img := range images {
		imgPart, err := s.prepareImagePart(img)
		if err != nil {
			return nil, err
		}
		parts = append(parts, imgPart)
	}

	// Add text prompt
	parts = append(parts, genai.NewPartFromText(prompt))

	// Create content
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	// Generate config
	config := s.buildGenerateConfig(opts)

	// Generate content
	resp, err := s.client.Models.GenerateContent(ctx, s.config.Model, contents, config)
	if err != nil {
		return nil, err
	}

	return s.parseResponse(resp), nil
}

// prepareImagePart creates a Gemini Part from an ImageInput
func (s *geminiService) prepareImagePart(image ImageInput) (*genai.Part, error) {
	// Handle URL-based images
	if image.URL != "" {
		return genai.NewPartFromURI(image.URL, s.detectMimeType(image)), nil
	}

	// Handle base64-encoded images
	if image.Data == "" {
		return nil, ErrInvalidImage
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(image.Data)
	if err != nil {
		// Try with padding variations
		data, err = base64.RawStdEncoding.DecodeString(image.Data)
		if err != nil {
			return nil, ErrInvalidImage
		}
	}

	// Check size
	if int64(len(data)) > s.config.MaxImageSize {
		return nil, ErrImageTooLarge
	}

	// Create inline data part
	return genai.NewPartFromBytes(data, s.detectMimeType(image)), nil
}

// detectMimeType determines the MIME type of an image
func (s *geminiService) detectMimeType(image ImageInput) string {
	if image.MimeType != "" {
		return image.MimeType
	}

	// Try to detect from URL extension
	if image.URL != "" {
		lower := strings.ToLower(image.URL)
		switch {
		case strings.HasSuffix(lower, ".png"):
			return "image/png"
		case strings.HasSuffix(lower, ".gif"):
			return "image/gif"
		case strings.HasSuffix(lower, ".webp"):
			return "image/webp"
		default:
			return "image/jpeg"
		}
	}

	// Default to JPEG
	return "image/jpeg"
}
