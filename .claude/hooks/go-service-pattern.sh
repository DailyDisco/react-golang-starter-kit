#!/bin/bash
# Hook: go-service-pattern
# Trigger: PostEdit on backend/internal/services/*.go
# Purpose: Validate Go service files follow project patterns

set -euo pipefail

FILE_PATH="${1:-}"
NEW_CONTENT="${2:-}"

# Skip test files and non-service files
if [[ "$FILE_PATH" == *"_test.go" ]] || [[ ! "$FILE_PATH" == *"_service.go" ]]; then
    exit 0
fi

WARNINGS=""
ERRORS=""

# Helper to add warning
warn() {
    WARNINGS="${WARNINGS}  - $1\n"
}

# Helper to add error
error() {
    ERRORS="${ERRORS}  - $1\n"
}

# Check for sentinel errors at top of file
if ! echo "$NEW_CONTENT" | grep -q "^var ($" || ! echo "$NEW_CONTENT" | grep -q "Err.*=.*errors.New"; then
    warn "Missing sentinel error definitions (var ( Err... = errors.New(...) ) block)"
fi

# Check for service struct with db field
if ! echo "$NEW_CONTENT" | grep -qE "type \w+Service struct \{"; then
    error "Missing service struct definition (type XxxService struct { ... })"
elif ! echo "$NEW_CONTENT" | grep -qE "db \*gorm\.DB"; then
    warn "Service struct should have 'db *gorm.DB' field"
fi

# Check for constructor function
SERVICE_NAME=$(echo "$NEW_CONTENT" | grep -oE "type (\w+)Service struct" | head -1 | sed 's/type \(\w*\)Service struct/\1/')
if [[ -n "$SERVICE_NAME" ]]; then
    if ! echo "$NEW_CONTENT" | grep -qE "func New${SERVICE_NAME}Service\(db \*gorm\.DB\)"; then
        warn "Missing constructor: New${SERVICE_NAME}Service(db *gorm.DB)"
    fi
fi

# Check that methods use context.Context as first param
# Look for methods on the service that don't start with (ctx context.Context
METHOD_PATTERN="func \(s \*\w+Service\) \w+\("
if echo "$NEW_CONTENT" | grep -qE "$METHOD_PATTERN"; then
    # Get methods without context as first param
    BAD_METHODS=$(echo "$NEW_CONTENT" | grep -oE "func \(s \*\w+Service\) \w+\([^)]*\)" | grep -v "ctx context.Context" | grep -v "func (s \*\w*Service) \w+()") || true
    if [[ -n "$BAD_METHODS" ]]; then
        warn "Service methods should accept context.Context as first parameter"
    fi
fi

# Check for proper error wrapping
if echo "$NEW_CONTENT" | grep -qE "return err$" && ! echo "$NEW_CONTENT" | grep -qE "fmt\.Errorf.*%w"; then
    warn "Errors should be wrapped with context: fmt.Errorf(\"description: %w\", err)"
fi

# Check for WithContext usage with db
if echo "$NEW_CONTENT" | grep -qE "s\.db\.(Create|First|Find|Save|Delete|Update|Where)" && ! echo "$NEW_CONTENT" | grep -qE "s\.db\.WithContext\(ctx\)"; then
    warn "Database operations should use WithContext(ctx)"
fi

# Output results
if [[ -n "$ERRORS" ]]; then
    echo "## Service Pattern Errors"
    echo ""
    echo -e "$ERRORS"
    echo ""
    echo "Please fix these issues before proceeding."
    exit 1
fi

if [[ -n "$WARNINGS" ]]; then
    echo "## Service Pattern Warnings"
    echo ""
    echo -e "$WARNINGS"
    echo ""
    echo "Consider addressing these patterns for consistency."
fi

exit 0
