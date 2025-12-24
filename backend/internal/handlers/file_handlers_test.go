package handlers

import (
	"testing"
)

func TestIsAllowedMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     bool
	}{
		// Allowed image types
		{"JPEG image", "image/jpeg", true},
		{"PNG image", "image/png", true},
		{"GIF image", "image/gif", true},
		{"WebP image", "image/webp", true},
		{"SVG image", "image/svg+xml", true},

		// Allowed document types
		{"PDF document", "application/pdf", true},
		{"Plain text", "text/plain", true},
		{"CSV file", "text/csv", true},
		{"JSON file", "application/json", true},
		{"XML file", "application/xml", true},
		{"XML text", "text/xml", true},
		{"Markdown", "text/markdown", true},
		{"Word document", "application/msword", true},
		{"Word docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
		{"Excel xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
		{"PowerPoint pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation", true},

		// MIME types with parameters (should still work)
		{"Text with charset", "text/plain; charset=utf-8", true},
		{"JSON with charset", "application/json; charset=utf-8", true},
		{"JPEG with boundary", "image/jpeg; boundary=something", true},

		// Case insensitivity
		{"Uppercase JPEG", "IMAGE/JPEG", true},
		{"Mixed case PNG", "Image/PNG", true},

		// Not allowed types
		{"Executable", "application/x-msdownload", false},
		{"Shell script", "application/x-sh", false},
		{"Binary", "application/octet-stream", false},
		{"JavaScript", "application/javascript", false},
		{"HTML", "text/html", false},
		{"PHP", "application/x-php", false},
		{"ZIP archive", "application/zip", false},
		{"TAR archive", "application/x-tar", false},
		{"RAR archive", "application/x-rar-compressed", false},
		{"7Z archive", "application/x-7z-compressed", false},

		// Empty and edge cases
		{"Empty string", "", false},
		{"Unknown type", "application/unknown", false},
		{"Random string", "not-a-mime-type", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedMimeType(tt.mimeType)
			if got != tt.want {
				t.Errorf("isAllowedMimeType(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestAllowedMimeTypes_Contains_CommonTypes(t *testing.T) {
	// Verify that common safe file types are in the allowed list
	commonTypes := []string{
		"image/jpeg",
		"image/png",
		"application/pdf",
		"text/plain",
	}

	for _, mimeType := range commonTypes {
		if !AllowedMimeTypes[mimeType] {
			t.Errorf("AllowedMimeTypes should contain %q", mimeType)
		}
	}
}

func TestAllowedMimeTypes_DoesNotContain_DangerousTypes(t *testing.T) {
	// Verify that dangerous file types are NOT in the allowed list
	dangerousTypes := []string{
		"application/x-msdownload",    // .exe
		"application/x-sh",            // shell scripts
		"application/x-php",           // PHP scripts
		"application/javascript",      // JavaScript
		"text/html",                   // HTML (can contain scripts)
		"application/x-httpd-php",     // PHP
		"application/x-python",        // Python scripts
		"application/x-perl",          // Perl scripts
		"application/x-ruby",          // Ruby scripts
		"application/x-msdos-program", // DOS executables
	}

	for _, mimeType := range dangerousTypes {
		if AllowedMimeTypes[mimeType] {
			t.Errorf("AllowedMimeTypes should NOT contain dangerous type %q", mimeType)
		}
	}
}
