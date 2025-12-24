import DOMPurify, { type Config } from "dompurify";

/**
 * Sanitization utilities for XSS protection
 *
 * Uses DOMPurify to sanitize user input and prevent XSS attacks.
 * Apply these functions to user-provided content before rendering or storing.
 */

/**
 * Default DOMPurify configuration for text-only content
 * Strips all HTML tags, keeping only plain text
 */
const TEXT_ONLY_CONFIG: Config = {
  ALLOWED_TAGS: [], // No HTML tags allowed
  ALLOWED_ATTR: [], // No attributes allowed
  KEEP_CONTENT: true, // Keep text content from stripped tags
  RETURN_TRUSTED_TYPE: false, // Return string, not TrustedHTML
};

/**
 * Configuration for basic formatted text (bold, italic, links)
 */
const BASIC_HTML_CONFIG: Config = {
  ALLOWED_TAGS: ["b", "i", "em", "strong", "a", "br", "p", "ul", "ol", "li"],
  ALLOWED_ATTR: ["href", "title", "target", "rel"],
  ALLOW_DATA_ATTR: false,
  ADD_ATTR: ["target"], // Allow target for links
  FORBID_TAGS: ["script", "style", "iframe", "form", "input", "object", "embed"],
  FORBID_ATTR: ["onerror", "onload", "onclick", "onmouseover", "onfocus", "onblur"],
  RETURN_TRUSTED_TYPE: false,
};

/**
 * Configuration for rich text content (more formatting options)
 */
const RICH_TEXT_CONFIG: Config = {
  ALLOWED_TAGS: [
    "b",
    "i",
    "em",
    "strong",
    "a",
    "br",
    "p",
    "ul",
    "ol",
    "li",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "blockquote",
    "pre",
    "code",
    "span",
    "div",
    "img",
  ],
  ALLOWED_ATTR: ["href", "title", "target", "rel", "src", "alt", "class", "id"],
  ALLOW_DATA_ATTR: false,
  FORBID_TAGS: ["script", "style", "iframe", "form", "input", "object", "embed", "base"],
  FORBID_ATTR: ["onerror", "onload", "onclick", "onmouseover", "onfocus", "onblur", "style"],
  RETURN_TRUSTED_TYPE: false,
};

/**
 * Sanitize text input by stripping all HTML
 * Use for: usernames, titles, single-line inputs
 *
 * @param input - The string to sanitize
 * @returns Sanitized plain text string
 *
 * @example
 * sanitizeText('<script>alert("xss")</script>Hello')
 * // Returns: 'Hello'
 */
export function sanitizeText(input: string | undefined | null): string {
  if (!input) return "";
  const sanitized = DOMPurify.sanitize(input, TEXT_ONLY_CONFIG);
  return sanitized.trim();
}

/**
 * Sanitize HTML with basic formatting allowed
 * Use for: comments, short descriptions
 *
 * @param input - The HTML string to sanitize
 * @returns Sanitized HTML string with basic formatting
 *
 * @example
 * sanitizeBasicHtml('<b>Hello</b><script>alert("xss")</script>')
 * // Returns: '<b>Hello</b>'
 */
export function sanitizeBasicHtml(input: string | undefined | null): string {
  if (!input) return "";
  return DOMPurify.sanitize(input, BASIC_HTML_CONFIG);
}

/**
 * Sanitize rich HTML content
 * Use for: blog posts, long-form content, rich text editor output
 *
 * @param input - The HTML string to sanitize
 * @returns Sanitized HTML string with rich formatting
 *
 * @example
 * sanitizeRichHtml('<h1>Title</h1><script>alert("xss")</script><p>Content</p>')
 * // Returns: '<h1>Title</h1><p>Content</p>'
 */
export function sanitizeRichHtml(input: string | undefined | null): string {
  if (!input) return "";
  return DOMPurify.sanitize(input, RICH_TEXT_CONFIG);
}

