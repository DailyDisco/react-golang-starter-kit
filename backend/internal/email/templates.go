package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

// TemplateManager handles email templates
type TemplateManager struct {
	templates map[string]*template.Template
}

// TemplateData contains common data available to all templates
type TemplateData struct {
	AppName      string
	SupportEmail string
	CurrentYear  int
	FrontendURL  string
	Data         map[string]interface{}
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() (*TemplateManager, error) {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
	}

	// Parse all templates
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".html")
		content, err := templateFS.ReadFile("templates/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", name, err)
		}

		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		tm.templates[name] = tmpl
	}

	return tm, nil
}

// Render renders a template with the given data
// Returns: subject, htmlBody, textBody, error
func (tm *TemplateManager) Render(name string, data map[string]interface{}) (string, string, string, error) {
	tmpl, ok := tm.templates[name]
	if !ok {
		return "", "", "", fmt.Errorf("template not found: %s", name)
	}

	// Add common template data
	templateData := TemplateData{
		AppName:      getAppName(),
		SupportEmail: getSupportEmail(),
		CurrentYear:  time.Now().Year(),
		FrontendURL:  GetFrontendURL(),
		Data:         data,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	htmlBody := buf.String()

	// Extract subject from template if present (<!-- SUBJECT: ... -->)
	subject := extractSubject(htmlBody)

	// Generate plain text version (basic HTML stripping)
	textBody := generatePlainText(htmlBody)

	return subject, htmlBody, textBody, nil
}

// HasTemplate checks if a template exists
func (tm *TemplateManager) HasTemplate(name string) bool {
	_, ok := tm.templates[name]
	return ok
}

func extractSubject(html string) string {
	const prefix = "<!-- SUBJECT: "
	const suffix = " -->"

	start := strings.Index(html, prefix)
	if start == -1 {
		return ""
	}

	start += len(prefix)
	end := strings.Index(html[start:], suffix)
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(html[start : start+end])
}

func generatePlainText(html string) string {
	text := html

	// Remove HTML comments
	for {
		start := strings.Index(text, "<!--")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "-->")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+3:]
	}

	// Replace common HTML elements with plain text equivalents
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n\n")
	text = strings.ReplaceAll(text, "</div>", "\n")
	text = strings.ReplaceAll(text, "</li>", "\n")
	text = strings.ReplaceAll(text, "</h1>", "\n\n")
	text = strings.ReplaceAll(text, "</h2>", "\n\n")
	text = strings.ReplaceAll(text, "</h3>", "\n")

	// Strip remaining HTML tags
	var result strings.Builder
	inTag := false
	for _, r := range text {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	// Clean up whitespace
	text = result.String()
	text = strings.TrimSpace(text)

	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	return text
}

func getAppName() string {
	if config != nil && config.SMTPFromName != "" {
		return config.SMTPFromName
	}
	return "React Go Starter"
}

func getSupportEmail() string {
	if config != nil && config.SMTPFrom != "" {
		return config.SMTPFrom
	}
	return "support@example.com"
}
