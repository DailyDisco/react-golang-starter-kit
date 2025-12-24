import { describe, expect, it } from "vitest";

import {
  escapeHtml,
  sanitizeBasicHtml,
  sanitizeEmail,
  sanitizeFilename,
  sanitizeObject,
  sanitizeRichHtml,
  sanitizeText,
  sanitizeUrl,
} from "./sanitize";

describe("sanitizeText", () => {
  it("should strip all HTML tags", () => {
    expect(sanitizeText("<script>alert('xss')</script>Hello")).toBe("Hello");
    expect(sanitizeText("<b>Bold</b> text")).toBe("Bold text");
    expect(sanitizeText("<img src=x onerror=alert(1)>")).toBe("");
  });

  it("should handle empty and null inputs", () => {
    expect(sanitizeText("")).toBe("");
    expect(sanitizeText(null)).toBe("");
    expect(sanitizeText(undefined)).toBe("");
  });

  it("should trim whitespace", () => {
    expect(sanitizeText("  hello  ")).toBe("hello");
  });

  it("should preserve normal text", () => {
    expect(sanitizeText("Hello World")).toBe("Hello World");
    expect(sanitizeText("user@example.com")).toBe("user@example.com");
  });
});

describe("sanitizeBasicHtml", () => {
  it("should allow basic formatting tags", () => {
    expect(sanitizeBasicHtml("<b>Bold</b>")).toBe("<b>Bold</b>");
    expect(sanitizeBasicHtml("<i>Italic</i>")).toBe("<i>Italic</i>");
    expect(sanitizeBasicHtml("<a href='https://example.com'>Link</a>")).toContain("<a");
  });

  it("should strip dangerous tags", () => {
    expect(sanitizeBasicHtml("<script>alert('xss')</script>")).toBe("");
    expect(sanitizeBasicHtml("<iframe src='evil.com'></iframe>")).toBe("");
    // Form and input tags are stripped, content is preserved
    expect(sanitizeBasicHtml("<form>content</form>")).not.toContain("<form");
  });

  it("should strip event handlers", () => {
    expect(sanitizeBasicHtml('<img onerror="alert(1)" src="x">')).not.toContain("onerror");
    expect(sanitizeBasicHtml('<div onclick="evil()">text</div>')).not.toContain("onclick");
  });
});

describe("sanitizeRichHtml", () => {
  it("should allow heading tags", () => {
    expect(sanitizeRichHtml("<h1>Title</h1>")).toBe("<h1>Title</h1>");
    expect(sanitizeRichHtml("<h2>Subtitle</h2>")).toBe("<h2>Subtitle</h2>");
  });

  it("should allow code blocks", () => {
    expect(sanitizeRichHtml("<pre><code>const x = 1;</code></pre>")).toContain("<code>");
  });

  it("should strip dangerous content", () => {
    expect(sanitizeRichHtml("<script>alert('xss')</script>")).toBe("");
    expect(sanitizeRichHtml("<style>body{display:none}</style>")).toBe("");
  });
});

describe("sanitizeUrl", () => {
  it("should allow safe URLs", () => {
    expect(sanitizeUrl("https://example.com")).toBe("https://example.com");
    expect(sanitizeUrl("http://example.com")).toBe("http://example.com");
    expect(sanitizeUrl("mailto:test@example.com")).toBe("mailto:test@example.com");
    expect(sanitizeUrl("/relative/path")).toBe("/relative/path");
  });

  it("should block dangerous protocols", () => {
    expect(sanitizeUrl("javascript:alert('xss')")).toBe("");
    expect(sanitizeUrl("data:text/html,<script>alert(1)</script>")).toBe("");
    expect(sanitizeUrl("vbscript:alert('xss')")).toBe("");
  });

  it("should handle case variations", () => {
    expect(sanitizeUrl("JAVASCRIPT:alert('xss')")).toBe("");
    expect(sanitizeUrl("JaVaScRiPt:alert('xss')")).toBe("");
  });

  it("should handle empty inputs", () => {
    expect(sanitizeUrl("")).toBe("");
    expect(sanitizeUrl(null)).toBe("");
    expect(sanitizeUrl(undefined)).toBe("");
  });

  it("should trim whitespace", () => {
    expect(sanitizeUrl("  https://example.com  ")).toBe("https://example.com");
  });
});

