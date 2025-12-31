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
import { useAuthStore } from "../../stores/auth-store";

export function useAIChat() {
  const { isAuthenticated } = useAuthStore();

  return useMutation({
    mutationFn: async ({ messages, options }: { messages: ChatMessage[]; options?: ChatOptions }) => {
      if (!isAuthenticated) {
        throw new Error("Authentication required.");
      }
      return AIService.chat(messages, options);
    },
    onError: (error: Error) => {
      logger.error("AI chat error", error);

      if (error.message.includes("not available")) {
        toast.error("AI service unavailable", {
          description: "Please configure your Gemini API key in settings",
        });
      } else if (error.message.includes("rate limit") || error.message.includes("429")) {
        toast.error("Rate limit exceeded", {
          description: "Please wait a moment before trying again",
        });
      } else if (error.message.includes("blocked")) {
        toast.error("Content blocked", {
          description: "Your message was blocked by safety filters",
        });
      } else {
        toast.error("AI request failed", {
          description: error.message || "An unexpected error occurred",
        });
      }
    },
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
    onError: (error: Error) => {
      logger.error("AI advanced chat error", error);

      if (error.message.includes("FUNCTION_CALLING_DISABLED")) {
        toast.error("Function calling disabled", {
          description: "This feature is not enabled on the server",
        });
      } else if (error.message.includes("JSON_MODE_DISABLED")) {
        toast.error("JSON mode disabled", {
          description: "This feature is not enabled on the server",
        });
      } else {
        toast.error("AI request failed", {
          description: error.message || "An unexpected error occurred",
        });
      }
    },
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
    onError: (error: Error) => {
      logger.error("AI image analysis error", error);

      if (error.message.includes("too large")) {
        toast.error("Image too large", {
          description: "Please select a smaller image (max 10MB)",
        });
      } else if (error.message.includes("invalid image")) {
        toast.error("Invalid image", {
          description: "Please select a valid image file",
        });
      } else {
        toast.error("Image analysis failed", {
          description: error.message || "An unexpected error occurred",
        });
      }
    },
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
    onError: (error: Error) => {
      logger.error("AI embeddings error", error);

      if (error.message.includes("too many texts")) {
        toast.error("Too many texts", {
          description: "Please reduce the number of texts (max 100)",
        });
      } else {
        toast.error("Embedding generation failed", {
          description: error.message || "An unexpected error occurred",
        });
      }
    },
  });
}
