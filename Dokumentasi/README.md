# VibeLog

**Jurnal prompt dan iterasi untuk vibe coder.**

VibeLog adalah aplikasi web untuk menyimpan prompt persis seperti yang dikirim ke AI, mencatat model dan tool yang dipakai, menilai hasilnya, lalu membuat iterasi berikutnya dalam satu klik. Entri yang sudah terfilter dapat diekspor menjadi prompt pack Markdown atau JSON untuk dipakai ulang di AI agent maupun editor.

Status: prototipe versi 1.0, dalam pengembangan aktif.
Semua data pada screenshot adalah data demo dari akun fiktif, bukan data pengguna asli.

---

## Ringkasan

Saat bekerja dengan AI, prompt yang bagus sering hilang di riwayat chat. VibeLog menyimpannya sebagai jurnal yang terstruktur: setiap entri berisi judul, prompt lengkap, model, tool, hasil (Berhasil, Sebagian, Gagal, Berjalan), rating 1 sampai 5, catatan hasil, dan tag. Riwayat percobaan tersimpan sebagai iterasi, sehingga pengguna tahu prompt mana yang berhasil dan apa yang diubah pada percobaan berikutnya.

Aplikasi ini adalah SPA Vue 3 dengan REST API Go dan database SQLite. Autentikasi memakai JWT, dengan dukungan login password, GitHub OAuth, dan magic link lewat email. Seluruh data dipisahkan per akun.

---

## Detail Deskripsi

### 1. Masalah yang diselesaikan

- Prompt yang berhasil tidak terdokumentasi dan sulit ditemukan kembali di riwayat chat.
- Tidak ada catatan model dan tool apa yang dipakai saat sebuah prompt berhasil.
- Iterasi prompt tercampur dengan percobaan lama tanpa penanda nomor iterasi.
- Prompt bagus tidak bisa dipakai ulang di agent atau editor lain dengan cepat.

### 2. Solusi

VibeLog menyatukan pencatatan prompt, penilaian hasil, dan ekspor dalam satu alur:

1. Simpan prompt tepat seperti yang dikirim ke AI.
2. Catat model, tool, hasil, rating, catatan, dan tag.
3. Klik **Next iteration** untuk menduplikasi entri sebagai percobaan baru. Nomor iterasi naik, entri baru menunjuk ke entri asal, hasil kembali ke In Progress, dan rating dikosongkan.
4. Filter berdasarkan proyek, model, tool, hasil, tag, atau kata kunci.
5. Ekspor hasil filter sebagai prompt pack Markdown atau JSON.

### 3. Fitur utama

**Autentikasi**
- Registrasi dan login dengan email dan password. Password di-hash dengan bcrypt.
- Login GitHub OAuth dengan verifikasi state untuk mencegah CSRF.
- Login magic link lewat email (Resend), token sekali pakai dengan masa berlaku 15 menit.
- Sesi JWT disimpan di browser dan dipulihkan saat aplikasi dibuka kembali.

**Proyek**
- CRUD proyek dengan nama, deskripsi, warna, dan jumlah entri.
- Filter proyek termasuk opsi **No project**.
- Menghapus proyek tidak menghapus entri jurnal. Entri dipindahkan ke No project.

**Entri jurnal**
- Judul, prompt lengkap, model AI, tool, hasil, rating 1 sampai 5, catatan hasil, dan maksimal 10 tag.
- Nomor iterasi dan tautan ke entri asal.
- Aksi Edit, Next iteration, Copy prompt, dan Delete.

**Pencarian dan filter**
- Pencarian judul, prompt, catatan, dan tag dengan debounce.
- Filter proyek, model, tool, hasil, dan tag.
- Tag populer dapat diklik untuk langsung memfilter.

**Ringkasan**
- Total entri, jumlah per hasil, rata-rata rating, dan jumlah entri yang dinilai.
- Distribusi per model, per tool, per proyek, dan tag teratas.

**Ekspor**
- Prompt pack Markdown yang dikelompokkan per proyek, siap ditempel ke AI agent atau editor.
- Ekspor JSON dari entri yang sedang terfilter.

### 4. Alur pengguna

1. Pengguna membuka landing page, lalu mendaftar atau masuk.
2. Membuat proyek, misalnya "Kasir API" atau "Portfolio Website".
3. Menyimpan prompt yang baru dikirim ke AI, lengkap dengan model, tool, dan hasilnya.
4. Memberi rating dan catatan setelah melihat hasil.
5. Menekan **Next iteration** untuk mencoba prompt yang sudah diperbaiki.
6. Memfilter entri yang berhasil, lalu mengekspornya sebagai prompt pack.

### 5. Arsitektur dan teknologi