/**
 * Sanitize URL to prevent javascript: protocol attacks
 * Use for: href attributes, external links
 *
 * @param url - The URL to sanitize
 * @returns Sanitized URL or empty string if dangerous
 *
 * @example
 * sanitizeUrl('javascript:alert("xss")')
 * // Returns: ''
 *
 * sanitizeUrl('https://example.com')
 * // Returns: 'https://example.com'
 */
export function sanitizeUrl(url: string | undefined | null): string {
  if (!url) return "";

  // Remove leading/trailing whitespace
  const trimmed = url.trim();

  // Check for dangerous protocols
  const lowerUrl = trimmed.toLowerCase();
  const dangerousProtocols = ["javascript:", "data:", "vbscript:", "file:"];

  for (const protocol of dangerousProtocols) {
    if (lowerUrl.startsWith(protocol)) {
      return "";
    }
  }

  // Allow only http, https, mailto, tel protocols
  const allowedProtocols = ["http://", "https://", "mailto:", "tel:"];
  const hasProtocol = allowedProtocols.some((p) => lowerUrl.startsWith(p));

  // If no protocol, assume relative URL (which is safe)
  // If has protocol, verify it's allowed
  if (lowerUrl.includes("://") && !hasProtocol) {
    return "";
  }

  return trimmed;
}

/**
 * Sanitize email address
 * Basic validation and sanitization
 *
 * @param email - The email to sanitize
 * @returns Sanitized email or empty string if invalid
 */
export function sanitizeEmail(email: string | undefined | null): string {
  if (!email) return "";

  const sanitized = sanitizeText(email).toLowerCase();

  // Basic email format check
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(sanitized)) {
    return "";
  }

  return sanitized;
}

/**
 * Escape HTML entities for safe text display
 * Use when you need to display raw text in HTML context
 *
 * @param input - The string to escape
 * @returns HTML-escaped string
 *
 * @example
 * escapeHtml('<script>alert("xss")</script>')
 * // Returns: '&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;'
 */
export function escapeHtml(input: string | undefined | null): string {
  if (!input) return "";

  const escapeMap: Record<string, string> = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#x27;",
    "/": "&#x2F;",
    "`": "&#x60;",
    "=": "&#x3D;",
  };

  return input.replace(/[&<>"'`=/]/g, (char) => escapeMap[char] || char);
}

/**
 * Sanitize object values recursively
 * Useful for sanitizing form data objects before submission
 *
 * @param obj - The object to sanitize
 * @returns Object with all string values sanitized
 */
export function sanitizeObject<T extends Record<string, unknown>>(obj: T): T {
  const result: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(obj)) {
    if (typeof value === "string") {
      result[key] = sanitizeText(value);
    } else if (Array.isArray(value)) {
      result[key] = value.map((item) =>
        typeof item === "string"
          ? sanitizeText(item)
          : typeof item === "object" && item !== null
            ? sanitizeObject(item as Record<string, unknown>)
            : item
      );
    } else if (typeof value === "object" && value !== null) {
      result[key] = sanitizeObject(value as Record<string, unknown>);
    } else {
      result[key] = value;
    }
  }

  return result as T;
}

/**
 * Strip potentially dangerous characters from filenames
 *
 * @param filename - The filename to sanitize
 * @returns Sanitized filename
 */
export function sanitizeFilename(filename: string | undefined | null): string {
  if (!filename) return "";

  // Remove path traversal attempts
  let sanitized = filename.replace(/\.\./g, "").replace(/[/\\]/g, "");

  // Remove null bytes and control characters
  sanitized = sanitized.replace(/[\x00-\x1f\x80-\x9f]/g, "");

  // Keep only alphanumeric, dots, hyphens, underscores
  sanitized = sanitized.replace(/[^a-zA-Z0-9._-]/g, "_");

  // Prevent empty or dot-only filenames
  if (!sanitized || sanitized === "." || sanitized === "..") {
    return "unnamed";
  }

  return sanitized;
}

// Export DOMPurify for advanced use cases
export { DOMPurify };
