#!/bin/sh
echo "Starting dev server..."
echo "PATH: $PATH"
echo "Checking npx..."
which npx
echo "Checking vite..."
which vite
echo "Listing node_modules/.bin..."
ls -la node_modules/.bin/ | head -10
echo "Running npx vite (routes will be generated automatically)..."
npx vite --host 0.0.0.0 --port 5193
