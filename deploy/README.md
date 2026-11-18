# Deploy VibeLog

Cara yang didukung: **satu container** yang melayani frontend + API dari port
yang sama. Panduan lengkap ada di [`../docs/SELF_HOST.md`](../docs/SELF_HOST.md).

```bash
cp .env.example .env    # set BASE_URL
docker compose up -d --build
curl -fsS http://localhost:8090/health
```

Data persisten ada di volume `vibelog-data` (`/data`). Backup:

```bash
docker compose exec vibelog vibelog backup /data/backups
```

## Reverse proxy

`nginx-vibelog.conf` adalah contoh nginx dengan TLS yang meneruskan semua
request ke container di `127.0.0.1:8090`. Setelah proxy aktif, isi
`BASE_URL=https://domain-kamu` dan `TRUSTED_PROXIES` dengan alamat proxy.

## Tanpa Docker

```bash
make build
APP_ENV=production BASE_URL=https://vibelog.example.com \
DATABASE_PATH=/var/lib/vibelog/vibelog.db ./vibelog
```

Jalankan sebagai service systemd, lalu taruh nginx/Caddy di depannya.
