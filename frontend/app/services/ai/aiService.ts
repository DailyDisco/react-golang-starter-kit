import { API_BASE_URL, authenticatedFetch, parseErrorResponse } from "../api/client";

// Types
export interface ChatMessage {
  role: "user" | "model" | "system" | "assistant";
  content: string;
}

export interface ChatOptions {
  systemPrompt?: string;
  temperature?: number;
  maxTokens?: number;
  topP?: number;
  topK?: number;
}

export interface ChatResponse {
  content: string;
  model: string;
  usage?: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
}

export interface StreamChunk {
  token?: string;
  done?: boolean;
  error?: string;
}

export interface ImageInput {
  data?: string; // Base64 encoded
  mimeType?: string;
  url?: string;
}

export interface AnalyzeImageRequest {
  image: ImageInput;
  prompt: string;
}

export interface EmbeddingsRequest {
  texts: string[];
}

export interface EmbeddingsResponse {
  embeddings: number[][];
  model: string;
}

export interface FunctionDeclaration {
  name: string;
  description: string;
  parameters?: Record<string, unknown>;
}

export interface AdvancedChatOptions extends ChatOptions {
  functions?: FunctionDeclaration[];
  toolConfig?: {
    functionCallingMode?: "auto" | "any" | "none";
  };
  jsonMode?: boolean;
  jsonSchema?: {
    type: string;
    properties?: Record<string, unknown>;
    required?: string[];
    description?: string;
  };
}

export interface FunctionCall {
  name: string;
  args: Record<string, unknown>;
}

export interface AdvancedChatResponse extends ChatResponse {
  functionCalls?: FunctionCall[];
}

export class AIService {
  /**
   * Send a chat message and get a response
   */
  static async chat(messages: ChatMessage[], options?: ChatOptions): Promise<ChatResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/ai/chat`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages,
        ...options,
      }),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to send chat message");
    }

    const data = await response.json();
    return data.data;
  }

  /**
   * Send a chat message with streaming response
   * Returns an async generator that yields tokens
   */
  static async *chatStream(messages: ChatMessage[], options?: ChatOptions): AsyncGenerator<StreamChunk, void, unknown> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/ai/chat/stream`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages,
        ...options,
      }),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to start streaming chat");
    }

    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error("No response body");
    }

    const decoder = new TextDecoder();
    let buffer = "";

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const data = line.slice(6);
            if (data === "[DONE]") {
              yield { done: true };
              return;
            }
            try {
              const parsed = JSON.parse(data);
              yield parsed;
            } catch {
              // Skip invalid JSON
            }
          } else if (line.startsWith("event: error")) {
            // Next line contains error data
            continue;
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Advanced chat with function calling and JSON mode support
   */
  static async chatAdvanced(messages: ChatMessage[], options?: AdvancedChatOptions): Promise<AdvancedChatResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/ai/chat/advanced`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages,
        ...options,
      }),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to send advanced chat message");
    }

    const data = await response.json();
    return data.data;
  }

  /**
   * Analyze an image with a prompt
   */
  static async analyzeImage(request: AnalyzeImageRequest): Promise<ChatResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/ai/analyze-image`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to analyze image");
    }

    const data = await response.json();
    return data.data;
  }

  /**
   * Generate embeddings for texts
   */
  static async generateEmbeddings(texts: string[]): Promise<EmbeddingsResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/ai/embeddings`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ texts }),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to generate embeddings");
    }

    const data = await response.json();
    return data.data;
  }

  /**
   * Convert a File to base64 for image analysis
   */
  static async fileToBase64(file: File): Promise<ImageInput> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result as string;
        // Remove data URL prefix (e.g., "data:image/png;base64,")
        const base64 = result.split(",")[1];
        resolve({
          data: base64,
          mimeType: file.type,
        });
      };
      reader.onerror = reject;
      reader.readAsDataURL(file);
    });
  }
}
