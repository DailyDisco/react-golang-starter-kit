package ai

import (
	"context"
	"encoding/json"

	"google.golang.org/genai"
)

// ChatAdvanced handles chat with function calling and JSON mode support
func (s *geminiService) ChatAdvanced(ctx context.Context, messages []Message, opts *ChatOptionsAdvanced) (*AdvancedResponse, error) {
	if len(messages) == 0 {
		return nil, ErrEmptyPrompt
	}

	// Validate messages
	if err := s.ValidateMessages(messages); err != nil {
		return nil, err
	}

	// Check feature flags
	if opts != nil {
		if len(opts.Functions) > 0 && !s.config.AllowFunctionCalling {
			return nil, ErrFunctionNotAllowed
		}
		if opts.JSONMode && !s.config.AllowJSONMode {
			return nil, ErrJSONModeNotAllowed
		}
	}

	// Build generation config with advanced options
	config := s.buildAdvancedConfig(opts)

	// Convert messages to Gemini format
	contents := s.messagesToContents(messages)

	// Generate content
	resp, err := s.client.Models.GenerateContent(ctx, s.config.Model, contents, config)
	if err != nil {
		return nil, err
	}

	return s.parseAdvancedResponse(resp), nil
}

// buildAdvancedConfig creates a GenerateContentConfig with all advanced options
func (s *geminiService) buildAdvancedConfig(opts *ChatOptionsAdvanced) *genai.GenerateContentConfig {
	// Start with base chat config
	var baseOpts *ChatOptions
	if opts != nil {
		baseOpts = &opts.ChatOptions
	}
	config := s.buildChatConfig(baseOpts)

	// Apply safety settings
	config.SafetySettings = s.buildSafetySettings()

	if opts == nil {
		return config
	}

	// Add function declarations if provided
	if len(opts.Functions) > 0 {
		tools := make([]*genai.Tool, 0, 1)
		funcDecls := make([]*genai.FunctionDeclaration, 0, len(opts.Functions))

		for _, fn := range opts.Functions {
			funcDecls = append(funcDecls, &genai.FunctionDeclaration{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  convertToSchema(fn.Parameters),
			})
		}

		tools = append(tools, &genai.Tool{
			FunctionDeclarations: funcDecls,
		})
		config.Tools = tools

		// Apply tool config if specified
		if opts.ToolConfig != nil {
			config.ToolConfig = &genai.ToolConfig{
				FunctionCallingConfig: &genai.FunctionCallingConfig{
					Mode: convertFunctionCallingMode(opts.ToolConfig.FunctionCallingMode),
				},
			}
		}
	}

	// Enable JSON mode if requested
	if opts.JSONMode {
		config.ResponseMIMEType = "application/json"

		// If schema provided, use it
		if opts.JSONSchema != nil {
			config.ResponseSchema = convertToSchema(map[string]interface{}{
				"type":        opts.JSONSchema.Type,
				"properties":  opts.JSONSchema.Properties,
				"required":    opts.JSONSchema.Required,
				"description": opts.JSONSchema.Description,
			})
		}
	}

	return config
}

// buildSafetySettings creates safety settings based on configuration
func (s *geminiService) buildSafetySettings() []*genai.SafetySetting {
	threshold := convertSafetyThreshold(s.config.SafetyLevel)

	// Apply same threshold to all harm categories
	return []*genai.SafetySetting{
		{Category: genai.HarmCategoryHateSpeech, Threshold: threshold},
		{Category: genai.HarmCategoryDangerousContent, Threshold: threshold},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: threshold},
		{Category: genai.HarmCategoryHarassment, Threshold: threshold},
	}
}

// convertSafetyThreshold converts our SafetyLevel to Gemini's HarmBlockThreshold
func convertSafetyThreshold(level SafetyLevel) genai.HarmBlockThreshold {
	switch level {
	case SafetyLevelNone:
		return genai.HarmBlockThresholdBlockNone
	case SafetyLevelLow:
		return genai.HarmBlockThresholdBlockOnlyHigh
	case SafetyLevelMedium:
		return genai.HarmBlockThresholdBlockMediumAndAbove
	case SafetyLevelHigh:
		return genai.HarmBlockThresholdBlockLowAndAbove
	case SafetyLevelBlockAll:
		return genai.HarmBlockThresholdBlockLowAndAbove // Most restrictive available
	default:
		return genai.HarmBlockThresholdBlockMediumAndAbove
	}
}

// convertFunctionCallingMode converts string to Gemini's FunctionCallingConfigMode
func convertFunctionCallingMode(mode string) genai.FunctionCallingConfigMode {
	switch mode {
	case "any":
		return genai.FunctionCallingConfigModeAny
	case "none":
		return genai.FunctionCallingConfigModeNone
	default:
		return genai.FunctionCallingConfigModeAuto
	}
}

// convertToSchema converts a map to Gemini's Schema type
func convertToSchema(params map[string]interface{}) *genai.Schema {
	if params == nil {
		return nil
	}

	// Convert to JSON and back to handle nested structures
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return nil
	}

	var schema genai.Schema
	if err := json.Unmarshal(jsonBytes, &schema); err != nil {
		return nil
	}

	return &schema
}

// parseAdvancedResponse extracts AdvancedResponse from Gemini's response
func (s *geminiService) parseAdvancedResponse(resp *genai.GenerateContentResponse) *AdvancedResponse {
	if resp == nil {
		return &AdvancedResponse{
			Response: Response{
				Content: "",
				Model:   s.config.Model,
			},
		}
	}

	result := &AdvancedResponse{
		Response: *s.parseResponse(resp),
	}

	// Extract function calls if present
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				result.FunctionCalls = append(result.FunctionCalls, FunctionCall{
					Name: part.FunctionCall.Name,
					Args: part.FunctionCall.Args,
				})
			}
		}
	}

	return result
}
