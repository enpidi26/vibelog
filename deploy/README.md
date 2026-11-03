# Deploy VibeLog

## Docker Compose (disarankan)

```bash
cp .env.example .env
# Isi JWT_SECRET dan FRONTEND_URL di .env, lalu:
docker compose up -d --build
```

Frontend dilayani nginx di port 80, backend di port 8080 di dalam jaringan
Compose. Data SQLite disimpan di volume `vibelog-data` (`/data/vibelog.db`).
Cadangkan volume itu secara berkala.

## nginx di VPS (tanpa Docker)

1. Build backend (`go build`) dan frontend (`npm run build`).
2. Jalankan backend sebagai service systemd dengan environment production.
3. Salin `deploy/nginx-vibelog.conf` ke konfigurasi nginx, ganti
   `vibelog.example.com` dengan domain sebenarnya, dan aktifkan HTTPS
   (misalnya dengan Certbot) sebelum dipakai publik.
4. Arahkan root nginx ke folder `frontend/dist`, dan proxy `/api` serta
   `/health` ke backend.

## Variabel production yang wajib benar

`APP_ENV=production`, `JWT_SECRET` minimal 32 karakter, `FRONTEND_URL`
sama persis dengan origin browser (termasuk https dan tanpa trailing slash),
dan `DATABASE_PATH` ke lokasi yang persisten dan tercadangkan.
