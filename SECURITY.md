# Keamanan VibeLog

## Yang dilindungi

- Password di-hash bcrypt sebelum masuk database; hash tidak pernah dikirim
  ke frontend.
- JWT HS256 dengan issuer `vibelog-api`, algoritma dikunci, masa berlaku
  dari `JWT_EXPIRY_HOURS`.
- Di mode production, server menolak secret kosong, nilai development, atau
  kurang dari 32 karakter. Bila `JWT_SECRET` tidak diisi, secret acak dibuat
  dan disimpan dengan mode `0600` di samping database.
- Setiap entri dan proyek selalu disaring `user_id` dari token. Tidak ada
  endpoint yang menerima user id dari client.
- CORS hanya mengizinkan origin `BASE_URL`. Origin lain tidak mendapat
  header CORS.
- Security headers: CSP, `X-Content-Type-Options: nosniff`,
  `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`,
  dan `Cache-Control: no-store` untuk respons API.
- Rate limit: auth 20 request/menit/IP, API 300 request/menit/IP, dengan
  pembersihan bucket lama. `X-Forwarded-For` hanya dipercaya dari
  `TRUSTED_PROXIES`.
- Body JSON dibatasi 1 MB, field tak dikenal ditolak, request timeout 30 detik.
- Magic link: token 32 byte acak, berlaku 15 menit, sekali pakai, dan hanya
  tautan terbaru yang aktif. Verifikasi lewat `POST` supaya pemindai email
  tidak membakar tautan. Akun baru dibatalkan bila email gagal dikirim.
- Menutup pendaftaran (`ALLOW_REGISTRATION=false`) juga menutup pembuatan
  akun lewat magic link dan GitHub untuk email yang belum terdaftar.
- Menghapus proyek tidak menghapus entri jurnal (foreign key
  `ON DELETE SET NULL`), supaya riwayat prompt tidak hilang tanpa sengaja.

## Yang harus dilakukan operator

- Jalankan di belakang HTTPS. Jangan expose database SQLite ke publik.
- Isi `TRUSTED_PROXIES` dengan IP/CIDR proxy saja. Jangan `0.0.0.0/0`.
- Isi `JWT_SECRET` sendiri bila ingin mengelola rotasi secret; kalau tidak,
  secret otomatis sudah aman. Jangan commit `.env`.
- Tutup pendaftaran setelah akun sendiri dibuat, kecuali instance memang
  untuk banyak pengguna.
- Cadangkan database secara berkala (`vibelog backup`) dan uji restore.
- Perbarui dependensi Go dan npm secara berkala (`go get -u`, `npm audit`).

## Melaporkan masalah keamanan

Jangan buka issue publik untuk kerentanan yang bisa dieksploitasi. Hubungi
pemilik deployment langsung dengan detail reproduksi dan dampaknya.
