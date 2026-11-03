# Verifikasi VibeLog (2026-11-01, diperbarui 2026-11-02)

Dokumen ini membedakan apa yang benar-benar diperiksa dari apa yang belum.
Jangan jadikan klaim di README sebagai bukti tanpa membaca bagian ini.

## Bug yang ditemukan saat preview dan sudah diperbaiki (2026-11-02)

- Ringkasan akun tanpa entri sempat mengembalikan HTTP 500. Penyebabnya:
  `SUM(CASE ...)` atas himpunan kosong menghasilkan NULL yang tidak bisa
  di-scan ke `int` di `store.go`. Perbaikannya memakai `COALESCE(..., 0)`.
  Tes regresi `TestSummaryEmptyUser` ditambahkan. Setelah perbaikan,
  `go test` menjadi 7/7 dan cek browser lokal kedua berjalan tanpa error
  konsol sama sekali.
- Tombol "Daftar gratis" di landing sempat kontras rendah (teks gelap di
  latar teal) karena selektor `.topbar a`. Diperbaiki dengan selektor
  `.topbar nav > a:not(.btn)` dan warna putih eksplisit untuk
  `.topbar .btn.primary`. Screenshot landing diambil ulang setelah
  perbaikan ini.

## Restyle gaya referensi (2026-11-02)

- Seluruh frontend di-restye total mengikuti DNA desain referensi
  ChainGPT Labs yang dikirim user: bingkai tebal membulat, grid garis
  halus, judul pixel Press Start 2P, aksen oranye, kartu statistik 2x2
  dengan sel panah, dan label mono kecil. Font self-hosted di
  `frontend/public/fonts/` (Press Start 2P, Space Grotesk, Space Mono).
- Yang disamakan hanya pola layout, tipografi, dan sistem grid. Brand,
  nama proyek, dan angka statistik referensi TIDAK disalin: semua teks
  dan angka berasal dari data akun VibeLog sendiri.
- Cek browser lokal setelah restyle: daftar, proyek, entri, perluas
  prompt, duplikasi Iterasi 2, reload (2 kartu bertahan), cari, dan
  lebar 390px tanpa overflow (0px), tanpa error konsol. Screenshot baru
  di `~/workspace/previews/vibelog/` (landing, dashboard atas, dashboard
  entri desktop, dashboard mobile). Backend tidak berubah pada restyle
  ini, kecuali tidak ada perubahan sama sekali di `backend/`.

## Yang diperiksa dan lolos

- `go vet` + `go test -count=1 ./...` di `backend/`: **7/7 lolos**
  (`TestConfigProductionRejectsWeakSecret`,
  `TestPasswordAndTokenRoundTrip`, `TestEntryValidation`,
  `TestAuthDuplicateEmailAndWrongPassword`,
  `TestProjectEntryFlowIsolationDuplicateSummaryExport`,
  `TestBuildMarkdownEscapesFence`). Termasuk isolasi antar user, duplikat
  iterasi 2, ringkasan, filter `project_id=0`, dan entri yang bertahan
  dengan `project_id` null setelah proyek dihapus.
- Cek API end-to-end dengan curl terhadap backend + frontend preview yang
  berjalan: **21/21 lolos**. Rinciannya: health, 401 tanpa token, register,
  email duplikat 409, password salah 401, `/me`, buat proyek, buat entri,
  duplikat iterasi 2, ringkasan total 2, filter q+outcome, ekspor Markdown
  dan JSON, validasi outcome 400, isolasi user lain 404, hapus proyek,
  entri bertahan dengan `project_id` null (cek ini sempat gagal di skrip
  karena bug quoting di skripnya, lalu diperiksa ulang manual dan lolos),
  filter tanpa proyek 2 entri, landing 200, rute `/app` dilayani SPA, dan
  mode production menolak `JWT_SECRET` pendek. Server uji dimatikan
  setelah pemeriksaan.
- Build frontend: `vue-tsc --noEmit && vite build` lolos. Satu error
  typecheck sempat muncul (`auth.user` di template Dashboard) dan sudah
  diperbaiki dengan destructure `useAuth`, lalu build diulang sampai lolos.
