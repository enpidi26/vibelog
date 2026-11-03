#!/usr/bin/env bash
# Menjalankan VibeLog (Go + Vue) dalam satu perintah, untuk pemakaian lokal.
# Cara pakai:  ./scripts/run.sh
# Lalu buka:   http://127.0.0.1:4173  (landing page, daftar, lalu pakai aplikasinya)
# Berhenti:    tekan Ctrl+C. Script ini membersihkan kedua servernya.
#
# Untuk production di server, pakai Docker Compose atau panduan VPS di
# deploy/README.md, bukan script ini.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ -d "$HOME/tools/go/bin" ]; then export PATH="$HOME/tools/go/bin:$PATH"; fi
if [ -d /usr/local/go/bin ]; then export PATH="/usr/local/go/bin:$PATH"; fi

command -v go >/dev/null || { echo "go tidak ditemukan. Pasang Go 1.24+ dulu: https://go.dev/dl/"; exit 1; }
command -v npm >/dev/null || { echo "npm tidak ditemukan. Pasang Node.js 20+ dulu: https://nodejs.org"; exit 1; }

export APP_ENV="${APP_ENV:-development}"
export HOST="${HOST:-127.0.0.1}"
export PORT="${PORT:-8080}"
export DATABASE_PATH="${DATABASE_PATH:-data/vibelog.db}"
export JWT_EXPIRY_HOURS="${JWT_EXPIRY_HOURS:-24}"
export FRONTEND_URL="${FRONTEND_URL:-http://127.0.0.1:4173}"
export TMPDIR="${TMPDIR:-$HOME/tmp}"
mkdir -p "$TMPDIR" 2>/dev/null || true
if [ -z "${JWT_SECRET:-}" ]; then
  if [ "$APP_ENV" = "production" ]; then
    echo "JWT_SECRET wajib diisi di mode production. Isi minimal 32 karakter, lalu jalankan lagi."
    exit 1
  fi
  export JWT_SECRET="dev-only-secret-change-me-in-production"
  echo "Memakai JWT_SECRET development bawaan. Jangan dipakai untuk production."
fi

mkdir -p data
echo "Membangun backend..."
(cd backend && go build -o vibelog-api .)
echo "Membangun frontend..."
(cd frontend && npm ci --silent && npm run build --silent)

cleanup() {
  echo "Menghentikan VibeLog..."
  kill "$BACKEND_PID" "$PREVIEW_PID" 2>/dev/null || true
}
trap cleanup INT TERM EXIT

./backend/vibelog-api &
BACKEND_PID=$!
(cd frontend && npm run preview -- --host 127.0.0.1 --port 4173) &
PREVIEW_PID=$!

echo ""
echo "VibeLog berjalan di http://127.0.0.1:4173"
echo "Daftar akun baru di halaman tersebut, lalu jurnal prompt bisa langsung dipakai."
echo "Tekan Ctrl+C untuk berhenti."
wait
