#!/bin/bash
set -e

VERSION=${1:-1.0.0}

echo "=== Nutricionist Build v$VERSION ==="

# 1. Build frontend
echo "[1/3] Building frontend..."
cd frontend
npm run build
cd ..

# 2. Build Windows 32-bit .exe
# El nombre del asset (nutricionist-windows-386.exe) debe coincidir EXACTO con
# el que busca updater/updater.go (assetNameFor) al revisar GitHub Releases.
echo "[2/3] Building Windows 32-bit exe..."
ASSET_NAME="nutricionist-windows-386.exe"
GOOS=windows GOARCH=386 CGO_ENABLED=0 \
  go build \
  -ldflags "-X main.Version=$VERSION -s -w" \
  -o "$ASSET_NAME" \
  .
cp "$ASSET_NAME" "nutricionist-win32-v$VERSION.exe"

# 3. Size report
SIZE=$(du -sh "$ASSET_NAME" | cut -f1)
echo "[3/3] Done! $ASSET_NAME ($SIZE) — copia legible: nutricionist-win32-v$VERSION.exe"
echo ""
echo "Para publicar (el asset debe llamarse EXACTO $ASSET_NAME para que el auto-updater lo encuentre):"
echo "  gh release create v$VERSION $ASSET_NAME --title \"v$VERSION\" --notes \"Descripcion de cambios\""