- Grep teks terlihat: tidak ada em-dash di `frontend/src`, README, dan
  dokumen. Tidak ada penyebutan Saku, Bukuku, atau Rafif di proyek ini.

## Yang TIDAK terbukti (diperbarui 2026-11-02)

- **Cek klik di browser lokal kini terbukti** lewat Chromium Playwright
  yang berjalan di VM yang sama (bukan tugas browser terkelola, yang
  gagal di langkah 1 karena sandbox-nya tidak bisa mencapai loopback VM).
  Alur yang diklik sungguhan: daftar akun baru, buat proyek, buat entri,
  perluas prompt, duplikasi jadi Iterasi 2, reload (2 kartu tetap ada),
  cari, dan lebar 390px tanpa overflow horizontal (0px), tanpa error
  konsol pada run kedua. Screenshot ada di `~/workspace/previews/vibelog/`.
  Cek ini memakai database sementara di `/tmp`, bukan database proyek.
- **Verifikasi sistem menyeluruh 2026-11-02 (user: "pastikan sistemnya
  benar-benar berjalan"):** zip di-extract bersih ke folder baru (62 file,
  tanpa node_modules/dist/database), `go test` 7/7 lolos dari hasil
  extract itu, lalu backend + frontend dibangun dari nol dan dijalankan.
  Uji API 27/27 lulus: register, email duplikat 409, pesan login gagal
  sama untuk email/password salah, /me, tanpa token 401, proyek dan
  entri CRUD, ringkasan, duplikasi iterasi 2, filter q/proyek/hasil/tag,
  ekspor Markdown dan JSON, isolasi antar user (404 + ringkasan 0),
  hapus proyek (entri bertahan, project_id null), hapus entri, dan
  validasi (outcome/rating/field tak dikenal 400). Restart backend
  dengan database yang sama: total entri tetap 2, token lama tetap
  berlaku (200), judul entri tetap ada. Mode production menolak
  JWT_SECRET pendek (exit 1, pesan jelas). Alur browser 19/19 langkah
  diklik sungguhan, 0 error konsol/API di luar 401 yang memang disengaja
  pada langkah password salah: landing, daftar, keluar, password salah
  (pesan tampil, tetap di login), login benar, 2 proyek, ubah nama
  proyek, buat entri, perluas prompt, duplikasi, ubah entri, cari,
  filter tag, ekspor Markdown terunduh berisi judul entri, hapus proyek,
  hapus entri, reload, keluar lalu login lagi (data tetap), dan lebar
  390px tanpa overflow. Catatan: selama pengujian, limit auth 20/menit
  per IP sempat tercapai oleh banyaknya percobaan uji dari mesin yang
  sama; itu perilaku pengaman yang benar, bukan bug, dan pengujian
  diulang setelah jeda. Tiga kegagalan awal pada skrip uji ternyata
  kesalahan skrip uji sendiri (bentuk respons array, DELETE 200, dan
  field form), bukan bug aplikasi, dan sudah diperbaiki di skrip lalu
  lulus semua.
- **Docker dan nginx tidak dijalankan.** Hanya sintaks/file yang ditulis;
  build image dan `docker compose up` belum dijalankan di mesin ini.
- **Windows dan MySQL tidak diuji.** Database yang dipakai SQLite satu
  instance. Jalankan di Windows lewat WSL atau Docker, dan uji sendiri
  sebelum production di Windows.
- **Tidak ada akun seed.** Aplikasi tidak membuat akun apa pun; akun uji
  yang dipakai saat verifikasi (`ana@test.id`, `budi@test.id`,
  `browser2@test.id` yang gagal dibuat karena browser tidak sampai) hanya
  ada di database sementara `/tmp` yang sudah tidak dipakai, bukan di
  database proyek.

## Cara user memverifikasi

```bash
./scripts/run.sh
# Buka http://127.0.0.1:4173, daftar, buat proyek, buat entri,
# klik Iterasi berikutnya, reload, lalu cek di lebar HP.
```
