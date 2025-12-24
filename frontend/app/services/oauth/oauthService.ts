import { logger } from "../../lib/logger";
import { API_BASE_URL, authenticatedFetch } from "../api/client";

export type OAuthProvider = "google" | "github";

export interface OAuthURLResponse {
  url: string;
  state: string;
}

export interface LinkedProvider {
  provider: OAuthProvider;
  email: string;
  linked_at: string;
}

export interface LinkedProvidersResponse {
  providers: LinkedProvider[];
}

/**
 * OAuth service for social login integration
 */
export const OAuthService = {
  /**
   * Get the OAuth authorization URL for a provider
   */
  async getOAuthURL(provider: OAuthProvider): Promise<OAuthURLResponse> {
    const response = await fetch(`${API_BASE_URL}/api/auth/oauth/${provider}`, {
      method: "GET",
      credentials: "include",
    });

    if (!response.ok) {
      const error = await response.text();
      logger.error(`Failed to get OAuth URL for ${provider}`, { error });
      throw new Error(`Failed to initialize ${provider} login`);
    }

    return response.json();
  },

  /**
   * Redirect user to OAuth provider for authentication
   */
  async initiateOAuth(provider: OAuthProvider): Promise<void> {
    try {
      const { url } = await this.getOAuthURL(provider);
      // Redirect to OAuth provider
      window.location.href = url;
    } catch (error) {
      logger.error(`Failed to initiate OAuth with ${provider}`, error);
      throw error;
    }
  },

  /**
   * Get linked OAuth providers for the current user
   */
  async getLinkedProviders(): Promise<LinkedProvider[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/auth/oauth/providers`);

    if (!response.ok) {
      if (response.status === 401) {
        return [];
      }
      throw new Error("Failed to get linked providers");
    }

    const data: LinkedProvidersResponse = await response.json();
    return data.providers || [];
  },

  /**
   * Unlink an OAuth provider from the current user
   */
  async unlinkProvider(provider: OAuthProvider): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/auth/oauth/${provider}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      const error = await response.text();
      logger.error(`Failed to unlink ${provider}`, { error });
      throw new Error(`Failed to unlink ${provider}`);
    }
  },

  /**
   * Check if a specific provider is configured on the backend
   */
  async isProviderConfigured(provider: OAuthProvider): Promise<boolean> {
    try {
      await this.getOAuthURL(provider);
      return true;
    } catch {
      return false;
    }
  },
};
