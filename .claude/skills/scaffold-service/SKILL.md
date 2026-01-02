---
name: scaffold-service
description: Scaffold a new Go service with handler, tests, and route registration. Use when adding new backend functionality.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(go:*), AskUserQuestion
context-files:
  - backend/internal/services/org_service.go
  - backend/internal/handlers/org_handlers.go
---

# Go Service Scaffolding

You are my Go service scaffolding assistant for this React + Go starter kit.

## Objective

Generate a complete Go service layer following project patterns:
- Service with sentinel errors and GORM database access
- HTTP handlers with proper error handling
- Table-driven unit tests
- Route registration snippet

## Hard Rules

1. Follow exact project patterns from existing services
2. Always use `context.Context` as first parameter in service methods
3. Wrap errors with context: `fmt.Errorf("description: %w", err)`
4. Define sentinel errors at top of service file
5. Use constructor pattern: `New{Entity}Service(db *gorm.DB)`
6. Handlers must call services, never access DB directly
7. Generate table-driven tests for all service methods

## Guided Workflow

### Phase 1: Gather Requirements

Ask the user:

1. **Entity name** (singular, lowercase): e.g., "notification", "payment", "audit"
2. **Operations needed**:
   - [ ] Create
   - [ ] GetByID
   - [ ] GetAll / List (with optional filters)
   - [ ] Update
   - [ ] Delete
   - [ ] Custom operations (describe)
3. **Related model fields** (optional): If creating a new model, ask for fields

### Phase 2: Generate Service

Create `backend/internal/services/{entity}_service.go`:

```go
package services

import (
	"context"
	"errors"
	"fmt"

	"react-golang-starter/internal/models"

	"gorm.io/gorm"
)

// Sentinel errors for {Entity}
var (
	Err{Entity}NotFound = errors.New("{entity} not found")
)

// {Entity}Service handles {entity} business logic
type {Entity}Service struct {
	db *gorm.DB
}

// New{Entity}Service creates a new {entity} service
func New{Entity}Service(db *gorm.DB) *{Entity}Service {
	return &{Entity}Service{db: db}
}

// Create creates a new {entity}
func (s *{Entity}Service) Create(ctx context.Context, {entity} *models.{Entity}) error {
	if err := s.db.WithContext(ctx).Create({entity}).Error; err != nil {
		return fmt.Errorf("create {entity}: %w", err)
	}
	return nil
}

// GetByID retrieves a {entity} by ID
func (s *{Entity}Service) GetByID(ctx context.Context, id uint) (*models.{Entity}, error) {
	var {entity} models.{Entity}
	if err := s.db.WithContext(ctx).First(&{entity}, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Err{Entity}NotFound
		}
		return nil, fmt.Errorf("get {entity} by id: %w", err)
	}
	return &{entity}, nil
}

// ... additional methods based on requirements
```

### Phase 3: Generate Handler

Create `backend/internal/handlers/{entity}_handlers.go`:

```go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"react-golang-starter/internal/services"

	"github.com/go-chi/chi/v5"
)

type {Entity}Handler struct {
	service *services.{Entity}Service
}

func New{Entity}Handler(service *services.{Entity}Service) *{Entity}Handler {
	return &{Entity}Handler{service: service}
}

func (h *{Entity}Handler) Get{Entity}(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	{entity}, err := h.service.GetByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.Err{Entity}NotFound) {
			WriteError(w, http.StatusNotFound, "{entity} not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get {entity}")
		return
	}

	WriteJSON(w, http.StatusOK, {entity})
}

// ... additional handlers based on requirements
```

### Phase 4: Generate Tests

Create `backend/internal/services/{entity}_service_test.go`:

```go
package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew{Entity}Service(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	service := New{Entity}Service(db)

	require.NotNil(t, service)
	assert.NotNil(t, service.db)
}

func Test{Entity}Service_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      uint
		setup   func(t *testing.T, db *gorm.DB)
		want    *models.{Entity}
		wantErr error
	}{
		{
			name: "returns {entity} when found",
			id:   1,
			setup: func(t *testing.T, db *gorm.DB) {
				// Create test {entity}
			},
			want:    &models.{Entity}{ID: 1},
			wantErr: nil,
		},
		{
			name:    "returns error when not found",
			id:      999,
			setup:   nil,
			want:    nil,
			wantErr: Err{Entity}NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tt.setup != nil {
				tt.setup(t, db)
			}

			service := New{Entity}Service(db)
			got, err := service.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.ID, got.ID)
		})
	}
}
```

### Phase 5: Route Registration

Provide snippet for `backend/cmd/main.go`:

```go
// Add to imports
// (service and handler already imported)

// Add to main() after other service initializations
{entity}Service := services.New{Entity}Service(db)
{entity}Handler := handlers.New{Entity}Handler({entity}Service)

// Add routes inside r.Route("/api/v1", func(r chi.Router) { ... })
r.Route("/{entities}", func(r chi.Router) {
	r.Get("/", {entity}Handler.List{Entities})
	r.Post("/", {entity}Handler.Create{Entity})
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", {entity}Handler.Get{Entity})
		r.Put("/", {entity}Handler.Update{Entity})
		r.Delete("/", {entity}Handler.Delete{Entity})
	})
})
```

### Phase 6: Verification

Run:
```bash
go build ./...
go test ./internal/services/... -run {Entity}
```

## Output Format

After each phase, show:
```
## Phase N Complete: {phase name}
- Created: {file path}
- Next: {what's next}
```

At the end:
```
## Scaffolding Complete

Files created:
- backend/internal/services/{entity}_service.go
- backend/internal/services/{entity}_service_test.go
- backend/internal/handlers/{entity}_handlers.go

Next steps:
1. Add model to backend/internal/models/models.go (if new)
2. Add route registration to backend/cmd/main.go
3. Run migrations if schema changed
```

## Constraints

- Entity names must be lowercase, singular (e.g., "notification" not "Notifications")
- PascalCase for types: "notification" → "Notification"
- File names use snake_case: "notification_service.go"
- Max 10 operations per service (split if more needed)
