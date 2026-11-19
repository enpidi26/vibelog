# VibeLog — prompt journal for vibe coders

VibeLog menyimpan prompt persis seperti yang kamu kirim ke AI, mencatat model
dan tool yang dipakai, menilai hasilnya, dan membuat iterasi berikutnya dalam
satu klik. Entri yang terfilter bisa diekspor menjadi prompt pack Markdown
atau JSON.

- **Self-host dulu**: satu container, satu volume, satu port. Frontend sudah
  di-embed ke dalam binary Go.
- **Tanpa layanan wajib**: SQLite, password + JWT. Magic link bisa memakai
  console, SMTP sendiri, atau Resend.
- **Data milikmu**: semua entri disaring per user; tidak ada telemetri.

Stack: Go (stdlib + SQLite), Vue 3 + TypeScript + Vite, TanStack Query.

## Self-host dalam 3 menit (Docker)

```bash
cp .env.example .env
# Edit BASE_URL menjadi alamat yang benar, mis. http://192.168.1.10:8090
docker compose up -d --build
```

Buka `http://localhost:8090`, daftar akun pertama, lalu pakai.

Tidak perlu mengisi `JWT_SECRET`: secret dibuat otomatis di volume
`vibelog-data` pada start pertama. Tidak perlu nginx dan tidak perlu
mengatur dua service.

Setelah akun sendiri dibuat, kunci pendaftaran:

```bash
# .env
ALLOW_REGISTRATION=false
```

lalu `docker compose up -d`.

### Magic link

Default `EMAIL_DRIVER=console` mencetak tautan masuk ke log container:

```bash
docker compose logs -f vibelog
```

Untuk mengirim email sungguhan, gunakan SMTP sendiri:

```env
EMAIL_DRIVER=smtp
EMAIL_FROM=VibeLog <vibelog@example.com>
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=vibelog@example.com
SMTP_PASSWORD=secret
SMTP_TLS=starttls
```

Port 465 memakai `SMTP_TLS=implicit`. `EMAIL_DRIVER=off` mematikan login
via email; login password tetap berjalan.

### HTTPS dan reverse proxy

Container mendengarkan di `127.0.0.1:8090` (ubah `PORT` di `.env`). Taruh
nginx/Caddy di depannya, arahkan ke port itu, lalu set:

```env
BASE_URL=https://vibelog.example.com
TRUSTED_PROXIES=172.16.0.0/12
```

Contoh konfigurasi nginx ada di `deploy/nginx-vibelog.conf`. `TRUSTED_PROXIES`
wajib diisi alamat proxy; tanpa itu semua pengunjung terlihat sebagai satu IP
dan kuota rate limit dibagi bersama.

Panduan operasional lengkap (backup, upgrade, restore, SMTP, troubleshooting):
[`docs/SELF_HOST.md`](docs/SELF_HOST.md). Latar produk dan screenshot:
[`docs/PRODUCT.md`](docs/PRODUCT.md).

## Backup

```bash
docker compose exec vibelog vibelog backup /data/backups
```

Perintah memakai `VACUUM INTO` sehingga aman dijalankan saat server hidup.
Arsip lama dipangkas otomatis (`BACKUP_KEEP`, default 7). Jadwalkan lewat cron
host:

```cron
0 3 * * * cd /srv/vibelog && docker compose exec -T vibelog vibelog backup /data/backups
```

Restore: hentikan container, ganti `/data/vibelog.db` dengan file arsip,
start lagi.

## Self-host tanpa Docker (binary + systemd)

Runtime-nya satu binary murni-Go: SQLite dan frontend sudah di dalamnya. Tidak
butuh gcc, Node, nginx, atau Docker di server.

Instal di Linux (release terbaru, otomatis membuat service systemd):

```bash
curl -fsSLO https://raw.githubusercontent.com/enpidi26/vibelog/main/scripts/install.sh
less install.sh                     # periksa dulu
VIBELOG_DRY_RUN=1 bash install.sh   # unduh + cek checksum saja
sudo bash install.sh
```

Script membuat user sistem `vibelog`, data di `/var/lib/vibelog`,
konfigurasi di `/etc/vibelog/vibelog.env`, lalu menyalakan service. Setelah itu:

```bash
sudo nano /etc/vibelog/vibelog.env   # set BASE_URL
sudo systemctl restart vibelog
journalctl -u vibelog -f             # magic link console tampil di sini
```

Manual tanpa script/systemd: unduh arsip dari halaman Releases, ekstrak, lalu

```bash
APP_ENV=production BASE_URL=http://localhost:8080 DATABASE_PATH=/var/lib/vibelog/vibelog.db ./vibelog
```

Build dari source (Go 1.26+, Node.js 22+, tanpa gcc):

```bash
make build          # frontend di-embed, lalu binary Go
./vibelog
```

Docker Compose tetap tersedia sebagai alternatif (bagian di atas).

## Konfigurasi

| Variabel | Default | Keterangan |
| --- | --- | --- |
| `APP_ENV` | `development` | `production` menolak secret lemah |
| `BASE_URL` | `http://localhost:8080` | URL publik; dipakai magic link, OAuth, CORS |
| `HOST` / `PORT` | `127.0.0.1` / `8080` | Alamat listen |
| `DATABASE_PATH` | `data/vibelog.db` | File SQLite |
| `JWT_SECRET` | dibuat otomatis | Minimal 32 karakter bila diisi |
| `JWT_SECRET_PATH` | sebelah database | Lokasi secret otomatis |
| `JWT_EXPIRY_HOURS` | `24` | Masa berlaku sesi, 1–720 |
| `ALLOW_REGISTRATION` | `true` | `false` menutup pendaftaran akun baru |
| `EMAIL_DRIVER` | `console` | `console`, `smtp`, `resend`, `off` |
| `EMAIL_FROM` | `VibeLog <no-reply@localhost>` | Alamat pengirim |
| `SMTP_HOST` / `SMTP_PORT` | — / `587` | SMTP |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | — | Kredensial SMTP |
| `SMTP_TLS` | `starttls` | `starttls`, `implicit`, `none` |
| `RESEND_API_KEY` | — | Untuk `EMAIL_DRIVER=resend` |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | — | OAuth opsional |
| `TRUSTED_PROXIES` | lokal | IP/CIDR proxy yang boleh mengisi `X-Forwarded-For` |
| `BACKUP_KEEP` | `7` | Jumlah arsip backup yang disimpan |

