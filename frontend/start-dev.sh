#!/bin/sh
echo "Starting Vite dev server..."
cd /app

# Check if vite is installed
if [ ! -f "./node_modules/.bin/vite" ]; then
  echo "Vite not found, installing..."
  npm install vite
fi

# Run vite
echo "Running vite..."
./node_modules/.bin/vite --host 0.0.0.0 --port 5173