| Lapisan | Teknologi |
| --- | --- |
| Frontend | Vue 3, TypeScript, Vite, Vue Router, TanStack Query |
| Backend | Go (standard library), REST API, JWT HS256 |
| Database | SQLite (WAL, foreign key aktif) |
| Email | Resend API untuk magic link |
| OAuth | GitHub OAuth App |
| Deployment | Docker Compose (Go + nginx), Vercel untuk frontend |

Struktur repositori:

```
backend/     API Go, store SQL, auth, middleware, validasi, ekspor
frontend/    SPA Vue, halaman, komponen, composable, tipe
deploy/      Konfigurasi Docker dan nginx
docs/        Catatan teknis dan verifikasi
scripts/     Skrip menjalankan aplikasi secara lokal
```

### 6. Model data

- `users`: akun, email, hash password, status verifikasi email.
- `projects`: proyek milik user, warna, deskripsi.
- `entries`: entri jurnal, relasi ke proyek dan entri induk, nomor iterasi, rating, tag.
- `otp_codes`: token sekali pakai untuk magic link.

Semua query entri dan proyek selalu disaring dengan `user_id` dari token, bukan dari input klien.

### 7. API utama

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| POST | `/api/auth/register` | Daftar akun, mengembalikan token |
| POST | `/api/auth/login` | Masuk dengan email dan password |
| GET | `/api/auth/github` | Mulai login GitHub OAuth |
| POST | `/api/auth/email/request` | Kirim magic link |
| GET | `/api/auth/email/verify` | Verifikasi magic link |
| GET | `/api/auth/me` | Profil dari token |
| GET, POST | `/api/projects` | Daftar dan buat proyek |
| PATCH, DELETE | `/api/projects/{id}` | Ubah dan hapus proyek |
| GET, POST | `/api/entries` | Daftar dengan filter, dan buat entri |
| GET, PATCH, DELETE | `/api/entries/{id}` | Baca, ubah, hapus entri |
| POST | `/api/entries/{id}/duplicate` | Buat iterasi berikutnya |
| GET | `/api/entries/summary` | Ringkasan jurnal |
| GET | `/api/entries/export` | Ekspor Markdown atau JSON |

Error selalu berbentuk `{"error":{"code":"...","message":"..."}}`.

### 8. Keamanan

- Password di-hash dengan bcrypt dan tidak pernah dikirim ke frontend.
- JWT memakai algoritma HS256 yang dikunci, issuer tetap, dan masa berlaku terbatas.
- Mode production menolak `JWT_SECRET` kosong, default, atau kurang dari 32 karakter.
- CORS hanya mengizinkan origin frontend yang dikonfigurasi.
- Security header: `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Cache-Control`.
- Rate limit untuk endpoint autentikasi dan API.
- Body JSON dibatasi 1 MB dan field yang tidak dikenal ditolak.
- Isolasi data per user diuji, termasuk percobaan akses lintas akun.

### 9. Batasan saat ini

- SQLite ditujukan untuk satu instance. Untuk multi-instance perlu Postgres/MySQL atau layanan SQLite terkelola.
- Rate limiter masih disimpan di memori per proses.
- Magic link membutuhkan `RESEND_API_KEY` dan domain terverifikasi di Resend.
- Login GitHub membutuhkan callback HTTPS yang terdaftar di GitHub OAuth App.
- Belum ada fitur ubah password, lupa password, atau pengaturan profil.

### 10. Menjalankan secara lokal

Prasyarat: Go 1.24+, Node.js 20+, dan gcc untuk driver SQLite.

```bash
./scripts/run.sh
# Buka http://127.0.0.1:4173
```

Variabel environment utama: `APP_ENV`, `HOST`, `PORT`, `DATABASE_PATH`, `JWT_SECRET`, `JWT_EXPIRY_HOURS`, `FRONTEND_URL`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `RESEND_API_KEY`.

---

## Isi Folder Dokumentasi

| File | Ukuran | Keterangan |
| --- | --- | --- |
| `01-landing.png` | 2880 x 1800 | Landing page: hero, fitur, dan prompt pack |
| `02-dashboard.png` | 2880 x 1800 | Dashboard: statistik, proyek, filter, dan entri |
| `03-entry-prompt.png` | 2880 x 1800 | Kartu entri dengan prompt terbuka dan tag |
| `thumbnail.png` | 2560 x 1440 | Thumbnail cover dengan kartu entri contoh |
| `README.md` | - | Dokumen ini |

Urutan carousel yang disarankan: `01-landing.png`, `02-dashboard.png`, `03-entry-prompt.png`.

Ketiga screenshot memakai viewport 1440 x 900 dengan skala 2x. Thumbnail memakai rasio 16:9, cocok untuk cover repositori, unggulan portofolio, atau kartu sosial.
