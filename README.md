# VibeLog, Jurnal Prompt dan Iterasi

VibeLog adalah jurnal prompt untuk vibe coder. Simpan prompt persis seperti
dikirim ke AI, catat model dan tool yang dipakai, nilai hasilnya, dan buat
iterasi berikutnya dalam satu klik. Entri yang terfilter bisa diekspor jadi
prompt pack Markdown atau JSON untuk dipakai ulang di AI agent atau editor.

Stack teknologi: **RESTful API, JWT, Vue, TypeScript, Auth,
Composables, TanStack Query**. Backend **Go** (stdlib + SQLite), frontend
**Vue 3 + TypeScript + Vite + Vue Router**.

## Fitur

- Auth JWT: daftar, masuk, profil. Password di-hash bcrypt. Token di
  `localStorage` dengan kunci `vibelog_token`.
- Proyek: CRUD, warna, jumlah entri. Menghapus proyek **tidak** menghapus
  entri jurnalnya, entri pindah ke Tanpa proyek.
- Entri jurnal: judul, prompt, model, tool, hasil (Berhasil, Sebagian,
  Gagal, Berjalan), rating 1-5, catatan hasil, tag, nomor iterasi.
- Iterasi berikutnya: duplikasi entri, nomor iterasi naik 1, parent menunjuk
  entri asal, hasil kembali ke Berjalan, rating dikosongkan.
- Filter: cari (debounce), proyek, model, tool, hasil, tag. Filter
  proyek "Tanpa proyek" memakai `project_id=0`.
- Ringkasan: total, per hasil, rata-rata rating, per model, per tool,
  per proyek, dan tag teratas. Ringkasan mengikuti filter proyek.
- Ekspor: Markdown dan JSON dari entri yang sedang terfilter, lewat
  endpoint server `GET /api/entries/export?format=md|json`.
- Isolasi data per user di semua endpoint. Akun lain tidak bisa membaca,
  mengubah, atau menghapus data akun lain (diuji).

## Menjalankan lokal

```bash
./scripts/run.sh
# Buka http://127.0.0.1:4173, daftar akun baru, lalu pakai aplikasinya.
```

Persyaratan: Go 1.24+, Node.js 20+, gcc (untuk driver SQLite mattn/go-sqlite3,
CGO). Di Windows, pakai WSL atau Docker. Database default `data/vibelog.db`.
Untuk memakai database lama, ganti nama file lama menjadi `vibelog.db`.

Variabel environment:

| Variabel | Default (dev) | Keterangan |
| --- | --- | --- |
| `APP_ENV` | `development` | `production` menolak JWT_SECRET kosong/default/pendek |
| `HOST` | `127.0.0.1` | Host backend |
| `PORT` | `8080` | Port backend |
| `DATABASE_PATH` | `data/vibelog.db` | Path file SQLite |
| `JWT_SECRET` | secret dev bawaan | Wajib di production, minimal 32 karakter |
| `JWT_EXPIRY_HOURS` | `24` | Masa berlaku token, 1-720 jam |
| `FRONTEND_URL` | `http://127.0.0.1:4173` | Satu-satunya origin CORS yang diizinkan |

## API

Semua endpoint di bawah `/api/` kecuali `/health` memakai header
`Authorization: Bearer <token>`. Error selalu JSON
`{"error":{"code":"...","message":"..."}}`.

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| GET | `/health` | Status server |
| POST | `/api/auth/register` | Daftar, mengembalikan token + user |
| POST | `/api/auth/login` | Masuk, pesan gagal disamakan untuk email/password salah |
| GET | `/api/auth/me` | Profil dari token |
| GET, POST | `/api/projects` | Daftar dan buat proyek |
| PATCH, DELETE | `/api/projects/{id}` | Ubah dan hapus proyek |
| GET, POST | `/api/entries` | Daftar (filter `q`, `project_id`, `model`, `tool`, `outcome`, `tag`, `limit`) dan buat entri |
| GET, PATCH, DELETE | `/api/entries/{id}` | Baca, ubah, hapus entri |
| POST | `/api/entries/{id}/duplicate` | Iterasi berikutnya |
| GET | `/api/entries/summary` | Ringkasan, opsional `project_id` |
| GET | `/api/entries/export` | Ekspor `format=md` atau `json`, mengikuti filter yang sama |

Validasi: judul maksimal 120 karakter, prompt dan catatan maksimal 10000
karakter, rating 1-5, maksimal 10 tag, body JSON maksimal 1 MB, field yang
tidak dikenal ditolak.

## Production

Lihat `deploy/README.md` untuk Docker Compose dan nginx. Ringkasnya:

1. Isi `JWT_SECRET` minimal 32 karakter yang acak, jangan pakai nilai dev.
2. Set `APP_ENV=production`, `FRONTEND_URL` ke domain frontend yang sebenarnya.
3. Jalankan lewat Docker Compose dengan volume data yang persisten.
4. Pakai HTTPS di depan aplikasi (nginx/Caddy/Cloudflare). Backend sendiri
   tidak mengakhiri TLS.
5. Cadangkan file `vibelog.db` secara berkala. Hapus akun seed/test sebelum
   dipakai publik, aplikasi ini tidak membuat akun seed.

Keamanan yang sudah diterapkan: bcrypt, JWT HS256 dengan issuer dan
algoritma yang dikunci, CORS satu origin, security headers, rate limit
(auth 20/menit/IP, API 300/menit/IP), timeout request 30 detik, graceful
shutdown, dan database WAL + foreign key. Lihat `SECURITY.md`.

## Pengembangan

```bash
make test   # go test + typecheck frontend
make build  # build backend dan frontend
```

Struktur: `backend/` (Go, satu package `main`, `store.go` untuk SQL,
`handlers.go` untuk HTTP), `frontend/src/` (`lib/api.ts` untuk fetch,
`composables/` untuk TanStack Query dan auth, `pages/` untuk rute,
`components/` untuk form dan kartu). Frontend memakai composables
`useAuth`, `useProjects`, `useEntries`, `useSummary`, dan
`useDebouncedValue`.

## Batasan yang diketahui

- Database SQLite untuk satu instance. Untuk multi-instance, pindah ke
  Postgres/MySQL dan ganti rate limiter in-memory dengan Redis.
- Rate limit in-memory per proses, tidak berbagi antar instance.
- Docker dan nginx belum dijalankan di mesin pembangun ini saat verifikasi
  awal, hanya sintaks dan build lokal yang diperiksa. Lihat
  `docs/verification.md` untuk bukti tiap klaim.
