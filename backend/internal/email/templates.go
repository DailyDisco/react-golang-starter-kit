package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	"time"
)

//go:embed templates/layouts/*.html templates/partials/*.html templates/emails/*.html
var templateFS embed.FS

// TemplateManager handles email templates with layout inheritance
type TemplateManager struct {
	templates map[string]*template.Template
	baseSet   *template.Template
}

// TemplateData contains common data available to all templates
type TemplateData struct {
	AppName      string
	SupportEmail string
	CurrentYear  int
	FrontendURL  string
	Data         map[string]interface{}
}

// NewTemplateManager creates a new template manager with layout support
func NewTemplateManager() (*TemplateManager, error) {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
	}

	// Parse base templates (layouts + partials)
	baseSet, err := tm.parseBaseTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to parse base templates: %w", err)
	}
	tm.baseSet = baseSet

	// Parse email templates
	if err := tm.parseEmailTemplates(); err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return tm, nil
}

// parseBaseTemplates loads all layouts and partials into a base template set
func (tm *TemplateManager) parseBaseTemplates() (*template.Template, error) {
	baseSet := template.New("base")

	// Load layouts
	layoutFiles, err := fs.Glob(templateFS, "templates/layouts/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob layouts: %w", err)
	}

	for _, file := range layoutFiles {
		content, err := templateFS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read layout %s: %w", file, err)
		}
		if _, err := baseSet.Parse(string(content)); err != nil {
			return nil, fmt.Errorf("failed to parse layout %s: %w", file, err)
		}
	}

	// Load partials
	partialFiles, err := fs.Glob(templateFS, "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob partials: %w", err)
	}

	for _, file := range partialFiles {
		content, err := templateFS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read partial %s: %w", file, err)
		}
		if _, err := baseSet.Parse(string(content)); err != nil {
			return nil, fmt.Errorf("failed to parse partial %s: %w", file, err)
		}
	}

	return baseSet, nil
}

// parseEmailTemplates loads all email templates, each inheriting from base set
func (tm *TemplateManager) parseEmailTemplates() error {
	emailFiles, err := fs.Glob(templateFS, "templates/emails/*.html")
	if err != nil {
		return fmt.Errorf("failed to glob emails: %w", err)
	}

	for _, file := range emailFiles {
		// Extract template name from path
		name := extractTemplateName(file)

		content, err := templateFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", name, err)
		}

		// Clone base set and add this template
		tmpl, err := tm.baseSet.Clone()
		if err != nil {
			return fmt.Errorf("failed to clone base for %s: %w", name, err)
		}
		if _, err := tmpl.Parse(string(content)); err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		tm.templates[name] = tmpl
	}

	return nil
}

// extractTemplateName gets the template name from a file path
func extractTemplateName(path string) string {
	// Remove directory prefix and .html suffix
	name := strings.TrimPrefix(path, "templates/emails/")
	return strings.TrimSuffix(name, ".html")
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

	// Check if template uses base layout (has "base" defined)
	if tmpl.Lookup("base") != nil {
		// Execute base template
		if err := tmpl.ExecuteTemplate(&buf, "base", templateData); err != nil {
			return "", "", "", fmt.Errorf("failed to execute base template: %w", err)
		}
	} else {
		// Execute as standalone template
		if err := tmpl.Execute(&buf, templateData); err != nil {
			return "", "", "", fmt.Errorf("failed to execute template: %w", err)
		}
	}

	htmlBody := buf.String()

	// Extract subject from "subject" template block
	var subject string
	if tmpl.Lookup("subject") != nil {
		var subjectBuf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&subjectBuf, "subject", templateData); err == nil {
			subject = strings.TrimSpace(subjectBuf.String())
		}
	}

	// Generate plain text version (basic HTML stripping)
	textBody := generatePlainText(htmlBody)

	return subject, htmlBody, textBody, nil
}

// HasTemplate checks if a template exists
func (tm *TemplateManager) HasTemplate(name string) bool {
	_, ok := tm.templates[name]
	return ok
}

// ListTemplates returns all available template names
func (tm *TemplateManager) ListTemplates() []string {
	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
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
