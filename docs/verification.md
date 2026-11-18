# Verifikasi rilis self-host (18 November 2026)

Dokumen ini mencatat apa yang benar-benar dijalankan dan apa yang belum.
Jangan menganggap klaim README sebagai bukti tanpa membaca bagian ini.

## Yang dijalankan

### Backend

- `go vet ./...` bersih, `gofmt -l .` kosong.
- `CGO_ENABLED=0 go test -count=1 ./...` **23/23 lolos**, termasuk:
  - migrasi bernomor idempoten dan tercatat di `schema_migrations`;
  - secret JWT otomatis dibuat, dipakai ulang antar restart, dan tidak
    ditulis bila `JWT_SECRET` diisi;
  - registrasi bisa ditutup (`403`) dan magic link tidak membuat akun saat
    registrasi ditutup;
  - magic link: request → verifikasi `POST` → token tidak bisa dipakai dua
    kali; request kedua membatalkan token pertama; `GET` tidak membakar token;
  - backup rotation menyisakan arsip terbaru;
  - filter `q` dan `tag` meng-escape `%` dan `_`;
  - ekspor memuat seluruh entri (250 entri, bukan hanya 100);
  - pesan MIME tidak mengandung CR ganda;
  - pragma SQLite aktif: WAL, foreign keys, busy timeout;
  - `/health` melaporkan status database;
  - isolasi antar user, duplikasi iterasi, ringkasan, dan ekspor tetap lolos
    seperti sebelumnya.
- `CGO_ENABLED=0 go build` menghasilkan satu binary dengan frontend di-embed,
  tanpa gcc dan tanpa CGO (driver SQLite murni-Go).

### Smoke test satu binary (production mode)

Dijalankan dengan `APP_ENV=production`, database kosong, tanpa `JWT_SECRET`:

| Cek | Hasil |
| --- | --- |
| Secret JWT otomatis di volume data | dibuat, mode 0600 |
| `/health` | 200, `{"db":"ok"}` |
| `/` dan `/app` | 200 `text/html` (SPA fallback) |
| `/api/nope` | 404 JSON |
| Register | 201 + JWT |
| Buat entri + ringkasan | 201 / data benar |
| Ekspor Markdown | memuat entri |
| Magic link (console) → `POST` verify | 200 + JWT |
| Replay token | 400 |
| `vibelog backup` | file backup dibuat |

### Frontend

- `npm run typecheck` bersih.
- `npm run build` sukses (122 modul, ±191 KB JS, ±20 KB CSS).
- Perubahan rilis: verifikasi magic link lewat `POST`, tanggal memakai zona
  waktu browser, bukan `Asia/Bangkok`.

## Yang TIDAK dijalankan di mesin ini

- **Build image Docker tidak dijalankan** karena Docker tidak tersedia di
  mesin ini. Dockerfile dan compose ditulis untuk build dari root repo;
  verifikasi build image harus dilakukan di mesin dengan Docker:
  `docker compose up -d --build`.
- **CI GitHub Actions sudah hijau** di Ubuntu (Go 1.26, Node 22,
  `CGO_ENABLED=0`): typecheck, build frontend, `go vet`, dan 23 test lolos.
- **Release binary dan `scripts/install.sh` belum dijalankan end-to-end.**
  Release workflow baru aktif saat tag `v*` di-push; `install.sh` butuh Linux
  dengan systemd dan akun root, yang tidak tersedia di mesin ini.
- **SMTP dan Resend tidak diuji ke server sungguhan.** Kode memakai
  `net/smtp` (STARTTLS/implicit) dan Resend API; yang diuji adalah jalur
  `console` dan `off`.
- **GitHub OAuth tidak diuji end-to-end** karena butuh aplikasi OAuth asli.
  Perubahan yang dibuat: pertukaran code lewat `POST` body, cookie `Secure`
  saat `BASE_URL` https, dan email GitHub yang terverifikasi ikut menandai
  akun lokal terverifikasi.
- **Reverse proxy nginx/Caddy tidak dijalankan** di mesin ini. Konfigurasi
  contoh ada di `deploy/nginx-vibelog.conf`; `TRUSTED_PROXIES` sudah diuji
  lewat unit test limiter.
- **Tidak ada unit test frontend.** Verifikasi UI masih manual.
- **Windows tanpa Docker tidak diuji** untuk produksi; jalur yang diuji di
  Windows adalah build dan smoke test binary.

## Cara memverifikasi ulang

```bash
cd backend && go test -count=1 ./...
cd frontend && npm run typecheck && npm run build
cd .. && ./scripts/run.sh
# Buka http://127.0.0.1:8080, daftar, buat entri, cek magic link di log.
```

Untuk Docker:

```bash
cp .env.example .env
docker compose up -d --build
curl -fsS http://localhost:8090/health
docker compose exec vibelog vibelog backup /data/backups
```
