#!/usr/bin/env bash
# Memasang VibeLog di Linux tanpa Docker: satu binary, satu service systemd.
#
#   sudo bash scripts/install.sh
#
# Variabel opsional:
#   VIBELOG_VERSION=1.2.3     pasang versi tertentu (default: release terbaru)
#   VIBELOG_REPO=owner/repo   sumber release
#   VIBELOG_BASE_URL=...      mirror unduhan (opsional)
#
# Script ini tidak menimpa database atau /etc/vibelog/vibelog.env yang sudah
# ada, jadi aman dijalankan ulang untuk upgrade.
set -euo pipefail

REPO="${VIBELOG_REPO:-enpidi26/vibelog}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
DATA_DIR="${DATA_DIR:-/var/lib/vibelog}"
CONF_DIR="${CONF_DIR:-/etc/vibelog}"
SERVICE_NAME="vibelog"

die() { echo "error: $*" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || die "jalankan sebagai root: sudo bash scripts/install.sh"
command -v curl >/dev/null || die "curl tidak ditemukan"
command -v tar >/dev/null || die "tar tidak ditemukan"
command -v sha256sum >/dev/null || die "sha256sum tidak ditemukan"

case "$(uname -s)" in
  Linux) ;;
  *) die "script ini untuk Linux. Untuk macOS/Windows, build dari source atau pakai Docker." ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) die "arsitektur $(uname -m) belum punya release. Build dari source: make build" ;;
esac

if [ -n "${VIBELOG_VERSION:-}" ]; then
  TAG="v${VIBELOG_VERSION#v}"
else
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)
fi
[ -n "${TAG:-}" ] || die "belum ada release di ${REPO}. Build dari source: make build"

VERSION="${TAG#v}"
ASSET="vibelog_${VERSION}_linux_${ARCH}.tar.gz"
BASE="${VIBELOG_BASE_URL:-https://github.com/${REPO}/releases/download/${TAG}}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Mengunduh ${ASSET}..."
curl -fsSL "$BASE/$ASSET" -o "$TMP/$ASSET"
curl -fsSL "$BASE/checksums.txt" -o "$TMP/checksums.txt"
(
  cd "$TMP"
  grep -F " $ASSET" checksums.txt | sha256sum -c - >/dev/null
) || die "checksum ${ASSET} tidak cocok; unduhan dibatalkan"

tar -xzf "$TMP/$ASSET" -C "$TMP"
[ -f "$TMP/vibelog" ] || die "arsip tidak berisi binary vibelog"
install -m 0755 "$TMP/vibelog" "$BIN_DIR/vibelog"

if ! id -u vibelog >/dev/null 2>&1; then
  useradd --system --user-group --home-dir "$DATA_DIR" --shell /usr/sbin/nologin vibelog
fi
install -d -o vibelog -g vibelog -m 0700 "$DATA_DIR"
install -d -m 0755 "$CONF_DIR"

if [ ! -f "$CONF_DIR/vibelog.env" ]; then
  cat > "$CONF_DIR/vibelog.env" <<EOF
APP_ENV=production
# Ganti dengan alamat publik yang dibuka di browser, tanpa trailing slash.
# Contoh: BASE_URL=https://vibelog.example.com
BASE_URL=http://localhost:8080
HOST=127.0.0.1
DATABASE_PATH=$DATA_DIR/vibelog.db
ALLOW_REGISTRATION=true
EMAIL_DRIVER=console
EOF
  chmod 600 "$CONF_DIR/vibelog.env"
  echo "Konfigurasi dibuat: $CONF_DIR/vibelog.env"
fi

if ! command -v systemctl >/dev/null; then
  echo ""
  echo "systemd tidak ditemukan. Jalankan manual:"
  echo "  set -a; . $CONF_DIR/vibelog.env; set +a; $BIN_DIR/vibelog"
  exit 0
fi

cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<'UNIT'
[Unit]
Description=VibeLog prompt journal
Documentation=https://github.com/enpidi26/vibelog
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=vibelog
Group=vibelog
WorkingDirectory=/var/lib/vibelog
EnvironmentFile=-/etc/vibelog/vibelog.env
ExecStart=/usr/local/bin/vibelog
Restart=on-failure
RestartSec=3
StateDirectory=vibelog
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
UNIT

systemctl daemon-reload
systemctl enable "$SERVICE_NAME" >/dev/null
systemctl restart "$SERVICE_NAME"

sleep 1
if ! systemctl is-active --quiet "$SERVICE_NAME"; then
  journalctl -u "$SERVICE_NAME" -n 30 --no-pager || true
  die "service gagal start"
fi

echo ""
echo "VibeLog ${VERSION} terpasang dan berjalan."
echo "  Konfigurasi : $CONF_DIR/vibelog.env"
echo "  Data        : $DATA_DIR"
echo "  Log         : journalctl -u $SERVICE_NAME -f"
echo ""
echo "Langkah berikutnya:"
echo "  1. Edit BASE_URL di $CONF_DIR/vibelog.env"
echo "  2. sudo systemctl restart $SERVICE_NAME"
echo "  3. Buka BASE_URL, daftar akun pertama."
echo "     Magic link (EMAIL_DRIVER=console) tercetak di journal."
echo "  4. Setelah akun jadi, set ALLOW_REGISTRATION=false lalu restart."
