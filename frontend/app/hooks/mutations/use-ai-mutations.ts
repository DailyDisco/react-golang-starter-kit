import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import { logger } from "../../lib/logger";
import {
  AIService,
  type AdvancedChatOptions,
  type AnalyzeImageRequest,
  type ChatMessage,
  type ChatOptions,
} from "../../services/ai/aiService";
import { ApiError } from "../../services/api/client";
import { useAuthStore } from "../../stores/auth-store";

// Helper to get error code from ApiError or fallback to checking message
function getErrorCode(error: Error): string {
  if (error instanceof ApiError) {
    return error.code;
  }
  return "UNKNOWN_ERROR";
}

// Common error messages used across AI mutations
const COMMON_AI_ERRORS: Record<string, { title: string; description: string }> = {
  SERVICE_UNAVAILABLE: {
    title: "AI service unavailable",
    description: "Please configure your Gemini API key in settings",
  },
  RATE_LIMITED: {
    title: "Rate limit exceeded",
    description: "Please wait a moment before trying again",
  },
  TIMEOUT: {
    title: "Request timed out",
    description: "The AI service took too long to respond",
  },
};

interface AIErrorConfig {
  logContext: string;
  defaultTitle: string;
  customErrors?: Record<string, { title: string; description: string }>;
}

/**
 * Creates an error handler for AI mutations with common error handling logic.
 */
function createAIErrorHandler(config: AIErrorConfig) {
  const { logContext, defaultTitle, customErrors = {} } = config;
  const errorMap = { ...COMMON_AI_ERRORS, ...customErrors };

  return (error: Error) => {
    logger.error(logContext, error);
    const code = getErrorCode(error);
    const errorInfo = errorMap[code];

    if (errorInfo) {
      toast.error(errorInfo.title, { description: errorInfo.description });
    } else {
      toast.error(defaultTitle, {
        description: error.message || "An unexpected error occurred",
      });
    }
  };
}

export function useAIChat() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async ({ messages, options }: { messages: ChatMessage[]; options?: ChatOptions }) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return AIService.chat(messages, options);
    },
    onError: createAIErrorHandler({
      logContext: "AI chat error",
      defaultTitle: "AI request failed",
      customErrors: {
        CONTENT_BLOCKED: {
          title: "Content blocked",
          description: "Your message was blocked by safety filters",
        },
      },
    }),
  });
}

export function useAIChatAdvanced() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async ({ messages, options }: { messages: ChatMessage[]; options?: AdvancedChatOptions }) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return AIService.chatAdvanced(messages, options);
    },
    onError: createAIErrorHandler({
      logContext: "AI advanced chat error",
      defaultTitle: "AI request failed",
      customErrors: {
        FUNCTION_CALLING_DISABLED: {
          title: "Function calling disabled",
          description: "This feature is not enabled on the server",
        },
        JSON_MODE_DISABLED: {
          title: "JSON mode disabled",
          description: "This feature is not enabled on the server",
        },
      },
    }),
  });
}

export function useAIAnalyzeImage() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async (request: AnalyzeImageRequest) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return AIService.analyzeImage(request);
    },
    onError: createAIErrorHandler({
      logContext: "AI image analysis error",
      defaultTitle: "Image analysis failed",
      customErrors: {
        IMAGE_TOO_LARGE: {
          title: "Image too large",
          description: "Please select a smaller image (max 10MB)",
        },
        INVALID_IMAGE: {
          title: "Invalid image",
          description: "Please select a valid image file",
        },
      },
    }),
  });
}

export function useAIEmbeddings() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async (texts: string[]) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return AIService.generateEmbeddings(texts);
    },
    onError: createAIErrorHandler({
      logContext: "AI embeddings error",
      defaultTitle: "Embedding generation failed",
      customErrors: {
        TOO_MANY_TEXTS: {
          title: "Too many texts",
          description: "Please reduce the number of texts (max 100)",
        },
      },
    }),
  });
}