describe("sanitizeEmail", () => {
  it("should accept valid emails", () => {
    expect(sanitizeEmail("user@example.com")).toBe("user@example.com");
    expect(sanitizeEmail("USER@EXAMPLE.COM")).toBe("user@example.com");
  });

  it("should reject invalid emails", () => {
    expect(sanitizeEmail("not-an-email")).toBe("");
    expect(sanitizeEmail("@example.com")).toBe("");
    expect(sanitizeEmail("user@")).toBe("");
  });

  it("should strip HTML from email and validate result", () => {
    // After stripping HTML, if the resulting string is a valid email, it's kept
    expect(sanitizeEmail("<script>alert(1)</script>user@example.com")).toBe("user@example.com");
    // If stripping HTML leaves an invalid email, it returns empty
    expect(sanitizeEmail("<script>user@example.com</script>")).toBe("");
  });
});

describe("escapeHtml", () => {
  it("should escape HTML special characters", () => {
    expect(escapeHtml("<script>")).toBe("&lt;script&gt;");
    expect(escapeHtml('"quoted"')).toBe("&quot;quoted&quot;");
    expect(escapeHtml("a & b")).toBe("a &amp; b");
  });

  it("should handle empty inputs", () => {
    expect(escapeHtml("")).toBe("");
    expect(escapeHtml(null)).toBe("");
    expect(escapeHtml(undefined)).toBe("");
  });

  it("should preserve normal text", () => {
    expect(escapeHtml("Hello World")).toBe("Hello World");
  });
});

describe("sanitizeObject", () => {
  it("should sanitize string values", () => {
    const input = {
      name: "<script>alert(1)</script>John",
      email: "john@example.com",
    };
    const result = sanitizeObject(input);
    expect(result.name).toBe("John");
    expect(result.email).toBe("john@example.com");
  });

  it("should handle nested objects", () => {
    const input = {
      user: {
        name: "<b>John</b>",
        profile: {
          bio: "<script>evil()</script>Hello",
        },
      },
    };
    const result = sanitizeObject(input);
    expect(result.user.name).toBe("John");
    expect(result.user.profile.bio).toBe("Hello");
  });

  it("should handle arrays", () => {
    const input = {
      tags: ["<script>", "normal", "<b>bold</b>"],
    };
    const result = sanitizeObject(input);
    expect(result.tags).toEqual(["", "normal", "bold"]);
  });

  it("should preserve non-string values", () => {
    const input = {
      count: 42,
      active: true,
      data: null,
    };
    const result = sanitizeObject(input);
    expect(result.count).toBe(42);
    expect(result.active).toBe(true);
    expect(result.data).toBe(null);
  });
});

describe("sanitizeFilename", () => {
  it("should remove path traversal attempts", () => {
    expect(sanitizeFilename("../../../etc/passwd")).toBe("etcpasswd");
    expect(sanitizeFilename("..\\..\\windows\\system32")).toBe("windowssystem32");
  });

  it("should remove dangerous characters", () => {
    expect(sanitizeFilename("file<script>.txt")).toBe("file_script_.txt");
    expect(sanitizeFilename("file\x00.txt")).toBe("file.txt");
  });

  it("should handle empty inputs", () => {
    expect(sanitizeFilename("")).toBe("");
    expect(sanitizeFilename(null)).toBe("");
    expect(sanitizeFilename(".")).toBe("unnamed");
    expect(sanitizeFilename("..")).toBe("unnamed");
  });

  it("should preserve valid filenames", () => {
    expect(sanitizeFilename("document.pdf")).toBe("document.pdf");
    expect(sanitizeFilename("my-file_v2.txt")).toBe("my-file_v2.txt");
  });
});
