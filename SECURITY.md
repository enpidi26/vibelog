# Keamanan VibeLog

## Yang dilindungi

- Password tidak pernah disimpan mentah. Password di-hash dengan bcrypt
  sebelum masuk database, dan hash tidak pernah dikirim ke frontend.
- Token JWT HS256 dengan issuer `vibelog-api`, algoritma dikunci ke HS256,
  dan masa berlaku dari `JWT_EXPIRY_HOURS`.
- Di mode production, server menolak jalan kalau `JWT_SECRET` kosong, masih
  nilai development, atau kurang dari 32 karakter.
- Setiap entri dan proyek selalu disaring `user_id` dari token. Tidak ada
  endpoint yang menerima user id dari client.
- CORS hanya mengizinkan origin `FRONTEND_URL`. Origin lain tidak mendapat
  header CORS.
- Security headers: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
  `Referrer-Policy: no-referrer`, `Cache-Control: no-store` untuk respons API.
- Rate limit: auth 20 request/menit/IP, API 300 request/menit/IP.
- Body JSON dibatasi 1 MB dan field yang tidak dikenal ditolak.
- Menghapus proyek tidak menghapus entri jurnal (foreign key
  `ON DELETE SET NULL`), supaya riwayat prompt tidak hilang tanpa sengaja.

## Yang harus dilakukan operator

- Isi `JWT_SECRET` yang acak dan rahasia, minimal 32 karakter. Jangan
  commit secret ke Git. Pakai `.env` yang tidak ter-commit atau secret
  manager deployment.
- Jalankan di belakang HTTPS. Jangan expose database SQLite ke publik.
- Cadangkan `vibelog.db` secara berkala dan uji restore-nya.
- Perbarui dependensi Go dan npm secara berkala (`go get -u`, `npm audit`).

## Melaporkan masalah keamanan

Jangan buka issue publik untuk kerentanan yang bisa dieksploitasi. Hubungi
pemilik deployment langsung dengan detail reproduksi dan dampaknya.
