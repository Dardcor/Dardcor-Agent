#!/bin/bash
set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Building Dardcor Agent Executable"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Step 1: Build Frontend
echo "[1/3] Building frontend assets..."
npm install
npm run build

# Step 2: Build Backend (Go)
echo "[2/3] Compiling Go binary..."
go build -ldflags="-s -w" -o dardcor main.go

# Step 3: Deployment Info
echo "[3/3] Build complete: dardcor"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Success!"
echo ""
echo "To install globally:"
echo "1. mkdir -p ~/.local/bin"
echo "2. mv dardcor ~/.local/bin/"
echo "3. Add export PATH=\"\$HOME/.local/bin:\$PATH\" to your shell profile"
echo ""
echo "After installation, you can run 'dardcor' from any terminal."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
