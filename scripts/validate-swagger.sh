#!/bin/bash
# Validate Swagger docs are up-to-date (called by lint-staged)
set -e

BACKEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../backend" && pwd)"

# Skip if no Go files staged
if ! git diff --cached --name-only | grep -q "^backend/.*\.go$"; then
    exit 0
fi

cd "$BACKEND_DIR"

# Skip if swag not installed (CI will catch issues)
if ! command -v swag &> /dev/null; then
    echo "Warning: swag not installed, skipping validation"
    exit 0
fi

# Skip if no docs exist yet
if [ ! -f "docs/swagger.json" ]; then
    exit 0
fi

# Compare current docs with fresh generation
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

swag init -g cmd/main.go -o "$TEMP_DIR" 2>/dev/null
sed -i '/LeftDelim:/d; /RightDelim:/d' "$TEMP_DIR/docs.go" 2>/dev/null || true

if ! diff -q docs/swagger.json "$TEMP_DIR/swagger.json" > /dev/null 2>&1; then
    echo "Swagger docs outdated, regenerating..."
    swag init -g cmd/main.go -o docs
    sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go 2>/dev/null || true
    git add docs/
    echo "Swagger docs regenerated and staged."
else
    echo "Swagger docs up-to-date."
fi