## API

Semua endpoint di bawah `/api/` (kecuali `/health`) memakai
`Authorization: Bearer <token>`. Error selalu JSON
`{"error":{"code":"...","message":"..."}}`.

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| GET | `/health` | Status server + database |
| GET | `/livez` | Liveness tanpa cek database |
| GET | `/api/meta` | Fitur yang aktif: email, GitHub, pendaftaran |
| POST | `/api/auth/register` | Daftar, mengembalikan token + user |
| POST | `/api/auth/login` | Masuk |
| GET | `/api/auth/me` | Profil dari token |
| GET | `/api/auth/github` | Mulai OAuth GitHub |
| GET | `/api/auth/github/callback` | Callback OAuth |
| POST | `/api/auth/email/request` | Kirim magic link |
| POST | `/api/auth/email/verify` | Tukar token magic link jadi JWT |
| GET, POST | `/api/projects` | Daftar dan buat proyek |
| PATCH, DELETE | `/api/projects/{id}` | Ubah dan hapus proyek |
| GET, POST | `/api/entries` | Daftar (filter `q`, `project_id`, `model`, `tool`, `outcome`, `tag`, `limit`) dan buat entri |
| GET, PATCH, DELETE | `/api/entries/{id}` | Baca, ubah, hapus entri |
| POST | `/api/entries/{id}/duplicate` | Iterasi berikutnya |
| GET | `/api/entries/summary` | Ringkasan, opsional `project_id` |
| GET | `/api/entries/export` | Ekspor `format=md` atau `json`, mengikuti filter, tanpa batas jumlah |

Magic link adalah token sekali pakai yang berlaku 15 menit. Tautan mengarah
ke halaman frontend dan ditukar lewat `POST`, sehingga pemindai email tidak
membakarnya sebelum diklik.

Validasi: judul maksimal 120 karakter, prompt dan catatan maksimal 10000
karakter, rating 1–5, maksimal 10 tag, body JSON maksimal 1 MB, field tak
dikenal ditolak.

## Keamanan

- Password bcrypt; JWT HS256 dengan issuer dan algoritma dikunci.
- Semua query parameterized dan disaring `user_id` dari token.
- CORS satu origin (`BASE_URL`), CSP, `X-Frame-Options: DENY`, `nosniff`.
- Rate limit per IP: auth 20/menit, API 300/menit, dengan prunning.
- Body limit 1 MB, timeout request 30 detik, graceful shutdown.
- SQLite WAL + foreign keys; migrasi bernomor yang aman untuk upgrade.
- Secret JWT otomatis ditulis dengan mode `0600`.

Detail dan cara melaporkan masalah: [`SECURITY.md`](SECURITY.md).

## Pengembangan

```bash
make test    # go test + typecheck frontend
make build   # frontend di-embed, lalu binary Go
```

Struktur:

```
cmd/vibelog/     entrypoint (server, backup, usage)
internal/        domain, database, auth, config, email, export, httpapi, netutil, ratelimit, web
frontend/src/    SPA Vue (lib, composables, pages, components)
deploy/          unit systemd + contoh reverse proxy
docs/            panduan self-host, produk, verifikasi
scripts/         install.sh, run.sh
```

Tidak ada paket `main` di luar `cmd/`; setiap paket punya tesnya sendiri.

## Release

Push tag `v*` (misalnya `v1.0.0`) untuk memicu GitHub Actions: binary linux
amd64/arm64, macOS, Windows, checksums, dan GitHub Release. `scripts/install.sh`
mengunduh release terbaru dari situ.

## Batasan saat ini

Cocok untuk self-host pribadi atau tim kecil. Belum untuk SaaS publik
multi-tenant:

- SQLite satu instance; belum ada mode multi-node.
- Belum ada ubah password, hapus akun, atau halaman profil. Akses bisa
dipulihkan lewat magic link bila email aktif.
- Daftar entri menampilkan 100 terbaru; pencarian dan ekspor tetap mencakup
seluruh data (ekspor tanpa batas).
- Rate limit disimpan di memori per proses.
- Belum ada import; tersedia ekspor Markdown/JSON.
- GitHub OAuth dan SMTP perlu diuji dengan kredensial nyata sebelum instance
dibuka ke publik.

Untuk instance publik: pasang HTTPS, isi `BASE_URL`, konfigurasi SMTP, lalu
set `ALLOW_REGISTRATION=false` setelah akun dibuat.

## Lisensi

Kode sumber memakai **MIT**. Lihat [`LICENSE`](LICENSE). MIT mengizinkan
fork, penggunaan komersial, dan redistribusi closed-source; itu konsekuensi
open source dan tidak bisa dicabut untuk salinan yang sudah beredar.

Nama, logo, dan identitas visual VibeLog tidak ikut dilisensikan lewat MIT.
Lihat [`TRADEMARK.md`](TRADEMARK.md). Aturan kontribusi ada di
[`CONTRIBUTING.md`](CONTRIBUTING.md).
