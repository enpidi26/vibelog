# Verifikasi rilis self-host (19 November 2026)

Dokumen ini mencatat apa yang benar-benar dijalankan dan apa yang belum.
Jangan menganggap klaim README sebagai bukti tanpa membaca bagian ini.

## Yang dijalankan

### Backend

- Struktur `cmd/vibelog/` + `internal/` (auth, config, database, domain,
  email, export, httpapi, netutil, ratelimit, web). Tidak ada paket `main`
  di luar `cmd/`, dan tidak ada file Go non-test di atas 250 baris.
- `go vet ./...` bersih, `gofmt -l cmd internal` kosong.
- `CGO_ENABLED=0 go test -count=1 ./...` **47/47 lolos** di 9 paket
  (`auth`, `config`, `database`, `domain`, `email`, `export`, `httpapi`,
  `netutil`, `ratelimit`), termasuk:
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
- `npm run build` sukses (130 modul, ±189,3 KB JS, ±16,1 KB CSS).
- Perubahan rilis: verifikasi magic link lewat `POST`, tanggal memakai zona
  waktu browser, bukan `Asia/Bangkok`.
- Halaman dashboard/landing dipecah menjadi komponen `DashboardTopbar`;
  `main.css` dipecah menjadi `base.css`, `layout.css`, dan
  `components.css` yang diimpor berurutan.
- Simplifikasi teks: landing dari 6 kartu berstatistik menjadi 4 kartu satu
  baris; CTA ganda dihapus; kartu entri kehilangan grid statistik ganda;
  header dashboard memakai satu menu Export; label form dipendekkan.
- Conditional UI: tombol GitHub, magic link, dan pendaftaran hanya tampil
  bila fiturnya aktif (`/api/meta`). Pendaftaran tertutup dan email `off`
  punya tampilan sendiri.
- Cleanup UX: blok ekspor duplikat di bawah dihapus, metadata ganda di kartu
  entri dihapus, ID parent tidak lagi ditampilkan, proyek aktif otomatis
  terpilih saat membuat entri, dan copy prompt punya fallback.
- `downloadBlob` menunda pelepasan blob URL agar unduhan tidak terpotong.
- Judul dokumen berubah per rute; footer landing menyebut lisensi MIT.

### End-to-end browser (Chrome DevTools Protocol)

Dijalankan terhadap binary produksi dengan database demo, viewport 1440x900
dan 390x844:

- **30/30 cek lolos**: guard `/app` tanpa token, register dari form, empty
  state, buat proyek, buat entri (proyek terpilih otomatis), edit, next
  iteration, filter pencarian/outcome/tag, copy prompt, ekspor Markdown,
  hapus iterasi, hapus proyek (entri bertahan), hapus entri, login password,
  password salah, magic link lewat UI + log console, register email duplikat,
  halaman 404, dan tidak ada overflow horizontal di 390px.
- **9/9 cek instance tertutup** (`ALLOW_REGISTRATION=false`, `EMAIL_DRIVER=off`):
  landing tanpa tombol daftar, login hanya password, halaman register
  menampilkan status tertutup, halaman email menjelaskan pengiriman mati.
- Catatan: Chrome headless membatalkan penulisan file unduhan. Yang diverifikasi
  adalah event `downloadWillBegin`, nama file `.md`, dan payload ekspor yang
  diambil halaman. Unduhan file nyata perlu dicek di browser biasa.

## Yang TIDAK dijalankan di mesin ini

- **Build image Docker tidak dijalankan** karena Docker tidak tersedia di
  mesin ini. Dockerfile dan compose ditulis untuk build dari root repo;
  verifikasi build image harus dilakukan di mesin dengan Docker:
  `docker compose up -d --build`.
- **CI GitHub Actions sudah hijau** di Ubuntu (Go 1.26, Node 22,
  `CGO_ENABLED=0`): typecheck, build frontend, `go vet`, dan 47 test lolos.
- **Release sudah dibuat dan workflow-nya hijau.** Tag `v1.0.0` dan `v1.0.1`
  menghasilkan binary linux amd64/arm64, macOS, Windows, plus checksums.
  `install.sh` diuji dari release sungguhan dengan `VIBELOG_DRY_RUN=1`:
  unduhan, checksum, dan ekstraksi lolos. Langkah `useradd` + systemd belum
  dijalankan karena mesin ini bukan Linux/root; itu bagian standar systemd.
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
go test -count=1 ./...
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
