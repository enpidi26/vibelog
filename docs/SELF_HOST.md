# Panduan self-host VibeLog

Panduan ini untuk satu instance VibeLog milik sendiri: satu container, satu
volume, satu port. Tidak ada layanan eksternal yang wajib.

## 1. Kebutuhan

- Linux dengan systemd (disarankan), atau
- Docker 24+ dan Docker Compose v2, atau
- Go 1.26+ dan Node.js 22+ bila build dari source.
- 256 MB RAM dan 1 GB disk sudah cukup untuk pemakaian pribadi.
- Domain + HTTPS bila instance dibuka ke internet.

## 2. Instalasi

### Binary + systemd (tanpa Docker)

```bash
curl -fsSLO https://raw.githubusercontent.com/enpidi26/vibelog/main/scripts/install.sh
less install.sh
sudo bash install.sh
```

Script memasang binary dari release terbaru, memverifikasi checksum, membuat
user `vibelog`, direktori data `/var/lib/vibelog`, konfigurasi
`/etc/vibelog/vibelog.env`, dan unit systemd `vibelog`. Aman dijalankan
ulang untuk upgrade; database dan env yang sudah ada tidak ditimpa.

```bash
sudo nano /etc/vibelog/vibelog.env   # BASE_URL
sudo systemctl restart vibelog
journalctl -u vibelog -f
```

Tanpa systemd, jalankan binary langsung:

```bash
APP_ENV=production BASE_URL=http://localhost:8080 DATABASE_PATH=/var/lib/vibelog/vibelog.db vibelog
```

### Docker (opsional)

```bash
git clone https://github.com/enpidi26/vibelog.git
cd vibelog
cp .env.example .env
```

Edit `.env` minimal:

```env
BASE_URL=http://192.168.1.10:8090   # alamat persis yang dibuka di browser
PORT=8090
```

Lalu:

```bash
docker compose up -d --build
docker compose logs -f vibelog
```

Buka `BASE_URL`, daftar akun pertama. Kalau `EMAIL_DRIVER=console`, tautan
magic link muncul di log.

Setelah akun jadi, tutup pendaftaran:

```env
ALLOW_REGISTRATION=false
```

```bash
docker compose up -d
```

### Tanpa Docker

```bash
make build
APP_ENV=production \
BASE_URL=https://vibelog.example.com \
DATABASE_PATH=/var/lib/vibelog/vibelog.db \
./backend/vibelog
```

Binary yang sama juga punya subcommand:

```bash
vibelog backup /var/lib/vibelog/backups
vibelog version
```

## 3. Data

Semua data ada di volume Docker `vibelog-data`, di dalam container pada
`/data`:

| File | Isi |
| --- | --- |
| `vibelog.db` | database SQLite (WAL) |
| `vibelog.db-wal`, `vibelog.db-shm` | journal WAL |
| `jwt_secret` | secret JWT yang dibuat otomatis (mode 0600) |
| `backups/` | arsip backup |

Menyalin `vibelog.db` saat server hidup tidak aman karena WAL. Selalu pakai
`vibelog backup`.

### Backup

```bash
docker compose exec vibelog vibelog backup /data/backups
```

- Memakai `VACUUM INTO`, konsisten walau server hidup.
- Menyimpan `vibelog-YYYYMMDD-HHMMSS.db` dengan mode `0600`.
- Menyisakan `BACKUP_KEEP` arsip terbaru (default 7).
- Salin juga `jwt_secret` bila ingin sesi login lama tetap valid setelah
  restore. Kalau tidak, semua pengguna cukup login ulang.

Jadwalkan dari host:

```cron
0 3 * * * cd /srv/vibelog && docker compose exec -T vibelog vibelog backup /data/backups
```

### Restore

```bash
docker compose stop
docker compose run --rm --entrypoint sh vibelog -c \
  'cp /data/backups/vibelog-20260101-030000.db /data/vibelog.db && rm -f /data/vibelog.db-wal /data/vibelog.db-shm'
docker compose up -d
```

## 4. Email / magic link

Login password selalu tersedia. Magic link opsional.

### console (default)

Tautan dicetak ke log container:

```bash
docker compose logs -f vibelog | grep 'auth/verify'
```

Cocok untuk instance pribadi di LAN atau saat baru memasang.

### SMTP

```env
EMAIL_DRIVER=smtp
EMAIL_FROM=VibeLog <vibelog@example.com>
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=vibelog@example.com
SMTP_PASSWORD=secret
SMTP_TLS=starttls
```

