#!/usr/bin/env bash
# Menjalankan VibeLog sebagai satu binary: frontend di-embed, satu port.
# Cara pakai:  ./scripts/run.sh
# Lalu buka:   http://127.0.0.1:8080
# Berhenti:    tekan Ctrl+C.
#
# Untuk server, pakai Docker Compose (lihat README).
set -euo pipefail
cd "$(dirname "$0")/.."

if [ -d "$HOME/tools/go/bin" ]; then export PATH="$HOME/tools/go/bin:$PATH"; fi
if [ -d /usr/local/go/bin ]; then export PATH="/usr/local/go/bin:$PATH"; fi

command -v go >/dev/null || { echo "go tidak ditemukan. Pasang Go 1.24+ dulu: https://go.dev/dl/"; exit 1; }
command -v npm >/dev/null || { echo "npm tidak ditemukan. Pasang Node.js 20+ dulu: https://nodejs.org"; exit 1; }

export APP_ENV="${APP_ENV:-development}"
export HOST="${HOST:-127.0.0.1}"
export PORT="${PORT:-8080}"
export BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
export DATABASE_PATH="${DATABASE_PATH:-data/vibelog.db}"
export EMAIL_DRIVER="${EMAIL_DRIVER:-console}"
mkdir -p data

if [ ! -d frontend/node_modules ]; then
  echo "Memasang dependency frontend..."
  (cd frontend && npm ci --silent)
fi

echo "Membangun frontend..."
(cd frontend && npm run build --silent)
rm -rf internal/web/dist && mkdir -p internal/web/dist
cp -r frontend/dist/. internal/web/dist/
touch internal/web/dist/.gitkeep

echo "Membangun backend..."
CGO_ENABLED=0 go build -trimpath -o vibelog ./cmd/vibelog

echo ""
echo "VibeLog berjalan di ${BASE_URL}"
echo "Magic link (EMAIL_DRIVER=console) tercetak di terminal ini."
echo "Tekan Ctrl+C untuk berhenti."
echo ""
exec ./vibelog