- Port 587: `SMTP_TLS=starttls`.
- Port 465: `SMTP_TLS=implicit`.
- `SMTP_TLS=none` hanya untuk server lokal tepercaya; kredensial ditolak
  tanpa TLS.
- `EMAIL_FROM` harus alamat yang diizinkan server SMTP.

### Resend (cloud)

```env
EMAIL_DRIVER=resend
EMAIL_FROM=VibeLog <vibelog@example.com>
RESEND_API_KEY=re_...
```

Domain pengirim harus sudah diverifikasi di Resend.

### Mematikan email

```env
EMAIL_DRIVER=off
```

Endpoint magic link membalas 503 dengan pesan jelas; UI tetap menampilkan
opsi email, jadi arahkan pengguna ke login password.

## 5. HTTPS dan reverse proxy

Jangan buka port container langsung ke internet tanpa TLS. Contoh nginx ada
di `deploy/nginx-vibelog.conf`; Caddy lebih singkat:

```caddy
vibelog.example.com {
    reverse_proxy 127.0.0.1:8090
}
```

Setelah proxy aktif:

```env
BASE_URL=https://vibelog.example.com
TRUSTED_PROXIES=172.16.0.0/12
```

`TRUSTED_PROXIES` adalah IP/CIDR proxy yang boleh mengisi
`X-Forwarded-For`. Untuk nginx di host yang sama dengan container, biasanya
`172.16.0.0/12` (network Docker) atau `127.0.0.1/32` bila memakai
`network_mode: host`. Tanpa nilai ini, semua pengunjung dianggap satu IP dan
kuota rate limit dibagi bersama.

Jangan set `TRUSTED_PROXIES=0.0.0.0/0`: siapa pun bisa memalsukan IP dan
melewati rate limit.

## 6. Upgrade

```bash
cd /srv/vibelog
git pull
docker compose up -d --build
```

Migrasi database berjalan otomatis saat start dan dicatat di tabel
`schema_migrations`. Tetap ambil backup sebelum upgrade:

```bash
docker compose exec vibelog vibelog backup /data/backups
```

Rollback aplikasi: checkout tag/commit sebelumnya dan build ulang. Migrasi
yang sudah jalan tidak dibatalkan otomatis; restore backup bila skema baru
tidak kompatibel.

## 7. GitHub OAuth (opsional)

1. GitHub → Settings → Developer settings → OAuth Apps → New OAuth App.
2. Homepage URL: `BASE_URL`.
3. Authorization callback URL: `BASE_URL/api/auth/github/callback`.
4. Isi `GITHUB_CLIENT_ID` dan `GITHUB_CLIENT_SECRET`, lalu restart.

Callback memakai `BASE_URL`, jadi `BASE_URL` harus persis sama dengan alamat
yang dibuka pengguna (termasuk skema dan tanpa trailing slash).

## 8. Checklist sebelum dipakai publik

- [ ] `BASE_URL` benar dan memakai `https://`.
- [ ] HTTPS aktif di reverse proxy.
- [ ] `ALLOW_REGISTRATION=false` setelah akun sendiri dibuat.
- [ ] `EMAIL_DRIVER` sesuai; kalau `console`, jangan kirim log ke publik.
- [ ] `TRUSTED_PROXIES` diisi alamat proxy, bukan `0.0.0.0/0`.
- [ ] Backup terjadwal dan pernah dicoba restore.
- [ ] `APP_ENV=production`.
- [ ] Port database tidak diekspos ke internet.

## 9. Troubleshooting

**Magic link mengarah ke localhost.** `BASE_URL` belum diubah. Set alamat
publik yang benar, lalu restart.

**Login magic link berhasil di API tapi UI kembali ke halaman login.** Token
disimpan di `localStorage` dengan kunci `vibelog_token`. Pastikan memakai
frontend dari container yang sama (bukan `frontend/dist` lama).

**Rate limit terasa terlalu ketat / semua orang kena limit.** Isi
`TRUSTED_PROXIES` dengan IP proxy.

**`database is locked`.** Jalankan satu instance saja terhadap satu file
SQLite. Jangan mount database yang sama ke dua container.

**Email tidak terkirim.** Cek `docker compose logs vibelog`. Untuk SMTP,
pastikan host/port/TLS benar dan kredensial diizinkan. Untuk Resend, cek
domain pengirim sudah diverifikasi.

**Lupa password.** Pakai magic link ke email yang sama, atau buat akun baru
bila pendaftaran dibuka. Tidak ada akun seed.
